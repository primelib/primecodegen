package openapiconvert

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RealHTTPClient struct{}

func (c *RealHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
