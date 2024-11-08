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

func TestCreateOperationTagsFromDocTitle(t *testing.T) {
	// arrange
	const spec = `
    openapi: 3.0.3
    info: 
      title: Sample API
      description: 
      version: v1.0.0
    paths:
      /sample-resource1:
        get:
          responses:
           '200':
             description: OK
        put:
          responses:
           '200':
             description: OK
          requestBody:
           content: 
    
        post:
          responses:
           '200':
             description: OK
          requestBody:
           content: 
      /sampel-resource2:
        get:
          responses:
           '200':
             description: OK
        put:
          responses:
           '200':
             description: OK
          requestBody:
           content: 
    
        post:
          responses:
           '200':
             description: OK
          requestBody:
           content: 
    `
	document, err := openapidocument.OpenDocument([]byte(spec))
	if err != nil {
		t.Fatalf("error creating document: %v", err)
	}
	v3doc, errors := document.BuildV3Model()
	assert.Equal(t, 0, len(errors))

	// act
	err = CreateOperationTagsFromDocTitle(v3doc)
	assert.NoError(t, err)

	// assert
	_, document, _, errors = document.RenderAndReload()
	assert.Equal(t, 0, len(errors))
	v3doc, errors = document.BuildV3Model()
	assert.Equal(t, 0, len(errors))
	// patchedAPISepc := v3doc.Model.RenderWithIndention(4)
	// fmt.Printf("Patched API spec: %s", string(patchedAPISepc))

	// Verify the document tag
	assert.Len(t, v3doc.Model.Tags, 1)
	assert.Equal(t, "Sample API", v3doc.Model.Tags[0].Name)
	assert.Equal(t, "See document description", v3doc.Model.Tags[0].Description)

	// Verify the tags on operations
	for path := v3doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			assert.Len(t, op.Value.Tags, 1)
			assert.Equal(t, "Sample API", op.Value.Tags[0])
		}
	}
}

func TestInlineAllOfHierarchies(t *testing.T) {
	// arrange
	const spec = `
    openapi: 3.0.0
    info:
      title: Sample API
      version: v1.0.0
    components:
        schemas:
            TestSchema:
                properties:
                  propertyA:
                    type: string
                    description: Description A
                  propertyB:
                    type: string
                    description: Description B
            TestSchemaWithReferences:
                allOf:
                    - $ref: '#/components/schemas/TestSchema'
                    - properties:
                        additionalPropertyC:
                          type: string
                          description: Description C
    `
	document, err := openapidocument.OpenDocument([]byte(spec))
	if err != nil {
		t.Fatalf("error creating document: %v", err)
	}
	v3doc, errors := document.BuildV3Model()
	assert.Equal(t, 0, len(errors))

	// act
	_ = InlineAllOfHierarchies(v3doc)

	// assert
	_, document, _, errors = document.RenderAndReload()
	assert.Equal(t, 0, len(errors))
	v3doc, errors = document.BuildV3Model()
	assert.Equal(t, 0, len(errors))

	propsToCheck := []string{"propertyA", "propertyB", "additionalPropertyC"}

	for schemaMapEntry := v3doc.Model.Components.Schemas.Oldest(); schemaMapEntry != nil; schemaMapEntry = schemaMapEntry.Next() {
		schema, err := schemaMapEntry.Value.BuildSchema()
		assert.NoError(t, err)
		assert.Nil(t, schema.AllOf, "allOf references should be deleted")
		if schemaMapEntry.Key == "TestSchemaWithReferences" {
			assert.Equal(t, 3, schema.Properties.Len())
			for _, prop := range propsToCheck {
				_, exists := schema.Properties.Get(prop)
				assert.True(t, exists, "Property \"%s\" is missing!", prop)
			}
		}
	}
}
