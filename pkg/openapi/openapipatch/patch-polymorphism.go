package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"
)

// SimplifyPolymorphicBooleans looks for booleans defined as polymorphic types and simplifies them
func SimplifyPolymorphicBooleans(doc *libopenapi.DocumentModel[v3.Document]) error {
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

// SimplifyAllOf merges allOf subschemas into the parent schema
func SimplifyAllOf(doc *libopenapi.DocumentModel[v3.Document]) error {
	walkedDoc := openapidocument.WalkDocument(doc)
	if len(walkedDoc.Schemas) == 0 {
		return nil
	}

	for _, schema := range walkedDoc.Schemas {
		if schema == nil || len(schema.AllOf) == 0 {
			continue
		}

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

	return nil
}
