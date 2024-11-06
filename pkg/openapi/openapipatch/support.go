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

// CreateOperationTagsFromDocTitle removes all tags and creates one new tag per API spec doc from document title setting it on each operation.
// Note: This patch must be applied before merging specs.
func CreateOperationTagsFromDocTitle(doc *libopenapi.DocumentModel[v3.Document]) error {

	// Remove all tags
	doc.Model.Tags = nil
	PruneOperationTags(doc)
	// Create tag from document title
	documenttag := doc.Model.Info.Title
	doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: documenttag, Description: "See document description"})
	// Set tag on each operation
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Str("tag", documenttag).Msg("operation is missing tags, adding default tag:")
				op.Value.Tags = append(op.Value.Tags, documenttag)
			} else {
				log.Warn().Strs("Operation Tag", op.Value.Tags).Msg("Found non-empty operation tag - ")
			}
		}
	}

	return nil
}

// FixOperationTags ensures all operations have tags, and that tags are documented in the document
func FixOperationTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	documentedTags := make(map[string]bool)
	for _, tag := range doc.Model.Tags {
		documentedTags[tag.Name] = true
	}

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Msg("operation is missing tags, adding default tag")
				op.Value.Tags = append(op.Value.Tags, "default")
			} else {
				// ensure all tags are documented
				for _, tag := range op.Value.Tags {
					if _, ok := documentedTags[tag]; !ok {
						log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Str("tag", tag).Msg("tag is not documented, adding to document")
						doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: tag})
						documentedTags[tag] = true
					}
				}
			}
		}
	}

	return nil
}

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

// Inlines properties of allOf-referenced schemas and removes allOf-references in schemas
func InlineAllOfHierarchies(doc *libopenapi.DocumentModel[v3.Document]) error {
	// component schemas
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		log.Debug().Str("schema", schema.Key).Msg("Inlining properties of allOf-referenced schemas in components.schema")
		_, err := openapidocument.InlineAllOf(schema.Value)
		if err != nil {
			return err
		}

		if schema.Value.Schema().Properties != nil {
			for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
				log.Debug().Str("Property", p.Key).Msg("Inlining properties of allOf in schema.Value.Schema().Properties")
				_, propErr := openapidocument.InlineAllOf(p.Value)
				if propErr != nil {
					return propErr
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
