package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeJSON(t *testing.T) {
	mergedData := make(map[string]interface{})

	jsonStringOne := `{
		"key1": "value1",
		"key2": {
			"nested1": "nestedValue1"
		}
	}`
	jsonStringTwo := `{
		"key2": {
			"nested2": "nestedValue2"
		},
		"key3": "value3"
	}`

	for _, str := range []string{jsonStringOne, jsonStringTwo} {
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(str), &jsonData)
		assert.NoError(t, err)

		MergeMaps(mergedData, jsonData)
	}

	// marshal merged data
	bytes, err := json.MarshalIndent(mergedData, "", "  ")
	assert.NoError(t, err)

	// verify
	expectedJSON := `{
		"key1": "value1",
		"key2": {
			"nested1": "nestedValue1",
			"nested2": "nestedValue2"
		},
		"key3": "value3"
	}`
	assert.JSONEq(t, expectedJSON, string(bytes))
}
