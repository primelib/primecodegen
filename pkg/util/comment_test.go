package util

import (
	"testing"
)

func TestCommentSingleLine(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"", ""},
		{"foo", "// foo"},
		{"foo\nbar", "// foo bar"},
		{"foo\nbar\nbaz", "// foo bar baz"},
	}

	for _, test := range tests {
		result := CommentSingleLine(test.input)
		if result != test.output {
			t.Errorf("Expected: %s, Got: %s for input: %v", test.output, result, test.input)
		}
	}
}
