package openapigenerator

import (
	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// CodeGenerator is the interface that all code generators must implement
type CodeGenerator interface {
	// Id returns a unique id for this generator
	Id() string

	// Description returns a human-readable description of the generator
	Description() string

	// Generate generates code
	Generate(opts GenerateOpts) error

	// TemplateData processes a openapi document and returns a data model to be used in the templates
	TemplateData(opts TemplateDataOpts) (DocumentModel, error)

	// ToClassName converts a name to a language-specific class name
	ToClassName(name string) string

	// ToFunctionName converts a name to a language-specific function name
	ToFunctionName(name string) string

	// ToPropertyName converts a name to a language-specific property name
	ToPropertyName(name string) string

	// ToParameterName converts a name to a language-specific parameter name
	ToParameterName(name string) string

	// ToCodeType converts a schema to a language-specific type
	ToCodeType(schema *base.Schema, schemaType CodeTypeSchemaType, required bool) (CodeType, error)

	// PostProcessType is used for post-processing a type (e.g. void type if the type is empty)
	PostProcessType(codeType CodeType) CodeType

	// IsPrimitiveType checks if a type is a primitive type
	IsPrimitiveType(input string) bool

	// TypeToImport returns the import path for a given type
	TypeToImport(typeName CodeType) string
}

type CodeTypeSchemaType string

const (
	CodeTypeSchemaParameter CodeTypeSchemaType = "parameter"
	CodeTypeSchemaProperty  CodeTypeSchemaType = "property"
	CodeTypeSchemaArray     CodeTypeSchemaType = "array"
	CodeTypeSchemaResponse  CodeTypeSchemaType = "response"
	CodeTypeSchemaParent    CodeTypeSchemaType = "parent"
)

type GenerateOpts struct {
	DryRun          bool
	Doc             *libopenapi.DocumentModel[v3.Document]
	OutputDir       string
	TemplateId      string
	PackageConfig   CommonPackages
	ArtifactGroupId string
	ArtifactId      string
	RepositoryUrl   string
	LicenseName     string
	LicenseUrl      string
}

type TemplateDataOpts struct {
	Doc           *libopenapi.DocumentModel[v3.Document]
	PackageConfig CommonPackages
}

type SchemaDefinition struct {
	Type   string
	Format string
}

var DefaultCodeType = CodeType{}

type CodeType struct {
	Name                 string // Name of the type
	Declaration          string // Declaration of the type
	QualifiedDeclaration string // Qualified declaration of the type (with package name, pointer, optional, etc.)
	Type                 string // Type of the type
	QualifiedType        string // Qualified type of the type (with package name)
	TypeArgs             []CodeType
	IsArray              bool
	IsList               bool
	IsMap                bool
	IsNullable           bool
	IsPointer            bool
	IsPostProcessed      bool
	ImportPath           string
}

func (ct CodeType) String() string {
	return ct.Name
}

func NewSimpleCodeType(name string, schema *base.Schema) CodeType {
	isNullable := ptr.ValueOrDefault(schema.Nullable, true) == true

	return CodeType{
		Name:       name,
		IsNullable: isNullable,
	}
}

func NewArrayCodeType(itemType CodeType, schema *base.Schema) CodeType {
	isNullable := ptr.ValueOrDefault(schema.Nullable, true) == true

	return CodeType{
		TypeArgs:   []CodeType{itemType},
		IsArray:    true,
		IsNullable: isNullable,
	}
}

func NewMapCodeType(keyType CodeType, valueType CodeType, schema *base.Schema) CodeType {
	isNullable := ptr.ValueOrDefault(schema.Nullable, true) == true

	return CodeType{
		TypeArgs:   []CodeType{keyType, valueType},
		IsNullable: isNullable,
		IsMap:      true,
	}
}
