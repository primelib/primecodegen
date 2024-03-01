package openapidocument

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

type SchemaMatchFunc func(schema *base.Schema) bool

func AllSchemasMatch(schemas []*base.SchemaProxy, f SchemaMatchFunc) bool {
	for _, schemaProxy := range schemas {
		if !f(schemaProxy.Schema()) {
			return false
		}
	}

	return true
}

func IsEnumSchema(s *base.Schema) bool {
	// 3.0 enum
	if len(s.Enum) > 0 {
		return true
	}

	// 3.1 enum with oneOf and const
	if s.OneOf != nil {
		if AllSchemasMatch(s.OneOf, func(s *base.Schema) bool {
			return s.Const != nil
		}) {
			return true
		}
	}

	return false
}
