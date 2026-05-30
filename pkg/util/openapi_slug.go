package util

import "strings"

func OpenAPIOperationSlug(method string, rawPath string) string {
	cleanMethod := strings.ToLower(strings.TrimSpace(method))
	cleanPath := strings.Trim(strings.TrimSpace(rawPath), "/")

	if cleanPath == "" {
		return ToSlug(cleanMethod)
	}

	segments := strings.Split(cleanPath, "/")
	cleanSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		segment = strings.TrimPrefix(segment, "{")
		segment = strings.TrimSuffix(segment, "}")
		if segment == "" {
			continue
		}
		cleanSegments = append(cleanSegments, segment)
	}

	if len(cleanSegments) == 0 {
		return ToSlug(cleanMethod)
	}

	return ToSlug(cleanMethod + "-" + strings.Join(cleanSegments, "-"))
}
