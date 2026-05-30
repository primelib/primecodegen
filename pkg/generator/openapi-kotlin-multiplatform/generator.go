package openapi_kotlin_multiplatform

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	openapi_kotlin "github.com/primelib/primecodegen/pkg/generator/openapi-kotlin"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/primelib/primecodegen/pkg/util"
)

type KotlinMultiplatformGenerator struct {
	baseGenerator *openapi_kotlin.KotlinGenerator
}

func (g *KotlinMultiplatformGenerator) Id() string {
	return "kotlin-multiplatform"
}

func (g *KotlinMultiplatformGenerator) Description() string {
	return "Generates Kotlin multiplatform client code with JsonElement dynamic mappings"
}

func (g *KotlinMultiplatformGenerator) Generate(opts openapigenerator.GenerateOpts) error {
	if opts.Doc == nil {
		return fmt.Errorf("document is required")
	}

	if opts.ArtifactGroupId == "" {
		return fmt.Errorf("artifact id is required, please set the --md-group-id flag")
	}
	if opts.ArtifactId == "" {
		return fmt.Errorf("artifact id is required, please set the --md-artifact-id flag")
	}

	if os.Getenv("PRIMECODEGEN_DEBUG_SPEC") == "true" {
		out, _ := opts.Doc.Model.Render()
		fmt.Print(string(out))
	}

	rootPackagePath := strings.ReplaceAll(opts.ArtifactGroupId+"."+opts.ArtifactId, "-", ".")
	opts.PackageConfig = openapigenerator.CommonPackages{
		Root:       rootPackagePath,
		Client:     rootPackagePath + ".client",
		Models:     rootPackagePath + ".models",
		Responses:  rootPackagePath + ".responses",
		Enums:      rootPackagePath + ".enums",
		Operations: rootPackagePath + ".operations",
		Auth:       rootPackagePath + ".auth",
	}

	templateData, err := g.TemplateData(openapigenerator.TemplateDataOpts{
		Doc:           opts.Doc,
		PackageConfig: opts.PackageConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to build template data in %s: %w", g.Id(), err)
	}

	files, err := openapigenerator.GenerateFiles("openapi-kotlin-"+opts.TemplateId, opts.OutputDir, templateData, templateapi.RenderOpts{
		DryRun:               opts.DryRun,
		Types:                nil,
		IgnoreFiles:          nil,
		IgnoreFileCategories: nil,
		Properties:           map[string]string{},
		TemplateFunctions: texttemplate.FuncMap{
			"toClassName":           g.ToClassName,
			"toFunctionName":        g.ToFunctionName,
			"toPropertyName":        g.ToPropertyName,
			"toParameterName":       g.ToParameterName,
			"isPrimitiveType":       g.IsPrimitiveType,
			"statusCodeToClassName": g.StatusCodeToClassName,
		},
	}, opts)
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	for _, f := range files {
		slog.Debug("Generated file", "file", f.File, "template-file", f.TemplateFile, "state", string(f.State))
	}
	slog.Info(fmt.Sprintf("Generated %d files", len(files)))

	oldFiles := openapigenerator.FilesListedInMetadata(opts.OutputDir)
	for _, f := range oldFiles {
		if _, ok := files[f]; !ok {
			slog.Debug("Removing obsolete file", "file", f)
			if !opts.DryRun {
				err = openapigenerator.RemoveGeneratedFile(opts.OutputDir, f)
				if err != nil {
					return fmt.Errorf("failed to remove generated file: %w", err)
				}
			}
		}
	}

	err = g.PostProcessing(files)
	if err != nil {
		return fmt.Errorf("failed to run post-processing: %w", err)
	}

	err = openapigenerator.WriteMetadata(opts.OutputDir, files)
	if err != nil {
		return errors.Join(openapigenerator.ErrFailedToWriteMetadata, err)
	}

	return nil
}

func (g *KotlinMultiplatformGenerator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	templateData, err := openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
	if err != nil {
		return templateData, err
	}
	return templateData, nil
}

func (g *KotlinMultiplatformGenerator) ToClassName(name string) string {
	return g.baseGenerator.ToClassName(name)
}

func (g *KotlinMultiplatformGenerator) ToFunctionName(name string) string {
	return g.baseGenerator.ToFunctionName(name)
}

func (g *KotlinMultiplatformGenerator) ToPropertyName(name string) string {
	return g.baseGenerator.ToPropertyName(name)
}

func (g *KotlinMultiplatformGenerator) ToParameterName(name string) string {
	return g.baseGenerator.ToParameterName(name)
}

func (g *KotlinMultiplatformGenerator) ToConstantName(name string) string {
	return g.baseGenerator.ToConstantName(name)
}

func (g *KotlinMultiplatformGenerator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	if schema == nil {
		return openapigenerator.DefaultCodeType, fmt.Errorf("schema is nil")
	}

	if util.CountExcluding(schema.Type, "null") > 1 {
		return jsonElementCodeType(), nil
	}

	switch {
	case slices.Contains(schema.Type, "string"):
		switch schema.Format {
		case "uri":
			return openapigenerator.CodeType{Name: "String"}, nil
		case "binary", "byte":
			return openapigenerator.CodeType{Name: "ByteArray"}, nil
		case "date", "date-time":
			return openapigenerator.CodeType{Name: "Instant", ImportPath: "kotlin.time"}, nil
		case "uuid":
			return openapigenerator.CodeType{Name: "String"}, nil
		default:
			return openapigenerator.CodeType{Name: "String"}, nil
		}

	case slices.Contains(schema.Type, "boolean"):
		return openapigenerator.NewSimpleCodeType("Boolean", schema), nil

	case slices.Contains(schema.Type, "integer"):
		switch schema.Format {
		case "int16":
			return openapigenerator.NewSimpleCodeType("Short", schema), nil
		case "int32":
			return openapigenerator.NewSimpleCodeType("Int", schema), nil
		case "int64":
			return openapigenerator.NewSimpleCodeType("Long", schema), nil
		case "uint16", "uint32":
			return openapigenerator.NewSimpleCodeType("Long", schema), nil
		case "uint64":
			return openapigenerator.NewSimpleCodeType("Long", schema), nil
		default:
			return openapigenerator.NewSimpleCodeType("Long", schema), nil
		}

	case slices.Contains(schema.Type, "number"):
		switch schema.Format {
		case "float":
			return openapigenerator.NewSimpleCodeType("Float", schema), nil
		case "double":
			return openapigenerator.NewSimpleCodeType("Double", schema), nil
		default:
			return openapigenerator.NewSimpleCodeType("Double", schema), nil
		}

	case slices.Contains(schema.Type, "array"):
		if schema.Items == nil || schema.Items.A == nil {
			return openapigenerator.DefaultCodeType, fmt.Errorf("array schema missing items definition")
		}
		arrayType, err := g.ToCodeType(schema.Items.A.Schema(), schemaType, true)
		if err != nil {
			return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format), err)
		}
		return openapigenerator.NewListCodeType(arrayType, schema), nil

	case slices.Contains(schema.Type, "object") || schema.Type == nil:
		if schema.PatternProperties != nil {
			var codeTypes []openapigenerator.CodeType
			for pp := schema.PatternProperties.First(); pp != nil; pp = pp.Next() {
				ppSchema := pp.Value().Schema()
				mappedType, err := g.ToCodeType(ppSchema, schemaType, true)
				if err != nil {
					return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled pattern properties type. schema: %s, format: %s", schema.Type, schema.Format), err)
				}
				codeTypes = append(codeTypes, mappedType)
			}

			valueType := jsonElementCodeType()
			if len(codeTypes) > 0 && openapigenerator.HaveSameCodeTypeName(codeTypes) {
				valueType = codeTypes[0]
			}

			return openapigenerator.NewMapCodeType(
				openapigenerator.NewSimpleCodeType("String", schema),
				valueType,
				schema,
			), nil
		} else if schema.AdditionalProperties != nil && schema.AdditionalProperties.IsA() {
			additionalPropertyType, err := g.ToCodeType(schema.AdditionalProperties.A.Schema(), schemaType, true)
			if err != nil {
				return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled additional properties type. schema: %s, format: %s", schema.Type, schema.Format), err)
			}

			return openapigenerator.NewMapCodeType(
				openapigenerator.NewSimpleCodeType("String", schema),
				additionalPropertyType,
				schema,
			), nil
		} else if schema.AdditionalProperties != nil && schema.AdditionalProperties.IsB() && schema.AdditionalProperties.B {
			return openapigenerator.NewMapCodeType(
				openapigenerator.NewSimpleCodeType("String", schema),
				jsonElementCodeType(),
				schema,
			), nil
		} else if schema.AdditionalProperties != nil && schema.AdditionalProperties.IsB() && !schema.AdditionalProperties.B && schema.Properties == nil {
			return jsonElementCodeType(), nil
		} else if schema.AdditionalProperties == nil && schema.Properties == nil {
			return jsonElementCodeType(), nil
		} else {
			if schema.Title == "" {
				return openapigenerator.DefaultCodeType, fmt.Errorf("schema does not have a title. schema: %s", schema.Type)
			}
			return openapigenerator.CodeType{Name: g.ToClassName(schema.Title)}, nil
		}

	case len(schema.Type) == 0 && len(schema.OneOf) > 0:
		codeTypes := make([]openapigenerator.CodeType, 0, len(schema.OneOf))
		for _, oneOfSchema := range schema.OneOf {
			codeType, err := g.ToCodeType(oneOfSchema.Schema(), schemaType, true)
			if err != nil {
				return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled oneOf type. schema: %s, format: %s", schema.Type, schema.Format), err)
			}
			codeTypes = append(codeTypes, codeType)
		}
		if openapigenerator.HaveSameCodeTypeName(codeTypes) {
			return codeTypes[0], nil
		}
		return jsonElementCodeType(), nil

	case len(schema.Type) == 0 && (len(schema.AnyOf) > 0 || len(schema.AllOf) > 0 || schema.Not != nil):
		return jsonElementCodeType(), nil

	default:
		return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
	}
}

func (g *KotlinMultiplatformGenerator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
	return g.baseGenerator.PostProcessType(codeType)
}

func (g *KotlinMultiplatformGenerator) IsPrimitiveType(input string) bool {
	return g.baseGenerator.IsPrimitiveType(input)
}

func (g *KotlinMultiplatformGenerator) TypeToImport(iType openapigenerator.CodeType) string {
	return g.baseGenerator.TypeToImport(iType)
}

func (g *KotlinMultiplatformGenerator) PostProcessing(files map[string]templateapi.RenderedFile) error {
	return g.baseGenerator.PostProcessing(files)
}

func (g *KotlinMultiplatformGenerator) StatusCodeToClassName(code string) string {
	return g.baseGenerator.StatusCodeToClassName(code)
}

func NewGenerator() *KotlinMultiplatformGenerator {
	return &KotlinMultiplatformGenerator{baseGenerator: openapi_kotlin.NewGenerator()}
}

func jsonElementCodeType() openapigenerator.CodeType {
	return openapigenerator.CodeType{Name: "JsonElement", ImportPath: "kotlinx.serialization.json"}
}
