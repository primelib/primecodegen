package openapigenerator

import (
	"slices"

	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

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
