package openapicmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/primelib/primecodegen/pkg/loader"
	"github.com/primelib/primecodegen/pkg/openapi/openapimerge"
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
	bytes, err := loader.InterfaceToYaml(mergedSpec)
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
