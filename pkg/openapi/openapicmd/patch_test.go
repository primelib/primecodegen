package openapicmd

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRunPatchCmdInvalidPath(t *testing.T) {
	_, err := runPatchCmd([]string{"missing-file.yaml"}, "", []string{}, []string{})
	assert.ErrorIs(t, err, util.ErrDocumentMerge)
	assert.ErrorIs(t, err, util.ErrFileMissing)
}
