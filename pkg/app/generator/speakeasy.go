package generator

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/primelib/primecodegen/pkg/tools/speakeasycli"
	"github.com/rs/zerolog/log"
)

const speakEasyWorkflowTemplate = `
workflowVersion: 1.0.0
speakeasyVersion: latest
sources:
  input-spec:
    inputs:
    - location: {{ .SpecFile }}
targets:
  sdk:
    target: {{ .GeneratorName }}
    source: input-spec
    codeSamples:
      output: codeSamples.yaml
`

const speakEasyGeneratorTemplate = `
configVersion: 2.0.0
generation:
  sdkClassName: {{ .SDKName }}
  maintainOpenAPIOrder: true
  usageSnippets:
    optionalPropertyRendering: withExample
  useClassNamesForArrayFields: true
  fixes:
    nameResolutionDec2023: true
    nameResolutionFeb2025: false
    parameterOrderingFeb2024: true
    requestResponseComponentNamesFeb2024: true
    securityFeb2025: false
  auth:
    oAuth2ClientCredentialsEnabled: false
    oAuth2PasswordEnabled: false
python:
  version: 0.1.0
  additionalDependencies:
    dev: {}
    main: {}
  authors:
    - Speakeasy
  clientServerStatusCodesAsErrors: true
  defaultErrorName: SDKError
  description: {{ .SDKDescription }}
  enableCustomCodeRegions: false
  enumFormat: enum
  fixFlags:
    responseRequiredSep2024: false
  flattenGlobalSecurity: false
  flattenRequests: false
  flatteningOrder: parameters-first
  imports:
    option: openapi
    paths:
      callbacks: models/callbacks
      errors: models/errors
      operations: models/operations
      shared: models/shared
      webhooks: models/webhooks
  inputModelSuffix: input
  maxMethodParams: 0
  methodArguments: require-security-and-request
  outputModelSuffix: output
  packageName: speakeasy-client-sdk-python
  projectUrls: {}
  pytestTimeout: 0
  responseFormat: envelope
  templateVersion: v2
`

type SpeakEasyGenerator struct {
	OutputName string   `json:"-" yaml:"-"`
	APISpec    string   `json:"-" yaml:"-"`
	Args       []string `json:"-" yaml:"-"`
	Config     SpeakEasyGeneratorConfig
}

type SpeakEasyGeneratorConfig struct {
	TemplateLanguage string                 `json:"templateLanguage" yaml:"templateLanguage"`
	Repository       appconf.RepositoryConf `json:"repository" yaml:"repository"`
}

// Name returns the name of the task
func (n *SpeakEasyGenerator) Name() string {
	return "speakeasy"
}

func (n *SpeakEasyGenerator) GetOutputName() string {
	return n.OutputName
}

func (n *SpeakEasyGenerator) Generate(opts GenerateOptions) error {
	// create dir
	_ = os.MkdirAll(opts.OutputDirectory, os.ModePerm)

	// create config file
	_ = os.MkdirAll(path.Join(opts.OutputDirectory, ".speakeasy"), os.ModePerm)
	_ = os.WriteFile(path.Join(opts.OutputDirectory, ".speakeasy", "workflow.yaml"), []byte(processTemplate(speakEasyWorkflowTemplate, n)), os.ModePerm)
	defer os.Remove(path.Join(opts.OutputDirectory, ".speakeasy", "workflow.yaml"))
	defer os.Remove(path.Join(opts.OutputDirectory, ".speakeasy", "workflow.lock"))
	_ = os.WriteFile(path.Join(opts.OutputDirectory, ".speakeasy", "gen.yaml"), []byte(processTemplate(speakEasyGeneratorTemplate, n)), os.ModePerm)
	defer os.Remove(path.Join(opts.OutputDirectory, ".speakeasy", "gen.yaml"))
	defer os.Remove(path.Join(opts.OutputDirectory, ".speakeasy", "gen.lock"))

	// generate
	err := n.generateCode(opts)
	if err != nil {
		return err
	}

	return nil
}

func (n *SpeakEasyGenerator) generateCode(opts GenerateOptions) error {
	_, err := exec.LookPath("speakeasy")
	if err != nil {
		return speakeasycli.ErrSpeakeasyNotInstalled
	}

	// generate
	cmd := exec.Command("speakeasy",
		"run", "sdk",
		"--verbose",
		"--minimal",
		"--skip-compile",
		"--skip-testing",
		"--skip-versioning",
		"--set-version", "0.1.0",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = opts.OutputDirectory
	log.Trace().Str("cmd", cmd.String()).Msg("calling speakeasy run to generate sdk")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func processTemplate(template string, gen *SpeakEasyGenerator) string {
	template = strings.ReplaceAll(template, "{{ .SpecFile }}", gen.APISpec)
	template = strings.ReplaceAll(template, "{{ .SDKName }}", gen.Config.Repository.Name)
	template = strings.ReplaceAll(template, "{{ .SDKDescription }}", gen.Config.Repository.Description)
	template = strings.ReplaceAll(template, "{{ .GeneratorName }}", gen.Config.TemplateLanguage)
	return template
}
