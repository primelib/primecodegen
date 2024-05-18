package openapicmd

import (
	"fmt"

	"github.com/primelib/primecodegen/pkg/commonmerge"
	"github.com/primelib/primecodegen/pkg/commonpatch"
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
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			if len(inputFiles) == 0 {
				log.Fatal().Msg("input specification is required")
			}
			for i, v := range inputFiles {
				inputFiles[i] = util.ResolvePath(v)
			}
			out, _ := cmd.Flags().GetString("output")
			out = util.ResolvePath(out)

			patches, _ := cmd.Flags().GetStringSlice("patch")
			patchFiles, _ := cmd.Flags().GetStringSlice("patch-file")
			log.Info().Strs("input", inputFiles).Strs("patch-ids", patches).Str("output-file", out).Msg("patching")

			// read and merge documents
			bytes, err := commonmerge.ReadAndMergeFiles(inputFiles)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to merge documents")
			}

			// patch document (external files)
			for _, patchFile := range patchFiles {
				patchedBytes, patchErr := commonpatch.ApplyPatchFile(bytes, patchFile)
				if patchErr != nil {
					log.Fatal().Err(patchErr).Msg("failed to apply patch file")
				}

				bytes = patchedBytes
			}

			// parse document
			doc, err := openapidocument.OpenDocument(bytes)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to open document")
			}
			v3doc, errs := doc.BuildV3Model()
			if len(errs) > 0 {
				log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
			}

			// patch document (built-in)
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
				outBytes, outErr := openapidocument.RenderDocument(doc)
				if outErr != nil {
					log.Fatal().Err(outErr).Msg("failed to render document")
				}
				fmt.Print(string(outBytes))
			}
		},
	}
	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringSliceP("patch", "p", []string{"generateOperationIds", "missingSchemaTitle"}, "Patch IDs to apply")
	cmd.Flags().StringSliceP("patch-file", "f", []string{}, "Patch files to apply (.patch, .jsonpatch)")
	cmd.AddCommand(PatchListCmd())

	return cmd
}
