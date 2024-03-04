package openapidocument

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

type SchemaMatchFunc func(schema *base.Schema) bool

func AllSchemasMatch(schemas []*base.SchemaProxy, f SchemaMatchFunc) bool {
	for _, schemaProxy := range schemas {
		if !f(schemaProxy.Schema()) {
			return false
		}
	}

	return true
}

func IsEnumSchema(s *base.Schema) bool {
	// 3.0 enum
	if len(s.Enum) > 0 {
		return true
	}

	// 3.1 enum with oneOf and const
	if s.OneOf != nil {
		if AllSchemasMatch(s.OneOf, func(s *base.Schema) bool {
			return s.Const != nil
		}) {
			return true
		}
	}

	return false
}

// MergeSchema merges the properties of the overwrite schema into the base schema (useful to process anyOf, oneOf, allOf)
func MergeSchema(baseSP *base.SchemaProxy, overwriteSP *base.SchemaProxy) error {
	result, err := baseSP.BuildSchema()
	if err != nil {
		return fmt.Errorf("error building a schema: %w", err)
	}
	override, err := overwriteSP.BuildSchema()
	if err != nil {
		return fmt.Errorf("error building schema: %w", err)
	}

	// merge properties
	if override.Type != nil && len(override.Type) > 0 {
		result.Type = override.Type
	}
	if override.Format != "" {
		result.Format = override.Format
	}
	if override.Description != "" {
		result.Description = override.Description
	}
	if override.Items != nil {
		result.Items = override.Items
	}
	if override.Properties != nil {
		if result.Properties == nil {
			result.Properties = override.Properties
		} else {
			for op := override.Properties.Oldest(); op != nil; op = op.Next() {
				result.Properties.Set(op.Key, op.Value)
			}
		}
	}
	/*
		if len(override.Required) > 0 {
			if result.Required == nil {
				result.Required = override.Required
			} else {
				for _, s := range override.Required {
					result.Required = append(result.Required, s)
				}
			}
		}
	*/

	return nil
}

func SimplifyPolymorphism(schemaProxy *base.SchemaProxy) (*base.SchemaProxy, error) {
	if schemaProxy.IsReference() { // skip references
		return nil, nil
	}

	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}

	// merge allOf schemas into base schema
	if schema.AllOf != nil && len(schema.AllOf) > 0 {
		for _, s := range schema.AllOf {
			err = MergeSchema(schemaProxy, s)
			if err != nil {
				return nil, fmt.Errorf("error merging allIn schema into base schema: %w", err)
			}
		}
	}

	return schemaProxy, nil
}
