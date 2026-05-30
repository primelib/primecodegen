package openapi_kotlin_multiplatform

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
)

func TestToCodeTypeDynamicFallbacksUseJsonElement(t *testing.T) {
	g := NewGenerator()

	multiType, err := g.ToCodeType(&base.Schema{Type: []string{"string", "integer"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", multiType.Name)
	assert.Equal(t, "kotlinx.serialization.json", multiType.ImportPath)

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
	assert.Equal(t, "kotlinx.serialization.json", addPropsTrue.TypeArgs[1].ImportPath)

	freeformObject, err := g.ToCodeType(&base.Schema{Type: []string{"object"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", freeformObject.Name)
	assert.Equal(t, "kotlinx.serialization.json", freeformObject.ImportPath)

	heterogeneousOneOf, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		OneOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", heterogeneousOneOf.Name)
	assert.Equal(t, "kotlinx.serialization.json", heterogeneousOneOf.ImportPath)
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

func TestToCodeTypePatternPropertiesUsesJsonFallbackForMixedTypes(t *testing.T) {
	g := NewGenerator()

	patternProperties := orderedmap.New[string, *base.SchemaProxy]()
	patternProperties.Set("^[a-z]+$", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
	patternProperties.Set("^[0-9]+$", base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}))

	codeType, err := g.ToCodeType(&base.Schema{
		Type:              []string{"object"},
		PatternProperties: patternProperties,
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.True(t, codeType.IsMap)
	assert.Len(t, codeType.TypeArgs, 2)
	assert.Equal(t, "JsonElement", codeType.TypeArgs[1].Name)
}

func TestToCodeTypePatternPropertiesUsesSingleMappedType(t *testing.T) {
	g := NewGenerator()

	patternProperties := orderedmap.New[string, *base.SchemaProxy]()
	patternProperties.Set("^[a-z]+$", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))

	codeType, err := g.ToCodeType(&base.Schema{
		Type:              []string{"object"},
		PatternProperties: patternProperties,
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.True(t, codeType.IsMap)
	assert.Len(t, codeType.TypeArgs, 2)
	assert.Equal(t, "String", codeType.TypeArgs[1].Name)
}

func TestToCodeTypeObjectWithAdditionalPropertiesFalseFallsBackToJsonElement(t *testing.T) {
	g := NewGenerator()

	codeType, err := g.ToCodeType(&base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: false,
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", codeType.Name)
}

func TestToCodeTypeCompositionFallbacksUseJsonElement(t *testing.T) {
	g := NewGenerator()

	anyOfType, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		AnyOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", anyOfType.Name)

	allOfType, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		AllOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", allOfType.Name)

	notType, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		Not:  base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "JsonElement", notType.Name)
}
