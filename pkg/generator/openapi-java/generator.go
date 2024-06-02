package openapi_java

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

type JavaGenerator struct {
	reservedWords  []string
	primitiveTypes []string
	typeToImport   map[string]string
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

	// required options
	if opts.ArtifactGroupId == "" {
		return fmt.Errorf("artifact id is required, please set the --md-group-id flag")
	}
	if opts.ArtifactId == "" {
		return fmt.Errorf("artifact id is required, please set the --md-artifact-id flag")
	}

	// print final spec
	if os.Getenv("PRIMECODEGEN_DEBUG_SPEC") == "true" {
		out, _ := opts.Doc.Model.Render()
		fmt.Print(string(out))
	}

	// build template data
	templateData, err := g.TemplateData(opts.Doc)
	if err != nil {
		return fmt.Errorf("failed to build template data: %w", err)
	}

	// set packages
	rootPackagePath := strings.ReplaceAll(opts.ArtifactGroupId+"."+opts.ArtifactId, "-", ".")
	templateData.Packages = openapigenerator.CommonPackages{
		Client:     rootPackagePath,
		Models:     rootPackagePath + ".models",
		Enums:      rootPackagePath + ".enums",
		Operations: rootPackagePath + ".operations",
		Auth:       rootPackagePath + ".auth",
	}

	// generate files
	files, err := openapigenerator.GenerateFiles(fmt.Sprintf("openapi-%s-%s", g.Id(), opts.TemplateId), opts.OutputDir, templateData, template.RenderOpts{
		DryRun:               opts.DryRun,
		Types:                nil,
		IgnoreFiles:          nil,
		IgnoreFileCategories: nil,
		Properties:           map[string]string{},
		PostProcess:          g.PostProcessContent,
		TemplateFunctions: texttemplate.FuncMap{
			"toClassName":     g.ToClassName,
			"toFunctionName":  g.ToFunctionName,
			"toPropertyName":  g.ToPropertyName,
			"toParameterName": g.ToParameterName,
		},
	}, opts)
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	for _, f := range files {
		log.Debug().Str("file", f.File).Str("template-file", f.TemplateFile).Str("state", string(f.State)).Msg("Generated file")
	}
	log.Info().Msgf("Generated %d files", len(files))

	// post-processing (formatting)
	err = g.PostProcessing(files)
	if err != nil {
		return fmt.Errorf("failed to run post-processing: %w", err)
	}

	return nil
}

func (g *JavaGenerator) TemplateData(doc *libopenapi.DocumentModel[v3.Document]) (openapigenerator.DocumentModel, error) {
	return openapigenerator.BuildTemplateData(doc, g)
}

func (g *JavaGenerator) ToClassName(name string) string {
	name = util.ToPascalCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Model"
	}
	return name
}

func (g *JavaGenerator) ToFunctionName(name string) string {
	name = util.ToCamelCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Func"
	}
	return name
}

func (g *JavaGenerator) ToPropertyName(name string) string {
	name = util.ToCamelCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}

	return name
}

func (g *JavaGenerator) ToParameterName(name string) string {
	name = util.ToCamelCase(name)

	if slices.Contains(g.reservedWords, name) {
		return name + "Prop"
	}

	return name
}

func (g *JavaGenerator) ToCodeType(schema *base.Schema, required bool) (string, error) {
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
	if slices.Contains(schema.Type, "number") && schema.Format == "" {
		return "double", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "float" {
		return "float", nil
	}
	if slices.Contains(schema.Type, "number") && schema.Format == "double" {
		return "double", nil
	}
	if slices.Contains(schema.Type, "array") {
		arrayType, err := g.ToCodeType(schema.Items.A.Schema(), true)
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
		return g.ToClassName(schema.Title), nil
	}

	return "", fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
}

func (g *JavaGenerator) PostProcessType(codeType string) string {
	// set type to void if empty
	if codeType == "" {
		return "void"
	}

	return codeType
}

func (g *JavaGenerator) IsPrimitiveType(input string) bool {
	return slices.Contains(g.primitiveTypes, input)
}

func (g *JavaGenerator) TypeToImport(typeName string) string {
	if typeName == "" {
		return ""
	}

	return g.typeToImport[typeName]
}

func (g *JavaGenerator) PostProcessContent(name string, content []byte) []byte {
	// clean imports
	if strings.HasSuffix(name, ".java") {
		content = CleanJavaImports(content)
	}

	return content
}

const fmtBinary = "google-java-format"

func (g *JavaGenerator) PostProcessing(files []template.RenderedFile) error {
	_, err := exec.LookPath(fmtBinary)
	if err != nil {
		slog.Warn(fmtBinary + " not found in PATH, skipping formatting")
		return nil
	}

	var formatFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.File, ".java") && f.State == template.FileRendered {
			formatFiles = append(formatFiles, f.File)
		}
	}

	slog.Debug("Post processing java files using "+fmtBinary, "file_len", len(files))
	cmd := exec.Command(fmtBinary, "-r", "--aosp")
	cmd.Args = append(cmd.Args, formatFiles...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s: %v", fmtBinary, err)
	}

	return nil
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
		primitiveTypes: []string{
			"String",
			"boolean",
			"int",
			"long",
			"float",
			"double",
			"byte",
			"char",
		},
		typeToImport: map[string]string{
			"OffsetDateTime": "java.time.OffsetDateTime",
		},
	}
}
