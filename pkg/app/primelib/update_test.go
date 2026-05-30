package primelib

import (
	"testing"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoCodeSamplesPatchesDisabledWithoutLLMs(t *testing.T) {
	conf := appconf.Configuration{
		Output: "sdk",
		Presets: appconf.PresetConf{
			Java: appconf.JavaLanguageOptions{Enabled: true},
		},
	}

	patches := autoCodeSamplesPatches(conf)
	assert.Nil(t, patches)
}

func TestAutoCodeSamplesPatchesForEnabledPresetLanguages(t *testing.T) {
	conf := appconf.Configuration{
		Output: "sdk",
		Presets: appconf.PresetConf{
			LLMs: appconf.LLMsOptions{Enabled: true},
			Java: appconf.JavaLanguageOptions{Enabled: true},
			Go:   appconf.GoLanguageOptions{Enabled: true},
		},
	}

	patches := autoCodeSamplesPatches(conf)
	require.Len(t, patches, 2)

	assert.Equal(t, "builtin", patches[0].Type)
	assert.Equal(t, "generate-code-samples-refs", patches[0].ID)
	assert.Equal(t, "java", patches[0].Config["language"])
	assert.Equal(t, "sdk/java", patches[0].Config["dir"])

	assert.Equal(t, "builtin", patches[1].Type)
	assert.Equal(t, "generate-code-samples-refs", patches[1].ID)
	assert.Equal(t, "go", patches[1].Config["language"])
	assert.Equal(t, "sdk/go", patches[1].Config["dir"])
}

func TestAutoCodeSamplesPatchesIncludesCustomPrimeCodeGen(t *testing.T) {
	conf := appconf.Configuration{
		Output: "out",
		Presets: appconf.PresetConf{
			LLMs: appconf.LLMsOptions{Enabled: true},
		},
		Generators: []appconf.GeneratorConf{
			{
				Enabled: true,
				Name:    "java-custom",
				Type:    appconf.GeneratorTypePrimeCodeGen,
				Config: map[string]interface{}{
					"templateLanguage": "java",
				},
			},
		},
	}

	patches := autoCodeSamplesPatches(conf)
	require.Len(t, patches, 1)
	assert.Equal(t, "java", patches[0].Config["language"])
	assert.Equal(t, "out/java-custom", patches[0].Config["dir"])
}
