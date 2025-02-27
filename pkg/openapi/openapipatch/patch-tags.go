package openapipatch

import (
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"
)

// CreateOperationTagsFromDocTitle removes all tags and creates one new tag per API spec doc from document title setting it on each operation.
// Note: This patch must be applied before merging specs.
func CreateOperationTagsFromDocTitle(doc *libopenapi.DocumentModel[v3.Document]) error {
	err := PruneDocumentTags(doc)
	if err != nil {
		return err
	}
	err = PruneOperationTags(doc)
	if err != nil {
		return err
	}

	specTitle := openapidocument.SpecTitle(doc, "default")
	doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: specTitle, Description: "See document description"})

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Str("tag", specTitle).Msg("operation is missing tags, adding default tag:")
				op.Value.Tags = append(op.Value.Tags, specTitle)
			} else {
				log.Warn().Strs("Operation Tag", op.Value.Tags).Msg("Found non-empty operation tag - ")
			}
		}
	}

	return nil
}

// RepairOperationTags ensures all operations have tags, and that tags are documented in the document
func RepairOperationTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	documentedTags := make(map[string]bool)
	for _, tag := range doc.Model.Tags {
		documentedTags[tag.Name] = true
	}

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Msg("operation is missing tags, adding default tag")
				op.Value.Tags = append(op.Value.Tags, "default")
			} else {
				// ensure all tags are documented
				for _, tag := range op.Value.Tags {
					if _, ok := documentedTags[tag]; !ok {
						log.Trace().Str("path", strings.ToUpper(op.Key)+" "+path.Key).Str("tag", tag).Msg("tag is not documented, adding to document")
						doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: tag})
						documentedTags[tag] = true
					}
				}
			}
		}
	}

	return nil
}

func PruneDocumentTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	doc.Model.Tags = nil
	return nil
}

func PruneOperationTags(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			op.Value.Tags = nil
		}
	}

	return nil
}

func PruneOperationTagsExceptFirst(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) > 1 {
				op.Value.Tags = op.Value.Tags[:1]
			}
		}
	}

	return nil
}
