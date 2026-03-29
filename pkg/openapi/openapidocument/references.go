package openapidocument

import (
	"fmt"

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
func CollectSchemas(doc *libopenapi.DocumentModel[v3.Document]) map[string]*base.Schema {
	schemas := make(map[string]*base.Schema)

	// component schemas
	for s := doc.Model.Components.Schemas.Oldest(); s != nil; s = s.Next() {
		ref := "#/components/schemas/" + s.Key
		schemas[ref] = s.Value.Schema()
	}

	// component parameters
	for s := doc.Model.Components.Parameters.Oldest(); s != nil; s = s.Next() {
		if s.Value.Schema != nil && !s.Value.Schema.IsReference() {
			ref := "#/components/parameters/" + s.Key
			schemas[ref] = s.Value.Schema.Schema()
		}
	}

	// requests parameters
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			for _, param := range op.Value.Parameters {
				if param.Schema != nil && !param.Schema.IsReference() {
					ref := "#/paths/" + path.Key + "/" + op.Key + "/parameters/" + param.Name
					schemas[ref] = param.Schema.Schema()
				}
			}
		}
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

// CollectSchemaProxies collects all schema proxies from the OpenAPI document.
func CollectSchemaProxies(doc *libopenapi.DocumentModel[v3.Document]) map[string]*base.SchemaProxy {
	proxies := make(map[string]*base.SchemaProxy)

	// Component Schemas
	if doc.Model.Components != nil && doc.Model.Components.Schemas != nil {
		for s := doc.Model.Components.Schemas.Oldest(); s != nil; s = s.Next() {
			ref := "#/components/schemas/" + s.Key
			proxies[ref] = s.Value
		}
	}

	// Component Parameters
	if doc.Model.Components != nil && doc.Model.Components.Parameters != nil {
		for s := doc.Model.Components.Parameters.Oldest(); s != nil; s = s.Next() {
			if s.Value.Schema != nil {
				ref := "#/components/parameters/" + s.Key
				proxies[ref] = s.Value.Schema
			}
		}
	}

	// Path/Operation Parameters
	if doc.Model.Paths != nil && doc.Model.Paths.PathItems != nil {
		for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
			for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
				for _, param := range op.Value.Parameters {
					if param.Schema != nil {
						ref := fmt.Sprintf("#/paths/%s/%s/parameters/%s", path.Key, op.Key, param.Name)
						proxies[ref] = param.Schema
					}
				}
			}
		}
	}

	return proxies
}
