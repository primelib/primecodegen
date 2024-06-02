package openapicmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/spf13/cobra"
)

func PatchListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{},
		Short:   "List available patches",
		Run: func(cmd *cobra.Command, args []string) {
			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintf(w, "ID\tDescription\n")
			for _, t := range openapipatch.V3Patchers {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Description)
			}
			_ = w.Flush()
		},
	}

	return cmd
}
