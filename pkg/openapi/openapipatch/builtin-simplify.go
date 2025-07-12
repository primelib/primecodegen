package openapipatch

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"
)

var MergePolymorphicSchemasPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "simplify-polymorphic-schemas",
	Description: "Merges polymorphic schemas (oneOf, anyOf, allOf) into a single schema",
	Func:        MergePolymorphicSchemas,
}

// MergePolymorphicSchemas merges polymorphic schemas (anyOf, oneOf, allOf) into a single flat schema
func MergePolymorphicSchemas(v3Model *libopenapi.DocumentModel[v3.Document], config string) error {
	// Remember derived schemata (key) to be replaced by their base schemata (value)
	derivedSchemaReplacementMap := make(map[string]string)

	// component schemas
	for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		var err error
		schema.Value, err = openapidocument.SimplifyPolymorphism(schema.Key, schema.Value, v3Model.Model.Components.Schemas, derivedSchemaReplacementMap)
		if err != nil {
			return err
		}
	}

	// TODO: Handle polymorphic responses, request bodies, parameter definitions

	// Delete empty schemas
	deleteEmptySchemas(v3Model, derivedSchemaReplacementMap)

	return nil
}

var MergePolymorphicPropertiesPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "simplify-polymorphic-properties",
	Description: "Merges polymorphic property values (anyOf, oneOf, allOf) into a single flat schema referenced by properties",
	Func:        MergePolymorphicProperties,
}

// MergePolymorphicProperties merges polymorphic property values (anyOf, oneOf, allOf) into a single flat schema referenced by resp. properties
func MergePolymorphicProperties(v3Model *libopenapi.DocumentModel[v3.Document], config string) error {
	// schema properties
	for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {

		if schema.Value.Schema().Properties != nil {
			for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
				/* TODO:
				Probably new method:
					A: Iterate over polymorphic properties
					B: Create a new schema for union of all properties in any-,all- and oneOf referenced schemas
					C: Iterate over all schema references introduced by any-,all- or oneOf and copy all of their properties into new schema
					D: Replace polymorphic relation (any-,all-, oneOf) with schema reference to newly created union schema
				Open: Avoid duplication for identical polymorphic relations in property values
				*/
			}
		}
	}

	return nil
}

var SimplifyPolymorphicBooleansPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "simplify-polymorphic-booleans",
	Description: "Merges polymorphic boolean schemas (oneOf, anyOf, allOf) into a single boolean schema",
	Func:        SimplifyPolymorphicBooleans,
}

// SimplifyPolymorphicBooleans looks for booleans defined as polymorphic types and simplifies them
func SimplifyPolymorphicBooleans(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	walkedDoc := openapidocument.WalkDocument(doc)
	if len(walkedDoc.Schemas) == 0 {
		return nil
	}

	for _, schema := range walkedDoc.Schemas {
		if schema == nil {
			continue
		}

		if len(schema.AnyOf) > 0 {
			var hasBoolean, hasStringEnum bool
			for _, sub := range schema.AnyOf {
				subSchema := sub.Schema()
				if subSchema == nil {
					continue
				}

				if len(subSchema.Type) == 1 && subSchema.Type[0] == "boolean" {
					hasBoolean = true
				}
				if len(subSchema.Type) == 1 && subSchema.Type[0] == "string" && len(subSchema.Enum) > 0 {
					hasStringEnum = true
				}
			}

			// Detect specific pattern and simplify
			if hasBoolean && hasStringEnum {
				schema.Type = []string{"boolean"}
				schema.AnyOf = nil
				schema.Enum = nil
				schema.Format = ""
				log.Trace().Str("schema", schema.Title).Msg("Simplified polymorphic boolean schema to plain boolean")
			}
		}
	}

	return nil
}

var SimplifyAllOfPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "simplify-all-of",
	Description: "Merges allOf subschemas into the parent schema",
	Func:        SimplifyAllOf,
}

// SimplifyAllOf merges allOf subschemas into the parent schema
func SimplifyAllOf(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	walkedDoc := openapidocument.WalkDocument(doc)
	if len(walkedDoc.Schemas) == 0 {
		return nil
	}

	for _, schema := range walkedDoc.Schemas {
		if schema == nil || len(schema.AllOf) == 0 {
			continue
		}

		err := simplifyAllOfRecursive(schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func simplifyAllOfRecursive(schema *base.Schema) error {
	if schema == nil {
		return nil
	}

	if len(schema.AllOf) > 0 {
		for _, sub := range schema.AllOf {
			subSchema := sub.Schema()
			if subSchema == nil {
				continue
			}

			_, err := openapidocument.MergeSchema(schema, subSchema)
			if err != nil {
				return err
			}
		}
		schema.AllOf = nil
	}

	if schema.Properties != nil {
		for p := schema.Properties.Oldest(); p != nil; p = p.Next() {
			if p.Value == nil {
				continue
			}

			err := simplifyAllOfRecursive(p.Value.Schema())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
