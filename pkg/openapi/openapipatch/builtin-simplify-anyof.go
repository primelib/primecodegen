package openapipatch

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/logging"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

var SimplifyAnyOfPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "simplify-any-of",
	Description:         "Merges anyOf subschemas into the parent schema to support non-polymorphic DTOs",
	PatchV3DocumentFunc: SimplifyAnyOf,
}

func SimplifyAnyOf(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for _, schemaProxy := range openapidocument.CollectSchemaProxies(doc) {
		schema := schemaProxy.Schema()
		if err := mergeAnyOfSchema(schema); err != nil {
			return err
		}
	}
	return nil
}

// mergeAnyOfSchema flattens anyOf into properties.
func mergeAnyOfSchema(schema *base.Schema) error {
	if schema == nil || len(schema.AnyOf) == 0 {
		return nil
	}

	for _, subProxy := range schema.AnyOf {
		subSchema := subProxy.Schema()
		if subSchema != nil {
			// recursion
			if err := mergeAnyOfSchema(subSchema); err != nil {
				return err
			}

			// merge allOf first
			if len(subSchema.AllOf) > 0 {
				err := mergeAllOfSchema(subSchema, false)
				if err != nil {
					return err
				}
			}

			//  merge schemas
			_, err := openapidocument.MergeSchema(schema, subSchema)
			if err != nil {
				return fmt.Errorf("failed to merge anyOf: %w", err)
			}
			logging.Trace("Merged anyOf sub-schema into parent", "parent", schema.Title)
		}
	}

	if len(schema.Type) == 0 && schema.Properties != nil {
		schema.Type = []string{"object"}
	}

	schema.AnyOf = nil

	return nil
}
