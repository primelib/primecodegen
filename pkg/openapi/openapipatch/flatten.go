package openapipatch

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func FlattenSchemas(doc *libopenapi.DocumentModel[v3.Document]) error {
	err := flattenInlineRequestBodies(doc)
	if err != nil {
		return err
	}

	err = flattenInlineResponses(doc)
	if err != nil {
		return err
	}

	err = flattenEnumsInComponentProperties(doc)
	if err != nil {
		return err
	}

	err = flattenInnerSchemas(doc)
	if err != nil {
		return err
	}

	err = flattenCallbacks(doc)
	if err != nil {
		return err
	}

	err = flattenWebhooks(doc)
	if err != nil {
		return err
	}

	return nil
}

// flattenInlineRequestBodies flattens inline request schemas into the components section of the document
func flattenInlineRequestBodies(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.OperationId == "" {
				return fmt.Errorf("operation id is required for operation [%s] of path[%s], you can use generateOperationId to ensure all operations have a id", op.Key, path.Key)
			}

			if op.Value.RequestBody != nil {
				err := processRequestBody(doc, op.Value, op.Value.RequestBody, "%sB%s", fmt.Sprintf("operation: %s / path: %s", op.Key, path.Key))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// flattenInlineResponses flattens inline response schemas into the components section of the document
func flattenInlineResponses(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.Responses.Codes == nil {
				continue
			}
			if op.Value.OperationId == "" {
				return fmt.Errorf("operation id is required for operation [%s] of path[%s], you can use generateOperationId to ensure all operations have a id", op.Key, path.Key)
			}

			for resp := op.Value.Responses.Codes.Oldest(); resp != nil; resp = resp.Next() {
				if resp.Value.Content == nil {
					continue
				}

				responseCount := op.Value.Responses.Codes.Len()
				for rb := resp.Value.Content.Oldest(); rb != nil; rb = rb.Next() {
					// fix for raw responses without schema (e.g. plain text, yaml)
					if rb.Value.Schema == nil {
						rb.Value.Schema = base.CreateSchemaProxy(&base.Schema{
							Type:        []string{"string"},
							Description: "Shemaless response",
						})
					}

					if rb.Value.Schema.IsReference() { // skip references
						continue
					}

					// move schema to components and replace with reference
					key := util.ToPascalCase(op.Value.OperationId)
					if responseCount > 1 { // if there are multiple responses, append response code to key
						key = key + "R" + resp.Key
					}
					log.Trace().Msg("moving response schema to components: " + key)
					if ref, err := moveSchemaIntoComponents(doc, key, rb.Value.Schema); err != nil {
						return fmt.Errorf("error moving schema to components: %w", err)
					} else if ref != nil {
						rb.Value.Schema = ref
					}
				}
			}
		}
	}

	return nil
}

// flattenEnumsInComponentProperties flattens enum values in component properties
func flattenEnumsInComponentProperties(doc *libopenapi.DocumentModel[v3.Document]) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		// TODO: check of a schema with that name already exists, skip if equal - change name if not equal
		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			if p.Value.Schema().Enum != nil {
				key := util.ToPascalCase(p.Key) + "Enum"
				log.Trace().Msg("moving property enum to components: " + key)
				if ref, err := moveSchemaIntoComponents(doc, key, p.Value); err != nil {
					return fmt.Errorf("error moving enum property schema to components: %w", err)
				} else if ref != nil {
					p.Value = ref
				}
			}
		}
	}

	return nil
}

// flattenInnerSchemas flattens inner component schemas in components
func flattenInnerSchemas(doc *libopenapi.DocumentModel[v3.Document]) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		valueSchema := schema.Value.Schema()

		// top-level schema
		if slices.Contains(valueSchema.Type, "array") && valueSchema.Items != nil {
			itemSchema := valueSchema.Items.A.Schema()
			if !valueSchema.Items.A.IsReference() && slices.Contains(itemSchema.Type, "object") {
				key := util.ToPascalCase(schema.Key) + "Item"
				log.Trace().Msg("moving top-level array inner schema to components: " + key)
				if ref, err := moveSchemaIntoComponents(doc, key, valueSchema.Items.A); err != nil {
					return fmt.Errorf("error moving top-level array object schema to components: %w", err)
				} else if ref != nil {
					valueSchema.Items.A = ref
				}
			}
		}

		// properties
		if valueSchema.Properties == nil {
			continue
		}
		for p := valueSchema.Properties.Oldest(); p != nil; p = p.Next() {
			propSchema := p.Value.Schema()
			if p.Value.IsReference() {
				continue
			}

			// inner objects
			if slices.Contains(propSchema.Type, "object") {
				key := util.ToPascalCase(p.Key)
				log.Trace().Msg("moving inner schema to components: " + key)
				if ref, err := moveSchemaIntoComponents(doc, key, p.Value); err != nil {
					return fmt.Errorf("error moving enum property schema to components: %w", err)
				} else if ref != nil {
					p.Value = ref
				}
			}

			// inner array objects
			if slices.Contains(propSchema.Type, "array") && propSchema.Items != nil {
				itemSchema := propSchema.Items.A.Schema()
				if !propSchema.Items.A.IsReference() && slices.Contains(itemSchema.Type, "object") {
					key := util.ToPascalCase(p.Key) + "Item"
					log.Trace().Msg("moving array inner schema to components: " + key)
					if ref, err := moveSchemaIntoComponents(doc, key, propSchema.Items.A); err != nil {
						return fmt.Errorf("error moving array object schema to components: %w", err)
					} else if ref != nil {
						propSchema.Items.A = ref
					}
				}
			}
		}
	}

	return nil
}

// flattenCallbacks flattens inline callback request schemas into the components section of the document
func flattenCallbacks(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.Callbacks == nil {
				continue
			}

			for callback := op.Value.Callbacks.Oldest(); callback != nil; callback = callback.Next() {
				for ce := callback.Value.Expression.Oldest(); ce != nil; ce = ce.Next() {
					for cop := ce.Value.GetOperations().Oldest(); cop != nil; cop = cop.Next() {
						if cop.Value.Responses.Codes == nil {
							continue
						}

						err := processRequestBody(doc, op.Value, op.Value.RequestBody, "%sWH%s", fmt.Sprintf("operation: %s / path: %s", op.Key, callback.Key))
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// flattenWebhooks flattens inline webhook request schemas into the components section of the document
func flattenWebhooks(doc *libopenapi.DocumentModel[v3.Document]) error {
	if doc.Model.Webhooks == nil {
		return nil
	}

	for webhook := doc.Model.Webhooks.Oldest(); webhook != nil; webhook = webhook.Next() {
		for op := webhook.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.Responses.Codes == nil {
				continue
			}

			err := processRequestBody(doc, op.Value, op.Value.RequestBody, "%sWH%s", fmt.Sprintf("operation: %s / path: %s", op.Key, webhook.Key))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processRequestBody processes the request body of an operation
func processRequestBody(doc *libopenapi.DocumentModel[v3.Document], operation *v3.Operation, requestBody *v3.RequestBody, schemaKeyTemplate string, location string) error {
	for rb := requestBody.Content.Oldest(); rb != nil; rb = rb.Next() {
		if rb.Value.Schema.IsReference() { // skip references
			continue
		}
		addSuffix := requestBody.Content.First().Key() != rb.Key // add suffix from the second request body onwards

		if operation.OperationId == "" {
			return fmt.Errorf("operation id is required [%s], you can use generateOperationId to ensure all operations have a id", location)
		}

		// move schema to components and replace with reference
		key := fmt.Sprintf(schemaKeyTemplate, util.ToPascalCase(operation.OperationId), util.Ternary(addSuffix, util.UpperCaseFirstLetter(util.ContentTypeToShortName(rb.Key)), ""))
		log.Trace().Msg("moving request schema to components: " + key)
		if ref, err := moveSchemaIntoComponents(doc, key, rb.Value.Schema); err != nil {
			return fmt.Errorf("error moving schema to components: %w", err)
		} else if ref != nil {
			rb.Value.Schema = ref
		}
	}

	return nil
}
