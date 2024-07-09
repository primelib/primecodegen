package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplateDryRun(t *testing.T) {
	config := Config{
		ID:          "openapi-go-httpclient",
		Description: "dummy template for a go model",
		Files: []File{
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "model.go",
				Type:            TypeModelEach,
			},
		},
	}

	files, err := RenderTemplate(config, "", TypeModelEach, map[string]string{
		"model": "User",
	}, RenderOpts{DryRun: true})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, filepath.Join("models", "model.go"), files[0].File)
	assert.Equal(t, FileDryRun, files[0].State)
}

func TestRenderTemplateFile(t *testing.T) {
	outputDir, err := os.MkdirTemp("", "generator-output")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	config := Config{
		ID:          "openapi-go-httpclient",
		Description: "dummy template for a go model",
		Files: []File{
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "model.go",
				Type:            TypeModelEach,
			},
		},
	}

	files, err := RenderTemplate(config, outputDir, TypeModelEach, map[string]string{
		"model": "User",
	}, RenderOpts{DryRun: false})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	// assert.Equal(t, filepath.Join("models", "model.go"), files[0])
	assert.Equal(t, FileRendered, files[0].State)
}
