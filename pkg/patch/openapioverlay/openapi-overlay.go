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
