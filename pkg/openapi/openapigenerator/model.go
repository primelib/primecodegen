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
	"github.com/primelib/primecodegen/pkg/util"
)

func BuildTemplateData(doc *libopenapi.DocumentModel[v3.Document], generator CodeGenerator, packageConfig CommonPackages) (DocumentModel, error) {
	specNameNode, _ := doc.Model.Info.Extensions.Get("x-name")
	if specNameNode == nil {
		return DocumentModel{}, fmt.Errorf("document is missing required x-name extension in info section")
	}

	specName := specNameNode.Value
	specName = strings.TrimSuffix(specName, "API")

	specTitle := doc.Model.Info.Title
	specTitle = strings.TrimSuffix(specTitle, "API")
	var template = DocumentModel{
		Name:             generator.ToClassName(specName),
		DisplayName:      specName,
		Title:            specTitle,
		Description:      doc.Model.Info.Description,
		APISpecVersion:   doc.Model.Info.Version,
		GeneratorVersion: "1.0.0", // TODO: introduce version constants
		Endpoints:        BuildEndpoints(doc),
		Auth:             BuildAuth(doc),
		Packages:         packageConfig,
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
			Name:          tag.Name,
			Type:          generator.ToClassName(template.Name + util.UpperCaseFirstLetter(tag.Name)),
			Description:   tag.Description,
			Operations:    []Operation{},
			Documentation: make([]Documentation, 0),
		}
		if _, ok := template.OperationsByTag[tag.Name]; ok {
			service.Operations = append(service.Operations, template.OperationsByTag[tag.Name]...)
		}

		if tag.ExternalDocs != nil {
			service.Documentation = append(service.Documentation, Documentation{
				Title: tag.ExternalDocs.Description,
				URL:   tag.ExternalDocs.URL,
			})
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

func BuildAuth(doc *libopenapi.DocumentModel[v3.Document]) (auth Auth) {
	if doc.Model.Components == nil || doc.Model.Components.SecuritySchemes == nil {
		return auth
	}

	for security := doc.Model.Components.SecuritySchemes.Oldest(); security != nil; security = security.Next() {
		securityValue := security.Value

		authMethodType := strings.ToLower(securityValue.Type)
		authMethod := AuthMethod{
			Name:        security.Key,
			Type:        authMethodType,
			Scheme:      strings.ToLower(securityValue.Scheme),
			Description: securityValue.Description,
		}

		switch authMethodType {
		case "apikey":
			if securityValue.In == "header" {
				authMethod.Variant = "apiKeyHeaderAuth"
				authMethod.HeaderParam = securityValue.Name
			} else if securityValue.In == "query" {
				authMethod.Variant = "apiKeyQueryAuth"
				authMethod.QueryParam = securityValue.Name
			}
		case "http":
			if securityValue.Scheme == "basic" {
				authMethod.Variant = "basicAuth"
			} else if securityValue.Scheme == "bearer" {
				authMethod.Variant = "bearerAuth"

			}
		case "oauth2":
			if securityValue.Flows != nil {
				if securityValue.Flows.ClientCredentials != nil {
					authMethod.Variant = "oauth2ClientCredentialAuth"
					authMethod.TokenUrl = securityValue.Flows.ClientCredentials.TokenUrl
				} else if securityValue.Flows.Password != nil {
					authMethod.Variant = "oauth2PasswordAuth"
					authMethod.TokenUrl = securityValue.Flows.Password.TokenUrl
				} else if securityValue.Flows.AuthorizationCode != nil {
					authMethod.Variant = "oauth2AuthorizationCodeAuth"
					authMethod.TokenUrl = securityValue.Flows.AuthorizationCode.TokenUrl
				} else if securityValue.Flows.Implicit != nil {
					authMethod.Variant = "oauth2ImplicitAuth"
					authMethod.TokenUrl = securityValue.Flows.Implicit.AuthorizationUrl
				}
			}
		}

		auth.Methods = append(auth.Methods, authMethod)
	}

	return auth
}

type OperationOpts struct {
	Generator     CodeGenerator
	Doc           *libopenapi.DocumentModel[v3.Document]
	PackageConfig CommonPackages
}

func BuildOperations(opts OperationOpts) ([]Operation, error) {
	if opts.Doc.Model.Paths == nil {
		return nil, nil
	}

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
			allParams := append(path.Value.Parameters, op.Value.Parameters...)
			for _, param := range allParams {
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

				explodeDelimiter, _ := delimiterFromStyle(param.Style)
				p := Parameter{
					Name:             gen.ToParameterName(param.Name),
					FieldName:        param.Name,
					In:               param.In,
					Description:      param.Description,
					Type:             pType,
					IsPrimitiveType:  gen.IsPrimitiveType(pType.Name),
					Explode:          getBoolValue(param.Explode, true),
					ExplodeDelimiter: explodeDelimiter,
					AllowedValues:    allowedValues,
					Required:         getBoolValue(param.Required, false),
					Deprecated:       param.Deprecated,
					DeprecatedReason: deprecatedReason,
					Stability:        getOrDefault(param.Extensions, "x-stability", "stable"),
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
			operation.ReturnTypeByCode = make(map[string]*CodeType)
			for resp := op.Value.Responses.Codes.Oldest(); resp != nil; resp = resp.Next() {
				if resp.Value.Content == nil {
					continue
				}
				if resp.Value.Content.First() == nil {
					continue
				}

				if resp.Key == "200" || resp.Key == "201" {
					respContent := resp.Value.Content.First()

					// return type
					responseType, err := gen.ToCodeType(respContent.Value().Schema.Schema(), CodeTypeSchemaResponse, false)
					if err != nil {
						return operations, fmt.Errorf("error converting type of [%s:%s:responseType:%s]: %w", path.Key, op.Key, resp.Key, err)
					}
					processedResponseType := gen.PostProcessType(responseType)
					operation.ReturnType = processedResponseType
					operation.ReturnTypeByCode[resp.Key] = &processedResponseType

					// accept header as static parameter
					mediaType := respContent.Key() // e.g. "application/json"
					if mediaType != "" {
						acceptType, err := gen.ToCodeType(&base.Schema{Type: []string{"string"}, Format: ""}, CodeTypeSchemaParameter, true)
						if err != nil {
							return operations, fmt.Errorf("error converting type of [%s:%s:acceptHeader]: %w", path.Key, op.Key, err)
						}

						headerParam := Parameter{
							Name:        gen.ToParameterName("Accept"),
							FieldName:   "Accept",
							In:          "header",
							Type:        acceptType,
							Required:    true,
							StaticValue: mediaType,
						}

						if !slices.Contains(addedParameters, gen.ToParameterName(headerParam.Name)) {
							operation.AddParameter(headerParam)
						}
					}
				} else {
					respContent := resp.Value.Content.First()

					// return type
					responseType, err := gen.ToCodeType(respContent.Value().Schema.Schema(), CodeTypeSchemaResponse, false)
					if err != nil {
						return operations, fmt.Errorf("error converting type of [%s:%s:responseType:%s]: %w", path.Key, op.Key, resp.Key, err)
					}
					processedResponseType := gen.PostProcessType(responseType)
					operation.ReturnTypeByCode[resp.Key] = &processedResponseType
				}
			}

			operation.PathSegments = BuildPathSegments(path.Key, operation.PathParameters)
			operation.Imports = uniqueSortImports(operation.Imports)
			operation.Extensions = op.Value.Extensions
			operations = append(operations, operation)
		}
	}

	return operations, nil
}

func BuildPathSegments(path string, pathParameters []Parameter) []PathSegment {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	var pathSegments []PathSegment

	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			parameterName := ""

			for _, p := range pathParameters {
				if p.FieldName == name {
					parameterName = p.Name
					break
				}
			}

			pathSegments = append(pathSegments, PathSegment{
				Value:         segment,
				IsParameter:   true,
				ParameterName: parameterName,
			})
		} else {
			pathSegments = append(pathSegments, PathSegment{
				Value:       segment,
				IsParameter: false,
			})
		}
	}

	return pathSegments
}

type ModelOpts struct {
	Generator     CodeGenerator
	Doc           *libopenapi.DocumentModel[v3.Document]
	PackageConfig CommonPackages
}

func BuildComponentModels(opts ModelOpts) ([]Model, error) {
	if opts.Doc.Model.Components == nil || opts.Doc.Model.Components.Schemas == nil {
		return nil, nil
	}

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
	if opts.Doc.Model.Components == nil || opts.Doc.Model.Components.Schemas == nil {
		return nil, nil
	}

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
			if v.Name == "" {
				v.Name = "None" // default name for empty enum value
			}
			v.Name = gen.ToConstantName(v.Name)
			add.AllowedValues[k] = v
		}
		add.Imports = uniqueSortImports(add.Imports)
		enums = append(enums, add)
	}

	return enums, nil
}

// PruneTypeAliases removes type aliases and replaces them with the actual type
// A type alias is identified by the absence of properties and the parent being a primitive type
func PruneTypeAliases(documentModel DocumentModel, primitiveTypes []string) DocumentModel {
	// build alias map
	aliasMap := make(map[string]CodeType)
	for _, model := range documentModel.Models {
		if model.IsTypeAlias {
			aliasMap[strings.ToLower(model.Name)] = model.Parent
		}
	}
	resolveType := func(t CodeType) CodeType {
		if resolved, ok := aliasMap[strings.ToLower(t.Name)]; ok {
			return resolved
		}
		return t
	}

	// model aliases
	for i := range documentModel.Models {
		for j := range documentModel.Models[i].Properties {
			documentModel.Models[i].Properties[j].Type =
				resolveType(documentModel.Models[i].Properties[j].Type)
		}
	}

	// operation aliases
	for i := range documentModel.Operations {
		for j := range documentModel.Operations[i].Parameters {
			documentModel.Operations[i].Parameters[j].Type =
				resolveType(documentModel.Operations[i].Parameters[j].Type)
		}

		documentModel.Operations[i].ReturnType =
			resolveType(documentModel.Operations[i].ReturnType)
	}

	// filter models
	filtered := make([]Model, 0, len(documentModel.Models))
	for _, model := range documentModel.Models {
		if _, isAlias := aliasMap[strings.ToLower(model.Name)]; !isAlias {
			filtered = append(filtered, model)
		}
	}
	documentModel.Models = filtered

	return documentModel
}
