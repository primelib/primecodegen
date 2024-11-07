package openapicmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primelib/primecodegen/pkg/openapi/openapiconvert"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func OpenAPIConvertCmd(httpClient openapiconvert.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-convert",
		Short:   "Converts input - into output format (currently Swagger 2.0 to OpenAPI 3.0 is supported)",
		GroupID: "openapi",
		Run: func(cmd *cobra.Command, args []string) {
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			converterUrl, _ := cmd.Flags().GetString("converter-url")
			formatIn, _ := cmd.Flags().GetString("format-in")
			formatOut, _ := cmd.Flags().GetString("format-out")
			log.Info().Strs("inputfiles", inputFiles).Msg("Spec files to be converted - ")
			log.Info().Str("input format", formatIn).Str("output format", formatOut).Msg("Formats: ")
			if converterUrl != "" {
				log.Info().Str("URL", converterUrl).Msg("Converter URL")
			}
			if len(inputFiles) == 0 {
				log.Fatal().Msg("input specification is required")
			}
			if formatIn == "" || formatOut == "" {
				log.Fatal().Msg("Input - and output format is required (--format-in swagger20 --format-out openapi30)")
			}
			out, _ := cmd.Flags().GetString("output-dir")
			log.Info().Str("Dir", out).Msg("Output")
			if formatIn == "swagger20" && formatOut == "openapi30" {
				for _, path := range inputFiles {
					filebasename := filepath.Base(path)
					filename := strings.TrimSuffix(filebasename, filepath.Ext(filebasename))
					filecontent, err := os.ReadFile(path)
					check(err)

					var openAPIJSON []byte
					var convertererr error

					openAPIJSON, convertererr = openapiconvert.ConvertSwaggerToOpenAPI(filecontent, converterUrl, httpClient)
					if convertererr != nil {
						log.Fatal().Err(convertererr).Msg("Error converting to OpenAPI 3.0")
					}

					// Unmarshal JSON data into a generic map (intermediate step for YAML)
					var yamlData map[string]interface{}
					if err := json.Unmarshal(openAPIJSON, &yamlData); err != nil {
						log.Fatal().Err(err).Str("output format", formatOut).Str("json", string(openAPIJSON)).Msg("Error unmarshaling spec to YAML - ")
						return
					}
					openAPIYAML, err := yaml.Marshal(yamlData)
					if err != nil {
						log.Fatal().Err(err).Str("output format", formatOut).Msg("Error marshaling spec to YAML - ")
						return
					}
					// write document
					fileoutname := out + "/" + filename + ".yaml"
					if err := os.WriteFile(fileoutname, openAPIYAML, 0644); err != nil {
						log.Fatal().Err(err).Str("output format", formatOut).Msg("Error writing YAML file - ")
					}
					fmt.Print(string(fileoutname) + "\n")
				}
			}
		},
	}

	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("output-dir", "o", "", "Output Directory")
	cmd.Flags().StringP("converter-url", "c", "", "URL to converter service")
	cmd.Flags().StringP("format-in", "f", "swagger20", "Input format (currently swagger20 is supported)")
	cmd.Flags().StringP("format-out", "r", "openapi30", "Output format (currently openapi30 is supported)")
	return cmd
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
