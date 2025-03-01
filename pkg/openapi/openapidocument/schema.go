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

	// merge properties of derived schemata referencing a base schema using allOf into the base schema
	if len(schema.AllOf) > 0 {
		for _, schemaRef := range schema.AllOf {
			err = mergeAllOf(schemaRef, schema, schemaName, schemas, schemataMap, "AllOf")
			if err != nil {
				return nil, fmt.Errorf("error merging schemas: %w", err)
			}
		}
		// delete allOf (needed by codegeneration)
		schema.AllOf = nil
	}

	// merge properties of derived schemata referenced using anyOf inside a base-schema into the base-schema
	if len(schema.AnyOf) > 0 {
		for _, schemaRef := range schema.AnyOf {
			err = mergeAnyOfOneOf(schemaRef, schema, schemaName, schemas, schemataMap, "AnyOf")
			if err != nil {
				return nil, fmt.Errorf("error merging schemas: %w", err)
			}
		}
		// delete anyOf (needed by codegeneration)
		schema.AnyOf = nil
	}

	// merge properties of derived schemata referenced using oneOf inside a base-schema into the base-schema
	if len(schema.OneOf) > 0 {
		for _, schemaRef := range schema.OneOf {
			err = mergeAnyOfOneOf(schemaRef, schema, schemaName, schemas, schemataMap, "OneOf")
			if err != nil {
				return nil, fmt.Errorf("error merging schemas: %w", err)
			}
		}
		// delete OneOf (needed by codegeneration)
		schema.OneOf = nil
	}

	return schemaProxy, nil
}

func MergeSchemaProxy(baseSP *base.SchemaProxy, overwriteSP *base.SchemaProxy) (*base.Schema, error) {
	result, err := baseSP.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}
	override, err := overwriteSP.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}

	return MergeSchema(result, override)
}

func MergeSchemaProxySchema(baseSP *base.SchemaProxy, override *base.Schema) (*base.Schema, error) {
	result, err := baseSP.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}

	return MergeSchema(result, override)
}

// MergeSchema merges the properties of the overwrite schema into the base schema (useful to process anyOf, oneOf, allOf)
func MergeSchema(result *base.Schema, override *base.Schema) (*base.Schema, error) {
	// merge schema attributes
	if override.SchemaTypeRef != "" {
		result.SchemaTypeRef = override.SchemaTypeRef
	}
	if override.ExclusiveMaximum != nil {
		result.ExclusiveMaximum = override.ExclusiveMaximum
	}
	if override.ExclusiveMinimum != nil {
		result.ExclusiveMinimum = override.ExclusiveMinimum
	}
	if len(override.Type) > 0 {
		if result.Type == nil {
			result.Type = override.Type
		} else {
			result.Type = append(result.Type, override.Type...)
		}
	}
	// AllOf: Copy props from derived schemas (defining allOf refs) into referenced schemas
	if len(override.AllOf) > 0 {
		for _, subSchemaSP := range override.AllOf {
			err := copyPropertiesIntoBaseSchema(result, subSchemaSP)
			if err != nil {
				return nil, fmt.Errorf("error copying properties from derived schema into base schemas: %w", err)
			}
		}
	}
	// AnyOf, OneOf: Copy props from referenced schemas into "this" schema defining one/anyOf refs
	if len(result.AnyOf) > 0 {
		for _, subSchemaSP := range result.AnyOf {
			err := copyPropertiesIntoBaseSchema(result, subSchemaSP)
			if err != nil {
				return nil, fmt.Errorf("error copying properties from one/anyOf referenced schema into base schema: %w", err)
			}
		}
	}
	// see above AnyOf ...
	if len(result.OneOf) > 0 {
		for _, subSchemaSP := range result.OneOf {
			err := copyPropertiesIntoBaseSchema(result, subSchemaSP)
			if err != nil {
				return nil, fmt.Errorf("error copying properties from one/anyOf referenced schema into base schema: %w", err)
			}
		}
	}
	if len(override.Examples) > 0 {
		if result.Examples == nil {
			result.Examples = override.Examples
		} else {
			result.Examples = append(result.Examples, override.Examples...)
		}
	}
	if len(override.PrefixItems) > 0 {
		if result.PrefixItems == nil {
			result.PrefixItems = override.PrefixItems
		} else {
			result.PrefixItems = append(result.PrefixItems, override.PrefixItems...)
		}
	}
	// 3.1 Specific properties
	if override.Contains != nil {
		result.Contains = override.Contains
	}
	if override.MinContains != nil {
		result.MinContains = override.MinContains
	}
	if override.MaxContains != nil {
		result.MaxContains = override.MaxContains
	}
	if override.If != nil {
		if result.If == nil {
			result.If = override.If
		}
	}
	if override.Else != nil {
		if result.Else == nil {
			result.Else = override.Else
		}
	}
	if override.Then != nil {
		if result.Then == nil {
			result.Then = override.Then
		}
	}
	// TODO: DependentSchemas, PatternProperties
	if override.PropertyNames != nil {
		if result.PropertyNames == nil {
			result.PropertyNames = override.PropertyNames
		}
	}
	if override.UnevaluatedItems != nil {
		if result.UnevaluatedItems == nil {
			result.UnevaluatedItems = override.UnevaluatedItems
		}
	}
	if override.UnevaluatedProperties != nil {
		if result.UnevaluatedProperties == nil {
			result.UnevaluatedProperties = override.UnevaluatedProperties
		}
	}
	if override.Items != nil {
		if result.Items == nil {
			result.Items = override.Items
		}
	}
	if override.Anchor != "" {
		result.Anchor = override.Anchor
	}
	// Compatible with all versions
	if override.Not != nil {
		if result.Not == nil {
			result.Not = override.Not
		}
	}
	if override.Properties != nil {
		for op := override.Properties.Oldest(); op != nil; op = op.Next() {
			bytes, _ := op.Value.Render()
			log.Trace().Str("key", op.Key).Interface("value", string(bytes)).Msg("Properties: ")
		}
		if result.Properties == nil {
			result.Properties = orderedmap.New[string, *base.SchemaProxy]()
		}
		for op := override.Properties.Oldest(); op != nil; op = op.Next() {
			result.Properties.Set(op.Key, op.Value)
		}
	}
	if override.Title != "" {
		if result.Title == "" {
			result.Title = override.Title
		}
	}
	if override.MultipleOf != nil {
		if result.MultipleOf == nil {
			result.MultipleOf = override.MultipleOf
		}
	}
	if override.Maximum != nil {
		if result.Maximum == nil {
			result.Maximum = override.Maximum
		}
	}
	if override.Minimum != nil {
		if result.Minimum == nil {
			result.Minimum = override.Minimum
		}
	}
	if override.MaxLength != nil {
		if result.MaxLength == nil {
			result.MaxLength = override.MaxLength
		}
	}
	if override.MinLength != nil {
		if result.MinLength == nil {
			result.MinLength = override.MinLength
		}
	}
	if override.Pattern != "" {
		if result.Pattern == "" {
			result.Pattern = override.Pattern
		}
	}
	if override.Format != "" {
		if result.Format == "" {
			result.Format = override.Format
		}
	}
	if override.MaxItems != nil {
		if result.MaxItems == nil {
			result.MaxItems = override.MaxItems
		}
	}
	if override.MinItems != nil {
		if result.MinItems == nil {
			result.MinItems = override.MinItems
		}
	}
	if override.UniqueItems != nil {
		if result.UniqueItems == nil {
			result.UniqueItems = override.UniqueItems
		}
	}
	if override.MaxProperties != nil {
		if result.MaxProperties == nil {
			result.MaxProperties = override.MaxProperties
		}
	}
	if override.MinProperties != nil {
		if result.MinProperties == nil {
			result.MinProperties = override.MinProperties
		}
	}
	if len(override.Required) > 0 {
		if result.Required == nil {
			result.Required = override.Required
		} else {
			result.Required = append(result.Required, override.Required...)
		}
	}
	if len(override.Enum) > 0 {
		if result.Enum == nil {
			result.Enum = override.Enum
		} else {
			result.Enum = append(result.Enum, override.Enum...)
		}
	}
	if override.AdditionalProperties != nil {
		if result.AdditionalProperties == nil {
			result.AdditionalProperties = override.AdditionalProperties
		}
	}
	if override.Description != "" {
		if result.Description == "" {
			result.Description = override.Description
		} else {
			result.Description = result.Description + "\n" + override.Description
		}
	}
	if override.Default != nil {
		if result.Default == nil {
			result.Default = override.Default
		}
	}
	if override.Const != nil {
		if result.Const == nil {
			result.Const = override.Const
		}
	}
	if override.Nullable != nil {
		if result.Nullable == nil {
			result.Nullable = override.Nullable
		}
	}
	if override.ReadOnly != nil {
		if result.ReadOnly == nil {
			result.ReadOnly = override.ReadOnly
		}
	}
	if override.WriteOnly != nil {
		if result.WriteOnly == nil {
			result.WriteOnly = override.WriteOnly
		}
	}
	if override.XML != nil {
		if result.XML == nil {
			result.XML = override.XML
		}
	}
	if override.ExternalDocs != nil {
		if result.ExternalDocs == nil {
			result.ExternalDocs = override.ExternalDocs
		}
	}
	if override.Example != nil {
		if result.Example == nil {
			result.Example = override.Example
		}
	}
	if override.Deprecated != nil {
		if result.Deprecated == nil {
			result.Deprecated = override.Deprecated
		}
	}
	// TODO: Extensions
	// Skip: low, ParentProxy

	return result, nil
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
		if AllSchemasMatch(s.OneOf, func(s *base.Schema) bool {
			return s.Const != nil
		}) {
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
