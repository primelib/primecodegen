package appcommon

import "github.com/primelib/primecodegen/pkg/app/specutil"

var (
	UpdateTaskName   = "update"
	GenerateTaskName = "generate"
)

const branchName = "feat/primecodegen-update"

type MergeRequestTemplateData struct {
	PlatformName string
	PlatformSlug string
	Name         string
	SpecDiff     *specutil.Diff
	Footer       bool
	FooterCustom string
}
