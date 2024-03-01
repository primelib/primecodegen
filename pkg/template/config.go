package template

var defaultSnippets = []string{"global-layout.gohtml"}

var allTemplates = []Config{
	{
		ID:          "openapi-go-client",
		Description: "OpenAPI Client for Go",
		Files: []File{
			{
				Description:     "model file",
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "{{ .Name }}.go",
				Type:            TypeModelEach,
			},
			{
				Description:     "model file",
				SourceTemplate:  "enum.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "models",
				TargetFileName:  "{{ .Name }}.go",
				Type:            TypeEnumEach,
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
