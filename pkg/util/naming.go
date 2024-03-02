package util

import (
	"slices"
	"strings"
	"unicode"
)

var replaceChars = []int32{'-', '_', ':'}

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
