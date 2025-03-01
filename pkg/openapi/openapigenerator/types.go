package openapigenerator

import (
	"strings"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

type DocumentModel struct {
	Name            string
	DisplayName     string
	Description     string
	Tags            map[string]Tag
	Endpoints       Endpoints
	Auth            Auth
	Services        map[string]Service
	Operations      []Operation
	OperationsByTag map[string][]Operation
	Models          []Model
	Enums           []Enum
	Packages        CommonPackages // Packages holds the import paths for output packages
}

type CommonPackages struct {
	Root       string
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
	RepositoryUrl   string // RepositoryUrl is the URL to the repository (without protocol or .git suffix)
	LicenseName     string // LicenseName is the name of the license (MIT, Apache-2.0, etc.)
	LicenseUrl      string // LicenseUrl is the URL to the license
}

type Endpoints []Endpoint

type Endpoint struct {
	Type        string
	URL         string
	Description string
}

func (e Endpoints) HasEndpointWithType(value string) bool {
	for _, ep := range e {
		if ep.Type == value {
			return true
		}
	}
	return false
}

func (e Endpoints) DefaultEndpoint() string {
	for _, ep := range e {
		return ep.URL
	}
	return ""
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

// Service represents a named collection of operations
type Service struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type,omitempty"` // Type returns the CodeType used for the service
	Description   string `yaml:"description,omitempty"`
	Operations    []Operation
	Documentation []Documentation `yaml:"documentation,omitempty"`
}

type Operation struct {
	Name                string          `yaml:"name,omitempty"`
	Path                string          `yaml:"path"`
	Method              string          `yaml:"method"`
	Summary             string          `yaml:"summary,omitempty"`     // Short description
	Description         string          `yaml:"description,omitempty"` // Long description
	Tag                 string          `yaml:"tag,omitempty"`
	Tags                []string        `yaml:"tags,omitempty"`
	ReturnType          CodeType        `yaml:"returnType,omitempty"`
	Deprecated          bool            `yaml:"deprecated,omitempty"`
	DeprecatedReason    string          `yaml:"deprecatedReason,omitempty"`
	Parameters          []Parameter     `yaml:"parameters,omitempty"`          // Parameters holds all parameters, including static ones that can not be overridden
	MutableParameters   []Parameter     `yaml:"mutableParameters,omitempty"`   // MutableParameters can be supplied by the user
	ImmutableParameters []Parameter     `yaml:"immutableParameters,omitempty"` // ImmutableParameters can not be overridden by the user
	PathParameters      []Parameter     `yaml:"pathParameters,omitempty"`
	QueryParameters     []Parameter     `yaml:"queryParameters,omitempty"`
	HeaderParameters    []Parameter     `yaml:"headerParameters,omitempty"`
	CookieParameters    []Parameter     `yaml:"cookieParameters,omitempty"`
	BodyParameter       *Parameter      `yaml:"bodyParameter,omitempty"`
	Imports             []string        `yaml:"imports,omitempty"`
	Documentation       []Documentation `yaml:"documentation,omitempty"`
	Stability           string          `yaml:"stability,omitempty"`
}

func (o *Operation) HasParametersWithType(paramType string) bool {
	for _, p := range o.Parameters {
		if p.In == paramType {
			return true
		}
	}

	return false
}

func (o *Operation) AddParameter(parameter Parameter) {
	// add parameter to parameter type list
	o.Parameters = append(o.Parameters, parameter)
	if parameter.StaticValue == "" {
		o.MutableParameters = append(o.MutableParameters, parameter)
	} else {
		o.ImmutableParameters = append(o.ImmutableParameters, parameter)
	}
	switch parameter.In {
	case "path":
		o.PathParameters = append(o.PathParameters, parameter)
	case "query":
		o.QueryParameters = append(o.QueryParameters, parameter)
	case "header":
		o.HeaderParameters = append(o.HeaderParameters, parameter)
	case "cookie":
		o.CookieParameters = append(o.CookieParameters, parameter)
	}

	// replace original FieldName in method path with parameter name
	o.Path = strings.Replace(o.Path, "{"+parameter.FieldName+"}", "{"+parameter.Name+"}", -1)
}

type Parameter struct {
	Name             string                                  `yaml:"name,omitempty"`
	FieldName        string                                  `yaml:"fieldName,omitempty"` // FieldName is the original name of the parameter
	In               string                                  `yaml:"in,omitempty"`
	Description      string                                  `yaml:"description,omitempty"`
	Type             CodeType                                `yaml:"type,omitempty"`
	IsPrimitiveType  bool                                    `yaml:"isPrimitiveType,omitempty"`
	Required         bool                                    `yaml:"required,omitempty"`
	AllowedValues    map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	StaticValue      string                                  `yaml:"staticValue,omitempty"`
	Deprecated       bool                                    `yaml:"deprecated,omitempty"`
	DeprecatedReason string                                  `yaml:"deprecatedReason,omitempty"`
}

type Model struct {
	Name             string     `yaml:"name"`
	Description      string     `yaml:"description,omitempty"`
	Parent           CodeType   `yaml:"parent,omitempty"`
	Properties       []Property `yaml:"properties,omitempty"`
	AnyOf            []Model    `yaml:"anyOf,omitempty"`
	AllOf            []Model    `yaml:"allOf,omitempty"`
	OneOf            []Model    `yaml:"oneOf,omitempty"`
	Imports          []string   `yaml:"imports,omitempty"`
	Deprecated       bool       `yaml:"deprecated,omitempty"`
	DeprecatedReason string     `yaml:"deprecatedReason,omitempty"`
	IsTypeAlias      bool       `yaml:"isTypeAlias,omitempty"`
}

type Enum struct {
	Name             string                                  `yaml:"name"`
	Description      string                                  `yaml:"description,omitempty"`
	Parent           CodeType                                `yaml:"parent,omitempty"`
	ValueType        CodeType                                `yaml:"valueType,omitempty"`
	AllowedValues    map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	Imports          []string                                `yaml:"imports,omitempty"`
	Deprecated       bool                                    `yaml:"deprecated,omitempty"`
	DeprecatedReason string                                  `yaml:"deprecatedReason,omitempty"`
}

type Property struct {
	Name            string                                  `yaml:"name" required:"true"`  // Name is the parameter name
	FieldName       string                                  `yaml:"fieldName,omitempty"`   // FieldName is the original name of the parameter
	Title           string                                  `yaml:"title,omitempty"`       // Title is the human-readable name of the parameter
	Description     string                                  `yaml:"description,omitempty"` // Description is the human-readable description of the parameter
	Type            CodeType                                `yaml:"type,omitempty"`
	IsPrimitiveType bool                                    `yaml:"isPrimitiveType,omitempty"`
	Nullable        bool                                    `yaml:"nullable,omitempty"`
	AllowedValues   map[string]openapidocument.AllowedValue `yaml:"allowedValues,omitempty"`
	Items           []Property                              `yaml:"items,omitempty"`
}

type Documentation struct {
	Title string `yaml:"title,omitempty"`
	URL   string `yaml:"url,omitempty"`
}
