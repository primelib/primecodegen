package openapidocument

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/rs/zerolog/log"
)

type SchemaMatchFunc func(schema *base.Schema) bool

func SimplifyPolymorphism(schemaName string, schemaProxy *base.SchemaProxy, schemas *orderedmap.Map[string, *base.SchemaProxy], schemataMap map[string]string) (*base.SchemaProxy, error) {
	schemataMap[schemaName] = ""
	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}

	if !IsPolymorphicSchema(schema) && schemaProxy.IsReference() {
		return nil, nil
	}

	// merge main schema
	if err = simplifyPolymorphismForSchema(schema, schemaName, schemas, schemataMap); err != nil {
		return nil, err
	}

	// merge property schemas
	if schema.Properties != nil {
		for op := schema.Properties.Oldest(); op != nil; op = op.Next() {
			propertySchema := op.Value.Schema()

			if propertySchema.Properties == nil && IsPolymorphicSchema(propertySchema) {
				log.Warn().Msg("polymorphic property detected, need to merge into base schema")
				if err := simplifyPolymorphismForSchema(propertySchema, op.Key, schemas, schemataMap); err != nil {
					return nil, err
				}
			}
		}
	}

	return schemaProxy, nil
}

func simplifyPolymorphismForSchema(schema *base.Schema, schemaName string, schemas *orderedmap.Map[string, *base.SchemaProxy], schemataMap map[string]string) error {
	if len(schema.AllOf) > 0 {
		for _, schemaRef := range schema.AllOf {
			if err := mergeAllOf(schemaRef, schema, schemaName, schemas, schemataMap, "AllOf"); err != nil {
				return fmt.Errorf("error merging schemas: %w", err)
			}
		}
		schema.AllOf = nil
	}

	if len(schema.AnyOf) > 0 {
		for _, schemaRef := range schema.AnyOf {
			if err := mergeAnyOfOneOf(schemaRef, schema, schemaName, schemas, schemataMap, "AnyOf"); err != nil {
				return fmt.Errorf("error merging schemas: %w", err)
			}
		}
		schema.AnyOf = nil
	}

	if len(schema.OneOf) > 0 {
		for _, schemaRef := range schema.OneOf {
			if err := mergeAnyOfOneOf(schemaRef, schema, schemaName, schemas, schemataMap, "OneOf"); err != nil {
				return fmt.Errorf("error merging schemas: %w", err)
			}
		}
		schema.OneOf = nil
	}

	return nil
}

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
		if AllSchemasMatch(s.OneOf, func(s *base.Schema) bool { return s.Const != nil }) {
			return true
		}
	}

	return false
}

// IsPolymorphicSchema returns true if the schema is a polymorphic schema (allOf, oneOf, anyOf)
func IsPolymorphicSchema(s *base.Schema) bool {
	if IsEnumSchema(s) {
		return false
	}

	if len(s.AllOf) > 1 {
		return true
	}
	if len(s.OneOf) > 1 {
		return true
	}
	if len(s.AnyOf) > 1 {
		return true
	}

	return false
}

func IsEmptySchema(schema *base.Schema) bool {
	if schema == nil {
		return true
	}

	return schema.Properties == nil &&
		schema.Type == nil &&
		schema.Items == nil &&
		schema.AdditionalProperties == nil &&
		schema.Enum == nil &&
		schema.AllOf == nil &&
		schema.AnyOf == nil &&
		schema.OneOf == nil
}

func mergeAllOf(schemaRef *base.SchemaProxy, schema *base.Schema, derivedSchemaName string, schemas *orderedmap.Map[string, *base.SchemaProxy], schemataMap map[string]string, polymorphicRel string) error {
	reference := schemaRef.GetReference()
	if reference != "" {
		baseSchemaName, _ := getSchemaNameFromLocalReference(reference)
		baseSP, present := schemas.Get(baseSchemaName)

		if !present {
			log.Fatal().Str("schema", baseSchemaName).Msg("base schema is missing in model")
		} else {
			log.Debug().Str("schema", polymorphicRel).Str("into base schema ref", reference).Msg("merging derived")
			mergedBaseSchema, err := MergeSchemaProxySchema(baseSP, schema)
			if err != nil {
				return fmt.Errorf("error merging %s schema into base schema: %w", polymorphicRel, err)
			}
			// update model
			renderedUpdatedBaseSchema, _ := mergedBaseSchema.Render()
			log.Trace().Str("schema", baseSchemaName).Str("rendered", string(renderedUpdatedBaseSchema)).Msg("Updated base")
			schemas.Set(baseSchemaName, base.CreateSchemaProxy(mergedBaseSchema))
			schemataMap[derivedSchemaName] = baseSchemaName
		}
	}

	return nil
}

func mergeAnyOfOneOf(schemaRef *base.SchemaProxy, baseSchema *base.Schema, baseSchemaName string, schemas *orderedmap.Map[string, *base.SchemaProxy], schemataMap map[string]string, polymorphicRel string) error {
	reference := schemaRef.GetReference()
	if reference != "" {
		composedSchemaName, _ := getSchemaNameFromLocalReference(reference)
		composedSP, present := schemas.Get(composedSchemaName)
		composedSchema, err := composedSP.BuildSchema()
		if err != nil {
			return fmt.Errorf("error building schema: %w", err)
		}
		if !present {
			log.Fatal().Str("schema", composedSchemaName).Msg("base schema is missing in model")
		} else {
			log.Debug().Str("schema", polymorphicRel).Str("into base schema ref", baseSchemaName).Msg("merging derived")
			mergedBaseSchema, err := MergeSchema(baseSchema, composedSchema)
			if err != nil {
				return fmt.Errorf("error merging %s schema into base schema: %w", polymorphicRel, err)
			}
			// update model
			renderedUpdatedBaseSchema, _ := mergedBaseSchema.Render()
			log.Trace().Str("schema", baseSchemaName).Str("rendered", string(renderedUpdatedBaseSchema)).Msg("Updated base")
			schemas.Set(baseSchemaName, base.CreateSchemaProxy(mergedBaseSchema))
			schemataMap[composedSchemaName] = baseSchemaName
		}
	}

	return nil
}

func copyPropertiesIntoBaseSchema(result *base.Schema, subSchemaSP *base.SchemaProxy) error {
	if subSchemaSP.IsReference() {
		subSchemaRef := subSchemaSP.GetReference()
		log.Debug().Str("ref", subSchemaRef).Msg("sub-schema reference, i.e. base schema:")
		return nil
	}

	subSchema, err := subSchemaSP.BuildSchema()
	if err != nil {
		return fmt.Errorf("error building schema: %w", err)
	}

	if subSchema.Properties != nil {
		for op := subSchema.Properties.Oldest(); op != nil; op = op.Next() {
			bytes, _ := op.Value.Render()
			log.Trace().Str("key", op.Key).Interface("value", string(bytes)).Msg("Extending Properties: ")
		}

		if result.Properties == nil {
			result.Properties = subSchema.Properties
		} else {
			for op := subSchema.Properties.Oldest(); op != nil; op = op.Next() {
				// Do not overwrite props already existent in base class
				if _, exists := result.Properties.Get(op.Key); exists {
					log.Trace().Str("key", op.Key).Msg("Property (key) already exists in base classe - skip copy:")
				} else {
					result.Properties.Set(op.Key, op.Value)
				}
			}
		}
	}

	return nil
}

func getSchemaNameFromLocalReference(ref string) (string, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return "", fmt.Errorf("invalid reference format")
	}

	return strings.TrimPrefix(ref, prefix), nil
}
