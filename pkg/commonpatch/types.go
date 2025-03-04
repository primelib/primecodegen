package commonpatch

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrUnsupportedPatchType = fmt.Errorf("unsupported patch type")
)

type PatchType string

const (
	PatchTypeJSONPatch      PatchType = "jsonpatch"
	PatchTypeGitPatch       PatchType = "git"
	PatchTypeOpenAPIOverlay PatchType = "openapi-overlay"
)

type PatchFile func(input []byte, patch []byte) ([]byte, error)

func ApplyPatch(patchType PatchType, input []byte, patchContent []byte) ([]byte, error) {
	switch patchType {
	case PatchTypeJSONPatch:
		return ApplyJSONPatch(input, patchContent)
	case PatchTypeGitPatch:
		return ApplyGitPatch(input, patchContent)
	case PatchTypeOpenAPIOverlay:
		return ApplyOpenAPIOverlay(input, patchContent)
	default:
		return nil, errors.Join(ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}
}

func ApplyPatchFile(input []byte, patchType string, patchFile string) ([]byte, error) {
	// read content
	content, err := os.ReadFile(patchFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// to enum
	var patchTypeEnum PatchType
	switch patchType {
	case "file":
		if strings.HasSuffix(patchFile, ".patch") {
			patchTypeEnum = PatchTypeGitPatch
		} else if strings.HasSuffix(patchFile, ".jsonpatch") {
			patchTypeEnum = PatchTypeJSONPatch
		} else {
			return nil, errors.Join(ErrUnsupportedPatchType, fmt.Errorf("file: %s", patchFile))
		}
	case string(PatchTypeJSONPatch), string(PatchTypeGitPatch), string(PatchTypeOpenAPIOverlay):
		patchTypeEnum = PatchType(patchType)
	default:
		return nil, errors.Join(ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}

	// process
	return ApplyPatch(patchTypeEnum, input, content)
}
