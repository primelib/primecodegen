package preset

import (
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type JavaLibraryGenerator struct {
	APISpec     string                      `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf      `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf    `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf        `json:"-" yaml:"-"`
	Opts        appconf.JavaLanguageOptions `json:"-" yaml:"-"`
}

func (n *JavaLibraryGenerator) Name() string {
	return "java-httpclient"
}

func (n *JavaLibraryGenerator) GetOutputName() string {
	return "java"
}

func (n *JavaLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	groupId, artifactId := suggestGroupAndArtifactId(n.Opts.GroupId, n.Opts.ArtifactId, n.Repository)

	slog.With("dir", opts.OutputDirectory, "spec", n.APISpec).With("coordinates", groupId+":"+artifactId).Info("generating java library")
	gen := generator.PrimeCodeGenGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrimeCodeGenGeneratorConfig{
			TemplateLanguage: "java",
			TemplateType:     "httpclient",
			Patches:          []string{},
			GroupId:          groupId,
			ArtifactId:       artifactId,
			Repository:       n.Repository,
			Maintainers:      n.Maintainers,
			Provider:         n.Provider,
		},
	}

	return gen.Generate(opts)
}

func suggestGroupAndArtifactId(groupId string, artifactId string, repository appconf.RepositoryConf) (string, string) {
	if groupId != "" && artifactId != "" {
		return groupId, artifactId
	}

	parsedURL, err := url.Parse(repository.URL)
	if err != nil {
		if groupId == "" {
			groupId = "com.example"
		}
		if artifactId == "" {
			artifactId = "unknown-artifact"
		}
		return groupId, artifactId
	}

	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	hostSegments := strings.Split(parsedURL.Host, ".")
	slices.Reverse(hostSegments)

	// generate groupId if missing
	if groupId == "" {
		groupId = strings.Join(hostSegments, ".")
		if len(segments) > 1 {
			groupId += "." + strings.Join(segments[:len(segments)-1], ".")
		}
		switch {
		case strings.HasPrefix(groupId, "com.github"):
			groupId = strings.Replace(groupId, "com.github", "io.github", 1)
		case strings.HasPrefix(groupId, "com.gitlab"):
			groupId = strings.Replace(groupId, "com.gitlab", "io.gitlab", 1)
		}
		if envGroup := os.Getenv("PRIMELIB_APP_JAVA_GROUP_ID"); envGroup != "" {
			groupId = envGroup
		}
	}

	// generate artifactId if missing
	if artifactId == "" && len(segments) > 0 {
		artifactId = segments[len(segments)-1]
	} else if artifactId == "" {
		artifactId = "unknown-artifact"
	}

	return groupId, artifactId
}
