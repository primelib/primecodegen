package openapipatch

import (
	"fmt"
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
    tags:
      - name: pet
        description: Everything about your Pets
        externalDocs:
          description: Find out more
          url: http://swagger.io      
    paths:
      /sample-resource1:
        get:
          tags:
            - pet        
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
          tags:
           - pet          
          responses:
           '200':
             description: OK
          requestBody:
           content: 
    
        post:
          tags:
            - pet  
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
	patchedAPISepc := v3doc.Model.RenderWithIndention(4)
	fmt.Printf("Patched API spec: %s", string(patchedAPISepc))

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

func TestMergePolymorphicSchemas(t *testing.T) {
	// arrange
	const spec = `
    openapi: 3.0.0
    info:
      title: Sample API
      version: v1.0.0
    components:
        schemas:
            BaseSchemaA:
                properties:
                  propertyA:
                    type: string
                    description: Description A
                  propertyB:
                    type: string
                    description: Description B
            BaseSchemaB:
                properties:
                  propertyF:
                    type: string
                    description: Description F
                  propertyG:
                    type: string
                    description: Description G
                  propertyH: 
                      $ref: '#/components/schemas/DerivedSchemaAny'                    
            DerivedSchemaAllOf:
                allOf:
                    - $ref: '#/components/schemas/BaseSchemaA'
                    - properties:
                        additionalPropertyC:
                          type: string
                          description: Description C
            DerivedSchemaOneOf:
                oneOf:
                    - $ref: '#/components/schemas/BaseSchemaA'
                    - $ref: '#/components/schemas/BaseSchemaB'
                    - properties:
                        additionalPropertyD:
                          type: string
                          description: Description D
            DerivedSchemaAny:
                anyOf:
                    - $ref: '#/components/schemas/BaseSchemaA'
                    - $ref: '#/components/schemas/BaseSchemaB'                    
                    - properties:
                        additionalPropertyE:
                          type: string
                          description: Description E                             
    `
	document, err := openapidocument.OpenDocument([]byte(spec))
	if err != nil {
		t.Fatalf("error creating document: %v", err)
	}
	v3Model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		t.Fatalf("error creating document: %v", errs)
	}

	// act
	_ = MergePolymorphicSchemas(v3Model)

	// assert
	_, document, v3model, errors := document.RenderAndReload()
	assert.Equal(t, 0, len(errors))

	propsBaseAToCheck := []string{"propertyA", "propertyB", "additionalPropertyC", "additionalPropertyD", "additionalPropertyE"}
	propsBaseBToCheck := []string{"propertyF", "propertyG", "propertyH", "additionalPropertyD", "additionalPropertyE"}

	_, present := v3model.Model.Components.Schemas.Get("DerivedSchemaAllOf")
	assert.False(t, present)
	_, present = v3model.Model.Components.Schemas.Get("DerivedSchemaOneOf")
	assert.False(t, present)
	_, present = v3model.Model.Components.Schemas.Get("DerivedSchemaAnyOf")
	assert.False(t, present)

	for schemaMapEntry := v3model.Model.Components.Schemas.Oldest(); schemaMapEntry != nil; schemaMapEntry = schemaMapEntry.Next() {
		schema, err := schemaMapEntry.Value.BuildSchema()
		assert.NoError(t, err)

		assert.Nil(t, schema.AllOf, "allOf references should be deleted")
		assert.Nil(t, schema.AnyOf, "anyOf references should be deleted")
		assert.Nil(t, schema.OneOf, "oneOf references should be deleted")

		if schemaMapEntry.Key == "BaseSchemaA" {
			assert.Equal(t, 5, schema.Properties.Len())

			for _, prop := range propsBaseAToCheck {
				_, exists := schema.Properties.Get(prop)
				assert.True(t, exists, "Property \"%s\" is missing!", prop)
			}
		}
		if schemaMapEntry.Key == "BaseSchemaB" {
			assert.Equal(t, 5, schema.Properties.Len())

			for _, prop := range propsBaseBToCheck {
				propSP, exists := schema.Properties.Get(prop)
				assert.True(t, exists, "Property \"%s\" is missing!", prop)
				if prop == "propertyH" {
					assert.Equal(t, "#/components/schemas/BaseSchemaB", propSP.GetReference())
				}
			}

		}
	}
}
