package openapigenerator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapiutil"
)

func BuildTemplateData(doc *libopenapi.DocumentModel[v3.Document], generator CodeGenerator, packageConfig CommonPackages) (DocumentModel, error) {
	title := doc.Model.Info.Title
	title = strings.TrimSuffix(title, "API")
	var template = DocumentModel{
		Name:        generator.ToClassName(title),
		DisplayName: title,
		Description: doc.Model.Info.Description,
		Endpoints:   BuildEndpoints(doc),
		Auth:        BuildAuth(doc),
		Packages:    packageConfig,
	}

	// all operations
	operations, err := BuildOperations(OperationOpts{
		Generator:     generator,
		Doc:           doc,
		PackageConfig: packageConfig,
	})
	if err != nil {
		return template, err
	}
	template.Operations = append(template.Operations, operations...)

	// operations by tag
	template.OperationsByTag = make(map[string][]Operation)
	for _, op := range operations {
		for _, tag := range op.Tags {
			template.OperationsByTag[tag] = append(template.OperationsByTag[tag], op)
		}
	}

	// services
	template.Services = make(map[string]Service)
	for _, tag := range doc.Model.Tags {
		service := Service{
			Name:        tag.Name,
			Type:        generator.ToClassName(template.Name + tag.Name),
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
		Generator:     generator,
		Doc:           doc,
		PackageConfig: packageConfig,
	})
	if err != nil {
		return template, err
	}
	template.Models = append(template.Models, models...)

	// enums
	enums, err := BuildEnums(ModelOpts{
		Generator:     generator,
		Doc:           doc,
		PackageConfig: packageConfig,
	})
	if err != nil {
		return template, err
	}
	template.Enums = append(template.Enums, enums...)

	return template, nil
}

func BuildEndpoints(doc *libopenapi.DocumentModel[v3.Document]) Endpoints {
	var endpoints Endpoints

	for _, server := range doc.Model.Servers {
		endpoint := Endpoint{
			Type:        "http",
			URL:         strings.TrimSuffix(server.URL, "/"),
			Description: server.Description,
		}
		if strings.HasPrefix(server.URL, "unix://") {
			endpoint.Type = "socket"
		}

		// TODO: add support for variables
		for epv := server.Variables.Oldest(); epv != nil; epv = epv.Next() {
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
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
	Generator     CodeGenerator
	Doc           *libopenapi.DocumentModel[v3.Document]
	PackageConfig CommonPackages
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
				ReturnType:       gen.PostProcessType(VoidCodeType),
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
			var addedParameters []string
			for _, param := range op.Value.Parameters {
				if slices.Contains(addedParameters, gen.ToParameterName(param.Name)) {
					continue
				}

				pSchema, err := param.Schema.BuildSchema()
				if err != nil {
					return operations, fmt.Errorf("error building property schema: %w", err)
				}

				pType, err := gen.ToCodeType(pSchema, CodeTypeSchemaParameter, ptr.Value(param.Required))
				if err != nil {
					return operations, fmt.Errorf("error converting type of [%s:%s:parameter:%s]: %w", path.Key, op.Key, param.Name, err)
				}
				pType = gen.PostProcessType(pType)

				deprecatedReason := ""
				deprecatedReasonNode := param.Extensions.GetOrZero("x-deprecated")
				if deprecatedReasonNode != nil {
					deprecatedReason = deprecatedReasonNode.Value
				}

				allowedValues, err := openapidocument.EnumToAllowedValues(pSchema)
				if err != nil {
					return operations, fmt.Errorf("error processing enum definitions: %w", err)
				}
				p := Parameter{
					Name:             gen.ToParameterName(param.Name),
					FieldName:        param.Name,
					In:               param.In,
					Description:      param.Description,
					Type:             pType,
					IsPrimitiveType:  gen.IsPrimitiveType(pType.Name),
					AllowedValues:    allowedValues,
					Required:         getBoolValue(param.Required, false),
					Deprecated:       param.Deprecated,
					DeprecatedReason: deprecatedReason,
				}
				operation.AddParameter(p)
				operation.Imports = append(operation.Imports, gen.TypeToImport(pType))

				addedParameters = append(addedParameters, gen.ToParameterName(param.Name))
			}

			// request body
			if rb := op.Value.RequestBody; rb != nil {
				requestBody := rb.Content.First()

				// content-type header
				contentType, err := gen.ToCodeType(&base.Schema{Type: []string{"string"}, Format: ""}, CodeTypeSchemaParameter, true)
				if err != nil {
					return operations, fmt.Errorf("error converting type of [%s:%s:contentType]: %w", path.Key, op.Key, err)
				}
				headerParam := Parameter{
					Name:        gen.ToParameterName("Content-Type"),
					FieldName:   "Content-Type",
					In:          "header",
					Type:        contentType,
					Required:    true,
					StaticValue: requestBody.Key(),
				}
				if !slices.Contains(addedParameters, gen.ToParameterName(headerParam.Name)) {
					operation.AddParameter(headerParam)
				}

				// body type
				bodyType, err := gen.ToCodeType(requestBody.Value().Schema.Schema(), CodeTypeSchemaResponse, false)
				if err != nil {
					return operations, fmt.Errorf("error converting type of [%s:%s:bodyType]: %w", path.Key, op.Key, err)
				}
				bodyParam := Parameter{
					Name:        "payload",
					In:          "body",
					Description: rb.Description,
					Type:        gen.PostProcessType(bodyType),
					Required:    true,
				}
				operation.AddParameter(bodyParam)
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
	Generator     CodeGenerator
	Doc           *libopenapi.DocumentModel[v3.Document]
	PackageConfig CommonPackages
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
			var addedProperties []string

			for p := s.Properties.Oldest(); p != nil; p = p.Next() {
				if slices.Contains(addedProperties, gen.ToPropertyName(p.Key)) {
					continue
				}

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
					Nullable:        openapiutil.IsSchemaNullable(pSchema),
					AllowedValues:   allowedValues,
				})
				add.Imports = append(add.Imports, gen.TypeToImport(pType))

				addedProperties = append(addedProperties, gen.ToPropertyName(p.Key))
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
		if len(add.Properties) == 0 && (add.Parent.Name != "" || add.Parent.IsArray || add.Parent.IsList || add.Parent.IsMap) {
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
			Name:             gen.ToClassName(s.Title),
			Description:      s.Description,
			ValueType:        vType,
			AllowedValues:    make(map[string]openapidocument.AllowedValue),
			Deprecated:       getBoolValue(s.Deprecated, false),
			DeprecatedReason: getOrDefault(s.Extensions, "x-deprecated", ""),
		}
		allowedValues, err := openapidocument.EnumToAllowedValues(s)
		if err != nil {
			return enums, fmt.Errorf("error building enum definitions: %w", err)
		}
		for k, v := range allowedValues {
			v.Name = gen.ToConstantName(v.Name)
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
