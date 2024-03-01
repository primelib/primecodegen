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

	var data []interface{}
	data = append(data, SupportOnceTemplate{
		GoModule: "github.com/primelib/primecodegen", // TODO: configurable
	})
	for _, op := range templateData.Operations {
		data = append(data, OperationEachTemplate{
			Package:   "operations", // TODO: need this from the generator
			Name:      op.OperationId,
			Operation: op,
		})
	}
	for _, model := range templateData.Models {
		data = append(data, ModelEachTemplate{
			Package: "types", // TODO: need this from the generator
			Name:    model.Name,
			Model:   model,
		})
	}
	for _, enum := range templateData.Enums {
		data = append(data, EnumEachTemplate{
			Package: "types", // TODO: need this from the generator
			Name:    enum.Name,
			Enum:    enum,
		})
	}

	// render files
	for _, d := range data {
		var renderedFiles []template.RenderedFile
		var renderErr error

		if _, ok := d.(SupportOnceTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeSupportOnce, d, renderOpts)
		}
		if _, ok := d.(OperationEachTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeOperationEach, d, renderOpts)
		}
		if _, ok := d.(ModelEachTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeModelEach, d, renderOpts)
		}
		if _, ok := d.(EnumEachTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeEnumEach, d, renderOpts)
		}

		if renderErr != nil {
			return nil, fmt.Errorf("failed to render template: %w", renderErr)
		}
		files = append(files, renderedFiles...)
	}

	return files, nil
}
