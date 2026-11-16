package krakend

import (
	"testing"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootKrakenDExtension(t *testing.T) {
	spec := []byte(`openapi: 3.0.3
info:
  title: test
  version: 1.0.0
  x-name: test
x-krakend:
  service:
    timeout: 45s
    extra_config:
      telemetry/logging:
        level: WARNING
paths: {}
`)

	doc := openapidocument.OpenV3DocumentForTest(spec)

	extension, err := rootKrakenDExtension(doc)
	require.NoError(t, err)

	service, ok := extension["service"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "45s", service["timeout"])
}

func TestApplyKrakenDOverrides(t *testing.T) {
	config := map[string]interface{}{
		"timeout": "30s",
		"endpoints": []interface{}{
			map[string]interface{}{
				"endpoint": "/pets",
				"method":   "GET",
				"backend": []interface{}{
					map[string]interface{}{
						"host": []interface{}{"https://api-prod.example.com"},
						"extra_config": map[string]interface{}{
							"qos/circuit-breaker": map[string]interface{}{"max_errors": 5},
						},
					},
				},
			},
		},
	}

	opExtNode := mustYAMLNode(t, map[string]interface{}{
		"endpoint": map[string]interface{}{
			"timeout": "5s",
		},
		"backend_extra_config": map[string]interface{}{
			"qos/ratelimit/proxy": map[string]interface{}{
				"max_rate": 100,
			},
		},
	})
	opExt := orderedmap.New[string, *yaml.Node]()
	opExt.Set(krakendExtensionKey, opExtNode)

	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{
				Method:     "get",
				Path:       "/pets",
				Extensions: opExt,
			},
		},
	}

	rootExtension := map[string]interface{}{
		"service": map[string]interface{}{
			"cache_ttl": "60s",
		},
		"endpoint_defaults": map[string]interface{}{
			"output_encoding": "json",
		},
		"backend_defaults": map[string]interface{}{
			"disable_host_sanitize": false,
		},
	}

	err := applyKrakenDOverrides(config, data, rootExtension)
	require.NoError(t, err)

	assert.Equal(t, "60s", config["cache_ttl"])

	endpoints := config["endpoints"].([]interface{})
	endpoint := endpoints[0].(map[string]interface{})
	assert.Equal(t, "json", endpoint["output_encoding"])
	assert.Equal(t, "5s", endpoint["timeout"])

	backends := endpoint["backend"].([]interface{})
	backend := backends[0].(map[string]interface{})
	assert.Equal(t, false, backend["disable_host_sanitize"])

	extraConfig := backend["extra_config"].(map[string]interface{})
	_, hasCircuitBreaker := extraConfig["qos/circuit-breaker"]
	assert.True(t, hasCircuitBreaker)
	_, hasRatelimit := extraConfig["qos/ratelimit/proxy"]
	assert.True(t, hasRatelimit)
}

func mustYAMLNode(t *testing.T, value interface{}) *yaml.Node {
	t.Helper()
	bytes, err := yaml.Marshal(value)
	require.NoError(t, err)

	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	require.NoError(t, err)
	require.NotEmpty(t, node.Content)

	return node.Content[0]
}
