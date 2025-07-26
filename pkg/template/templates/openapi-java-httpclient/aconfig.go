package openapi_java_httpclient

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
)

var Template = templateapi.Config{
	ID:          "openapi-java-httpclient",
	Description: "OpenAPI Client for Java",
	Files: []templateapi.File{
		// factory
		{
			SourceTemplate:  "api_factory.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Root | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}Factory.java",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "api_factoryspec.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Root | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}FactorySpec.java",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// api
		{
			SourceTemplate:  "api_main_default.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}Api.java",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "api_main_consumer.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}ConsumerApi.java",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// services
		{
			SourceTemplate:  "api_service_default.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Service.Type }}Api.java",
			Type:            templateapi.TypeAPIEach,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "api_service_consumer.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Service.Type }}ConsumerApi.java",
			Type:            templateapi.TypeAPIEach,
			Kind:            templateapi.KindAPI,
		},
		// operations
		{
			SourceTemplate:  "operation.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Operations | toFilePath }}",
			TargetFileName:  "{{ .Operation.Name }}OperationSpec.java",
			Type:            templateapi.TypeOperationEach,
			Kind:            templateapi.KindAPI,
		},
		// model
		{
			SourceTemplate:  "model.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Models | toFilePath }}",
			TargetFileName:  "{{ .Name }}.java",
			Type:            templateapi.TypeModelEach,
			Kind:            templateapi.KindModel,
		},
		{
			SourceTemplate:  "enum.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "src/main/java/{{ .Common.Packages.Enums | toFilePath }}",
			TargetFileName:  "{{ .Name }}.java",
			Type:            templateapi.TypeEnumEach,
			Kind:            templateapi.KindModel,
		},
		// support files - docs
		{
			Description:    "README.md",
			SourceTemplate: "readme.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "README.md",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindDocumentation,
		},
		// support files - gradle
		{
			SourceTemplate: "build.gradle.kts.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "build.gradle.kts",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindBuildSystem,
		},
		{
			SourceTemplate: "settings.gradle.kts.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "settings.gradle.kts",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindBuildSystem,
		},
		{
			SourceTemplate:  "libs.versions.toml.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "gradle",
			TargetFileName:  "libs.versions.toml",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		{
			SourceTemplate: "gradle.properties.gohtml",
			Snippets:       templateapi.DefaultSnippets,
			TargetFileName: "gradle.properties",
			Type:           templateapi.TypeSupportOnce,
			Kind:           templateapi.KindBuildSystem,
		},
		// gradle wrapper
		{
			SourceFile:      "gradle/gradle-wrapper.properties",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "gradle/wrapper",
			TargetFileName:  "gradle-wrapper.properties",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		{
			SourceFile:      "gradle/gradle-wrapper.jar",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "gradle/wrapper",
			TargetFileName:  "gradle-wrapper.jar",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		{
			SourceFile:      "gradle/gradlew",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "",
			TargetFileName:  "gradlew",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		{
			SourceFile:      "gradle/gradlew.bat",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "",
			TargetFileName:  "gradlew.bat",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
	},
}
