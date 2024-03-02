package util

import (
	"fmt"
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

func TestLowerCaseFirstLetter(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Hello world", "hello world"},
		{"", ""},
		{"This is a test", "this is a test"},
		{"Another test", "another test"},
		{"another Test", "another Test"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := LowerCaseFirstLetter(tc.input)
			if result != tc.expected {
				t.Errorf("LowerCaseFirstLetter(%s) = %s; expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestTrimNonASCII(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world"},
		{"", ""},
		{"This is a test", "This is a test"},
		{"Another test", "Another test"},
		{"another Test", "another Test"},
		{"hello world! こんにちは", "hello world! "},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := TrimNonASCII(tc.input)
			if result != tc.expected {
				t.Errorf("TrimNonASCII(%s) = %s; expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFindCommonStrPrefix(t *testing.T) {
	testCases := []struct {
		input    []string
		expected string
	}{
		{[]string{"apple", "app", "ape"}, "ap"},
		{[]string{"apple", "banana", "peach"}, ""},
		{[]string{"hello", "hey", "hi"}, "h"},
		{[]string{"abcd", "abcde", "abcdef"}, "abcd"},
		{[]string{"same", "same", "same"}, "same"},
		{[]string{"", "hello", "hey"}, ""},
		{[]string{"prefix", "prefixes", "prefixed"}, "prefix"},
		{[]string{"/api/v1/books", "/api/v1/chapters", "/api/v1/authors"}, "/api/v1/"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Input: %v", testCase.input), func(t *testing.T) {
			result := FindCommonStrPrefix(testCase.input)
			if result != testCase.expected {
				t.Errorf("Expected prefix: %s, but got: %s", testCase.expected, result)
			}
		})
	}
}
