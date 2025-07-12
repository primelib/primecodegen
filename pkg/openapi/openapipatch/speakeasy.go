package openapipatch

import (
	"errors"

	"github.com/primelib/primecodegen/pkg/tools/speakeasycli"
	"github.com/primelib/primecodegen/pkg/util"
)

var SpeakeasyRemoveUnusedPatch = BuiltInPatcher{
	Type:          "speakeasy",
	ID:            "remove-unused",
	Description:   "Given an OpenAPI file, remove all unused options",
	PatchFileFunc: SpeakeasyRemoveUnused,
}

func SpeakeasyRemoveUnused(inputFile string, config map[string]interface{}) ([]byte, error) {
	patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(inputFile, "remove-unused")
	if patchErr != nil {
		return nil, errors.Join(util.ErrFailedToPatchDocument, patchErr)
	}

	return patchedBytes, nil
}

var SpeakeasyCleanupPatch = BuiltInPatcher{
	Type:          "speakeasy",
	ID:            "cleanup",
	Description:   "Cleanup the formatting of a given OpenAPI document",
	PatchFileFunc: SpeakeasyCleanup,
}

func SpeakeasyCleanup(inputFile string, config map[string]interface{}) ([]byte, error) {
	patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(inputFile, "cleanup")
	if patchErr != nil {
		return nil, errors.Join(util.ErrFailedToPatchDocument, patchErr)
	}

	return patchedBytes, nil
}

var SpeakeasyFormatPatch = BuiltInPatcher{
	Type:          "speakeasy",
	ID:            "format",
	Description:   "Format an OpenAPI document to be more human-readable",
	PatchFileFunc: SpeakeasyFormat,
}

func SpeakeasyFormat(inputFile string, config map[string]interface{}) ([]byte, error) {
	patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(inputFile, "format")
	if patchErr != nil {
		return nil, errors.Join(util.ErrFailedToPatchDocument, patchErr)
	}

	return patchedBytes, nil
}

var SpeakeasyNormalizePatch = BuiltInPatcher{
	Type:          "speakeasy",
	ID:            "normalize",
	Description:   "Normalize an OpenAPI document to be more human-readable",
	PatchFileFunc: SpeakeasyNormalize,
}

func SpeakeasyNormalize(inputFile string, config map[string]interface{}) ([]byte, error) {
	patchedBytes, patchErr := speakeasycli.SpeakEasyTransformCommand(inputFile, "normalize")
	if patchErr != nil {
		return nil, errors.Join(util.ErrFailedToPatchDocument, patchErr)
	}

	return patchedBytes, nil
}
