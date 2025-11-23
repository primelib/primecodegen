package template

import (
	"github.com/primelib/primecodegen/pkg/template/templateapi"
	openapi_default_scaffolding "github.com/primelib/primecodegen/pkg/template/templates/openapi-default-scaffolding"
	openapi_go_httpclient "github.com/primelib/primecodegen/pkg/template/templates/openapi-go-httpclient"
	openapi_java_httpclient "github.com/primelib/primecodegen/pkg/template/templates/openapi-java-httpclient"
	openapi_kotlin_httpclient "github.com/primelib/primecodegen/pkg/template/templates/openapi-kotlin-httpclient"
)

var defaultSnippets = []string{"global-layout.gohtml"}

var allTemplates = map[string]templateapi.Config{
	openapi_go_httpclient.Template.ID:       openapi_go_httpclient.Template,
	openapi_java_httpclient.Template.ID:     openapi_java_httpclient.Template,
	openapi_kotlin_httpclient.Template.ID:   openapi_kotlin_httpclient.Template,
	openapi_default_scaffolding.Template.ID: openapi_default_scaffolding.Template,
}
