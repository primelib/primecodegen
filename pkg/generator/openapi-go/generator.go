package openapi_go

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

type GoGenerator struct {
	reservedWords  []string
	primitiveTypes []string
	typeToImport   map[string]string
}

func (g *GoGenerator) Id() string {
	return "go"
}

func (g *GoGenerator) Description() string {
	return "Generates Go client code"
}

func (g *GoGenerator) Generate(opts openapigenerator.GenerateOpts) error {
	// check opts
	if opts.Doc == nil {
		return fmt.Errorf("document is required")
	}

	// required options
	if opts.ArtifactId == "" {
		return fmt.Errorf("artifact id is required, please set the --md-artifact-id flag")
	}

	// set packages
	opts.PackageConfig = openapigenerator.CommonPackages{
		Root:       "client",
		Client:     "client",
		Models:     "models",
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
		return fmt.Errorf("failed to build template data: %w", err)
	}

	// generate files
	files, err := openapigenerator.GenerateFiles(fmt.Sprintf("openapi-%s-%s", g.Id(), opts.TemplateId), opts.OutputDir, templateData, template.RenderOpts{
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

func (g *GoGenerator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
}

func (g *GoGenerator) ToClassName(name string) string {
	name = util.ToPascalCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Model"
	}
	return name
}

func (g *GoGenerator) ToFunctionName(name string) string {
	name = util.ToPascalCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Func"
	}
	return name
}

func (g *GoGenerator) ToPropertyName(name string) string {
	name = util.ToPascalCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}
	return name
}

func (g *GoGenerator) ToParameterName(name string) string {
	name = util.ToCamelCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}
	return name
}

func (g *GoGenerator) ToConstantName(name string) string {
	name = util.ToPascalCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}
	return name
}

func (g *GoGenerator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	if schema == nil {
		return openapigenerator.DefaultCodeType, fmt.Errorf("schema is nil")
	}
	isNullable := ptr.ValueOrDefault(schema.Nullable, true) == true

	// multiple types
	if util.CountExcluding(schema.Type, "null") > 1 {
		return openapigenerator.CodeType{Name: "interface{}", IsNullable: isNullable}, nil
	}

	// normal types
	switch {
	case slices.Contains(schema.Type, "string"):
		switch schema.Format {
		case "uri":
			return openapigenerator.NewSimpleCodeType("string", schema), nil
		case "binary", "byte":
			return openapigenerator.NewArrayCodeType(openapigenerator.NewSimpleCodeType("byte", schema), schema), nil
		case "date", "date-time":
			return openapigenerator.NewSimpleCodeType("string", schema), nil
		default:
			return openapigenerator.NewSimpleCodeType("string", schema), nil
		}
	case slices.Contains(schema.Type, "boolean"):
		return openapigenerator.NewSimpleCodeType("bool", schema), nil
	case slices.Contains(schema.Type, "integer"):
		switch schema.Format {
		case "int16":
			return openapigenerator.NewSimpleCodeType("int16", schema), nil
		case "int32":
			return openapigenerator.NewSimpleCodeType("int32", schema), nil
		case "int64":
			return openapigenerator.NewSimpleCodeType("int64", schema), nil
		case "uint16":
			return openapigenerator.NewSimpleCodeType("uint16", schema), nil
		case "uint32":
			return openapigenerator.NewSimpleCodeType("uint32", schema), nil
		case "uint64":
			return openapigenerator.NewSimpleCodeType("uint64", schema), nil
		default:
			return openapigenerator.NewSimpleCodeType("int64", schema), nil
		}
	case slices.Contains(schema.Type, "number"):
		switch schema.Format {
		case "float":
			return openapigenerator.NewSimpleCodeType("float32", schema), nil
		case "double":
			return openapigenerator.NewSimpleCodeType("float64", schema), nil
		default:
			return openapigenerator.NewSimpleCodeType("float64", schema), nil
		}
	case slices.Contains(schema.Type, "array"):
		arrayType, err := g.ToCodeType(schema.Items.A.Schema(), schemaType, true)
		if err != nil {
			return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format), err)
		}
		arrayType = g.PostProcessType(arrayType)

		return openapigenerator.NewArrayCodeType(arrayType, schema), nil
	case slices.Contains(schema.Type, "object"):
		// exception for maps
		if schemaType == openapigenerator.CodeTypeSchemaParent {
			if schema.AdditionalProperties != nil && schema.Properties == nil {
				additionalPropertyType, err := g.ToCodeType(schema.AdditionalProperties.A.Schema(), schemaType, true)
				if err != nil {
					return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled additional properties type. schema: %s, format: %s", schema.Type, schema.Format), err)
				}
				additionalPropertyType = g.PostProcessType(additionalPropertyType)

				return openapigenerator.NewMapCodeType(openapigenerator.NewSimpleCodeType("string", schema), additionalPropertyType, schema), nil
			} else if schema.AdditionalProperties == nil && schema.Properties == nil {
				return openapigenerator.NewSimpleCodeType("interface{}", schema), nil
			}
		}

		// exception for any data
		if schema.Properties == nil && len(schema.OneOf) == 0 && len(schema.AnyOf) == 0 && len(schema.AllOf) == 0 && schema.AdditionalProperties == nil {
			return openapigenerator.NewSimpleCodeType("interface{}", schema), nil
		}

		if schema.Title == "" {
			return openapigenerator.DefaultCodeType, fmt.Errorf("schema does not have a title. schema: %s", schema.Type)
		}
		return openapigenerator.CodeType{Name: g.ToClassName(schema.Title), IsNullable: isNullable, ImportPath: "models"}, nil // TODO: import path
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
		} else {
			return openapigenerator.NewSimpleCodeType("interface{}", schema), nil
		}
	default:
		return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
	}
}

func (g *GoGenerator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
	if codeType.IsPostProcessed {
		return codeType
	}

	// PostProcess TypeArgs
	for i, typeArg := range codeType.TypeArgs {
		codeType.TypeArgs[i] = g.PostProcessType(typeArg)
	}

	// Qualifier
	qualifier := ""
	if codeType.ImportPath != "" {
		parts := strings.Split(codeType.ImportPath, "/")
		qualifier = parts[len(parts)-1] + "."
	}

	// FullyQualifiedName
	switch {
	case codeType.IsArray:
		codeType.Declaration = "[]" + codeType.TypeArgs[0].Declaration
		codeType.QualifiedDeclaration = "[]" + qualifier + codeType.TypeArgs[0].QualifiedDeclaration
		codeType.Type = "[]" + codeType.TypeArgs[0].Type
		codeType.QualifiedType = "[]" + qualifier + codeType.TypeArgs[0].Type
	case codeType.IsMap:
		codeType.Declaration = "map[" + codeType.TypeArgs[0].Declaration + "]" + codeType.TypeArgs[1].Declaration
		codeType.QualifiedDeclaration = "map[" + codeType.TypeArgs[0].QualifiedDeclaration + "]" + qualifier + codeType.TypeArgs[1].QualifiedDeclaration
		codeType.Type = "map[" + codeType.TypeArgs[0].Type + "]" + codeType.TypeArgs[1].Type
		codeType.QualifiedType = "map[" + codeType.TypeArgs[0].Type + "]" + qualifier + codeType.TypeArgs[1].QualifiedType
	default:
		codeType.Declaration = codeType.Name
		codeType.QualifiedDeclaration = qualifier + codeType.Name
		codeType.Type = codeType.Name
		codeType.QualifiedType = qualifier + codeType.Name
	}

	// pointer
	if !codeType.IsMap && !codeType.IsArray && codeType.IsNullable {
		codeType.IsPointer = true
	}
	if codeType.IsPointer {
		codeType.Declaration = "*" + codeType.Declaration
		codeType.QualifiedDeclaration = "*" + codeType.QualifiedDeclaration
	}

	codeType.IsPostProcessed = true
	return codeType
}

func (g *GoGenerator) IsPrimitiveType(input string) bool {
	return slices.Contains(g.primitiveTypes, input)
}

func (g *GoGenerator) TypeToImport(iType openapigenerator.CodeType) string {
	typeName := iType.Name
	if typeName == "" {
		return ""
	}
	typeName = strings.Replace(typeName, "*", "", -1)

	return g.typeToImport[typeName]
}

const gofmtBinary = "gofmt"
const goimportsBinary = "goimports"

func (g *GoGenerator) PostProcessing(outputDir string) error {
	if os.Getenv("PRIMECODEGEN_SKIP_POST_PROCESSING") == "true" {
		slog.Debug("Skipping post processing java files")
		return nil
	}

	// run gofmt
	if openapigenerator.IsBinaryAvailable(gofmtBinary) {
		cmd := exec.Command(gofmtBinary, "-s", "-w", outputDir)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running %s: %v", gofmtBinary, err)
		}
	}

	// run goimports
	if openapigenerator.IsBinaryAvailable(goimportsBinary) {
		cmd := exec.Command(goimportsBinary, "-w", outputDir)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running %s: %v", goimportsBinary, err)
		}
	}

	return nil
}

func NewGenerator() *GoGenerator {
	// references: https://openapi-generator.tech/docs/generators/go
	return &GoGenerator{
		reservedWords: []string{
			"bool",
			"break",
			"byte",
			"case",
			"chan",
			"complex128",
			"complex64",
			"const",
			"continue",
			"default",
			"defer",
			"else",
			"error",
			"fallthrough",
			"float32",
			"float64",
			"for",
			"func",
			"go",
			"goto",
			"if",
			"import",
			"int",
			"int16",
			"int32",
			"int64",
			"int8",
			"interface",
			"map",
			"nil",
			"package",
			"range",
			"return",
			"rune",
			"select",
			"string",
			"struct",
			"switch",
			"type",
			"uint",
			"uint16",
			"uint32",
			"uint64",
			"uint8",
			"uintptr",
			"var",
		},
		primitiveTypes: []string{
			"string",
			"bool",
			"int",
			"int32",
			"int64",
			"float32",
			"float64",
			"byte",
			"rune",
			"time.Time",
		},
		typeToImport: map[string]string{
			"time.Time": "time",
		},
	}
}
