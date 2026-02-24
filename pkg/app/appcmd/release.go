package appcmd

import (
	"log/slog"
	"os"

	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/primelib/primecodegen/pkg/app/tasks/createtag"
	"github.com/spf13/cobra"
)

func ReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-release",
		Aliases: []string{"r"},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("failed to configure platform from environment", "err", err)
			os.Exit(1)
			}

			// execute
			err = vcsapp.ExecuteTasks(platform, []taskcommon.Task{
				createtag.NewTask(),
			})
			if err != nil {
				slog.Error("failed to execute release task", "err", err)
			os.Exit(1)
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	return cmd
}
