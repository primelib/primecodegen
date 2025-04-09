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
	"fix-oas-300-version": {
		ID:   "fix-oas-300-version",
		Func: FixOAS300Version,
	},
	"fix-oas-301-version": {
		ID:   "fix-oas-301-version",
		Func: FixOAS301Version,
	},
	"fix-oas-302-version": {
		ID:   "fix-oas-302-version",
		Func: FixOAS302Version,
	},
	"fix-oas-303-version": {
		ID:   "fix-oas-303-version",
		Func: FixOAS303Version,
	},
	"fix-oas-304-version": {
		ID:   "fix-oas-304-version",
		Func: FixOAS304Version,
	},
	"fix-oas-310-version": {
		ID:   "fix-oas-310-version",
		Func: FixOAS310Version,
	},
	"fix-oas-311-version": {
		ID:   "fix-oas-311-version",
		Func: FixOAS311Version,
	},
	"flatten-components": {
		ID:   "flatten-components",
		Func: FlattenSchemas,
	},
	"simplify-polymorphic-schemas": {
		ID:   "simplify-polymorphic-schemas",
		Func: MergePolymorphicSchemas,
	},
	"simplify-polymorphic-booleans": {
		ID:   "simplify-polymorphic-booleans",
		Func: SimplifyPolymorphicBooleans,
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
	"add-idempotency-key": {
		ID:   "add-idempotency-key",
		Func: AddIdempotencyKey,
	},
}

var EmbeddedPatchers = []sharedpatch.SpecPatch{
	// builtin transformations
	{
		Type:        "builtin",
		ID:          "fix-oas-300-version",
		Description: "Fixes specs authored in OpenAPI 3.0.0 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-301-version",
		Description: "Fixes specs authored in OpenAPI 3.0.1 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-302-version",
		Description: "Fixes specs authored in OpenAPI 3.0.2 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-303-version",
		Description: "Fixes specs authored in OpenAPI 3.0.3 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-304-version",
		Description: "Fixes specs authored in OpenAPI 3.0.4 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-310-version",
		Description: "Fixes specs authored in OpenAPI 3.1.0 format but mistakenly labeled as a different version, without converting schema content.",
	},
	{
		Type:        "builtin",
		ID:          "fix-oas-311-version",
		Description: "Fixes specs authored in OpenAPI 3.1.1 format but mistakenly labeled as a different version, without converting schema content.",
	},
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
	{
		Type:        "builtin",
		ID:          "add-idempotency-key",
		Description: "Adds an idempotency key to all POST operations in the OpenAPI document (see https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header)",
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
