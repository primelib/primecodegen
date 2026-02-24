package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type KotlinLibraryGenerator struct {
	APISpec     string                        `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf        `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf      `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf          `json:"-" yaml:"-"`
	Opts        appconf.KotlinLanguageOptions `json:"-" yaml:"-"`
}

func (n *KotlinLibraryGenerator) Name() string {
	return "kotlin-httpclient"
}

func (n *KotlinLibraryGenerator) GetOutputName() string {
	return "kotlin"
}

func (n *KotlinLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	groupId, artifactId := suggestGroupAndArtifactId(n.Opts.GroupId, n.Opts.ArtifactId, n.Repository)

	slog.Info("generating kotlin library", "dir", opts.OutputDirectory, "spec", n.APISpec)
	gen := generator.PrimeCodeGenGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrimeCodeGenGeneratorConfig{
			TemplateLanguage: "kotlin",
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
