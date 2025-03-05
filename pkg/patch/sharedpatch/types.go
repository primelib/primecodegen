package sharedpatch

import (
	"fmt"
)

var (
	ErrFailedToReadPatchFile = fmt.Errorf("failed to read patch file")
	ErrUnsupportedPatchType  = fmt.Errorf("unsupported patch type")
)

type PatchType string

const (
	PatchTypeJSONPatch      PatchType = "jsonpatch"
	PatchTypeGitPatch       PatchType = "git"
	PatchTypeOpenAPIOverlay PatchType = "openapi-overlay"
)

type PatchFile func(input []byte, patch []byte) ([]byte, error)
