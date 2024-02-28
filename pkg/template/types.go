package template

type RenderOpts struct {
	DryRun      bool     // DryRun will not write any files to disk
	Scopes      []Scope  // Scopes is a list of scopes that should be rendered (empty means all)
	IgnoreFiles []string // IgnoreFiles is a list of file names that should not be rendered
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

type File struct {
	Description     string   // Description is a human-readable description of the template
	SourceTemplate  string   // SourceTemplate is the path to the template file
	Snippets        []string // Snippets is a list of paths to files that contain snippets that can be used in the template
	TargetDirectory string   // TargetDirectory is the directory where the rendered file will be saved
	TargetFileName  string   // TargetFileName contains the template for the file name
	Scope           Scope    // Scope can be used as filter to only render files with a given scope
	Iterator        Iterator // Iterator can be used to render the same template multiple times with different data
	// TODO: allow to filter or transform template data per file
}

type Scope string

const (
	ScopeAPI       Scope = "api"
	ScopeOperation Scope = "operation"
	ScopeModel     Scope = "model"
	ScopeSupport   Scope = "support"
)

type Iterator string

const (
	IteratorOnce          Iterator = "once"
	IteratorEachAPI       Iterator = "each_api"
	IteratorEachOperation Iterator = "each_operation"
	IteratorEachModel     Iterator = "each_model"
)
