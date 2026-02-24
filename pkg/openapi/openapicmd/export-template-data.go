package openapicmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GenerateTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-export-template-data",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Exports the template data usually passed to the code generator to render templates",
		Run: func(cmd *cobra.Command, args []string) {
			// validate input
			in, _ := cmd.Flags().GetString("input")
			out, _ := cmd.Flags().GetString("output")
			generatorId, _ := cmd.Flags().GetString("generator")
			patches, _ := cmd.Flags().GetStringArray("patches")
			in = util.ResolvePath(in)
			out = util.ResolvePath(out)
			if in == "" {
				slog.Error("input specification is required")
				os.Exit(1)
			}
			slog.Info("generating", "input", in, "output", out)

			// patch
			bytes, err := os.ReadFile(in)
			if err != nil {
				slog.Error("failed to read document", "err", err)
				os.Exit(1)
			}

			bytes, err = openapipatch.ApplyPatches(bytes, sharedpatch.ParsePatchSpecsFromStrings(patches))
			if err != nil {
				slog.Error("failed to apply input patches", "err", err)
				os.Exit(1)
			}

			// open document
			doc, err := openapidocument.OpenDocument(bytes)
			if err != nil {
				slog.Error("failed to open document", "err", err)
				os.Exit(1)
			}
			v3doc, err := doc.BuildV3Model()
			if err != nil {
				slog.Error("failed to build v3 high level model", "err", err)
				os.Exit(1)
			}

			// run generator
			gen, err := openapigenerator.GeneratorById(generatorId, generators)
			if err != nil {
				slog.Error("failed to find generator with provided id", "err", err, "generator-id", generatorId)
				os.Exit(1)
			}

			// build template data
			slog.Info("generating template data", "generator-id", gen.Id(), "output-file", out)
			templateData, err := gen.TemplateData(openapigenerator.TemplateDataOpts{
				Doc: v3doc,
			})
			if err != nil {
				slog.Error("failed to transform spec into template data for the generator", "err", err, "generator-id", gen.Id())
				os.Exit(1)
			}
			templateDataYaml, err := yaml.Marshal(templateData)
			if err != nil {
				slog.Error("failed to marshal template data", "err", err, "generator-id", gen.Id())
				os.Exit(1)
			}

			if out == "" {
				fmt.Print(string(templateDataYaml))
			} else {
				err = os.WriteFile(out, templateDataYaml, 0644)
				if err != nil {
					slog.Error("failed to write template data to file", "err", err)
					os.Exit(1)
				}
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().StringP("input", "i", "", "Input Specification")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringP("generator", "g", "", "Code Generation Generator ID")
	cmd.Flags().StringArray("patches", openapigenerator.DefaultCodeGenerationPatches, "Code Generation Patches")

	return cmd
}
