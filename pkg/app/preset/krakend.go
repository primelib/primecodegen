package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type KrakenDGenerator struct {
	APISpec     string                   `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf   `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf     `json:"-" yaml:"-"`
	Opts        appconf.KrakenDOptions   `json:"-" yaml:"-"`
}

func (n *KrakenDGenerator) Name() string {
	return "krakend-config"
}

func (n *KrakenDGenerator) GetOutputName() string {
	return "krakend"
}

func (n *KrakenDGenerator) Generate(opts generator.GenerateOptions) error {
	slog.Info("generating krakend config", "dir", opts.OutputDirectory, "spec", n.APISpec)
	gen := generator.PrimeCodeGenGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrimeCodeGenGeneratorConfig{
			TemplateLanguage: "krakend",
			TemplateType:     "config",
			Patches:          []string{},
			GroupId:          "gateway",
			ArtifactId:       "krakend",
			Repository:       n.Repository,
			Maintainers:      n.Maintainers,
			Provider:         n.Provider,
			GeneratorNames:   opts.GeneratorNames,
			GeneratorOutputs: opts.GeneratorOutputs,
		},
	}

	return gen.Generate(opts)
}
