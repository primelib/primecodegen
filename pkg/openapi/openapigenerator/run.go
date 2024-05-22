package openapigenerator

import (
	"fmt"
	"strings"

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

	common := GlobalTemplate{
		GeneratorProperties: renderOpts.Properties,
		Auth:                templateData.Auth,
		Packages:            templateData.Packages,
		Operations:          templateData.Operations,
		Models:              templateData.Models,
		Enums:               templateData.Enums,
	}
	metadata := Metadata{
		ArtifactGroupId: "",
		ArtifactId:      "",
		Name:            strings.TrimSpace(templateData.Name),
		DisplayName:     strings.TrimSpace(templateData.DisplayName),
		Description:     templateData.Description,
	}

	var data []interface{}
	data = append(data, SupportOnceTemplate{
		Metadata: metadata,
		Common:   common,
	})
	data = append(data, APIOnceTemplate{
		Metadata: metadata,
		Common:   common,
		Package:  common.Packages.Client,
	})
	for tag, ops := range templateData.OperationsByTag {
		tagDescription := ""
		if tagData, ok := templateData.Tags[tag]; ok {
			tagDescription = tagData.Description
		}

		data = append(data, APIEachTemplate{
			Metadata:       metadata,
			Common:         common,
			Package:        common.Packages.Client,
			TagName:        tag,
			TagDescription: tagDescription,
			TagOperations:  ops,
		})
	}
	for _, op := range templateData.Operations {
		data = append(data, OperationEachTemplate{
			Metadata:  metadata,
			Common:    common,
			Package:   common.Packages.Operations,
			Name:      op.Name,
			Operation: op,
		})
	}
	for _, model := range templateData.Models {
		data = append(data, ModelEachTemplate{
			Metadata: metadata,
			Common:   common,
			Package:  common.Packages.Models,
			Name:     model.Name,
			Model:    model,
		})
	}
	for _, enum := range templateData.Enums {
		data = append(data, EnumEachTemplate{
			Metadata: metadata,
			Common:   common,
			Package:  common.Packages.Models,
			Name:     enum.Name,
			Enum:     enum,
		})
	}

	// render files
	for _, d := range data {
		var renderedFiles []template.RenderedFile
		var renderErr error

		if _, ok := d.(SupportOnceTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeSupportOnce, d, renderOpts)
		}
		if _, ok := d.(APIOnceTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeAPIOnce, d, renderOpts)
		}
		if _, ok := d.(APIEachTemplate); ok {
			renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeAPIEach, d, renderOpts)
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
