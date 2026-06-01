package krakend

import (
	"github.com/gosimple/slug"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"go.yaml.in/yaml/v4"
	"slices"
	"strconv"
	"strings"
)

type GenerateKrakenConfigOpts struct {
	DebugEndpoint          bool
	EchoEndpoint           bool
	PrometheusExporterPort *int
	OTELExporterPort       *int
	BackendHost            string
	BackendBasePath        string
	Environment            string
	APIName                string
	CircuitBreaker         bool
}

func generateKrakenConfig(data openapigenerator.DocumentModel, opts GenerateKrakenConfigOpts) KrakenDConfig {
	config := KrakenDConfig{
		Schema:            "https://www.krakend.io/schema/krakend.json",
		Version:           3,
		Name:              buildGatewayName(opts.APIName),
		Timeout:           "3000ms",
		ReadTimeout:       "10s",
		WriteTimeout:      "10s",
		IdleTimeout:       "10s",
		ReadHeaderTimeout: "10s",
		Port:              8080,
		CacheTTL:          "300s",
		OutputEncoding:    "json",
		DebugEndpoint:     opts.DebugEndpoint,
		EchoEndpoint:      opts.EchoEndpoint,
		DisableRest:       true,
		ExtraConfig: ExtraConfig{
			CORS: &CORSConfig{
				AllowOrigins:  []string{"*"},
				ExposeHeaders: []string{"Content-Length"},
				MaxAge:        "12h",
				AllowMethods:  []string{"GET", "HEAD", "POST"},
			},
			BotDetector: &BotDetectorConfig{
				EmptyUserAgentIsBot: true,
				CacheSize:           10000,
			},
			AuthValidator: &AuthValidatorConfig{
				SharedCacheDuration: 3600,
			},
		},
		Endpoints: []Endpoint{},
	}
	if opts.Environment != "" {
		config.Name = config.Name + " (" + opts.Environment + ")"
	}

	// endpoints
	for _, op := range data.Operations {
		endpointEncoding, backendEncoding := operationEncodings(op)
		ep := Endpoint{
			Endpoint:       op.Path,
			Method:         strings.ToUpper(op.Method),
			OutputEncoding: endpointEncoding,
		}
		opSlug := slug.Make(op.Path + "-" + op.Method)

		for _, qp := range op.QueryParameters {
			ep.InputQueryStrings = append(ep.InputQueryStrings, qp.Name)
		}
		ep.InputQueryStrings = slices.Compact(ep.InputQueryStrings)

		ep.InputHeaders = append(ep.InputHeaders, "Content-Type", "Accept", "User-Agent")
		for _, hp := range op.HeaderParameters {
			ep.InputHeaders = append(ep.InputHeaders, hp.Name)
		}
		ep.InputHeaders = slices.Compact(ep.InputHeaders)
		/*
			ep.ExtraConfig = map[string]interface{}{
				"auth/validator": map[string]interface{}{
					"cache":               true,
					"jwk_url":             "jwkUri",
					"issuer":              "issuerUrl", // optional?
					"alg":                 "RS256",     // optional?
					"roles_key_is_nested": true,
					"scopes_key":          "scopesKey",
					"scopes_matcher":      "any",
					"roles_key":           "rolesKey.nested.abc",
					"roles":               []string{"admin", "user"},
					"scopes":              []string{"scopea", "scopeb"},
				},
			}
		*/

		// backend config
		backendExtraConfig := map[string]interface{}{}
		circuitBreakerEnabled := opts.CircuitBreaker
		if value, exists := operationExtensionBool(op.Extensions, "x-krakend-circuit-breaker-enabled"); exists {
			circuitBreakerEnabled = value
		}

		if circuitBreakerEnabled {
			backendExtraConfig["qos/http-cache"] = map[string]interface{}{
				"shared": true,
			}
			backendExtraConfig["qos/circuit-breaker"] = map[string]interface{}{
				"interval":          60,
				"name":              "circuit-breaker-" + opSlug,
				"timeout":           10,
				"max_errors":        1,
				"log_status_change": true,
			}
			backendExtraConfig["qos/ratelimit/proxy"] = map[string]interface{}{
				"max_rate": 100,
				"capacity": 100,
				"every":    "30s",
			}
		}

		backendURLPattern := op.Path
		if opts.BackendBasePath != "" {
			backendURLPattern = strings.TrimSuffix(opts.BackendBasePath, "/") + op.Path
		}

		ep.Backend = append(ep.Backend, Backend{
			SD:                  "static",
			Host:                []string{opts.BackendHost},
			DisableHostSanitize: false,
			IsCollection:        false,
			Target:              "",
			URLPattern:          backendURLPattern,
			Encoding:            backendEncoding,
			Method:              strings.ToUpper(op.Method),
			ExtraConfig:         backendExtraConfig,
		})

		config.Endpoints = append(config.Endpoints, ep)
	}

	return config
}

func operationExtensionBool(extensions *orderedmap.Map[string, *yaml.Node], key string) (bool, bool) {
	if extensions == nil {
		return false, false
	}
	node, exists := extensions.Get(key)
	if !exists || node == nil {
		return false, false
	}

	var asBool bool
	if err := node.Decode(&asBool); err == nil {
		return asBool, true
	}

	var asString string
	if err := node.Decode(&asString); err == nil {
		parsed, parseErr := strconv.ParseBool(strings.TrimSpace(asString))
		if parseErr == nil {
			return parsed, true
		}
	}

	return false, false
}

func buildGatewayName(apiName string) string {
	apiName = strings.TrimSpace(apiName)
	if apiName == "" {
		return "API Client SDK"
	}
	return apiName + " Client SDK"
}

func operationEncodings(op openapigenerator.Operation) (string, string) {
	mediaType := responseMediaType(op)
	if strings.Contains(mediaType, "json") {
		if op.ReturnType.IsArray || op.ReturnType.IsList {
			return "json-collection", "json"
		}
		if isJSONObjectReturnType(op.ReturnType) {
			return "json", "json"
		}
		return "no-op", "no-op"
	}
	if strings.Contains(mediaType, "xml") {
		return "no-op", "xml"
	}

	return "no-op", "no-op"
}

func responseMediaType(op openapigenerator.Operation) string {
	for _, p := range op.Parameters {
		if strings.EqualFold(p.In, "header") && strings.EqualFold(p.FieldName, "Accept") && p.StaticValue != "" {
			return strings.ToLower(strings.TrimSpace(p.StaticValue))
		}
	}
	return ""
}

func isJSONObjectReturnType(returnType openapigenerator.CodeType) bool {
	if returnType.IsVoid {
		return false
	}
	if returnType.IsMap {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(returnType.Name))
	if name == "" {
		return false
	}
	if name == "object" || name == "map" {
		return true
	}
	if isPrimitiveName(name) {
		return false
	}

	return true
}

func isPrimitiveName(name string) bool {
	switch name {
	case "string", "bool", "boolean", "byte", "int", "int32", "int64", "float", "float32", "float64", "number", "integer":
		return true
	default:
		return false
	}
}
