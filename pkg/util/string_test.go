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

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"HelloWorld", "HelloWorld"},
		{"", ""},
	}

	for _, test := range tests {
		result := ToPascalCase(test.input)
		if result != test.expected {
			t.Errorf("ToPascalCase(%s) returned %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"helloWorld", "hello_world"},
		{"hello_world", "hello_world"},
		{"hello-world", "hello_world"},
		{"", ""},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToSnakeCase(%s) returned %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello-world"},
		{"helloWorld", "hello-world"},
		{"hello_world", "hello-world"},
		{"hello-world", "hello-world"},
		{"", ""},
	}

	for _, test := range tests {
		result := ToKebabCase(test.input)
		if result != test.expected {
			t.Errorf("ToKebabCase(%s) returned %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "helloWorld"},
		{"helloWorld", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"", ""},
	}

	for _, test := range tests {
		result := ToCamelCase(test.input)
		if result != test.expected {
			t.Errorf("ToCamelCase(%s) returned %s, expected %s", test.input, result, test.expected)
		}
	}
}
