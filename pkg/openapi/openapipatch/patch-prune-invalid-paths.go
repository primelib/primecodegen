package openapipatch

import (
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func PruneInvalidPaths(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		if strings.HasSuffix(url, "/*") {
			doc.Model.Paths.PathItems.Delete(url)
		}
	}

	return nil
}
