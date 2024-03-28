package openapigenerator

type SupportOnceTemplate struct {
	ProjectName string // Name of the project
	GoModule    string
}

type APIOnceTemplate struct {
	ProjectName string // Name of the project
	Package     string
	Operations  []Operation
}

type APIEachTemplate struct {
	ProjectName    string // Name of the project
	Package        string
	TagName        string // Name of the operation tag
	TagDescription string // Description of the operation tag
	Operations     []Operation
}

type OperationEachTemplate struct {
	ProjectName string // Name of the project
	Package     string
	Name        string
	Operation   Operation
}

type OperationsOnceTemplate struct {
	ProjectName string // Name of the project
	Operations  []Operation
}

type ModelEachTemplate struct {
	ProjectName string // Name of the project
	Package     string
	Name        string
	Model       Model
}

type ModelsOnceTemplate struct {
	ProjectName string // Name of the project
	Models      []Model
}

type EnumEachTemplate struct {
	ProjectName string // Name of the project
	Package     string
	Name        string
	Enum        Enum
}
