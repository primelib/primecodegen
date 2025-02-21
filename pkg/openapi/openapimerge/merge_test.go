package openapimerge

import (
	"fmt"
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/stretchr/testify/assert"
)

func TestMergeOpenAPI3(t *testing.T) {
	// Arrange
	inputEmptySpec := []byte("openapi: 3.0.1\ninfo:\ntitle:\nversion:\nsummary:\ndescription:\ncontact:\nextensions:\nlicense:\ntermsOfService:\npaths: {}\ncomponents: {}")
	inputFile1 := []byte("openapi: 3.0.0\ninfo:\n  title: Test API 1\n  version: 1.0.0\n  description: This API implements A\n")
	inputFile2 := []byte("openapi: 3.0.0\ninfo:\n  title: Test API 2\n  version: 1.0.0\n  description: This API implements B\n")
	inputFile3 := []byte("openapi: 3.0.0\ninfo:\n  title:\n  version:\n  description:\n")
	specs := [][]byte{inputEmptySpec, inputFile1, inputFile2, inputFile3}

	// Merge
	v3Model, err := MergeOpenAPI3(specs)
	assert.NoError(t, err)

	// Convert
	yamlData, err := openapidocument.RenderV3Document(v3Model)
	assert.NoError(t, err)
	outputData := string(yamlData)

	// Assert
	fmt.Println("Merged Spec: " + outputData)
	assert.NoError(t, err)
	assert.Contains(t, outputData, "openapi: 3.0.1")
	assert.Contains(t, outputData, "Test API 2")
	assert.Contains(t, outputData, "title: Test API 1")
	assert.Contains(t, outputData, "TEST API 1 \\n\\nThis API implements A")
	assert.Contains(t, outputData, "TEST API 2 \\n\\nThis API implements B")
}
