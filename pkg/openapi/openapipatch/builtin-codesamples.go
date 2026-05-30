package openapipatch

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/util"
	"go.yaml.in/yaml/v4"
)

var GenerateCodeSamplesRefsPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "generate-code-samples-refs",
	Description:         "Generates x-codeSamples refs from operation slugs",
	PatchV3DocumentFunc: GenerateCodeSamplesRefs,
}

type codeSampleSource struct {
	Ref string `yaml:"$ref"`
}

type codeSample struct {
	Lang   string           `yaml:"lang"`
	Label  string           `yaml:"label,omitempty"`
	Source codeSampleSource `yaml:"source"`
}

func GenerateCodeSamplesRefs(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	baseDir, err := getStringConfig(config, "dir")
	if err != nil {
		return err
	}

	refPrefix, ok := getOptionalStringConfig(config, "ref-prefix")
	if !ok || strings.TrimSpace(refPrefix) == "" {
		refPrefix = "snippets"
	}

	lang, _ := getOptionalStringConfig(config, "language")
	lang = strings.TrimSpace(lang)

	extension, _ := getOptionalStringConfig(config, "extension")
	extension = strings.TrimPrefix(strings.TrimSpace(extension), ".")

	if extension == "" {
		extension = extensionForLanguage(lang)
	}
	if extension == "" {
		return fmt.Errorf("missing extension: set config.language or config.extension")
	}

	if lang == "" {
		lang = inferLanguage(extension)
		if lang == "" {
			lang = extension
		}
	}

	if doc.Model.Paths == nil {
		return nil
	}

	label := buildCodeSampleLabel(doc, lang)

	for p := doc.Model.Paths.PathItems.Oldest(); p != nil; p = p.Next() {
		for op := p.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			slug := util.OpenAPIOperationSlug(op.Key, p.Key)
			fileName := slug + "." + extension
			filePath := buildCodeSampleFilePath(baseDir, refPrefix, fileName)
			if _, statErr := os.Stat(filePath); statErr != nil {
				slog.Warn("code sample file missing or inaccessible, skipping", "file", filePath, "path", p.Key, "method", strings.ToUpper(op.Key), "err", statErr)
				continue
			}

			ref := buildCodeSampleRef(baseDir, refPrefix, fileName)

			if op.Value.Extensions == nil {
				op.Value.Extensions = orderedmap.New[string, *yaml.Node]()
			}

			existingSamples := make([]codeSample, 0)
			if existingNode, exists := op.Value.Extensions.Get("x-codeSamples"); exists && existingNode != nil {
				if err = existingNode.Decode(&existingSamples); err != nil {
					return fmt.Errorf("failed to decode existing x-codeSamples for %s %s: %w", strings.ToUpper(op.Key), p.Key, err)
				}
			}

			alreadyPresent := false
			for _, sample := range existingSamples {
				if sample.Source.Ref == ref {
					alreadyPresent = true
					break
				}
			}

			if !alreadyPresent {
				existingSamples = append(existingSamples, codeSample{
					Lang:  lang,
					Label: label,
					Source: codeSampleSource{
						Ref: ref,
					},
				})
			}

			node, nodeErr := marshalCodeSamplesNode(existingSamples)
			if nodeErr != nil {
				return fmt.Errorf("failed to render x-codeSamples for operation %s %s: %w", strings.ToUpper(op.Key), p.Key, nodeErr)
			}
			op.Value.Extensions.Set("x-codeSamples", node)
		}
	}

	return nil
}

func buildCodeSampleRef(baseDir string, refPrefix string, fileName string) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(baseDir) != "" {
		parts = append(parts, strings.Trim(strings.TrimSpace(baseDir), "/"))
	}
	if strings.TrimSpace(refPrefix) != "" {
		parts = append(parts, strings.Trim(strings.TrimSpace(refPrefix), "/"))
	}
	parts = append(parts, strings.Trim(strings.TrimSpace(fileName), "/"))
	return path.Clean(path.Join(parts...))
}

func buildCodeSampleFilePath(baseDir string, refPrefix string, fileName string) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(baseDir) != "" {
		parts = append(parts, strings.TrimSpace(baseDir))
	}
	if strings.TrimSpace(refPrefix) != "" {
		parts = append(parts, strings.Trim(strings.TrimSpace(refPrefix), "/"))
	}
	parts = append(parts, strings.Trim(strings.TrimSpace(fileName), "/"))
	return filepath.Clean(filepath.Join(parts...))
}

func marshalCodeSamplesNode(samples []codeSample) (*yaml.Node, error) {
	bytes, err := yaml.Marshal(samples)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	if err = yaml.Unmarshal(bytes, &node); err != nil {
		return nil, err
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0], nil
	}

	return &node, nil
}

func inferLanguage(extension string) string {
	switch strings.ToLower(extension) {
	case "java":
		return "java"
	case "kt", "kts":
		return "kotlin"
	case "go":
		return "go"
	case "ts":
		return "typescript"
	case "js", "mjs", "cjs":
		return "javascript"
	case "py":
		return "python"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "cs":
		return "csharp"
	case "swift":
		return "swift"
	case "sh":
		return "bash"
	case "curl":
		return "curl"
	case "http":
		return "http"
	default:
		return ""
	}
}

func extensionForLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "java":
		return "java"
	case "kotlin":
		return "kt"
	case "go":
		return "go"
	case "typescript":
		return "ts"
	case "javascript":
		return "js"
	case "python":
		return "py"
	case "ruby":
		return "rb"
	case "php":
		return "php"
	case "csharp":
		return "cs"
	case "swift":
		return "swift"
	case "bash":
		return "sh"
	case "curl":
		return "curl"
	case "http":
		return "http"
	default:
		return ""
	}
}

func buildCodeSampleLabel(doc *libopenapi.DocumentModel[v3.Document], lang string) string {
	apiName := "API"
	if doc != nil && doc.Model.Info != nil {
		if doc.Model.Info.Extensions != nil {
			if xName, ok := doc.Model.Info.Extensions.Get("x-name"); ok && xName != nil && strings.TrimSpace(xName.Value) != "" {
				apiName = strings.TrimSpace(xName.Value)
			} else if strings.TrimSpace(doc.Model.Info.Title) != "" {
				apiName = strings.TrimSpace(doc.Model.Info.Title)
			}
		} else if strings.TrimSpace(doc.Model.Info.Title) != "" {
			apiName = strings.TrimSpace(doc.Model.Info.Title)
		}
	}

	if apiName == "" {
		apiName = "API"
	}

	language := displayLanguage(lang)
	return fmt.Sprintf("%s %s SDK", apiName, language)
}

func displayLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "java":
		return "Java"
	case "kotlin":
		return "Kotlin"
	case "go":
		return "Go"
	case "typescript":
		return "TypeScript"
	case "javascript":
		return "JavaScript"
	case "python":
		return "Python"
	case "ruby":
		return "Ruby"
	case "php":
		return "PHP"
	case "csharp":
		return "C#"
	case "swift":
		return "Swift"
	case "bash":
		return "Bash"
	case "curl":
		return "cURL"
	case "http":
		return "HTTP"
	default:
		return strings.ToUpper(strings.TrimSpace(language))
	}
}
