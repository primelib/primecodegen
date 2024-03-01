package openapidocument

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

type AllowedValue struct {
	Value       string `yaml:"value"`
	Description string `yaml:"description,omitempty"`
	Name        string `yaml:"name,omitempty"`
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

// EnumToAllowedValues converts an enum schema to a list of allowed values
func EnumToAllowedValues(s *base.Schema) (map[string]AllowedValue, error) {
	allowedValues := make(map[string]AllowedValue)
	// 3.1 enum with enum
	if s.Enum != nil {
		for _, e := range s.Enum {
			allowedValues[e.Value] = AllowedValue{Value: e.Value, Name: e.Value}
		}
	}

	// 3.1 enum with oneOf
	if s.OneOf != nil {
		for _, o := range s.OneOf {
			os, err := o.BuildSchema()
			if err != nil {
				return allowedValues, fmt.Errorf("error building oneOf schema: %w", err)
			}

			if os.Const != nil {
				allowedValues[os.Const.Value] = AllowedValue{Value: os.Const.Value, Name: os.Title, Description: os.Description}
			}
		}
	}

	// primecodegen specific extension
	if yamlNode, ok := s.Extensions.Get("x-primecodegen-enum"); ok {
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
	if yamlNode, ok := s.Extensions.Get("x-ms-enum"); ok {
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
