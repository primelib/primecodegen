package util

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// JSONToYAML converts JSON data to YAML data
func JSONToYAML(jsonData []byte) ([]byte, error) {
	var intermediate map[string]interface{}
	if err := json.Unmarshal(jsonData, &intermediate); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	yamlData, err := yaml.Marshal(intermediate)
	if err != nil {
		return nil, fmt.Errorf("error marshaling YAML: %w", err)
	}
	return yamlData, nil
}
