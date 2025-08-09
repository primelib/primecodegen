package openapipatch

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"
)

func moveSchemaIntoComponents(doc *libopenapi.DocumentModel[v3.Document], key string, schema *base.SchemaProxy) (*base.SchemaProxy, error) {
	if schema.IsReference() { // skip references
		return nil, nil
	}

	// check for key conflict
	if existingSchema, present := doc.Model.Components.Schemas.Get(key); present {
		if openapidocument.Compare(schema, existingSchema) {
			// match: return ref to existing schema
			return base.CreateSchemaProxyRef("#/components/schemas/" + key), nil
		} else {
			// mismatch: append suffix to avoid conflict
			key = key + openapidocument.HashSchema(schema)
		}
	}

	// add schema to components
	s, err := schema.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}
	doc.Model.Components.Schemas.Set(key, base.CreateSchemaProxy(s))

	// return reference to new schema
	return base.CreateSchemaProxyRef("#/components/schemas/" + key), nil
}

func deleteEmptySchemas(v3Model *libopenapi.DocumentModel[v3.Document], schemataMap map[string]string) {
	var keysForDeletion []string

	for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		log.Trace().Str("components.schema", schema.Key).Msg("checking for empty schemas")
		value, present := v3Model.Model.Components.Schemas.Get(schema.Key)
		if !present {
			continue
		}
		if openapidocument.IsEmptySchema(value.Schema()) {
			keysForDeletion = append(keysForDeletion, schema.Key)
		}
	}
	for deleteKeyIdx := range keysForDeletion {
		derivedSchemaReplacement := schemataMap[keysForDeletion[deleteKeyIdx]]
		log.Info().Str("key", keysForDeletion[deleteKeyIdx]).Str("replacement", derivedSchemaReplacement).Msg("Replacement for empty schema")
		log.Info().Str("key", keysForDeletion[deleteKeyIdx]).Msg("Deleting empty schema")
		replaceEmptySchemaRefsByBaseSchemaRefs(keysForDeletion[deleteKeyIdx], derivedSchemaReplacement, v3Model)
		v3Model.Model.Components.Schemas.Delete(keysForDeletion[deleteKeyIdx])
	}
}

// Replace refs to schemas merged into their base-schemas with refs to these base-schemas everywhere
func replaceEmptySchemaRefsByBaseSchemaRefs(derivedEmptySchema string, baseSchemaReplacement string, v3Model *libopenapi.DocumentModel[v3.Document]) error {
	derivedEmptySchemaRef := "#/components/schemas/" + derivedEmptySchema
	baseSchemaReplacementRef := "#/components/schemas/" + baseSchemaReplacement

	log.Debug().Str("schema", derivedEmptySchema).Msg("checking references of empty")

	// properties
	for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {

		if schema.Value.Schema().Properties != nil {
			for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
				if p.Value.GetReference() == derivedEmptySchemaRef {
					log.Info().Str("schema", derivedEmptySchema).Str("with base schema", baseSchemaReplacement).Msg("replacing derived empty")
					schemaRefReplacementSP := base.CreateSchemaProxyRef(baseSchemaReplacementRef)
					p.Value = schemaRefReplacementSP
				}
			}
		}
	}
	// component.responses
	for response := v3Model.Model.Components.Responses.Oldest(); response != nil; response = response.Next() {
		// TODO
	}
	// component.requestbodies
	for reqBody := v3Model.Model.Components.RequestBodies.Oldest(); reqBody != nil; reqBody = reqBody.Next() {
		// TODO
	}

	// component.headers
	for header := v3Model.Model.Components.Headers.Oldest(); header != nil; header = header.Next() {
		// TODO
	}
	// component.parameters
	for param := v3Model.Model.Components.Parameters.Oldest(); param != nil; param = param.Next() {
		// TODO
	}

	return nil
}
