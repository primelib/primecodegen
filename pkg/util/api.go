package util

import (
	"regexp"
	"strings"
)

func URLRemovePathParams(url string) string {
	re := regexp.MustCompile(`{[^}]+}`)
	return re.ReplaceAllString(url, "")
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
	}

	return input
}

func ToOperationId(method string, url string) string {
	operationID := strings.Replace(url, "/api", "", 1)
	operationID = strings.Replace(operationID, "/oauth2/", "/OAuth2/", 1)
	operationID = URLRemovePathParams(operationID)

	// get version and remove it from the operationID
	version := ParseURLAPIVersion(url)
	operationID = strings.Replace(operationID, "/v"+version+"/", "", 1)

	return strings.ToLower(method) + CapitalizeAfterChars(operationID, []int32{'/', '-', ':'}, true) + "V" + version
}
