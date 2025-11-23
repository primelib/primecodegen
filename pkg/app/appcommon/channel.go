package appcommon

import (
	"slices"
	"strings"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

var githubAlphaNamespaces = []string{
	"primelib",
	"philippheuer",
}

var gitlabAlphaNamespaces = []string{
	"primelib",
	"philippheuer",
}

// GetChannel returns the channel of a given repository
func GetChannel(platform api.Platform, repo api.Repository) string {
	if platform.Slug() == "github" && slices.Contains(githubAlphaNamespaces, strings.ToLower(repo.Namespace)) && slices.Contains(repo.Topics, "primecodegen-alpha") {
		return "alpha"
	} else if platform.Slug() == "gitlab" && slices.Contains(gitlabAlphaNamespaces, strings.ToLower(repo.Namespace)) && slices.Contains(repo.Topics, "primecodegen-alpha") {
		return "alpha"
	} else if slices.Contains(repo.Topics, "primecodegen-beta") {
		return "beta"
	}

	return "production"
}
