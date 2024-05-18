package commonpatch

import (
	"bytes"
	"fmt"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

func ApplyGitPatch(input []byte, patchContent []byte) ([]byte, error) {
	files, _, err := gitdiff.Parse(bytes.NewReader(patchContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse git patch: %w", err)
	}

	// apply the changes in the patch to a source file
	var output bytes.Buffer
	if err = gitdiff.Apply(&output, bytes.NewReader(input), files[0]); err != nil {
		return nil, fmt.Errorf("failed to apply git patch: %w", err)
	}

	return output.Bytes(), nil
}
