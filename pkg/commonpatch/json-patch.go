package commonpatch

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

func ApplyJSONPatch(input []byte, patchContent []byte) ([]byte, error) {
	patch, err := jsonpatch.DecodePatch(patchContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json patch: %w", err)
	}

	modified, err := patch.Apply(input)
	if err != nil {
		return nil, fmt.Errorf("failed to apply json patch: %w", err)
	}

	return modified, nil
}
