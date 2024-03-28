package openapigenerator

import (
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
	TemplateData(doc *libopenapi.DocumentModel[v3.Document]) (DocumentModel, error)

	// ToClassName converts a name to a language-specific class name
	ToClassName(name string) string

	// ToFunctionName converts a name to a language-specific function name
	ToFunctionName(name string) string

	// ToPropertyName converts a name to a language-specific property name
	ToPropertyName(name string) string

	// ToParameterName converts a name to a language-specific parameter name
	ToParameterName(name string) string

	// ToCodeType converts a schema to a language-specific type
	ToCodeType(schema *base.Schema, required bool) (string, error)

	// IsPrimitiveType checks if a type is a primitive type
	IsPrimitiveType(input string) bool

	// TypeToImport returns the import path for a given type
	TypeToImport(typeName string) string
}

type GenerateOpts struct {
	DryRun     bool
	Doc        *libopenapi.DocumentModel[v3.Document]
	OutputDir  string
	TemplateId string
}

type SchemaDefinition struct {
	Type   string
	Format string
}
