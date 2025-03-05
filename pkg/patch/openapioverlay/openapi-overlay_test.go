package openapioverlay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyOpenAPIOverlay(t *testing.T) {
	input := []byte(`openapi: 3.0.0
info:
  title: Sample API
  version: 0.1.9
paths:
  /example:
    get:
      summary: Example endpoint
`)
	patchContent := []byte(`overlay: 1.0.0
info:
  title: Fix up API description
  version: 1.0.1
actions:
  - target: $.info
    update:
      title: Public-facing API Title
      description: >-
        Description fields allow for longer explanations and support Markdown
        so that you can add links or formatting where it's useful.
  - target: $.paths['/example'].get
    remove: true
`)
	expected := `openapi: 3.0.0
info:
  title: Public-facing API Title
  version: 0.1.9
  description: >-
    Description fields allow for longer explanations and support Markdown so that you can add links or formatting where it's useful.
paths:
  /example: {}
`

	result, err := ApplyOpenAPIOverlay(input, patchContent)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestValidateOpenAPIOverlay(t *testing.T) {
	validPatchContent := []byte(`
overlay: 1.0.0
info:
  title: Fix up API description
  version: 1.0.1
actions:
  - target: $.info
    update:
      title: Public-facing API Title
`)

	invalidPatchContent := []byte(`
overlay: 0.1.0
info:
actions: []
`)

	err := ValidateOpenAPIOverlay(validPatchContent)
	assert.NoError(t, err)

	err = ValidateOpenAPIOverlay(invalidPatchContent)
	assert.Error(t, err)
}
