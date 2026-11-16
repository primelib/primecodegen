package krakend

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
)

const (
	EnvironmentDev   = "dev"
	EnvironmentStage = "stage"
	EnvironmentProd  = "prod"
)

var environmentOrder = []string{EnvironmentDev, EnvironmentStage, EnvironmentProd}

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

type ResolvedEndpoint struct {
	Host        string // host including the scheme, e.g. https://api.example.com
	BasePath    string
	RawURL      string
	Environment string
}

func resolveEnvironmentEndpoints(endpoints openapigenerator.Endpoints) map[string]ResolvedEndpoint {
	resolved := map[string]ResolvedEndpoint{}

	for _, ep := range endpoints {
		environment := detectEnvironment(ep)
		if environment == "" {
			continue
		}

		normalized, ok := normalizeEndpointURL(ep.URL)
		if !ok || normalized.Host == "" {
			continue
		}

		if _, exists := resolved[environment]; exists {
			continue
		}

		resolved[environment] = ResolvedEndpoint{
			Host:        normalized.Scheme + "://" + normalized.Host,
			BasePath:    normalized.BasePath,
			RawURL:      strings.TrimSpace(ep.URL),
			Environment: environment,
		}
	}

	return resolved
}

type normalizedEndpoint struct {
	Scheme   string
	Host     string
	BasePath string
}

func normalizeEndpointURL(raw string) (normalizedEndpoint, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return normalizedEndpoint{}, false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return normalizedEndpoint{}, false
	}

	host := strings.TrimSpace(parsed.Host)
	if host == "" {
		return normalizedEndpoint{}, false
	}

	basePath := strings.TrimSpace(parsed.Path)
	basePath = strings.TrimSuffix(basePath, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme == "" {
		scheme = "https"
	}

	return normalizedEndpoint{Scheme: scheme, Host: host, BasePath: basePath}, true
}

func detectEnvironment(endpoint openapigenerator.Endpoint) string {
	content := strings.ToLower(strings.TrimSpace(endpoint.URL + " " + endpoint.Description))
	if content == "" {
		return ""
	}

	tokens := map[string]bool{}
	for _, token := range nonAlphaNumRegex.Split(content, -1) {
		if token != "" {
			tokens[token] = true
		}
	}

	// also tokenize without separators to match spellings like pre-production / pre_prod
	normalized := strings.NewReplacer("-", "", "_", "").Replace(content)
	for _, token := range nonAlphaNumRegex.Split(normalized, -1) {
		if token == "preprod" || token == "preproduction" {
			tokens[token] = true
		}
	}

	if tokens["dev"] || tokens["development"] {
		return EnvironmentDev
	}

	if tokens["stage"] || tokens["staging"] || tokens["uat"] || tokens["preprod"] || tokens["preproduction"] {
		return EnvironmentStage
	}

	if tokens["prod"] || tokens["production"] || tokens["live"] {
		return EnvironmentProd
	}

	return ""
}
