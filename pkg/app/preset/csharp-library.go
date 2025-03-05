package preset

import (
	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/app/generator"
	"github.com/rs/zerolog/log"
)

type CSharpLibraryGenerator struct {
	APISpec     string                        `json:"-" yaml:"-"`
	Repository  appconf.RepositoryConf        `json:"-" yaml:"-"`
	Maintainers []appconf.MaintainerConf      `json:"-" yaml:"-"`
	Opts        appconf.CSharpLanguageOptions `json:"-" yaml:"-"`
}

func (n *CSharpLibraryGenerator) Name() string {
	return "csharp-httpclient"
}

func (n *CSharpLibraryGenerator) GetOutputName() string {
	return "csharp"
}

func (n *CSharpLibraryGenerator) Generate(opts generator.GenerateOptions) error {
	log.Info().Str("dir", opts.OutputDirectory).Str("spec", n.APISpec).Msg("generating csharp library")

	gen := generator.OpenAPIGenerator{
		OutputName: n.GetOutputName(),
		APISpec:    n.APISpec,
		Config: generator.OpenAPIGeneratorConfig{
			GeneratorName:         "csharp",
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
