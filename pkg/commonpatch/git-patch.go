package commonpatch

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

var (
	ErrFailedToParseGitPatch = fmt.Errorf("failed to parse git patch")
	ErrFailedToApplyGitPatch = fmt.Errorf("failed to apply git patch")
)

func ApplyGitPatch(input []byte, patchContent []byte) ([]byte, error) {
	files, _, err := gitdiff.Parse(bytes.NewReader(patchContent))
	if err != nil {
		return nil, errors.Join(ErrFailedToParseGitPatch, err)
	}

	// apply the changes in the patch to a source file
	var output bytes.Buffer
	if err = gitdiff.Apply(&output, bytes.NewReader(input), files[0]); err != nil {
		return nil, errors.Join(ErrFailedToApplyGitPatch, err)
	}

	return output.Bytes(), nil
}
