package generator

import (
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

type Config struct {
	Directory  string
	Platform   api.Platform
	Repository api.Repository
}

type GenerateOptions struct {
	ProjectDirectory string
	OutputDirectory  string
}

// Generator provides a common interface for all generators
type Generator interface {
	Name() string                        // Name returns the name of the generator
	GetOutputName() string               // GetOutputName returns the name of the output dir for e.g. multi-language SDKs
	Generate(opts GenerateOptions) error // Generate runs the code generation
}
