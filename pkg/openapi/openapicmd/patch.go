package openapicmd

import (
	"fmt"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func PatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-patch",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Patch OpenAPI Specification for Code Generation",
		Long:    "Transform an OpenAPI Specification to be compatible with code generation tools and clean up common issues",
		Run: func(cmd *cobra.Command, args []string) {
			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
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
			for _, t := range openapipatch.V3Patchers {
				log.Debug().Str("id", t.ID).Msg("applying spec patches")
				patchErr := t.Func(v3doc)
				if patchErr != nil {
					log.Fatal().Err(patchErr).Str("id", t.ID).Msg("failed to patch document")
				}
			}

			// write document
			if out != "" {
				writeErr := openapidocument.RenderDocumentFile(doc, out)
				if writeErr != nil {
					log.Fatal().Err(writeErr).Msg("failed to write document")
				}
			} else {
				bytes, err := openapidocument.RenderDocument(doc)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to render document")
				}
				fmt.Print(string(bytes))
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().StringP("input", "i", "", "Input Specification")
	cmd.Flags().StringP("output", "o", "", "Output Directory")

	return cmd
}

// TODO: subcommand to list all options to patch the openapi specs
