package preset

import (
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/rs/zerolog/log"
)

type PythonLibraryGenerator struct {
	APISpec     string                        `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf        `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf      `json:"-" yaml:"-"`
	Provider    appconf.ProviderConf          `json:"-" yaml:"-"`
	Opts        appconf.PythonLanguageOptions `json:"-" yaml:"-"`
}

func (n *PythonLibraryGenerator) Name() string {
	return "python-httpclient"
}

func (n *PythonLibraryGenerator) GetOutputName() string {
	return "python"
}

func (n *PythonLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	log.Info().Str("dir", opts.OutputDirectory).Str("spec", n.APISpec).Msg("generating python library")

	gen := generator.OpenAPIGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Config: generator.OpenAPIGeneratorConfig{
			GeneratorName:         "python",
			EnablePostProcessFile: false,
			GlobalProperty:        nil,
			AdditionalProperties: map[string]interface{}{
				"library":        "urllib3",
				"projectName":    n.Repository.Name,
				"packageName":    n.Opts.PypiPackageName,
				"packageUrl":     n.Repository.URL,
				"packageVersion": "",
			},
			IgnoreFiles: []string{
				"README.md",
				"tox.ini",
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
