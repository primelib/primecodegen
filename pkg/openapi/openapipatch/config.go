package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
)

type V3Config struct {
	ID   string
	Func func(doc *libopenapi.DocumentModel[v3.Document]) error
}

var V3Patchers = map[string]V3Config{
	"flatten-components": {
		ID:   "flatten-components",
		Func: FlattenSchemas,
	},
	"simplify-polymorphic-schemas": {
		ID:   "simplify-polymorphic-schemas",
		Func: MergePolymorphicSchemas,
	},
	"fix-operation-tags": {
		ID:   "fix-operation-tags",
		Func: RepairOperationTags,
	},
	"fix-missing-schema-title": {
		ID:   "fix-missing-schema-title",
		Func: MissingSchemaTitle,
	},
	"fix-remove-common-operation-id-prefix": {
		ID:   "fix-remove-common-operation-id-prefix",
		Func: PruneCommonOperationIdPrefix,
	},
	"prune-operation-tags-keep-first": {
		ID:   "prune-operation-tags-keep-first",
		Func: PruneOperationTagsExceptFirst,
	},
	"prune-operation-tags": {
		ID:   "pruneOperationTags",
		Func: PruneOperationTags,
	},
	"prune-invalid-paths": {
		ID:   "prune-invalid-paths",
		Func: PruneInvalidPaths,
	},
	"generate-tag-from-doc-title": {
		ID:   "generate-tag-from-doc-title",
		Func: CreateOperationTagsFromDocTitle,
	},
	"generate-operation-id": {
		ID:   "generate-operation-id",
		Func: GenerateOperationIds,
	},
}

var EmbeddedPatchers = []sharedpatch.SpecPatch{
	// builtin transformations
	{
		Type:        "builtin",
		ID:          "flatten-components",
		Description: "Flattens inline request bodies and response schemas into the components section of the document",
	},
	{
		Type:        "builtin",
		ID:          "simplify-polymorphic-schemas",
		Description: "Merges polymorphic schemas (oneOf, anyOf, allOf) into a single schema",
	},
	{
		Type:        "builtin",
		ID:          "fix-operation-tags",
		Description: "Ensures all operations have at least one tag, and that tags are documented in the document",
	},
	{
		Type:        "builtin",
		ID:          "fix-missing-schema-title",
		Description: "Adds a title to all schemas that are missing a title",
	},
	{
		Type:        "builtin",
		ID:          "fix-remove-common-operation-id-prefix",
		Description: "Removes common prefixes from operation IDs",
	},
	{
		Type:        "builtin",
		ID:          "prune-operation-tags-keep-first",
		Description: "Removes all tags from operations except the first one",
	},
	{
		Type:        "builtin",
		ID:          "prune-operation-tags",
		Description: "Removes all tags from operations",
	},
	{
		Type:        "builtin",
		ID:          "prune-invalid-paths",
		Description: "Removes all paths that are invalid (e.g. empty path, path with invalid characters)",
	},
	{
		Type:        "builtin",
		ID:          "generate-tag-from-doc-title",
		Description: "Removes all tags and createsone tag based on the document title, useful when merging multiple specs",
	},
	{
		Type:        "builtin",
		ID:          "generate-operation-id",
		Description: "Generates operation IDs for all operations (overwrites existing IDs)",
	},
	// speakeasy transformations
	{
		Type:        "speakeasy",
		ID:          "remove-unused",
		Description: "Given an OpenAPI file, remove all unused options",
	},
	{
		Type:        "speakeasy",
		ID:          "cleanup",
		Description: "Cleanup the formatting of a given OpenAPI document",
	},
	{
		Type:        "speakeasy",
		ID:          "format",
		Description: "Format an OpenAPI document to be more human-readable",
	},
	{
		Type:        "speakeasy",
		ID:          "normalize",
		Description: "Normalize an OpenAPI document to be more human-readable",
	},
}
