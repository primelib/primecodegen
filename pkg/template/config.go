package template

var defaultSnippets = []string{"global-layout.gohtml"}

var allTemplates = map[string]Config{
	"openapi-go-httpclient": {
		ID:          "openapi-go-httpclient",
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
				TargetFileName:  "service-{{ .Service.Name }}.go",
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
	"openapi-java-httpclient": {
		ID:          "openapi-java-httpclient",
		Description: "OpenAPI Client for Java",
		Files: []File{
			// factory
			{
				SourceTemplate:  "api_factory.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Root | toFilePath }}",
				TargetFileName:  "{{ .Metadata.Name }}Factory.java",
				Type:            TypeAPIOnce,
			},
			{
				SourceTemplate:  "api_factoryspec.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Root | toFilePath }}",
				TargetFileName:  "{{ .Metadata.Name }}FactorySpec.java",
				Type:            TypeAPIOnce,
			},
			// api
			{
				SourceTemplate:  "api_main_default.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
				TargetFileName:  "{{ .Metadata.Name }}Api.java",
				Type:            TypeAPIOnce,
			},
			{
				SourceTemplate:  "api_main_consumer.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
				TargetFileName:  "{{ .Metadata.Name }}ConsumerApi.java",
				Type:            TypeAPIOnce,
			},
			// services
			{
				SourceTemplate:  "api_service_default.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
				TargetFileName:  "{{ .Service.Type }}Api.java",
				Type:            TypeAPIEach,
			},
			{
				SourceTemplate:  "api_service_consumer.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
				TargetFileName:  "{{ .Service.Type }}ConsumerApi.java",
				Type:            TypeAPIEach,
			},
			// operations
			{
				SourceTemplate:  "operation.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Operations | toFilePath }}",
				TargetFileName:  "{{ .Operation.Name }}OperationSpec.java",
				Type:            TypeOperationEach,
			},
			// model
			{
				SourceTemplate:  "model.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Models | toFilePath }}",
				TargetFileName:  "{{ .Name }}.java",
				Type:            TypeModelEach,
			},
			{
				SourceTemplate:  "enum.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "src/main/java/{{ .Common.Packages.Enums | toFilePath }}",
				TargetFileName:  "{{ .Name }}.java",
				Type:            TypeEnumEach,
			},
			// support files - gradle
			{
				SourceTemplate:  "build.gradle.kts.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "build.gradle.kts",
				Type:            TypeSupportOnce,
			},
			{
				SourceTemplate:  "settings.gradle.kts.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "settings.gradle.kts",
				Type:            TypeSupportOnce,
			},
			{
				SourceTemplate:  "libs.versions.toml.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "gradle",
				TargetFileName:  "libs.versions.toml",
				Type:            TypeSupportOnce,
			},
			{
				SourceTemplate:  "gradle-wrapper.properties.gohtml",
				Snippets:        defaultSnippets,
				TargetDirectory: "gradle/wrapper",
				TargetFileName:  "gradle-wrapper.properties",
				Type:            TypeSupportOnce,
			},
			{
				SourceUrl:       "https://github.com/PhilippHeuer/events4j/raw/main/gradle/wrapper/gradle-wrapper.jar",
				Snippets:        defaultSnippets,
				TargetDirectory: "gradle/wrapper",
				TargetFileName:  "gradle-wrapper.jar",
				Type:            TypeSupportOnce,
			},
			{
				SourceUrl:       "https://raw.githubusercontent.com/PhilippHeuer/events4j/main/gradlew",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "gradlew",
				Type:            TypeSupportOnce,
			},
			{
				SourceUrl:       "https://raw.githubusercontent.com/PhilippHeuer/events4j/main/gradlew.bat",
				Snippets:        defaultSnippets,
				TargetDirectory: "",
				TargetFileName:  "gradlew.bat",
				Type:            TypeSupportOnce,
			},
		},
	},
}
