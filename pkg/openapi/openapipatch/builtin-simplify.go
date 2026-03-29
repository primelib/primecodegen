package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/logging"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

var MergePolymorphicSchemasPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "simplify-polymorphic-schemas",
	Description:         "Merges polymorphic schemas (oneOf, anyOf, allOf) into a single schema",
	PatchV3DocumentFunc: MergePolymorphicSchemas,
}

// MergePolymorphicSchemas merges polymorphic schemas (anyOf, oneOf, allOf) into a single flat schema
func MergePolymorphicSchemas(v3Model *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for _, schemaProxy := range openapidocument.CollectSchemas(v3Model) {
		// simplify allOf
		if err := mergeAllOfSchema(schemaProxy, false); err != nil {
			return err
		}
		// simplify anyOf
		if err := mergeAnyOfSchema(schemaProxy); err != nil {
			return err
		}
	}

	// component schemas
	derivedSchemaReplacementMap := make(map[string]string)
	if v3Model != nil && v3Model.Model.Components != nil {
		for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
			var err error
			schema.Value, err = openapidocument.SimplifyPolymorphism(schema.Key, schema.Value, v3Model.Model.Components.Schemas, derivedSchemaReplacementMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var MergePolymorphicPropertiesPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "simplify-polymorphic-properties",
	Description:         "Merges polymorphic property values (anyOf, oneOf, allOf) into a single flat schema referenced by properties",
	PatchV3DocumentFunc: MergePolymorphicProperties,
}

// MergePolymorphicProperties merges polymorphic property values (anyOf, oneOf, allOf) into a single flat schema referenced by resp. properties
func MergePolymorphicProperties(v3Model *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
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
	Type:                "builtin",
	ID:                  "simplify-polymorphic-booleans",
	Description:         "Merges polymorphic boolean schemas (oneOf, anyOf, allOf) into a single boolean schema",
	PatchV3DocumentFunc: SimplifyPolymorphicBooleans,
}

// SimplifyPolymorphicBooleans looks for booleans defined as polymorphic types and simplifies them
func SimplifyPolymorphicBooleans(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
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
				logging.Trace("Simplified polymorphic boolean schema to plain boolean", "schema", schema.Title)
			}
		}
	}

	return nil
}
