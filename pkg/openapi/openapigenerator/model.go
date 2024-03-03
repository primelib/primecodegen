package openapigenerator

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
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
	template.Operations = append(template.Operations, operations...)

	models, err := BuildComponentModels(ModelOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Models = append(template.Models, models...)

	enums, err := BuildEnums(ModelOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Enums = append(template.Enums, enums...)

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
				OperationId: gen.ToFunctionName(op.Value.OperationId),
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
					return operations, fmt.Errorf("error converting type of [%s:%s:parameter:%s]: %w", path.Key, op.Key, param.Name, err)
				}

				allowedValues, err := openapidocument.EnumToAllowedValues(pSchema)
				if err != nil {
					return operations, fmt.Errorf("error processing enum definitions: %w", err)
				}
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:            gen.ToParameterName(param.Name),
					FieldName:       param.Name,
					In:              param.In,
					Description:     param.Description,
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

func BuildComponentModels(opts ModelOpts) ([]Model, error) {
	var models []Model
	gen := opts.Generator

	for schema := opts.Doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		s, err := schema.Value.BuildSchema()
		if err != nil {
			return models, fmt.Errorf("error building component schema: %w", err)
		}

		if openapidocument.IsEnumSchema(s) {
			continue
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
					return models, fmt.Errorf("error converting type of [%s:object:%s]: %w", schema.Key, p.Key, err)
				}

				allowedValues, err := openapidocument.EnumToAllowedValues(pSchema)
				if err != nil {
					return models, fmt.Errorf("error processing enum definitions: %w", err)
				}
				add.Properties = append(add.Properties, Property{
					Name:            gen.ToPropertyName(p.Key),
					FieldName:       p.Key,
					Description:     pSchema.Description,
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
				return models, fmt.Errorf("error converting type of [%s:array]: %w", schema.Key, err)
			}

			add.Parent = mParent
		} else {
			mType, err := gen.ToCodeType(s)
			if err != nil {
				return models, fmt.Errorf("error converting type of [%s]: %w", schema.Key, err)
			}

			add.Parent = mType
			add.Imports = append(add.Imports, gen.TypeToImport(add.Parent))
		}

		add.Imports = cleanImports(add.Imports)
		models = append(models, add)

	}

	return models, nil
}

func BuildEnums(opts ModelOpts) ([]Enum, error) {
	var enums []Enum
	gen := opts.Generator

	for schema := opts.Doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		s, err := schema.Value.BuildSchema()
		if err != nil {
			return enums, fmt.Errorf("error building component schema: %w", err)
		}

		if !openapidocument.IsEnumSchema(s) {
			continue
		}

		vType, err := gen.ToCodeType(s)
		if err != nil {
			return enums, fmt.Errorf("error converting type of [%s]: %w", schema.Key, err)
		}

		add := Enum{
			Name:          gen.ToClassName(s.Title),
			Description:   s.Description,
			ValueType:     vType,
			AllowedValues: make(map[string]openapidocument.AllowedValue),
		}
		allowedValues, err := openapidocument.EnumToAllowedValues(s)
		if err != nil {
			return enums, fmt.Errorf("error building enum definitions: %w", err)
		}
		for k, v := range allowedValues {
			v.Name = gen.ToPropertyName(v.Name)
			add.AllowedValues[k] = v
		}
		add.Imports = cleanImports(add.Imports)
		enums = append(enums, add)
	}

	return enums, nil
}
