package openapipatch

import (
	"errors"
	"fmt"
	"os"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func ApplyPatches(input []byte, patches []sharedpatch.SpecPatch) ([]byte, error) {
	inputFormat := util.DetectJSONOrYAML(input)

	for _, p := range patches {
		log.Info().Str("id", p.String()).Msg("applying patch to spec")

		if patcher, ok := EmbeddedPatcherMap[p.Type+":"+p.ID]; ok {
			// In-Memory Patcher (libopenapi)
			if patcher.PatchV3DocumentFunc != nil {
				doc, err := openapidocument.OpenDocument(input)
				if err != nil {
					return input, err
				}

				v3doc, err := doc.BuildV3Model()
				if err != nil {
					return input, fmt.Errorf("failed to build v3 high level model: %w", err)
				}

				patchErr := patcher.PatchV3DocumentFunc(v3doc, p.Config)
				if patchErr != nil {
					return input, fmt.Errorf("failed to patch document with [%s]: %w", patcher.ID, patchErr)
				}

				bytes, err := openapidocument.RenderV3ModelFormat(v3doc, inputFormat)
				if err != nil {
					return input, errors.Join(util.ErrRenderDocument, err)
				}
				input = bytes
				continue
			}

			// File-based Patch (external tool call)
			if patcher.PatchFileFunc != nil {
				tempFile, err := os.CreateTemp("", "input-*.yaml")
				if err != nil {
					return input, errors.Join(util.ErrFailedToPatchDocument, err)
				}
				defer os.Remove(tempFile.Name())

				_, err = tempFile.Write(input)
				if err != nil {
					return input, errors.Join(util.ErrFailedToPatchDocument, err)
				}
				err = tempFile.Close()
				if err != nil {
					return input, errors.Join(util.ErrFailedToPatchDocument, err)
				}

				patchedBytes, patchErr := patcher.PatchFileFunc(tempFile.Name(), p.Config)
				if patchErr != nil {
					return input, patchErr
				}

				input = patchedBytes
				continue
			}
		} else {
			patchedBytes, patchErr := patch.ApplyPatchFile(input, p)
			if patchErr != nil {
				return input, errors.Join(util.ErrFailedToPatchDocument, patchErr)
			}

			input = patchedBytes
		}
	}

	return input, nil
}
