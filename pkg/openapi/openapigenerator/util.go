package openapigenerator

import (
	"slices"
)

func getBoolValue(ptrToBool *bool, defaultValue bool) bool {
	if ptrToBool != nil {
		return *ptrToBool
	}
	return defaultValue
}

func cleanImports(imports []string) (out []string) {
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
