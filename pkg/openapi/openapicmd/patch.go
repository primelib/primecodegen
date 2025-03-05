package openapicmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/openapi/openapimerge"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/patch"
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
				log.Fatal().Msg("input specification is required")
			}
			inputPatches, _ := cmd.Flags().GetStringSlice("input-patch")
			out, _ := cmd.Flags().GetString("output")
			patches, _ := cmd.Flags().GetStringSlice("patch")

			// run patch command
			stdout, err := runPatchCmd(inputFiles, out, inputPatches, patches)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to patch document")
			} else if stdout != "" {
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
				Headers: []string{"ID", "Description"},
				Rows:    [][]interface{}{},
			}
			for _, repo := range openapipatch.V3Patchers {
				data.Rows = append(data.Rows, []interface{}{
					repo.ID,
					repo.Description,
				})
			}

			// filter columns
			if len(columns) > 0 {
				data = clioutputwriter.FilterColumns(data, columns)
			}

			// print
			err := clioutputwriter.PrintData(os.Stdout, data, clioutputwriter.Format(format))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to print data")
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
				log.Fatal().Msg("no patches specified")
			}

			errorCount := 0
			for _, arg := range args {
				parts := strings.SplitN(arg, ":", 2)
				if len(parts) != 2 {
					log.Error().Str("patch", arg).Msg("invalid patch file syntax, expected is <patchType>:<patchFile>")
					errorCount++
					continue
				}
				patchType := strings.Split(arg, ":")[0]
				patchFile := strings.Split(arg, ":")[1]

				err := patch.ValidatePatchFile(patchType, patchFile)
				if err != nil {
					log.Error().Err(err).Str("patchType", patchType).Str("patchFile", patchFile).Msg("patch is invalid")
					errorCount++
				}
				log.Info().Str("patchType", patchType).Str("patchFile", patchFile).Msg("patch is valid")
			}

			if errorCount > 0 {
				os.Exit(1)
			}
		},
	}

	return cmd
}

// runPatchCmd runs the patch command
//
// Parameters:
//   - inputFiles: input specification files
//   - output: output file
//   - inputPatches: patches to apply to the input specification(s) pre-merge
//   - patches: patches to apply to the merged specification
func runPatchCmd(inputFiles []string, output string, inputPatches, patches []string) (string, error) {
	log.Info().Strs("input", inputFiles).Strs("input-patches", inputPatches).Strs("patches", patches).Str("output-file", output).Msg("patching")
	for i, v := range inputFiles {
		inputFiles[i] = util.ResolvePath(v)
	}
	output = util.ResolvePath(output)

	if len(inputPatches) > 0 {
		for _, f := range inputFiles {
			bytes, err := os.ReadFile(f)
			if err != nil {
				return "", errors.Join(util.ErrReadDocumentFromFile, err)
			}

			bytes, err = openapipatch.ApplyPatches(bytes, inputPatches)
			if err != nil {
				return "", errors.Join(util.ErrFailedToPatchDocument, err)
			}

			err = os.WriteFile(f, bytes, 0644)
			if err != nil {
				return "", errors.Join(util.ErrWriteDocumentToFile, err)
			}
		}
	}

	// read and merge documents
	mergedSpec, err := openapimerge.MergeOpenAPI3Files(inputFiles)
	if err != nil {
		return "", errors.Join(util.ErrDocumentMerge, err)
	}
	bytes, err := loader.InterfaceToYaml(mergedSpec.Model)
	if err != nil {
		return "", errors.Join(util.ErrRenderDocument, err)
	}

	// patch document
	bytes, err = openapipatch.ApplyPatches(bytes, patches)
	if err != nil {
		return "", errors.Join(util.ErrFailedToPatchDocument, err)
	}

	// write document
	if output != "" {
		err = os.WriteFile(output, bytes, 0644)
		if err != nil {
			return "", errors.Join(util.ErrWriteDocumentToFile, err)
		}
	} else {
		return string(bytes), nil
	}

	return "", nil
}
