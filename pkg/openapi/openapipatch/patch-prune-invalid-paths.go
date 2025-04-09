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

func PruneUnusualPaths(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key

		segments := strings.Split(url, "/")
		shouldDelete := false

		for _, segment := range segments {
			if len(segment) > 1 && (strings.HasPrefix(segment, "_") || strings.HasSuffix(segment, "_")) {
				shouldDelete = true
				break
			}
		}

		if shouldDelete {
			doc.Model.Paths.PathItems.Delete(url)
		}
	}

	return nil
}
