package krakend

import (
	"strings"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ensureSpecName relaxes the global info.x-name requirement for KrakenD.
// The name is only used for the gateway display name and the telemetry service name, so fall back to info.title (or leave it empty, which the config builders handle).
func ensureSpecName(doc *libopenapi.DocumentModel[v3.Document]) {
	if doc == nil {
		return
	}
	if doc.Model.Info == nil {
		doc.Model.Info = &base.Info{}
	}
	if doc.Model.Info.Extensions == nil {
		doc.Model.Info.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if node, ok := doc.Model.Info.Extensions.Get("x-name"); ok && node != nil && strings.TrimSpace(node.Value) != "" {
		return
	}

	doc.Model.Info.Extensions.Set("x-name", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: strings.TrimSpace(doc.Model.Info.Title)})
}
