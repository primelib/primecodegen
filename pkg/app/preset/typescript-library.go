package preset

import (
	"log/slog"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
)

type TypeScriptLibraryGenerator struct {
	APISpec     string                            `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf            `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf          `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf              `json:"-" yaml:"-"`
	Opts        appconf.TypescriptLanguageOptions `json:"-" yaml:"-"`
}

func (n *TypeScriptLibraryGenerator) Name() string {
	return "typescript-httpclient"
}

func (n *TypeScriptLibraryGenerator) GetOutputName() string {
	return "typescript"
}

func (n *TypeScriptLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	slog.Info("generating python library", "dir", opts.OutputDirectory, "spec", n.APISpec)

	gen := generator.OpenAPIGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Config: generator.OpenAPIGeneratorConfig{
			GeneratorName:         "typescript-axios",
			EnablePostProcessFile: false,
			GlobalProperty:        nil,
			AdditionalProperties: map[string]interface{}{
				"projectName": n.Repository.Name,
			},
			IgnoreFiles: []string{
				"README.md",
				".travis.yml",
				"appveyor.yml",
				".gitlab-ci.yml",
				".gitignore",
				"git_push.sh",
				".github/*",
				"docs/*",
			},
			Repository:  n.Repository,
			Maintainers: n.Maintainers,
		},
	}

	return gen.Generate(opts)
}
