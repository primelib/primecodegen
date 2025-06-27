package openapi_java

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
)

var commonPackages = openapigenerator.CommonPackages{
	Root:       "",
	Client:     "",
	Models:     "",
	Enums:      "",
	Operations: "",
	Auth:       "",
}

var (
	//go:embed specs/operation-basic.yaml
	operationBasic []byte
	//go:embed specs/model-basic.yaml
	modelBasic []byte
	//go:embed specs/model-array-of-string.yaml
	modelArrayOfString []byte
	//go:embed specs/model-array-of-map.yaml
	modelArrayOfMap []byte
	//go:embed specs/model-array-oneof.yaml
	modelArrayOfOneOf []byte
	//go:embed specs/callback-basic.yaml
	callbackBasic []byte
	//go:embed specs/webhook-basic.yaml
	webhookBasic []byte
)

func TestOperationBasic(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(operationBasic)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Operations, 1)
	assert.Equal(t, "CreateBook", templateData.Operations[0].Name)
	assert.Equal(t, "post", templateData.Operations[0].Method)
	assert.Equal(t, "/books", templateData.Operations[0].Path)
	assert.Len(t, templateData.Models, 1)
	assert.Equal(t, "BookDto", templateData.Models[0].Name)
	assert.Equal(t, true, templateData.Models[0].IsTypeAlias)
	assert.Equal(t, "Object", templateData.Models[0].Parent.QualifiedType)
}

func TestBasicModel(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(modelBasic)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 1)
	assert.Equal(t, "BookDto", templateData.Models[0].Name)
	assert.Equal(t, "title", templateData.Models[0].Properties[0].Name)
	assert.Equal(t, "String", templateData.Models[0].Properties[0].Type.QualifiedType)
	assert.Equal(t, "author", templateData.Models[0].Properties[1].Name)
	assert.Equal(t, "String", templateData.Models[0].Properties[1].Type.QualifiedType)
	assert.Equal(t, "year", templateData.Models[0].Properties[2].Name)
	assert.Equal(t, "Long", templateData.Models[0].Properties[2].Type.QualifiedType)
	assert.Equal(t, "price", templateData.Models[0].Properties[3].Name)
	assert.Equal(t, "Double", templateData.Models[0].Properties[3].Type.QualifiedType)
}

func TestArrayOfStringModel(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(modelArrayOfString)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 1)
	assert.Equal(t, "BookDto", templateData.Models[0].Name)
	assert.Equal(t, true, templateData.Models[0].IsTypeAlias)
	assert.Equal(t, "List<String>", templateData.Models[0].Parent.QualifiedType)
}

func TestArrayOfMapModel(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(modelArrayOfMap)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 2)
	assert.Equal(t, "BookDto", templateData.Models[0].Name)
	assert.Equal(t, true, templateData.Models[0].IsTypeAlias)
	assert.Equal(t, "List<Map<String, String>>", templateData.Models[0].Parent.QualifiedType)
}

func TestArrayOfOneOf(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(modelArrayOfOneOf)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 1)
	assert.Equal(t, "BookDto", templateData.Models[0].Name)
	assert.Equal(t, true, templateData.Models[0].IsTypeAlias)
	assert.Equal(t, "List<Object>", templateData.Models[0].Parent.QualifiedType)
}

func TestCallbackBasic(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(callbackBasic)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 1)
}

func TestWebhookBasic(t *testing.T) {
	// arrange
	v3doc := openapidocument.OpenV3DocumentForTest(webhookBasic)

	// act
	templateData, err := openapigenerator.BuildTemplateData(v3doc, NewGenerator(), commonPackages)
	assert.NoError(t, err)
	assert.NotNil(t, templateData)

	// assert
	assert.Len(t, templateData.Models, 1)
}

func dumpJSON(v interface{}) {
	j, _ := json.Marshal(v)
	fmt.Print(string(j))
}
