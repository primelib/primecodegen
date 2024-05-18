package commonpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyJSONPatch(t *testing.T) {
	input := []byte(`{ "foo": "bar" }`)
	patchContent := []byte(`[
		{ "op": "add", "path": "/baz", "value": "qux" },
		{ "op": "remove", "path": "/foo" }
	]`)
	expected := []byte(`{"baz":"qux"}`)

	result, err := ApplyJSONPatch(input, patchContent)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
