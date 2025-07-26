package templateapi

import (
	"text/template"
)

type RenderOpts struct {
	DryRun               bool                                     // DryRun will not write any files to disk
	Types                []Type                                   // Types can be used to only render files of specific types
	IgnoreFiles          []string                                 // IgnoreFiles is a list of file names that should not be rendered
	IgnoreFileCategories []string                                 // IgnoreFileCategories is a list of file categories that should not be rendered
	Properties           map[string]string                        // User-defined properties that can be used in the templates
	PostProcess          func(name string, content []byte) []byte // PostProcess is a function that can be used to post-process file output
	TemplateFunctions    template.FuncMap                         // TemplateFunctions is a map of additional functions that can be used in the templates
}

type RenderedFile struct {
	File         string
	TemplateFile string
	State        FileState
}

type FileState string

const (
	FileDryRun       FileState = "dry-run"
	FileSkippedName  FileState = "skipped-by-name"
	FileSkippedScope FileState = "skipped-by-scope"
	FileRendered     FileState = "rendered"
)

type Config struct {
	ID          string // ID is a unique identifier for the template, should be a combination of the spec type, generator and template name (openapi-go-client, asyncapi-java-client, etc.)
	Description string // Description is a human-readable description, only used to list available templates
	Files       []File // Files is a list of files that will be rendered
}

func (c Config) FilesByType(t Type) []File {
	var files []File
	for _, f := range c.Files {
		if f.Type == t {
			files = append(files, f)
		}
	}
	return files
}

type File struct {
	Description     string   // Description is a human-readable description of the template
	SourceTemplate  string   // SourceTemplate is the path to the template file
	SourceFile      string   // SourceFile is the path to a file that will be copied as is
	SourceUrl       string   // SourceUrl is the URL where the template or binary file can be downloaded from
	Snippets        []string // Snippets is a list of paths to files that contain snippets that can be used in the template
	TargetDirectory string   // TargetDirectory is the directory where the rendered file will be saved
	TargetFileName  string   // TargetFileName contains the template for the file name
	Type            Type     // Type is the type of the template
	Kind            Kind     // Kind is the kind of the template, can be used to filter which templates to render
	Category        []string // Category is a list of categories that the template belongs to, can be used to filter which templates to render
	// TODO: allow to filter or transform template data per file
}

type Type string

const (
	TypeAPIOnce       Type = "api_once"
	TypeAPIEach       Type = "api_each"
	TypeOperationEach Type = "operation_each"
	TypeModelEach     Type = "model_each"
	TypeEnumEach      Type = "enum_each"
	TypeSupportOnce   Type = "support_once"
)

type Kind string

const (
	KindAPI           Kind = "api"
	KindModel         Kind = "model"
	KindDocumentation Kind = "documentation"
	KindBuildSystem   Kind = "buildsystem"
)
