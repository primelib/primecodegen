package openapi_kotlin

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
)

func TestToCodeTypeDynamicFallbacksUseAny(t *testing.T) {
	g := NewGenerator()

	multiType, err := g.ToCodeType(&base.Schema{Type: []string{"string", "integer"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "Any", multiType.Name)

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
	assert.Equal(t, "Any", addPropsTrue.TypeArgs[1].Name)

	freeformObject, err := g.ToCodeType(&base.Schema{Type: []string{"object"}}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "Any", freeformObject.Name)

	heterogeneousOneOf, err := g.ToCodeType(&base.Schema{
		Type: []string{},
		OneOf: []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}}),
		},
	}, openapigenerator.CodeTypeSchemaProperty, true)
	assert.NoError(t, err)
	assert.Equal(t, "Any", heterogeneousOneOf.Name)
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
