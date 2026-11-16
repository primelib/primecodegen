package krakend

import (
	"fmt"

	"github.com/gosimple/slug"
	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"slices"
	"strconv"
	"strings"
)

type GenerateKrakenConfigOpts struct {
	Config            Config
	Environment       string
	APIName           string
	BackendHost       string // backend host including the scheme, e.g. https://api.example.com
	BackendBasePath   string
	EndpointPrefix    string                   // root x-krakend.endpoint_prefix
	Auth              map[string]interface{}   // root x-krakend.auth, base auth/validator config for secured operations
	ClientCredentials map[string]interface{}   // root x-krakend.client_credentials, base auth/client-credentials config for backends
	Operations        map[string]OperationMeta // keyed by operationKey(method, path)
}

func generateKrakenConfig(data openapigenerator.DocumentModel, opts GenerateKrakenConfigOpts) (KrakenDConfig, error) {
	conf := opts.Config
	config := KrakenDConfig{
		Schema:            "https://www.krakend.io/schema/krakend.json",
		Version:           3,
		Name:              buildGatewayName(opts.APIName),
		Timeout:           conf.Timeout,
		ReadTimeout:       conf.ReadTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		Port:              conf.Port,
		CacheTTL:          conf.CacheTTL,
		OutputEncoding:    "json",
		DebugEndpoint:     slices.Contains(conf.DebugEndpointEnvironments, opts.Environment),
		EchoEndpoint:      slices.Contains(conf.EchoEndpointEnvironments, opts.Environment),
		DisableRest:       true,
		Endpoints:         []Endpoint{},
	}
	if opts.Environment != "" {
		config.Name = config.Name + " (" + opts.Environment + ")"
	}
	if conf.CORSEnabled {
		config.ExtraConfig.CORS = &CORSConfig{
			AllowOrigins:  conf.CORSAllowOrigins,
			ExposeHeaders: []string{"Content-Length"},
			MaxAge:        "12h",
			AllowMethods:  operationMethods(data.Operations),
		}
	}
	if conf.BotDetectorEnabled {
		config.ExtraConfig.BotDetector = &BotDetectorConfig{
			EmptyUserAgentIsBot: true,
			CacheSize:           10000,
		}
	}
	if conf.HideVersionHeader || len(conf.LoggerSkipPaths) > 0 || len(conf.TrustedProxies) > 0 {
		config.ExtraConfig.Router = &RouterConfig{
			HideVersionHeader: conf.HideVersionHeader,
			LoggerSkipPaths:   conf.LoggerSkipPaths,
		}
		if len(conf.TrustedProxies) > 0 {
			config.ExtraConfig.Router.ForwardedByClientIP = true
			config.ExtraConfig.Router.RemoteIPHeaders = []string{"X-Forwarded-For"}
			config.ExtraConfig.Router.TrustedProxies = conf.TrustedProxies
		}
	}
	if conf.SecurityHeadersEnabled {
		config.ExtraConfig.SecurityHTTP = &SecurityHTTPConfig{
			FrameDeny:            true,
			ContentTypeNosniff:   true,
			BrowserXSSFilter:     true,
			ReferrerPolicy:       "same-origin",
			STSSeconds:           conf.STSSeconds,
			STSIncludeSubdomains: conf.STSSeconds > 0,
		}
	}
	if conf.LoggingLevel != "" {
		config.ExtraConfig.Logging = &LoggingConfig{
			Level:  conf.LoggingLevel,
			Prefix: "[KRAKEND]",
			Stdout: true,
			Format: conf.LoggingFormat,
		}
		if conf.LoggingFormat == "logstash" {
			config.ExtraConfig.Logstash = &LogstashConfig{Enabled: true}
		}
	}
	if conf.GELFAddress != "" {
		config.ExtraConfig.GELF = &GELFConfig{Address: conf.GELFAddress}
	}
	config.ExtraConfig.OpenTelemetry = openTelemetryConfig(conf, buildServiceName(opts.APIName))

	// endpoints
	exposed, err := exposedOperations(data.Operations)
	if err != nil {
		return config, err
	}
	seen := map[string]bool{}
	for _, e := range exposed {
		op, settings := e.Operation, e.Settings
		meta := opts.Operations[operationKey(op.Method, op.Path)]
		opSlug := slug.Make(op.Path + "-" + op.Method)

		endpointEncoding, backendEncoding := operationEncodings(op)
		if settings.OutputEncoding != "" {
			endpointEncoding = settings.OutputEncoding
		}
		if settings.BackendEncoding != "" {
			backendEncoding = settings.BackendEncoding
		}

		ep := Endpoint{
			Endpoint:        endpointPath(opts.EndpointPrefix, op, settings),
			Method:          strings.ToUpper(op.Method),
			OutputEncoding:  endpointEncoding,
			Timeout:         settings.Timeout,
			CacheTTL:        settings.CacheTTL,
			ConcurrentCalls: settings.ConcurrentCalls,
		}
		key := operationKey(ep.Method, ep.Endpoint)
		if seen[key] {
			return config, fmt.Errorf("duplicate gateway endpoint %s (check %s.path and endpoint_prefix)", key, krakendExtensionKey)
		}
		seen[key] = true

		// forwarded query strings and headers, krakend drops everything that is not listed
		for _, qp := range op.QueryParameters {
			ep.InputQueryStrings = appendUnique(ep.InputQueryStrings, qp.Name)
		}
		ep.InputHeaders = appendUnique(ep.InputHeaders, "Content-Type", "Accept", "User-Agent")
		for _, hp := range op.HeaderParameters {
			ep.InputHeaders = appendUnique(ep.InputHeaders, hp.Name)
		}
		if len(op.CookieParameters) > 0 {
			ep.InputHeaders = appendUnique(ep.InputHeaders, "Cookie")
		}
		if meta.Security != nil {
			ep.InputHeaders = appendUnique(ep.InputHeaders, meta.Security.ForwardHeaders...)
			ep.InputQueryStrings = appendUnique(ep.InputQueryStrings, meta.Security.ForwardQueryStrings...)
		}
		ep.InputHeaders = mergeInputList(ep.InputHeaders, settings.InputHeaders)
		ep.InputQueryStrings = mergeInputList(ep.InputQueryStrings, settings.InputQueryStrings)

		// endpoint config
		endpointExtraConfig := map[string]interface{}{}
		if validator := authValidatorConfig(opts.Auth, meta.Security, settings.Auth); validator != nil {
			endpointExtraConfig["auth/validator"] = validator
		}
		if settings.RateLimit.resolve(false) {
			endpointExtraConfig["qos/ratelimit/router"] = settings.RateLimit.apply(map[string]interface{}{
				"max_rate": 100,
			})
		}
		if settings.Validation.resolve(conf.ValidateRequestBody) && (meta.RequestBodySchema != nil || len(settings.Validation.settings) > 0) {
			schema := map[string]interface{}{}
			deepMergeMap(schema, meta.RequestBodySchema)
			endpointExtraConfig["validation/json-schema"] = settings.Validation.apply(schema)
		}
		if len(endpointExtraConfig) > 0 {
			ep.ExtraConfig = endpointExtraConfig
		}

		// backend config
		qosEnabled := conf.CircuitBreakerEnabled
		if value, exists := operationExtensionBool(op.Extensions, "x-krakend-circuit-breaker-enabled"); exists {
			qosEnabled = value
		}

		backendExtraConfig := map[string]interface{}{}
		if settings.HTTPCache.resolve(qosEnabled) {
			backendExtraConfig["qos/http-cache"] = settings.HTTPCache.apply(map[string]interface{}{
				"shared": true,
			})
		}
		if settings.CircuitBreaker.resolve(qosEnabled) {
			backendExtraConfig["qos/circuit-breaker"] = settings.CircuitBreaker.apply(map[string]interface{}{
				"interval":          60,
				"name":              "circuit-breaker-" + opSlug,
				"timeout":           10,
				"max_errors":        1,
				"log_status_change": true,
			})
		}
		if settings.ProxyRateLimit.resolve(qosEnabled) {
			backendExtraConfig["qos/ratelimit/proxy"] = settings.ProxyRateLimit.apply(map[string]interface{}{
				"max_rate": 100,
				"capacity": 100,
				"every":    "30s",
			})
		}

		if settings.BackendHTTP.resolve(conf.ReturnErrorCode && backendEncoding != "no-op") {
			backendExtraConfig["backend/http"] = settings.BackendHTTP.apply(map[string]interface{}{
				"return_error_code": true,
			})
		}
		if clientCredentials := clientCredentialsConfig(opts.ClientCredentials, meta.Security, settings.ClientCredentials); clientCredentials != nil {
			backendExtraConfig["auth/client-credentials"] = clientCredentials
		}

		backendHost, backendBasePath := opts.BackendHost, opts.BackendBasePath
		if resolved, ok := operationBackend(meta.Servers, opts.Environment); ok {
			backendHost, backendBasePath = resolved.Host, resolved.BasePath
		}
		backendURLPattern := op.Path
		if backendBasePath != "" {
			backendURLPattern = strings.TrimSuffix(backendBasePath, "/") + op.Path
		}

		ep.Backend = append(ep.Backend, Backend{
			SD:                  "static",
			Host:                []string{backendHost},
			DisableHostSanitize: false,
			IsCollection:        endpointEncoding == "json-collection",
			URLPattern:          backendURLPattern,
			Encoding:            backendEncoding,
			Method:              strings.ToUpper(op.Method),
			ExtraConfig:         backendExtraConfig,
		})

		config.Endpoints = append(config.Endpoints, ep)
	}

	return config, nil
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
		return "API Gateway"
	}
	return apiName + " Gateway"
}

func buildServiceName(apiName string) string {
	name := slug.Make(apiName)
	if name == "" {
		return "krakend"
	}
	return name + "-gateway"
}

// operationMethods returns the distinct HTTP methods used by the operations, used for the CORS allow list
func operationMethods(operations []openapigenerator.Operation) []string {
	var methods []string
	for _, op := range operations {
		method := strings.ToUpper(op.Method)
		if method != "" && !slices.Contains(methods, method) {
			methods = append(methods, method)
		}
	}
	if len(methods) == 0 {
		return []string{"GET", "HEAD", "POST"}
	}
	slices.Sort(methods)
	return methods
}

func openTelemetryConfig(conf Config, serviceName string) *OpenTelemetryConfig {
	if conf.PrometheusPort == 0 && conf.OTLPHost == "" {
		return nil
	}

	otel := &OpenTelemetryConfig{
		ServiceName:           serviceName,
		MetricReportingPeriod: 60,
		Layers: OTLayers{
			Backend: OTLayerBackend{
				Metrics: OTStageDetails{RoundTrip: true, ReadPayload: true},
				Traces:  OTStageDetails{RoundTrip: true, ReadPayload: true},
			},
		},
	}
	if conf.PrometheusPort > 0 {
		otel.Exporters.Prometheus = append(otel.Exporters.Prometheus, PrometheusExporter{
			Name: "prometheus",
			Port: conf.PrometheusPort,
		})
	}
	if conf.OTLPHost != "" {
		otel.TraceSampleRate = conf.TraceSampleRate
		otel.Exporters.OTLP = append(otel.Exporters.OTLP, OTLPExporter{
			Name: "otlp",
			Host: conf.OTLPHost,
			Port: conf.OTLPPort,
		})
	}

	return otel
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

	// KrakenD requires no-op on both the endpoint and the backend, this includes xml since re-encoding would alter the payload
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
