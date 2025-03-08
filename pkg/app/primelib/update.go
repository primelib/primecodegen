package primelib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/openapi/openapicmd"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/patch/openapioverlay"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/rs/zerolog/log"
)

// Update will update the openapi spec and apply patches
func Update(dir string, conf appconf.Configuration, repository api.Repository) error {
	spec := conf.Spec
	specFile := filepath.Join(dir, conf.Spec.File)
	log.Debug().Strs("spec-urls", spec.UrlSlice()).Str("spec-format", string(spec.Type)).Str("spec-file", specFile).Msg("processing module")

	// remove old file
	_ = os.Remove(specFile)

	// download spec sources
	targetSpecDir := spec.GetSourcesDir(dir)
	var specFiles []string
	var specFilesType []appconf.SpecType
	var tempFiles []string
	defer func() {
		for _, f := range tempFiles {
			_ = os.Remove(f)
		}
	}()

	// download spec sources
	for _, s := range spec.Sources {
		log.Debug().Str("url", s.URL).Str("type", string(s.Type)).Msg("fetching spec")
		var targetFile string
		var bytes []byte
		var err error

		// fetch spec
		if s.File != "" && s.URL == "" {
			bytes, err = os.ReadFile(filepath.Join(targetSpecDir, s.File))
		} else if s.URL != "" {
			bytes, err = openapidocument.FetchSpec(s)
		}
		if err != nil {
			return fmt.Errorf("failed to fetch spec: %w", err)
		}

		if s.File != "" && s.URL != "" {
			targetFile = filepath.Join(targetSpecDir, s.File)
		} else {
			tempFile, err := os.CreateTemp("", "api-spec-*.yaml")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			tempFiles = append(tempFiles, tempFile.Name())
			targetFile = tempFile.Name()
		}

		// write to file
		err = os.WriteFile(targetFile, bytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write api spec to file: %w", err)
		}
		specFiles = append(specFiles, targetFile)
		specFilesType = append(specFilesType, s.Type)
	}

	// spec type conversions
	for i, f := range specFiles {
		// convert from swagger to openapi
		if spec.Type == appconf.SpecTypeOpenAPI3 && specFilesType[i] == appconf.SpecTypeSwagger2 {
			log.Debug().Str("file", f).Msg("converting from swagger to openapi")
			output, err := openapicmd.ConvertSpec(f, "swagger2", "openapi3", "")
			if err != nil {
				return fmt.Errorf("failed to convert swagger to openapi: %w", err)
			}

			err = os.WriteFile(f, output, 644)
			if err != nil {
				return fmt.Errorf("failed to write converted spec to file: %w", err)
			}
		}
	}

	// openapi processing
	if spec.Type == appconf.SpecTypeOpenAPI3 {
		log.Debug().Strs("files", specFiles).Str("output", specFile).Msg("merging and patching openapi spec")

		// inputPatches
		inputPatches, inputPatchTempFiles, err := processPatches(spec.InputPatches)
		tempFiles = append(tempFiles, inputPatchTempFiles...)
		if err != nil {
			return fmt.Errorf("failed to process patches: %w", err)
		}

		// apply default overlay
		ov := openapioverlay.CreateInfoOverlay(repository.Name, repository.Description, repository.LicenseName, repository.LicenseURL)
		specPatch, err := patch.NewContentPatch("openapi-overlay", ov)
		if err != nil {
			return fmt.Errorf("failed to create overlay info patch: %w", err)
		}

		spec.Patches = append([]sharedpatch.SpecPatch{specPatch}, spec.Patches...)

		// patches
		patches, patchTempFiles, err := processPatches(spec.Patches)
		tempFiles = append(tempFiles, patchTempFiles...)
		if err != nil {
			return fmt.Errorf("failed to process patches: %w", err)
		}

		// merge and patch
		_, err = openapicmd.Patch(specFiles, specFile, inputPatches, patches)
		if err != nil {
			return fmt.Errorf("failed to patch openapi spec: %w", err)
		}
	}

	return nil
}

func processPatches(patchesList []sharedpatch.SpecPatch) ([]string, []string, error) {
	var patches []string
	var tempFiles []string

	for _, p := range patchesList {
		if p.ID != "" {
			if p.Type != "" {
				patches = append(patches, p.Type+":"+p.ID)
			} else {
				patches = append(patches, p.ID)
			}
		} else if p.File != "" {
			patches = append(patches, p.Type+":"+p.File)
		} else if p.Content != "" {
			tmpFile, err := os.CreateTemp("", "patch-*."+p.Type)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
			}
			tempFiles = append(tempFiles, tmpFile.Name())

			if _, err = tmpFile.Write([]byte(p.Content)); err != nil {
				tmpFile.Close()
				return nil, nil, fmt.Errorf("failed to write patch content to temp file: %w", err)
			}
			tmpFile.Close()

			patches = append(patches, p.Type+":"+tmpFile.Name())
		}
	}

	return patches, tempFiles, nil
}
