package createtag

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

type PrimeLibTagCreateTask struct {
}

// Name returns the name of the task
func (n PrimeLibTagCreateTask) Name() string {
	return "release"
}

// Execute runs the task
func (n PrimeLibTagCreateTask) Execute(ctx taskcommon.TaskContext) error {
	content, err := ctx.Platform.FileContent(ctx.Repository, ctx.Repository.DefaultBranch, appconf.ConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to get primelib.yaml content: %w", err)
	}

	// load config
	conf, err := appconf.LoadConfig(content)
	if err != nil {
		return fmt.Errorf("failed to load primelib.yaml: %w", err)
	}

	// requires modules
	if !conf.HasGenerator() {
		return fmt.Errorf("no generators enabeld")
	}

	// skip if auto release is disabled
	/*
		if !config.Release {
			log.Debug().Str("repo", ctx.Repository.Namespace+"/"+ctx.Repository.Name).Msg("release creation is disabled, skipping")
			return nil
		}
	*/

	// check if last tag has a release
	tagList, err := ctx.Platform.Tags(ctx.Repository, 5)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}
	for _, release := range tagList {
		if release.CommitHash == ctx.Repository.CommitHash {
			log.Debug().Msg("latest commit already has a tag, skipping")
			return nil
		}
	}

	// find the latest two releases
	var lastRelease *api.Tag
	for _, tag := range tagList {
		if lastRelease == nil {
			lastRelease = &tag
		}
	}
	log.Debug().Interface("tag", lastRelease).Msg("found last tag")

	// get next version
	nextVersion := []string{"0.1.0"}
	/*
		if lastRelease != nil {
			for _, module := range config.Modules {
				specFile := path.Join(ctx.Directory, conf.Spec.File)
				// get old version of spec file
				oldFile, err := os.CreateTemp("", "primelib-spec")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				oldContent, err := ctx.Platform.FileContent(ctx.Repository, lastRelease.CommitHash, specFile)
				if err != nil {
					return fmt.Errorf("failed to get spec file content: %w", err)
				}
				_, err = oldFile.WriteString(oldContent)
				if err != nil {
					return fmt.Errorf("failed to write to temp file: %w", err)
				}

				// get new version of spec file
				newFile, err := os.CreateTemp("", "primelib-spec")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				currentContent, err := ctx.Platform.FileContent(ctx.Repository, lastRelease.CommitHash, specFile)
				if err != nil {
					return fmt.Errorf("failed to get spec file content: %w", err)
				}
				_, err = newFile.WriteString(currentContent)

				// determinate the next version number
				version, err := spec.BumpVersion(conf.Spec.Type, oldFile.Name(), newFile.Name(), lastRelease.Name)
				if err != nil {
					return fmt.Errorf("failed to bump version: %w", err)
				}

				nextVersion = append(nextVersion, version)
			}
		}
	*/

	// find highest version
	version := util.FindHighestVersion(nextVersion)

	// create tag
	err = ctx.Platform.CreateTag(ctx.Repository, "v"+version, ctx.Repository.CommitHash, "")
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	log.Info().Str("repository", ctx.Repository.Namespace+"/"+ctx.Repository.Name).Str("tag", "v"+version).Msg("created tag")

	return nil
}

func NewTask() PrimeLibTagCreateTask {
	return PrimeLibTagCreateTask{}
}
