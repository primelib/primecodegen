package openapiconvert

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/primelib/primecodegen/pkg/tools/speakeasycli"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

const (
	converterEndpointEnvVar = "PRIMECODEGEN_SWAGGER_CONVERTER"
	converterEndpoint       = "https://converter.swagger.io/api/convert"
)

var (
	ErrInvalidInputFormat    = fmt.Errorf("invalid input format")
	ErrInvalidOutputFormat   = fmt.Errorf("invalid output format")
	ErrUnsupportedConversion = fmt.Errorf("unsupported conversion")
	SupportedInputFormats    = []string{"swagger20"}
	SupportedOutputFormats   = []string{"openapi30", "openapi30-json"}
)

// ConvertSpec converts an input specification file to the desired output format.
func ConvertSpec(inputPath, formatIn, formatOut, converter string) ([]byte, error) {
	var result []byte

	// validate parameters
	if !slices.Contains(SupportedInputFormats, formatIn) {
		return nil, errors.Join(ErrInvalidInputFormat, fmt.Errorf("unsupported format: %s, supported are %s", formatOut, strings.Join(SupportedInputFormats, ", ")))
	}
	if !slices.Contains(SupportedOutputFormats, formatOut) {
		return nil, errors.Join(ErrInvalidOutputFormat, fmt.Errorf("unsupported output format: %s, supported are %s", formatOut, strings.Join(SupportedOutputFormats, ", ")))
	}

	// read file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", inputPath, err)
	}

	// convert
	if formatIn == "swagger20" && strings.HasPrefix(formatOut, "openapi30") {
		if converter == "" || converter == "openapi-converter" {
			result, err = ConvertSwaggerToOpenAPIUsingSwaggerConverter(data, converter)
			if err != nil {
				return result, err
			}
		} else if converter == "speakeasy" {
			result, err = ConvertSwaggerToOpenAPIUsingSpeakeasy(data)
			if err != nil {
				return result, err
			}
		} else {
			return nil, errors.Join(ErrUnsupportedConversion, fmt.Errorf("converter %s does not exist", converter))
		}

		// ConvertSwaggerToOpenAPIUsingSwaggerConverter returns json, convert to YAML if needed
		if formatOut == "openapi30" {
			result, err = util.JSONToYAML(result)
			if err != nil {
				return result, err
			}
		}
	} else {
		return nil, errors.Join(ErrUnsupportedConversion, fmt.Errorf("from %s to %s", formatIn, formatOut))
	}

	return result, nil
}

func ConvertSwaggerToOpenAPIUsingSwaggerConverter(swaggerData []byte, converterUrl string) ([]byte, error) {
	if converterUrl == "" {
		converterUrl, _ = os.LookupEnv(converterEndpointEnvVar)
		if converterUrl == "" {
			converterUrl = converterEndpoint
		}
	}
	log.Debug().Str("url", converterUrl).Msg("Using swagger converter endpoint for openapi conversion")

	req, err := http.NewRequest("POST", converterUrl, bytes.NewBuffer(swaggerData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func ConvertSwaggerToOpenAPIUsingSpeakeasy(swaggerData []byte) ([]byte, error) {
	// write to temp file
	tempFile, err := os.CreateTemp("", "swagger-*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(swaggerData)
	if err != nil {
		return nil, err
	}

	err = tempFile.Close()
	if err != nil {
		return nil, err
	}

	// convert
	output, err := speakeasycli.SpeakEasySwaggerConvertCommand(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return output, nil
}
