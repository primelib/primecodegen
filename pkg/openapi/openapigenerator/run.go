package openapigenerator

import (
	"fmt"

	"github.com/primelib/primecodegen/pkg/template"
)

func GeneratorById(id string, allGenerators []CodeGenerator) (CodeGenerator, error) {
	for _, g := range allGenerators {
		if g.Id() == id {
			return g, nil
		}
	}

	return nil, fmt.Errorf("generator with id %s not found", id)
}

func GenerateFiles(templateId string, outputDir string, templateData DocumentModel, renderOpts template.RenderOpts) ([]template.RenderedFile, error) {
	var files []template.RenderedFile

	// support files
	suppFiles, err := template.RenderTemplateById(templateId, outputDir, template.ScopeSupport, SupportOnceTemplate{
		GoModule: "github.com/primelib/primecodegen", // TODO: configurable
	}, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	files = append(files, suppFiles...)

	// operations
	for _, op := range templateData.Operations {
		data := OperationEachTemplate{
			Package:   "operations", // TODO: need this from the generator
			Name:      op.OperationId,
			Operation: op,
		}
		opFiles, err := template.RenderTemplateById(templateId, outputDir, template.ScopeOperation, data, renderOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		files = append(files, opFiles...)
	}

	// models
	for _, model := range templateData.Models {
		data := ModelEachTemplate{
			Package: "types", // TODO: need this from the generator
			Name:    model.Name,
			Model:   model,
		}
		modelFiles, err := template.RenderTemplateById(templateId, outputDir, template.ScopeModel, data, renderOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		files = append(files, modelFiles...)
	}

	return files, nil
}
