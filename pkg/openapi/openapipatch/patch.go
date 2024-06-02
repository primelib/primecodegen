package openapipatch

import (
	"errors"
	"fmt"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rs/zerolog/log"
)

// PatchV3 applies a list of patches to the given document and returns the modified document
func PatchV3(patchIds []string, doc libopenapi.Document, v3doc *libopenapi.DocumentModel[v3.Document]) (libopenapi.Document, *libopenapi.DocumentModel[v3.Document], error) {
	for _, patchId := range patchIds {
		if patch, ok := V3Patchers[patchId]; ok {
			log.Debug().Str("id", patch.ID).Msg("running spec patcher")
			patchErr := patch.Func(v3doc)
			if patchErr != nil {
				return doc, v3doc, fmt.Errorf("failed to patch document with [%s]: %w", patch.ID, patchErr)
			}

			// reload document
			var errs []error
			_, doc, _, errs = doc.RenderAndReload()
			if len(errs) > 0 {
				return doc, v3doc, fmt.Errorf("failed to reload document after patching: %w", errors.Join(errs...))
			}
			v3doc, errs = doc.BuildV3Model()
			if len(errs) > 0 {
				return doc, v3doc, fmt.Errorf("failed to build v3 high level model: %w", errors.Join(errs...))
			}
		} else {
			return doc, v3doc, fmt.Errorf("patch with given id not found: %s", patchId)
		}
	}

	// reload document
	_, _, _, _ = doc.RenderAndReload()

	return doc, v3doc, nil
}
