package openapipatch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func ApplyPatches(input []byte, patches []string) ([]byte, error) {
	for _, p := range patches {
		// file-based patch
		if strings.Contains(p, ":") {
			parts := strings.SplitN(p, ":", 2)
			if len(parts) != 2 {
				return input, errors.New("invalid patch file syntax")
			}
			patchType := parts[0]
			patchFile := parts[1]

			log.Debug().Str("patchType", patchType).Str("patchFile", patchFile).Msg("applying patch to spec")
			patchedBytes, patchErr := patch.ApplyPatchFile(input, patchType, patchFile)
			if patchErr != nil {
				return input, errors.Join(util.ErrFailedToPatchDocument, patchErr)
			}

			input = patchedBytes
			continue
		}

		// builtin patch
		if patcher, ok := V3Patchers[p]; ok {
			log.Debug().Str("id", patcher.ID).Msg("applying patch to spec")

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
		}
	}

	return input, nil
}
