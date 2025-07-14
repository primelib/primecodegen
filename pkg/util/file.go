package util

import (
	"bytes"
)

func DetectJSONOrYAML(input []byte) (format string) {
	trimmed := bytes.TrimLeft(input, " \t\r\n")
	if len(trimmed) == 0 {
		return "yaml"
	}

	switch trimmed[0] {
	case '{', '[':
		return "json"
	default:
		return "yaml"
	}
}
