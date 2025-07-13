package openapidocument

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rs/zerolog/log"
)

func VisitAllSchemas(
	doc *libopenapi.DocumentModel[v3.Document],
	visitor func(name string, schema *base.SchemaProxy) *base.SchemaProxy,
) {
	if doc == nil {
		return
	}

	if doc.Model.Paths != nil && doc.Model.Paths.PathItems != nil {
		for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
			if path.Value == nil {
				log.Warn().Str("path", path.Key).Msg("Path item is nil, skipping")
				continue
			}
			for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
				if op.Value == nil {
					log.Warn().Str("operation", op.Key).Msg("Operation is nil, skipping")
					continue
				}

				// Visit request body
				if op.Value.RequestBody != nil && op.Value.RequestBody.Content != nil {
					for contentType := op.Value.RequestBody.Content.Oldest(); contentType != nil; contentType = contentType.Next() {
						if contentType.Value.Schema != nil {
							contentType.Value.Schema = visitNestedSchemas(contentType.Key, contentType.Value.Schema, visitor)
						}
					}
				}

				// Visit responses
				if op.Value.Responses != nil {
					for response := op.Value.Responses.Codes.Oldest(); response != nil; response = response.Next() {
						if response.Value != nil && response.Value.Content != nil {
							for contentType := response.Value.Content.Oldest(); contentType != nil; contentType = contentType.Next() {
								if contentType.Value.Schema != nil {
									contentType.Value.Schema = visitNestedSchemas(contentType.Key, contentType.Value.Schema, visitor)
								}
							}
						}
					}
				}
			}
		}
	}

	if doc.Model.Components.Schemas != nil {
		for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
			if schema.Value != nil {
				schema.Value = visitNestedSchemas(schema.Key, schema.Value, visitor)
			}
		}
	}
	if doc.Model.Components.Responses != nil {
		for response := doc.Model.Components.Responses.Oldest(); response != nil; response = response.Next() {
			if response.Value != nil && response.Value.Content != nil {
				for contentType := response.Value.Content.Oldest(); contentType != nil; contentType = contentType.Next() {
					if contentType.Value.Schema != nil {
						contentType.Value.Schema = visitNestedSchemas(contentType.Key, contentType.Value.Schema, visitor)
					}
				}
			}
		}
	}
	if doc.Model.Components.Parameters != nil {
		for parameter := doc.Model.Components.Parameters.Oldest(); parameter != nil; parameter = parameter.Next() {
			if parameter.Value != nil && parameter.Value.Schema != nil {
				parameter.Value.Schema = visitNestedSchemas(parameter.Key, parameter.Value.Schema, visitor)
			}
		}
	}
	if doc.Model.Components.RequestBodies != nil {
		for requestBody := doc.Model.Components.RequestBodies.Oldest(); requestBody != nil; requestBody = requestBody.Next() {
			if requestBody.Value != nil && requestBody.Value.Content != nil {
				for contentType := requestBody.Value.Content.Oldest(); contentType != nil; contentType = contentType.Next() {
					if contentType.Value.Schema != nil {
						contentType.Value.Schema = visitNestedSchemas(contentType.Key, contentType.Value.Schema, visitor)
					}
				}
			}
		}
	}
	if doc.Model.Components.Headers != nil {
		for header := doc.Model.Components.Headers.Oldest(); header != nil; header = header.Next() {
			if header.Value != nil && header.Value.Schema != nil {
				header.Value.Schema = visitNestedSchemas(header.Key, header.Value.Schema, visitor)
			}
		}
	}
}

func visitNestedSchemas(
	key string,
	schema *base.SchemaProxy,
	visitor func(name string, schema *base.SchemaProxy) *base.SchemaProxy,
) *base.SchemaProxy {
	if schema == nil {
		return nil
	}

	// call visitor for the current schema
	schema = visitor(key, schema)

	// abort if the schema is a reference
	if schema.IsReference() {
		return schema
	}

	// resolve full schema
	s := schema.Schema()

	// Visit properties
	if s.Properties != nil {
		for prop := s.Properties.Oldest(); prop != nil; prop = prop.Next() {
			prop.Value = visitNestedSchemas(prop.Key, prop.Value, visitor)
		}
	}

	// Visit allOf, anyOf, oneOf
	for _, subSchema := range s.AllOf {
		subSchema = visitNestedSchemas("", subSchema, visitor)
	}
	for _, subSchema := range s.AnyOf {
		subSchema = visitNestedSchemas("", subSchema, visitor)
	}
	for _, subSchema := range s.OneOf {
		subSchema = visitNestedSchemas("", subSchema, visitor)
	}

	// Visit items (for arrays)
	if s.Items != nil {
		if s.Items.IsA() {
			s.Items.A = visitNestedSchemas("", s.Items.A, visitor)
		}
	}

	// Visit additionalProperties (freeform objects)
	if s.AdditionalProperties != nil {
		if s.AdditionalProperties.IsA() {
			s.AdditionalProperties.A = visitNestedSchemas("", s.AdditionalProperties.A, visitor)
		}
	}

	// Not
	if s.Not != nil {
		s.Not = visitNestedSchemas("", s.Not, visitor)
	}

	return schema
}
