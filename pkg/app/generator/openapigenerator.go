package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/primelib/primecodegen/pkg/app/appconf"
	"github.com/rs/zerolog/log"
)

type OpenAPIGenerator struct {
	OutputName string   `json:"-" yaml:"-"`
	APISpec    string   `json:"-" yaml:"-"`
	Args       []string `json:"-" yaml:"-"`
	Config     OpenAPIGeneratorConfig
}

type OpenAPIGeneratorConfig struct {
	GeneratorName         string                   `json:"generatorName" yaml:"generatorName"`
	InvokerPackage        string                   `json:"invokerPackage" yaml:"invokerPackage"`
	ApiPackage            string                   `json:"apiPackage" yaml:"apiPackage"`
	ModelPackage          string                   `json:"modelPackage" yaml:"modelPackage"`
	EnablePostProcessFile bool                     `json:"enablePostProcessFile" yaml:"enablePostProcessFile"`
	GlobalProperty        map[string]interface{}   `json:"globalProperty" yaml:"globalProperty"`
	AdditionalProperties  map[string]interface{}   `json:"additionalProperties" yaml:"additionalProperties"`
	IgnoreFiles           []string                 `json:"ignoreFiles" yaml:"ignoreFiles"`
	Repository            appconf.RepositoryConf   `json:"repository" yaml:"repository"`
	Maintainers           []appconf.MaintainerConf `json:"maintainers" yaml:"maintainers"`
}

// openApiGeneratorArgumentAllowList is a list of arguments that are allowed to be passed to the openapi generator
var openApiGeneratorArgumentAllowList = []string{
	// spec validation
	"--skip-validate-spec",
	// normalizer - see https://openapi-generator.tech/docs/customization/#openapi-normalizer
	"--openapi-normalizer",
	"SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=true",
	"SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=false",
	"SIMPLIFY_BOOLEAN_ENUM=true",
	"SIMPLIFY_BOOLEAN_ENUM=false",
	"SIMPLIFY_ONEOF_ANYOF=true",
	"SIMPLIFY_ONEOF_ANYOF=false",
	"ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=true",
	"ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=false",
	"REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=true",
	"REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=false",
	"REF_AS_PARENT_IN_ALLOF=true",
	"REF_AS_PARENT_IN_ALLOF=false",
	"REMOVE_ANYOF_ONEOF_AND_KEEP_PROPERTIES_ONLY=true",
	"REMOVE_ANYOF_ONEOF_AND_KEEP_PROPERTIES_ONLY=false",
	"KEEP_ONLY_FIRST_TAG_IN_OPERATION=true",
	"KEEP_ONLY_FIRST_TAG_IN_OPERATION=false",
	"SET_TAGS_FOR_ALL_OPERATIONS=true",
	"SET_TAGS_FOR_ALL_OPERATIONS=false",
	"DISABLE_ALL=true",
}

func (n *OpenAPIGenerator) Name() string {
	return "openapi-generator"
}

func (n *OpenAPIGenerator) GetOutputName() string {
	return n.OutputName
}

func (n *OpenAPIGenerator) Generate(opts GenerateOptions) error {
	// create dir
	_ = os.MkdirAll(opts.OutputDirectory, os.ModePerm)

	// write ignore file
	err := n.writeIgnoreFilesFile(opts)
	if err != nil {
		return err
	}

	// generate
	err = n.generateCode(opts)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	return nil
}

func (n *OpenAPIGenerator) writeIgnoreFilesFile(opts GenerateOptions) error {
	ignoreFile := filepath.Join(opts.OutputDirectory, ".openapi-generator-ignore")

	if len(n.Config.IgnoreFiles) > 0 {
		err := os.WriteFile(ignoreFile, []byte(strings.Join(n.Config.IgnoreFiles, "\n")), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *OpenAPIGenerator) generateCode(opts GenerateOptions) error {
	// auto generate config
	tempConfigFile, tmpErr := os.CreateTemp("", "openapi-generator.json")
	if tmpErr != nil {
		return fmt.Errorf("failed to create temporary config openapi-generator.json: %w", tmpErr)
	}
	defer tempConfigFile.Close()

	// config
	configFile := path.Join(opts.OutputDirectory, "openapi-generator.json")
	if _, fileErr := os.Stat(configFile); os.IsNotExist(fileErr) {
		// set defaults and missing properties
		n.Config.EnablePostProcessFile = true

		// marshal config
		bytes, err := json.MarshalIndent(n.Config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// write to temp file
		err = os.WriteFile(tempConfigFile.Name(), bytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		configFile = tempConfigFile.Name()
	}

	// default user args
	if len(n.Args) == 0 {
		n.Args = []string{
			"--openapi-normalizer", "SIMPLIFY_ANYOF_STRING_AND_ENUM_STRING=true",
			"--openapi-normalizer", "SIMPLIFY_BOOLEAN_ENUM=true",
			"--openapi-normalizer", "SIMPLIFY_ONEOF_ANYOF=true",
			"--openapi-normalizer", "ADD_UNSIGNED_TO_INTEGER_WITH_INVALID_MAX_VALUE=true",
			"--openapi-normalizer", "REFACTOR_ALLOF_WITH_PROPERTIES_ONLY=true",
		}
	}

	// all user args must be present in the allow list
	for _, arg := range n.Args {
		if !slices.Contains(openApiGeneratorArgumentAllowList, arg) {
			return fmt.Errorf("openapi generator argument not allowed: %s", arg)
		}
	}

	var args []string
	if strings.HasPrefix(n.Config.GeneratorName, "primecodegen-") {
		args = append(args, "prime-generate")
		args = append(args, "-e", "auto")
	} else {
		args = append(args, "generate")
	}

	// primecodegen bin and args
	executable := "openapi-generator-cli"
	if binPath := os.Getenv("OPENAPI_GENERATOR_BIN"); binPath != "" {
		executable = binPath
	}
	args = append(args, []string{
		"-i", n.APISpec,
		"-o", opts.OutputDirectory,
		"-c", configFile,
		"--skip-validate-spec",
	}...)
	args = append(args, n.Args...)

	cmd := exec.Command(executable, args...)
	cmd.Dir = opts.ProjectDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Trace().Str("command", cmd.String()).Msg("executing code generation")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute code generation: %w", err)
	}

	return nil
}
