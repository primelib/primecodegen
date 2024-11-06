package openapidocument

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/rs/zerolog/log"
)

type SchemaMatchFunc func(schema *base.Schema) bool

// AllSchemasMatch returns true if the given function returns true for all input schemas
func AllSchemasMatch(schemas []*base.SchemaProxy, f SchemaMatchFunc) bool {
	for _, schemaProxy := range schemas {
		if !f(schemaProxy.Schema()) {
			return false
		}
	}

	return true
}

// IsEnumSchema returns true if the schema is an enum schema
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

// IsPolypmorphicSchema returns true if the schema is a polymorphic schema (oneOf, anyOf)
func IsPolypmorphicSchema(s *base.Schema) bool {
	if IsEnumSchema(s) {
		return false
	}

	if len(s.OneOf) > 1 {
		return true
	}
	if len(s.AnyOf) > 1 {
		return true
	}

	return false
}

// MergeSchema merges the properties of the overwrite schema into the base schema (useful to process anyOf, oneOf, allOf)
func MergeSchema(baseSP *base.SchemaProxy, overwriteSP *base.SchemaProxy) error {
	result, err := baseSP.BuildSchema()
	if err != nil {
		return fmt.Errorf("error building schema: %w", err)
	}
	override, err := overwriteSP.BuildSchema()
	if err != nil {
		return fmt.Errorf("error building schema: %w", err)
	}

	// merge properties
	if len(override.Type) > 0 {
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
	if override.Nullable != nil {
		result.Nullable = override.Nullable
	}
	if override.Properties != nil {
		for op := override.Properties.Oldest(); op != nil; op = op.Next() {
			bytes, _ := op.Value.Render()
			log.Trace().Str("key", op.Key).Interface("value", string(bytes)).Msg("Properties: ")
		}
		if result.Properties == nil {
			result.Properties = override.Properties
		} else {
			for op := override.Properties.Oldest(); op != nil; op = op.Next() {
				result.Properties.Set(op.Key, op.Value)
			}
		}
	}
	if len(override.Required) > 0 {
		if result.Required == nil {
			result.Required = override.Required
		} else {
			result.Required = append(result.Required, override.Required...)
		}
	}

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
	if len(schema.AllOf) > 0 {
		for _, s := range schema.AllOf {
			err = MergeSchema(schemaProxy, s)
			if err != nil {
				return nil, fmt.Errorf("error merging allIn schema into base schema: %w", err)
			}
		}
	}

	return schemaProxy, nil
}

func InlineAllOf(schemaProxy *base.SchemaProxy) (*base.SchemaProxy, error) {
	if schemaProxy.IsReference() { // skip references
		return nil, nil
	}
	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}
	// merge allOf schemas into base schema
	if len(schema.AllOf) > 0 {
		for _, schemaRef := range schema.AllOf {
			err = MergeSchema(schemaProxy, schemaRef)
			if err != nil {
				return nil, fmt.Errorf("error merging allOf schema into base schema: %w", err)
			}
		}
		// delete allOf (needed by codegeneration)
		schema.AllOf = nil
	}

	return schemaProxy, nil
}
