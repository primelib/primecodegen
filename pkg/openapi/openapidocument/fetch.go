package openapidocument

import (
	"fmt"
	"regexp"

	"github.com/primelib/primecodegen/pkg/util"
)

// FetchSpec will download the spec from the source
func FetchSpec(format SourceType, url string) ([]byte, error) {
	switch format {
	case "", SourceTypeSpec:
		return fetchSpecFromURL(url)
	case SourceTypeSwaggerUI:
		return fetchSpecFromSwaggerUI(url)
	case SourceTypeRedoc:
		return fetchSpecFromRedoc(url)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", format)
	}
}

func fetchSpecFromURL(url string) ([]byte, error) {
	content, err := util.DownloadBytes(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download spec source: %w", err)
	}
	return content, nil
}

func fetchSpecFromSwaggerUI(url string) ([]byte, error) {
	swaggerJsUrl := url + "/swagger-ui-init.js"
	content, err := util.DownloadBytes(swaggerJsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to download spec source: %w", err)
	}

	re := regexp.MustCompile(`"swaggerDoc":([\S\s]*?),[\n\s]*"customOptions"`)
	match := re.FindStringSubmatch(string(content))
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to extract spec from swagger-ui-init.js")
	}

	return []byte(match[1]), nil
}

func fetchSpecFromRedoc(url string) ([]byte, error) {
	content, err := util.DownloadBytes(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download Redoc page: %w", err)
	}

	// Regex to extract OpenAPI spec from Redoc's state variable
	re := regexp.MustCompile(`const __redoc_state = .+"data":([\S\s]*?)},"searchIndex`) // Extract JSON payload
	match := re.FindStringSubmatch(string(content))
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to extract OpenAPI spec from Redoc page")
	}

	return []byte(match[1]), nil
}
