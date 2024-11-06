package openapicmd

import (
	"fmt"
	"os"

	"github.com/primelib/primecodegen/pkg/openapi/openapimerge"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// This command reads all files (yaml and json) and merges them into a single OpenAPI spec document.
// The output format is YAML
func MergeLibOpenAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-merge",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Merge OpenAPI specifications using libopenapi for code generation",
		Long:    "Merge OpenAPI specifications using libopenapi as stand-alone command to be compatible with code generation tools",

		Run: func(cmd *cobra.Command, args []string) {
			// inputs
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			log.Info().Strs("Input Files", inputFiles).Msg("Merging")
			if len(inputFiles) == 0 {
				log.Fatal().Msg("input specification is required")
			}
			out, _ := cmd.Flags().GetString("output")
			emptyspec, _ := cmd.Flags().GetString("empty")
			output := util.ResolvePath(out)
			mergedSpec, err := openapimerge.MergeOpenAPISpecs(emptyspec, inputFiles)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to merge api specs")
			}
			yamlDate, err := yaml.Marshal(mergedSpec.Model)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to unmarshal api specs to YAML")
			}
			// write document
			if output != "" {
				err = os.WriteFile(output, yamlDate, 0644)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to write document to file")
				}
			} else {
				fmt.Println(string(yamlDate))
				return
			}
		},
	}
	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("empty", "e", "", "Empty OpenAPI 3.0 Specification (YAML or JSON for building up a clean info block)")
	cmd.Flags().StringP("output", "o", "", "Output File (Merged Specifications)")

	return cmd
}
