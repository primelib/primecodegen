package openapigenerator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
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
			Name:          tag.Name,
			Type:          generator.ToClassName(template.Name + tag.Name),
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
					operation.ReturnType = gen.PostProcessType(responseType)

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

					break
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
			Name:             gen.ToClassName(s.Title),
			Description:      s.Description,
			Deprecated:       getBoolValue(s.Deprecated, false),
			DeprecatedReason: getOrDefault(s.Extensions, "x-deprecated", ""),
		}

		// --- oneOf: union / sum type ---
		if len(s.OneOf) > 0 && !openapidocument.IsEnumSchema(s) {
			add.IsOneOf = true
			for _, sp := range s.OneOf {
				oneOfSchema, schemaErr := sp.BuildSchema()
				if schemaErr != nil || oneOfSchema == nil {
					continue
				}
				codeType, ctErr := gen.ToCodeType(oneOfSchema, CodeTypeSchemaProperty, false)
				if ctErr != nil {
					continue
				}
				codeType = gen.PostProcessType(codeType)
				add.OneOf = append(add.OneOf, codeType)
				add.Imports = append(add.Imports, gen.TypeToImport(codeType))
			}
			if s.Discriminator != nil {
				add.Discriminator = buildDiscriminatorModel(s.Discriminator, gen, opts.Doc.Model.Components.Schemas)
			}
		}

		// --- anyOf: union type ---
		if len(s.AnyOf) > 0 {
			add.IsAnyOf = true
			for _, sp := range s.AnyOf {
				anyOfSchema, schemaErr := sp.BuildSchema()
				if schemaErr != nil || anyOfSchema == nil {
					continue
				}
				codeType, ctErr := gen.ToCodeType(anyOfSchema, CodeTypeSchemaProperty, false)
				if ctErr != nil {
					continue
				}
				codeType = gen.PostProcessType(codeType)
				add.AnyOf = append(add.AnyOf, codeType)
				add.Imports = append(add.Imports, gen.TypeToImport(codeType))
			}
			if s.Discriminator != nil && add.Discriminator == nil {
				add.Discriminator = buildDiscriminatorModel(s.Discriminator, gen, opts.Doc.Model.Components.Schemas)
			}
		}

		// --- allOf: inheritance / composition ---
		if len(s.AllOf) > 0 {
			add.IsAllOf = true
			// collect required fields from both the root schema and any inline allOf sub-schemas
			allRequired := append([]string{}, s.Required...)
			for _, sp := range s.AllOf {
				if !sp.IsReference() {
					if inlineS := sp.Schema(); inlineS != nil {
						allRequired = append(allRequired, inlineS.Required...)
					}
				}
			}

			for _, sp := range s.AllOf {
				allOfSchema, schemaErr := sp.BuildSchema()
				if schemaErr != nil || allOfSchema == nil {
					continue
				}
				if sp.IsReference() {
					// referenced schema → record as a parent/interface type
					codeType, ctErr := gen.ToCodeType(allOfSchema, CodeTypeSchemaParent, false)
					if ctErr != nil {
						continue
					}
					codeType = gen.PostProcessType(codeType)
					add.AllOf = append(add.AllOf, codeType)
					add.Imports = append(add.Imports, gen.TypeToImport(codeType))
				} else {
					// inline sub-schema → merge its properties directly
					if allOfSchema.Properties != nil {
						props, imports, propErr := buildModelProperties(allOfSchema, gen, allRequired, add.Properties)
						if propErr != nil {
							return models, propErr
						}
						add.Properties = append(add.Properties, props...)
						add.Imports = append(add.Imports, imports...)
					}
				}
			}
		}

		// --- object: direct properties ---
		if slices.Contains(s.Type, "object") && s.Properties != nil {
			props, imports, propErr := buildModelProperties(s, gen, s.Required, add.Properties)
			if propErr != nil {
				return models, propErr
			}
			add.Properties = append(add.Properties, props...)
			add.Imports = append(add.Imports, imports...)
		} else if slices.Contains(s.Type, "array") {
			// array type alias
			mParent, ctErr := gen.ToCodeType(s, CodeTypeSchemaArray, false)
			if ctErr != nil {
				return models, fmt.Errorf("error converting type of [%s:array]: %w", schema.Key, ctErr)
			}
			add.Parent = gen.PostProcessType(mParent)
		} else if !add.IsOneOf && !add.IsAnyOf && !add.IsAllOf {
			// fallback: simple type alias or unrecognised type
			mParent, ctErr := gen.ToCodeType(s, CodeTypeSchemaParent, false)
			if ctErr != nil {
				return models, fmt.Errorf("error converting type of [%s]: %w", schema.Key, ctErr)
			}
			mParent = gen.PostProcessType(mParent)
			add.Parent = mParent
			add.Imports = append(add.Imports, gen.TypeToImport(add.Parent))
		}

		// mark as a type alias when the model has no properties and no polymorphic structure
		if len(add.Properties) == 0 && !add.IsOneOf && !add.IsAnyOf && !add.IsAllOf &&
			(add.Parent.Name != "" || add.Parent.IsArray || add.Parent.IsList || add.Parent.IsMap) {
			add.IsTypeAlias = true
		}

		add.Imports = uniqueSortImports(add.Imports)
		models = append(models, add)
	}

	return models, nil
}

// buildModelProperties converts the properties of an OpenAPI schema into template Property values.
// existing is the list already accumulated so that duplicate property names are skipped.
func buildModelProperties(s *base.Schema, gen CodeGenerator, required []string, existing []Property) ([]Property, []string, error) {
	var properties []Property
	var imports []string

	if s.Properties == nil {
		return properties, imports, nil
	}

	for p := s.Properties.Oldest(); p != nil; p = p.Next() {
		propName := gen.ToPropertyName(p.Key)

		// skip duplicates already present in existing or already added in this call
		alreadyAdded := false
		for _, ep := range existing {
			if ep.Name == propName {
				alreadyAdded = true
				break
			}
		}
		if !alreadyAdded {
			for _, np := range properties {
				if np.Name == propName {
					alreadyAdded = true
					break
				}
			}
		}
		if alreadyAdded {
			continue
		}

		pSchema, pErr := p.Value.BuildSchema()
		if pErr != nil {
			return properties, imports, fmt.Errorf("error building property schema: %w", pErr)
		}

		pType, err := gen.ToCodeType(pSchema, CodeTypeSchemaProperty, false)
		if err != nil {
			return properties, imports, fmt.Errorf("error converting type of property [%s]: %w", p.Key, err)
		}
		pType = gen.PostProcessType(pType)

		allowedValues, err := openapidocument.EnumToAllowedValues(pSchema)
		if err != nil {
			return properties, imports, fmt.Errorf("error processing enum definitions: %w", err)
		}

		properties = append(properties, Property{
			Name:            propName,
			FieldName:       p.Key,
			Description:     pSchema.Description,
			Title:           pSchema.Title,
			Type:            pType,
			IsPrimitiveType: gen.IsPrimitiveType(pType.Name),
			Required:        slices.Contains(required, p.Key),
			Nullable:        openapiutil.IsSchemaNullable(pSchema),
			ReadOnly:        getBoolValue(pSchema.ReadOnly, false),
			WriteOnly:       getBoolValue(pSchema.WriteOnly, false),
			AllowedValues:   allowedValues,
		})
		imports = append(imports, gen.TypeToImport(pType))
	}

	return properties, imports, nil
}

// buildDiscriminatorModel converts a libopenapi Discriminator into a DiscriminatorModel.
// It resolves each mapping value (a $ref) to a CodeType using the component schemas.
func buildDiscriminatorModel(disc *base.Discriminator, gen CodeGenerator, schemas *orderedmap.Map[string, *base.SchemaProxy]) *DiscriminatorModel {
	dm := &DiscriminatorModel{
		PropertyName: disc.PropertyName,
		Mapping:      make(map[string]CodeType),
	}
	if disc.Mapping == nil {
		return dm
	}
	for entry := disc.Mapping.Oldest(); entry != nil; entry = entry.Next() {
		ref := entry.Value
		// extract schema key from "#/components/schemas/<Key>" or a bare name
		schemaKey := ref
		if idx := strings.LastIndex(ref, "/"); idx >= 0 {
			schemaKey = ref[idx+1:]
		}
		// resolve the component schema to a CodeType
		if sp, ok := schemas.Get(schemaKey); ok {
			if resolved, err := sp.BuildSchema(); err == nil && resolved != nil {
				codeType, ctErr := gen.ToCodeType(resolved, CodeTypeSchemaProperty, false)
				if ctErr == nil {
					dm.Mapping[entry.Key] = gen.PostProcessType(codeType)
					continue
				}
			}
		}
		// fallback: construct a plain CodeType from the schema key name
		dm.Mapping[entry.Key] = gen.PostProcessType(CodeType{Name: gen.ToClassName(schemaKey)})
	}
	return dm
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
