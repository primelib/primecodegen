package openapicmd

import (
	"fmt"
	"os"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GenerateTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-export-template-data",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Exports the template data usually passed to the code generator to render templates",
		Run: func(cmd *cobra.Command, args []string) {
			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
			generatorId, _ := cmd.Flags().GetString("generator")
			in = util.ResolvePath(in)
			out = util.ResolvePath(out)
			if in == "" {
				log.Fatal().Msg("input specification is required")
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
			doc, v3doc, err = openapipatch.PatchV3(generatorPatches, doc, v3doc)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to patch document")
			}

			// run generator
			gen, err := openapigenerator.GeneratorById(generatorId, generators)
			if err != nil {
				log.Fatal().Err(err).Str("generator-id", generatorId).Msg("failed to find generator with provided id")
			}

			// build template data
			log.Info().Str("generator-id", gen.Id()).Str("output-file", out).Msg("generating template data")
			templateData, err := gen.TemplateData(v3doc)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to transform spec into template data for the generator")
			}
			templateDataYaml, err := yaml.Marshal(templateData)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to marshal template data")
			}

			if out == "" {
				fmt.Print(string(templateDataYaml))
			} else {
				err = os.WriteFile(out, templateDataYaml, 0644)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to write template data to file")
				}
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().StringP("input", "i", "", "Input Specification")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringP("generator", "g", "", "Code Generation Generator ID")

	return cmd
}
