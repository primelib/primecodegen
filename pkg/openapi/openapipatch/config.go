package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/util"
)

type BuiltInPatcher struct {
	Type        string `yaml:"type"`
	ID          string `yaml:"id,omitempty"`
	Description string `yaml:"description,omitempty"`
	Func        func(doc *libopenapi.DocumentModel[v3.Document], config string) error
}

var EmbeddedPatchers = []BuiltInPatcher{
	// builtin transformations
	// - fix oas version
	FixOAS300VersionPatch,
	FixOAS301VersionPatch,
	FixOAS302VersionPatch,
	FixOAS303VersionPatch,
	FixOAS304VersionPatch,
	FixOAS310VersionPatch,
	FixOAS311VersionPatch,
	// - fix invalid configurations / values
	FixInvalidMaxValuePatch,
	FixOperationTagsPatch,
	FixMissingSchemaTitlePatch,
	FixRemoveCommonOperationIdPrefixPatch,
	// - simplification
	FlattenComponentsPatch,
	MergePolymorphicSchemasPatch,
	SimplifyPolymorphicBooleansPatch,
	MergePolymorphicPropertiesPatch,
	SimplifyAllOfPatch,
	// - pruning
	PruneInvalidPathsPatch,
	PruneUnusualPathsPatch,
	PruneDocumentTagsPatch,
	PruneOperationTagsPatch,
	PruneOperationTagsExceptFirstPatch,
	// - generation
	GenerateTagFromDocTitlePatch,
	GenerateOperationIdsPatch,
	GenerateMissingOperationIdsPatch,
	AddIdempotencyKeyPatch,
	// speakeasy transformations
	SpeakeasyRemoveUnusedPatch,
	SpeakeasyCleanupPatch,
	SpeakeasyFormatPatch,
	SpeakeasyNormalizePatch,
}

var EmbeddedPatcherMap = util.SliceToMapWithKeyFunc(EmbeddedPatchers, func(p BuiltInPatcher) string {
	return p.Type + ":" + p.ID
})
