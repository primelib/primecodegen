package appcmd

import (
	"github.com/primelib/primecodegen/pkg/app/appcommon"
	"github.com/spf13/cobra"
)

func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-update",
		Short:   "Task: Fetch API specs, apply patches, merge files, output optimized spec",
		Aliases: []string{"u"},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			expr, _ := cmd.Flags().GetString("expr")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			tasks := []string{appcommon.UpdateTaskName}
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
