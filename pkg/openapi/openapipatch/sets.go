package openapipatch

import (
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
)

type PatchSet struct {
	Id     string                            `yaml:"id"`
	Config map[string]map[string]interface{} `yaml:"config,omitempty"`
}

type PatchPreset struct {
	Id      string                  `yaml:"id"`
	Patches []sharedpatch.SpecPatch `yaml:"patches"`
}

var allPatchSets = map[string]PatchPreset{
	"code-generation": {
		Id: "code-generation",
		Patches: []sharedpatch.SpecPatch{
			PrunePathPrefixPatch.ToSpecPatch(),
			GenerateOperationIdsPatch.ToSpecPatch(),
			MergePolymorphicSchemasPatch.ToSpecPatch(),
			FlattenComponentsPatch.ToSpecPatch(),
			FixMissingSchemaTitlePatch.ToSpecPatch(),
			FixCommonPatch.ToSpecPatch(),
		},
	},
}

func ResolvePatchSets(patchSets []PatchSet) []sharedpatch.SpecPatch {
	var resolvedPatches []sharedpatch.SpecPatch

	// resolve sets
	for _, patchSet := range patchSets {
		if patchSetConfig, patchSetConfigOk := allPatchSets[patchSet.Id]; patchSetConfigOk {
			for _, patch := range patchSetConfig.Patches {
				// apply config per patch ID if available
				if cfg, cfgOk := patchSet.Config[patch.ID]; cfgOk {
					patch.Config = cfg
				}

				// apply patch
				resolvedPatches = append(resolvedPatches, patch)
			}
		}
	}

	return resolvedPatches
}
