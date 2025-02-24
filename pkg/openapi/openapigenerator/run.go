package openapigenerator

import (
	"fmt"
	"os"
	"path"
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

func GenerateFiles(templateId string, outputDir string, templateData DocumentModel, renderOpts template.RenderOpts, generatorOpts GenerateOpts) ([]template.RenderedFile, error) {
	log.Debug().Str("template-id", templateId).Str("output-dir", outputDir).Msg("Generating files")
	var files []template.RenderedFile
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
			Metadata:       metadata,
			Common:         common,
			Package:        common.Packages.Client,
			TagName:        service.Name,
			TagDescription: service.Description,
			TagOperations:  service.Operations,
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
	group := parallelizer.NewGroup(parallelizer.WithPoolSize(32))
	defer group.Close()

	for _, d := range data {
		group.Add(func() error {
			var renderedFiles []template.RenderedFile
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
			files = append(files, renderedFiles...)
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

func RemoveFilesListedInMetadata(outputDir string) error {
	writtenFiles := path.Join(outputDir, ".openapi-generator", "FILES")
	log.Debug().Str("output-dir", outputDir).Str("lookup-file", writtenFiles).Msg("Clearing generated files")

	// open the file for reading
	file, err := os.Open(writtenFiles)
	if err != nil {
		return nil // no metadata file found, no files to remove
	}
	defer file.Close()

	// read each file name from the file
	var files []string
	scanner := yaml.NewDecoder(file)
	for scanner.Decode(&files) == nil {
		for _, f := range files {
			absFile := path.Join(outputDir, f)

			if fileInfo, err := os.Stat(absFile); err == nil && fileInfo.Mode().IsRegular() {
				remErr := os.Remove(absFile)
				if remErr != nil {
					return fmt.Errorf("failed to remove file: %w", remErr)
				}
			}
		}
	}

	return nil
}

// WriteMetadata generates metadata about the generated files for the output directory
func WriteMetadata(outputDir string, files []template.RenderedFile) error {
	writtenFiles := path.Join(outputDir, ".openapi-generator", "FILES")

	// ensure output directory exists
	err := os.MkdirAll(path.Dir(writtenFiles), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// open the file for writing
	file, err := os.Create(writtenFiles)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// write each file name to the file
	for _, f := range files {
		if f.State == template.FileRendered {
			relativeFile := strings.TrimPrefix(strings.TrimPrefix(f.File, outputDir), "/")
			_, err := file.WriteString(relativeFile + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}

	return nil
}
