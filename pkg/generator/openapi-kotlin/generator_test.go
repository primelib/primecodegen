package openapi_kotlin

import (
	"testing"

	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
)

func TestToCodeTypeDynamicFallbacksUseJsonElement(t *testing.T) {
	g := NewGenerator()

	multiType, err := g.ToCodeType(&base.Schema{Type: []string{"string", "integer"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", multiType.Name)

	addPropsTrue, err := g.ToCodeType(&base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: true,
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.True(t, addPropsTrue.IsMap)
	assert.Len(t, addPropsTrue.TypeArgs, 2)
	assert.Equal(t, "JsonElement", addPropsTrue.TypeArgs[1].Name)

	freeformObject, err := g.ToCodeType(&base.Schema{Type: []string{"object"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", freeformObject.Name)

	heterogeneousOneOf, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		OneOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", heterogeneousOneOf.Name)
}

func TestToCodeTypeNullableUnionPrefersConcreteType(t *testing.T) {
	g := NewGenerator()

	codeType, err := g.ToCodeType(&base.Schema{Type: []string{"string", "null"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "String", codeType.Name)
}

func TestToCodeTypeArrayRequiresItems(t *testing.T) {
	g := NewGenerator()

	_, err := g.ToCodeType(&base.Schema{Type: []string{"array"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.ErrorContains(t, err, "array schema missing items definition")
}

func TestToCodeTypeOneOfWithoutType(t *testing.T) {
	g := NewGenerator()

	// libopenapi leaves Type nil if the schema has no type, which previously matched the object case first
	codeType, err := g.ToCodeType(&base.Schema{
		OneOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "String", codeType.Name)
}

func TestToCodeTypeDateFormats(t *testing.T) {
	g := NewGenerator()

	date, err := g.ToCodeType(&base.Schema{Type: []string{"string"}, Format: "date"}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "LocalDate", date.Name)
	assert.Equal(t, "kotlinx.datetime", date.ImportPath)

	dateTime, err := g.ToCodeType(&base.Schema{Type: []string{"string"}, Format: "date-time"}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "Instant", dateTime.Name)
}

func TestToCodeTypeUint64(t *testing.T) {
	g := NewGenerator()

	codeType, err := g.ToCodeType(&base.Schema{Type: []string{"integer"}, Format: "uint64"}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "ULong", codeType.Name)
}

func TestToCodeTypePropertiesTakePrecedenceOverAdditionalProperties(t *testing.T) {
	g := NewGenerator()

	properties := orderedmap.New[string, *base.SchemaProxy]()
	properties.Set("name", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
	codeType, err := g.ToCodeType(&base.Schema{
		Type:       []string{"object"},
		Title:      "User",
		Properties: properties,
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: true,
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.False(t, codeType.IsMap)
	assert.Equal(t, "User", codeType.Name)
}

func TestPostProcessTypeQualifiesTypeArguments(t *testing.T) {
	g := NewGenerator()

	codeType := g.PostProcessType(openapigenerator.NewListCodeType(openapigenerator.CodeType{Name: "Instant", ImportPath: "kotlin.time"}, &base.Schema{}))
	assert.Equal(t, "List<Instant>", codeType.Declaration)
	assert.Equal(t, "List<kotlin.time.Instant>", codeType.QualifiedType)
}
