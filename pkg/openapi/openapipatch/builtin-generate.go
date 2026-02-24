package openapipatch

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/logging"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/util"
)

var GenerateTagFromDocTitlePatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "generate-tag-from-doc-title",
	Description:         "Removes all tags and createsone tag based on the document title, useful when merging multiple specs",
	PatchV3DocumentFunc: GenerateTagFromDocTitle,
}

// GenerateTagFromDocTitle removes all tags and creates one new tag per API spec doc from document title setting it on each operation.
// Note: This patch must be applied before merging specs.
func GenerateTagFromDocTitle(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	err := PruneDocumentTags(doc, make(map[string]interface{}))
	if err != nil {
		return err
	}
	err = PruneOperationTags(doc, make(map[string]interface{}))
	if err != nil {
		return err
	}

	specTitle := openapidocument.SpecTitle(doc, "default")
	doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: specTitle, Description: "See document description"})

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				logging.Trace("operation is missing tags, adding default tag:", "path", strings.ToUpper(op.Key)+" "+path.Key, "tag", specTitle)
				op.Value.Tags = append(op.Value.Tags, specTitle)
			} else {
				slog.Warn("Found non-empty operation tag - ", "Operation Tag", op.Value.Tags)
			}
		}
	}

	return nil
}

var GenerateOperationIdsPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "generate-operation-id",
	Description:         "Generates operation IDs for all operations (overwrites existing IDs)",
	PatchV3DocumentFunc: GenerateOperationIds,
}

func GenerateOperationIds(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// validate config
	trimPrefix, _ := getOptionalStringConfig(config, "trim-prefix")

	// call
	return generateOperationIds(doc, true, trimPrefix)
}

var GenerateMissingOperationIdsPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "generate-missing-operation-id",
	Description:         "Generates operation IDs for all operations that are missing an ID (does not overwrite existing IDs)",
	PatchV3DocumentFunc: GenerateMissingOperationIds,
}

func GenerateMissingOperationIds(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// validate config
	trimPrefix, _ := getOptionalStringConfig(config, "trim-prefix")

	// call
	return generateOperationIds(doc, false, trimPrefix)
}

func generateOperationIds(doc *libopenapi.DocumentModel[v3.Document], replaceExisting bool, trimPrefix string) error {
	var usedOperationIds []string

	if doc.Model.Paths == nil {
		return nil
	}
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if !replaceExisting && op.Value.OperationId != "" {
				usedOperationIds = append(usedOperationIds, op.Value.OperationId)
				continue
			}

			input := strings.TrimPrefix(url, trimPrefix)
			generatedOperationId := util.ToOperationId(op.Key, input)

			if slices.Contains(usedOperationIds, generatedOperationId) {
				slog.Warn("Duplicated operation id for method", "path", url, "operation", strings.ToUpper(op.Key))
			}
			usedOperationIds = append(usedOperationIds, generatedOperationId)

			logging.Trace("replacing operation id with generated id", "path", strings.ToUpper(op.Key)+" "+url, "operation-id", generatedOperationId, "original-operation-id", op.Value.OperationId)
			op.Value.OperationId = generatedOperationId
		}
	}

	return nil
}
