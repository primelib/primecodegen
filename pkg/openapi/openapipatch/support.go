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

func PruneOperationTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.Tags = nil
		}
	}

	return nil
}

func PruneOperationTagsExceptFirst(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) > 1 {
				op.Value.Tags = op.Value.Tags[:1]
			}
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

func GenerateOperationIds(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			originalOperationId := op.Value.OperationId
			op.Value.OperationId = util.ToOperationId(op.Key, url)

			log.Trace().Str("path", strings.ToUpper(op.Key)+" "+url).Str("operation-id", op.Value.OperationId).Str("original-operation-id", originalOperationId).Msg("replacing operation id with generated id")
		}
	}

	return nil
}

// MergePolymorphicSchemas merges polymorphic schemas (anyOf, oneOf, allOf) into a single flat schema
func MergePolymorphicSchemas(doc *libopenapi.DocumentModel[v3.Document]) error {
	// component schemas
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		// TODO: remove
		log.Debug().Str("schema", schema.Key).Msg("merging components.schema")

		_, err := openapidocument.SimplifyPolymorphism(schema.Value)
		if err != nil {
			return err
		}

		if schema.Value.Schema().Properties != nil {
			for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
				_, propErr := openapidocument.SimplifyPolymorphism(p.Value)
				if propErr != nil {
					return propErr
				}
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
