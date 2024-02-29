package util

import (
	"strings"
)

// CommentSingleLine returns a single line comment
func CommentSingleLine(comment string) string {
	if comment == "" {
		return comment
	}

	return "// " + strings.Replace(strings.TrimSpace(comment), "\n", " ", -1)
}
