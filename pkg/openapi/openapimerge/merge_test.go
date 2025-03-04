package openapimerge

import (
	"fmt"
	"testing"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/stretchr/testify/assert"
)

func TestMergeOpenAPI3CrossVersion(t *testing.T) {
	// Arrange
	spec1 := []byte("openapi: 3.0.0\ninfo:\n  title: Test API 1\n  version: 1.0.0\n  description: This API implements A\n")
	spec2 := []byte("openapi: 3.1.0\ninfo:\n  title: Test API 2\n  version: 1.0.0\n  description: This API implements B\n")
	specs := [][]byte{spec1, spec2}

	// Merge
	_, err := MergeOpenAPI3(specs)
	assert.ErrorIs(t, err, ErrOpenAPICrossVersionMergeUnsupported)
}

func TestMergeOpenAPI3Info(t *testing.T) {
	// Arrange
	spec1 := []byte("openapi: 3.0.0\ninfo:\n  title: Test API 1\n  version: 1.0.0\n  description: This API implements A\n")
	spec2 := []byte("openapi: 3.0.0\ninfo:\n  title: Test API 2\n  version: 1.0.0\n  description: This API implements B\n")
	spec3 := []byte("openapi: 3.0.0\ninfo:\n  title:\n  version:\n  description:\n")
	specs := [][]byte{spec1, spec2, spec3}

	// Merge
	v3Model, err := MergeOpenAPI3(specs)
	assert.NoError(t, err)

	// Convert
	yamlData, err := loader.InterfaceToYaml(v3Model.Model)
	assert.NoError(t, err)
	outputData := string(yamlData)

	// Assert
	fmt.Println("Merged Spec: " + outputData)
	assert.NoError(t, err)

	expectedYaml := `openapi: 3.0.0
info:
  title: Test API 1, Test API 2
  description: |-
    # TEST API 1

    This API implements A

    # TEST API 2

    This API implements B
  version: |-
    (Test API 1) 1.0.0
    (Test API 2) 1.0.0
`
	assert.Equal(t, expectedYaml, outputData, "The merged spec YAML did not match the expected output")
}
