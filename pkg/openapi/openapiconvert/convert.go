package openapiconvert

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func ConvertSwaggerToOpenAPI(swaggerData []byte, converterUrl string) ([]byte, error) {

	swaggerConverterEnvVar := "PRIMECODEGN_SWAGGER_CONVERTER"
	var url string
	var urlEnvIsSet bool

	if converterUrl == "" {
		url, urlEnvIsSet = os.LookupEnv(swaggerConverterEnvVar)
		log.Trace().Bool("Env var is present", urlEnvIsSet).Str("URL", url).Msg("Converter from env: ")
		if !urlEnvIsSet {
			// URL of public Swagger Converter
			url = "https://converter.swagger.io/api/convert"
		}
	} else {
		url = converterUrl
	}
	log.Debug().Str("URL", url).Msg("Swagger 2.0 OpenAPI 3.0 converter used: ")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(swaggerData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
