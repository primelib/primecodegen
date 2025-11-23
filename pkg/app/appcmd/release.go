package appcmd

import (
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/primelib/primecodegen/pkg/app/tasks/createtag"
	"github.com/rs/zerolog/log"
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
				log.Fatal().Err(err).Msg("failed to configure platform from environment")
			}

			// execute
			err = vcsapp.ExecuteTasks(platform, []taskcommon.Task{
				createtag.NewTask(),
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to execute release task")
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	return cmd
}
