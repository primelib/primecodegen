package generator

import (
	"os"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/openapi/openapicmd"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
)

type PrimeCodeGenGenerator struct {
	OutputName string   `json:"-" yaml:"-"`
	APISpec    string   `json:"-" yaml:"-"`
	Args       []string `json:"-" yaml:"-"`
	Config     PrimeCodeGenGeneratorConfig
}

type PrimeCodeGenGeneratorConfig struct {
	TemplateLanguage string                   `json:"templateLanguage" yaml:"templateLanguage"`
	TemplateType     string                   `json:"templateType" yaml:"templateType"`
	Patches          []string                 `json:"patches" yaml:"patches"`
	GroupId          string                   `json:"groupId" yaml:"groupId"`
	ArtifactId       string                   `json:"artifactId" yaml:"artifactId"`
	Repository       appconf.RepositoryConf   `json:"repository" yaml:"repository"`
	Maintainers      []appconf.MaintainerConf `json:"maintainers" yaml:"maintainers"`
	Provider         appconf.ProviderConf     `json:"provider" yaml:"provider"`
	GeneratorNames   []string                 `json:"generatorNames" yaml:"generatorNames"`
	GeneratorOutputs []string                 `json:"generatorOutputs" yaml:"generatorOutputs"`
}

// Name returns the name of the task
func (n *PrimeCodeGenGenerator) Name() string {
	return "primecodegen"
}

func (n *PrimeCodeGenGenerator) GetOutputName() string {
	return n.OutputName
}

func (n *PrimeCodeGenGenerator) Generate(opts GenerateOptions) error {
	// create dir
	_ = os.MkdirAll(opts.OutputDirectory, os.ModePerm)

	// generate
	err := n.generateCode(opts)
	if err != nil {
		return err
	}

	return nil
}

func (n *PrimeCodeGenGenerator) generateCode(opts GenerateOptions) error {
	// generate
	return openapicmd.Generate(n.APISpec, n.Config.Patches, n.Config.TemplateLanguage, n.Config.TemplateType, opts.OutputDirectory, openapigenerator.GenerateOpts{
		ArtifactGroupId:  n.Config.GroupId,
		ArtifactId:       n.Config.ArtifactId,
		RepositoryUrl:    n.Config.Repository.URL,
		LicenseName:      n.Config.Repository.LicenseName,
		LicenseUrl:       n.Config.Repository.LicenseURL,
		Provider:         n.Config.Provider,
		GeneratorNames:   n.Config.GeneratorNames,
		GeneratorOutputs: n.Config.GeneratorOutputs,
	})
}
