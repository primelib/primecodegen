package krakend

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfigDefaults(t *testing.T) {
	conf, err := parseConfig(nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"dev", "stage", "prod"}, conf.Environments)
	assert.Equal(t, 8080, conf.Port)
	assert.Equal(t, "", conf.CacheTTL)
	assert.Equal(t, []string{"dev"}, conf.DebugEndpointEnvironments)
	assert.Equal(t, []string{"dev"}, conf.EchoEndpointEnvironments)
	assert.Equal(t, []string{"*"}, conf.CORSAllowOrigins)
	assert.False(t, conf.CircuitBreakerEnabled)
	assert.Equal(t, 9090, conf.PrometheusPort)
	assert.Equal(t, "", conf.OTLPHost)
}

func TestParseConfigOverrides(t *testing.T) {
	conf, err := parseConfig(map[string]string{
		PropEnvironments:          "prod, stage",
		PropServerPort:            "9000",
		PropCORSAllowOrigins:      "https://a.example.com,https://b.example.com",
		PropCircuitBreakerEnabled: "true",
		PropPrometheusPort:        "0",
		PropOTLPHost:              "otel-collector",
		PropLoggingLevel:          "warning",
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"prod", "stage"}, conf.Environments)
	assert.Equal(t, 9000, conf.Port)
	assert.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, conf.CORSAllowOrigins)
	assert.True(t, conf.CircuitBreakerEnabled)
	assert.Equal(t, 0, conf.PrometheusPort)
	assert.Equal(t, "otel-collector", conf.OTLPHost)
	assert.Equal(t, "WARNING", conf.LoggingLevel)
}

func TestParseConfigInvalid(t *testing.T) {
	_, err := parseConfig(map[string]string{PropServerPort: "abc"})
	assert.ErrorContains(t, err, PropServerPort)

	_, err = parseConfig(map[string]string{PropEnvironments: "dev,qa"})
	assert.ErrorContains(t, err, "qa")

	_, err = parseConfig(map[string]string{PropEnvironments: ""})
	assert.Error(t, err)

	_, err = parseConfig(map[string]string{PropTraceSampleRate: "2"})
	assert.Error(t, err)
}

func TestTemplatePropertiesRegistered(t *testing.T) {
	resolved, err := openapigenerator.ResolveTemplateProperties(TemplateId, map[string]string{PropServerPort: "9000"})
	require.NoError(t, err)
	assert.Equal(t, "9000", resolved[PropServerPort])
	assert.Equal(t, "dev,stage,prod", resolved[PropEnvironments])

	_, err = openapigenerator.ResolveTemplateProperties(TemplateId, map[string]string{"unknown": "x"})
	assert.Error(t, err)
}

func TestGenerateKrakenConfigDebugEndpointsPerEnvironment(t *testing.T) {
	dev := mustGenerate(t, openapigenerator.DocumentModel{}, GenerateKrakenConfigOpts{Config: DefaultConfig(), Environment: EnvironmentDev})
	prod := mustGenerate(t, openapigenerator.DocumentModel{}, GenerateKrakenConfigOpts{Config: DefaultConfig(), Environment: EnvironmentProd})

	assert.True(t, dev.DebugEndpoint)
	assert.True(t, dev.EchoEndpoint)
	assert.False(t, prod.DebugEndpoint)
	assert.False(t, prod.EchoEndpoint)
}

func TestGenerateKrakenConfigServiceExtraConfig(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get"},
			{Path: "/pets", Method: "post"},
			{Path: "/pets/{id}", Method: "delete"},
		},
	}

	config := mustGenerate(t, data, GenerateKrakenConfigOpts{Config: DefaultConfig(), APIName: "Pet Store"})
	require.NotNil(t, config.ExtraConfig.CORS)
	assert.Equal(t, []string{"DELETE", "GET", "POST"}, config.ExtraConfig.CORS.AllowMethods)
	require.NotNil(t, config.ExtraConfig.OpenTelemetry)
	assert.Equal(t, "pet-store-gateway", config.ExtraConfig.OpenTelemetry.ServiceName)
	assert.Len(t, config.ExtraConfig.OpenTelemetry.Exporters.Prometheus, 1)
	assert.Empty(t, config.ExtraConfig.OpenTelemetry.Exporters.OTLP)

	conf := DefaultConfig()
	conf.CORSEnabled = false
	conf.BotDetectorEnabled = false
	conf.PrometheusPort = 0
	conf.LoggingLevel = ""
	config = mustGenerate(t, data, GenerateKrakenConfigOpts{Config: conf})
	assert.Nil(t, config.ExtraConfig.CORS)
	assert.Nil(t, config.ExtraConfig.BotDetector)
	assert.Nil(t, config.ExtraConfig.OpenTelemetry)
	assert.Nil(t, config.ExtraConfig.Logging)
}

func TestResolveEnvironmentEndpoints(t *testing.T) {
	resolved := resolveEnvironmentEndpoints(openapigenerator.Endpoints{
		{URL: "https://api-dev.example.com/v1/"},
		{URL: "http://uat.example.com"},
		{URL: "https://api.example.com", Description: "Production"},
	})

	require.Len(t, resolved, 3)
	assert.Equal(t, "https://api-dev.example.com", resolved[EnvironmentDev].Host)
	assert.Equal(t, "/v1", resolved[EnvironmentDev].BasePath)
	assert.Equal(t, "http://uat.example.com", resolved[EnvironmentStage].Host)
	assert.Equal(t, "https://api.example.com", resolved[EnvironmentProd].Host)

	resolved = resolveEnvironmentEndpoints(openapigenerator.Endpoints{{URL: "https://api.pre-production.example.com"}})
	assert.Contains(t, resolved, EnvironmentStage)
}

func TestGenerateKrakenConfigServiceModules(t *testing.T) {
	conf, err := parseConfig(map[string]string{
		PropRouterTrustedProxies:      "10.0.0.0/8",
		PropSecurityHeadersSTSSeconds: "31536000",
		PropLoggingFormat:             "logstash",
		PropLoggingGELFAddress:        "graylog:12201",
	})
	require.NoError(t, err)

	config := mustGenerate(t, openapigenerator.DocumentModel{}, GenerateKrakenConfigOpts{Config: conf})
	require.NotNil(t, config.ExtraConfig.Router)
	assert.True(t, config.ExtraConfig.Router.HideVersionHeader)
	assert.True(t, config.ExtraConfig.Router.ForwardedByClientIP)
	assert.Equal(t, []string{"10.0.0.0/8"}, config.ExtraConfig.Router.TrustedProxies)
	assert.Equal(t, []string{"/__health"}, config.ExtraConfig.Router.LoggerSkipPaths)
	require.NotNil(t, config.ExtraConfig.SecurityHTTP)
	assert.Equal(t, 31536000, config.ExtraConfig.SecurityHTTP.STSSeconds)
	assert.Equal(t, "logstash", config.ExtraConfig.Logging.Format)
	require.NotNil(t, config.ExtraConfig.Logstash)
	assert.Equal(t, "graylog:12201", config.ExtraConfig.GELF.Address)

	_, err = parseConfig(map[string]string{PropLoggingFormat: "json"})
	assert.Error(t, err)
}
