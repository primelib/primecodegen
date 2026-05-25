package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type LLMSGenerator struct {
	APISpec     string                       `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf       `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf     `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf         `json:"-" yaml:"-"`
	Opts        appconf.PrintingPressOptions `json:"-" yaml:"-"`
}

func (n *LLMSGenerator) Name() string {
	return "llms"
}

func (n *LLMSGenerator) GetOutputName() string {
	return "llms"
}

func (n *LLMSGenerator) Generate(opts generator.GenerateOptions) error {
	slog.Info("generating printingpress llms", "dir", opts.OutputDirectory, "spec", n.APISpec)
	gen := generator.PrintingPressGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config: generator.PrintingPressGeneratorConfig{
			NoHtml: true,
			NoJson: true,
		},
	}

	return gen.Generate(opts)
}
