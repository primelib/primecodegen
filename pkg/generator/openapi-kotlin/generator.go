package openapi_kotlin

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

type KotlinGenerator struct {
	reservedWords  []string
	reservedRunes  []rune
	primitiveTypes []string
	boxedTypes     map[string]string
	typeToImport   map[string]string
	symbolMappings map[string]string
}

func (g *KotlinGenerator) Id() string {
	return "kotlin"
}

func (g *KotlinGenerator) Description() string {
	return "Generates Kotlin client code"
}

func (g *KotlinGenerator) Generate(opts openapigenerator.GenerateOpts) error {
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
		Root:       rootPackagePath,
		Client:     rootPackagePath + ".client",
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
	err = g.PostProcessing(files)
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

func (g *KotlinGenerator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	templateData, err := openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
	if err != nil {
		return templateData, err
	}
	templateData = openapigenerator.PruneTypeAliases(templateData, g.primitiveTypes)
	return templateData, nil
}

func (g *KotlinGenerator) ToClassName(name string) string {
	name = g.sanitizeName(name)
	if slices.Contains(g.reservedWords, name) {
		name = name + "Model"
	}

	return util.ToPascalCase(name)
}

func (g *KotlinGenerator) ToFunctionName(name string) string {
	name = g.sanitizeName(name)
	if slices.Contains(g.reservedWords, name) {
		name = name + "Func"
	}

	return util.ToCamelCase(name)
}

func (g *KotlinGenerator) ToPropertyName(name string) string {
	if strings.HasPrefix(name, "_") {
		name = "additional" + util.ToUpperCamelCase(name)
	}

	name = g.sanitizeName(name)
	if slices.Contains(g.reservedWords, name) {
		name = name + "Prop"
	}

	return util.ToCamelCase(name)
}

func (g *KotlinGenerator) ToParameterName(name string) string {
	if strings.HasPrefix(name, "_") {
		name = "additional" + util.ToUpperCamelCase(name)
	}

	name = g.sanitizeName(name)
	if slices.Contains(g.reservedWords, name) {
		name = name + "Prop"
	}

	return util.ToCamelCase(name)
}

func (g *KotlinGenerator) ToConstantName(name string) string {
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "p" + name
	}

	name = g.sanitizeName(name)
	if slices.Contains(g.reservedWords, name) {
		name = name + "_CONST"
	}

	return util.ToUpperSnakeCase(name)
}

func (g *KotlinGenerator) sanitizeName(name string) string {
	// special case: starts with a digit
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "p" + name
	}

	// symbol/operator mappings (e.g. "=" → "EQUALS")
	if rep, ok := g.symbolMappings[name]; ok {
		name = rep
	}

	// reserved runes
	name = strings.Map(func(r rune) rune {
		for _, rr := range g.reservedRunes {
			if r == rr {
				return '_'
			}
		}
		return r
	}, name)

	return name
}

func (g *KotlinGenerator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	if schema == nil {
		return openapigenerator.DefaultCodeType, fmt.Errorf("schema is nil")
	}

	// multiple types (e.g., ["string", "null"])
	if util.CountExcluding(schema.Type, "null") > 1 {
		return openapigenerator.CodeType{Name: "Any"}, nil
	}

	switch {
	case slices.Contains(schema.Type, "string"):
		switch schema.Format {
		case "uri":
			return openapigenerator.CodeType{Name: "String"}, nil
		case "binary", "byte":
			// Kotlin multiplatform binary → ByteArray
			return openapigenerator.CodeType{Name: "ByteArray"}, nil
		case "date", "date-time":
			// Kotlinx.datetime.Instant (multiplatform)
			return openapigenerator.CodeType{Name: "Instant", ImportPath: "kotlinx.datetime"}, nil
		case "uuid":
			// Kotlin UUID — may need expect/actual; using String by default for KMP safety
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
			// Kotlin does not have unsigned 32-bit/64-bit types cross-platform yet in KMP serialization
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
		arrayType, err := g.ToCodeType(schema.Items.A.Schema(), schemaType, true)
		if err != nil {
			return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled array type. schema: %s, format: %s", schema.Type, schema.Format), err)
		}
		// Kotlin List<T>
		return openapigenerator.NewListCodeType(arrayType, schema), nil

	case slices.Contains(schema.Type, "object") || schema.Type == nil:
		// handle map-like (additionalProperties) schemas
		if schema.PatternProperties != nil {
			pp := schema.PatternProperties.First()
			ppSchema := pp.Value().Schema()

			additionalPropertyType, err := g.ToCodeType(ppSchema, schemaType, true)
			if err != nil {
				return openapigenerator.DefaultCodeType, errors.Join(fmt.Errorf("unhandled pattern properties type. schema: %s, format: %s", schema.Type, schema.Format), err)
			}

			return openapigenerator.NewMapCodeType(
				openapigenerator.NewSimpleCodeType("String", schema),
				additionalPropertyType,
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
				openapigenerator.NewSimpleCodeType("Any", schema),
				schema,
			), nil
		} else if schema.AdditionalProperties == nil && schema.Properties == nil {
			return openapigenerator.CodeType{Name: "Any"}, nil
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
		return openapigenerator.CodeType{Name: "Any"}, nil

	default:
		return openapigenerator.DefaultCodeType, fmt.Errorf("unhandled type. schema: %s, format: %s", schema.Type, schema.Format)
	}
}

func (g *KotlinGenerator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
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
	if codeType.IsList && len(codeType.TypeArgs) != 1 {
		log.Fatal().Interface("codeType", codeType).Msgf("List type must have exactly one type argument.")
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

func (g *KotlinGenerator) IsPrimitiveType(input string) bool {
	return slices.Contains(g.primitiveTypes, input)
}

func (g *KotlinGenerator) TypeToImport(iType openapigenerator.CodeType) string {
	typeName := iType.Name

	if typeName == "" {
		return ""
	}

	return g.typeToImport[typeName]
}

const googleJavaFormatBinary = "google-java-format"

func (g *KotlinGenerator) PostProcessing(files map[string]templateapi.RenderedFile) error {
	if os.Getenv("PRIMECODEGEN_SKIP_POST_PROCESSING") == "true" {
		slog.Debug("Skipping post processing kotlin files")
		return nil
	}

	// TODO: cli tool for kotlin formatting

	return nil
}

func (g *KotlinGenerator) BoxType(codeType string, box bool) string {
	if !box {
		return codeType
	}

	if boxedType, ok := g.boxedTypes[codeType]; ok {
		return boxedType
	}

	return codeType
}

func NewGenerator() *KotlinGenerator {
	// references: https://openapi-generator.tech/docs/generators/java/
	return &KotlinGenerator{
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
		reservedRunes: []rune{
			'$',  // dollar sign
			'#',  // hash
			'%',  // percent
			'^',  // caret
			'&',  // ampersand
			'*',  // asterisk
			'(',  // open parenthesis
			')',  // close parenthesis
			'+',  // plus
			'=',  // equals
			'/',  // forward slash
			'\\', // backslash
			'|',  // pipe
			'~',  // tilde
			'`',  // backtick
			'!',  // exclamation mark
			'<',  // less than
			'>',  // greater than
			',',  // comma
			':',  // colon
			';',  // semicolon
			' ',  // space
			'?',  // question mark
			'"',  // double quote
			'\'', // single quote
			'{',  // open curly brace
			'}',  // close curly brace
			'[',  // open square bracket
			']',  // close square bracket
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
		symbolMappings: map[string]string{
			"=":  "EQUALS",
			"!=": "NOT_EQUALS",
			">":  "GREATER_THAN",
			"<":  "LESS_THAN",
			">=": "GREATER_OR_EQUALS",
			"<=": "LESS_OR_EQUALS",
			"~":  "TILDE",
			"~=": "MATCHES",
		},
	}
}
