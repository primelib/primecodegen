package template

var defaultSnippets = []string{"global-layout.gohtml"}

var allTemplates = []Config{
	{
		ID:          "openapi-go-client",
		Description: "OpenAPI Client for Go",
		Files: []File{
			{
				Description:     "client",
				SourceTemplate:  "client.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "client.go",
				Type:            TypeAPIOnce,
			},
			{
				Description:     "service per tag with all operations",
				SourceTemplate:  "service.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "service-{{ .TagName }}.go",
				Type:            TypeAPIEach,
			},
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "pkgs/models",
				TargetFileName:  "{{ .Name }}.go",
				Type:            TypeModelEach,
			},
			{
				Description:     "model file",
				SourceTemplate:  "enum.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "pkgs/models",
				TargetFileName:  "{{ .Name }}.go",
				Type:            TypeEnumEach,
			},
			{
				Description:     "operation",
				SourceTemplate:  "operation.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "pkgs/operations",
				TargetFileName:  "{{ .Name }}.go",
				Type:            TypeOperationEach,
			},
			// support files
			{
				Description:    "go.mod",
				SourceTemplate: "gomod.gohtml",
				Snippets:       defaultSnippets,
				TargetFileName: "go.mod",
				Type:           TypeSupportOnce,
			},
			{
				Description:    "go.sum",
				SourceTemplate: "gosum.gohtml",
				Snippets:       defaultSnippets,
				TargetFileName: "go.sum",
				Type:           TypeSupportOnce,
			},
		},
	},
}
