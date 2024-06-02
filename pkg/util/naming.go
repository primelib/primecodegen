package util

import (
	"slices"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/iancoleman/strcase"
)

var replaceChars = []int32{'-', '_', ':', ' '}

func init() {
	// configure acronyms
	strcase.ConfigureAcronym("API", "api")
	strcase.ConfigureAcronym("HTML", "html")

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
	return strcase.ToCamel(input)
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
