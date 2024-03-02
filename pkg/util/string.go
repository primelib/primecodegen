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

	return strings.ToUpper(input[0:1]) + input[1:]
}

// LowerCaseFirstLetter lowercases the first letter of the input string
func LowerCaseFirstLetter(input string) string {
	if input == "" {
		return input
	}

	return strings.ToLower(input[0:1]) + input[1:]
}

// TrimNonASCII removes non-ASCII characters from the input string
func TrimNonASCII(input string) string {
	return strings.Map(func(r rune) rune {
		if r > 127 {
			return -1
		}
		return r
	}, input)
}

// CapitalizeAfterChars removes the characters in the chars slice and capitalizes the next character
func CapitalizeAfterChars(input string, chars []int32, capitalizeFirst bool) string {
	var strBuilder strings.Builder
	shouldCapitalize := capitalizeFirst
	for _, char := range input {
		if slices.Contains(chars, char) {
			shouldCapitalize = true
			continue
		}
		if shouldCapitalize {
			strBuilder.WriteRune(unicode.ToUpper(char))
			shouldCapitalize = false
		} else {
			strBuilder.WriteRune(char)
		}
	}

	return strBuilder.String()
}

var replaceChars = []int32{'-', '_', ':'}

func ToPascalCase(input string) string {
	return CapitalizeAfterChars(input, replaceChars, true)
}

func ToSnakeCase(input string) string {
	var strBuilder strings.Builder
	for i, char := range input {
		if slices.Contains(replaceChars, char) {
			strBuilder.WriteRune('_')
			continue
		}
		if i > 0 && unicode.IsUpper(char) {
			strBuilder.WriteRune('_')
		}
		strBuilder.WriteRune(unicode.ToLower(char))
	}
	return strBuilder.String()
}

func ToKebabCase(input string) string {
	var strBuilder strings.Builder
	for i, char := range input {
		if slices.Contains(replaceChars, char) {
			strBuilder.WriteRune('-')
			continue
		}
		if i > 0 && unicode.IsUpper(char) {
			strBuilder.WriteRune('-')
		}
		strBuilder.WriteRune(unicode.ToLower(char))
	}
	return strBuilder.String()
}

func ToCamelCase(input string) string {
	return LowerCaseFirstLetter(ToPascalCase(input))
}
