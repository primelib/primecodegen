package openapipatch

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/tools/speakeasycli"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func ApplyPatches(input []byte, patches []string) ([]byte, error) {
	for _, p := range patches {
		var patchType string
		var patchFile string
		if strings.Contains(p, ":") {
			parts := strings.SplitN(p, ":", 2)
			if len(parts) != 2 {
				return input, errors.New("invalid patch file syntax")
			}
			patchType = parts[0]
			patchFile = parts[1]
		} else {
			patchType = "builtin"
			patchFile = p
		}
		log.Debug().Str("patchType", patchType).Str("patchFile", patchFile).Msg("applying patch to spec")

		if patchType == "builtin" {
			if patcher, ok := V3Patchers[p]; ok {
				doc, err := openapidocument.OpenDocument(input)
				if err != nil {
					return input, err
				}

				v3doc, errs := doc.BuildV3Model()
				if len(errs) > 0 {
					return input, fmt.Errorf("failed to build v3 high level model: %w", errors.Join(errs...))
				}

				patchErr := patcher.Func(v3doc)
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
		} else if patchType == "speakeasy" {
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

			patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(tempFile.Name(), patchFile)
			if patchErr != nil {
				return input, errors.Join(util.ErrFailedToPatchDocument, patchErr)
			}

			input = patchedBytes
		} else {
			patchedBytes, patchErr := patch.ApplyPatchFile(input, patchType, patchFile)
			if patchErr != nil {
				return input, errors.Join(util.ErrFailedToPatchDocument, patchErr)
			}

			input = patchedBytes
		}
	}

	return input, nil
}
