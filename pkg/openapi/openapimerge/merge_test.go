package openapimerge

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMergeLibOpenAPICmd(t *testing.T) {

	// Arrange
	inputFile1 := createTempFile(t, "input1.yaml", "openapi: 3.0.0\ninfo:\n  title: Test API 1\n  version: 1.0.0\n  description: This API implements A\n")
	inputFile2 := createTempFile(t, "input2.yaml", "openapi: 3.0.0\ninfo:\n  title: Test API 2\n  version: 1.0.0\n  description: This API implements B\n")
	inputFile3 := createTempFile(t, "input3.yaml", "openapi: 3.0.0\ninfo:\n  title:\n  version:\n  description:\n")
	inputEmptySpec := createTempFile(t, "", "openapi: 3.0.1\ninfo:\ntitle:\nversion:\nsummary:\ndescription:\ncontact:\nextensions:\nlicense:\ntermsOfService:\npaths: {}\ncomponents: {}")
	defer os.Remove(inputFile1)
	defer os.Remove(inputFile2)
	defer os.Remove(inputFile3)
	defer os.Remove(inputEmptySpec)
	paths := []string{inputFile1, inputFile2, inputFile3}

	// Act
	v3Model, err := MergeOpenAPISpecs(inputEmptySpec, paths)

	// Assert
	assert.NoError(t, err)
	yamlDate, err := yaml.Marshal(v3Model.Model)
	assert.NoError(t, err)
	outputData := string(yamlDate)
	fmt.Println("Merged Spec: " + outputData)
	assert.NoError(t, err)
	assert.Contains(t, string(outputData), "openapi: 3.0.1")
	assert.Contains(t, string(outputData), "Test API 2")
	assert.Contains(t, string(outputData), "title: Test API 1")
	assert.Contains(t, string(outputData), "TEST API 1 \\n\\nThis API implements A")
	assert.Contains(t, string(outputData), "TEST API 2 \\n\\nThis API implements B")
}

func createTempFile(t *testing.T, name, content string) string {
	file, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return file.Name()
}
