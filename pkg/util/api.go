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
		if paramName == "id" {
			return ""
		}
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
	}

	return input
}

func ToOperationId(method string, url string) string {
	operationID := strings.Replace(url, "/api", "", 1)
	operationID = strings.Replace(operationID, "/oauth2/", "/OAuth2/", 1)
	operationID = convertPathParameterToSingularIfFollowedByVariable(operationID)
	//operationID = URLRemovePathParams(operationID)
	operationID = URLPathParamAddByPrefix(operationID)

	// get version and remove it from the operationID
	version := ParseURLAPIVersion(url)
	operationID = strings.Replace(operationID, "/v"+version+"/", "", 1)
	operationID = strings.Replace(operationID, "*", "", 1)

	return strings.ToLower(method) + CapitalizeAfterChars(operationID, []int32{'/', '-', ':'}, true) + "V" + version
}

func convertPathParameterToSingularIfFollowedByVariable(path string) string {
	sections := strings.Split(path, "/")
	for i := 0; i < len(sections)-1; i++ {
		nextSection := sections[i+1]
		currentSection := sections[i]

		if strings.HasPrefix(nextSection, "{") {
			currentSection = toSingular(currentSection)
		}

		sections[i] = currentSection
	}
	return strings.Join(sections, "/")
}

func toSingular(word string) string {
	suffixes := map[string]string{
		"ies": "y",
		"s":   "",
	}
	irregularForms := map[string]string{}

	// irregular forms
	if val, ok := irregularForms[word]; ok {
		return val
	}

	// regular forms
	for pluralSuffix, singularSuffix := range suffixes {
		if strings.HasSuffix(word, pluralSuffix) {
			return strings.TrimSuffix(word, pluralSuffix) + singularSuffix
		}
	}

	return word
}
