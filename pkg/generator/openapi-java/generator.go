package openapi_java

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type JavaGenerator struct {
	reservedWords []string
}

func (g *JavaGenerator) Id() string {
	return "java"
}

func (g *JavaGenerator) Description() string {
	return "Generates Java client code"
}

func (g *JavaGenerator) Generate(opts openapigenerator.GenerateOpts) error {
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
		Scopes:      nil,
		IgnoreFiles: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	for _, f := range files {
		log.Debug().Str("file", f.File).Str("template-file", f.TemplateFile).Str("state", string(f.State)).Msg("Generated file")
	}
	log.Info().Msgf("Generated %d files", len(files))

	return nil
}

func (g *JavaGenerator) TemplateData(doc *libopenapi.DocumentModel[v3.Document]) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(doc, g)
}

func (g *JavaGenerator) ToClassName(name string) string {
	// uppercase first letter and remove special characters
	name = util.CapitalizeAfterChars(name, []int32{'-', '_'}, true)

	if slices.Contains(g.reservedWords, name) {
		return name + "Model"
	}
	return name
}

func (g *JavaGenerator) ToPropertyName(name string) string {
	// uppercase first letter and remove special characters
	name = util.CapitalizeAfterChars(name, []int32{'-', '_'}, true)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}

	return name
}

func (g *JavaGenerator) ToParameterName(name string) string {
	// uppercase first letter and remove special characters
	name = util.CapitalizeAfterChars(name, []int32{'-', '_'}, false)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}

	return name
}

func (g *JavaGenerator) ToCodeType(schema *base.Schema) (string, error) {
	// multiple types
	if util.CountExcluding(schema.Type, "null") > 1 {
		return "Object", nil
	}

	// normal types
	if slices.Contains(schema.Type, "string") && schema.Format == "" {
		return "String", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "uri" {
		return "String", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "binary" {
		return "byte[]", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "byte" {
		return "byte[]", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "date" {
		return "Timestamp", nil
	}
	if slices.Contains(schema.Type, "string") && schema.Format == "date-time" {
		return "Timestamp", nil
	}
	if slices.Contains(schema.Type, "boolean") {
		return "boolean", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "" {
		return "int", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int32" {
		return "int", nil
	}
	if slices.Contains(schema.Type, "integer") && schema.Format == "int64" {
		return "long", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "float" {
		return "float", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "double" {
		return "double", nil
	}
	if slices.Contains(schema.Type, "array") {
		arrayType, err := g.ToCodeType(schema.Items.A.Schema())
		if err != nil {
			return "", fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format)
		}
		return "List<" + arrayType + ">", nil
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

func NewGenerator() *JavaGenerator {
	// references: https://openapi-generator.tech/docs/generators/go
	return &JavaGenerator{
		reservedWords: []string{
			"abstract",
			"assert",
			"boolean",
			"break",
			"byte",
			"case",
			"catch",
			"char",
			"class",
			"const",
			"continue",
			"default",
			"do",
			"double",
			"else",
			"enum",
			"extends",
			"final",
			"finally",
			"float",
			"for",
			"goto",
			"if",
			"implements",
			"import",
			"instanceof",
			"int",
			"interface",
			"list",
			"long",
			"native",
			"new",
			"null",
			"object",
			"offsetdatetime",
			"package",
			"private",
			"protected",
			"public",
			"return",
			"short",
			"static",
			"strictfp",
			"stringutil",
			"super",
			"switch",
			"synchronized",
			"this",
			"throw",
			"throws",
			"transient",
			"try",
			"void",
			"volatile",
			"while",
		},
	}
}
