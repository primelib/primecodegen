package openapipatch

import (
	"errors"
	"fmt"
	"os"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/primelib/primecodegen/pkg/tools/speakeasycli"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func ApplyPatches(input []byte, patches []sharedpatch.SpecPatch) ([]byte, error) {
	for _, p := range patches {
		log.Info().Str("id", p.String()).Msg("applying patch to spec")

		if p.Type == "builtin" {
			if patcher, ok := EmbeddedPatcherMap[p.Type+":"+p.ID]; ok {
				doc, err := openapidocument.OpenDocument(input)
				if err != nil {
					return input, err
				}

				v3doc, errs := doc.BuildV3Model()
				if len(errs) > 0 {
					return input, fmt.Errorf("failed to build v3 high level model: %w", errors.Join(errs...))
				}

				patchErr := patcher.Func(v3doc, p.Config)
				if patchErr != nil {
					return input, fmt.Errorf("failed to patch document with [%s]: %w", patcher.ID, patchErr)
				}

				bytes, err := loader.InterfaceToYaml(v3doc.Model)
				if err != nil {
					return input, errors.Join(util.ErrRenderDocument, err)
				}

				input = bytes
			} else {
				return input, errors.Join(util.ErrFailedToPatchDocument, fmt.Errorf("builtin patch [%s] is not supported", p))
			}
		} else if p.Type == "speakeasy" {
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

			patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(tempFile.Name(), p.File)
			if patchErr != nil {
				return input, errors.Join(util.ErrFailedToPatchDocument, patchErr)
			}

			input = patchedBytes
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
