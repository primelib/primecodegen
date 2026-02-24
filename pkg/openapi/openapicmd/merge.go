package openapicmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapimerge"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/spf13/cobra"
)

func MergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-merge",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Merge multiple OpenAPI 3 Specifications into one",
		Run: func(cmd *cobra.Command, args []string) {
			// inputs
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			if len(inputFiles) == 0 {
				slog.Error("input specification is required")
				os.Exit(1)
			}
			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")
			output = util.ResolvePath(output)
			slog.Info("Merging Specifications", "input", inputFiles, "output", output)

			// read and merge documents
			mergedSpec, err := openapimerge.MergeOpenAPI3Files(inputFiles)
			if err != nil {
				slog.Error("failed to merge api specs", "err", err)
				os.Exit(1)
			}

			// render
			rendered, err := openapidocument.RenderV3ModelFormat(mergedSpec, format)
			if err != nil {
				slog.Error("failed to render document", "err", err)
				os.Exit(1)
			}

			// output
			if output == "" {
				fmt.Println(rendered)
			} else {
				err = filesystem.SaveFileText(output, string(rendered))
				if err != nil {
					slog.Error("failed to save output file", "err", err)
					os.Exit(1)
				}
				slog.Info("Saved", "file", output)
			}
		},
	}
	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("empty", "e", "", "Empty OpenAPI 3.0 Specification (YAML or JSON for building up a clean info block)")
	cmd.Flags().StringP("format", "f", "yaml", "Output Format (yaml|json)")
	cmd.Flags().StringP("output", "o", "", "Output File (Merged Specifications)")

	return cmd
}
