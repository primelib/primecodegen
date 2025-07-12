package openapipatch

import (
	"slices"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

var FixInvalidMaxValuePatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-invalid-max-value",
	Description:         "Fixes integer and long schemas where the maximum value is out of bounds for the type, e.g. max: 9223372036854775807 for long.",
	PatchV3DocumentFunc: FixInvalidMaxValue,
}

// FixInvalidMaxValue fixes integers and longs, where the maximum value is out of bounds for the type
func FixInvalidMaxValue(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			s := p.Value.Schema()
			if slices.Contains(s.Type, "integer") && p.Value.Schema().Maximum != nil {
				if *p.Value.Schema().Maximum > 2147483647 {
					p.Value.Schema().Maximum = ptr.Ptr(float64(2147483647))
					log.Trace().Str("schema", schema.Key).Str("property", p.Key).Msg("fixing maximum value for integer")
				}
			}
		}
	}

	return nil
}

var FixOperationTagsPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-operation-tags",
	Description:         "Ensures all operations have at least one tag, and that tags are documented in the document",
	PatchV3DocumentFunc: FixOperationTags,
}

// FixOperationTags ensures all operations have tags, and that tags are documented in the document
func FixOperationTags(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
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

var FixMissingSchemaTitlePatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-missing-schema-title",
	Description:         "Adds a title to all schemas that are missing a title",
	PatchV3DocumentFunc: FixMissingSchemaTitle,
}

// FixMissingSchemaTitle fills in missing schema titles with the schema key
func FixMissingSchemaTitle(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Title == "" {
			schema.Value.Schema().Title = schema.Key
			log.Trace().Str("schema", schema.Key).Msg("missing schema title, setting to schema key")
		}
	}

	return nil
}

var FixRemoveCommonOperationIdPrefixPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-remove-common-operation-id-prefix",
	Description:         "Removes common prefixes from operation IDs",
	PatchV3DocumentFunc: FixRemoveCommonOperationIdPrefix,
}

// FixRemoveCommonOperationIdPrefix sets the operation IDs of all operations and fixes some commonly seen issues.
func FixRemoveCommonOperationIdPrefix(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
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
