package openapigenerator

type GlobalTemplate struct {
	GeneratorProperties map[string]string
	Endpoints           Endpoints
	Auth                Auth
	Packages            CommonPackages
	Services            map[string]Service
	Operations          []Operation
	Models              []Model
	Enums               []Enum
}

func (g GlobalTemplate) HasParametersWithType(paramType string) bool {
	for _, o := range g.Operations {
		if o.HasParametersWithType(paramType) {
			return true
		}
	}

	return false
}

type SupportOnceTemplate struct {
	Metadata Metadata       // Metadata for the template, e.g. artifact group, ID, etc.
	Common   GlobalTemplate // Common template data, e.g. API name, project name, etc.
}

type APIOnceTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Package  string
}

type APIEachTemplate struct {
	Metadata       Metadata // Metadata for the template, like artifact group and ID
	Common         GlobalTemplate
	Package        string
	TagName        string // Name of the operation tag
	TagType        string // Type returns the CodeType used for the service
	TagDescription string // Description of the operation tag
	TagOperations  []Operation
}

type OperationEachTemplate struct {
	Metadata  Metadata // Metadata for the template, like artifact group and ID
	Common    GlobalTemplate
	Package   string
	Name      string
	Operation Operation
}

type OperationsOnceTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
}

type ModelEachTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Package  string
	Name     string
	Model    Model
}

type ModelsOnceTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Models   []Model
}

type EnumEachTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Package  string
	Name     string
	Enum     Enum
}
