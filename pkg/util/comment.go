package util

import (
	"strings"
)

// CommentSingleLine returns a single line comment, replacing newlines with spaces
func CommentSingleLine(comment string) string {
	if comment == "" {
		return comment
	}

	comment = strings.Replace(comment, "\r\n", " ", -1)
	comment = strings.Replace(comment, "\n", " ", -1)

	return strings.TrimSpace(comment)
}
