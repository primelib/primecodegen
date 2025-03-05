package patch

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

func ApplyPatchFile(input []byte, patchType string, patchFile string) ([]byte, error) {
	// read content
	content, err := os.ReadFile(patchFile)
	if err != nil {
		return nil, errors.Join(sharedpatch.ErrFailedToReadPatchFile, err)
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
			return nil, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("file: %s", patchFile))
		}
	case string(sharedpatch.PatchTypeJSONPatch), string(sharedpatch.PatchTypeGitPatch), string(sharedpatch.PatchTypeOpenAPIOverlay):
		patchTypeEnum = sharedpatch.PatchType(patchType)
	default:
		return nil, errors.Join(sharedpatch.ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
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
