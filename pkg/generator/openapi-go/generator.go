package openapi_go

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
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

	// TODO: remove - render final document pre-template generation
	// out, _ := opts.Doc.Model.Render()
	// fmt.Print(string(out))

	// build template data
	templateData, err := g.TemplateData(opts.Doc)
	if err != nil {
		return fmt.Errorf("failed to build template data: %w", err)
	}

	// TODO: remove this - render template data passed to render files
	bytes, _ := yaml.Marshal(templateData)
	fmt.Print(string(bytes))

	// generate files
	files, err := openapigenerator.GenerateFiles(fmt.Sprintf("openapi-%s-%s", g.Id(), opts.TemplateId), opts.OutputDir, templateData, template.RenderOpts{
		DryRun:      opts.DryRun,
		Types:       nil,
		IgnoreFiles: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	for _, f := range files {
		log.Debug().Str("file", f.File).Str("template-file", f.TemplateFile).Str("state", string(f.State)).Msg("Generated file")
	}
	log.Info().Msgf("Generated %d files", len(files))

	// post-processing (formatting)
	err = g.PostProcessing(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to run post-processing: %w", err)
	}

	return nil
}

func (g *GoGenerator) TemplateData(doc *libopenapi.DocumentModel[v3.Document]) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(doc, g)
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

func (g *GoGenerator) ToCodeType(schema *base.Schema) (string, error) {
	// multiple types
	if util.CountExcluding(schema.Type, "null") > 1 {
		return "interface{}", nil
	}

	// normal types
	if slices.Contains(schema.Type, "string") && schema.Format == "" {
		return "string", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "uri" {
		return "string", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "binary" {
		return "[]byte", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "byte" {
		return "[]byte", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "date" {
		return "time.Time", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "date-time" {
		return "time.Time", nil
	}
	if slices.Contains(schema.Type, "boolean") {
		return "bool", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "" {
		return "int64", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int32" {
		return "int32", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int64" {
		return "int64", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "" {
		return "float64", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "float" {
		return "float32", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "double" {
		return "float64", nil
	}
	if slices.Contains(schema.Type, "array") {
		arrayType, err := g.ToCodeType(schema.Items.A.Schema())
		if err != nil {
			return "", fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format)
		}
		return "[]" + arrayType, nil
	}
	if slices.Contains(schema.Type, "object") && schema.AdditionalProperties == nil && schema.Properties == nil {
		return "interface{}", nil
	}
	if slices.Contains(schema.Type, "object") && schema.AdditionalProperties != nil {
		keyType := "string"

		additionalProperties, err := g.ToCodeType(schema.AdditionalProperties.A.Schema())
		if err != nil {
			return "", fmt.Errorf("unhandled additional properties type. schema: %s, format: %s: %w", schema.Type, schema.Format, err)
		}

		return "map[" + keyType + "]" + additionalProperties, nil
	}
	if slices.Contains(schema.Type, "object") {
		if schema.Title == "" {
			// TODO: ensure all schemas have a title
			// return "", fmt.Errorf("schema does not have a title. schema: %s", schema.Type)
		}

		return g.ToClassName(schema.Title), nil
	}

	return "", fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
}

func (g *GoGenerator) IsPrimitiveType(input string) bool {
	return slices.Contains(g.primitiveTypes, input)
}

func (g *GoGenerator) TypeToImport(typeName string) string {
	if typeName == "" {
		return ""
	}
	typeName = strings.Replace(typeName, "*", "", -1)

	return g.typeToImport[typeName]
}

func (g *GoGenerator) PostProcessing(outputDir string) error {
	// run gofmt
	cmd := exec.Command("gofmt", "-s", "-w", outputDir)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running gofmt: %v", err)
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
		},
		typeToImport: map[string]string{
			"time.Time": "time",
		},
	}
}
