package openapipatch

import (
	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
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

	// store original paths and values to avoid modifying the map while iterating
	var originalPathKeys []string
	var originalValues []*v3.PathItem
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		if path.Key == "" {
			continue
		}

		originalPathKeys = append(originalPathKeys, path.Key)
		originalValues = append(originalValues, path.Value)
	}

	// delete the original paths
	for _, key := range originalPathKeys {
		doc.Model.Paths.PathItems.Delete(key)
	}

	// add the prefix to each original path key and re-add the path items
	for i, originalPathKey := range originalPathKeys {
		newPathKey := prefix + originalPathKey
		log.Trace().Str("originalPath", originalPathKey).Str("newPath", newPathKey).Msg("changing path")
		if _, exists := doc.Model.Paths.PathItems.Get(newPathKey); !exists {
			doc.Model.Paths.PathItems.Set(newPathKey, originalValues[i])
		} else {
			log.Error().Str("originalPath", originalPathKey).Str("newPath", newPathKey).Msg("path already exists, skipping")
		}
	}

	return nil
}
