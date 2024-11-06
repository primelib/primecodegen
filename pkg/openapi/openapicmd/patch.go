package openapicmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/primelib/primecodegen/pkg/commonmerge"
	"github.com/primelib/primecodegen/pkg/commonpatch"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/openapi/openapipatch"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func PatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "openapi-patch",
		Aliases: []string{},
		GroupID: "openapi",
		Short:   "Patch OpenAPI Specification for Code Generation",
		Long:    "Transform an OpenAPI Specification to be compatible with code generation tools and clean up common issues",
		Run: func(cmd *cobra.Command, args []string) {
			// inputs
			inputFiles, _ := cmd.Flags().GetStringSlice("input")
			if len(inputFiles) == 0 {
				log.Fatal().Msg("input specification is required")
			}
			out, _ := cmd.Flags().GetString("output")
			patches, _ := cmd.Flags().GetStringSlice("patch")
			patchFiles, _ := cmd.Flags().GetStringSlice("patch-file")

			// run patch command
			log.Info().Strs("input", inputFiles).Strs("patch-ids", patches).Str("output-file", out).Msg("patching")
			stdout, err := runPatchCmd(inputFiles, out, patches, patchFiles)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to patch document")
			} else if stdout != "" {
				fmt.Printf("%s", stdout)
			}
		},
	}

	cmd.Flags().StringSliceP("input", "i", []string{}, "Input Specification(s) (YAML or JSON)")
	cmd.Flags().StringP("output", "o", "", "Output File")
	cmd.Flags().StringSliceP("patch", "p", []string{"generateOperationIds", "missingSchemaTitle"}, "Patch IDs to apply")
	cmd.Flags().StringSliceP("patch-file", "f", []string{}, "Patch files to apply (.patch, .jsonpatch)")
	cmd.AddCommand(PatchListCmd())

	return cmd
}

func runPatchCmd(inputFiles []string, output string, patches []string, patchFiles []string) (string, error) {
	for i, v := range inputFiles {
		inputFiles[i] = util.ResolvePath(v)
	}
	output = util.ResolvePath(output)

	// Check for pre-merge patches and apply them
	_, patches, inputFiles, err := applyPreMergePatches(inputFiles, output, patches)
	if err != nil {
		return "", errors.Join(util.ErrDocumentMerge, err)
	}
	if len(patches) == 0 {
		return "", nil
	}
	// read and merge documents
	bytes, err := commonmerge.ReadAndMergeFiles(inputFiles)
	if err != nil {
		return "", errors.Join(util.ErrDocumentMerge, err)
	}

	// patch document (external files)
	for _, patchFile := range patchFiles {
		patchedBytes, patchErr := commonpatch.ApplyPatchFile(bytes, patchFile)
		if patchErr != nil {
			return "", errors.Join(util.ErrFailedToPatchDocument, patchErr)
		}

		bytes = patchedBytes
	}

	// parse document
	doc, err := openapidocument.OpenDocument(bytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open document")
	}
	v3doc, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
	}

	// patch document (built-in)
	doc, v3doc, err = openapipatch.PatchV3(patches, doc, v3doc)
	if err != nil {
		return "", errors.Join(util.ErrFailedToPatchDocument, err)
	}

	// write document
	if output != "" {
		writeErr := openapidocument.RenderDocumentFile(doc, output)
		if writeErr != nil {
			return "", errors.Join(util.ErrWriteDocumentToFile, writeErr)
		}
	} else {
		outBytes, outErr := openapidocument.RenderDocument(doc)
		if outErr != nil {
			return "", errors.Join(util.ErrWriteDocumentToStdout, outErr)
		}
		return string(outBytes), nil
	}

	return "", nil
}

func applyPreMergePatches(inputFiles []string, output string, patches []string) (string, []string, []string, error) {

	var resultFiles []string

	for _, v := range patches {
		if v == "createOperationTagsFromDocTitle" {
			var err error
			_, patches, resultFiles, err = createOperationTagsFromDocTitle(inputFiles, output, patches, v)
			if err != nil {
				return "", []string{}, []string{}, fmt.Errorf("failed to create operation tags: %w", err)
			}
			inputFiles = resultFiles
		}
	}
	return "", patches, inputFiles, nil
}

// For each spec replace all operation tags with a tag generated from the spec title
func createOperationTagsFromDocTitle(inputFiles []string, output string, patches []string, patch string) (string, []string, []string, error) {

	var resultFiles []string

	for _, path := range inputFiles {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return "", []string{}, []string{}, fmt.Errorf("failed to read file %s: %w", path, err)
		}
		doc, err := openapidocument.OpenDocument(bytes)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open document")
		}
		v3docmodel, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
		}
		doc, v3docmodel, err = openapipatch.PatchV3([]string{patch}, doc, v3docmodel)
		if err != nil {
			return "", []string{}, []string{}, errors.Join(util.ErrFailedToPatchDocument, err)
		}
		// write document
		if output != "" {
			filebasename := filepath.Base(path)
			filePath := filepath.Join(output, filebasename)
			writeErr := openapidocument.RenderDocumentFile(doc, filePath)
			if writeErr != nil {
				return "", []string{}, []string{}, errors.Join(util.ErrWriteDocumentToFile, writeErr)
			}
			resultFiles = append(resultFiles, filePath)
		} else {
			outBytes, outErr := openapidocument.RenderDocument(doc)
			if outErr != nil {
				return "", []string{}, []string{}, errors.Join(util.ErrWriteDocumentToStdout, outErr)
			}
			return string(outBytes), []string{}, []string{}, nil
		}
	}
	// Delete this patch from list
	for i, v := range patches {
		if v == "createOperationTagsFromDocTitle" {
			patches = append(patches[:i], patches[i+1:]...)
		}
	}
	if len(patches) == 0 {
		return "", []string{}, resultFiles, nil
	}

	return "", patches, resultFiles, nil
}
