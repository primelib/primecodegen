package openapidocument

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func OpenDocumentFile(file string) (libopenapi.Document, error) {
	// read the file
	input, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	// config
	conf := datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BasePath:              filepath.Dir(file),
		//BaseURL:               baseURL,
	}

	// create a new document from specification bytes
	document, err := libopenapi.NewDocumentWithConfiguration(input, &conf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create document from spec")
	}

	return document, nil
}

func OpenDocument(input []byte) (libopenapi.Document, error) {
	return OpenDocumentWithBaseDir(input, "")
}

func OpenDocumentWithBaseDir(input []byte, baseDir string) (libopenapi.Document, error) {
	// config
	conf := datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BasePath:              baseDir,
	}

	// create a new document from specification bytes
	document, err := libopenapi.NewDocumentWithConfiguration(input, &conf)
	if err != nil {
		return nil, errors.Join(util.ErrOpenDocument, err)
	}

	return document, nil
}

// OpenV3DocumentForTest opens a document and builds a v3 model, asserting no errors (panics on error, for usage in tests only!)
func OpenV3DocumentForTest(bytes []byte) *libopenapi.DocumentModel[v3.Document] {
	document, err := OpenDocument(bytes)
	if err != nil {
		panic(err)
	}
	v3doc, err := document.BuildV3Model()
	if err != nil {
		panic(err)
	}

	return v3doc
}

// RenderDocument renders the document as bytes
func RenderDocument(doc libopenapi.Document) ([]byte, error) {
	bytes, err := doc.Render()
	if err != nil {
		return nil, fmt.Errorf("failed to render document: %w", err)
	}

	return bytes, nil
}

// RenderDocumentFile writes the document into a file
func RenderDocumentFile(doc libopenapi.Document, file string) error {
	bytes, err := doc.Render()
	if err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	if file == "" {
		return fmt.Errorf("output file is required")
	}

	err = os.WriteFile(file, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write document to file: %w", err)
	}
	return nil
}

func EmptyDocument() libopenapi.DocumentModel[v3.Document] {
	doc, err := OpenDocument([]byte("openapi: 3.0.0"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create empty document")
	}
	v3doc, err := doc.BuildV3Model()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create empty v3 document")
	}
	return *v3doc
}
