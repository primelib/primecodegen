package appcmd

import (
	"os"
	"path"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/primelib"
	"github.com/primelib/primecodegen/pkg/app/tasks/codegeneration"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-generate",
		Aliases: []string{"g"},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := cmd.Flags().GetString("dir")

			if dir == "" {
				generateApp()
			} else {
				generateLocal(dir)
			}
		},
	}
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without making any changes")
	cmd.Flags().String("dir", "", "Directory of the project for local code generation")

	return cmd
}

func generateApp() {
	// platform
	platform, err := vcsapp.GetPlatformFromEnvironment()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to configure platform from environment")
	}

	// execute
	err = vcsapp.ExecuteTasks(platform, []taskcommon.Task{
		codegeneration.NewTask(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to execute generate task")
	}
}

func generateLocal(dir string) {
	configPath := path.Join(dir, "primelib.yaml")
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config-path", configPath).Msg("failed to read primelib.yaml")
	}

	// load config
	conf, err := appconf.LoadConfig(string(bytes))
	if err != nil {
		log.Fatal().Err(err).Str("config-path", configPath).Msg("failed to parse primelib.yaml")
	}

	// for each module
	log.Info().Str("dir", dir).Str("config", configPath).Msg("running local generation")
	genErr := primelib.Generate(dir, conf, api.Repository{
		Name:        conf.Repository.Name,
		Description: conf.Repository.Description,
		LicenseName: conf.Repository.LicenseName,
		LicenseURL:  conf.Repository.LicenseURL,
	})
	if genErr != nil {
		log.Fatal().Err(genErr).Msg("failed to generate code")
	}
}
