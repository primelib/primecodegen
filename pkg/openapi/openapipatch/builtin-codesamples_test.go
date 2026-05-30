package openapipatch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeSamplesRefsAddsDeterministicRef(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /features:
    get:
      operationId: get-features
      responses:
        '200':
          description: OK
`

	doc := openapidocument.OpenV3DocumentForTest([]byte(spec))
	require.NoError(t, os.MkdirAll(filepath.Join("sdk", "java", "snippets"), 0755))
	t.Cleanup(func() {
		_ = os.RemoveAll("sdk")
	})
	require.NoError(t, os.WriteFile(filepath.Join("sdk", "java", "snippets", "get-features.java"), []byte("sample"), 0644))

	err := GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir":        "sdk/java",
		"language":   "java",
		"ref-prefix": "snippets",
	})
	require.NoError(t, err)

	op := getOperationForTest(t, doc, "/features", "get")
	node, exists := op.Extensions.Get("x-codeSamples")
	require.True(t, exists)

	var samples []codeSample
	require.NoError(t, node.Decode(&samples))
	require.Len(t, samples, 1)
	assert.Equal(t, "java", samples[0].Lang)
	assert.Equal(t, "Sample API Java SDK", samples[0].Label)
	assert.Equal(t, "sdk/java/snippets/get-features.java", samples[0].Source.Ref)
}

func TestGenerateCodeSamplesRefsAppendsWithoutReplacing(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /features:
    get:
      operationId: get-features
      x-codeSamples:
        - lang: go
          label: existing-go
          source:
            inline: existing sample
      responses:
        '200':
          description: OK
`

	doc := openapidocument.OpenV3DocumentForTest([]byte(spec))
	require.NoError(t, os.MkdirAll(filepath.Join("sdk", "java", "snippets"), 0755))
	t.Cleanup(func() {
		_ = os.RemoveAll("sdk")
	})
	require.NoError(t, os.WriteFile(filepath.Join("sdk", "java", "snippets", "get-features.java"), []byte("sample"), 0644))

	err := GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir":        "sdk/java",
		"language":   "java",
		"ref-prefix": "snippets",
	})
	require.NoError(t, err)

	op := getOperationForTest(t, doc, "/features", "get")
	node, exists := op.Extensions.Get("x-codeSamples")
	require.True(t, exists)

	var samples []codeSample
	require.NoError(t, node.Decode(&samples))
	require.Len(t, samples, 2)
	assert.Equal(t, "", samples[0].Source.Ref)
	assert.Equal(t, "sdk/java/snippets/get-features.java", samples[1].Source.Ref)
}

func TestGenerateCodeSamplesRefsDedupesByRef(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /features:
    get:
      operationId: get-features
      responses:
        '200':
          description: OK
`

	doc := openapidocument.OpenV3DocumentForTest([]byte(spec))
	require.NoError(t, os.MkdirAll(filepath.Join("sdk", "java", "snippets"), 0755))
	t.Cleanup(func() {
		_ = os.RemoveAll("sdk")
	})
	require.NoError(t, os.WriteFile(filepath.Join("sdk", "java", "snippets", "get-features.java"), []byte("sample"), 0644))

	err := GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir":        "sdk/java",
		"language":   "java",
		"ref-prefix": "snippets",
	})
	require.NoError(t, err)

	err = GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir":        "sdk/java",
		"language":   "java",
		"ref-prefix": "snippets",
	})
	require.NoError(t, err)

	op := getOperationForTest(t, doc, "/features", "get")
	node, exists := op.Extensions.Get("x-codeSamples")
	require.True(t, exists)

	var samples []codeSample
	require.NoError(t, node.Decode(&samples))
	require.Len(t, samples, 1)
	assert.Equal(t, "sdk/java/snippets/get-features.java", samples[0].Source.Ref)
}

func TestGenerateCodeSamplesRefsRequiresLanguageOrExtension(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /features:
    get:
      operationId: get-features
      responses:
        '200':
          description: OK
`

	doc := openapidocument.OpenV3DocumentForTest([]byte(spec))
	err := GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir": "sdk/java",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing extension")
}

func TestGenerateCodeSamplesRefsSkipsWhenFileMissing(t *testing.T) {
	const spec = `
openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /missing:
    get:
      operationId: get-missing
      responses:
        '200':
          description: OK
`

	doc := openapidocument.OpenV3DocumentForTest([]byte(spec))
	require.NoError(t, os.MkdirAll(filepath.Join("sdk", "java", "snippets"), 0755))
	t.Cleanup(func() {
		_ = os.RemoveAll("sdk")
	})

	err := GenerateCodeSamplesRefs(doc, map[string]interface{}{
		"dir":        "sdk/java",
		"language":   "java",
		"ref-prefix": "snippets",
	})
	require.NoError(t, err)

	op := getOperationForTest(t, doc, "/missing", "get")
	_, exists := op.Extensions.Get("x-codeSamples")
	assert.False(t, exists)
}

func getOperationForTest(t *testing.T, doc *libopenapi.DocumentModel[v3.Document], path string, method string) *v3.Operation {
	t.Helper()

	pathItem, ok := doc.Model.Paths.PathItems.Get(path)
	require.True(t, ok)

	for op := pathItem.GetOperations().Oldest(); op != nil; op = op.Next() {
		if op.Key == method {
			return op.Value
		}
	}

	t.Fatalf("operation not found for %s %s", method, path)
	return nil
}
