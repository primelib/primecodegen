package openapigenerator

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/parallelizer"
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

func GenerateFiles(templateId string, outputDir string, templateData DocumentModel, renderOpts template.RenderOpts, generatorOpts GenerateOpts) (map[string]template.RenderedFile, error) {
	log.Debug().Str("template-id", templateId).Str("output-dir", outputDir).Msg("Generating files")
	files := make(map[string]template.RenderedFile)
	var filesMutex sync.Mutex

	// print template data
	if os.Getenv("PRIMECODEGEN_DEBUG_TEMPLATEDATA") == "true" {
		bytes, _ := yaml.Marshal(templateData)
		fmt.Print(string(bytes))
	}

	// global template data
	common := GlobalTemplate{
		GeneratorProperties: renderOpts.Properties,
		Endpoints:           templateData.Endpoints,
		Auth:                templateData.Auth,
		Packages:            templateData.Packages,
		Services:            templateData.Services,
		Operations:          templateData.Operations,
		Models:              templateData.Models,
		Enums:               templateData.Enums,
	}
	metadata := Metadata{
		ArtifactGroupId: generatorOpts.ArtifactGroupId,
		ArtifactId:      generatorOpts.ArtifactId,
		Name:            strings.TrimSpace(templateData.Name),
		DisplayName:     strings.TrimSpace(templateData.DisplayName),
		Description:     templateData.Description,
		RepositoryUrl:   generatorOpts.RepositoryUrl,
		LicenseName:     generatorOpts.LicenseName,
		LicenseUrl:      generatorOpts.LicenseUrl,
	}
	if metadata.ArtifactId == "" {
		metadata.ArtifactId = util.ToSlug(metadata.Name)
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
	log.Debug().Str("templateId", templateId).Str("outputDir", outputDir).Int("files", len(data)).Msg("rendering template files")
	group := parallelizer.NewGroup(parallelizer.WithPoolSize(6))
	defer group.Close()

	for _, d := range data {
		group.Add(func() error {
			var renderedFiles map[string]template.RenderedFile
			var renderErr error

			switch d.(type) {
			case SupportOnceTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeSupportOnce, d, renderOpts)
			case APIOnceTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeAPIOnce, d, renderOpts)
			case APIEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeAPIEach, d, renderOpts)
			case OperationEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeOperationEach, d, renderOpts)
			case ModelEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeModelEach, d, renderOpts)
			case EnumEachTemplate:
				renderedFiles, renderErr = template.RenderTemplateById(templateId, outputDir, template.TypeEnumEach, d, renderOpts)
			}

			if renderErr != nil {
				return fmt.Errorf("failed to render template: %w", renderErr)
			}

			filesMutex.Lock()
			for k, v := range renderedFiles {
				files[k] = v
			}
			filesMutex.Unlock()

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return files, nil
}
