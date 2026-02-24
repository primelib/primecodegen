package util

import (
	"log/slog"

	"github.com/pb33f/libopenapi/orderedmap"
)

func AppendOrSetString(destStr *string, srcStr, prefix, separator string) {
	if srcStr == "" {
		return
	}

	if *destStr != "" {
		*destStr += separator + prefix + srcStr
	} else {
		*destStr = prefix + srcStr
	}
}

func MergeComponentMap[V any](destMap, srcMap *orderedmap.Map[string, V], componentType string) {
	for item := srcMap.First(); item != nil; item = item.Next() {
		name, value := item.Key(), item.Value()
		if _, exists := destMap.Get(name); !exists {
			destMap.Set(name, value)
		} else {
			slog.Error("Component already exists", "component", name, "type", componentType)
			// TODO: Handle duplicate (rename | prefix)
		}
	}
}
