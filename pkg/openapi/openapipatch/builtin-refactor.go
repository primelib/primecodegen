package openapipatch

import (
	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

var AddIdempotencyKeyPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "add-idempotency-key",
	Description:         "Adds an idempotency key to all POST operations in the OpenAPI document (see https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header)",
	PatchV3DocumentFunc: AddIdempotencyKey,
}

// AddIdempotencyKey adds an idempotency key to all POST operations in the OpenAPI document - see https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/
func AddIdempotencyKey(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Key == "post" {
				if op.Value.Parameters == nil {
					op.Value.Parameters = []*v3.Parameter{}
				}

				log.Trace().Str("path", path.Key).Str("op", op.Key).Msg("adding idempotency key as header parameter")
				op.Value.Parameters = append(op.Value.Parameters, &v3.Parameter{
					Name:        "Idempotency-Key",
					In:          "header",
					Description: "A unique key to ensure idempotency of the request",
					Required:    ptr.True(),
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:   []string{"string"},
						Format: "uuid",
					}),
				})
			}
		}
	}

	return nil
}

var SetOperationTagPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "set-operation-tag",
	Description:         "Sets a tag for all operations in the OpenAPI document",
	PatchV3DocumentFunc: SetOperationTag,
}

func SetOperationTag(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// validate config
	tag, err := getStringConfig(config, "tag")
	if err != nil {
		return err
	}

	// set tag for all operations
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.Tags = []string{tag}
		}
	}

	return nil
}

var AddPathPrefixPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "add-path-prefix",
	Description:         "Adds a prefix to all paths in the OpenAPI document",
	PatchV3DocumentFunc: AddPathPrefix,
}

func AddPathPrefix(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// validate config
	prefix, err := getStringConfig(config, "prefix")
	if err != nil {
		return err
	}

	// rename path keys
	_ = util.RenameOrderedMapKeys(
		doc.Model.Paths.PathItems,
		func(oldKey string) string {
			return prefix + oldKey
		},
	)

	return nil
}

var AddComponentSchemaPrefixPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "add-component-schema-prefix",
	Description:         "Adds a prefix to all component schemas in the OpenAPI document",
	PatchV3DocumentFunc: AddComponentSchemaPrefix,
}

func AddComponentSchemaPrefix(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// validate config
	prefix, err := getStringConfig(config, "prefix")
	if err != nil {
		return err
	}

	// rename keys and update references
	referenceMapping := util.RenameOrderedMapKeys(
		doc.Model.Components.Schemas,
		func(oldKey string) string {
			return prefix + oldKey
		},
	)
	refMapping := make(map[string]string)
	for oldKey, newKey := range referenceMapping {
		refMapping["#/components/schemas/"+oldKey] = "#/components/schemas/" + newKey
	}
	updateAllSchemaRefs(doc, refMapping)

	return nil
}

func updateAllSchemaRefs(
	doc *libopenapi.DocumentModel[v3.Document],
	referenceMapping map[string]string,
) {
	log.Trace().Int("numRefs", len(referenceMapping)).Msg("updating schema references in document")
	openapidocument.VisitAllSchemas(doc, func(name string, schema *base.SchemaProxy) *base.SchemaProxy {
		if schema.IsReference() {
			if newReference, ok := referenceMapping[schema.GetReference()]; ok {
				log.Trace().Str("oldRef", schema.GetReference()).Str("newRef", newReference).Msg("updating schema reference")
				schema = base.CreateSchemaProxyRef(newReference)
			}
		}
		return schema
	})
}
