package openapidocument

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func SpecTitle(doc *libopenapi.DocumentModel[v3.Document], defaultTitle string) string {
	if doc.Model.Info != nil && doc.Model.Info.Title != "" {
		return doc.Model.Info.Title
	}

	return defaultTitle
}
