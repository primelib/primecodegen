package util

import (
	"bytes"
	"strings"
)

var newLineReplacer = strings.NewReplacer(
	"\r\n", " ", // windows newlines
	"\n", " ", // unix newlines
	"\u2028", " ", // unicode line separator
	"\u2029", " ", // unicode paragraph separator
)

// CommentSingleLine returns a single line comment, replacing newlines with spaces
func CommentSingleLine(comment string) string {
	if comment == "" {
		return comment
	}

	comment = newLineReplacer.Replace(comment)
	return strings.TrimSpace(comment)
}

func CommentMultiLine(comment string, prefix string) string {
	var output bytes.Buffer

	lines := strings.Split(comment, "\n")
	for i, line := range lines {
		if i != 0 {
			output.WriteString(prefix)
		}
		output.WriteString(line)
		if i < len(lines)-1 {
			output.WriteString("\n")
		}
	}

	return output.String()
}
