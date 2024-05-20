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
	PatchTypeJSONPatch PatchType = "jsonpatch"
	PatchTypeGitPatch  PatchType = "git"
)

type PatchFile func(input []byte, patch []byte) ([]byte, error)

func ApplyPatch(patchType PatchType, input []byte, patchContent []byte) ([]byte, error) {
	switch patchType {
	case PatchTypeJSONPatch:
		return ApplyJSONPatch(input, patchContent)
	case PatchTypeGitPatch:
		return ApplyGitPatch(input, patchContent)
	default:
		return nil, errors.Join(ErrUnsupportedPatchType, fmt.Errorf("type: %s", patchType))
	}
}

func ApplyPatchFile(input []byte, patchFile string) ([]byte, error) {
	// read content
	content, err := os.ReadFile(patchFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// process
	if strings.HasSuffix(patchFile, ".patch") {
		return ApplyPatch(PatchTypeGitPatch, input, content)
	} else if strings.HasSuffix(patchFile, ".jsonpatch") {
		return ApplyPatch(PatchTypeJSONPatch, input, content)
	}

	return nil, errors.Join(ErrUnsupportedPatchType, fmt.Errorf("file: %s", patchFile))
}
