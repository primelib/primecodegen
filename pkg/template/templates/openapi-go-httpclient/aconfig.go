package openapi_go_httpclient

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
)

var Template = templateapi.Config{
	ID:          "openapi-go-httpclient",
	Description: "OpenAPI Client for Go",
	Files: []templateapi.File{
		{
			Description:     "client",
			SourceTemplate:  "client.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "",
			TargetFileName:  "client.go",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			Description:     "service per tag with all operations",
			SourceTemplate:  "service.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "",
			TargetFileName:  "service-{{ .Service.Name }}.go",
			Type:            templateapi.TypeAPIEach,
			Kind:            templateapi.KindAPI,
		},
		{
			Description:     "operation",
			SourceTemplate:  "operation.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "pkgs/operations",
			TargetFileName:  "{{ .Name }}.go",
			Type:            templateapi.TypeOperationEach,
			Kind:            templateapi.KindAPI,
		},
		// models
		{
			Description:     "model file",
			SourceTemplate:  "model.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "pkgs/models",
			TargetFileName:  "{{ .Name }}.go",
			Type:            templateapi.TypeModelEach,
			Kind:            templateapi.KindModel,
		},
		{
			Description:     "model file",
			SourceTemplate:  "enum.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "pkgs/enums",
			TargetFileName:  "{{ .Name }}.go",
			Type:            templateapi.TypeEnumEach,
			Kind:            templateapi.KindModel,
		},
		// support files´- docs
		{
			Description:    "README.md",
			SourceTemplate: "readme.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "README.md",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
		// support files - go.mod
		{
			Description:    "go.mod",
			SourceTemplate: "gomod.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "go.mod",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindBuildSystem,
		},
		{
			Description:    "go.sum",
			SourceTemplate: "gosum.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "go.sum",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindBuildSystem,
		},
	},
}
