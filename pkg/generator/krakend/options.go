package krakend

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
)

// TemplateId is the template id used to resolve --tpl-prop overrides (`-g krakend -t config`)
const TemplateId = "openapi-krakend-config"

const (
	PropEnvironments              = "environments"
	PropServerPort                = "server.port"
	PropServerTimeout             = "server.timeout"
	PropServerReadTimeout         = "server.readTimeout"
	PropServerWriteTimeout        = "server.writeTimeout"
	PropServerIdleTimeout         = "server.idleTimeout"
	PropServerReadHeaderTimeout   = "server.readHeaderTimeout"
	PropServerCacheTTL            = "server.cacheTTL"
	PropDebugEndpointEnvironments = "debugEndpoint.environments"
	PropEchoEndpointEnvironments  = "echoEndpoint.environments"
	PropCORSEnabled               = "cors.enabled"
	PropCORSAllowOrigins          = "cors.allowOrigins"
	PropBotDetectorEnabled        = "botDetector.enabled"
	PropCircuitBreakerEnabled     = "circuitBreaker.enabled"
	PropLoggingLevel              = "logging.level"
	PropPrometheusPort            = "telemetry.prometheus.port"
	PropOTLPHost                  = "telemetry.otlp.host"
	PropOTLPPort                  = "telemetry.otlp.port"
	PropTraceSampleRate           = "telemetry.traceSampleRate"
	PropLoggingFormat             = "logging.format"
	PropLoggingGELFAddress        = "logging.gelfAddress"
	PropBackendReturnErrorCode    = "backend.returnErrorCode"
	PropValidationRequestBody     = "validation.requestBody"
	PropSecurityHeadersEnabled    = "securityHeaders.enabled"
	PropSecurityHeadersSTSSeconds = "securityHeaders.stsSeconds"
	PropRouterHideVersionHeader   = "router.hideVersionHeader"
	PropRouterTrustedProxies      = "router.trustedProxies"
	PropRouterLoggerSkipPaths     = "router.loggerSkipPaths"
)

var propertyDefs = []openapigenerator.TemplatePropertyDef{
	{Key: PropEnvironments, Description: "Comma-separated environments to generate configs for (subset of dev,stage,prod)", DefaultValue: "dev,stage,prod"},
	{Key: PropServerPort, Description: "Port the gateway listens on", DefaultValue: "8080"},
	{Key: PropServerTimeout, Description: "Default timeout for the whole request pipeline per endpoint", DefaultValue: "3s"},
	{Key: PropServerReadTimeout, Description: "Maximum duration for reading the entire request", DefaultValue: "10s"},
	{Key: PropServerWriteTimeout, Description: "Maximum duration before timing out writes of the response", DefaultValue: "10s"},
	{Key: PropServerIdleTimeout, Description: "Maximum time to wait for the next request when keep-alives are enabled", DefaultValue: "10s"},
	{Key: PropServerReadHeaderTimeout, Description: "Maximum duration for reading request headers", DefaultValue: "10s"},
	{Key: PropServerCacheTTL, Description: "Default Cache-Control max-age for all endpoints, empty to disable", DefaultValue: ""},
	{Key: PropDebugEndpointEnvironments, Description: "Comma-separated environments that expose the /__debug endpoint", DefaultValue: EnvironmentDev},
	{Key: PropEchoEndpointEnvironments, Description: "Comma-separated environments that expose the /__echo endpoint", DefaultValue: EnvironmentDev},
	{Key: PropCORSEnabled, Description: "Enable the security/cors module", DefaultValue: "true", Boolean: true},
	{Key: PropCORSAllowOrigins, Description: "Comma-separated allowed CORS origins", DefaultValue: "*"},
	{Key: PropBotDetectorEnabled, Description: "Enable the security/bot-detector module", DefaultValue: "true", Boolean: true},
	{Key: PropCircuitBreakerEnabled, Description: "Enable circuit breaker, http cache and proxy rate limit for all operations (override per operation with x-krakend-circuit-breaker-enabled)", DefaultValue: "false", Boolean: true},
	{Key: PropLoggingLevel, Description: "telemetry/logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL), empty to disable", DefaultValue: "INFO"},
	{Key: PropPrometheusPort, Description: "Port of the OpenTelemetry Prometheus exporter, 0 to disable", DefaultValue: "9090"},
	{Key: PropOTLPHost, Description: "Host of the OTLP collector, empty to disable", DefaultValue: ""},
	{Key: PropOTLPPort, Description: "Port of the OTLP collector (gRPC)", DefaultValue: "4317"},
	{Key: PropTraceSampleRate, Description: "Trace sample rate between 0 and 1 (only used when the OTLP exporter is enabled)", DefaultValue: "0.1"},
	{Key: PropLoggingFormat, Description: "telemetry/logging format: default or logstash (enables telemetry/logstash)", DefaultValue: "default"},
	{Key: PropLoggingGELFAddress, Description: "Graylog GELF UDP address (host:port) for telemetry/gelf, empty to disable", DefaultValue: ""},
	{Key: PropBackendReturnErrorCode, Description: "Return the backend status code and body on errors (backend/http return_error_code) for non no-op backends", DefaultValue: "true", Boolean: true},
	{Key: PropValidationRequestBody, Description: "Validate required JSON request bodies against the spec schema (validation/json-schema)", DefaultValue: "false", Boolean: true},
	{Key: PropSecurityHeadersEnabled, Description: "Enable security/http response headers (frame deny, nosniff, xss filter, referrer policy)", DefaultValue: "true", Boolean: true},
	{Key: PropSecurityHeadersSTSSeconds, Description: "Strict-Transport-Security max-age in seconds, 0 to disable", DefaultValue: "0"},
	{Key: PropRouterHideVersionHeader, Description: "Hide the X-KrakenD-Version response header", DefaultValue: "true", Boolean: true},
	{Key: PropRouterTrustedProxies, Description: "Comma-separated trusted proxy CIDRs, enables client IP detection via X-Forwarded-For (needed for ip based rate limits behind a load balancer)", DefaultValue: ""},
	{Key: PropRouterLoggerSkipPaths, Description: "Comma-separated paths excluded from the access log", DefaultValue: "/__health"},
}

func init() {
	defs := make([]openapigenerator.TemplatePropertyDef, len(propertyDefs))
	for i, def := range propertyDefs {
		def.EnvVar = "PRIMECODEGEN_TPL_OPENAPI_KRAKEND_CONFIG_" + strings.ToUpper(strings.ReplaceAll(def.Key, ".", "_"))
		defs[i] = def
	}
	propertyDefs = defs
	openapigenerator.RegisterTemplateProperties(TemplateId, propertyDefs)
}

// Config holds the generator settings that are not derived from the OpenAPI spec
type Config struct {
	Environments              []string
	Port                      int
	Timeout                   string
	ReadTimeout               string
	WriteTimeout              string
	IdleTimeout               string
	ReadHeaderTimeout         string
	CacheTTL                  string
	DebugEndpointEnvironments []string
	EchoEndpointEnvironments  []string
	CORSEnabled               bool
	CORSAllowOrigins          []string
	BotDetectorEnabled        bool
	CircuitBreakerEnabled     bool
	LoggingLevel              string
	PrometheusPort            int
	OTLPHost                  string
	OTLPPort                  int
	TraceSampleRate           float64
	LoggingFormat             string
	GELFAddress               string
	ReturnErrorCode           bool
	ValidateRequestBody       bool
	SecurityHeadersEnabled    bool
	STSSeconds                int
	HideVersionHeader         bool
	TrustedProxies            []string
	LoggerSkipPaths           []string
}

// DefaultConfig returns the config built from the property defaults
func DefaultConfig() Config {
	conf, _ := parseConfig(nil)
	return conf
}

// parseConfig builds the config from resolved template properties, missing keys fall back to their defaults
func parseConfig(props map[string]string) (Config, error) {
	values := map[string]string{}
	for _, def := range propertyDefs {
		values[def.Key] = def.DefaultValue
	}
	for k, v := range props {
		values[k] = strings.TrimSpace(v)
	}

	p := propertyParser{values: values}
	conf := Config{
		Environments:              p.environments(PropEnvironments),
		Port:                      p.int(PropServerPort),
		Timeout:                   values[PropServerTimeout],
		ReadTimeout:               values[PropServerReadTimeout],
		WriteTimeout:              values[PropServerWriteTimeout],
		IdleTimeout:               values[PropServerIdleTimeout],
		ReadHeaderTimeout:         values[PropServerReadHeaderTimeout],
		CacheTTL:                  values[PropServerCacheTTL],
		DebugEndpointEnvironments: p.environments(PropDebugEndpointEnvironments),
		EchoEndpointEnvironments:  p.environments(PropEchoEndpointEnvironments),
		CORSEnabled:               p.bool(PropCORSEnabled),
		CORSAllowOrigins:          splitList(values[PropCORSAllowOrigins]),
		BotDetectorEnabled:        p.bool(PropBotDetectorEnabled),
		CircuitBreakerEnabled:     p.bool(PropCircuitBreakerEnabled),
		LoggingLevel:              strings.ToUpper(values[PropLoggingLevel]),
		PrometheusPort:            p.int(PropPrometheusPort),
		OTLPHost:                  values[PropOTLPHost],
		OTLPPort:                  p.int(PropOTLPPort),
		TraceSampleRate:           p.float(PropTraceSampleRate),
		LoggingFormat:             strings.ToLower(values[PropLoggingFormat]),
		GELFAddress:               values[PropLoggingGELFAddress],
		ReturnErrorCode:           p.bool(PropBackendReturnErrorCode),
		ValidateRequestBody:       p.bool(PropValidationRequestBody),
		SecurityHeadersEnabled:    p.bool(PropSecurityHeadersEnabled),
		STSSeconds:                p.int(PropSecurityHeadersSTSSeconds),
		HideVersionHeader:         p.bool(PropRouterHideVersionHeader),
		TrustedProxies:            splitList(values[PropRouterTrustedProxies]),
		LoggerSkipPaths:           splitList(values[PropRouterLoggerSkipPaths]),
	}
	if p.err != nil {
		return conf, p.err
	}

	if len(conf.Environments) == 0 {
		return conf, fmt.Errorf("property %s must contain at least one environment", PropEnvironments)
	}
	if conf.LoggingFormat != "default" && conf.LoggingFormat != "logstash" {
		return conf, fmt.Errorf("property %s must be default or logstash", PropLoggingFormat)
	}
	if conf.TraceSampleRate < 0 || conf.TraceSampleRate > 1 {
		return conf, fmt.Errorf("property %s must be between 0 and 1", PropTraceSampleRate)
	}

	return conf, nil
}

type propertyParser struct {
	values map[string]string
	err    error
}

func (p *propertyParser) fail(key string, format string, args ...interface{}) {
	if p.err == nil {
		p.err = fmt.Errorf("invalid value %q for property %s: %s", p.values[key], key, fmt.Sprintf(format, args...))
	}
}

func (p *propertyParser) int(key string) int {
	value, err := strconv.Atoi(p.values[key])
	if err != nil || value < 0 {
		p.fail(key, "expected a non-negative integer")
	}
	return value
}

func (p *propertyParser) float(key string) float64 {
	value, err := strconv.ParseFloat(p.values[key], 64)
	if err != nil {
		p.fail(key, "expected a number")
	}
	return value
}

func (p *propertyParser) bool(key string) bool {
	value, err := strconv.ParseBool(p.values[key])
	if err != nil {
		p.fail(key, "expected a boolean")
	}
	return value
}

func (p *propertyParser) environments(key string) []string {
	environments := splitList(strings.ToLower(p.values[key]))
	for _, env := range environments {
		if !slices.Contains(environmentOrder, env) {
			p.fail(key, "unknown environment %q (expected one of: %s)", env, strings.Join(environmentOrder, ", "))
		}
	}
	return environments
}

func splitList(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" && !slices.Contains(out, part) {
			out = append(out, part)
		}
	}
	return out
}
