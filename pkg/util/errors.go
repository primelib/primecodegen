package util

import (
	"fmt"
)

var (
	ErrReadDocumentFromFile   = fmt.Errorf("file is missing")
	ErrOpenDocument           = fmt.Errorf("failed to open document")
	ErrNoFilesSpecified       = fmt.Errorf("no files specified")
	ErrDocumentMerge          = fmt.Errorf("failed to merge documents")
	ErrFailedToPatchDocument  = fmt.Errorf("failed to patch document")
	ErrRenderDocument         = fmt.Errorf("failed to render document")
	ErrGenerateOpenAPIV3Model = fmt.Errorf("failed to generate openapi v3 model")
	ErrWriteDocumentToFile    = fmt.Errorf("failed to write document to file")
	ErrNoGeneratorWithId      = fmt.Errorf("no generator with specified id")
	ErrWriteDocumentToStdout  = fmt.Errorf("failed to write document to stdout")
	ErrJSONMarshal            = fmt.Errorf("failed to marshal into JSON")
	ErrSwagger2OpenAPI30      = fmt.Errorf("failed to convert API spec from Swagger 2.0 to OpenAPI 3.0")
)
