package openapigenerator

import (
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

type DocumentModel struct {
	Name            string
	DisplayName     string
	Description     string
	Tags            map[string]Tag
	Operations      []Operation
	OperationsByTag map[string][]Operation
	Models          []Model
	Enums           []Enum
	Packages        CommonPackages // Packages holds the import paths for output packages
	Auth            Auth
}

type CommonPackages struct {
	Client     string
	Models     string
	Enums      string
	Operations string
	Auth       string
}

type Metadata struct {
	ArtifactGroupId string
	ArtifactId      string
	Name            string
	DisplayName     string
	Description     string
}

type Auth struct {
	Methods []AuthMethod
}

func (a Auth) HasAuth() bool {
	return len(a.Methods) > 0
}

func (a Auth) HasAuthMethod(name string) bool {
	for _, m := range a.Methods {
		if m.Name == name {
			return true
		}
	}
	return false
}

func (a Auth) HasAuthScheme(scheme string) bool {
	for _, m := range a.Methods {
		if m.Scheme == scheme {
			return true
		}
	}
	return false
}

type AuthMethod struct {
	Name   string
	Scheme string
}

type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type Operation struct {
	Name             string          `yaml:"name,omitempty"`
	Path             string          `yaml:"path"`
	Method           string          `yaml:"method"`
	Summary          string          `yaml:"summary,omitempty"`     // Short description
	Description      string          `yaml:"description,omitempty"` // Long description
	Tag              string          `yaml:"tag,omitempty"`
	Tags             []string        `yaml:"tags,omitempty"`
	ReturnType       string          `yaml:"returnType,omitempty"`
	Deprecated       bool            `yaml:"deprecated,omitempty"`
	DeprecatedReason string          `yaml:"deprecatedReason,omitempty"`
	Parameters       []Parameter     `yaml:"parameters,omitempty"`
	PathParameters   []Parameter     `yaml:"pathParameters,omitempty"`
	QueryParameters  []Parameter     `yaml:"queryParameters,omitempty"`
	HeaderParameters []Parameter     `yaml:"headerParameters,omitempty"`
	CookieParameters []Parameter     `yaml:"cookieParameters,omitempty"`
	Imports          []string        `yaml:"imports,omitempty"`
	Documentation    []Documentation `yaml:"documentation,omitempty"`
	Stability        string          `yaml:"stability,omitempty"`
}

func (o Operation) HasParametersWithType(paramType string) bool {
	for _, p := range o.Parameters {
		if p.In == paramType {
			return true
		}
	}

	return false
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
	Name             string     `yaml:"name"`
	Description      string     `yaml:"description,omitempty"`
	Parent           string     `yaml:"parent,omitempty"`
	Properties       []Property `yaml:"properties,omitempty"`
	AnyOf            []Model    `yaml:"anyOf,omitempty"`
	AllOf            []Model    `yaml:"allOf,omitempty"`
	OneOf            []Model    `yaml:"oneOf,omitempty"`
	Imports          []string   `yaml:"imports,omitempty"`
	Deprecated       bool       `yaml:"deprecated,omitempty"`
	DeprecatedReason string     `yaml:"deprecatedReason,omitempty"`
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
