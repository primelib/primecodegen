package openapidocument

type SourceType string

const (
	SourceTypeSpec      SourceType = "spec"
	SourceTypeSwaggerUI SourceType = "swagger-ui"
	SourceTypeRedoc     SourceType = "redoc"
)

type SpecType string

const (
	SpecTypeOpenAPI3 SpecType = "openapi3"
	SpecTypeSwagger2 SpecType = "swagger2"
)
