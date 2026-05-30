package openapigenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTemplateProperties(t *testing.T) {
	parsed, err := ParseTemplateProperties([]string{
		"gradle.configurationPlugin.id=com.example.configuration",
		"gradle.configurationPlugin.version=1.2.3",
	})
	require.NoError(t, err)
	assert.Equal(t, "com.example.configuration", parsed["gradle.configurationPlugin.id"])
	assert.Equal(t, "1.2.3", parsed["gradle.configurationPlugin.version"])
}

func TestParseTemplatePropertiesInvalid(t *testing.T) {
	_, err := ParseTemplateProperties([]string{"gradle.configurationPlugin.id"})
	assert.ErrorContains(t, err, "expected key=value")

	_, err = ParseTemplateProperties([]string{"=value"})
	assert.ErrorContains(t, err, "key must not be empty")
}

func TestResolveTemplatePropertiesDefaults(t *testing.T) {
	resolved, err := ResolveTemplateProperties("openapi-java-httpclient", map[string]string{})
	require.NoError(t, err)

	assert.Equal(t, "me.philippheuer.configuration", resolved["gradle.configurationPlugin.id"])
	assert.Equal(t, "0.20.1", resolved["gradle.configurationPlugin.version"])
	assert.Equal(t, "projectConfiguration", resolved["gradle.projectConfiguration.blockName"])
	assert.Equal(t, "", resolved["gradle.pluginManagement.repositoryUrl"])
}

func TestResolveTemplatePropertiesEnvAndCliPrecedence(t *testing.T) {
	t.Setenv("PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_CONFIGURATIONPLUGIN_ID", "com.env.configuration")
	t.Setenv("PRIMECODEGEN_TPL_OPENAPI_JAVA_HTTPCLIENT_GRADLE_PLUGINMANAGEMENT_REPOSITORYURL", "https://env.example/maven")

	resolved, err := ResolveTemplateProperties("openapi-java-httpclient", map[string]string{
		"gradle.configurationPlugin.id":         "com.cli.configuration",
		"gradle.pluginManagement.repositoryUrl": "https://cli.example/maven",
	})
	require.NoError(t, err)

	assert.Equal(t, "com.cli.configuration", resolved["gradle.configurationPlugin.id"])
	assert.Equal(t, "https://cli.example/maven", resolved["gradle.pluginManagement.repositoryUrl"])
}

func TestResolveTemplatePropertiesUnknownKey(t *testing.T) {
	_, err := ResolveTemplateProperties("openapi-java-httpclient", map[string]string{
		"gradle.unknown": "x",
	})

	assert.ErrorContains(t, err, "unknown --tpl-prop key")
	assert.ErrorContains(t, err, "allowed")
}

func TestResolveTemplatePropertiesUnsupportedTemplate(t *testing.T) {
	_, err := ResolveTemplateProperties("openapi-go-httpclient", map[string]string{
		"gradle.configurationPlugin.id": "x",
	})
	assert.ErrorContains(t, err, "does not support --tpl-prop overrides")
}
