package template

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sync"
	"text/template"

	"github.com/primelib/primecodegen/pkg/template/templateapi"
	"github.com/primelib/primecodegen/pkg/util"
	"log/slog"
)

//go:embed templates/*
var templateFS embed.FS

func RenderTemplateById(templateId string, outputDir string, templateType templateapi.Type, data interface{}, opts templateapi.RenderOpts) (map[string]templateapi.RenderedFile, error) {
	templateConfig, exists := allTemplates[templateId]
	if !exists {
		return nil, errors.Join(templateapi.ErrTemplateNotFound, fmt.Errorf("template id not found: %s", templateId))
	}

	return RenderTemplate(templateConfig, outputDir, templateType, data, opts)
}

// RenderTemplate renders the template with the provided data and returns the rendered files
func RenderTemplate(config templateapi.Config, outputDir string, templateType templateapi.Type, data interface{}, opts templateapi.RenderOpts) (map[string]templateapi.RenderedFile, error) {
	files := make(map[string]templateapi.RenderedFile)
	var filesMutex sync.Mutex
	templateFiles := config.FilesByType(templateType)

	// pre-load all template files
	tmpl := make(map[string]*template.Template)
	for _, file := range config.Files {
		if file.SourceTemplate == "" {
			continue
		}

		t, err := loadTemplate(config.ID, append([]string{file.SourceTemplate}, file.Snippets...), opts.TemplateFunctions)
		if err != nil {
			return nil, errors.Join(templateapi.ErrFailedToParseTemplate, fmt.Errorf("template in %s, file %s: %w", config.ID, file.SourceTemplate, err))
		}
		tmpl[file.SourceTemplate] = t
	}

	// render templates
	var waitGroup sync.WaitGroup
	sem := make(chan struct{}, 6)
	errCh := make(chan error, 1)
	for _, file := range templateFiles {
		file := file
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var renderedContent bytes.Buffer
			if file.SourceTemplate != "" {
				t := tmpl[file.SourceTemplate]

				err := t.Execute(&renderedContent, data)
				if err != nil {
					select {
					case errCh <- errors.Join(templateapi.ErrFailedToRenderTemplate, fmt.Errorf("template in %s, file %s: %w", config.ID, file.SourceTemplate, err)):
					default:
					}
					return
				}
			} else if file.SourceFile != "" {
				content, err := readTemplateFile([]string{config.ID, "_global"}, file.SourceFile)
				if err != nil {
					select {
					case errCh <- errors.Join(templateapi.ErrFailedToCopyTemplateFile, fmt.Errorf("failed to read template file %s: %w", file.SourceFile, err)):
					default:
					}
					return
				}
				renderedContent.Write([]byte(content))
			} else if file.SourceUrl != "" {
				out, err := util.DownloadBytes(file.SourceUrl)
				if err != nil {
					select {
					case errCh <- errors.Join(templateapi.ErrFailedToDownloadTemplateFile, fmt.Errorf("failed to download template from %s: %w", file.SourceUrl, err)):
					default:
					}
					return
				}
				renderedContent.Write(out)
			} else {
				select {
				case errCh <- errors.Join(templateapi.ErrTemplateFileOrUrlIsRequired, errors.New("template id: "+file.TargetDirectory+"/"+file.TargetFileName)):
				default:
				}
				return
			}

			// variables in dir or name
			resolvedDir, err := resolveName(file.TargetDirectory, data)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to resolve directory name: %w", err):
				default:
				}
				return
			}
			resolvedFile, err := resolveName(file.TargetFileName, data)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to resolve file name: %w", err):
				default:
				}
				return
			}

			// write to file
			// TODO: allow variables in target file name
			targetDir := filepath.Join(outputDir, resolvedDir)
			targetFile := filepath.Join(targetDir, resolvedFile)
			skippedByScope := len(opts.Types) > 0 && !slices.Contains(opts.Types, file.Type)
			skippedByName := slices.Contains(opts.IgnoreFiles, file.TargetFileName)
			output := renderedContent.Bytes()
			if opts.PostProcess != nil {
				output = opts.PostProcess(resolvedFile, output)
			}

			var state templateapi.FileState
			if opts.DryRun {
				state = templateapi.FileDryRun
			} else if skippedByName {
				state = templateapi.FileSkippedName
			} else if skippedByScope {
				state = templateapi.FileSkippedScope
			} else {
				err = os.MkdirAll(targetDir, 0755)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("failed to create directory %s: %w", targetDir, err):
					default:
					}
					return
				}

				err = os.WriteFile(targetFile, output, 0644)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("failed to write rendered file %s: %w", targetFile, err):
					default:
					}
					return
				}
				state = templateapi.FileRendered
			}
			slog.Debug("Rendered file", "template-id", config.ID, "file", targetFile)

			filesMutex.Lock()
			files[targetFile] = templateapi.RenderedFile{File: targetFile, TemplateFile: file.SourceTemplate, State: state}
			filesMutex.Unlock()
		}()
	}

	waitGroup.Wait()
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("failed to render template: %w", err)
	default:
	}

	return files, nil
}

func loadTemplate(templateId string, files []string, customFunctions template.FuncMap) (*template.Template, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}
	name := files[0]
	lookupTemplates := []string{templateId, "_global"}

	tmpl := template.New(name)
	tmpl.Funcs(templateapi.TemplateFunctions)
	if customFunctions != nil {
		tmpl.Funcs(customFunctions)
	}
	for _, f := range files {
		err := loadTemplateById(tmpl, lookupTemplates, f)
		if err != nil {
			return nil, err
		}
	}

	if len(tmpl.Templates()) > 0 {
		return tmpl, nil
	}
	return nil, fmt.Errorf("neither embedded filesystem nor PRIMECODEGEN_TEMPLATE_DIR environment variable is set")
}

// loadTemplateById reads a template file from either the local filesystem or the embedded filesystem.
//
// It searches for the file in the given lookupTemplates directories, following the order of priority.
//
// Parameters:
//   - lookupTemplates: A list of template directories to search in, ordered by priority.
//   - templateFile: The name of the file to read.
func loadTemplateById(tmpl *template.Template, lookupTemplates []string, templateFile string) error {
	// read contents of the template file
	content, err := readTemplateFile(lookupTemplates, templateFile)
	if err != nil {
		return err
	}

	// parse the template
	_, err = tmpl.Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", templateFile, err)
	}

	return nil
}

// readTemplateFile reads a template file from either the local filesystem or the embedded filesystem.
//
// It searches for the file in the given lookupTemplates directories, following the order of priority.
//
// Parameters:
//   - lookupTemplates: A list of template directories to search in, ordered by priority.
//   - templateFile: The name of the file to read.
func readTemplateFile(lookupTemplates []string, templateFile string) ([]byte, error) {
	// check local filesystem (PRIMECODEGEN_TEMPLATE_DIR has priority to allow easy customization of templates)
	templateDir := os.Getenv("PRIMECODEGEN_TEMPLATE_DIR")
	if templateDir != "" {
		for _, currentTemplateId := range lookupTemplates {
			file := filepath.Join(templateDir, currentTemplateId, templateFile)
			content, err := os.ReadFile(file)
			if err == nil {
				return content, nil
			}
		}
	}

	// check embedded filesystem
	for _, currentTemplateId := range lookupTemplates {
		embedFSFile := path.Join("templates", currentTemplateId, templateFile)
		content, err := templateFS.ReadFile(embedFSFile)
		if err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("template file %s not found in either embedded filesystem or PRIMECODEGEN_TEMPLATE_DIR", templateFile)
}

// resolveName resolves the file name by executing the template with the provided data
func resolveName(input string, data interface{}) (string, error) {
	tmpl, err := template.New("name").Funcs(templateapi.TemplateFunctions).Parse(input)
	if err != nil {
		return "", err
	}

	var tplOutput bytes.Buffer
	err = tmpl.Execute(&tplOutput, data)
	if err != nil {
		return "", err
	}

	return tplOutput.String(), nil
}
