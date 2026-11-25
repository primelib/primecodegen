package openapigenerator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// HaveSameCodeTypeName checks if all code types are structurally the same type (name, container kind and type arguments)
func HaveSameCodeTypeName(codeTypes []CodeType) bool {
	if len(codeTypes) == 0 {
		return false
	}
	for _, ct := range codeTypes {
		if !isSameCodeType(ct, codeTypes[0]) {
			return false
		}
	}
	return true
}

func isSameCodeType(a CodeType, b CodeType) bool {
	if a.Name != b.Name || a.IsArray != b.IsArray || a.IsList != b.IsList || a.IsMap != b.IsMap || len(a.TypeArgs) != len(b.TypeArgs) {
		return false
	}
	for i := range a.TypeArgs {
		if !isSameCodeType(a.TypeArgs[i], b.TypeArgs[i]) {
			return false
		}
	}
	return true
}

func getBoolValue(ptrToBool *bool, defaultValue bool) bool {
	if ptrToBool != nil {
		return *ptrToBool
	}
	return defaultValue
}

func uniqueSortImports(imports []string) (out []string) {
	// unique imports
	visited := make(map[string]bool)
	for _, imp := range imports {
		if _, ok := visited[imp]; !ok {
			visited[imp] = true

			if imp != "" {
				out = append(out, imp)
			}
		}
	}

	// sort imports
	slices.Sort(out)

	return out
}

func getOrDefault(extensions *orderedmap.Map[string, *yaml.Node], key string, defaultValue string) string {
	if extensions == nil {
		return defaultValue
	}

	if node, ok := extensions.Get(key); ok {
		return node.Value
	}

	return defaultValue
}

func delimiterFromStyle(style string) (string, error) {
	switch style {
	case "pipeDelimited":
		return "|", nil
	case "spaceDelimited":
		return " ", nil
	case "form", "":
		return ",", nil
	default:
		return ",", fmt.Errorf("unsupported style: %s", style)
	}
}

// IsBinaryMediaType reports whether a media type carries a raw (non-textual) payload, e.g. application/octet-stream, image/png or application/pdf.
// Wildcards (*/*) are not considered binary, as they are frequently used for structured responses.
func IsBinaryMediaType(mediaType string) bool {
	mt := strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0]))
	switch {
	case mt == "", mt == "*/*":
		return false
	case strings.HasPrefix(mt, "text/"), strings.HasPrefix(mt, "multipart/"):
		return false
	case mt == "application/json", strings.HasSuffix(mt, "+json"), strings.HasSuffix(mt, "/x-ndjson"):
		return false
	case mt == "application/xml", strings.HasSuffix(mt, "+xml"):
		return false
	case mt == "application/yaml", mt == "application/x-yaml", strings.HasSuffix(mt, "+yaml"):
		return false
	case mt == "application/x-www-form-urlencoded", mt == "application/javascript", mt == "application/graphql":
		return false
	}
	return true
}

// ContentSchema returns the schema of a request/response media type entry.
// Raw binary media types whose schema is missing or a plain string (e.g. `application/pdf: {schema: {type: string}}`)
// are normalized to `type: string, format: binary`, so generators map them to a byte type instead of a text type.
// Returns nil if no schema can be determined.
func ContentSchema(mediaType string, content *v3.MediaType) *base.Schema {
	var schema *base.Schema
	if content != nil && content.Schema != nil {
		schema = content.Schema.Schema()
	}

	if IsBinaryMediaType(mediaType) {
		if schema == nil || (slices.Contains(schema.Type, "string") && schema.Format == "" && len(schema.Enum) == 0) {
			return &base.Schema{Type: []string{"string"}, Format: "binary"}
		}
		return schema
	}

	if schema == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(mediaType)), "text/") {
		return &base.Schema{Type: []string{"string"}}
	}
	return schema
}
