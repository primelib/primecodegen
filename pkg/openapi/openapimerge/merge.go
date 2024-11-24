package openapimerge

import (
	"fmt"
	"github.com/cidverse/cidverseutils/filesystem"
	"os"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/rs/zerolog/log"

	str "strings"
)

// MergeOpenAPISpecs merges multiple OpenAPI specs into a single OpenAPI spec
func MergeOpenAPISpecs(emptySpec string, paths []string) (*libopenapi.DocumentModel[v3.Document], error) {
	var mergedSpec *libopenapi.DocumentModel[v3.Document]

	for _, path := range paths {
		// Load OpenAPI spec file
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}
		// Parse OpenAPI spec
		doc, err := openapidocument.OpenDocument(bytes)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open document")
		}
		v3Model, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			log.Fatal().Errs("spec", errs).Msgf("failed to build v3 high level model")
		}
		// Initialize mergedSpec if it is nil
		if mergedSpec == nil {
			if emptySpec != "" && filesystem.FileExists(emptySpec) {
				log.Trace().Str("filepath", emptySpec).Msg("Creating empty doc model from empty spec in")
				// Empty OpenAPI spec from file (for clean info-block build-up)
				_, mergedSpec, _ = CreateEmptySpec(emptySpec)
			} else {
				mergedSpec = v3Model
			}
		}
		// Merge OpenAPI elements
		mergeInfo(&mergedSpec.Model, &v3Model.Model)
		mergeServers(&mergedSpec.Model, &v3Model.Model)
		mergeTags(&mergedSpec.Model, &v3Model.Model)
		mergePaths(&mergedSpec.Model, &v3Model.Model)
		mergeComponents(&mergedSpec.Model, &v3Model.Model)

		// Reload document
		_, doc, _, errs = doc.RenderAndReload()
		if len(errs) > 0 {
			log.Error().Errs("spec", errs).Msgf("failed to reload document after patching")
		}
		v3Model, errs = doc.BuildV3Model()
		if len(errs) > 0 {
			log.Error().Errs("spec", errs).Msgf("failed to build v3 high level model")
		}
	}

	return mergedSpec, nil
}

func mergeInfo(dest, src *v3.Document) {
	if src.Info.Title != "" {
		if dest.Info.Title != "" {
			dest.Info.Title = dest.Info.Title + ", " + src.Info.Title
		} else {
			dest.Info.Title = src.Info.Title
		}
	}
	if src.Info.Version != "" {
		if dest.Info.Version != "" {
			dest.Info.Version = dest.Info.Version + "\n" + src.Info.Version + " (" + src.Info.Title + ")"
		} else {
			dest.Info.Version = src.Info.Version + " (" + src.Info.Title + ")"
		}
	}
	if src.Info.Summary != "" {
		if dest.Info.Summary != "" {
			dest.Info.Summary = dest.Info.Summary + "\n\n" + str.ToUpper(src.Info.Title) + ": " + src.Info.Summary
		} else {
			dest.Info.Summary = str.ToUpper(src.Info.Title) + ": " + src.Info.Summary
		}
	}
	if src.Info.Description != "" {
		if dest.Info.Description != "" {
			dest.Info.Description = dest.Info.Description + "\n\n" + str.ToUpper(src.Info.Title) + " \n\n" + src.Info.Description
		} else {
			dest.Info.Description = str.ToUpper(src.Info.Title) + " \n\n" + src.Info.Description
		}
	}
	if src.Info.TermsOfService != "" {
		if dest.Info.TermsOfService != "" {
			dest.Info.TermsOfService = dest.Info.TermsOfService + "\n\n" + str.ToUpper(src.Info.Title) + " \n\n" + src.Info.TermsOfService
		} else {
			dest.Info.TermsOfService = str.ToUpper(src.Info.Title) + " \n\n" + src.Info.TermsOfService
		}
	}
	if src.Info.Contact != nil {
		if dest.Info.Contact != nil {
			dest.Info.Contact.Name = dest.Info.Contact.Name + "\n" + str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.Name
			dest.Info.Contact.Email = dest.Info.Contact.Email + "\n" + str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.Email
			dest.Info.Contact.URL = dest.Info.Contact.URL + "\n" + str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.URL
		} else {
			dest.Info.Contact = &base.Contact{
				Name:  str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.Name,
				Email: str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.Email,
				URL:   str.ToUpper(src.Info.Title) + ": " + src.Info.Contact.URL,
			}
		}
	}
	if src.Info.License != nil {
		if dest.Info.License != nil {
			dest.Info.License.Name = dest.Info.License.Name + "\n" + src.Info.Title + ": " + src.Info.License.Name
			dest.Info.License.URL = dest.Info.License.URL + "\n" + src.Info.Title + ": " + src.Info.License.URL
			dest.Info.License.Identifier = dest.Info.License.Identifier + "\n" + src.Info.Title + ": " + src.Info.License.Identifier
		} else {
			dest.Info.License = &base.License{
				Name:       src.Info.Title + ": " + src.Info.License.Name,
				URL:        src.Info.Title + ": " + src.Info.License.URL,
				Identifier: src.Info.Title + ": " + src.Info.License.Identifier,
			}
		}
	}
}

func mergeServers(dest, src *v3.Document) {
	dest.Servers = append(dest.Servers, src.Servers...)
}

func mergeTags(dest, src *v3.Document) {
	dest.Tags = append(dest.Tags, src.Tags...)
}

func mergePaths(dest, src *v3.Document) {
	if src.Paths != nil {
		if dest.Paths == nil {
			dest.Paths = src.Paths
			return
		} else {
			for pathairs := src.Paths.PathItems.First(); pathairs != nil; pathairs = pathairs.Next() {
				pathname := pathairs.Key()
				pathitem := pathairs.Value()
				if _, present := dest.Paths.PathItems.Get(pathname); !present {
					dest.Paths.PathItems.Set(pathname, pathitem)
				} else {
					log.Error().Str("mergePaths: Path Item already exists: ", pathname)
					// TODO: Handle duplicate (rename|prefix)
				}
			}
			return
		}
	}
}

func mergeComponents(dest, src *v3.Document) {
	if src.Components == nil {
		return
	}
	if dest.Components != nil {
		// Schema
		for schema := src.Components.Schemas.First(); schema != nil; schema = schema.Next() {
			schemaname := schema.Key()
			schemavalue := schema.Value()
			if _, present := dest.Components.Schemas.Get(schemaname); !present {
				dest.Components.Schemas.Set(schemaname, schemavalue)
			} else {
				log.Error().Str("Schema already exists: ", schemaname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Security Schema
		for securityschema := src.Components.SecuritySchemes.First(); securityschema != nil; securityschema = securityschema.Next() {
			securityschemaname := securityschema.Key()
			securityschemavalue := securityschema.Value()
			if _, present := dest.Components.SecuritySchemes.Get(securityschemaname); !present {
				dest.Components.SecuritySchemes.Set(securityschemaname, securityschemavalue)
			} else {
				log.Error().Str("Security Schema already exists: ", securityschemaname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Responses
		for response := src.Components.Responses.First(); response != nil; response = response.Next() {
			responsename := response.Key()
			responsevalue := response.Value()
			if _, present := dest.Components.Responses.Get(responsename); !present {
				dest.Components.Responses.Set(responsename, responsevalue)
			} else {
				log.Error().Str("Response already exists: ", responsename)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Parameters
		for parameter := src.Components.Parameters.First(); parameter != nil; parameter = parameter.Next() {
			responsename := parameter.Key()
			responsevalue := parameter.Value()
			if _, present := dest.Components.Parameters.Get(responsename); !present {
				dest.Components.Parameters.Set(responsename, responsevalue)
			} else {
				log.Error().Str("Parameter already exists: ", responsename)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Examples
		for example := src.Components.Examples.First(); example != nil; example = example.Next() {
			examplename := example.Key()
			examplevalue := example.Value()
			if _, present := dest.Components.Examples.Get(examplename); !present {
				dest.Components.Examples.Set(examplename, examplevalue)
			} else {
				log.Error().Str("Example already exists: ", examplename)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Request Bodies
		for requestbody := src.Components.RequestBodies.First(); requestbody != nil; requestbody = requestbody.Next() {
			requestbodyname := requestbody.Key()
			requestbodyvalue := requestbody.Value()
			if _, present := dest.Components.RequestBodies.Get(requestbodyname); !present {
				dest.Components.RequestBodies.Set(requestbodyname, requestbodyvalue)
			} else {
				log.Error().Str("Request Body already exists: ", requestbodyname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Headers
		for header := src.Components.Headers.First(); header != nil; header = header.Next() {
			headername := header.Key()
			headervalue := header.Value()
			if _, present := dest.Components.Headers.Get(headername); !present {
				dest.Components.Headers.Set(headername, headervalue)
			} else {
				log.Error().Str("Header already exists: ", headername)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Links
		for link := src.Components.Links.First(); link != nil; link = link.Next() {
			linkname := link.Key()
			linkvalue := link.Value()
			if _, present := dest.Components.Links.Get(linkname); !present {
				dest.Components.Links.Set(linkname, linkvalue)
			} else {
				log.Error().Str("Link already exists: ", linkname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Callbacks
		for callback := src.Components.Callbacks.First(); callback != nil; callback = callback.Next() {
			callbackname := callback.Key()
			callbackvalue := callback.Value()
			if _, present := dest.Components.Callbacks.Get(callbackname); !present {
				dest.Components.Callbacks.Set(callbackname, callbackvalue)
			} else {
				log.Error().Str("Callback already exists: ", callbackname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
		// Path Items
		for pathitem := src.Components.PathItems.First(); pathitem != nil; pathitem = pathitem.Next() {
			pathitemname := pathitem.Key()
			pathitemvalue := pathitem.Value()
			if _, present := dest.Components.PathItems.Get(pathitemname); !present {
				dest.Components.PathItems.Set(pathitemname, pathitemvalue)
			} else {
				log.Error().Str("Path Item already exists: ", pathitemname)
				// TODO: Handle duplicate (rename|prefix)
			}
		}
	} else {
		dest.Components = src.Components
		return
	}
}

// Create an empty OpenAPI spec to be filled with specs to be merged
func CreateEmptySpec(path string) (libopenapi.Document, *libopenapi.DocumentModel[v3.Document], error) {
	_, error := os.Stat(path)
	if os.IsNotExist(error) {
		return nil, nil, fmt.Errorf("file does not exist %s", path)
	}
	// Load the OpenAPI spec file
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	// Parse the OpenAPI spec
	doc, err := openapidocument.OpenDocument(bytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open document")
	}
	v3Model, errors := doc.BuildV3Model()
	// if anything went wrong when building the v3 model, a slice of errors will be returned
	if len(errors) > 0 {
		for i := range errors {
			fmt.Printf("error: %e\n", errors[i])
		}
		panic(fmt.Sprintf("cannot create v3 model from document: %d errors reported", len(errors)))
	}

	return doc, v3Model, nil
}
