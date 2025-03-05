package primelib

import (
	"fmt"
	"path/filepath"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/primelib/primecodegen/pkg/app/preset"
	"github.com/rs/zerolog/log"
)

func Generate(dir string, conf appconf.Configuration, repository api.Repository) error {
	spec := conf.Spec
	specFile := filepath.Join(dir, conf.Spec.File)
	log.Debug().Strs("spec-urls", spec.UrlSlice()).Str("spec-file", specFile).Msg("processing module")

	// prepare generators
	generators := preset.Generators(specFile, conf)

	// execute generators
	for _, gen := range generators {
		outputDir := filepath.Join(dir, conf.Output)
		if conf.MultiLanguage() {
			outputDir = filepath.Join(outputDir, gen.GetOutputName())
		}

		log.Info().Str("generator", gen.Name()).Str("projectDir", dir).Str("outputDir", outputDir).Msg("running code generator")
		err := gen.Generate(generator.GenerateOptions{
			ProjectDirectory: dir,
			OutputDirectory:  outputDir,
		})
		if err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}
		log.Info().Str("generator", gen.Name()).Msg("code generation completed")
	}

	return nil
}
