package openapigenerator

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
)

func TestIsBinaryMediaType(t *testing.T) {
	for _, mt := range []string{"application/octet-stream", "image/png", "application/pdf", "application/zip", "audio/mpeg"} {
		assert.True(t, IsBinaryMediaType(mt), mt)
	}
	for _, mt := range []string{"", "*/*", "application/json", "application/problem+json", "application/xml", "text/plain", "text/csv; charset=utf-8", "application/yaml", "multipart/form-data", "application/x-www-form-urlencoded"} {
		assert.False(t, IsBinaryMediaType(mt), mt)
	}
}

func TestContentSchema(t *testing.T) {
	// binary media type without schema
	s := ContentSchema("image/png", &v3.MediaType{})
	assert.Equal(t, []string{"string"}, s.Type)
	assert.Equal(t, "binary", s.Format)

	// binary media type with plain string schema
	s = ContentSchema("application/pdf", &v3.MediaType{Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}})})
	assert.Equal(t, "binary", s.Format)

	// textual media type keeps plain string
	s = ContentSchema("text/plain", &v3.MediaType{Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}})})
	assert.Equal(t, "", s.Format)

	// textual media type without schema
	s = ContentSchema("text/plain", &v3.MediaType{})
	assert.Equal(t, []string{"string"}, s.Type)

	// json without schema
	assert.Nil(t, ContentSchema("application/json", &v3.MediaType{}))
}
