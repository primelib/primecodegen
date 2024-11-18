package openapipatch

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
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

	return nil
}

func flattenInlineRequestBodies(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if op.Value.OperationId == "" {
				return fmt.Errorf("operation id is required for operation [%s] of path[%s], you can use generateOperationId to ensure all operations have a id", op.Key, path.Key)
			}

			if op.Value.RequestBody != nil {
				for rb := op.Value.RequestBody.Content.Oldest(); rb != nil; rb = rb.Next() {
					if rb.Value.Schema.IsReference() { // skip references
						continue
					}
					addSuffix := op.Value.RequestBody.Content.First().Key() != rb.Key // add suffix from the second request body onwards

					// move schema to components and replace with reference
					key := util.ToPascalCase(op.Value.OperationId) + "B" + util.Ternary(addSuffix, util.UpperCaseFirstLetter(util.ContentTypeToShortName(rb.Key)), "")
					log.Trace().Msg("moving request schema to components: " + key)
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

func flattenInlineResponses(doc *libopenapi.DocumentModel[v3.Document]) error {
	// flatten inline responses
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

func flattenInnerSchemas(doc *libopenapi.DocumentModel[v3.Document]) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			if !p.Value.IsReference() && slices.Contains(p.Value.Schema().Type, "object") {
				key := util.ToPascalCase(p.Key)
				log.Trace().Msg("moving inner schema to components: " + key)
				if ref, err := moveSchemaIntoComponents(doc, key, p.Value); err != nil {
					return fmt.Errorf("error moving enum property schema to components: %w", err)
				} else if ref != nil {
					p.Value = ref
				}
			}

		}

		// flatten anyOf
		s := schema.Value.Schema()
		if len(s.AnyOf) > 0 {

		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			// simplify polymorphism
			_, err := openapidocument.SimplifyPolymorphism(p.Value)
			if err != nil {
				return err
			}

			/*
				// flatten anyOf
				if p.Value.Schema().AnyOf != nil {
					for i, s := range p.Value.Schema().AnyOf {
						if s.IsReference() {
							continue
						}

						// move schema to components and replace with reference
						key := schema.Key + "AnyOf" + fmt.Sprint(i)
						log.Trace().Msg("moving anyOf schema to components: " + key)
						if ref, err := moveSchemaIntoComponents(doc, key, s); err != nil {
							return fmt.Errorf("error moving schema to components: %w", err)
						} else if ref != nil {
							p.Value.Schema().AnyOf[i] = ref
						}
					}
				}

				// flatten oneOf
				if p.Value.Schema().OneOf != nil {
					for i, s := range p.Value.Schema().OneOf {
						if s.IsReference() {
							continue
						}

						// move schema to components and replace with reference
						key := schema.Key + "OneOf" + fmt.Sprint(i)
						log.Trace().Msg("moving oneOf schema to components: " + key)
						if ref, err := moveSchemaIntoComponents(doc, key, s); err != nil {
							return fmt.Errorf("error moving schema to components: %w", err)
						} else if ref != nil {
							p.Value.Schema().OneOf[i] = ref
						}
					}
				}

				// flatten allOf
				if p.Value.Schema().AllOf != nil {
					for i, s := range p.Value.Schema().AllOf {
						if s.IsReference() {
							continue
						}

						// move schema to components and replace with reference
						key := schema.Key + "AllOf" + fmt.Sprint(i)
						log.Trace().Msg("moving allOf schema to components: " + key)
						if ref, err := moveSchemaIntoComponents(doc, key, s); err != nil {
							return fmt.Errorf("error moving schema to components: %w", err)
						} else if ref != nil {
							p.Value.Schema().AllOf[i] = ref
						}
					}
				}
			*/
		}
	}

	return nil
}
