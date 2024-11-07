package openapicmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapiconvert"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOpenAPIConvertCmd(t *testing.T) {
	// arrange
	mockClient := new(openapiconvert.MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"openapi": "3.0.0"}`)),
	}
	mockClient.On("Do", mock.Anything).Return(mockResponse, nil)

	inputFile := createTempFile(t, "input.yaml", `{"swagger": "2.0"}`)
	defer os.Remove(inputFile)
	outputDir, err := os.MkdirTemp("", "output")
	assert.NoError(t, err)
	defer os.RemoveAll(outputDir)

	cmd := OpenAPIConvertCmd(mockClient)
	cmd.SetArgs([]string{
		"--input", inputFile,
		"--output-dir", outputDir,
		"--converter-url", "http://mock-converter-url",
		"--format-in", "swagger20",
		"--format-out", "openapi30",
	})

	// act
	err = cmd.Execute()
	assert.NoError(t, err)

	// assert
	outputFile := filepath.Join(outputDir, "input.yaml")
	outputData, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(outputData), "openapi: 3.0.0")

	mockClient.AssertExpectations(t)
}
