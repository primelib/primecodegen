package patch

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/patch/gitpatch"
	"github.com/primelib/primecodegen/pkg/patch/jsonpatch"
	"github.com/primelib/primecodegen/pkg/patch/openapioverlay"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
)

func ApplyPatch(patchType sharedpatch.PatchType, input []byte, patchContent []byte) ([]byte, error) {
	switch patchType {
	case sharedpatch.PatchTypeJSONPatch:
		return jsonpatch.ApplyJSONPatch(input, patchContent)
	case sharedpatch.PatchTypeGitPatch:
		return gitpatch.ApplyGitPatch(input, patchContent)
	case sharedpatch.PatchTypeOpenAPIOverlay:
		return openapioverlay.ApplyOpenAPIOverlay(input, patchContent)
	default:
		return nil, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}
}

func ApplyPatchFile(input []byte, patch sharedpatch.SpecPatch) ([]byte, error) {
	var content []byte
	var err error
	if patch.Content != "" {
		content = []byte(patch.Content)
	} else if patch.File != "" {
		content, err = os.ReadFile(patch.File)
		if err != nil {
			return nil, errors.Join(sharedpatch.ErrFailedToReadPatchFile, err)
		}
	} else {
		return nil, sharedpatch.ErrExternalPatchMustHaveContentOrFile
	}

	// to enum
	var patchTypeEnum sharedpatch.PatchType
	switch patch.Type {
	case "file":
		if strings.HasSuffix(patch.File, ".jsonpatch") {
			patchTypeEnum = sharedpatch.PatchTypeJSONPatch
		} else if strings.HasSuffix(patch.File, ".patch") {
			patchTypeEnum = sharedpatch.PatchTypeGitPatch
		} else {
			return nil, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("file: %s", patch.File))
		}
	case string(sharedpatch.PatchTypeJSONPatch), string(sharedpatch.PatchTypeGitPatch), string(sharedpatch.PatchTypeOpenAPIOverlay):
		patchTypeEnum = sharedpatch.PatchType(patch.Type)
	default:
		return nil, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patch.Type))
	}

	// process
	return ApplyPatch(patchTypeEnum, input, content)
}

func ValidatePatch(patchType sharedpatch.PatchType, patchContent []byte) error {
	switch patchType {
	case sharedpatch.PatchTypeJSONPatch:
		return jsonpatch.ValidateJSONPatch(patchContent)
	case sharedpatch.PatchTypeGitPatch:
		return gitpatch.ValidateGitPatch(patchContent)
	case sharedpatch.PatchTypeOpenAPIOverlay:
		return openapioverlay.ValidateOpenAPIOverlay(patchContent)
	default:
		return errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}
}

func ValidatePatchFile(patchType string, patchFile string) error {
	// read content
	content, err := os.ReadFile(patchFile)
	if err != nil {
		return errors.Join(sharedpatch.ErrFailedToReadPatchFile, err)
	}

	// to enum
	var patchTypeEnum sharedpatch.PatchType
	switch patchType {
	case "file":
		if strings.HasSuffix(patchFile, ".jsonpatch") {
			patchTypeEnum = sharedpatch.PatchTypeJSONPatch
		} else if strings.HasSuffix(patchFile, ".patch") {
			patchTypeEnum = sharedpatch.PatchTypeGitPatch
		} else {
			return errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("file: %s", patchFile))
		}
	case string(sharedpatch.PatchTypeJSONPatch), string(sharedpatch.PatchTypeGitPatch), string(sharedpatch.PatchTypeOpenAPIOverlay):
		patchTypeEnum = sharedpatch.PatchType(patchType)
	default:
		return errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}

	// process
	return ValidatePatch(patchTypeEnum, content)
}

func NewContentPatch(patchType sharedpatch.PatchType, content interface{}) (sharedpatch.SpecPatch, error) {
	switch patchType {
	case sharedpatch.PatchTypeOpenAPIOverlay:
		bytes, err := loader.InterfaceToYaml(content)
		if err != nil {
			return sharedpatch.SpecPatch{}, err
		}

		return sharedpatch.SpecPatch{
			Type:    string(patchType),
			Content: string(bytes),
		}, nil
	default:
		return sharedpatch.SpecPatch{}, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}
}
