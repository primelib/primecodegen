package openapigenerator

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/primelib/primecodegen/pkg/template"
	"github.com/rs/zerolog/log"
)

func FilesListedInMetadata(outputDir string) []string {
	writtenFiles := path.Join(outputDir, ".openapi-generator", "FILES")
	log.Debug().Str("output-dir", outputDir).Str("lookup-file", writtenFiles).Msg("Listing generated files")

	file, err := os.Open(writtenFiles)
	if err != nil {
		return nil
	}
	defer file.Close()

	var files []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, path.Join(outputDir, line))
		}
	}

	if err = scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading file metadata")
		return nil
	}

	return files
}

func RemoveGeneratedFile(outputDir string, file string) error {
	if !path.IsAbs(file) {
		file = path.Join(outputDir, file)
	}

	if fileInfo, err := os.Stat(file); err == nil && fileInfo.Mode().IsRegular() {
		remErr := os.Remove(file)
		if remErr != nil {
			return fmt.Errorf("failed to remove file: %w", remErr)
		}
	}

	return nil
}

// WriteMetadata generates metadata about the generated files for the output directory
func WriteMetadata(outputDir string, files map[string]template.RenderedFile) error {
	writtenFiles := path.Join(outputDir, ".openapi-generator", "FILES")

	// ensure output directory exists
	err := os.MkdirAll(path.Dir(writtenFiles), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// open the file for writing
	file, err := os.Create(writtenFiles)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// write each file name to the file
	for _, f := range files {
		if f.State == template.FileRendered {
			relativeFile := strings.TrimPrefix(strings.TrimPrefix(f.File, outputDir), "/")
			_, err = file.WriteString(relativeFile + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}

	return nil
}
