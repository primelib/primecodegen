package util

import (
	"slices"
	"strings"
	"unicode"
)

// FirstNonEmptyString returns the first non-empty string from the input strings
func FirstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// UpperCaseFirstLetter capitalizes the first letter of the input string
func UpperCaseFirstLetter(input string) string {
	if input == "" {
		return input
	}

	firstChar := strings.ToUpper(string(input[0]))
	return firstChar + input[1:]
}

// CapitalizeAfterChars removes the characters in the chars slice and capitalizes the next character
func CapitalizeAfterChars(input string, chars []int32, capitalizeFirst bool) string {
	var modifiedURL strings.Builder
	shouldCapitalize := capitalizeFirst
	for _, char := range input {
		if slices.Contains(chars, char) {
			shouldCapitalize = true
			continue
		}
		if shouldCapitalize {
			modifiedURL.WriteRune(unicode.ToUpper(char))
			shouldCapitalize = false
		} else {
			modifiedURL.WriteRune(char)
		}
	}

	return modifiedURL.String()
}
