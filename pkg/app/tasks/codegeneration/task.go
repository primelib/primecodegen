package codegeneration

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/primelib"
	"github.com/primelib/primecodegen/pkg/app/specutil"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

const branchName = "feat/primelib-generate"

//go:embed templates/description.gohtml
var descriptionTemplate []byte

type PrimeLibGenerateTask struct{}

// Name returns the name of the task
func (n PrimeLibGenerateTask) Name() string {
	return "generate"
}

// Execute runs the task
func (n PrimeLibGenerateTask) Execute(ctx taskcommon.TaskContext) error {
	content, err := ctx.Platform.FileContent(ctx.Repository, ctx.Repository.DefaultBranch, appconf.ConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to get primelib.yaml content: %w", err)
	}

	// load config
	config, err := appconf.LoadConfig(content)
	if err != nil {
		return fmt.Errorf("failed to load primelib.yaml: %w", err)
	}

	// create temp directory (override, so we can run the modules individually)
	tempDir, err := os.MkdirTemp("", "vcs-app-*")
	if err != nil {
		return fmt.Errorf("failed to prepare temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	ctx.Directory = tempDir

	// create helper
	helper := simpletask.New(ctx)

	// clone repository
	err = helper.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	branch := branchName
	commitSuffix := ""

	// create and checkout new branch
	err = helper.CreateBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// store original spec file
	specFile := path.Join(ctx.Directory, config.Spec.File)
	originalSpecFile, err := os.CreateTemp("", "primelib-openapi-*"+filepath.Ext(specFile))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = util.CopyFile(specFile, originalSpecFile.Name())
	defer os.Remove(originalSpecFile.Name())

	// update spec
	err = primelib.Update(ctx.Directory, config, ctx.Repository)
	if err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	// generate
	err = primelib.Generate(ctx.Directory, config, ctx.Repository)
	if err != nil {
		return fmt.Errorf("failed to generate: %w", err)
	}

	// store updated spec file
	diff, err := specutil.DiffSpec("openapi", originalSpecFile.Name(), specFile)
	if err != nil {
		log.Warn().Err(err).Msg("failed to diff spec file")
	}
	if len(diff.OpenAPI) > 15 {
		diff.OpenAPI = diff.OpenAPI[:15] // limit to the first n changes, sorted by level
	}

	// commit message and description
	changes, err := helper.VCSClient.UncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to get uncommitted changes: %w", err)
	}
	filteredChanges := filterChanges(changes)
	commitMessage := fmt.Sprintf("feat: update generated code%s", commitSuffix)
	if slices.Contains(changes, specFile) {
		commitMessage = fmt.Sprintf("feat: update openapi spec%s", commitSuffix)
	}
	description, err := vcsapp.Render(string(descriptionTemplate), map[string]interface{}{
		"PlatformName": ctx.Platform.Name(),
		"PlatformSlug": ctx.Platform.Slug(),
		"Module":       config.Repository.Name,
		"SpecUpdated":  true,
		"CodeUpdated":  len(filteredChanges) > 1,
		"SpecDiff":     diff,
		"Footer":       os.Getenv("PRIMEAPP_FOOTER_HIDE") != "true",
		"FooterCustom": os.Getenv("PRIMEAPP_FOOTER_CUSTOM"),
	})
	if err != nil {
		return fmt.Errorf("failed to render description template: %w", err)
	}

	// do not commit if only .openapi-generator/FILES changed
	if len(filteredChanges) == 0 {
		log.Info().Int("total-changes", len(changes)).Int("actual-changes", len(filteredChanges)).Msg("no changes detected, skipping commit and merge request")
		return nil
	}

	// commit push and create or update merge request
	err = helper.CommitPushAndMergeRequest(commitMessage, commitMessage, string(description), "")
	if err != nil {
		return fmt.Errorf("failed to commit push and create or update merge request: %w", err)
	}

	return nil
}

func filterChanges(changes []string) []string {
	var filtered []string

	for _, change := range changes {
		// .openapi-generator/FILES is a file that is always changed, so we can ignore it
		if strings.HasSuffix(change, ".openapi-generator/FILES") {
			continue
		}

		filtered = append(filtered, change)
	}

	return filtered
}

func NewTask() PrimeLibGenerateTask {
	return PrimeLibGenerateTask{}
}

func toModuleName(input string) string {
	if input != "" && input != "root" {
		return strings.ToLower(input)
	}

	return ""
}
