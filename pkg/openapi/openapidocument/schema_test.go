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

	err := MergeSchema(base.CreateSchemaProxy(&b), base.CreateSchemaProxy(&ow))
	assert.NoError(t, err)

	assert.Equal(t, ow.Type, b.Type)
	assert.Equal(t, ow.Format, b.Format)
	assert.Equal(t, ow.Description, b.Description)
	assert.Equal(t, ow.Required, b.Required)
}
