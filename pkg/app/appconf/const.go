package appconf

// ConfigFileName is the default name of the configuration file
const ConfigFileName = "primelib.yaml"

type GeneratorType string

const (
	GeneratorTypeOpenApiGenerator GeneratorType = "openapi-generator"
	GeneratorTypePrimeCodeGen     GeneratorType = "primecodegen"
	GeneratorTypeSpeakEasy        GeneratorType = "speakeasy"
)
