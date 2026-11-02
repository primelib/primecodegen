package openapi_java

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
)

func toPostProcessedCodeType(t *testing.T, g *JavaGenerator, schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) openapigenerator.CodeType {
	t.Helper()
	codeType, err := g.ToCodeType(schema, schemaType, required)
	assert.NoError(t, err)
	return g.PostProcessType(codeType)
}

func TestToCodeTypeBinaryIsPrimitiveByteArray(t *testing.T) {
	g := NewGenerator()

	for _, format := range []string{"binary", "byte"} {
		codeType := toPostProcessedCodeType(t, g, &base.Schema{Type: []string{"string"}, Format: format}, openapigenerator.CodeTypeSchemaResponse, false)
		assert.Equal(t, "byte[]", codeType.Declaration, format)
		assert.Equal(t, "byte[]", codeType.QualifiedType, format)
	}
}

func TestToCodeTypeDateFormats(t *testing.T) {
	g := NewGenerator()

	date := toPostProcessedCodeType(t, g, &base.Schema{Type: []string{"string"}, Format: "date"}, openapigenerator.CodeTypeSchemaProperty, false)
	assert.Equal(t, "LocalDate", date.Declaration)
	assert.Equal(t, "java.time.LocalDate", date.QualifiedType)

	dateTime := toPostProcessedCodeType(t, g, &base.Schema{Type: []string{"string"}, Format: "date-time"}, openapigenerator.CodeTypeSchemaProperty, false)
	assert.Equal(t, "Instant", dateTime.Declaration)
	assert.Equal(t, "java.time.Instant", dateTime.QualifiedType)
}

func TestToCodeTypeBoxing(t *testing.T) {
	g := NewGenerator()
	notNullable := false
	intSchema := func() *base.Schema {
		return &base.Schema{Type: []string{"integer"}, Format: "int32", Nullable: &notNullable}
	}

	// optional values can be absent, even if not nullable
	optional := toPostProcessedCodeType(t, g, intSchema(), openapigenerator.CodeTypeSchemaProperty, false)
	assert.Equal(t, "Integer", optional.Declaration)

	// required and not nullable values are always present
	required := toPostProcessedCodeType(t, g, intSchema(), openapigenerator.CodeTypeSchemaProperty, true)
	assert.Equal(t, "int", required.Declaration)

	// parameters are validated using null checks
	requiredParam := toPostProcessedCodeType(t, g, intSchema(), openapigenerator.CodeTypeSchemaParameter, true)
	assert.Equal(t, "Integer", requiredParam.Declaration)

	// type arguments must be boxed
	list := toPostProcessedCodeType(t, g, &base.Schema{
		Type:  []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(intSchema())},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.Equal(t, "List<Integer>", list.Declaration)
}

func TestPostProcessTypeQualifiesTypeArguments(t *testing.T) {
	g := NewGenerator()

	list := toPostProcessedCodeType(t, g, &base.Schema{
		Type:  []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}, Format: "date-time"})},
	}, openapigenerator.CodeTypeSchemaResponse, false)
	assert.Equal(t, "List<Instant>", list.Declaration)
	assert.Equal(t, "List<java.time.Instant>", list.QualifiedType)

	uuidMap := toPostProcessedCodeType(t, g, &base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			A: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}, Format: "uuid"}),
		},
	}, openapigenerator.CodeTypeSchemaResponse, false)
	assert.Equal(t, "Map<String, java.util.UUID>", uuidMap.QualifiedType)
}

func TestToCodeTypePropertiesTakePrecedenceOverAdditionalProperties(t *testing.T) {
	g := NewGenerator()

	properties := orderedmap.New[string, *base.SchemaProxy]()
	properties.Set("name", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
	codeType := toPostProcessedCodeType(t, g, &base.Schema{
		Type:       []string{"object"},
		Title:      "User",
		Properties: properties,
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: true,
		},
	}, openapigenerator.CodeTypeSchemaProperty, false)
	assert.False(t, codeType.IsMap)
	assert.Equal(t, "User", codeType.Declaration)
}

func TestToCodeTypeOneOfComparesFullType(t *testing.T) {
	g := NewGenerator()

	listOf := func(title string) *base.SchemaProxy {
		properties := orderedmap.New[string, *base.SchemaProxy]()
		properties.Set("id", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
		return base.CreateSchemaProxy(&base.Schema{
			Type:  []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{Type: []string{"object"}, Title: title, Properties: properties})},
		})
	}

	// List<A> and List<B> have the same (empty) name, but are different types
	codeType := toPostProcessedCodeType(t, g, &base.Schema{OneOf: []*base.SchemaProxy{listOf("A"), listOf("B")}}, openapigenerator.CodeTypeSchemaProperty, false)
	assert.Equal(t, "Object", codeType.Declaration)
}
