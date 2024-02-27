package openapipatch

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/stretchr/testify/assert"
)

func TestPruneOperationTags(t *testing.T) {
	// parse spec
	const spec = `
    openapi: 3.0.0
    info:
      title: Sample API
      version: 1.0.0
    paths:
      /test:
        get:
          tags:
            - test
          responses:
            '200':
              description: OK
    `
	document, err := openapidocument.OpenDocument([]byte(spec))
	if err != nil {
		t.Fatalf("error creating document: %v", err)
	}
	v3doc, errors := document.BuildV3Model()
	assert.Equal(t, 0, len(errors))

	// prune operation tags
	_ = PruneOperationTags(v3doc)

	// check if tags are pruned
	v, _ := v3doc.Model.Paths.PathItems.Get("/test")
	assert.Nil(t, v.Get.Tags, "tags should be pruned")
}
