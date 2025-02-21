package openapiconvert

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

const (
	converterEndpointEnvVar = "PRIMECODEGEN_SWAGGER_CONVERTER"
	converterEndpoint       = "https://converter.swagger.io/api/convert"
)

// ConvertSpec converts an input specification file to the desired output format.
func ConvertSpec(inputPath, formatIn, formatOut string) ([]byte, error) {
	var result []byte

	// read file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", inputPath, err)
	}

	// convert
	if formatIn == "swagger20" && (formatOut == "openapi30" || formatOut == "openapi30-json") {
		openAPIJSON, err := ConvertSwaggerToOpenAPI30(data, "")
		if err != nil {
			return nil, fmt.Errorf("error converting to OpenAPI 3.0: %w", err)
		}
		result = openAPIJSON

		// convert to YAML if needed
		if formatOut == "openapi30" {
			result, err = util.JSONToYAML(result)
			if err != nil {
				return result, err
			}
		}
	} else {
		return nil, fmt.Errorf("unsupported conversion: %s to %s", formatIn, formatOut)
	}

	return result, nil
}

func ConvertSwaggerToOpenAPI30(swaggerData []byte, converterUrl string) ([]byte, error) {
	if converterUrl == "" {
		converterUrl, _ = os.LookupEnv(converterEndpointEnvVar)
		if converterUrl == "" {
			converterUrl = converterEndpoint
		}
	}
	log.Debug().Str("url", converterUrl).Msg("Using swagger converter endpoint for openapi conversion")

	client := &http.Client{}
	req, err := http.NewRequest("POST", converterUrl, bytes.NewBuffer(swaggerData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	openapiData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return openapiData, nil
}
