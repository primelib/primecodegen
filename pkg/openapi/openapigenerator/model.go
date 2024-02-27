package openapigenerator

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func BuildTemplateData(doc *libopenapi.DocumentModel[v3.Document], generator CodeGenerator) (DocumentModel, error) {
	var template = DocumentModel{}

	operations, err := BuildOperations(OperationOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Operations = operations

	models, err := BuildModels(ModelOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Models = models

	return template, nil
}

type OperationOpts struct {
	Generator CodeGenerator
	Doc       *libopenapi.DocumentModel[v3.Document]
}

func BuildOperations(opts OperationOpts) ([]Operation, error) {
	var operations []Operation

	for path := opts.Doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			// operation
			operation := Operation{
				Path:        path.Key,
				Method:      op.Key,
				Summary:     op.Value.Summary,
				Description: op.Value.Description,
				Tags:        op.Value.Tags,
				OperationId: op.Value.OperationId,
				Deprecated:  getBoolValue(op.Value.Deprecated, false),
				// DeprecatedReason: op.Value.Extensions.Get("x-deprecated"),
			}

			// operation parameters
			for _, param := range op.Value.Parameters {
				pSchema, err := param.Schema.BuildSchema()
				if err != nil {
					return operations, fmt.Errorf("error building property schema: %w", err)
				}

				pType, err := opts.Generator.ToCodeType(param.Schema.Schema())
				if err != nil {
					return operations, fmt.Errorf("error converting type: %w", err)
				}

				pKind := KindVar
				if len(pSchema.Enum) > 0 {
					pKind = KindEnum
				}
				var allowedValues []string
				for _, e := range pSchema.Enum {
					allowedValues = append(allowedValues, e.Value)
				}
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:          param.Name,
					In:            param.In,
					Description:   param.Description,
					Kind:          pKind,
					Type:          pType,
					AllowedValues: allowedValues,
					Required:      getBoolValue(param.Required, false),
					Deprecated:    param.Deprecated,
					// DeprecatedReason: param.Value.Extensions.Get("x-deprecated"),
				})
			}

			// request body
			if rb := op.Value.RequestBody; rb != nil {
				// TODO: set correct type for request body
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        "payload",
					In:          "body",
					Description: rb.Description,
					Kind:        KindVar,
					Type:        "string",
				})
			}

			operations = append(operations, operation)
		}
	}

	return operations, nil
}

type ModelOpts struct {
	Generator CodeGenerator
	Doc       *libopenapi.DocumentModel[v3.Document]
}

func BuildModels(opts ModelOpts) ([]Model, error) {
	var models []Model

	for schema := opts.Doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		s, err := schema.Value.BuildSchema()
		if err != nil {
			return models, fmt.Errorf("error building component schema: %w", err)
		}

		add := Model{
			Name:        opts.Generator.ToClassName(s.Title),
			Description: s.Description,
		}
		if slices.Contains(s.Type, "object") && s.Properties != nil {
			for p := s.Properties.Oldest(); p != nil; p = p.Next() {
				pSchema, pErr := p.Value.BuildSchema()
				if pErr != nil {
					return models, fmt.Errorf("error building property schema: %w", err)
				}

				pType, err := opts.Generator.ToCodeType(pSchema)
				if err != nil {
					return models, fmt.Errorf("error converting type: %w", err)
				}

				pKind := KindVar
				if len(pSchema.Enum) > 0 {
					pKind = KindEnum
				}
				var allowedValues []string
				for _, e := range pSchema.Enum {
					allowedValues = append(allowedValues, e.Value)
				}
				add.Properties = append(add.Properties, Property{
					Name:          p.Key,
					Kind:          pKind,
					Title:         pSchema.Title,
					Type:          pType,
					Nullable:      getBoolValue(pSchema.Nullable, slices.Contains(pSchema.Type, "null")), // 3.1 uses null type, 3.0 uses nullable
					AllowedValues: allowedValues,
				})
			}
		} else if slices.Contains(s.Type, "array") {
			mParent, err := opts.Generator.ToCodeType(s)
			if err != nil {
				return models, fmt.Errorf("error converting type: %w", err)
			}

			add.Parent = mParent
		} else {
			mType, err := opts.Generator.ToCodeType(s)
			if err != nil {
				return models, fmt.Errorf("error converting type: %w", err)
			}

			add.Parent = mType
		}

		models = append(models, add)

	}

	return models, nil
}
