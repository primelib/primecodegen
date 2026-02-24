package primelib

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/primelib/primecodegen/pkg/app/preset"
)

func Generate(dir string, conf appconf.Configuration, repository api.Repository) error {
	spec := conf.Spec
	specFile := filepath.Join(dir, conf.Spec.File)
	slog.Debug("processing module", "spec-urls", spec.UrlSlice(), "spec-file", specFile)

	// prepare generators
	generators := preset.Generators(specFile, conf)
	if len(generators) == 0 {
		return nil
	}

	// generator names
	var generatorNames []string
	var generatorOutputs []string
	for _, gen := range generators {
		generatorNames = append(generatorNames, gen.Name())
		if gen.GetOutputName() != "" && gen.GetOutputName() != "root" {
			generatorOutputs = append(generatorOutputs, gen.GetOutputName())
		}
	}

	// execute generators
	slog.With("generators", generatorNames).Info("starting code generation")
	for _, gen := range generators {
		outputDir := filepath.Join(dir, conf.Output)
		if gen.GetOutputName() == "root" {
			outputDir = dir
		} else if conf.MultiLanguage() {
			outputDir = filepath.Join(outputDir, gen.GetOutputName())
		}

		slog.Info("running code generator", "generator", gen.Name(), "projectDir", dir, "outputDir", outputDir)
		err := gen.Generate(generator.GenerateOptions{
			ProjectDirectory: dir,
			OutputDirectory:  outputDir,
			GeneratorNames:   generatorNames,
			GeneratorOutputs: generatorOutputs,
		})
		if err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}
		slog.Info("code generation completed", "generator", gen.Name())
	}

	return nil
}
