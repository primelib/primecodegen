package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/stretchr/testify/assert"
)

func TestRenderTemplateDryRun(t *testing.T) {
	config := templateapi.Config{
		ID:          "openapi-go-httpclient",
		Description: "dummy template for a go model",
		Files: []templateapi.File{
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "model.go",
				Type:            templateapi.TypeModelEach,
			},
		},
	}

	files, err := RenderTemplate(config, "", templateapi.TypeModelEach, map[string]string{
		"model": "User",
	}, templateapi.RenderOpts{DryRun: true})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	fileKey := filepath.Join("models", "model.go")
	assert.Equal(t, fileKey, files[fileKey].File)
	assert.Equal(t, templateapi.FileDryRun, files[fileKey].State)
}

func TestRenderTemplateFile(t *testing.T) {
	outputDir, err := os.MkdirTemp("", "generator-output")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	config := templateapi.Config{
		ID:          "openapi-go-httpclient",
		Description: "dummy template for a go model",
		Files: []templateapi.File{
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "model.go",
				Type:            templateapi.TypeModelEach,
			},
		},
	}

	files, err := RenderTemplate(config, outputDir, templateapi.TypeModelEach, map[string]string{
		"model": "User",
	}, templateapi.RenderOpts{DryRun: false})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	//assert.Equal(t, filepath.Join("models", "model.go"), files["test"])
	//fileKey := filepath.Join("models", "model.go")
	//assert.Equal(t, FileRendered, files[fileKey].State)
}
