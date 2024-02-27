package openapi_go

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"gopkg.in/yaml.v3"
)

type GoGenerator struct {
	reservedWords []string
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

	// TODO: select template for go / limit generated files via options

	// TODO: iterate over all operations and models and generate the files

	return nil
}

func (g *GoGenerator) TemplateData(doc *libopenapi.DocumentModel[v3.Document]) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(doc, g)
}

func (g *GoGenerator) ToClassName(name string) string {
	if slices.Contains(g.reservedWords, name) {
		return name + "Model"
	}
	return name
}

func (g *GoGenerator) ToPropertyName(name string) string {
	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}
	return name
}

func (g *GoGenerator) ToCodeType(schema *base.Schema) (string, error) {
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
		return "int32", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int32" {
		return "int32", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int64" {
		return "int64", nil
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
	if slices.Contains(schema.Type, "object") {
		if schema.Title == "" {
			// TODO: ensure all schemas have a title
			// return "", fmt.Errorf("schema does not have a title. schema: %s", schema.Type)
		}
		return schema.Title, nil
	}

	return "", fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
}

func NewGoGenerator() *GoGenerator {
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
	}
}