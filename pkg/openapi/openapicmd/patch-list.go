package openapicmd

import (
	"fmt"
	"os"

	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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
