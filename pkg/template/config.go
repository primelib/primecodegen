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
				Scope:           ScopeModel,
				Iterator:        IteratorEachModel,
			},
			// support files
			{
				Description:    "go.mod",
				SourceTemplate: "gomod.gohtml",
				Snippets:       defaultSnippets,
				TargetFileName: "go.mod",
				Scope:          ScopeSupport,
				Iterator:       IteratorOnce,
			},
			{
				Description:    "go.sum",
				SourceTemplate: "gosum.gohtml",
				Snippets:       defaultSnippets,
				TargetFileName: "go.sum",
				Scope:          ScopeSupport,
				Iterator:       IteratorOnce,
			},
		},
	},
}
