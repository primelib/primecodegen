package openapicmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primelib/primecodegen/pkg/openapi/openapiconvert"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func ConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-convert",
		Short:   "Convert between OpenAPI Specification formats",
		GroupID: "openapi",
		Run: func(cmd *cobra.Command, args []string) {
			// input
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			formatIn, _ := cmd.Flags().GetString("format-in")
			formatOut, _ := cmd.Flags().GetString("format-out")
			if len(inputFiles) == 0 {
				log.Fatal().Msg("input specification is required")
			}
			outputDir, _ := cmd.Flags().GetString("output-dir")

			// convert
			for _, path := range inputFiles {
				converted, err := openapiconvert.ConvertSpec(path, formatIn, formatOut)
				if err != nil {
					log.Fatal().Err(err).Str("input format", formatIn).Str("output format", formatOut).Msg("Error converting OpenAPI Specification")
				}

				// write document (stdout or file)
				filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				filePath := outputDir + "/" + filename + ".yaml"
				if outputDir == "" {
					fmt.Printf("%s", converted)
				} else {
					if err = os.WriteFile(filePath, converted, 0644); err != nil {
						log.Fatal().Err(err).Str("output format", formatOut).Msg("Error writing YAML file")
					}
					log.Info().Str("input", path).Str("output", filePath).Msg("Converted OpenAPI Specification")
				}
			}
		},
	}

	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("output-dir", "o", "", "Output Directory")
	cmd.Flags().StringP("format-in", "f", "swagger20", fmt.Sprintf("Input format (supported: %s)", strings.Join(openapiconvert.SupportedInputFormats, ", ")))
	cmd.Flags().StringP("format-out", "r", "openapi30", fmt.Sprintf("Output format (supported: %s)", strings.Join(openapiconvert.SupportedOutputFormats, ", ")))
	return cmd
}
