package openapioverlay

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/speakeasy-api/openapi-overlay/pkg/overlay"
	"gopkg.in/yaml.v3"
)

var (
	ErrFailedToDecodeOpenAPIOverlay   = fmt.Errorf("failed to decode openapi overlay")
	ErrFailedToApplyOpenAPIOverlay    = fmt.Errorf("failed to apply openapi overlay")
	ErrFailedToValidateOpenAPIOverlay = fmt.Errorf("failed to validate openapi overlay")
)

// ParseOpenAPIOverlay parses the overlay bytes into an overlay object
func ParseOpenAPIOverlay(patchContent []byte) (*overlay.Overlay, error) {
	if len(patchContent) == 0 {
		return nil, errors.New("patch content is empty")
	}

	var ol overlay.Overlay
	if err := yaml.NewDecoder(bytes.NewReader(patchContent)).Decode(&ol); err != nil {
		return nil, err
	}

	return &ol, nil
}

// ApplyOpenAPIOverlay applies the overlay content to the input
func ApplyOpenAPIOverlay(input []byte, patchContent []byte) ([]byte, error) {
	o, err := ParseOpenAPIOverlay(patchContent)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeOpenAPIOverlay, err)
	}

	yn, err := loader.YamlNodeFromString(input)
	if err != nil {
		return nil, err
	}

	err = o.ApplyTo(yn)
	if err != nil {
		return nil, errors.Join(ErrFailedToApplyOpenAPIOverlay, err)
	}

	return loader.InterfaceToYaml(yn)
}

// ValidateOpenAPIOverlay validates the overlay content
func ValidateOpenAPIOverlay(patchContent []byte) error {
	o, err := ParseOpenAPIOverlay(patchContent)
	if err != nil {
		return errors.Join(ErrFailedToDecodeOpenAPIOverlay, err)
	}

	err = o.Validate()
	if err != nil {
		return errors.Join(ErrFailedToValidateOpenAPIOverlay, err)
	}

	return nil
}

func CreateInfoOverlay(name string, description string, licenseName string, licenseUrl string) overlay.Overlay {
	ov := overlay.Overlay{
		Extensions:      nil,
		Version:         "1.0.0",
		JSONPathVersion: "",
		Info: overlay.Info{
			Title:   "Info Overlay",
			Version: "1.0.0",
		},
		Actions: []overlay.Action{},
	}
	if name != "" {
		ov.Actions = append(ov.Actions, overlay.Action{
			Target: "$.info",
			Update: loader.YamlNodeFromInterfaceNoErr(map[string]interface{}{
				"title": name,
			}),
		})
	}
	if description != "" {
		ov.Actions = append(ov.Actions, overlay.Action{
			Target: "$.info",
			Update: loader.YamlNodeFromInterfaceNoErr(map[string]interface{}{
				"description": description,
			}),
		})
	}
	if licenseName != "" && licenseUrl != "" {
		ov.Actions = append(ov.Actions, overlay.Action{
			Target: "$.info",
			Update: loader.YamlNodeFromInterfaceNoErr(map[string]interface{}{
				"license": map[string]interface{}{
					"name": licenseName,
					"url":  licenseUrl,
				},
			}),
		})
	}

	return ov
}
