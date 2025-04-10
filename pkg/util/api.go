package util

import (
	"regexp"
	"strings"
)

func URLRemovePathParams(url string) string {
	re := regexp.MustCompile(`{[^}]+}`)
	return re.ReplaceAllString(url, "")
}

// URLPathParamAddByPrefix converts path parameters to By{ParamName}
func URLPathParamAddByPrefix(path string) string {
	re := regexp.MustCompile(`{([^}]+)}`)
	return re.ReplaceAllStringFunc(path, func(match string) string {
		paramName := strings.Trim(match, "{}")
		/*
			if paramName == "id" {
				return ""
			}
		*/
		return "By" + strings.Title(paramName)
	})
}

func ParseURLAPIVersion(url string) string {
	re := regexp.MustCompile(`/[vV]([0-9]+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 2 {
		return matches[1]
	}
	return "1"
}

func ContentTypeToShortName(input string) string {
	if input == "application/json" {
		return "json"
	} else if input == "application/xml" {
		return "xml"
	} else if input == "application/yaml" {
		return "yaml"
	} else if input == "application/x-www-form-urlencoded" {
		return "form"
	} else if input == "multipart/form-data" {
		return "form"
	} else if input == "application/octet-stream" {
		return "bytes"
	} else if input == "application/hal+json" {
		return "haljson"
	}

	return input
}
