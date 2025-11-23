package openapipatch

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/primelib/primecodegen/pkg/util"
)

type BuiltInPatcher struct {
	Type                string `yaml:"type"`
	ID                  string `yaml:"id,omitempty"`
	Description         string `yaml:"description,omitempty"`
	PatchV3DocumentFunc func(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error
	PatchFileFunc       func(inputFile string, config map[string]interface{}) ([]byte, error)
}

func (bip BuiltInPatcher) ToSpecPatch() sharedpatch.SpecPatch {
	return sharedpatch.SpecPatch{
		Type: bip.Type,
		ID:   bip.ID,
	}
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
	FixCommonPatch,
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
	// - refactoring / modifications
	SetEndpointPatch,
	AddIdempotencyKeyPatch,
	SetOperationTagPatch,
	AddPathPrefixPatch,
	PrunePathPrefixPatch,
	AddComponentSchemaPrefixPatch,
	// speakeasy transformations
	SpeakeasyRemoveUnusedPatch,
	SpeakeasyCleanupPatch,
	SpeakeasyFormatPatch,
	SpeakeasyNormalizePatch,
}

var EmbeddedPatcherMap = util.SliceToMapWithKeyFunc(EmbeddedPatchers, func(p BuiltInPatcher) string {
	return p.Type + ":" + p.ID
})

func getStringConfig(config map[string]interface{}, key string) (string, error) {
	val, ok := config[key]
	if !ok {
		return "", fmt.Errorf("missing config key: %s", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("config key %q must be a string", key)
	}
	return s, nil
}

func getOptionalStringConfig(config map[string]interface{}, key string) (string, bool) {
	val, ok := config[key]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func getOptionalBoolConfig(config map[string]interface{}, key string) (bool, bool) {
	val, ok := config[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	if !ok {
		return false, false
	}
	return b, true
}
