package util

import (
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/rs/zerolog/log"
)

func RenameOrderedMapKeys[V any](
	m *orderedmap.Map[string, V],
	transformKey func(string) string,
) (mapping map[string]string) {
	if m == nil {
		return nil
	}

	// Collect all keys + values
	var originalKeys []string
	var originalValues []V

	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		if pair.Key == "" {
			continue
		}
		originalKeys = append(originalKeys, pair.Key)
		originalValues = append(originalValues, pair.Value)
	}

	// Delete original keys
	for _, key := range originalKeys {
		m.Delete(key)
	}

	// Insert new keys and track mapping
	mapping = make(map[string]string)
	for i, oldKey := range originalKeys {
		newKey := transformKey(oldKey)

		if _, exists := m.Get(newKey); !exists {
			m.Set(newKey, originalValues[i])
		} else {
			log.Error().Str("oldKey", oldKey).Str("newKey", newKey).Msg("Key already exists, skipping")
			continue
		}

		mapping[oldKey] = newKey
	}

	return mapping
}
