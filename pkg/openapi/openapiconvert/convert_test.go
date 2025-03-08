package openapiconvert

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestConvertSwaggerToOpenAPI(t *testing.T) {
	// arrange
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterResponder("POST", converterEndpoint, httpmock.NewStringResponder(200, `{"openapi": "3.0.0"}`))
	swaggerData := []byte(`{"swagger": "2.0"}`)

	// act
	result, err := ConvertSwaggerToOpenAPIUsingSwaggerConverter(swaggerData, "")

	// assert
	assert.NoError(t, err)
	assert.JSONEq(t, `{"openapi": "3.0.0"}`, string(result))
}
