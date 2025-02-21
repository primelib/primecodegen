package openapidocument

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rs/zerolog/log"
)

func EmptyDocument() libopenapi.DocumentModel[v3.Document] {
	doc, err := OpenDocument([]byte("openapi: 3.0.0"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create empty document")
	}
	v3doc, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		log.Fatal().Errs("spec", errs).Msgf("failed to create empty v3 document")
	}
	return *v3doc
}
