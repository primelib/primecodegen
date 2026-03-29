package openapi_kotlin_httpclient

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
)

var Template = templateapi.Config{
	ID:          "openapi-kotlin-httpclient",
	Description: "OpenAPI Server for Kotlin Spring",
	Files: []templateapi.File{
		// core - main
		{
			SourceTemplate:  "build.gradle.kts.core.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core",
			TargetFileName:  "build.gradle.kts",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		// core - factory - common
		{
			SourceTemplate:  "api_factoryspec.common.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}FactorySpec.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// core - factory - jvm
		{
			SourceTemplate:  "api_factory.jvm.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Root | toFilePath }}",
			TargetFileName:  "Jvm{{ .Metadata.Name }}Factory.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "api_factoryspec.jvm.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Root | toFilePath }}",
			TargetFileName:  "Jvm{{ .Metadata.Name }}FactorySpec.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// core - main api - common
		{
			SourceTemplate:  "api_main.common.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Metadata.Name }}Api.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// core - main api - jvm
		{
			SourceTemplate:  "api_main.jvm.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "Jvm{{ .Metadata.Name }}Api.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "operation.jvm.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Operations | toFilePath }}",
			TargetFileName:  "{{ .Operation.Name }}OperationSpec.kt",
			Type:            templateapi.TypeOperationEach,
			Kind:            templateapi.KindAPI,
		},
		// core - service api - common
		{
			SourceTemplate:  "api_service.common.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "{{ .Service.Type }}Api.kt",
			Type:            templateapi.TypeAPIEach,
			Kind:            templateapi.KindAPI,
		},
		// core - service api - jvm
		{
			SourceTemplate:  "api_service.jvm.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Client | toFilePath }}",
			TargetFileName:  "Jvm{{ .Service.Type }}Api.kt",
			Type:            templateapi.TypeAPIEach,
			Kind:            templateapi.KindAPI,
		},
		// core - model
		{
			SourceTemplate:  "model.common.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Models | toFilePath }}",
			TargetFileName:  "{{ .Name }}.kt",
			Type:            templateapi.TypeModelEach,
			Kind:            templateapi.KindModel,
		},
		{
			SourceTemplate:  "enum.common.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Enums | toFilePath }}",
			TargetFileName:  "{{ .Name }}.kt",
			Type:            templateapi.TypeEnumEach,
			Kind:            templateapi.KindModel,
		},
		// core - model - jvm
		/*
			{
				SourceTemplate:  "model.jvm.gohtml",
				Snippets:        templateapi.DefaultSnippets,
				TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Models | toFilePath }}/jvm",
				TargetFileName:  "{{ .Name }}.kt",
				Type:            templateapi.TypeModelEach,
				Kind:            templateapi.KindModel,
			},
			{
				SourceTemplate:  "enum.jvm.gohtml",
				Snippets:        templateapi.DefaultSnippets,
				TargetDirectory: "core/src/jvmMain/kotlin/{{ .Common.Packages.Enums | toFilePath }}/jvm",
				TargetFileName:  "{{ .Name }}.kt",
				Type:            templateapi.TypeEnumEach,
				Kind:            templateapi.KindModel,
			},
		*/
		// core - operation response models
		{
			SourceTemplate:  "response.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Responses | toFilePath }}",
			TargetFileName:  "{{ .Operation.Name }}Response.kt",
			Type:            templateapi.TypeOperationEach,
			Kind:            templateapi.KindAPI,
		},
		// core - auth
		{
			SourceTemplate:  "auth_api.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "AuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "auth_apikey.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "ApiKeyAuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "auth_basic.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "BasicAuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "auth_bearer.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "BearerAuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "auth_oauth2client.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "OAuth2ClientCredentialAuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "auth_oauth2user.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "core/src/commonMain/kotlin/{{ .Common.Packages.Root | toFilePath }}/auth",
			TargetFileName:  "OAuth2UserCredentialAuthMethod.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		// module - spring
		{
			SourceTemplate:  "build.gradle.kts.spring.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "spring",
			TargetFileName:  "build.gradle.kts",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindBuildSystem,
		},
		{
			SourceTemplate:  "spring_auto_configuration.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "spring/src/main/java/{{ .Common.Packages.Root | toFilePath }}/spring",
			TargetFileName:  "{{ .Metadata.Name }}SpringAutoConfiguration.kt",
			Type:            templateapi.TypeAPIOnce,
			Kind:            templateapi.KindAPI,
		},
		{
			SourceTemplate:  "spring_auto_configuration_imports.gohtml",
			Snippets:        templateapi.DefaultSnippets,
			TargetDirectory: "spring/src/main/resources/META-INF/spring",
			TargetFileName:  "org.springframework.boot.autoconfigure.AutoConfiguration.imports",
			Type:            templateapi.TypeSupportOnce,
			Kind:            templateapi.KindAPI,
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
			SourceTemplate: "build.gradle.kts.root.gohtml",
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
