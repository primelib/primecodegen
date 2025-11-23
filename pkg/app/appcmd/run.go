package appcmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"slices"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/primelib/primecodegen/pkg/app/appcommon"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/primelib"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-run",
		Short:   "Task: Fetch API specs, apply patches, merge files, output optimized spec",
		Aliases: []string{},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			expr, _ := cmd.Flags().GetString("expr")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			tasks := []string{appcommon.UpdateTaskName, appcommon.GenerateTaskName}
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

func runRemote(channel string, filterExpr string, dryRun bool, tasks []string) {
	// platform
	platform, err := vcsapp.GetPlatformFromEnvironment()
	if err != nil {
		slog.Error("Failed to configure platform from environment", "err", err)
		os.Exit(1)
	}

	// list repositories
	repos, err := platform.Repositories(api.RepositoryListOpts{
		IncludeBranches:   true,
		IncludeCommitHash: true,
	})
	if err != nil {
		slog.Error("Failed to list repositories", "err", err)
		os.Exit(1)
	}

	// execute task for each repository
	for _, repo := range repos {
		slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Evaluating repository for processing")

		if filterExpr != "" {
			e := regexp.MustCompile(filterExpr)
			if !e.Match([]byte(repo.Name)) {
				slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to regex mismatch")
				continue
			}
		}

		// only process repositories with a matching channel value#
		if channel != "all" && appcommon.GetChannel(platform, repo) != channel {
			slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to channel mismatch")
			continue
		}

		slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("repo_channel", channel).With("platform", platform.Name()).Info("running workflow update task")
		err = appcommon.ProcessRepository(platform, repo, dryRun, tasks)
		if err != nil {
			slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).With("err", err).Warn("Failed to process repository")
		}
	}
}

func runLocal(dir string, dryRun bool, tasks []string) {
	configPath := path.Join(dir, appconf.ConfigFileName)
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config-path", configPath).Msg("failed to read " + appconf.ConfigFileName)
	}

	// load config
	conf, err := appconf.LoadConfig(string(bytes))
	if err != nil {
		log.Fatal().Err(err).Str("config-path", configPath).Msg("failed to parse " + appconf.ConfigFileName)
	}

	// update specifications
	if slices.Contains(tasks, appcommon.UpdateTaskName) && !dryRun {
		log.Info().Str("dir", dir).Str("config", configPath).Msg("running local specification update")
		err = primelib.Update(dir, conf, api.Repository{
			Name:        conf.Repository.Name,
			Description: conf.Repository.Description,
			LicenseName: conf.Repository.LicenseName,
			LicenseURL:  conf.Repository.LicenseURL,
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to update spec")
		}
	}

	// generate code
	if slices.Contains(tasks, appcommon.GenerateTaskName) && !dryRun {
		log.Info().Str("dir", dir).Str("config", configPath).Msg("running local code generation")
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
}
