package openapipatch

import (
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

func GenerateOperationIds(doc *libopenapi.DocumentModel[v3.Document]) error {
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			originalOperationId := op.Value.OperationId
			op.Value.OperationId = util.ToOperationId(op.Key, url)

			log.Trace().Str("path", strings.ToUpper(op.Key)+" "+url).Str("operation-id", op.Value.OperationId).Str("original-operation-id", originalOperationId).Msg("replacing operation id with generated id")
		}
	}

	return nil
}
