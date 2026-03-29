package openapipatch

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/logging"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

var SimplifyAllOfPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "simplify-all-of",
	Description:         "Merges allOf subschemas into the parent schema",
	PatchV3DocumentFunc: SimplifyAllOf,
}

func SimplifyAllOf(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for _, schemaProxy := range openapidocument.CollectSchemas(doc) {
		if err := mergeAllOfSchema(schemaProxy, false); err != nil {
			return err
		}
	}
	return nil
}

var SimplifyInlineAllOfPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "simplify-inline-all-of",
	Description:         "Merges allOf subschemas into the parent schema, but only for inline schemas (i.e. not $ref)",
	PatchV3DocumentFunc: SimplifyInlineAllOf,
}

func SimplifyInlineAllOf(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for _, schemaProxy := range openapidocument.CollectSchemas(doc) {
		if err := mergeAllOfSchema(schemaProxy, true); err != nil {
			return err
		}
	}
	return nil
}

// mergeAllOfSchema is the core logic shared by both the Global and Inline-only patches.
func mergeAllOfSchema(schema *base.Schema, mergeOnlyInlines bool) error {
	if schema == nil || len(schema.AllOf) == 0 {
		return nil
	}

	var preserved []*base.SchemaProxy
	for _, subProxy := range schema.AllOf {
		if mergeOnlyInlines && subProxy.IsReference() {
			preserved = append(preserved, subProxy)
			continue
		}

		subSchema := subProxy.Schema()
		if subSchema != nil {
			// recurse first so we merge "deep to shallow"
			if err := mergeAllOfSchema(subSchema, mergeOnlyInlines); err != nil {
				return err
			}

			// merge schema into parent
			_, err := openapidocument.MergeSchema(schema, subSchema)
			if err != nil {
				return fmt.Errorf("failed to merge allOf: %w", err)
			}
			logging.Trace("Merged allOf sub-schema", "parent", schema.Title, "isRef", subProxy.IsReference())
		}
	}

	if len(schema.Type) == 0 && schema.Properties != nil {
		schema.Type = []string{"object"}
	}

	schema.AllOf = preserved
	return nil
}
