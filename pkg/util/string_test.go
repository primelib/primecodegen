package util

import (
	"testing"
)

func TestFirstNonEmptyString(t *testing.T) {
	tests := []struct {
		input  []string
		output string
	}{
		{[]string{"", "foo", "bar"}, "foo"},
		{[]string{"", "", ""}, ""},
		{[]string{"", "baz", "qux"}, "baz"},
		{[]string{"hello", "", "world"}, "hello"},
	}

	for _, test := range tests {
		result := FirstNonEmptyString(test.input...)
		if result != test.output {
			t.Errorf("Expected: %s, Got: %s for input: %v", test.output, result, test.input)
		}
	}
}

func TestUpperCaseFirstLetter(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello world"},
		{"", ""},
		{"this is a test", "This is a test"},
		{"another test", "Another test"},
		{"Another Test", "Another Test"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := UpperCaseFirstLetter(tc.input)
			if result != tc.expected {
				t.Errorf("UpperCaseFirstLetter(%s) = %s; expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCapitalizeAfterChars(t *testing.T) {
	testCases := []struct {
		input           string
		chars           []int32
		capitalizeFirst bool
		expected        string
	}{
		{"hello/world", []int32{'/'}, false, "helloWorld"},
		{"_hello", []int32{'_'}, false, "Hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := CapitalizeAfterChars(tc.input, tc.chars, tc.capitalizeFirst)
			if result != tc.expected {
				t.Errorf("CharToCapitalize(%s, %v, %t) = %s; expected %s", tc.input, tc.chars, tc.capitalizeFirst, result, tc.expected)
			}
		})
	}
}
