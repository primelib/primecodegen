package openapipatch

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixOperationTags_LeavesUntaggedOperationsWithoutDefault(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /untagged:
    get:
      responses:
        '200':
          description: OK
`

	document, err := openapidocument.OpenDocument([]byte(spec))
	require.NoError(t, err)
	v3doc, err := document.BuildV3Model()
	require.NoError(t, err)

	err = FixOperationTags(v3doc, nil)
	require.NoError(t, err)

	pathItem, ok := v3doc.Model.Paths.PathItems.Get("/untagged")
	require.True(t, ok)
	assert.Empty(t, pathItem.Get.Tags)
	assert.Empty(t, v3doc.Model.Tags)
}

func TestFixOperationTags_DocumentsExistingOperationTags(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /pets:
    get:
      tags:
        - pets
      responses:
        '200':
          description: OK
`

	document, err := openapidocument.OpenDocument([]byte(spec))
	require.NoError(t, err)
	v3doc, err := document.BuildV3Model()
	require.NoError(t, err)

	err = FixOperationTags(v3doc, nil)
	require.NoError(t, err)

	require.Len(t, v3doc.Model.Tags, 1)
	assert.Equal(t, "pets", v3doc.Model.Tags[0].Name)
}
