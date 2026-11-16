package krakend

import (
	"testing"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKrakenConfigUsesBackendHostForAllOperations(t *testing.T) {
	data := openapigenerator.DocumentModel{
		DisplayName: "Petstore",
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get"},
			{Path: "/pets", Method: "post"},
		},
	}

	conf := DefaultConfig()
	conf.CircuitBreakerEnabled = true
	config := mustGenerate(t, data, GenerateKrakenConfigOpts{
		Config:          conf,
		BackendHost:     "https://api-stage.example.com",
		BackendBasePath: "/api/v1",
		Environment:     EnvironmentStage,
		APIName:         "Petstore",
	})

	require.Len(t, config.Endpoints, 2)
	assert.Equal(t, "Petstore Gateway (stage)", config.Name)

	for _, ep := range config.Endpoints {
		require.Len(t, ep.Backend, 1)
		assert.Equal(t, []string{"https://api-stage.example.com"}, ep.Backend[0].Host)
		assert.Equal(t, "/api/v1"+ep.Endpoint, ep.Backend[0].URLPattern)
		assert.Equal(t, false, ep.Backend[0].IsCollection)
		assert.Equal(t, "", ep.Backend[0].Target)
		assert.Contains(t, ep.Backend[0].ExtraConfig, "qos/http-cache")
		ratelimit := ep.Backend[0].ExtraConfig["qos/ratelimit/proxy"].(map[string]interface{})
		assert.Equal(t, 100, ratelimit["max_rate"])
		assert.Equal(t, 100, ratelimit["capacity"])
		assert.Equal(t, "30s", ratelimit["every"])
	}
}

func TestGenerateKrakenConfigEncodingSelection(t *testing.T) {
	data := openapigenerator.DocumentModel{
		DisplayName: "Demo",
		Operations: []openapigenerator.Operation{
			{Path: "/one", Method: "get", ReturnType: openapigenerator.CodeType{Name: "Pet"}, Parameters: []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/json"}}},
			{Path: "/many", Method: "get", ReturnType: openapigenerator.CodeType{IsArray: true}, Parameters: []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/json"}}},
			{Path: "/primitive", Method: "get", ReturnType: openapigenerator.CodeType{Name: "string"}, Parameters: []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/json"}}},
			{Path: "/xml", Method: "get", Parameters: []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/xml"}}},
			{Path: "/bin", Method: "get", Parameters: []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/octet-stream"}}},
		},
	}

	config := mustGenerate(t, data, GenerateKrakenConfigOpts{Config: DefaultConfig(), BackendHost: "https://api.example.com", APIName: "Demo"})
	require.Len(t, config.Endpoints, 5)

	assert.Equal(t, "json", config.Endpoints[0].OutputEncoding)
	assert.Equal(t, "json", config.Endpoints[0].Backend[0].Encoding)

	assert.Equal(t, "json-collection", config.Endpoints[1].OutputEncoding)
	assert.Equal(t, "json", config.Endpoints[1].Backend[0].Encoding)
	assert.True(t, config.Endpoints[1].Backend[0].IsCollection)

	assert.Equal(t, "no-op", config.Endpoints[2].OutputEncoding)
	assert.Equal(t, "no-op", config.Endpoints[2].Backend[0].Encoding)

	assert.Equal(t, "no-op", config.Endpoints[3].OutputEncoding)
	assert.Equal(t, "no-op", config.Endpoints[3].Backend[0].Encoding)

	assert.Equal(t, "no-op", config.Endpoints[4].OutputEncoding)
	assert.Equal(t, "no-op", config.Endpoints[4].Backend[0].Encoding)
}

func TestGenerateKrakenConfigDefaultTimeouts(t *testing.T) {
	config := mustGenerate(t, openapigenerator.DocumentModel{}, GenerateKrakenConfigOpts{Config: DefaultConfig(), BackendHost: "https://api.example.com", APIName: "Demo"})

	assert.Equal(t, "3s", config.Timeout)
	assert.Equal(t, "10s", config.ReadTimeout)
	assert.Equal(t, "10s", config.WriteTimeout)
	assert.Equal(t, "10s", config.IdleTimeout)
	assert.Equal(t, "10s", config.ReadHeaderTimeout)
}

func TestGenerateKrakenConfigCircuitBreakerPerOperationExtension(t *testing.T) {
	withCB := openapigenerator.Operation{Path: "/with", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{"x-krakend-circuit-breaker-enabled": true})}
	withoutCB := openapigenerator.Operation{Path: "/without", Method: "get"}
	withDisabledCB := openapigenerator.Operation{Path: "/off", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{"x-krakend-circuit-breaker-enabled": false})}

	config := mustGenerate(t, openapigenerator.DocumentModel{Operations: []openapigenerator.Operation{withCB, withoutCB, withDisabledCB}}, GenerateKrakenConfigOpts{
		Config:      DefaultConfig(),
		BackendHost: "https://api.example.com",
		APIName:     "Demo",
	})

	require.Len(t, config.Endpoints, 3)
	assert.Contains(t, config.Endpoints[0].Backend[0].ExtraConfig, "qos/circuit-breaker")
	assert.Contains(t, config.Endpoints[0].Backend[0].ExtraConfig, "qos/ratelimit/proxy")
	assert.NotContains(t, config.Endpoints[1].Backend[0].ExtraConfig, "qos/circuit-breaker")
	assert.NotContains(t, config.Endpoints[2].Backend[0].ExtraConfig, "qos/circuit-breaker")
}

func operationExtensions(t *testing.T, value map[string]interface{}) *orderedmap.Map[string, *yaml.Node] {
	t.Helper()
	content, err := yaml.Marshal(value)
	require.NoError(t, err)

	var root yaml.Node
	err = yaml.Unmarshal(content, &root)
	require.NoError(t, err)
	require.NotEmpty(t, root.Content)

	extensions := orderedmap.New[string, *yaml.Node]()
	mapNode := root.Content[0]
	for i := 0; i+1 < len(mapNode.Content); i += 2 {
		key := mapNode.Content[i].Value
		val := mapNode.Content[i+1]
		extensions.Set(key, val)
	}

	return extensions
}

func mustGenerate(t *testing.T, data openapigenerator.DocumentModel, opts GenerateKrakenConfigOpts) KrakenDConfig {
	t.Helper()
	config, err := generateKrakenConfig(data, opts)
	require.NoError(t, err)
	return config
}
