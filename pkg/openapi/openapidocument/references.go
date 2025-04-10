package openapidocument

import (
	"github.com/pb33f/doctor/model"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type WalkDocumentResult struct {
	Schemas []*base.Schema
}

func WalkDocument(doc *libopenapi.DocumentModel[v3.Document]) WalkDocumentResult {
	result := WalkDocumentResult{
		Schemas: make([]*base.Schema, 0),
	}

	walker := model.NewDrDocument(doc)
	for _, s := range walker.Schemas {
		// might be able to fill missing titles with s.Name
		result.Schemas = append(result.Schemas, s.Value)
	}

	return result
}

// CollectSchemas collects all schemas from the OpenAPI document (doesn't follow recursion, flatten should be used first).
func CollectSchemas(doc *libopenapi.DocumentModel[v3.Document]) []*base.Schema {
	var schemas []*base.Schema

	// requests parameters
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			for _, param := range op.Value.Parameters {
				if param.Schema != nil && !param.Schema.IsReference() {
					schemas = append(schemas, collectFromSchema(param.Schema)...)
				}
			}
		}
	}

	// components
	for s := doc.Model.Components.Parameters.Oldest(); s != nil; s = s.Next() {
		if s.Value.Schema != nil && !s.Value.Schema.IsReference() {
			schemas = append(schemas, collectFromSchema(s.Value.Schema)...)
		}
	}
	for s := doc.Model.Components.Schemas.Oldest(); s != nil; s = s.Next() {
		schemas = append(schemas, collectFromSchema(s.Value)...)
	}

	return schemas
}

func CollectOperations(doc *libopenapi.DocumentModel[v3.Document]) []*v3.Operation {
	var operations []*v3.Operation

	// requests parameters
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			operations = append(operations, op.Value)
		}
	}

	return operations
}

func collectFromSchema(schema *base.SchemaProxy) []*base.Schema {
	schemas := []*base.Schema{schema.Schema()}

	// properties
	if schema.Schema().Properties != nil {
		for p := schema.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			schemas = append(schemas, p.Value.Schema())
		}
	}

	return schemas
}
