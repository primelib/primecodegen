package openapiconvert

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

const (
	converterEndpointEnvVar = "PRIMECODEGEN_SWAGGER_CONVERTER"
	converterEndpoint       = "https://converter.swagger.io/api/convert"
)

func ConvertSwaggerToOpenAPI(swaggerData []byte, converterUrl string) ([]byte, error) {
	if converterUrl == "" {
		converterUrl, _ = os.LookupEnv(converterEndpointEnvVar)
		if converterUrl == "" {
			converterUrl = converterEndpoint
		}
	}
	log.Debug().Str("endpoint", converterUrl).Msg("Converting Swagger to OpenAPI using external service")

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
