package commonpatch

import (
	"errors"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

var (
	ErrFailedToDecodeJSONPatch = fmt.Errorf("failed to decode json patch")
	ErrFailedToApplyJSONPatch  = fmt.Errorf("failed to apply json patch")
)

func ApplyJSONPatch(input []byte, patchContent []byte) ([]byte, error) {
	patch, err := jsonpatch.DecodePatch(patchContent)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeJSONPatch, err)
	}

	modified, err := patch.Apply(input)
	if err != nil {
		return nil, errors.Join(ErrFailedToApplyJSONPatch, err)
	}

	return modified, nil
}
