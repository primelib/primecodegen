package openapigenerator

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"log/slog"

	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/primelib/primecodegen/pkg/util"
	"gopkg.in/yaml.v3"
)

func GeneratorById(id string, allGenerators []CodeGenerator) (CodeGenerator, error) {
	for _, g := range allGenerators {
		if g.Id() == id {
			return g, nil
		}
	}

	return nil, fmt.Errorf("generator with id %s not found", id)
}

func GenerateFiles(templateId string, outputDir string, templateData DocumentModel, renderOpts templateapi.RenderOpts, generatorOpts GenerateOpts) (map[string]templateapi.RenderedFile, error) {
	slog.Debug("Generating files", "template-id", templateId, "output-dir", outputDir)
	files := make(map[string]templateapi.RenderedFile)
	var filesMutex sync.Mutex

	// print template data
	if os.Getenv("PRIMECODEGEN_DEBUG_TEMPLATEDATA") == "true" {
		bytes, _ := yaml.Marshal(templateData)
		fmt.Print(string(bytes))
	}

	properties := map[string]string{}
	for key, value := range renderOpts.Properties {
		properties[key] = value
	}
	for key, value := range generatorOpts.TemplateProperties {
		properties[key] = value
	}

	// global template data
	common := GlobalTemplate{
		GeneratorProperties: properties,
		Endpoints:           templateData.Endpoints,
		Auth:                templateData.Auth,
		Packages:            templateData.Packages,
		Services:            templateData.Services,
		Operations:          templateData.Operations,
		Models:              templateData.Models,
		Enums:               templateData.Enums,
	}
	metadata := Metadata{
		ArtifactGroupId:  generatorOpts.ArtifactGroupId,
		ArtifactId:       generatorOpts.ArtifactId,
		Name:             strings.TrimSpace(templateData.Name),
		DisplayName:      strings.TrimSpace(templateData.DisplayName),
		Title:            templateData.Title,
		Description:      templateData.Description,
		APISpecVersion:   templateData.APISpecVersion,
		GeneratorVersion: templateData.GeneratorVersion,
		RepositoryUrl:    generatorOpts.RepositoryUrl,
		LicenseName:      generatorOpts.LicenseName,
		LicenseUrl:       generatorOpts.LicenseUrl,
		GeneratorNames:   generatorOpts.GeneratorNames,
		GeneratorOutputs: generatorOpts.GeneratorOutputs,
	}
	if metadata.ArtifactId == "" {
		metadata.ArtifactId = util.ToSlug(metadata.Name)
	}

	var data []interface{}
	data = append(data, SupportOnceTemplate{
		Metadata: metadata,
		Provider: generatorOpts.Provider,
		Common:   common,
	})
	data = append(data, APIOnceTemplate{
		Metadata: metadata,
		Common:   common,
		Package:  common.Packages.Client,
	})

	for _, service := range templateData.Services {
		data = append(data, APIEachTemplate{
			Metadata: metadata,
			Common:   common,
			Package:  common.Packages.Client,
			Service:  service,
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
			Package:  common.Packages.Enums,
			Name:     enum.Name,
			Enum:     enum,
		})
	}

	// render files
	slog.Debug("rendering template files", "templateId", templateId, "outputDir", outputDir, "files", len(data))
	var waitGroup sync.WaitGroup
	sem := make(chan struct{}, 6)
	errCh := make(chan error, 1)

	for _, d := range data {
		d := d
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var renderedFiles map[string]templateapi.RenderedFile
			var renderErr error

			switch d.(type) {
			case SupportOnceTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeSupportOnce, d, renderOpts)
			case APIOnceTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeAPIOnce, d, renderOpts)
			case APIEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeAPIEach, d, renderOpts)
			case OperationEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeOperationEach, d, renderOpts)
			case ModelEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeModelEach, d, renderOpts)
			case EnumEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, templateapi.TypeEnumEach, d, renderOpts)
			}

			if renderErr != nil {
				select {
				case errCh <- fmt.Errorf("failed to render template: %w", renderErr):
				default:
				}
				return
			}

			filesMutex.Lock()
			for k, v := range renderedFiles {
				files[k] = v
			}
			filesMutex.Unlock()
		}()
	}

	waitGroup.Wait()
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("failed to render template: %w", err)
	default:
	}

	return files, nil
}
