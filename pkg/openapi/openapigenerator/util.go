package openapigenerator

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// HaveSameCodeTypeName checks if all code types are structurally the same type (name, container kind and type arguments)
func HaveSameCodeTypeName(codeTypes []CodeType) bool {
	if len(codeTypes) == 0 {
		return false
	}
	for _, ct := range codeTypes {
		if !isSameCodeType(ct, codeTypes[0]) {
			return false
		}
	}
	return true
}

func isSameCodeType(a CodeType, b CodeType) bool {
	if a.Name != b.Name || a.IsArray != b.IsArray || a.IsList != b.IsList || a.IsMap != b.IsMap || len(a.TypeArgs) != len(b.TypeArgs) {
		return false
	}
	for i := range a.TypeArgs {
		if !isSameCodeType(a.TypeArgs[i], b.TypeArgs[i]) {
			return false
		}
	}
	return true
}

func getBoolValue(ptrToBool *bool, defaultValue bool) bool {
	if ptrToBool != nil {
		return *ptrToBool
	}
	return defaultValue
}

func uniqueSortImports(imports []string) (out []string) {
	// unique imports
	visited := make(map[string]bool)
	for _, imp := range imports {
		if _, ok := visited[imp]; !ok {
			visited[imp] = true

			if imp != "" {
				out = append(out, imp)
			}
		}
	}

	// sort imports
	slices.Sort(out)

	return out
}

func getOrDefault(extensions *orderedmap.Map[string, *yaml.Node], key string, defaultValue string) string {
	if extensions == nil {
		return defaultValue
	}

	if node, ok := extensions.Get(key); ok {
		return node.Value
	}

	return defaultValue
}

func delimiterFromStyle(style string) (string, error) {
	switch style {
	case "pipeDelimited":
		return "|", nil
	case "spaceDelimited":
		return " ", nil
	case "form", "":
		return ",", nil
	default:
		return ",", fmt.Errorf("unsupported style: %s", style)
	}
}
