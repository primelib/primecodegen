package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type ScaffoldingGenerator struct {
	APISpec     string                     `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf     `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf   `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf       `json:"-" yaml:"-"`
	Opts        appconf.ScaffoldingOptions `json:"-" yaml:"-"`
}

func (n *ScaffoldingGenerator) Name() string {
	return "scaffolding"
}

func (n *ScaffoldingGenerator) GetOutputName() string {
	return "root"
}

func (n *ScaffoldingGenerator) Generate(opts generator.GenerateOptions) error {
	slog.Info("generating scaffolding", "dir", opts.OutputDirectory, "spec", n.APISpec)
	gen := generator.PrimeCodeGenGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrimeCodeGenGeneratorConfig{
			TemplateLanguage: "default",
			TemplateType:     "scaffolding",
			Patches:          []string{},
			GroupId:          "scaffolding",
			ArtifactId:       "scaffolding",
			Repository:       n.Repository,
			Maintainers:      n.Maintainers,
			Provider:         n.Provider,
			GeneratorNames:   opts.GeneratorNames,
			GeneratorOutputs: opts.GeneratorOutputs,
		},
	}

	return gen.Generate(opts)
}
