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
	// General Tech & Web
	"ID", "API", "GIT", "VCS", "UUID", "GUID", "REST", "SOAP",

	// Business
	"VAT", "SLA", "SLO", "IBAN", "BIC",

	// TM Forum (TMF) & Telecom Specific
	"TMF", "SID", "TAM", "eTOM", "IP", "VPN", "SIM", "IMEI", "IMSI", "CPE",
	"FVO", "MVO", // first-value object and mutation-value object
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

func ToUpperSnakeCase(input string) string {
	return strcase.ToScreamingSnake(input)
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

func ToUpperCamelCase(input string) string {
	return strcase.ToCamel(input)
}

func ToSlug(input string) string {
	return slug.MakeLang(input, "en")
}

// UppercaseAcronyms replaces acronyms in the input string with their uppercase form
func UppercaseAcronyms(input string) string {
	for _, acronym := range acronyms {
		upperAcro := strings.ToUpper(acronym)

		search := input
		var result strings.Builder
		i := 0
		for i < len(search) {
			if strings.HasPrefix(strings.ToLower(search[i:]), strings.ToLower(acronym)) {
				endIdx := i + len(acronym)

				// Boundary Check:
				// Is it the start of the string OR was the previous char NOT a lowercase letter?
				// Is it the end of the string OR is the next char NOT a lowercase letter?
				isStart := i == 0 || !unicode.IsLower(rune(search[i-1]))
				isEnd := endIdx == len(search) || !unicode.IsLower(rune(search[endIdx]))

				if isStart && isEnd {
					result.WriteString(upperAcro)
					i = endIdx
					continue
				}
			}
			result.WriteByte(search[i])
			i++
		}
		input = result.String()
	}
	return input
}
