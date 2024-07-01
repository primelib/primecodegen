package openapigenerator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
)

func BuildTemplateData(doc *libopenapi.DocumentModel[v3.Document], generator CodeGenerator, packageConfig CommonPackages) (DocumentModel, error) {
	title := doc.Model.Info.Title
	title = strings.TrimSuffix(title, "API")
	var template = DocumentModel{
		Name:        generator.ToClassName(title),
		DisplayName: title,
		Description: doc.Model.Info.Description,
		Auth:        BuildAuth(doc),
		Packages:    packageConfig,
	}

	// all operations
	operations, err := BuildOperations(OperationOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Operations = append(template.Operations, operations...)

	// operations by tag
	template.OperationsByTag = make(map[string][]Operation)
	for _, op := range operations {
		template.OperationsByTag[op.Tag] = append(template.OperationsByTag[op.Tag], op)
	}

	// services
	template.Services = make(map[string]Service)
	for _, tag := range doc.Model.Tags {
		service := Service{
			Name:        tag.Name,
			Description: tag.Description,
			Operations:  []Operation{},
		}
		if _, ok := template.OperationsByTag[tag.Name]; ok {
			service.Operations = append(service.Operations, template.OperationsByTag[tag.Name]...)
		}

		template.Services[tag.Name] = service
	}

	// models
	models, err := BuildComponentModels(ModelOpts{
		Generator: generator,
		Doc:       doc,
	})
	if err != nil {
		return template, err
	}
	template.Models = append(template.Models, models...)

	// enums
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

func BuildAuth(doc *libopenapi.DocumentModel[v3.Document]) Auth {
	var auth Auth

	for security := doc.Model.Components.SecuritySchemes.Oldest(); security != nil; security = security.Next() {
		auth.Methods = append(auth.Methods, AuthMethod{
			Name:   security.Key,
			Scheme: strings.ToLower(security.Value.Scheme),
		})
	}

	return auth
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
				Name:             gen.ToClassName(op.Value.OperationId),
				Path:             path.Key,
				Method:           op.Key,
				Summary:          op.Value.Summary,
				Description:      op.Value.Description,
				Tag:              "default",
				Tags:             op.Value.Tags,
				ReturnType:       CodeType{},
				Deprecated:       getBoolValue(op.Value.Deprecated, false),
				DeprecatedReason: getOrDefault(op.Value.Extensions, "x-deprecated", ""),
				Documentation:    make([]Documentation, 0),
				Stability:        getOrDefault(op.Value.Extensions, "x-stability", "stable"),
			}
			if len(op.Value.Tags) > 0 {
				operation.Tag = op.Value.Tags[0]
			}

			// external docs
			if op.Value.ExternalDocs != nil {
				operation.Documentation = append(operation.Documentation, Documentation{
					Title: op.Value.ExternalDocs.Description,
					URL:   op.Value.ExternalDocs.URL,
				})
			}

			// operation parameters
			for _, param := range op.Value.Parameters {
				pSchema, err := param.Schema.BuildSchema()
				if err != nil {
					return operations, fmt.Errorf("error building property schema: %w", err)
				}

				pType, err := gen.ToCodeType(pSchema, CodeTypeSchemaParameter, ptr.Value(param.Required))
				if err != nil {
					return operations, fmt.Errorf("error converting type of [%s:%s:parameter:%s]: %w", path.Key, op.Key, param.Name, err)
				}
				pType = gen.PostProcessType(pType)

				allowedValues, err := openapidocument.EnumToAllowedValues(pSchema)
				if err != nil {
					return operations, fmt.Errorf("error processing enum definitions: %w", err)
				}
				p := Parameter{
					Name:            gen.ToParameterName(param.Name),
					FieldName:       param.Name,
					In:              param.In,
					Description:     param.Description,
					Type:            pType,
					IsPrimitiveType: gen.IsPrimitiveType(pType.Name),
					AllowedValues:   allowedValues,
					Required:        getBoolValue(param.Required, false),
					Deprecated:      param.Deprecated,
					// DeprecatedReason: param.Value.Extensions.Get("x-deprecated"),
				}
				operation.Parameters = append(operation.Parameters, p)
				if p.In == "path" {
					operation.PathParameters = append(operation.PathParameters, p)
				} else if p.In == "query" {
					operation.QueryParameters = append(operation.QueryParameters, p)
				} else if p.In == "header" {
					operation.HeaderParameters = append(operation.HeaderParameters, p)
				} else if p.In == "cookie" {
					operation.CookieParameters = append(operation.CookieParameters, p)
				}
				operation.Imports = append(operation.Imports, gen.TypeToImport(pType))
			}

			// request body
			if rb := op.Value.RequestBody; rb != nil {
				// TODO: set correct type for request body
				payloadType := rb.Content.First().Value().Schema.Schema().Title

				bodyParam := Parameter{
					Name:        "payload",
					In:          "body",
					Description: rb.Description,
					Type:        gen.PostProcessType(CodeType{Name: payloadType}),
					Required:    true,
				}
				operation.BodyParameter = &bodyParam
				operation.Parameters = append(operation.Parameters, bodyParam)
			}

			// response type
			for resp := op.Value.Responses.Codes.Oldest(); resp != nil; resp = resp.Next() {
				if resp.Value.Content == nil {
					continue
				}

				if resp.Value.Content.First() == nil {
					continue
				}

				if resp.Key == "200" || resp.Key == "201" {
					responseType, err := gen.ToCodeType(resp.Value.Content.First().Value().Schema.Schema(), CodeTypeSchemaResponse, false)
					if err != nil {
						return operations, fmt.Errorf("error converting type of [%s:%s:responseType:%s]: %w", path.Key, op.Key, resp.Key, err)
					}
					operation.ReturnType = gen.PostProcessType(responseType)
					break
				}
			}

			operation.Imports = uniqueSortImports(operation.Imports)
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

				pType, err := gen.ToCodeType(pSchema, CodeTypeSchemaProperty, false)
				if err != nil {
					return models, fmt.Errorf("error converting type of [%s:object:%s]: %w", schema.Key, p.Key, err)
				}
				pType = gen.PostProcessType(pType)

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
					IsPrimitiveType: gen.IsPrimitiveType(pType.Name),
					Nullable:        getBoolValue(pSchema.Nullable, slices.Contains(pSchema.Type, "null")), // 3.1 uses null type, 3.0 uses nullable
					AllowedValues:   allowedValues,
				})

				add.Imports = append(add.Imports, gen.TypeToImport(pType))
			}
		} else if slices.Contains(s.Type, "array") {
			mParent, err := gen.ToCodeType(s, CodeTypeSchemaArray, false)
			if err != nil {
				return models, fmt.Errorf("error converting type of [%s:array]: %w", schema.Key, err)
			}
			mParent = gen.PostProcessType(mParent)

			add.Parent = mParent
		} else {
			mParent, err := gen.ToCodeType(s, CodeTypeSchemaParent, false)
			if err != nil {
				return models, fmt.Errorf("error converting type of [%s]: %w", schema.Key, err)
			}
			mParent = gen.PostProcessType(mParent)

			add.Parent = mParent
			add.Imports = append(add.Imports, gen.TypeToImport(add.Parent))
		}
		if len(add.Properties) == 0 && add.Parent.Name != "" {
			add.IsTypeAlias = true
		}

		add.Imports = uniqueSortImports(add.Imports)
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

		vType, err := gen.ToCodeType(s, CodeTypeSchemaProperty, true)
		if err != nil {
			return enums, fmt.Errorf("error converting type of [%s]: %w", schema.Key, err)
		}
		vType = gen.PostProcessType(vType)

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
		add.Imports = uniqueSortImports(add.Imports)
		enums = append(enums, add)
	}

	return enums, nil
}

// PruneTypeAliases removes type aliases and replaces it with the actual type
// A type alias is identified by not having properties and the parent being a primitive type
func PruneTypeAliases(documentModel DocumentModel, primitiveTypes []string) DocumentModel {
	var typeAliasModels []Model
	for _, model := range documentModel.Models {
		if model.IsTypeAlias {
			typeAliasModels = append(typeAliasModels, model)
		}
	}

	// fix types (replace type alias with primitive type)
	for i, model := range documentModel.Models {
		for j, property := range model.Properties {
			for _, typeAliasModel := range typeAliasModels {
				if property.Type.Name == typeAliasModel.Name {
					documentModel.Models[i].Properties[j].Type = typeAliasModel.Parent
					break
				}
			}
		}
	}
	for i, op := range documentModel.Operations {
		for j, param := range op.Parameters {
			for _, typeAliasModel := range typeAliasModels {
				if param.Type.Name == typeAliasModel.Name {
					documentModel.Operations[i].Parameters[j].Type = typeAliasModel.Parent
					break
				}
			}
		}

		for _, typeAliasModel := range typeAliasModels {
			if op.ReturnType.Name == typeAliasModel.Name {
				documentModel.Operations[i].ReturnType = typeAliasModel.Parent
				break
			}
		}
	}

	// remove type alias models
	for _, typeAliasModel := range typeAliasModels {
		for i, model := range documentModel.Models {
			if strings.ToLower(model.Name) == strings.ToLower(typeAliasModel.Name) {
				documentModel.Models = append(documentModel.Models[:i], documentModel.Models[i+1:]...)
			}
		}
	}

	return documentModel
}
