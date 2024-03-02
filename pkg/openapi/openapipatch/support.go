package openapipatch

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rs/zerolog/log"
)

func PruneOperationTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.Tags = nil
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
	commonPrefix := findPrefix(operationIds)
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

func GenerateOperationIds(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.OperationId = toOperationId(op.Key, url)
			log.Trace().Str("path", strings.ToUpper(op.Key)+" "+url).Str("operation-id", op.Value.OperationId).Msg("replacing operation id with generated id")
		}
	}

	return nil
}

func FlattenSchemas(doc *libopenapi.DocumentModel[v3.Document]) error {
	// flatten inline requests bodies
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.OperationId == "" {
				return fmt.Errorf("operation id is required for operation [%s], you can use generateOperationId to ensure all operations have a id", op.Key)
			}

			if op.Value.RequestBody != nil {
				for rb := op.Value.RequestBody.Content.Oldest(); rb != nil; rb = rb.Next() {
					if rb.Value.Schema.IsReference() { // skip references
						continue
					}

					// move schema to components and replace with reference
					key := op.Value.OperationId + "B" + strings.ToUpper(contentTypeToStr(rb.Key))
					log.Trace().Msg("moving request schema to components: " + key)
					if ref, err := moveSchemaIntoComponents(doc, key, rb.Value.Schema); err != nil {
						return fmt.Errorf("error moving schema to components: %w", err)
					} else if ref != nil {
						rb.Value.Schema = ref
					}
				}
			}
		}
	}

	// flatten inline responses
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.Responses.Codes == nil {
				continue
			}
			if op.Value.OperationId == "" {
				return fmt.Errorf("operation id is required for operation [%s], you can use generateOperationId to ensure all operations have a id", op.Key)
			}

			for resp := op.Value.Responses.Codes.Oldest(); resp != nil; resp = resp.Next() {
				if resp.Value.Content == nil {
					continue
				}

				responseCount := op.Value.Responses.Codes.Len()
				for rb := resp.Value.Content.Oldest(); rb != nil; rb = rb.Next() {
					// fix for raw responses without schema (e.g. plain text, yaml)
					if rb.Value.Schema == nil {
						rb.Value.Schema = base.CreateSchemaProxy(&base.Schema{
							Type:        []string{"string"},
							Description: "Shemaless response",
						})
					}

					if rb.Value.Schema.IsReference() { // skip references
						continue
					}

					// move schema to components and replace with reference
					key := op.Value.OperationId
					if responseCount > 1 { // if there are multiple responses, append response code to key
						key = key + "R" + resp.Key
					}
					log.Trace().Msg("moving response schema to components: " + key)
					if ref, err := moveSchemaIntoComponents(doc, key, rb.Value.Schema); err != nil {
						return fmt.Errorf("error moving schema to components: %w", err)
					} else if ref != nil {
						rb.Value.Schema = ref
					}
				}
			}
		}
	}

	// flatten inner schemas inside of components
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			// TODO: flatten inner schemas
			// out, _ := p.Value.Render()
			// fmt.Println(string(out))
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
