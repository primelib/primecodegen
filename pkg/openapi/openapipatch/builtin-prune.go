package openapipatch

import (
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var PruneInvalidPathsPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "prune-invalid-paths",
	Description: "Removes all paths that are invalid (e.g. empty path, path with invalid characters)",
	Func:        PruneInvalidPaths,
}

func PruneInvalidPaths(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key

		if strings.HasSuffix(url, "/*") {
			doc.Model.Paths.PathItems.Delete(url)
		}
	}

	return nil
}

var PruneUnusualPathsPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "prune-unusual-paths",
	Description: "Removes all paths that are unusual (e.g. path parameters with underscores, ...)",
	Func:        PruneUnusualPaths,
}

func PruneUnusualPaths(doc *libopenapi.DocumentModel[v3.Document], config string) error {
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

var PruneDocumentTagsPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "prune-document-tags",
	Description: "Removes all tags from the document",
	Func:        PruneDocumentTags,
}

func PruneDocumentTags(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	doc.Model.Tags = nil
	return nil
}

var PruneOperationTagsPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "prune-operation-tags",
	Description: "Removes all tags from operations",
	Func:        PruneOperationTags,
}

func PruneOperationTags(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.Tags = nil
		}
	}

	return nil
}

var PruneOperationTagsExceptFirstPatch = BuiltInPatcher{
	Type:        "builtin",
	ID:          "prune-operation-tags-keep-first",
	Description: "Removes all tags from operations except the first one",
	Func:        PruneOperationTagsExceptFirst,
}

func PruneOperationTagsExceptFirst(doc *libopenapi.DocumentModel[v3.Document], config string) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) > 1 {
				op.Value.Tags = op.Value.Tags[:1]
			}
		}
	}

	return nil
}
