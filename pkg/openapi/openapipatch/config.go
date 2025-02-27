package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type V3Config struct {
	ID             string
	Description    string
	Enabled        bool
	CodeGeneration bool // required for code generation
	Func           func(doc *libopenapi.DocumentModel[v3.Document]) error
}

var V3Patchers = map[string]V3Config{
	"createOperationTagsFromDocTitle": {
		ID:             "createOperationTagsFromDocTitle",
		Description:    "Removes all tags and creates one new tag per API spec from the document title, setting it on each operation. This patch will be applied before merging specs.",
		Enabled:        true,
		CodeGeneration: false,
		Func:           CreateOperationTagsFromDocTitle,
	},
	"fixOperationTags": {
		ID:             "repairOperationTags",
		Description:    "Ensures all operations have at least one tag, and that tags are documented in the document",
		Enabled:        true,
		CodeGeneration: false,
		Func:           RepairOperationTags,
	},
	"pruneOperationTags": {
		ID:             "pruneOperationTags",
		Description:    "Removes all tags from operations",
		Enabled:        true,
		CodeGeneration: false,
		Func:           PruneOperationTags,
	},
	"pruneOperationTagsExceptFirst": {
		ID:             "pruneOperationTagsExceptOne",
		Description:    "Removes all tags from operations except the first one",
		Enabled:        false,
		CodeGeneration: false,
		Func:           PruneOperationTagsExceptFirst,
	},
	"pruneCommonOperationIdPrefix": {
		ID:             "pruneCommonOperationIdPrefix",
		Description:    "Removes common prefixes from operation IDs",
		Enabled:        true,
		CodeGeneration: false,
		Func:           PruneCommonOperationIdPrefix,
	},
	"generateOperationIds": {
		ID:             "generateOperationIds",
		Description:    "Generates operation IDs for all operations (overwrites existing IDs)",
		Enabled:        true,
		CodeGeneration: false,
		Func:           GenerateOperationIds,
	},
	"flattenSchemas": {
		ID:             "flattenSchemas",
		Description:    "Flattens inline request bodies and response schemas into the components section of the document",
		Enabled:        true,
		CodeGeneration: true,
		Func:           FlattenSchemas,
	},
	"mergePolymorphicSchemas": {
		ID:             "mergePolymorphicSchemas",
		Description:    "Merges polymorphic schemas (oneOf, anyOf, allOf) into a single schema",
		Enabled:        true,
		CodeGeneration: true,
		Func:           MergePolymorphicSchemas,
	},
	"missingSchemaTitle": {
		ID:             "missingSchemaTitle",
		Description:    "Adds a title to all schemas that are missing a title",
		Enabled:        true,
		CodeGeneration: true,
		Func:           MissingSchemaTitle,
	},
	"pruneInvalidPaths": {
		ID:             "pruneInvalidPaths",
		Description:    "Removes all paths that are invalid (e.g. empty path, path with invalid characters)",
		Enabled:        true,
		CodeGeneration: true,
		Func:           PruneInvalidPaths,
	},
}
