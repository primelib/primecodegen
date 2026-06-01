package krakend

type KrakenDConfig struct {
	Schema              string      `json:"$schema"`
	Version             int         `json:"version"`
	Name                string      `json:"name"`
	ExtraConfig         ExtraConfig `json:"extra_config"`
	Timeout             string      `json:"timeout"`
	ReadTimeout         string      `json:"read_timeout,omitempty"`
	WriteTimeout        string      `json:"write_timeout,omitempty"`
	IdleTimeout         string      `json:"idle_timeout,omitempty"`
	ReadHeaderTimeout   string      `json:"read_header_timeout,omitempty"`
	Port                int         `json:"port,omitempty"`
	CacheTTL            string      `json:"cache_ttl"`
	DisableKeepAlives   bool        `json:"disable_keep_alives,omitempty"`
	DisableCompression  bool        `json:"disable_compression,omitempty"`
	MaxIdleConns        int         `json:"max_idle_connections,omitempty"`
	MaxIdleConnsPerHost int         `json:"max_idle_connections_per_host,omitempty"`
	OutputEncoding      string      `json:"output_encoding"`
	DebugEndpoint       bool        `json:"debug_endpoint"`
	EchoEndpoint        bool        `json:"echo_endpoint"`
	DisableRest         bool        `json:"disable_rest"`
	Endpoints           []Endpoint  `json:"endpoints"`
}

type ExtraConfig struct {
	Router        *RouterConfig        `json:"router,omitempty"`
	CORS          *CORSConfig          `json:"security/cors,omitempty"`
	BotDetector   *BotDetectorConfig   `json:"security/bot-detector,omitempty"`
	AuthValidator *AuthValidatorConfig `json:"auth/validator,omitempty"`
	Logging       *LoggingConfig       `json:"telemetry/logging,omitempty"`
	Logstash      *LogstashConfig      `json:"telemetry/logstash,omitempty"`
	OpenTelemetry *OpenTelemetryConfig `json:"telemetry/opentelemetry,omitempty"`
}

type AuthValidatorConfig struct {
	SharedCacheDuration int `json:"shared_cache_duration,omitempty"`
}

type RouterConfig struct {
	HideVersionHeader   bool     `json:"hide_version_header,omitempty"`
	LoggerSkipPaths     []string `json:"logger_skip_paths,omitempty"`
	ForwardedByClientIP bool     `json:"forwarded_by_client_ip,omitempty"`
	RemoteIPHeaders     []string `json:"remote_ip_headers,omitempty"`
	TrustedProxies      []string `json:"trusted_proxies,omitempty"`
}

type CORSConfig struct {
	AllowOrigins  []string `json:"allow_origins"`
	ExposeHeaders []string `json:"expose_headers"`
	MaxAge        string   `json:"max_age"`
	AllowMethods  []string `json:"allow_methods"`
}

type BotDetectorConfig struct {
	EmptyUserAgentIsBot bool     `json:"empty_user_agent_is_bot"`
	CacheSize           int      `json:"cache_size"`
	Deny                []string `json:"deny,omitempty"`
}

type LoggingConfig struct {
	Level          string `json:"level"`
	Prefix         string `json:"prefix"`
	Syslog         bool   `json:"syslog"`
	Stdout         bool   `json:"stdout"`
	Format         string `json:"format"`
	SyslogFacility string `json:"syslog_facility"`
}

type LogstashConfig struct {
	Enabled bool `json:"enabled"`
}

type OpenTelemetryConfig struct {
	TraceSampleRate       float64     `json:"trace_sample_rate"`
	MetricReportingPeriod int         `json:"metric_reporting_period"`
	Exporters             OTExporters `json:"exporters"`
	Layers                OTLayers    `json:"layers"`
}

type OTExporters struct {
	Prometheus []PrometheusExporter `json:"prometheus"`
	OTLP       []OTLPExporter       `json:"otlp"`
}

type PrometheusExporter struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type OTLPExporter struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	UseHTTP        bool   `json:"use_http"`
	DisableMetrics bool   `json:"disable_metrics"`
	DisableTraces  bool   `json:"disable_traces"`
}

type OTLayers struct {
	Global  OTLayerGlobal  `json:"global"`
	Proxy   OTLayerProxy   `json:"proxy"`
	Backend OTLayerBackend `json:"backend"`
}

type OTLayerGlobal struct {
	DisableMetrics     bool `json:"disable_metrics"`
	DisableTraces      bool `json:"disable_traces"`
	DisablePropagation bool `json:"disable_propagation"`
}

type OTLayerProxy struct {
	DisableMetrics bool `json:"disable_metrics"`
	DisableTraces  bool `json:"disable_traces"`
}

type OTLayerBackend struct {
	Metrics OTStageDetails `json:"metrics"`
	Traces  OTStageDetails `json:"traces"`
}

type OTStageDetails struct {
	DisableStage       bool     `json:"disable_stage"`
	RoundTrip          bool     `json:"round_trip"`
	ReadPayload        bool     `json:"read_payload"`
	DetailedConnection bool     `json:"detailed_connection"`
	StaticAttributes   []string `json:"static_attributes"`
}

type Endpoint struct {
	Endpoint          string                 `json:"endpoint"`
	Method            string                 `json:"method"`
	OutputEncoding    string                 `json:"output_encoding"`
	Backend           []Backend              `json:"backend"`
	InputQueryStrings []string               `json:"input_query_strings,omitempty"`
	InputHeaders      []string               `json:"input_headers,omitempty"`
	ExtraConfig       map[string]interface{} `json:"extra_config,omitempty"`
}

type Backend struct {
	SD                  string                 `json:"sd"`
	Host                []string               `json:"host"`
	URLPattern          string                 `json:"url_pattern"`
	Encoding            string                 `json:"encoding"`
	Method              string                 `json:"method"`
	DisableHostSanitize bool                   `json:"disable_host_sanitize"`
	IsCollection        bool                   `json:"is_collection,omitempty"`
	Target              string                 `json:"target,omitempty"`
	ExtraConfig         map[string]interface{} `json:"extra_config,omitempty"`
}
