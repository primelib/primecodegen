package template

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	openapi_go_httpclient "github.com/primelib/primecodegen/pkg/template/templates/openapi-go-httpclient"
	openapi_java_httpclient "github.com/primelib/primecodegen/pkg/template/templates/openapi-java-httpclient"
)

var defaultSnippets = []string{"global-layout.gohtml"}

var allTemplates = map[string]templateapi.Config{
	openapi_go_httpclient.Template.ID:   openapi_go_httpclient.Template,
	openapi_java_httpclient.Template.ID: openapi_java_httpclient.Template,
}
