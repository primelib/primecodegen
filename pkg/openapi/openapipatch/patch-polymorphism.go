package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"
)

// SimplifyPolymorphicBooleans looks for booleans defined as polymorphic types and simplifies them
func SimplifyPolymorphicBooleans(doc *libopenapi.DocumentModel[v3.Document]) error {
	allSchemas := openapidocument.CollectSchemas(doc)
	if len(allSchemas) == 0 {
		return nil
	}

	for _, schema := range allSchemas {
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
				log.Info().Str("schema", schema.Title).Msg("Simplified polymorphic boolean schema to plain boolean")
			}
		}
	}

	return nil
}
