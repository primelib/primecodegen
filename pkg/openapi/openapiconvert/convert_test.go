package openapiconvert

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConvertSwaggerToOpenAPI(t *testing.T) {

	// arrange
	mockClient := new(MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"openapi": "3.0.0"}`)),
	}
	mockClient.On("Do", mock.Anything).Return(mockResponse, nil)
	swaggerData := []byte(`{"swagger": "2.0"}`)
	converterUrl := "http://mock-converter-url"

	// act
	result, err := ConvertSwaggerToOpenAPI(swaggerData, converterUrl, mockClient)

	// assert
	assert.NoError(t, err)
	assert.JSONEq(t, `{"openapi": "3.0.0"}`, string(result))

	mockClient.AssertExpectations(t)
}
