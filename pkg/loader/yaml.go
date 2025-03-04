package loader

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YamlNodeFromFile will load and parse a YAML or JSON file from the given path.
func YamlNodeFromFile(path string) (*yaml.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema from path %q: %w", path, err)
	}

	var yn yaml.Node
	err = yaml.NewDecoder(file).Decode(&yn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema at path %q: %w", path, err)
	}

	return &yn, nil
}

// YamlNodeFromString will load and parse a YAML or JSON file from the given input.
func YamlNodeFromString(input []byte) (*yaml.Node, error) {
	var yn yaml.Node
	err := yaml.Unmarshal(input, &yn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema from input: %w", err)
	}

	return &yn, nil
}

func InterfaceToYaml(payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to render data: %w", err)
	}

	return buf.Bytes(), nil
}
