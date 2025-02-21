package openapidocument

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/stretchr/testify/assert"
)

func TestMergeSchema(t *testing.T) {
	b := base.Schema{Type: []string{"string"}}
	ow := base.Schema{
		Type:        []string{"string", "null"},
		Format:      "uuid",
		Description: "new description",
		Required:    []string{"name"},
	}

	mergedSchema, err := MergeSchemaProxy(base.CreateSchemaProxy(&b), base.CreateSchemaProxy(&ow))
	assert.NoError(t, err)

	assert.NotNil(t, mergedSchema)
	assert.Equal(t, len(ow.Type), 2)
	assert.Equal(t, len(b.Type), 3)
	assert.Equal(t, len(mergedSchema.Type), 3)
	assert.Equal(t, ow.Format, b.Format)
	assert.Equal(t, ow.Description, b.Description)
	assert.Equal(t, ow.Required, b.Required)
}
