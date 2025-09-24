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
	} else if input == "text/plain" {
		return "text"
	} else if input == "text/html" {
		return "html"
	} else if input == "text/xml" {
		return "xml"
	} else if input == "text/csv" {
		return "csv"
	} else if input == "image/png" {
		return "png"
	} else if input == "image/jpeg" {
		return "jpeg"
	} else if input == "image/gif" {
		return "gif"
	} else if input == "image/svg+xml" {
		return "svg"
	} else if input == "application/pdf" {
		return "pdf"
	} else if input == "application/zip" {
		return "zip"
	} else if input == "application/gzip" {
		return "gzip"
	} else if input == "application/vnd.api+json" {
		return "json"
	}

	return input
}
