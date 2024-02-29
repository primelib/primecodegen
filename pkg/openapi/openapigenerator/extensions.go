package openapigenerator

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func isMutable(schema *base.Schema) (bool, error) {
	// see https://azure.github.io/autorest/extensions/#x-ms-mutability
	if mxMutability, ok := schema.Extensions.Get("x-ms-mutability"); ok {
		var values []string
		err := mxMutability.Decode(&values)
		if err != nil {
			return false, fmt.Errorf("unable to decode x-ms-mutability: %w", err)
		}

		// possible values: create, read, update

	}

	return false, nil
}

type pcgEnum map[string]struct {
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
}

type msEnum struct {
	Name          string `yaml:"name"`
	ModelAsString bool   `yaml:"modelAsString"`
	Values        []struct {
		Value       string `yaml:"value"`
		Description string `yaml:"description"`
		Name        string `yaml:"name"`
	}
}

// extensionEnumDefinitions processes the x-ms-enum extension
func extensionEnumDefinitions(schema *base.Schema, allowedValues map[string]AllowedValue) (map[string]AllowedValue, error) {
	// primecodegen specific extension
	if yamlNode, ok := schema.Extensions.Get("x-primecodegen-enum"); ok {
		var enumSpec pcgEnum
		err := yamlNode.Decode(&enumSpec)
		if err != nil {
			return allowedValues, fmt.Errorf("unable to decode x-pcg-enum: %w", err)
		}

		for k, v := range enumSpec {
			if _, ok = allowedValues[k]; !ok {
				allowedValues[k] = AllowedValue{
					Value:       k,
					Name:        v.Name,
					Description: v.Description,
				}
			}
		}
	}

	// see https://azure.github.io/autorest/extensions/#x-ms-enum
	if yamlNode, ok := schema.Extensions.Get("x-ms-enum"); ok {
		var enumSpec msEnum
		err := yamlNode.Decode(&enumSpec)
		if err != nil {
			return allowedValues, fmt.Errorf("unable to decode x-ms-enum: %w", err)
		}

		for _, v := range enumSpec.Values {
			if _, ok = allowedValues[v.Value]; !ok {
				allowedValues[v.Value] = AllowedValue{
					Value:       v.Value,
					Name:        v.Name,
					Description: v.Description,
				}
			}
		}
	}

	return allowedValues, nil
}
