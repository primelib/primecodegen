package openapicmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapimerge"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/patch"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/spf13/cobra"
)

func PatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-patch",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Patch OpenAPI Specification for Code Generation",
		Long: util.TrimSpaceEachLine(`
			Patch OpenAPI Specification(s) to improve quality and compatibility with code generation tooling.

			Each patch must be specified as <patchType>:<patchFile>, where:
			  - <patchType> is the type of patch (e.g., "openapi-overlay").
			  - <patchFile> is the path to the patch file.

			The following patch types are supported:
			  - openapi-overlay - https://github.com/OAI/Overlay-Specification/blob/main/versions/1.0.0.md
			  - json-patch - https://tools.ietf.org/html/rfc6902
              - git-patch - https://git-scm.com/docs/git-apply
		`),
		Run: func(cmd *cobra.Command, args []string) {
			// inputs
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			if len(inputFiles) == 0 {
				slog.Error("input specification is required")
				os.Exit(1)
			}
			inputPatches, _ := cmd.Flags().GetStringSlice("input-patch")
			out, _ := cmd.Flags().GetString("output")
			patches, _ := cmd.Flags().GetStringSlice("patch")

			// run patch command
			stdout, err := Patch(inputFiles, out, sharedpatch.ParsePatchSpecsFromStrings(inputPatches), sharedpatch.ParsePatchSpecsFromStrings(patches))
			if err != nil {
				slog.Error("failed to patch document", "err", err)
				os.Exit(1)
			} else if len(stdout) > 0 {
				fmt.Printf("%s", stdout)
			}
		},
	}

	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringSlice("input-patch", []string{}, "Patches to apply to the input specification(s) pre-merge (<patchId>, file:<name>.patch, file:<name>.jsonpatch)")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringSliceP("patch", "p", []string{}, "Patches to apply in order (<patchId>, file:<name>.patch, file:<name>.jsonpatch)")

	cmd.AddCommand(PatchListCmd())
	cmd.AddCommand(PatchValidateCmd())
	cmd.AddCommand(PatchGenerateCmd())

	return cmd
}

func PatchListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{},
		Short:   "List available patches",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			columns, _ := cmd.Flags().GetStringSlice("columns")

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"TYPE", "ID", "Description"},
				Rows:    [][]interface{}{},
			}
			for _, p := range openapipatch.EmbeddedPatchers {
				data.Rows = append(data.Rows, []interface{}{
					p.Type,
					p.ID,
					p.Description,
				})
			}

			// filter columns
			if len(columns) > 0 {
				data = clioutputwriter.FilterColumns(data, columns)
			}

			// print
			err := clioutputwriter.PrintData(os.Stdout, data, clioutputwriter.Format(format))
			if err != nil {
				slog.Error("failed to print data", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))
	cmd.Flags().StringSliceP("columns", "c", []string{}, "columns to display")

	return cmd
}

func PatchValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate <patchType>:<patchFile>",
		Aliases: []string{},
		Short:   "validate patch(es)",
		Long: util.TrimSpaceEachLine(`
			Validates patch files to ensure they conform to the expected format.

			Each patch must be specified as <patchType>:<patchFile>, where:
			  - <patchType> is the type of patch (e.g., "openapi-overlay").
			  - <patchFile> is the path to the patch file.

			The following patch types are supported:
			  - openapi-overlay - https://github.com/OAI/Overlay-Specification/blob/main/versions/1.0.0.md
			  - json-patch - https://tools.ietf.org/html/rfc6902
              - git-patch - https://git-scm.com/docs/git-apply
		`),
		Example: "validate openapi-overlay:dir/patch.yaml json-patch:dir/patch.json",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				slog.Error("no patches specified")
				os.Exit(1)
			}

			errorCount := 0
			for _, arg := range args {
				parts := strings.SplitN(arg, ":", 2)
				if len(parts) != 2 {
					slog.Error("invalid patch file syntax, expected is <patchType>:<patchFile>", "patch", arg)
					errorCount++
					continue
				}
				patchType := strings.Split(arg, ":")[0]
				patchFile := strings.Split(arg, ":")[1]

				err := patch.ValidatePatchFile(patchType, patchFile)
				if err != nil {
					slog.Error("patch is invalid", "err", err, "patchType", patchType, "patchFile", patchFile)
					errorCount++
				}
				slog.Info("patch is valid", "patchType", patchType, "patchFile", patchFile)
			}

			if errorCount > 0 {
				os.Exit(1)
			}
		},
	}

	return cmd
}

// Patch runs the patch command
//
// Parameters:
//   - inputFiles: input specification files
//   - output: output file
//   - inputPatches: patches to apply to the input specification(s) pre-merge
//   - patches: patches to apply to the merged specification
func Patch(inputFiles []string, output string, inputPatches []sharedpatch.SpecPatch, patches []sharedpatch.SpecPatch) ([]byte, error) {
	slog.Info("patching", "input", inputFiles, "input-patches", sharedpatch.SpecPatchesToStringSlice(inputPatches), "patches", sharedpatch.SpecPatchesToStringSlice(patches), "output-file", output)
	for i, v := range inputFiles {
		inputFiles[i] = util.ResolvePath(v)
	}
	output = util.ResolvePath(output)

	if len(inputPatches) > 0 {
		for _, f := range inputFiles {
			bytes, err := os.ReadFile(f)
			if err != nil {
				return nil, errors.Join(util.ErrReadDocumentFromFile, err)
			}

			bytes, err = openapipatch.ApplyPatches(bytes, inputPatches)
			if err != nil {
				return nil, errors.Join(util.ErrFailedToPatchDocument, err)
			}

			err = os.WriteFile(f, bytes, 0644)
			if err != nil {
				return nil, errors.Join(util.ErrWriteDocumentToFile, err)
			}
		}
	}

	// read and merge documents
	mergedSpec, err := openapimerge.MergeOpenAPI3Files(inputFiles)
	if err != nil {
		return nil, errors.Join(util.ErrDocumentMerge, err)
	}
	bytes, err := openapidocument.RenderV3ModelFormat(mergedSpec, "yaml")
	if err != nil {
		return nil, errors.Join(util.ErrRenderDocument, err)
	}

	// patch document
	bytes, err = openapipatch.ApplyPatches(bytes, patches)
	if err != nil {
		return nil, errors.Join(util.ErrFailedToPatchDocument, err)
	}

	// write document
	if output != "" {
		// convert
		if strings.HasSuffix(output, ".json") || strings.HasSuffix(output, ".jsonc") {
			bytes, err = openapidocument.ConvertDocument(bytes, "json")
			if err != nil {
				return nil, errors.Join(util.ErrRenderDocument, err)
			}
		} else if strings.HasSuffix(output, ".yaml") || strings.HasSuffix(output, ".yml") {
			bytes, err = openapidocument.ConvertDocument(bytes, "yaml")
			if err != nil {
				return nil, errors.Join(util.ErrRenderDocument, err)
			}
		}

		// write
		err = os.WriteFile(output, bytes, 0644)
		if err != nil {
			return nil, errors.Join(util.ErrWriteDocumentToFile, err)
		}
	} else {
		return bytes, nil
	}

	return nil, nil
}

func PatchGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{},
		Short:   "Generates a requested patch",
		Run: func(cmd *cobra.Command, args []string) {
			// inputs
			inputFile, _ := cmd.Flags().GetString("input")
			if inputFile == "" {
				slog.Error("input specification is required")
				os.Exit(1)
			}
			outputFile, _ := cmd.Flags().GetString("output")
			if outputFile == "" {
				slog.Error("output file is required")
				os.Exit(1)
			}
			if len(args) == 0 {
				slog.Error("patch id argument is required")
				os.Exit(1)
			}

			// open document
			document, err := openapidocument.OpenDocumentFile(inputFile)
			if err != nil {
				slog.Error("failed to open document", "err", err)
				os.Exit(1)
			}
			v3Model, err := document.BuildV3Model()
			if err != nil {
				slog.Error("failed to build v3 model", "err", err)
				os.Exit(1)
			}

			// generate openapi overlay
			bytes, err := openapipatch.GenerateOpenAPIOverlay(v3Model, args[0])
			if err != nil {
				slog.Error("failed to generate patch", "err", err)
				os.Exit(1)
			}

			// write patch file
			err = os.WriteFile(outputFile, bytes, 0644)
			if err != nil {
				slog.Error("failed to write patch to file", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP("input", "i", "", "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("output", "o", "", "Output File")

	return cmd
}
