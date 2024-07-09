package openapi_java

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/openapi/openapiutil"
	"github.com/primelib/primecodegen/pkg/template"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

type JavaGenerator struct {
	reservedWords  []string
	primitiveTypes []string
	boxedTypes     map[string]string
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

	// set packages
	rootPackagePath := strings.ReplaceAll(opts.ArtifactGroupId+"."+opts.ArtifactId, "-", ".")
	opts.PackageConfig = openapigenerator.CommonPackages{
		Client:     rootPackagePath,
		Models:     rootPackagePath + ".models",
		Enums:      rootPackagePath + ".enums",
		Operations: rootPackagePath + ".operations",
		Auth:       rootPackagePath + ".auth",
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
		PostProcess:          g.PostProcessContent,
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

	// post-processing (formatting)
	err = g.PostProcessing(files)
	if err != nil {
		return fmt.Errorf("failed to run post-processing: %w", err)
	}

	return nil
}

func (g *JavaGenerator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	templateData, err := openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
	if err != nil {
		return templateData, err
	}
	templateData = openapigenerator.PruneTypeAliases(templateData, g.primitiveTypes)
	return templateData, nil
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

func (g *JavaGenerator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	// multiple types
	if util.CountExcluding(schema.Type, "null") > 1 {
		return openapigenerator.CodeType{Name: "Object"}, nil
	}

	// nullable
	isNullable := openapiutil.IsSchemaNullable(schema)

	// normal types
	switch {
	case slices.Contains(schema.Type, "string") && schema.Format == "":
		return openapigenerator.CodeType{Name: "String"}, nil
	case slices.Contains(schema.Type, "string") && schema.Format == "uri":
		return openapigenerator.CodeType{Name: "String"}, nil
	case slices.Contains(schema.Type, "string") && schema.Format == "binary":
		return openapigenerator.CodeType{TypeArgs: []openapigenerator.CodeType{openapigenerator.NewSimpleCodeType(g.BoxType("byte", isNullable), schema)}, IsArray: true}, nil
	case slices.Contains(schema.Type, "string") && schema.Format == "byte":
		return openapigenerator.CodeType{TypeArgs: []openapigenerator.CodeType{openapigenerator.NewSimpleCodeType(g.BoxType("byte", isNullable), schema)}, IsArray: true}, nil
	case slices.Contains(schema.Type, "string") && schema.Format == "date":
		return openapigenerator.CodeType{Name: "Instant", ImportPath: "java.time"}, nil
	case slices.Contains(schema.Type, "string") && schema.Format == "date-time":
		return openapigenerator.CodeType{Name: "Instant", ImportPath: "java.time"}, nil
	case slices.Contains(schema.Type, "boolean"):
		return openapigenerator.NewSimpleCodeType(g.BoxType("boolean", isNullable), schema), nil
	case slices.Contains(schema.Type, "integer") && schema.Format == "":
		return openapigenerator.NewSimpleCodeType(g.BoxType("int", isNullable), schema), nil
	case slices.Contains(schema.Type, "integer") && schema.Format == "int32":
		return openapigenerator.NewSimpleCodeType(g.BoxType("int", isNullable), schema), nil
	case slices.Contains(schema.Type, "integer") && schema.Format == "int64":
		return openapigenerator.NewSimpleCodeType(g.BoxType("long", isNullable), schema), nil
	case slices.Contains(schema.Type, "number") && schema.Format == "":
		return openapigenerator.NewSimpleCodeType(g.BoxType("double", isNullable), schema), nil
	case slices.Contains(schema.Type, "number") && schema.Format == "float":
		return openapigenerator.NewSimpleCodeType(g.BoxType("float", isNullable), schema), nil
	case slices.Contains(schema.Type, "number") && schema.Format == "double":
		return openapigenerator.NewSimpleCodeType(g.BoxType("double", isNullable), schema), nil
	case slices.Contains(schema.Type, "array"):
		arrayType, err := g.ToCodeType(schema.Items.A.Schema(), schemaType, true)
		if err != nil {
			return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format)
		}

		isArrayTypeNullable := openapiutil.IsSchemaNullable(schema.Items.A.Schema())
		if isArrayTypeNullable {
			return openapigenerator.NewListCodeType(arrayType, schema), nil
		}
		return openapigenerator.NewArrayCodeType(arrayType, schema), nil
	case slices.Contains(schema.Type, "object"):
		if schema.AdditionalProperties != nil {
			additionalPropertyType, err := g.ToCodeType(schema.AdditionalProperties.A.Schema(), schemaType, true)
			if err != nil {
				return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled additional properties type. schema: %s, format: %s: %w", schema.Type, schema.Format, err)
			}

			return openapigenerator.NewMapCodeType(openapigenerator.NewSimpleCodeType("String", schema), additionalPropertyType, schema), nil
		} else if schema.AdditionalProperties == nil && schema.Properties == nil {
			return openapigenerator.CodeType{Name: "Object"}, nil
		} else {
			if schema.Title == "" {
				return openapigenerator.DefaultCodeType, fmt.Errorf("schema does not have a title. schema: %s", schema.Type)
			}
			return openapigenerator.CodeType{Name: g.ToClassName(schema.Title)}, nil // TODO: import path
		}
	default:
		return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
	}
}

func (g *JavaGenerator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
	if codeType.IsPostProcessed {
		return codeType
	}

	// VoidType
	if codeType.IsVoid {
		codeType.Declaration = "void"
		codeType.QualifiedDeclaration = "void"
		codeType.Type = "void"
		codeType.QualifiedType = "void"
		codeType.IsPostProcessed = true
		return codeType
	}

	// PostProcess TypeArgs
	for i, typeArg := range codeType.TypeArgs {
		codeType.TypeArgs[i] = g.PostProcessType(typeArg)
	}

	// Validate
	if codeType.IsArray && len(codeType.TypeArgs) != 1 {
		log.Fatal().Interface("codeType", codeType).Msgf("Array type must have exactly one type argument.")
		return codeType
	}
	if codeType.IsMap && len(codeType.TypeArgs) != 2 {
		log.Fatal().Interface("codeType", codeType).Msgf("Map type must have exactly two type arguments.")
		return codeType
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
		codeType.Declaration = codeType.TypeArgs[0].Declaration + "[]"
		codeType.QualifiedDeclaration = qualifier + codeType.TypeArgs[0].QualifiedDeclaration + "[]"
		codeType.Type = codeType.TypeArgs[0].Type + "[]"
		codeType.QualifiedType = qualifier + codeType.TypeArgs[0].Type + "[]"
	case codeType.IsList:
		codeType.Declaration = "List<" + codeType.TypeArgs[0].Declaration + ">"
		codeType.QualifiedDeclaration = "List<" + qualifier + codeType.TypeArgs[0].QualifiedDeclaration + ">"
		codeType.Type = "List<" + codeType.TypeArgs[0].Type + ">"
		codeType.QualifiedType = "List<" + qualifier + codeType.TypeArgs[0].Type + ">"
	case codeType.IsMap:
		codeType.Declaration = "Map<" + codeType.TypeArgs[0].Declaration + ", " + codeType.TypeArgs[1].Declaration + ">"
		codeType.QualifiedDeclaration = "Map<" + codeType.TypeArgs[0].QualifiedDeclaration + ", " + qualifier + codeType.TypeArgs[1].QualifiedDeclaration + ">"
		codeType.Type = "Map<" + codeType.TypeArgs[0].Type + ", " + codeType.TypeArgs[1].Type + ">"
		codeType.QualifiedType = "Map<" + codeType.TypeArgs[0].Type + ", " + qualifier + codeType.TypeArgs[1].QualifiedType + ">"
	default:
		codeType.Declaration = g.BoxType(codeType.Name, codeType.IsNullable)
		codeType.QualifiedDeclaration = qualifier + g.BoxType(codeType.Name, codeType.IsNullable)
		codeType.Type = g.BoxType(codeType.Name, codeType.IsNullable)
		codeType.QualifiedType = qualifier + g.BoxType(codeType.Name, codeType.IsNullable)
	}

	codeType.IsPostProcessed = true
	return codeType
}

func (g *JavaGenerator) IsPrimitiveType(input string) bool {
	return slices.Contains(g.primitiveTypes, input)
}

func (g *JavaGenerator) TypeToImport(iType openapigenerator.CodeType) string {
	typeName := iType.Name

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

func (g *JavaGenerator) BoxType(codeType string, box bool) string {
	if !box {
		return codeType
	}

	if boxedType, ok := g.boxedTypes[codeType]; ok {
		return boxedType
	}

	return codeType
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
			"boolean",
			"int",
			"long",
			"short",
			"float",
			"double",
			"byte",
			"char",
		},
		boxedTypes: map[string]string{
			"boolean": "Boolean",
			"int":     "Integer",
			"long":    "Long",
			"short":   "Short",
			"float":   "Float",
			"double":  "Double",
			"byte":    "Byte",
			"char":    "Character",
		},
		typeToImport: map[string]string{
			"OffsetDateTime": "java.time.OffsetDateTime",
		},
	}
}
