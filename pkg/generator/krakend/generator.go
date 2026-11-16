package krakend

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
	"github.com/primelib/primecodegen/pkg/template/templateapi"
)

type Generator struct {
	primitiveTypes []string
	typeToImport   map[string]string
}

func (g *Generator) Id() string {
	return "krakend"
}

func (g *Generator) Description() string {
	return "Generates a configuration for KrakenD"
}

func (g *Generator) Generate(opts openapigenerator.GenerateOpts) error {
	// check opts
	if opts.Doc == nil {
		return fmt.Errorf("document is required")
	}

	// print final spec
	if os.Getenv("PRIMECODEGEN_DEBUG_SPEC") == "true" {
		out, _ := opts.Doc.Model.Render()
		fmt.Print(string(out))
	}

	// set packages
	rootPackagePath := strings.ReplaceAll(opts.ArtifactGroupId+"."+opts.ArtifactId, "-", ".")
	opts.PackageConfig = openapigenerator.CommonPackages{
		Root:       rootPackagePath,
		Client:     rootPackagePath + ".client",
		Models:     rootPackagePath + ".models",
		Responses:  rootPackagePath + ".responses",
		Enums:      rootPackagePath + ".enums",
		Operations: rootPackagePath + ".operations",
		Auth:       rootPackagePath + ".auth",
	}

	// build template data
	templateData, err := g.TemplateData(openapigenerator.TemplateDataOpts{
		Doc:           opts.Doc,
		PackageConfig: opts.PackageConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to build template data in %s: %w", g.Id(), err)
	}

	// generator config
	conf, err := parseConfig(opts.TemplateProperties)
	if err != nil {
		return fmt.Errorf("invalid %s generator config: %w", g.Id(), err)
	}

	// resolve backend per environment
	environmentEndpoints := resolveEnvironmentEndpoints(templateData.Endpoints)
	if len(environmentEndpoints) == 0 {
		return fmt.Errorf("no environment endpoint found in OpenAPI servers (expected one of: %s)", strings.Join(environmentOrder, ", "))
	}
	rootExtension, err := rootKrakenDExtension(opts.Doc)
	if err != nil {
		return fmt.Errorf("failed to parse %s extension from root: %w", krakendExtensionKey, err)
	}
	endpointPrefix, _ := rootExtension["endpoint_prefix"].(string)
	rootAuth, _ := toMap(rootExtension["auth"])
	rootClientCredentials, _ := toMap(rootExtension["client_credentials"])
	operationMeta := buildOperationMeta(opts.Doc)

	// write config files
	_ = os.MkdirAll(opts.OutputDir, 0755)

	var generatedEnvironments []string
	for _, environment := range environmentOrder {
		resolvedEndpoint, ok := environmentEndpoints[environment]
		if !ok || !slices.Contains(conf.Environments, environment) {
			continue
		}

		krakendConf, genErr := generateKrakenConfig(templateData, GenerateKrakenConfigOpts{
			Config:            conf,
			Environment:       environment,
			APIName:           templateData.DisplayName,
			BackendHost:       resolvedEndpoint.Host,
			BackendBasePath:   resolvedEndpoint.BasePath,
			EndpointPrefix:    endpointPrefix,
			Auth:              rootAuth,
			ClientCredentials: rootClientCredentials,
			Operations:        operationMeta,
		})
		if genErr != nil {
			return fmt.Errorf("failed to generate %s config: %w", environment, genErr)
		}

		configMap, mapErr := configToMap(krakendConf)
		if mapErr != nil {
			return fmt.Errorf("failed to convert %s config to generic map: %w", environment, mapErr)
		}
		if overrideErr := applyKrakenDOverrides(configMap, templateData, rootExtension); overrideErr != nil {
			return fmt.Errorf("failed to apply %s overrides for %s config: %w", krakendExtensionKey, environment, overrideErr)
		}

		output, marshalErr := json.MarshalIndent(configMap, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal %s config: %w", environment, marshalErr)
		}

		filename := fmt.Sprintf("krakend.%s.json", environment)
		err = filesystem.SaveFileText(path.Join(opts.OutputDir, filename), string(output))
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
		generatedEnvironments = append(generatedEnvironments, environment)
	}

	if len(generatedEnvironments) == 0 {
		return fmt.Errorf("none of the configured environments (%s) has a matching server in the OpenAPI spec", strings.Join(conf.Environments, ", "))
	}
	slog.Info("generated KrakenD configs", "environments", strings.Join(generatedEnvironments, ","), "output-dir", opts.OutputDir)

	return nil
}

func (g *Generator) TemplateData(opts openapigenerator.TemplateDataOpts) (openapigenerator.DocumentModel, error) {
	ensureSpecName(opts.Doc)
	templateData, err := openapigenerator.BuildTemplateData(opts.Doc, g, opts.PackageConfig)
	if err != nil {
		return templateData, err
	}
	templateData = openapigenerator.PruneTypeAliases(templateData, g.primitiveTypes)
	return templateData, nil
}

func (g *Generator) ToClassName(name string) string {
	return name
}

func (g *Generator) ToFunctionName(name string) string {
	return name
}

func (g *Generator) ToPropertyName(name string) string {
	return name
}

func (g *Generator) ToParameterName(name string) string {
	return name
}

func (g *Generator) ToConstantName(name string) string {
	return name
}

func (g *Generator) ToCodeType(schema *base.Schema, schemaType openapigenerator.CodeTypeSchemaType, required bool) (openapigenerator.CodeType, error) {
	return openapigenerator.CodeType{}, nil
}

func (g *Generator) PostProcessType(codeType openapigenerator.CodeType) openapigenerator.CodeType {
	return codeType
}

func (g *Generator) IsPrimitiveType(input string) bool {
	return false
}

func (g *Generator) TypeToImport(iType openapigenerator.CodeType) string {
	typeName := iType.Name

	if typeName == "" {
		return ""
	}

	return g.typeToImport[typeName]
}

func (g *Generator) PostProcessing(files map[string]templateapi.RenderedFile) error {
	return nil
}

func NewGenerator() *Generator {
	return &Generator{}
}
