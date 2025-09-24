package openapigenerator

import (
	"slices"

	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

func HaveSameCodeTypeName(codeTypes []CodeType) bool {
	if len(codeTypes) == 0 {
		return false
	}
	firstName := codeTypes[0].Name
	for _, ct := range codeTypes {
		if ct.Name != firstName {
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
