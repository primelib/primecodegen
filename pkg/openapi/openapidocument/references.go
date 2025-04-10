package openapidocument

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// CollectSchemas collects all schemas from the OpenAPI document (doesn't follow recursion, flatten should be used first).
func CollectSchemas(doc *libopenapi.DocumentModel[v3.Document]) []*base.Schema {
	var schemas []*base.Schema

	// requests parameters
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			for _, param := range op.Value.Parameters {
				if param.Schema != nil && !param.Schema.IsReference() {
					schemas = append(schemas, param.Schema.Schema())
				}
			}
		}
	}

	// components
	for s := doc.Model.Components.Parameters.Oldest(); s != nil; s = s.Next() {
		if s.Value.Schema != nil && !s.Value.Schema.IsReference() {
			schemas = append(schemas, s.Value.Schema.Schema())
		}
	}
	for s := doc.Model.Components.Schemas.Oldest(); s != nil; s = s.Next() {
		schemas = append(schemas, s.Value.Schema())

		// properties
		if s.Value.Schema().Properties != nil {
			for p := s.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
				schemas = append(schemas, p.Value.Schema())
			}
		}
	}

	return schemas
}
