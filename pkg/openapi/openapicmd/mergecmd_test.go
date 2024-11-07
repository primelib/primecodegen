package openapicmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeLibOpenAPICmd(t *testing.T) {
	cmd := MergeLibOpenAPICmd()

	// Create temporary files for input and output
	inputFile1 := createTempFile(t, "input1.yaml", "openapi: 3.0.0\ninfo:\n  title: Test API 1\n  version: 1.0.0\n  description: This API implements A\n")
	inputFile2 := createTempFile(t, "input2.yaml", "openapi: 3.0.0\ninfo:\n  title: Test API 2\n  version: 1.0.0\n  description: This API implements B\n")
	inputFile3 := createTempFile(t, "input3.yaml", "openapi: 3.0.0\ninfo:\n  title:\n  version:\n  description:\n")
	outputFile := filepath.Join(os.TempDir(), "output.yaml")
	defer os.Remove(inputFile1)
	defer os.Remove(inputFile2)
	defer os.Remove(inputFile3)
	defer os.Remove(outputFile)

	// Set flags
	cmd.SetArgs([]string{
		"--input", inputFile1,
		"--input", inputFile2,
		"--empty", inputFile3,
		"--output", outputFile,
	})

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute command
	err := cmd.Execute()
	assert.NoError(t, err)

	// Check output file
	outputData, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	printFileToStdout(t, outputFile)
	assert.Contains(t, string(outputData), "openapi: 3.0.0")
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

func printFileToStdout(t *testing.T, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	fmt.Printf("Content of %s:\n%s\n", filePath, string(data))
}
