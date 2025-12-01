package appconf

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/patch/sharedpatch"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Output       string `yaml:"output,omitempty" jsonschema_description:"output directory for the generated code"`
	OutputSubDir bool   `yaml:"outputSubDir,omitempty" jsonschema_description:"create a subdirectory for each generator in the output directory"`

	Repository  RepositoryConf   `yaml:"repository"`
	Maintainers []MaintainerConf `yaml:"maintainers"`
	Provider    ProviderConf     `yaml:"provider"`

	Generators []GeneratorConf `yaml:"generators"` // Generators can be used to fully customize the generation process
	Presets    PresetConf      `yaml:"presets"`    // Presets are pre-configured generators for specific languages

	Spec Spec `yaml:"spec"`
}

func (c Configuration) HasGenerator() bool {
	return (c.Presets.EnabledCount() + len(c.Generators)) > 0
}

func (c Configuration) MultiLanguage() bool {
	return c.OutputSubDir || (c.Presets.EnabledCount()+len(c.Generators)) > 1
}

type RepositoryConf struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	URL           string `yaml:"url"`
	InceptionYear int    `yaml:"inceptionYear"`
	LicenseName   string `yaml:"licenseName"`
	LicenseURL    string `yaml:"licenseURL"`
}

type MaintainerConf struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
	URL   string `yaml:"url"`
}

type ProviderConf struct {
	ProductDescription string   `yaml:"productDescription"`
	Organizations      []string `yaml:"organizations"`
	Documentation      []Link   `yaml:"documentation"`
	Specifications     []Link   `yaml:"specifications"`
}

type Link struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type GeneratorConf struct {
	Enabled   bool                   `yaml:"enabled"`   // Enable the generator
	Name      string                 `yaml:"name"`      // Name of the generator
	Type      GeneratorType          `yaml:"type"`      // Type of the generator
	Arguments []string               `yaml:"arguments"` // Arguments that are passed to the generator command
	Config    map[string]interface{} `yaml:"config"`    // Config that is passed to the generator
}

// PresetConf are pre-configured generators for specific languages
type PresetConf struct {
	Scaffolding ScaffoldingOptions        `yaml:"scaffolding"`
	Go          GoLanguageOptions         `yaml:"go"`
	Java        JavaLanguageOptions       `yaml:"java"`
	Kotlin      KotlinLanguageOptions     `yaml:"kotlin"`
	Python      PythonLanguageOptions     `yaml:"python"`
	CSharp      CSharpLanguageOptions     `yaml:"csharp"`
	Typescript  TypescriptLanguageOptions `yaml:"typescript"`
}

func (c PresetConf) EnabledCount() int {
	enabledCount := 0

	if c.Go.Enabled {
		enabledCount++
	}
	if c.Java.Enabled {
		enabledCount++
	}
	if c.Kotlin.Enabled {
		enabledCount++
	}
	if c.Python.Enabled {
		enabledCount++
	}
	if c.Typescript.Enabled {
		enabledCount++
	}

	return enabledCount
}

type OpenApiGeneratorOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type PrimeCodeGenOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type ScaffoldingOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type GoLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	ModuleName string `yaml:"module"`
}

type JavaLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	GroupId    string `yaml:"groupId"`
	ArtifactId string `yaml:"artifactId"`
}

type KotlinLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	GroupId    string `yaml:"groupId"`
	ArtifactId string `yaml:"artifactId"`
}

type PythonLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	PypiPackageName string `yaml:"pypiPackageName"`
}

type CSharpLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`
}

type TypescriptLanguageOptions struct {
	Enabled     bool     `yaml:"enabled"`
	IgnoreFiles []string `yaml:"ignoreFiles"`

	NpmOrg  string `yaml:"npmOrg"`
	NpmName string `yaml:"npmName"`
}

type Spec struct {
	// File is the path to the openapi specification file
	File string `yaml:"file" default:"openapi.yaml" required:"true"`
	// SourcesDir is the directory where specifications are stored
	SourcesDir string `yaml:"sourcesDir"`
	// Sources contains one or multiple sources to specifications
	Sources []SpecSource `yaml:"sources" required:"true"`
	// Type is the format of the api specification
	Type openapidocument.SpecType `yaml:"type" required:"true"`
	// InputPatches are applied to the source specifications before merging
	InputPatches []sharedpatch.SpecPatch `yaml:"inputPatches"`
	// PatchSets are the named patch sets that are applied to the specification
	PatchSets []openapipatch.PatchSet `yaml:"patchSets"`
	// Patches are the patches that are applied to the specification
	Patches []sharedpatch.SpecPatch `yaml:"patches"`
}

func (s Spec) UrlSlice() []string {
	urls := make([]string, len(s.Sources))
	for i, u := range s.Sources {
		urls[i] = u.URL
	}
	return urls
}

func (s Spec) GetSourcesDir(rootDir string) string {
	if s.SourcesDir == "" {
		return rootDir
	}

	if filepath.IsAbs(s.SourcesDir) {
		return s.SourcesDir
	}

	return filepath.Join(rootDir, s.SourcesDir)
}

type SpecSource struct {
	File    string                     `yaml:"file"` // File path to the openapi specification
	URL     string                     `yaml:"url"`  // URL to the openapi specification
	Format  openapidocument.SourceType `yaml:"format" default:"spec"`
	Type    openapidocument.SpecType   `yaml:"type"`
	Patches []sharedpatch.SpecPatch    `yaml:"patches"` // Patches are the patches that are applied to the specification
}

type GeneratorConfig struct {
	GeneratorName         string                 `json:"generatorName" yaml:"generatorName"`
	InvokerPackage        string                 `json:"invokerPackage" yaml:"invokerPackage"`
	ApiPackage            string                 `json:"apiPackage" yaml:"apiPackage"`
	ModelPackage          string                 `json:"modelPackage" yaml:"modelPackage"`
	EnablePostProcessFile bool                   `json:"enablePostProcessFile" yaml:"enablePostProcessFile"`
	GlobalProperty        map[string]interface{} `json:"globalProperty" yaml:"globalProperty"`
	AdditionalProperties  map[string]interface{} `json:"additionalProperties" yaml:"additionalProperties"`
}

type GeneratorArgs struct {
	// UserArgs are the arguments that are passed to the generator
	OpenAPIGeneratorArgs []string `yaml:"openapi_generator"`
}

func LoadConfig(content string) (Configuration, error) {
	config := Configuration{}
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return Configuration{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// spec defaults
	for i, _ := range config.Spec.Sources {
		if config.Spec.Sources[i].Format == "" {
			config.Spec.Sources[i].Format = openapidocument.SourceTypeSpec
		}
	}
	if config.Spec.File == "" {
		config.Spec.File = "openapi.yaml"
	}
	for i := range config.Spec.InputPatches {
		defaultPatchType(&config.Spec.InputPatches[i])
	}
	for i := range config.Spec.Patches {
		defaultPatchType(&config.Spec.Patches[i])
	}
	for i := range config.Spec.Sources {
		for j := range config.Spec.Sources[i].Patches {
			defaultPatchType(&config.Spec.Sources[i].Patches[j])
		}
	}

	// auto-add specification links
	var specLinks []string
	for _, s := range config.Spec.Sources {
		if s.URL != "" {
			if slices.Contains(specLinks, s.URL) {
				continue
			}
			specLinks = append(specLinks, s.URL)
		}
	}
	for _, l := range specLinks {
		config.Provider.Specifications = append(config.Provider.Specifications, Link{
			Name: l,
			URL:  l,
		})
	}

	return config, nil
}

func defaultPatchType(p *sharedpatch.SpecPatch) {
	if p.Type == "" {
		p.Type = "builtin"
	}
}
