package openapipatch

import (
	"fmt"
	"log/slog"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/primelib/primecodegen/pkg/llm"
	"github.com/primelib/primecodegen/pkg/util"
	"github.com/speakeasy-api/openapi/overlay"
	"gopkg.in/yaml.v3"
)

func GenerateOpenAPIOverlay(doc *libopenapi.DocumentModel[v3.Document], id string) ([]byte, error) {
	if id == "llm-operation-id-overlay" {
		return LLMOperationIDPatch(doc)
	}

	return nil, fmt.Errorf("unknown patch id %s", id)
}

func LLMOperationIDPatch(doc *libopenapi.DocumentModel[v3.Document]) ([]byte, error) {
	// const
	systemMessage := `
		You are an expert OpenAPI specification assistant.
		Given an HTTP method and URL path, generate a concise and descriptive operation ID following these rules:
		- Use camelCase, no special chars.
		- Map methods: GET→get, POST→create, PUT→update, PATCH→patch, DELETE→delete and singularize resource names.
		- GET on a root resource (no path params) → use list instead of get and pluralize.
		- Use ByXyz for path parameters only.
		- Skip generic prefixes (e.g., "admin", "api").
		- Add version suffix (V1) only if in path.
		- Collapse long nested paths into meaningful names.

		Output only the operation ID.
	`

	// build overlay
	ol := overlay.Overlay{
		Version: "1.0.0",
		Info: overlay.Info{
			Title:   "PrimeCodeGen Patch - [LLM Operation IDs]",
			Version: "1.0.0",
		},
		Actions: make([]overlay.Action, 0),
	}

	// iterate paths and operations to build actions
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		url := path.Key
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			userMessage := fmt.Sprintf("Request: %s %s\nSummary: %s\nDescription: %s", op.Key, url, util.Ellipsize(op.Value.Summary, 100), util.Ellipsize(op.Value.Description, 100))

			suggestedOperationId, err := llm.LLMChatCompletion(systemMessage, userMessage)
			if err != nil {
				slog.Error("failed to generate operation ID using LLM", "method", fmt.Sprintf("%s %s", op.Key, url), "err", err)
				continue
			}
			slog.Info("Operation ID generated", "operation-id", suggestedOperationId, "method", fmt.Sprintf("%s %s", op.Key, url))

			ol.Actions = append(ol.Actions, overlay.Action{
				Target: fmt.Sprintf("$.paths['%s'].%s", url, op.Key),
				Update: yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "operationId",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: suggestedOperationId,
							Tag:   "!!str",
						},
					},
				},
			})
		}
	}

	// render overlay to bytes
	out, err := ol.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
