package openapipatch

import (
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
)

type PatchSet struct {
	Name    string                  `yaml:"name"`
	Patches []sharedpatch.SpecPatch `yaml:"patches"`
}

var patchSets = []PatchSet{
	{
		Name: "code-generation",
		Patches: []sharedpatch.SpecPatch{
			GenerateOperationIdsPatch.ToSpecPatch(),
			MergePolymorphicSchemasPatch.ToSpecPatch(),
			FlattenComponentsPatch.ToSpecPatch(),
			FixMissingSchemaTitlePatch.ToSpecPatch(),
			FixCommonPatch.ToSpecPatch(),
		},
	},
}

func ResolvePatchSets(patchSetNames []string) []sharedpatch.SpecPatch {
	var resolvedPatches []sharedpatch.SpecPatch
	for _, setName := range patchSetNames {
		for _, ps := range patchSets {
			if ps.Name == setName {
				resolvedPatches = append(resolvedPatches, ps.Patches...)
			}
		}
	}
	return resolvedPatches
}
