package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type PrintingPressGenerator struct {
	APISpec     string                       `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf       `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf     `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf         `json:"-" yaml:"-"`
	Opts        appconf.PrintingPressOptions `json:"-" yaml:"-"`
}

func (n *PrintingPressGenerator) Name() string {
	return "printingpress"
}

func (n *PrintingPressGenerator) GetOutputName() string {
	return "printingpress"
}

func (n *PrintingPressGenerator) Generate(opts generator.GenerateOptions) error {
	slog.Info("generating printingpress", "dir", opts.OutputDirectory, "spec", n.APISpec)
	gen := generator.PrintingPressGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Args:       []string{},
		Config:     generator.PrintingPressGeneratorConfig{},
	}

	return gen.Generate(opts)
}
