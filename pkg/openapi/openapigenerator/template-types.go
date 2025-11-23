package openapigenerator

import "github.com/primelib/primecodegen/pkg/app/appconf"

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
	Metadata Metadata             // Metadata for the template, e.g. artifact group, ID, etc.
	Provider appconf.ProviderConf // Provider contains information about the product or company providing the API
	Common   GlobalTemplate       // Common template data, e.g. API name, project name, etc.
}

type APIOnceTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Package  string
}

type APIEachTemplate struct {
	Metadata Metadata // Metadata for the template, like artifact group and ID
	Common   GlobalTemplate
	Package  string
	Service  Service
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
