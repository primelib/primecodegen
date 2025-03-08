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
	PatchTypeSpeakEasy      PatchType = "speakeasy"
)

type SpecPatch struct {
	Type        string `yaml:"type"`
	ID          string `yaml:"id"`
	File        string `yaml:"file"`
	Content     string `yaml:"content"`
	Description string `yaml:"description,omitempty"`
}

type PatchFile func(input []byte, patch []byte) ([]byte, error)
