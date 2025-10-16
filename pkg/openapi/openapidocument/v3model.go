package openapidocument

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"
)

func RenderV3ModelFormat(doc *libopenapi.DocumentModel[v3.Document], format string) ([]byte, error) {
	if format == "yaml" {
		var buf bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&buf)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(doc.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to render data: %w", err)
		}

		return buf.Bytes(), nil
	} else if format == "json" {
		output, err := doc.Model.RenderJSON("  ")
		if err != nil {
			return nil, fmt.Errorf("failed to render data: %w", err)
		}

		return output, nil
	}

	return nil, fmt.Errorf("unsupported format: %s", format)
}

// ConvertDocument converts an OpenAPI document byte array to the specified format (yaml or json)
func ConvertDocument(bytes []byte, format string) ([]byte, error) {
	slog.Debug("Converting OpenAPI document format", "format", format)

	doc, err := OpenDocument(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}

	v3doc, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build v3 high level model: %w", err)
	}

	return RenderV3ModelFormat(v3doc, format)
}

func SpecTitle(doc *libopenapi.DocumentModel[v3.Document], defaultTitle string) string {
	if doc.Model.Info != nil && doc.Model.Info.Title != "" {
		return doc.Model.Info.Title
	}

	return defaultTitle
}
