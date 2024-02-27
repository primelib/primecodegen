package openapigenerator

type DocumentModel struct {
	Operations []Operation
	Models     []Model
}

type Operation struct {
	Path             string      `yaml:"path"`
	Method           string      `yaml:"method"`
	Summary          string      `yaml:"summary,omitempty"`     // Short description
	Description      string      `yaml:"description,omitempty"` // Long description
	Tags             []string    `yaml:"tags,omitempty"`
	OperationId      string      `yaml:"operationId,omitempty"`
	Deprecated       bool        `yaml:"deprecated,omitempty"`
	DeprecatedReason string      `yaml:"deprecatedReason,omitempty"`
	Parameters       []Parameter `yaml:"parameters,omitempty"`
}

type Parameter struct {
	Name             string       `yaml:"name,omitempty"`
	In               string       `yaml:"in,omitempty"`
	Description      string       `yaml:"description,omitempty"`
	Kind             PropertyKind `yaml:"kind,omitempty"`
	Type             string       `yaml:"type,omitempty"`
	Required         bool         `yaml:"required,omitempty"`
	AllowedValues    []string     `yaml:"allowedValues,omitempty"`
	Deprecated       bool         `yaml:"deprecated,omitempty"`
	DeprecatedReason string       `yaml:"deprecatedReason,omitempty"`
}

type Model struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description,omitempty"`
	Parent      string     `yaml:"parent,omitempty"`
	Properties  []Property `yaml:"properties,omitempty"`
}

type Property struct {
	Name          string       `yaml:"name" required:"true"` // Name is the parameter name
	Title         string       `yaml:"title,omitempty"`      // Title is the human-readable name of the parameter
	Kind          PropertyKind `yaml:"kind,omitempty"`       // Kind is the type of the parameter
	Type          string       `yaml:"type,omitempty"`
	Nullable      bool         `yaml:"nullable,omitempty"`
	AllowedValues []string     `yaml:"allowedValues,omitempty"`
	Items         []Property   `yaml:"items,omitempty"`
}

type PropertyKind string

const (
	KindVar  PropertyKind = "var"
	KindEnum PropertyKind = "enum"
)
