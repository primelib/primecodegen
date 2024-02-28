package openapicmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
			// list patches
			list, _ := cmd.Flags().GetBool("list")
			if list {
				listPatches()
				return
			}

			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
			patches, _ := cmd.Flags().GetStringSlice("patch")
			in = util.ResolvePath(in)
			out = util.ResolvePath(out)
			if in == "" {
				log.Fatal().Msg("input specification is required")
			}
			log.Info().Str("input", in).Strs("patch-ids", patches).Str("output-file", out).Msg("patching")

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
			doc, v3doc, err = openapipatch.PatchV3(patches, doc, v3doc)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to patch document")
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
	cmd.Flags().StringP("input", "i", "", "Input Specification")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringSliceP("patch", "p", []string{"generateOperationIds", "missingSchemaTitle"}, "Patch IDs to apply")
	cmd.Flags().BoolP("list", "l", false, "List available patches")

	return cmd
}

func listPatches() {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintf(w, "ID\tDescription\n")
	for _, t := range openapipatch.V3Patchers {
		fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Description)
	}
	w.Flush()
}
