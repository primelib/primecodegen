package openapidocument

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"gopkg.in/yaml.v3"
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
		var buf bytes.Buffer
		jsonEncoder := json.NewEncoder(&buf)
		jsonEncoder.SetIndent("", "  ")
		err := jsonEncoder.Encode(doc.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to render data: %w", err)
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unsupported format: %s", format)
}

func SpecTitle(doc *libopenapi.DocumentModel[v3.Document], defaultTitle string) string {
	if doc.Model.Info != nil && doc.Model.Info.Title != "" {
		return doc.Model.Info.Title
	}

	return defaultTitle
}
