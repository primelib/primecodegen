package openapi_default_scaffolding

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
)

var Template = templateapi.Config{
	ID:          "openapi-default-scaffolding",
	Description: "Scaffolding Project Files",
	Files: []templateapi.File{
		{
			Description:    "README.md",
			SourceTemplate: "readme.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "README.md",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
		{
			Description:    "LICENSE",
			SourceTemplate: "license.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "LICENSE",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
		{
			Description:    "gitignore",
			SourceTemplate: "gitignore.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: ".gitignore",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
		{
			Description:    "justfile",
			SourceTemplate: "justfile.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "justfile",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
	},
}
