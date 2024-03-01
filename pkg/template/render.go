package template

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"slices"
)

//go:embed templates/*
var templateFS embed.FS

func RenderTemplateById(templateId string, outputDir string, templateType Type, data interface{}, opts RenderOpts) ([]RenderedFile, error) {
	for _, t := range allTemplates {
		if t.ID == templateId {
			return RenderTemplate(t, outputDir, templateType, data, opts)
		}
	}

	return nil, fmt.Errorf("template with id %s not found", templateId)
}

// RenderTemplate renders the template with the provided data and returns the rendered files
func RenderTemplate(config Config, outputDir string, templateType Type, data interface{}, opts RenderOpts) ([]RenderedFile, error) {
	var files []RenderedFile

	// pre-load all template files
	tmpl := make(map[string]*template.Template)
	for _, file := range config.Files {
		t, err := loadTemplate(config.ID, append([]string{file.SourceTemplate}, file.Snippets...))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template in %s, file %s: %w", config.ID, file.SourceTemplate, err)
		}
		tmpl[file.SourceTemplate] = t
	}

	// render templates
	// TODO: concurrency
	for _, file := range config.Files {
		if file.Type != templateType {
			continue
		}

		// render
		t := tmpl[file.SourceTemplate]

		var renderedContent bytes.Buffer
		err := t.Execute(&renderedContent, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render template in %s, file %s: %w", config.ID, file.SourceTemplate, err)
		}

		// variables in dir or name
		resolvedDir, err := resolveName(file.TargetDirectory, data)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve directory name: %w", err)
		}
		resolvedFile, err := resolveName(file.TargetFileName, data)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve file name: %w", err)
		}

		// write to file
		// TODO: allow variables in target file name
		targetDir := filepath.Join(outputDir, resolvedDir)
		targetFile := filepath.Join(targetDir, resolvedFile)
		skippedByScope := len(opts.Types) > 0 && !slices.Contains(opts.Types, file.Type)
		skippedByName := slices.Contains(opts.IgnoreFiles, file.TargetFileName)

		var state FileState
		if opts.DryRun {
			state = FileDryRun
		} else if skippedByName {
			state = FileSkippedName
		} else if skippedByScope {
			state = FileSkippedScope
		} else {
			err = os.MkdirAll(targetDir, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			err = os.WriteFile(targetFile, renderedContent.Bytes(), 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to write rendered file %s: %w", targetFile, err)
			}
			state = FileRendered
		}
		files = append(files, RenderedFile{File: targetFile, TemplateFile: file.SourceTemplate, State: state})
	}

	return files, nil
}

func loadTemplate(templateId string, files []string) (*template.Template, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}
	name := files[0]
	lookupTemplates := []string{templateId, "_global"}

	tmpl := template.New(name)
	for _, f := range files {
		err := loadTemplateById(tmpl, lookupTemplates, f)
		if err != nil {
			return nil, err
		}
	}

	if len(tmpl.Templates()) > 0 {
		tmpl.Funcs(templateFunctions)
		return tmpl, nil
	}
	return nil, fmt.Errorf("neither embedded filesystem nor PRIMECODEGEN_TEMPLATE_DIR environment variable is set")
}

func loadTemplateById(tmpl *template.Template, lookupTemplates []string, templateFile string) error {
	// local filesystem (PRIMECODEGEN_TEMPLATE_DIR has priority to allow easy customization of templates)
	templateDir := os.Getenv("PRIMECODEGEN_TEMPLATE_DIR")
	if templateDir != "" {
		for _, currentTemplateId := range lookupTemplates {
			file := filepath.Join(templateDir, currentTemplateId, templateFile)
			if _, err := os.Stat(file); err == nil {
				_, err = tmpl.ParseFiles(file)
				if err != nil {
					return fmt.Errorf("failed to parse template file %s: %w", file, err)
				}
				return nil
			}
		}
	}

	// embedded filesystem
	for _, currentTemplateId := range lookupTemplates {
		embedFSFile := path.Join("templates", currentTemplateId, templateFile)
		if _, err := templateFS.ReadFile(embedFSFile); err == nil {
			_, err = tmpl.ParseFS(templateFS, embedFSFile)
			if err != nil {
				return fmt.Errorf("failed to parse embedded template file %s: %w", templateFile, err)
			}
			return nil
		}
	}

	return fmt.Errorf("neither embedded filesystem nor PRIMECODEGEN_TEMPLATE_DIR provides template file %s", templateFile)
}

func resolveName(input string, data interface{}) (string, error) {
	tmpl, err := template.New("name").Parse(input)
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
