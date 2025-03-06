package openapimerge

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/rs/zerolog/log"
)

var (
	ErrOpenAPICrossVersionMergeUnsupported = fmt.Errorf("cross-version merge for openapi specs is not supported")
)

// MergeOpenAPI3Files merges multiple OpenAPI spec files into a single OpenAPI document
func MergeOpenAPI3Files(paths []string) (*libopenapi.DocumentModel[v3.Document], error) {
	var specs [][]byte

	for _, path := range paths {
		spec, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}
		specs = append(specs, spec)
	}

	return MergeOpenAPI3(specs)
}

// MergeOpenAPI3 merges multiple OpenAPI specs into a single OpenAPI spec
func MergeOpenAPI3(specs [][]byte) (*libopenapi.DocumentModel[v3.Document], error) {
	var mergedSpec = ptr.Ptr(openapidocument.EmptyDocument())
	specVersion := ""

	if len(specs) == 1 {
		// open document
		doc, err := openapidocument.OpenDocument(specs[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open document")
		}

		// build v3 model
		v3Model, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			return mergedSpec, errors.Join(util.ErrGenerateOpenAPIV3Model, errors.Join(errs...))
		}

		mergedSpec = v3Model
	} else if len(specs) > 1 {
		for _, spec := range specs {
			// open document
			doc, err := openapidocument.OpenDocument(spec)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to open document")
			}

			if specVersion == "" {
				specVersion = doc.GetVersion()
			} else if specVersion != doc.GetVersion() {
				return mergedSpec, errors.Join(ErrOpenAPICrossVersionMergeUnsupported, fmt.Errorf("spec version mismatch: %s != %s", specVersion, doc.GetVersion()))
			}

			// build v3 model
			v3Model, errs := doc.BuildV3Model()
			if len(errs) > 0 {
				return mergedSpec, errors.Join(util.ErrGenerateOpenAPIV3Model, errors.Join(errs...))
			}

			// merge elements
			mergeInfo(&mergedSpec.Model, &v3Model.Model)
			mergeServers(&mergedSpec.Model, &v3Model.Model)
			mergeTags(&mergedSpec.Model, &v3Model.Model)
			mergePaths(&mergedSpec.Model, &v3Model.Model)
			mergeComponents(&mergedSpec.Model, &v3Model.Model)

			// reload document
			_, doc, _, errs = doc.RenderAndReload()
			if len(errs) > 0 {
				return mergedSpec, errors.Join(util.ErrRenderDocument, errors.Join(errs...))
			}
			v3Model, errs = doc.BuildV3Model()
			if len(errs) > 0 {
				return mergedSpec, errors.Join(util.ErrGenerateOpenAPIV3Model, errors.Join(errs...))
			}
		}
	}

	return mergedSpec, nil
}

func mergeInfo(dest, src *v3.Document) {
	if src.Info == nil {
		return
	}
	if dest.Info == nil {
		dest.Info = &base.Info{}
	}

	titleUpper := strings.ToUpper(src.Info.Title)
	util.AppendOrSetString(&dest.Info.Title, src.Info.Title, "", ", ")
	util.AppendOrSetString(&dest.Info.Version, src.Info.Version, "("+src.Info.Title+") ", "\n")
	util.AppendOrSetString(&dest.Info.Summary, src.Info.Summary, titleUpper+": ", "\n\n")
	util.AppendOrSetString(&dest.Info.Description, src.Info.Description, "# "+titleUpper+"\n\n", "\n\n")
	util.AppendOrSetString(&dest.Info.TermsOfService, src.Info.TermsOfService, titleUpper+"\n\n", "\n\n")

	if src.Info.Contact != nil {
		if dest.Info.Contact == nil {
			dest.Info.Contact = &base.Contact{}
		}
		util.AppendOrSetString(&dest.Info.Contact.Name, src.Info.Contact.Name, titleUpper+": ", "\n")
		util.AppendOrSetString(&dest.Info.Contact.Email, src.Info.Contact.Email, titleUpper+": ", "\n")
		util.AppendOrSetString(&dest.Info.Contact.URL, src.Info.Contact.URL, titleUpper+": ", "\n")
	}

	if src.Info.License != nil {
		if dest.Info.License == nil {
			dest.Info.License = &base.License{}
		}
		util.AppendOrSetString(&dest.Info.License.Name, src.Info.License.Name, src.Info.Title+": ", "\n")
		util.AppendOrSetString(&dest.Info.License.URL, src.Info.License.URL, src.Info.Title+": ", "\n")
		util.AppendOrSetString(&dest.Info.License.Identifier, src.Info.License.Identifier, src.Info.Title+": ", "\n")
	}
}

func mergeServers(dest, src *v3.Document) {
	dest.Servers = append(dest.Servers, src.Servers...)
}

func mergeTags(dest, src *v3.Document) {
	dest.Tags = append(dest.Tags, src.Tags...)
}

func mergePaths(dest, src *v3.Document) {
	if src.Paths == nil {
		return
	}
	if dest.Paths == nil {
		dest.Paths = src.Paths
		return
	}

	for pathItem := src.Paths.PathItems.First(); pathItem != nil; pathItem = pathItem.Next() {
		pathName, pathValue := pathItem.Key(), pathItem.Value()

		if _, exists := dest.Paths.PathItems.Get(pathName); !exists {
			dest.Paths.PathItems.Set(pathName, pathValue)
		} else {
			log.Error().Str("path", pathName).Msg("mergePaths: Path Item already exists")
			// TODO: Handle duplicate (rename | prefix)
		}
	}
}

func mergeComponents(dest, src *v3.Document) {
	if src.Components == nil {
		return
	}
	if dest.Components == nil {
		dest.Components = src.Components
		return
	}

	// Merge all component types
	util.MergeComponentMap(dest.Components.Schemas, src.Components.Schemas, "Schema")
	util.MergeComponentMap(dest.Components.SecuritySchemes, src.Components.SecuritySchemes, "Security Schema")
	util.MergeComponentMap(dest.Components.Responses, src.Components.Responses, "Response")
	util.MergeComponentMap(dest.Components.Parameters, src.Components.Parameters, "Parameter")
	util.MergeComponentMap(dest.Components.Examples, src.Components.Examples, "Example")
	util.MergeComponentMap(dest.Components.RequestBodies, src.Components.RequestBodies, "Request Body")
	util.MergeComponentMap(dest.Components.Headers, src.Components.Headers, "Header")
	util.MergeComponentMap(dest.Components.Links, src.Components.Links, "Link")
	util.MergeComponentMap(dest.Components.Callbacks, src.Components.Callbacks, "Callback")
	util.MergeComponentMap(dest.Components.PathItems, src.Components.PathItems, "Path Item")
}
