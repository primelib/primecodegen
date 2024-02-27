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

var V3Patchers = []V3Config{
	{
		ID:             "pruneOperationTags",
		Description:    "Removes all tags from operations",
		Enabled:        true,
		CodeGeneration: false,
		Func:           PruneOperationTags,
	},
	{
		ID:             "pruneCommonOperationIdPrefix",
		Description:    "Removes common prefixes from operation IDs",
		Enabled:        true,
		CodeGeneration: false,
		Func:           PruneCommonOperationIdPrefix,
	},
	{
		ID:             "generateOperationIds",
		Description:    "Generates operation IDs for all operations (overwrites existing IDs)",
		Enabled:        true,
		CodeGeneration: false,
		Func:           GenerateOperationIds,
	},
	{
		ID:             "flattenSchemas",
		Description:    "Flattens inline request bodies and response schemas into the components section of the document",
		Enabled:        true,
		CodeGeneration: true,
		Func:           FlattenSchemas,
	},
	{
		ID:             "missingSchemaTitle",
		Description:    "Adds a title to all schemas that are missing a title",
		Enabled:        true,
		CodeGeneration: true,
		Func:           MissingSchemaTitle,
	},
}
