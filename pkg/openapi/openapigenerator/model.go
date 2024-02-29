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
	gen := opts.Generator

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

				pType, err := gen.ToCodeType(param.Schema.Schema())
				if err != nil {
					return operations, fmt.Errorf("error converting type: %w", err)
				}

				pKind := KindVar
				if len(pSchema.Enum) > 0 {
					pKind = KindEnum
				}
				allowedValues := make(map[string]AllowedValue)
				for _, e := range pSchema.Enum {
					allowedValues[e.Value] = AllowedValue{Value: e.Value}
				}
				allowedValues, err = extensionEnumDefinitions(pSchema, allowedValues)
				if err != nil {
					return operations, fmt.Errorf("error processing enum definitions: %w", err)
				}
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:            gen.ToParameterName(param.Name),
					FieldName:       param.Name,
					In:              param.In,
					Description:     param.Description,
					Kind:            pKind,
					Type:            pType,
					IsPrimitiveType: gen.IsPrimitiveType(pType),
					AllowedValues:   allowedValues,
					Required:        getBoolValue(param.Required, false),
					Deprecated:      param.Deprecated,
					// DeprecatedReason: param.Value.Extensions.Get("x-deprecated"),
				})
				operation.Imports = append(operation.Imports, gen.TypeToImport(pType))
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

			operation.Imports = cleanImports(operation.Imports)
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
	gen := opts.Generator

	for schema := opts.Doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		s, err := schema.Value.BuildSchema()
		if err != nil {
			return models, fmt.Errorf("error building component schema: %w", err)
		}

		add := Model{
			Name:        gen.ToClassName(s.Title),
			Description: s.Description,
		}
		if slices.Contains(s.Type, "object") && s.Properties != nil {
			for p := s.Properties.Oldest(); p != nil; p = p.Next() {
				pSchema, pErr := p.Value.BuildSchema()
				if pErr != nil {
					return models, fmt.Errorf("error building property schema: %w", err)
				}

				pType, err := gen.ToCodeType(pSchema)
				if err != nil {
					return models, fmt.Errorf("error converting type: %w", err)
				}

				pKind := KindVar
				if len(pSchema.Enum) > 0 {
					pKind = KindEnum
				}
				allowedValues := make(map[string]AllowedValue)
				for _, e := range pSchema.Enum {
					allowedValues[e.Value] = AllowedValue{Value: e.Value}
				}
				allowedValues, err = extensionEnumDefinitions(pSchema, allowedValues)
				if err != nil {
					return models, fmt.Errorf("error processing enum definitions: %w", err)
				}
				add.Properties = append(add.Properties, Property{
					Name:            gen.ToPropertyName(p.Key),
					FieldName:       p.Key,
					Description:     pSchema.Description,
					Kind:            pKind,
					Title:           pSchema.Title,
					Type:            pType,
					IsPrimitiveType: gen.IsPrimitiveType(pType),
					Nullable:        getBoolValue(pSchema.Nullable, slices.Contains(pSchema.Type, "null")), // 3.1 uses null type, 3.0 uses nullable
					AllowedValues:   allowedValues,
				})
				add.Imports = append(add.Imports, gen.TypeToImport(pType))
			}
		} else if slices.Contains(s.Type, "array") {
			mParent, err := gen.ToCodeType(s)
			if err != nil {
				return models, fmt.Errorf("error converting type: %w", err)
			}

			add.Parent = mParent
		} else {
			mType, err := gen.ToCodeType(s)
			if err != nil {
				return models, fmt.Errorf("error converting type: %w", err)
			}

			add.Parent = mType
			add.Imports = append(add.Imports, gen.TypeToImport(add.Parent))
		}

		add.Imports = cleanImports(add.Imports)
		models = append(models, add)

	}

	return models, nil
}
