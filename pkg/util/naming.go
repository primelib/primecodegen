package util

import (
	"slices"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/iancoleman/strcase"
)

var replaceChars = []int32{'-', '_', ':', ' '}

var acronyms = []string{
	"id",
	"api",
	"vcs",
	"git",
}

func init() {
	// configure custom rune substitutions for slug
	slug.CustomRuneSub = map[rune]string{
		'_': "-",
	}
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

func ToPascalCase(input string) string {
	return UppercaseAcronyms(strcase.ToCamel(input))
}

func ToSnakeCase(input string) string {
	return strcase.ToSnake(input)
}

func ToKebabCase(input string) string {
	return strcase.ToKebab(input)
}

func ToCamelCase(input string) string {
	return strcase.ToLowerCamel(input)
}

func ToSlug(input string) string {
	return slug.MakeLang(input, "en")
}

// UppercaseAcronyms replaces acronyms in the input string with their uppercase form
func UppercaseAcronyms(input string) string {
	for i, _ := range input {
		// check if any acronym starts at this index
		for _, acronym := range acronyms {
			if strings.HasPrefix(strings.ToLower(input[i:]), strings.ToLower(acronym)) {
				acronymEnd := i + len(acronym)
				if acronymEnd < len(input) {
					// uppercase acronym
					input = input[:i] + strings.ToUpper(acronym) + input[acronymEnd:]

					// uppercase following character, if lowercase
					if acronymEnd < len(input) && unicode.IsLower(rune(input[acronymEnd])) {
						input = input[:acronymEnd] + strings.ToUpper(string(input[acronymEnd])) + input[acronymEnd+1:]
					}
				}
			}
		}
	}

	return input
}
