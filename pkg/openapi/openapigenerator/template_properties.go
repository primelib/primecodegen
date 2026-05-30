package openapigenerator

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type TemplatePropertyDef struct {
	Key          string
	Description  string
	DefaultValue string
	EnvVar       string
	Boolean      bool
}

var templatePropertyRegistry = map[string][]TemplatePropertyDef{
	"openapi-java-httpclient": {
		{
			Key:          "gradle.configurationPlugin.id",
			Description:  "Gradle configuration plugin id in libs.versions.toml",
			DefaultValue: "me.philippheuer.configuration",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_CONFIGURATIONPLUGIN_ID",
		},
		{
			Key:          "gradle.configurationPlugin.version",
			Description:  "Gradle configuration plugin version in libs.versions.toml",
			DefaultValue: "0.20.1",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_CONFIGURATIONPLUGIN_VERSION",
		},
		{
			Key:          "gradle.projectConfiguration.blockName",
			Description:  "Gradle extension block name for project configuration",
			DefaultValue: "projectConfiguration",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_PROJECTCONFIGURATION_BLOCKNAME",
		},
		{
			Key:          "gradle.pluginManagement.repositoryUrl",
			Description:  "Custom pluginManagement repository URL replacing official plugin repositories",
			DefaultValue: "",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_PLUGINMANAGEMENT_REPOSITORYURL",
		},
		{
			Key:          "gradle.pomMetadata.enabled",
			Description:  "Enable writing custom pom metadata block in root build.gradle.kts",
			DefaultValue: "true",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_POMMETADATA_ENABLED",
			Boolean:      true,
		},
	},
	"openapi-kotlin-httpclient": {
		{
			Key:          "gradle.configurationPlugin.id",
			Description:  "Gradle configuration plugin id in libs.versions.toml",
			DefaultValue: "me.philippheuer.configuration",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_KOTLIN_HTTPCLIENT_GRADLE_CONFIGURATIONPLUGIN_ID",
		},
		{
			Key:          "gradle.configurationPlugin.version",
			Description:  "Gradle configuration plugin version in libs.versions.toml",
			DefaultValue: "0.20.1",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_KOTLIN_HTTPCLIENT_GRADLE_CONFIGURATIONPLUGIN_VERSION",
		},
		{
			Key:          "gradle.projectConfiguration.blockName",
			Description:  "Gradle extension block name for project configuration",
			DefaultValue: "projectConfiguration",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_KOTLIN_HTTPCLIENT_GRADLE_PROJECTCONFIGURATION_BLOCKNAME",
		},
		{
			Key:          "gradle.pluginManagement.repositoryUrl",
			Description:  "Custom pluginManagement repository URL replacing official plugin repositories",
			DefaultValue: "",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_KOTLIN_HTTPCLIENT_GRADLE_PLUGINMANAGEMENT_REPOSITORYURL",
		},
		{
			Key:          "gradle.pomMetadata.enabled",
			Description:  "Enable writing custom pom metadata block in root build.gradle.kts",
			DefaultValue: "true",
			EnvVar:       "PRIMECODEGEN_TPL_OPENAPI_KOTLIN_HTTPCLIENT_GRADLE_POMMETADATA_ENABLED",
			Boolean:      true,
		},
	},
}

func ParseTemplateProperties(values []string) (map[string]string, error) {
	properties := map[string]string{}

	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --tpl-prop value %q, expected key=value", raw)
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid --tpl-prop value %q, key must not be empty", raw)
		}

		properties[key] = strings.TrimSpace(parts[1])
	}

	return properties, nil
}

func ResolveTemplateProperties(templateId string, provided map[string]string) (map[string]string, error) {
	defs := templatePropertyRegistry[templateId]
	if len(defs) == 0 {
		if len(provided) > 0 {
			return nil, fmt.Errorf("template %s does not support --tpl-prop overrides", templateId)
		}
		return map[string]string{}, nil
	}

	byKey := map[string]TemplatePropertyDef{}
	out := map[string]string{}

	for _, def := range defs {
		byKey[def.Key] = def
		out[def.Key] = def.DefaultValue
		if envValue, ok := os.LookupEnv(def.EnvVar); ok {
			out[def.Key] = envValue
		}
	}

	for key, value := range provided {
		def, ok := byKey[key]
		if !ok {
			return nil, fmt.Errorf("unknown --tpl-prop key %q for template %s (allowed: %s)", key, templateId, strings.Join(AllowedTemplatePropertyKeys(templateId), ", "))
		}
		if def.Boolean {
			parsedValue, parseErr := strconv.ParseBool(value)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid boolean value %q for --tpl-prop key %q", value, key)
			}
			value = strconv.FormatBool(parsedValue)
		}
		out[key] = value
	}

	for _, def := range defs {
		if !def.Boolean {
			continue
		}
		value := out[def.Key]
		parsedValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean value %q for key %q from default/env/cli", value, def.Key)
		}
		out[def.Key] = strconv.FormatBool(parsedValue)
	}

	return out, nil
}

func AllowedTemplatePropertyKeys(templateId string) []string {
	defs := templatePropertyRegistry[templateId]
	keys := make([]string, 0, len(defs))
	for _, def := range defs {
		keys = append(keys, def.Key)
	}
	sort.Strings(keys)
	return keys
}
