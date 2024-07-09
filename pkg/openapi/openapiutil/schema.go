package openapiutil

import (
	"slices"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func IsSchemaNullable(schema *base.Schema) bool {
	return ptr.ValueOrDefault(schema.Nullable, slices.Contains(schema.Type, "null")) // 3.1 uses null type, 3.0 uses nullable
}
