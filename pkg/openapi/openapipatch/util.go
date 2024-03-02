package openapipatch

import (
	"regexp"
	"strings"

	"github.com/primelib/primecodegen/pkg/util"
)

func findPrefix(strs []string) string {
	if len(strs) <= 1 {
		return ""
	}

	prefix := strs[0]
	for _, str := range strs[1:] {
		for !strings.HasPrefix(str, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return "" // If the prefix becomes empty, there's no common prefix
			}
		}
	}

	return prefix
}

func toOperationId(method string, url string) string {
	operationID := strings.Replace(url, "/api", "", 1)
	operationID = strings.Replace(operationID, "/oauth2/", "/OAuth2/", 1)
	operationID = removePathParams(operationID)

	// get version and remove it from the operationID
	version := extractApiVersionVersionFromUrl(url)
	operationID = strings.Replace(operationID, "/v"+version+"/", "", 1)

	return strings.ToLower(method) + util.CapitalizeAfterChars(operationID, []int32{'/', '-', ':'}, true) + "V" + version
}

func removePathParams(url string) string {
	re := regexp.MustCompile(`{[^}]+}`)
	return re.ReplaceAllString(url, "")
}

func extractApiVersionVersionFromUrl(url string) string {
	re := regexp.MustCompile(`/[vV]([0-9]+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 2 {
		return matches[1]
	}
	return "1"
}

func contentTypeToStr(input string) string {
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
	}

	return input
}
