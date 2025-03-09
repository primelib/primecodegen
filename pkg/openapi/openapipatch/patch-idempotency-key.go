package openapipatch

import (
	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rs/zerolog/log"
)

// AddIdempotencyKey adds an idempotency key to all POST operations in the OpenAPI document - see https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/
func AddIdempotencyKey(doc *libopenapi.DocumentModel[v3.Document]) error {
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
