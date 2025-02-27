package openapipatch

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

// MergePolymorphicSchemas merges polymorphic schemas (anyOf, oneOf, allOf) into a single flat schema
func MergePolymorphicSchemas(v3Model *libopenapi.DocumentModel[v3.Document]) error {
	// Remember derived schemata (key) to be replaced by their base schemata (value)
	derivedSchemaReplacementMap := make(map[string]string)

	// component schemas
	for schema := v3Model.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		log.Debug().Str("components.schema", schema.Key).Msg("merging")
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

// MergePolymorphicProperties merges polymorphic property values (anyOf, oneOf, allOf) into a single flat schema referenced by resp. properties
func MergePolymorphicProperties(v3Model *libopenapi.DocumentModel[v3.Document]) error {
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

// MissingSchemaTitle fills in missing schema titles with the schema key
func MissingSchemaTitle(doc *libopenapi.DocumentModel[v3.Document]) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Title == "" {
			schema.Value.Schema().Title = schema.Key
			log.Trace().Str("schema", schema.Key).Msg("missing schema title, setting to schema key")
		}
	}

	return nil
}

// PruneCommonOperationIdPrefix sets the operation IDs of all operations and fixes some commonly seen issues.
func PruneCommonOperationIdPrefix(doc *libopenapi.DocumentModel[v3.Document]) error {
	var operationIds []string

	// scan all current operation IDs
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			operationIds = append(operationIds, op.Value.OperationId)
		}
	}

	// detect common prefix
	commonPrefix := util.FindCommonStrPrefix(operationIds)
	if commonPrefix != "" {
		log.Debug().Str("prefix", commonPrefix).Msg("found common operation id prefix, removing it from all operation IDs")
		for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
			for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
				op.Value.OperationId = strings.TrimPrefix(op.Value.OperationId, commonPrefix)
			}
		}
	}

	return nil
}

// InvalidMaxValue fixes integers and longs, where the maximum value is out of bounds for the type
func InvalidMaxValue(doc *libopenapi.DocumentModel[v3.Document]) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			s := p.Value.Schema()
			if slices.Contains(s.Type, "integer") && p.Value.Schema().Maximum != nil {
				if *p.Value.Schema().Maximum > 2147483647 {
					// p.Value.Schema().Maximum = float64(2147483647)
					log.Trace().Str("schema", schema.Key).Str("property", p.Key).Msg("fixing maximum value for integer")
				}
			}
		}
	}

	return nil
}

func moveSchemaIntoComponents(doc *libopenapi.DocumentModel[v3.Document], key string, schema *base.SchemaProxy) (*base.SchemaProxy, error) {
	if schema.IsReference() { // skip references
		return nil, nil
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
		if isEmptySchema(value.Schema()) {
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

func isEmptySchema(schema *base.Schema) bool {
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
