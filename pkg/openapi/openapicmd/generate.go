package openapicmd

import (
	openapi_go "github.com/primelib/primecodegen/pkg/generator/openapi-go"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var generators = []openapigenerator.CodeGenerator{
	openapi_go.NewGoGenerator(),
}

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-generate",
		Aliases: []string{},
		GroupID: "openapi",
		Run: func(cmd *cobra.Command, args []string) {
			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
			generatorId, _ := cmd.Flags().GetString("generator")
			templateId, _ := cmd.Flags().GetString("template")
			in = util.ResolvePath(in)
			out = util.ResolvePath(out)
			if in == "" {
				log.Fatal().Msg("input specification is required")
			}
			if out == "" {
				log.Fatal().Msg("output directory is required")
			}
			log.Info().Str("input", in).Str("output", out).Msg("generating")

			// open document
			doc, err := openapidocument.OpenDocumentFile(in)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to open document")
			}
			v3doc, errs := doc.BuildV3Model()
			if len(errs) > 0 {
				log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
			}

			// patch document
			for _, t := range openapipatch.V3Patchers {
				log.Debug().Str("id", t.ID).Msg("applying spec patches")
				patchErr := t.Func(v3doc)
				if patchErr != nil {
					log.Fatal().Err(patchErr).Str("id", t.ID).Msg("failed to patch document")
				}

				// reload document
				_, doc, _, errs = doc.RenderAndReload()
				if len(errs) > 0 {
					log.Fatal().Errs("spec", errs).Msgf("failed to reload document after patching")
				}
				v3doc, errs = doc.BuildV3Model()
				if len(errs) > 0 {
					log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
				}
			}

			// run generator
			gen, err := openapigenerator.GeneratorById(generatorId, generators)
			if err != nil {
				log.Fatal().Err(err).Str("generator-id", generatorId).Str("template", templateId).Msg("failed to run generator")
			}
			generatorOpts := openapigenerator.GenerateOpts{
				DryRun:     false,
				Doc:        v3doc,
				OutputDir:  out,
				TemplateId: templateId,
			}
			log.Info().Str("generator-id", gen.Id()).Str("template", templateId).Bool("dry-run", generatorOpts.DryRun).Str("output-dir", generatorOpts.OutputDir).Msg("running generator")
			err = gen.Generate(generatorOpts)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to transform spec into template data for the generator")
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().StringP("input", "i", "", "Input Specification")
	cmd.Flags().StringP("output", "o", "", "Output Directory")
	cmd.Flags().StringP("generator", "g", "", "Code Generation Generator ID")
	cmd.Flags().StringP("template", "t", "", "Code Generation Template ID")

	return cmd
}
