package krakend

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationSettingsEndpoint(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/internal", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{"skip": true},
			})},
			{Path: "/pets", Method: "get", QueryParameters: []openapigenerator.Parameter{{Name: "limit"}}, Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{
					"path":                "animals",
					"timeout":             "5s",
					"cache_ttl":           "60s",
					"concurrent_calls":    2,
					"input_headers":       []string{"X-Tenant"},
					"input_query_strings": []string{"*"},
					"output_encoding":     "string",
				},
			})},
		},
	}

	config := mustGenerate(t, data, GenerateKrakenConfigOpts{Config: DefaultConfig(), EndpointPrefix: "/petstore/", BackendHost: "https://api.example.com"})
	require.Len(t, config.Endpoints, 1)
	ep := config.Endpoints[0]
	assert.Equal(t, "/petstore/animals", ep.Endpoint)
	assert.Equal(t, "/pets", ep.Backend[0].URLPattern)
	assert.Equal(t, "5s", ep.Timeout)
	assert.Equal(t, "60s", ep.CacheTTL)
	assert.Equal(t, 2, ep.ConcurrentCalls)
	assert.Equal(t, "string", ep.OutputEncoding)
	assert.Contains(t, ep.InputHeaders, "X-Tenant")
	assert.Equal(t, []string{"*"}, ep.InputQueryStrings)
}

func TestOperationSettingsQoSToggles(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{
					"circuit_breaker": false,
					"http_cache":      map[string]interface{}{"shared": false},
					"rate_limit":      map[string]interface{}{"max_rate": 10, "client_max_rate": 1, "strategy": "ip"},
				},
			})},
		},
	}
	conf := DefaultConfig()
	conf.CircuitBreakerEnabled = true

	config := mustGenerate(t, data, GenerateKrakenConfigOpts{Config: conf, BackendHost: "https://api.example.com"})
	ep := config.Endpoints[0]
	assert.NotContains(t, ep.Backend[0].ExtraConfig, "qos/circuit-breaker")
	assert.Contains(t, ep.Backend[0].ExtraConfig, "qos/ratelimit/proxy")
	assert.Equal(t, false, ep.Backend[0].ExtraConfig["qos/http-cache"].(map[string]interface{})["shared"])

	rateLimit := ep.ExtraConfig["qos/ratelimit/router"].(map[string]interface{})
	assert.Equal(t, 10, rateLimit["max_rate"])
	assert.Equal(t, "ip", rateLimit["strategy"])
}

func TestOperationSettingsInvalid(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{"circuit_breaker": []string{"x"}},
			})},
		},
	}
	_, err := generateKrakenConfig(data, GenerateKrakenConfigOpts{Config: DefaultConfig()})
	assert.ErrorContains(t, err, "GET /pets")
}

func TestOperationSettingsDuplicateEndpoint(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get"},
			{Path: "/animals", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{"path": "/pets"},
			})},
		},
	}
	_, err := generateKrakenConfig(data, GenerateKrakenConfigOpts{Config: DefaultConfig()})
	assert.ErrorContains(t, err, "duplicate")
}

func TestOperationSecurityAndServers(t *testing.T) {
	spec := []byte(`openapi: 3.0.3
info:
  title: test
  version: 1.0.0
security:
  - oauth: [pets:read]
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
  /health:
    get:
      security: []
      responses:
        "200":
          description: ok
  /legacy:
    servers:
      - url: https://legacy-prod.example.com/v0
    get:
      security:
        - apiKey: []
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    oauth:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/token
          scopes:
            pets:read: read pets
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
`)
	doc := openapidocument.OpenV3DocumentForTest(spec)
	meta := buildOperationMeta(doc)

	require.NotNil(t, meta["GET /pets"].Security)
	assert.True(t, meta["GET /pets"].Security.Bearer)
	assert.Equal(t, []string{"pets:read"}, meta["GET /pets"].Security.Scopes)
	assert.Nil(t, meta["GET /health"].Security)
	require.NotNil(t, meta["GET /legacy"].Security)
	assert.Equal(t, []string{"X-API-Key"}, meta["GET /legacy"].Security.ForwardHeaders)
	assert.False(t, meta["GET /legacy"].Security.Bearer)

	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get"},
			{Path: "/health", Method: "get"},
			{Path: "/legacy", Method: "get"},
		},
	}
	config := mustGenerate(t, data, GenerateKrakenConfigOpts{
		Config:      DefaultConfig(),
		Environment: EnvironmentProd,
		BackendHost: "https://api.example.com",
		Auth:        map[string]interface{}{"alg": "RS256", "jwk_url": "https://auth.example.com/jwks"},
		Operations:  meta,
	})

	pets, health, legacy := config.Endpoints[0], config.Endpoints[1], config.Endpoints[2]
	validator := pets.ExtraConfig["auth/validator"].(map[string]interface{})
	assert.Equal(t, "https://auth.example.com/jwks", validator["jwk_url"])
	assert.Equal(t, []string{"pets:read"}, validator["scopes"])
	assert.Equal(t, "scope", validator["scopes_key"])
	assert.Contains(t, pets.InputHeaders, "Authorization")

	assert.NotContains(t, health.ExtraConfig, "auth/validator")
	assert.NotContains(t, health.InputHeaders, "Authorization")

	assert.NotContains(t, legacy.ExtraConfig, "auth/validator")
	assert.Contains(t, legacy.InputHeaders, "X-API-Key")
	assert.Equal(t, []string{"https://legacy-prod.example.com"}, legacy.Backend[0].Host)
	assert.Equal(t, "/v0/legacy", legacy.Backend[0].URLPattern)
}

func TestApplyKrakenDOverridesPrecedence(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{
					"timeout":   "5s",
					"cache_ttl": "10s",
					"endpoint":  map[string]interface{}{"cache_ttl": "20s"},
				},
			})},
			{Path: "/owners", Method: "get"},
		},
	}
	config, err := generateKrakenConfig(data, GenerateKrakenConfigOpts{Config: DefaultConfig(), BackendHost: "https://api.example.com"})
	require.NoError(t, err)
	configMap, err := configToMap(config)
	require.NoError(t, err)

	err = applyKrakenDOverrides(configMap, data, map[string]interface{}{
		"endpoint_defaults": map[string]interface{}{"timeout": "1s"},
	})
	require.NoError(t, err)

	endpoints := configMap["endpoints"].([]interface{})
	pets := endpoints[0].(map[string]interface{})
	owners := endpoints[1].(map[string]interface{})
	assert.Equal(t, "5s", pets["timeout"])
	assert.Equal(t, "20s", pets["cache_ttl"])
	assert.Equal(t, "1s", owners["timeout"])
}

func TestGenerateKrakenConfigDerivedBackendAndValidation(t *testing.T) {
	jsonAccept := []openapigenerator.Parameter{{In: "header", FieldName: "Accept", StaticValue: "application/json"}}
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "post", ReturnType: openapigenerator.CodeType{Name: "Pet"}, Parameters: jsonAccept},
			{Path: "/pets/{id}", Method: "put", ReturnType: openapigenerator.CodeType{Name: "Pet"}, Parameters: jsonAccept, Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{"validation": false, "client_credentials": false},
			})},
			{Path: "/files", Method: "get"},
		},
	}
	schema := map[string]interface{}{"type": "object", "required": []interface{}{"name"}}
	security := &OperationSecurity{ForwardHeaders: []string{"Authorization"}, TokenURL: "https://auth.example.com/token"}
	conf := DefaultConfig()
	conf.ValidateRequestBody = true

	config := mustGenerate(t, data, GenerateKrakenConfigOpts{
		Config:            conf,
		BackendHost:       "https://api.example.com",
		ClientCredentials: map[string]interface{}{"client_id": "gateway", "client_secret": "secret"},
		Operations: map[string]OperationMeta{
			"POST /pets":     {RequestBodySchema: schema, Security: security},
			"PUT /pets/{id}": {RequestBodySchema: schema, Security: security},
		},
	})
	post, put, files := config.Endpoints[0], config.Endpoints[1], config.Endpoints[2]

	assert.Equal(t, schema, post.ExtraConfig["validation/json-schema"])
	assert.Equal(t, map[string]interface{}{"return_error_code": true}, post.Backend[0].ExtraConfig["backend/http"])
	clientCredentials := post.Backend[0].ExtraConfig["auth/client-credentials"].(map[string]interface{})
	assert.Equal(t, "https://auth.example.com/token", clientCredentials["token_url"])
	assert.Equal(t, "gateway", clientCredentials["client_id"])

	assert.NotContains(t, put.ExtraConfig, "validation/json-schema")
	assert.NotContains(t, put.Backend[0].ExtraConfig, "auth/client-credentials")

	// no-op backends pass status codes through without backend/http
	assert.NotContains(t, files.Backend[0].ExtraConfig, "backend/http")
}

func TestApplyKrakenDOverridesNamespaceShortcuts(t *testing.T) {
	data := openapigenerator.DocumentModel{
		Operations: []openapigenerator.Operation{
			{Path: "/pets", Method: "get", Extensions: operationExtensions(t, map[string]interface{}{
				"x-krakend": map[string]interface{}{
					"cel":         []interface{}{map[string]interface{}{"check_expr": "'x' in req_headers"}},
					"proxy":       map[string]interface{}{"static": map[string]interface{}{"strategy": "errored", "data": map[string]interface{}{}}},
					"martian":     map[string]interface{}{"header.Modifier": map[string]interface{}{"scope": []interface{}{"request"}, "name": "X-Gateway", "value": "krakend"}},
					"lua_backend": map[string]interface{}{"sources": []interface{}{"backend.lua"}},
				},
			})},
		},
	}
	config, err := generateKrakenConfig(data, GenerateKrakenConfigOpts{Config: DefaultConfig(), BackendHost: "https://api.example.com"})
	require.NoError(t, err)
	configMap, err := configToMap(config)
	require.NoError(t, err)

	err = applyKrakenDOverrides(configMap, data, map[string]interface{}{
		"revoker": map[string]interface{}{"N": 10000000, "P": 0.0000001, "hash_name": "optimal", "TTL": 1500, "port": 1234, "token_keys": []interface{}{"jti"}},
	})
	require.NoError(t, err)

	assert.Contains(t, configMap["extra_config"], "auth/revoker")
	endpoint := configMap["endpoints"].([]interface{})[0].(map[string]interface{})
	endpointExtraConfig := endpoint["extra_config"].(map[string]interface{})
	assert.Contains(t, endpointExtraConfig, "validation/cel")
	assert.Contains(t, endpointExtraConfig, "proxy")
	backendExtraConfig := endpoint["backend"].([]interface{})[0].(map[string]interface{})["extra_config"].(map[string]interface{})
	assert.Contains(t, backendExtraConfig, "modifier/martian")
	assert.Contains(t, backendExtraConfig, "modifier/lua-backend")
}

func TestBuildOperationMetaRequestBodyAndTokenURL(t *testing.T) {
	spec := []byte(`openapi: 3.0.3
info:
  title: test
  version: 1.0.0
paths:
  /pets:
    post:
      security:
        - oauth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Pet"
      responses:
        "201":
          description: created
    put:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Pet"
      responses:
        "200":
          description: ok
components:
  schemas:
    Pet:
      type: object
      required: [name]
      properties:
        name:
          type: string
  securitySchemes:
    oauth:
      type: oauth2
      flows:
        clientCredentials:
          tokenUrl: https://auth.example.com/token
          scopes: {}
`)
	meta := buildOperationMeta(openapidocument.OpenV3DocumentForTest(spec))

	post := meta["POST /pets"]
	require.NotNil(t, post.RequestBodySchema)
	assert.Equal(t, "object", post.RequestBodySchema["type"])
	assert.Contains(t, post.RequestBodySchema["properties"], "name")
	require.NotNil(t, post.Security)
	assert.Equal(t, "https://auth.example.com/token", post.Security.TokenURL)

	// optional request bodies are not validated
	assert.Nil(t, meta["PUT /pets"].RequestBodySchema)
}
