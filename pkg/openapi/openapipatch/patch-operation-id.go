package openapipatch

import (
	"slices"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func GenerateOperationIds(doc *libopenapi.DocumentModel[v3.Document]) error {
	return generateOperationIds(doc, true)
}

func GenerateMissingOperationIds(doc *libopenapi.DocumentModel[v3.Document]) error {
	return generateOperationIds(doc, false)
}

func generateOperationIds(doc *libopenapi.DocumentModel[v3.Document], replaceExisting bool) error {
	var usedOperationIds []string

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if !replaceExisting && op.Value.OperationId != "" {
				usedOperationIds = append(usedOperationIds, op.Value.OperationId)
				continue
			}

			generatedOperationId := util.ToOperationId(op.Key, url)

			if slices.Contains(usedOperationIds, generatedOperationId) {
				log.Warn().Str("path", url).Str("operation", strings.ToUpper(op.Key)).Msg("Duplicated operation id for method")
			}
			usedOperationIds = append(usedOperationIds, generatedOperationId)

			op.Value.OperationId = generatedOperationId
			log.Trace().Str("path", strings.ToUpper(op.Key)+" "+url).Str("operation-id", op.Value.OperationId).Msg("replacing operation id with generated id")
		}
	}

	return nil
}

// PruneCommonOperationIdPrefix sets the operation IDs of all operations and fixes some commonly seen issues.
func PruneCommonOperationIdPrefix(doc *libopenapi.DocumentModel[v3.Document]) error {
	var operationIds []string

	// scan all current operation IDs
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			operationIds = append(operationIds, op.Value.OperationId)
		}
	}

	// detect common prefix
	commonPrefix := util.FindCommonStrPrefix(operationIds)
	if commonPrefix != "" {
		log.Trace().Str("prefix", commonPrefix).Msg("removing common prefix from operation IDs")
		for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
			for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
				op.Value.OperationId = strings.TrimPrefix(op.Value.OperationId, commonPrefix)
			}
		}
	}

	return nil
}
