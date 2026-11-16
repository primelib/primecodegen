package krakend

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
)

// operationSettings are the typed keys of the operation-level x-krakend extension
type operationSettings struct {
	Skip              bool     `yaml:"skip"`
	Path              string   `yaml:"path"`
	Timeout           string   `yaml:"timeout"`
	CacheTTL          string   `yaml:"cache_ttl"`
	ConcurrentCalls   int      `yaml:"concurrent_calls"`
	InputHeaders      []string `yaml:"input_headers"`
	InputQueryStrings []string `yaml:"input_query_strings"`
	OutputEncoding    string   `yaml:"output_encoding"`
	BackendEncoding   string   `yaml:"backend_encoding"`
	Auth              toggle   `yaml:"auth"`
	RateLimit         toggle   `yaml:"rate_limit"`
	ProxyRateLimit    toggle   `yaml:"proxy_rate_limit"`
	CircuitBreaker    toggle   `yaml:"circuit_breaker"`
	HTTPCache         toggle   `yaml:"http_cache"`
	Validation        toggle   `yaml:"validation"`
	BackendHTTP       toggle   `yaml:"backend_http"`
	ClientCredentials toggle   `yaml:"client_credentials"`
}

// endpointNamespaceShortcuts map operation x-krakend keys to endpoint extra_config namespaces, values are passed through
var endpointNamespaceShortcuts = map[string]string{
	"signer":            "auth/signer",
	"cel":               "validation/cel",
	"lua_proxy":         "modifier/lua-proxy",
	"lua_endpoint":      "modifier/lua-endpoint",
	"proxy":             "proxy",
	"req_resp_modifier": "plugin/req-resp-modifier",
	"bot_detector":      "security/bot-detector",
}

// backendNamespaceShortcuts map operation x-krakend keys to backend extra_config namespaces, values are passed through
var backendNamespaceShortcuts = map[string]string{
	"backend_cel":               "validation/cel",
	"lua_backend":               "modifier/lua-backend",
	"martian":                   "modifier/martian",
	"backend_proxy":             "proxy",
	"http_client_plugin":        "plugin/http-client",
	"backend_req_resp_modifier": "plugin/req-resp-modifier",
}

// serviceNamespaceShortcuts map root x-krakend keys to service extra_config namespaces, values are passed through
var serviceNamespaceShortcuts = map[string]string{
	"revoker":            "auth/revoker",
	"http_server_plugin": "plugin/http-server",
}

// namespaceShortcuts collects the extra_config entries for the shortcut keys present in the extension
func namespaceShortcuts(extension map[string]interface{}, shortcuts map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for key, namespace := range shortcuts {
		if value, ok := extension[key]; ok {
			out[namespace] = value
		}
	}
	return out
}

// toggle is a setting that accepts either a boolean or an object with settings (which implies enabled)
type toggle struct {
	set      bool
	enabled  bool
	settings map[string]interface{}
}

func (t *toggle) UnmarshalYAML(node *yaml.Node) error {
	t.set = true
	if node.Kind == yaml.ScalarNode {
		if err := node.Decode(&t.enabled); err != nil {
			return fmt.Errorf("line %d: expected a boolean or an object", node.Line)
		}
		return nil
	}

	settings := map[string]interface{}{}
	if err := node.Decode(&settings); err != nil {
		return fmt.Errorf("line %d: expected a boolean or an object", node.Line)
	}
	t.enabled = true
	t.settings = normalizeMap(settings)
	return nil
}

// resolve returns the toggle state, or the fallback if the toggle is not set
func (t toggle) resolve(fallback bool) bool {
	if t.set {
		return t.enabled
	}
	return fallback
}

// apply returns the defaults deep-merged with the toggle settings
func (t toggle) apply(defaults map[string]interface{}) map[string]interface{} {
	deepMergeMap(defaults, t.settings)
	return defaults
}

func parseOperationSettings(op openapigenerator.Operation) (operationSettings, error) {
	var settings operationSettings
	if op.Extensions == nil {
		return settings, nil
	}
	node, ok := op.Extensions.Get(krakendExtensionKey)
	if !ok || node == nil {
		return settings, nil
	}
	if err := node.Decode(&settings); err != nil {
		return settings, fmt.Errorf("invalid %s extension on %s %s: %w", krakendExtensionKey, strings.ToUpper(op.Method), op.Path, err)
	}
	return settings, nil
}

type exposedOperation struct {
	Operation openapigenerator.Operation
	Settings  operationSettings
}

// exposedOperations returns all operations that are not skipped, in the order of the generated endpoints
func exposedOperations(operations []openapigenerator.Operation) ([]exposedOperation, error) {
	var exposed []exposedOperation
	for _, op := range operations {
		settings, err := parseOperationSettings(op)
		if err != nil {
			return nil, err
		}
		if settings.Skip {
			continue
		}
		exposed = append(exposed, exposedOperation{Operation: op, Settings: settings})
	}
	return exposed, nil
}

// endpointPath returns the public gateway path of an operation
func endpointPath(prefix string, op openapigenerator.Operation, settings operationSettings) string {
	p := op.Path
	if settings.Path != "" {
		p = "/" + strings.TrimPrefix(settings.Path, "/")
	}
	prefix = strings.Trim(prefix, "/")
	if prefix != "" {
		p = "/" + prefix + p
	}
	return p
}

// OperationMeta holds operation details that are not part of the generator-neutral template data
type OperationMeta struct {
	Servers           openapigenerator.Endpoints // operation or path level servers
	Security          *OperationSecurity         // nil if the operation is public
	RequestBodySchema map[string]interface{}     // inlined JSON schema of a required JSON request body
}

type OperationSecurity struct {
	ForwardHeaders      []string // headers carrying credentials, e.g. Authorization or api key headers
	ForwardQueryStrings []string // query strings carrying credentials
	Bearer              bool     // a mandatory requirement uses a bearer token (oauth2, openIdConnect, http bearer)
	Scopes              []string // scopes of the first bearer requirement
	TokenURL            string   // token url of the first oauth2 client credentials flow
}

// buildOperationMeta collects per-operation servers and security requirements from the document, keyed by operationKey
func buildOperationMeta(doc *libopenapi.DocumentModel[v3.Document]) map[string]OperationMeta {
	meta := map[string]OperationMeta{}
	if doc == nil || doc.Model.Paths == nil || doc.Model.Paths.PathItems == nil {
		return meta
	}

	schemes := map[string]*v3.SecurityScheme{}
	if doc.Model.Components != nil && doc.Model.Components.SecuritySchemes != nil {
		for s := doc.Model.Components.SecuritySchemes.Oldest(); s != nil; s = s.Next() {
			schemes[s.Key] = s.Value
		}
	}

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			m := OperationMeta{}

			servers := op.Value.Servers
			if len(servers) == 0 {
				servers = path.Value.Servers
			}
			for _, s := range servers {
				if s != nil {
					m.Servers = append(m.Servers, openapigenerator.Endpoint{URL: s.URL, Description: s.Description})
				}
			}

			// a missing operation security inherits the document security, an empty list makes the operation public
			requirements := op.Value.Security
			if requirements == nil {
				requirements = doc.Model.Security
			}
			m.Security = buildOperationSecurity(requirements, schemes)
			m.RequestBodySchema = requestBodySchema(op.Value)

			meta[operationKey(op.Key, path.Key)] = m
		}
	}

	return meta
}

func buildOperationSecurity(requirements []*base.SecurityRequirement, schemes map[string]*v3.SecurityScheme) *OperationSecurity {
	sec := &OperationSecurity{}
	found := false
	optional := false

	for _, req := range requirements {
		if req == nil || req.Requirements == nil {
			optional = true
			continue
		}

		entries := 0
		for r := req.Requirements.Oldest(); r != nil; r = r.Next() {
			entries++
			scheme := schemes[r.Key]
			if scheme == nil {
				continue
			}
			found = true

			bearer := false
			switch strings.ToLower(scheme.Type) {
			case "apikey":
				switch strings.ToLower(scheme.In) {
				case "header":
					sec.ForwardHeaders = appendUnique(sec.ForwardHeaders, scheme.Name)
				case "query":
					sec.ForwardQueryStrings = appendUnique(sec.ForwardQueryStrings, scheme.Name)
				case "cookie":
					sec.ForwardHeaders = appendUnique(sec.ForwardHeaders, "Cookie")
				}
			case "http":
				sec.ForwardHeaders = appendUnique(sec.ForwardHeaders, "Authorization")
				bearer = strings.EqualFold(scheme.Scheme, "bearer")
			case "oauth2", "openidconnect":
				sec.ForwardHeaders = appendUnique(sec.ForwardHeaders, "Authorization")
				bearer = true
				if sec.TokenURL == "" && scheme.Flows != nil && scheme.Flows.ClientCredentials != nil {
					sec.TokenURL = scheme.Flows.ClientCredentials.TokenUrl
				}
			}

			if bearer && !sec.Bearer {
				sec.Bearer = true
				sec.Scopes = r.Value
			}
		}
		// an empty requirement ({}) allows anonymous access
		if entries == 0 {
			optional = true
		}
	}

	if !found {
		return nil
	}
	if optional {
		sec.Bearer = false
		sec.Scopes = nil
	}
	return sec
}

// requestBodySchema returns the inlined schema of a required JSON request body, nil if there is none or it can not be inlined
func requestBodySchema(op *v3.Operation) map[string]interface{} {
	if op == nil || op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}
	if op.RequestBody.Required == nil || !*op.RequestBody.Required {
		return nil // krakend rejects empty bodies when a schema is set
	}

	for c := op.RequestBody.Content.Oldest(); c != nil; c = c.Next() {
		if !strings.Contains(strings.ToLower(c.Key), "json") || c.Value == nil || c.Value.Schema == nil {
			continue
		}
		schema := c.Value.Schema.Schema()
		if schema == nil {
			return nil
		}
		rendered, err := schema.RenderInline()
		if err != nil {
			slog.Warn("skipping request body validation, schema can not be inlined", "operation", op.OperationId, "err", err)
			return nil
		}
		out := map[string]interface{}{}
		if err := yaml.Unmarshal(rendered, &out); err != nil {
			return nil
		}
		return normalizeMap(out)
	}

	return nil
}

// clientCredentialsConfig builds the backend auth/client-credentials config, nil if the gateway should not fetch tokens
func clientCredentialsConfig(root map[string]interface{}, sec *OperationSecurity, opClientCredentials toggle) map[string]interface{} {
	if opClientCredentials.set && !opClientCredentials.enabled {
		return nil
	}
	if !opClientCredentials.set && (len(root) == 0 || sec == nil || sec.TokenURL == "") {
		return nil
	}

	conf := map[string]interface{}{}
	if sec != nil && sec.TokenURL != "" {
		conf["token_url"] = sec.TokenURL
	}
	deepMergeMap(conf, root)
	deepMergeMap(conf, opClientCredentials.settings)
	if _, ok := conf["token_url"]; !ok {
		return nil
	}
	return conf
}

// operationBackend resolves the operation-level server for an environment, if the operation defines servers
func operationBackend(servers openapigenerator.Endpoints, environment string) (ResolvedEndpoint, bool) {
	if len(servers) == 0 {
		return ResolvedEndpoint{}, false
	}
	if resolved, ok := resolveEnvironmentEndpoints(servers)[environment]; ok {
		return resolved, true
	}

	// a single server without environment marker applies to all environments
	if len(servers) == 1 && detectEnvironment(servers[0]) == "" {
		if normalized, ok := normalizeEndpointURL(servers[0].URL); ok {
			return ResolvedEndpoint{
				Host:        normalized.Scheme + "://" + normalized.Host,
				BasePath:    normalized.BasePath,
				RawURL:      strings.TrimSpace(servers[0].URL),
				Environment: environment,
			}, true
		}
	}

	return ResolvedEndpoint{}, false
}

// authValidatorConfig builds the endpoint auth/validator config, nil if the endpoint should not validate tokens
func authValidatorConfig(rootAuth map[string]interface{}, sec *OperationSecurity, opAuth toggle) map[string]interface{} {
	if opAuth.set && !opAuth.enabled {
		return nil
	}
	if !opAuth.set && (sec == nil || !sec.Bearer) {
		return nil
	}
	if len(rootAuth) == 0 && len(opAuth.settings) == 0 {
		return nil
	}

	validator := map[string]interface{}{}
	deepMergeMap(validator, rootAuth)
	if sec != nil && len(sec.Scopes) > 0 {
		validator["scopes"] = sec.Scopes
		validator["scopes_matcher"] = "all"
		if _, ok := validator["scopes_key"]; !ok {
			validator["scopes_key"] = "scope"
		}
	}
	deepMergeMap(validator, opAuth.settings)
	return validator
}

// mergeInputList appends the extra entries, a "*" entry forwards everything
func mergeInputList(list []string, extra []string) []string {
	for _, e := range extra {
		if strings.TrimSpace(e) == "*" {
			return []string{"*"}
		}
	}
	for _, e := range extra {
		list = appendUnique(list, strings.TrimSpace(e))
	}
	return list
}

func appendUnique(list []string, values ...string) []string {
	for _, v := range values {
		if v == "" {
			continue
		}
		exists := false
		for _, existing := range list {
			if strings.EqualFold(existing, v) {
				exists = true
				break
			}
		}
		if !exists {
			list = append(list, v)
		}
	}
	return list
}
