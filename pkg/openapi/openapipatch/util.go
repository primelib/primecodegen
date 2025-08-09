package openapipatch

import (
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/util"
)

func isObjectSchema(schema *base.Schema) bool {
	if schema == nil {
		return false
	}
	return slices.Contains(schema.Type, "object") || (len(schema.Type) == 0 && schema.Properties != nil)
}

func isPrimitive(schema *base.Schema) bool {
	if schema == nil {
		return false
	}

	types := schema.Type

	// Handle union types like ["string", "null"]
	if util.CountExcluding(types, "null") > 1 {
		return false
	}

	switch {
	case slices.Contains(types, "string"),
		slices.Contains(types, "boolean"),
		slices.Contains(types, "integer"),
		slices.Contains(types, "number"):
		return true

	case slices.Contains(types, "array"):
		if schema.Items == nil {
			return false
		}
		return isPrimitive(schema.Items.A.Schema())

	case slices.Contains(types, "object"):
		// Object with no defined properties or just additionalProperties = true
		if schema.Properties == nil && schema.AdditionalProperties == nil {
			return true
		}

		// Map with primitive values
		if schema.Properties == nil && schema.AdditionalProperties != nil {
			return isPrimitive(schema.AdditionalProperties.A.Schema())
		}

		return false

	case len(types) == 0 && len(schema.OneOf) > 0:
		return false
	}

	return false
}
