package preset

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/rs/zerolog/log"
)

type GoLibraryGenerator struct {
	APISpec     string                    `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf    `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf  `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf      `json:"-" yaml:"-"`
	Opts        appconf.GoLanguageOptions `json:"-" yaml:"-"`
}

func (n *GoLibraryGenerator) Name() string {
	return "go-httpclient"
}

func (n *GoLibraryGenerator) GetOutputName() string {
	return "go"
}

func (n *GoLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	moduleName := suggestGoModuleName(n.Opts.ModuleName, n.Repository, opts.ProjectDirectory, opts.OutputDirectory)

	log.Info().Str("dir", opts.OutputDirectory).Str("spec", n.APISpec).Msg("generating go library")
	gen := generator.PrimeCodeGenGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrimeCodeGenGeneratorConfig{
			TemplateLanguage: "go",
			TemplateType:     "httpclient",
			Patches:          []string{},
			ArtifactId:       moduleName,
			Repository:       n.Repository,
			Maintainers:      n.Maintainers,
			Provider:         n.Provider,
		},
	}

	return gen.Generate(opts)
}

func suggestGoModuleName(moduleName string, repository appconf.RepositoryConf, projectDirectory string, outputDirectory string) string {
	if moduleName != "" {
		return moduleName
	}

	// trim protocol prefix
	parsedURL, err := url.Parse(repository.URL)
	if err != nil {
		return "example.com/unknown-module"
	}
	moduleName = parsedURL.Host + parsedURL.Path

	// append relative path in case output directory is not the project directory
	relPath, err := filepath.Rel(projectDirectory, outputDirectory)
	if err == nil && relPath != "." {
		moduleName = filepath.Join(moduleName, relPath)
	}

	// replace all backslashes with forward slashes
	moduleName = strings.ReplaceAll(moduleName, "\\", "/")

	return moduleName
}
