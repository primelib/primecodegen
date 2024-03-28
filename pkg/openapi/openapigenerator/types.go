package openapigenerator

import (
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

type DocumentModel struct {
	Tags            map[string]Tag
	Operations      []Operation
	OperationsByTag map[string][]Operation
	Models          []Model
	Enums           []Enum
}

type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type Operation struct {
	Path             string          `yaml:"path"`
	Method           string          `yaml:"method"`
	Summary          string          `yaml:"summary,omitempty"`     // Short description
	Description      string          `yaml:"description,omitempty"` // Long description
	Tag              string          `yaml:"tag,omitempty"`
	Tags             []string        `yaml:"tags,omitempty"`
	OperationId      string          `yaml:"operationId,omitempty"`
	Deprecated       bool            `yaml:"deprecated,omitempty"`
	DeprecatedReason string          `yaml:"deprecatedReason,omitempty"`
	Parameters       []Parameter     `yaml:"parameters,omitempty"`
	PathParameters   []Parameter     `yaml:"pathParameters,omitempty"`
	QueryParameters  []Parameter     `yaml:"queryParameters,omitempty"`
	HeaderParameters []Parameter     `yaml:"headerParameters,omitempty"`
	CookieParameters []Parameter     `yaml:"cookieParameters,omitempty"`
	Imports          []string        `yaml:"imports,omitempty"`
	Documentation    []Documentation `yaml:"documentation,omitempty"`
}

type Parameter struct {
	Name             string                                  `yaml:"name,omitempty"`
	FieldName        string                                  `yaml:"fieldName,omitempty"` // FieldName is the original name of the parameter
	In               string                                  `yaml:"in,omitempty"`
	Description      string                                  `yaml:"description,omitempty"`
	Type             string                                  `yaml:"type,omitempty"`
	IsPrimitiveType  bool                                    `yaml:"isPrimitiveType,omitempty"`
	Required         bool                                    `yaml:"required,omitempty"`
	AllowedValues    map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	Deprecated       bool                                    `yaml:"deprecated,omitempty"`
	DeprecatedReason string                                  `yaml:"deprecatedReason,omitempty"`
}

type Model struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description,omitempty"`
	Parent      string     `yaml:"parent,omitempty"`
	Properties  []Property `yaml:"properties,omitempty"`
	Imports     []string   `yaml:"imports,omitempty"`
}

type Enum struct {
	Name          string                                  `yaml:"name"`
	Description   string                                  `yaml:"description,omitempty"`
	Parent        string                                  `yaml:"parent,omitempty"`
	ValueType     string                                  `yaml:"ValueType,omitempty"`
	AllowedValues map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	Imports       []string                                `yaml:"imports,omitempty"`
}

type Property struct {
	Name            string                                  `yaml:"name" required:"true"`  // Name is the parameter name
	FieldName       string                                  `yaml:"fieldName,omitempty"`   // FieldName is the original name of the parameter
	Title           string                                  `yaml:"title,omitempty"`       // Title is the human-readable name of the parameter
	Description     string                                  `yaml:"description,omitempty"` // Description is the human-readable description of the parameter
	Type            string                                  `yaml:"type,omitempty"`
	IsPrimitiveType bool                                    `yaml:"isPrimitiveType,omitempty"`
	Nullable        bool                                    `yaml:"nullable,omitempty"`
	AllowedValues   map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	Items           []Property                              `yaml:"items,omitempty"`
}

type Documentation struct {
	Title string `yaml:"title,omitempty"`
	URL   string `yaml:"url,omitempty"`
}
