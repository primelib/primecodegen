package appcommon

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/gosimple/slug"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/primelib"
	"github.com/primelib/primecodegen/pkg/app/specutil"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

//go:embed templates/description.gohtml
var descriptionTemplate []byte

func ProcessRepository(platform api.Platform, repo api.Repository, dryRun bool, tasks []string) error {
	// create temp directory
	tempDir, err := os.MkdirTemp("", "primecodegen-app-*")
	if err != nil {
		return fmt.Errorf("failed to prepare temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// run task task
	taskContext := taskcommon.TaskContext{
		Directory:  filepath.Join(tempDir, slug.Make(repo.Name)),
		Platform:   platform,
		Repository: repo,
	}
	err = os.MkdirAll(taskContext.Directory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := taskContext.Platform.FileContent(taskContext.Repository, taskContext.Repository.DefaultBranch, appconf.ConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to get %s content: %w", appconf.ConfigFileName, err)
	}

	// load config
	conf, err := appconf.LoadConfig(content)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", appconf.ConfigFileName, err)
	}

	// create helper
	helper := simpletask.New(taskContext)

	// clone repository
	err = helper.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// create and checkout new branch
	branch := branchName
	err = helper.CreateBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// store original spec file
	specFile := path.Join(taskContext.Directory, conf.Spec.File)
	originalSpecFile, err := os.CreateTemp("", "primelib-openapi-*"+filepath.Ext(conf.Spec.File))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	err = util.CopyFile(specFile, originalSpecFile.Name())
	if err != nil {
		return fmt.Errorf("failed to copy spec file: %w", err)
	}
	defer os.Remove(originalSpecFile.Name())

	// update spec
	if slices.Contains(tasks, UpdateTaskName) {
		err = primelib.Update(taskContext.Directory, conf, taskContext.Repository)
		if err != nil {
			return fmt.Errorf("failed to update spec: %w", err)
		}
	}

	// diff spec files
	diff, err := specutil.DiffSpec("openapi", originalSpecFile.Name(), specFile)
	if err != nil {
		slog.With("err", err).Warn("failed to diff spec files")
	}

	// code generation
	if slices.Contains(tasks, GenerateTaskName) {
		err = primelib.Generate(taskContext.Directory, conf, taskContext.Repository)
		if err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}
	}

	// commit message and description
	commitMessage := "feat: update spec"
	if conf.HasGenerator() {
		commitMessage = "feat: update spec and generated code"
	}
	description, err := vcsapp.Render(string(descriptionTemplate), MergeRequestTemplateData{
		PlatformName: taskContext.Platform.Name(),
		PlatformSlug: taskContext.Platform.Slug(),
		Name:         conf.Repository.Name,
		SpecDiff:     &diff,
		Footer:       os.Getenv("PRIMEAPP_FOOTER_HIDE") != "true",
		FooterCustom: os.Getenv("PRIMEAPP_FOOTER_CUSTOM"),
	})
	if err != nil {
		return fmt.Errorf("failed to render description template: %w", err)
	}

	changes, err := helper.VCSClient.UncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to get uncommitted changes: %w", err)
	}
	if len(changes) == 0 {
		log.Info().Int("total-changes", len(changes)).Msg("no changes detected, skipping commit and merge request")
		return nil
	}

	// commit push and create or update merge request
	err = helper.CommitPushAndMergeRequest(commitMessage, commitMessage, string(description), "")
	if err != nil {
		return fmt.Errorf("failed to commit push and create or update merge request: %w", err)
	}

	return nil
}
