package sharedpatch

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	ErrFailedToReadPatchFile              = fmt.Errorf("failed to read patch file")
	ErrUnsupportedPatchType               = fmt.Errorf("unsupported patch type")
	ErrExternalPatchMustHaveContentOrFile = fmt.Errorf("external patch must have either content or file specified")
)

type PatchType string

const (
	PatchTypeJSONPatch      PatchType = "jsonpatch"
	PatchTypeGitPatch       PatchType = "git"
	PatchTypeOpenAPIOverlay PatchType = "openapi-overlay"
	PatchTypeSpeakEasy      PatchType = "speakeasy"
)

type SpecPatch struct {
	Type        string                 `yaml:"type"`
	ID          string                 `yaml:"id,omitempty"`
	File        string                 `yaml:"file,omitempty"`
	Content     string                 `yaml:"content,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty"` // JSON or YAML config for the patch
	Description string                 `yaml:"description,omitempty"`
}

func (p SpecPatch) String() string {
	if p.Type != "" && p.Content != "" {
		return fmt.Sprintf("%s:<content>", p.Type)
	}
	if p.Type != "" && p.File != "" && p.ID == "" {
		return fmt.Sprintf("%s:%s", p.Type, p.File)
	}
	if p.Type != "" && p.ID != "" && p.File == "" {
		return fmt.Sprintf("%s:%s", p.Type, p.ID)
	}
	if p.Type == "" && p.File != "" {
		return fmt.Sprintf("file:%s", p.File)
	}

	return fmt.Sprintf("%s:%s:%s", p.Type, p.ID, p.File)
}

type PatchFile func(input []byte, patch []byte) ([]byte, error)

func ParsePatchSpecsFromStrings(patches []string) []SpecPatch {
	var specs []SpecPatch
	for _, p := range patches {
		var patchType string
		var patchFile string
		if strings.Contains(p, ":") {
			parts := strings.SplitN(p, ":", 2)
			if len(parts) != 2 {
				log.Fatal().Msg("invalid patch file syntax")
			}
			patchType = parts[0]
			patchFile = parts[1]
		} else {
			patchType = "builtin"
			patchFile = p
		}
		log.Debug().Str("patchType", patchType).Str("patchFile", patchFile).Msg("adding patch to spec")

		if patchType == "builtin" {
			specs = append(specs, SpecPatch{
				Type: patchType,
				ID:   patchFile,
			})
		} else {
			specs = append(specs, SpecPatch{
				Type: patchType,
				File: patchFile,
			})
		}
	}
	return specs
}

func SpecPatchesToStringSlice(patches []SpecPatch) []string {
	var patchStrings []string
	for _, p := range patches {
		patchStrings = append(patchStrings, p.String())
	}
	return patchStrings
}
