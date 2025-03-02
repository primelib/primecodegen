package openapicmd

import (
	"fmt"
	"os"

	openapi_go "github.com/primelib/primecodegen/pkg/generator/openapi-go"
	openapi_java "github.com/primelib/primecodegen/pkg/generator/openapi-java"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var generators = []openapigenerator.CodeGenerator{
	openapi_go.NewGenerator(),
	openapi_java.NewGenerator(),
}

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-generate",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Generates code based on the requested generator and template",
		Run: func(cmd *cobra.Command, args []string) {
			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
			generatorId, _ := cmd.Flags().GetString("generator")
			templateId, _ := cmd.Flags().GetString("template")
			patches, _ := cmd.Flags().GetStringArray("patches")
			in = util.ResolvePath(in)
			out = util.ResolvePath(out)
			if in == "" {
				log.Fatal().Msg("input specification is required")
			}
			if out == "" {
				log.Fatal().Msg("output directory is required")
			}
			log.Info().Str("input", in).Str("output", out).Msg("generating")

			// metadata
			metadataGroupId, _ := cmd.Flags().GetString("md-group-id")
			metadataArtifactId, _ := cmd.Flags().GetString("md-artifact-id")
			metadataRepositoryUrl, _ := cmd.Flags().GetString("md-repository-url")
			metadataLicenseName, _ := cmd.Flags().GetString("md-license-name")
			metadataLicenseUrl, _ := cmd.Flags().GetString("md-license-url")

			// patch
			bytes, err := os.ReadFile(in)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to read document")
			}

			bytes, err = openapipatch.ApplyPatches(bytes, patches)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to apply input patches")
			}

			// open document
			doc, err := openapidocument.OpenDocument(bytes)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to open document")
			}
			v3doc, errs := doc.BuildV3Model()
			if len(errs) > 0 {
				log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
			}

			// print final spec
			if os.Getenv("PRIMECODEGEN_DEBUG_SPEC") == "true" {
				out, _ := v3doc.Model.Render()
				fmt.Print(string(out))
			}

			// run generator
			gen, err := openapigenerator.GeneratorById(generatorId, generators)
			if err != nil {
				log.Fatal().Err(err).Str("generator-id", generatorId).Str("template", templateId).Msg("failed to run generator")
			}
			generatorOpts := openapigenerator.GenerateOpts{
				DryRun:          false,
				Doc:             v3doc,
				OutputDir:       out,
				TemplateId:      templateId,
				ArtifactGroupId: metadataGroupId,
				ArtifactId:      metadataArtifactId,
				RepositoryUrl:   metadataRepositoryUrl,
				LicenseName:     metadataLicenseName,
				LicenseUrl:      metadataLicenseUrl,
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
	cmd.Flags().StringArray("patches", openapigenerator.DefaultCodeGenerationPatches, "Code Generation Patches")
	cmd.Flags().String("md-group-id", "", "Artifact Group ID")
	cmd.Flags().String("md-artifact-id", "", "Artifact ID")
	cmd.Flags().String("md-repository-url", "", "Repository URL (without protocol or .git suffix)")
	cmd.Flags().String("md-license-name", "", "License Name")
	cmd.Flags().String("md-license-url", "", "License URL")

	return cmd
}
