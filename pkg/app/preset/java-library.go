package preset

import (
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/rs/zerolog/log"
)

type JavaLibraryGenerator struct {
	APISpec     string                      `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf      `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf    `json:"-" yaml:"-"`
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

	log.Info().Str("dir", opts.OutputDirectory).Str("spec", n.APISpec).Msg("generating java library")
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
		},
	}

	return gen.Generate(opts)
}

func suggestGroupAndArtifactId(groupId string, artifactId string, repository appconf.RepositoryConf) (string, string) {
	if groupId != "" || artifactId != "" {
		return groupId, artifactId
	}

	// split into segments
	parsedURL, err := url.Parse(repository.URL)
	if err != nil {
		return "com.example", "unknown-artifact"
	}
	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	hostSegments := strings.Split(parsedURL.Host, ".")
	slices.Reverse(hostSegments)
	if len(segments) == 0 {
		return "com.example", "unknown-artifact"
	}

	// group id
	groupId = strings.Join(hostSegments, ".")
	if len(segments) > 1 {
		groupId += "." + strings.Join(segments[:len(segments)-1], ".") // Include all but last segment
	}
	if strings.HasPrefix(groupId, "com.github") {
		groupId = strings.Replace(groupId, "com.github", "io.github", 1)
	} else if strings.HasPrefix(groupId, "com.gitlab") {
		groupId = strings.Replace(groupId, "com.gitlab", "io.gitlab", 1)
	}

	// artifact id
	artifactId = segments[len(segments)-1]

	// override with env vars
	if os.Getenv("PRIMELIB_APP_JAVA_GROUP_ID") != "" {
		groupId = os.Getenv("PRIMELIB_APP_JAVA_GROUP_ID")
	}

	return groupId, artifactId
}
