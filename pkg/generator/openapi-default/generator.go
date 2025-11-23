package openapi_default

import (
	"errors"
	"fmt"
	texttemplate "text/template"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/rs/zerolog/log"
)

type DefaultGenerator struct {
}

func (g *DefaultGenerator) Id() string {
	return "default"
}

func (g *DefaultGenerator) Description() string {
	return "Generates Scaffolding files"
}

func (g *DefaultGenerator) Generate(opts openapigenerator.GenerateOpts) error {
	// check opts
	if opts.Doc == nil {
		return fmt.Errorf("document is required")
	}

	// set packages
	opts.PackageConfig = openapigenerator.CommonPackages{
		Root:       "client",
		Client:     "client",
		Models:     "models",
		Responses:  "responses",
		Enums:      "enums",
		Operations: "operations",
		Auth:       "auth",
	}

	// build template data
	templateData, err := g.TemplateData(openapigenerator.TemplateDataOpts{
		Doc:           opts.Doc,
		PackageConfig: opts.PackageConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to build template data in %s: %w", g.Id(), err)
	}

	// generate files
	files, err := openapigenerator.GenerateFiles(fmt.Sprintf("openapi-%s-%s", g.Id(), opts.TemplateId), opts.OutputDir, templateData, templateapi.RenderOpts{
		DryRun:               opts.DryRun,
		Types:                nil,
		IgnoreFiles:          nil,
		IgnoreFileCategories: nil,
		Properties:           map[string]string{},
		TemplateFunctions: texttemplate.FuncMap{
			"toClassName":     g.ToClassName,
			"toFunctionName":  g.ToFunctionName,
			"toPropertyName":  g.ToPropertyName,
			"toParameterName": g.ToParameterName,
			"isPrimitiveType": g.IsPrimitiveType,
		},
	}, opts)
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	for _, f := range files {
		log.Debug().Str("file", f.File).Str("template-file", f.TemplateFile).Str("state", string(f.State)).Msg("Generated file")
	}
	log.Info().Msgf("Generated %d files", len(files))

	// delete old files (oldfiles - files)
	oldFiles := openapigenerator.FilesListedInMetadata(opts.OutputDir)
	for _, f := range oldFiles {
		if _, ok := files[f]; !ok {
			log.Debug().Str("file", f).Msg("Removing obsolete file")
			if !opts.DryRun {
				err = openapigenerator.RemoveGeneratedFile(opts.OutputDir, f)
				if err != nil {
					return fmt.Errorf("failed to remove generated file: %w", err)
				}
			}
		}
	}

	// post-processing (formatting)
	err = g.PostProcessing(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to run post-processing: %w", err)
	}

	// write metadata
	err = openapigenerator.WriteMetadata(opts.OutputDir, files)
	if err != nil {
		return errors.Join(openapigenerator.ErrFailedToWriteMetadata, err)
	}

	return nil
}

func (g *DefaultGenerator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
}

func (g *DefaultGenerator) ToClassName(name string) string {
	return name
}

func (g *DefaultGenerator) ToFunctionName(name string) string {
	return name
}

func (g *DefaultGenerator) ToPropertyName(name string) string {
	return name
}

func (g *DefaultGenerator) ToParameterName(name string) string {
	return name
}

func (g *DefaultGenerator) ToConstantName(name string) string {
	return name
}

func (g *DefaultGenerator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	return openapigenerator.DefaultCodeType, nil
}

func (g *DefaultGenerator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
	return openapigenerator.DefaultCodeType
}

func (g *DefaultGenerator) IsPrimitiveType(input string) bool {
	return false
}

func (g *DefaultGenerator) TypeToImport(iType openapigenerator.CodeType) string {
	return ""
}

func (g *DefaultGenerator) PostProcessing(outputDir string) error {
	return nil
}

func NewGenerator() *DefaultGenerator {
	return &DefaultGenerator{}
}
