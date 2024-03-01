package openapigenerator

type SupportOnceTemplate struct {
	GoModule string
}

type OperationEachTemplate struct {
	Package   string
	Name      string
	Operation Operation
}

type OperationsOnceTemplate struct {
	Operations []Operation
}

type ModelEachTemplate struct {
	Package string
	Name    string
	Model   Model
}

type ModelsOnceTemplate struct {
	Models []Model
}

type EnumEachTemplate struct {
	Package string
	Name    string
	Enum    Enum
}
