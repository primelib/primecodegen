package appcmd

import (
	"github.com/primelib/primecodegen/pkg/app/appcommon"
	"github.com/spf13/cobra"
)

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-generate",
		Aliases: []string{"g"},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			expr, _ := cmd.Flags().GetString("expr")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			tasks := []string{appcommon.GenerateTaskName}
			if dir == "" {
				runRemote(channel, expr, dryRun, tasks)
			} else {
				runLocal(dir, dryRun, tasks)
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().String("dir", "", "Directory of the project for local code generation")
	cmd.Flags().StringP("channel", "c", "", "Channel")
	cmd.Flags().StringP("expr", "e", "", "Regex expression to filter repositories")
	return cmd
}
