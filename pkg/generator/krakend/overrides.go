package krakend

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/openapi/openapigenerator"
)

const krakendExtensionKey = "x-krakend"

func rootKrakenDExtension(doc *libopenapi.DocumentModel[v3.Document]) (map[string]interface{}, error) {
	if doc == nil || doc.Model.Extensions == nil {
		return map[string]interface{}{}, nil
	}

	return extensionMap(doc.Model.Extensions, krakendExtensionKey)
}

func operationKrakenDExtension(op openapigenerator.Operation) (map[string]interface{}, error) {
	if op.Extensions == nil {
		return map[string]interface{}{}, nil
	}

	return extensionMap(op.Extensions, krakendExtensionKey)
}

func extensionMap(extensions *orderedmap.Map[string, *yaml.Node], key string) (map[string]interface{}, error) {
	if extensions == nil {
		return map[string]interface{}{}, nil
	}

	node, ok := extensions.Get(key)
	if !ok || node == nil {
		return map[string]interface{}{}, nil
	}

	decoded := map[string]interface{}{}
	if err := node.Decode(&decoded); err == nil {
		return normalizeMap(decoded), nil
	}

	var generic interface{}
	if err := node.Decode(&generic); err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", key, err)
	}
	normalized := normalizeValue(generic)
	mapValue, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s must be an object", key)
	}

	return mapValue, nil
}

// applyKrakenDOverrides merges the x-krakend overrides into the generated config.
// Precedence (lowest to highest): generated values < root endpoint_defaults/backend_defaults < typed operation keys and namespace shortcuts < raw operation endpoint/backend overrides.
func applyKrakenDOverrides(config map[string]interface{}, data openapigenerator.DocumentModel, rootExtension map[string]interface{}) error {
	if serviceExtraConfig := namespaceShortcuts(rootExtension, serviceNamespaceShortcuts); len(serviceExtraConfig) > 0 {
		mergeExtraConfig(config, serviceExtraConfig)
	}
	if service, ok := toMap(rootExtension["service"]); ok {
		deepMergeMap(config, service)
	}
	endpointDefaults, _ := toMap(rootExtension["endpoint_defaults"])
	backendDefaults, _ := toMap(rootExtension["backend_defaults"])

	exposed, err := exposedOperations(data.Operations)
	if err != nil {
		return err
	}

	endpointsRaw, ok := config["endpoints"].([]interface{})
	if !ok {
		return nil
	}
	if len(endpointsRaw) != len(exposed) {
		return fmt.Errorf("generated %d endpoints for %d operations, %s.service must not override endpoints", len(endpointsRaw), len(exposed), krakendExtensionKey)
	}

	for i, endpointRaw := range endpointsRaw {
		endpointMap, ok := toMap(endpointRaw)
		if !ok {
			continue
		}
		op, settings := exposed[i].Operation, exposed[i].Settings

		opExtension, err := operationKrakenDExtension(op)
		if err != nil {
			return err
		}

		deepMergeMap(endpointMap, endpointDefaults)
		deepMergeMap(endpointMap, typedEndpointOverrides(settings))
		mergeExtraConfig(endpointMap, namespaceShortcuts(opExtension, endpointNamespaceShortcuts))
		if endpointOverride, ok := toMap(opExtension["endpoint"]); ok {
			deepMergeMap(endpointMap, endpointOverride)
		}
		if endpointExtraConfig, ok := toMap(opExtension["endpoint_extra_config"]); ok {
			mergeExtraConfig(endpointMap, endpointExtraConfig)
		}

		backendsRaw, _ := endpointMap["backend"].([]interface{})
		if len(backendsRaw) > 0 {
			if backendMap, ok := toMap(backendsRaw[0]); ok {
				deepMergeMap(backendMap, backendDefaults)
				if settings.BackendEncoding != "" {
					backendMap["encoding"] = settings.BackendEncoding
				}
				mergeExtraConfig(backendMap, namespaceShortcuts(opExtension, backendNamespaceShortcuts))
				if backendOverride, ok := toMap(opExtension["backend"]); ok {
					deepMergeMap(backendMap, backendOverride)
				}
				if backendExtraConfig, ok := toMap(opExtension["backend_extra_config"]); ok {
					mergeExtraConfig(backendMap, backendExtraConfig)
				}
				backendsRaw[0] = backendMap
			}
			endpointMap["backend"] = backendsRaw
		}

		endpointsRaw[i] = endpointMap
	}

	config["endpoints"] = endpointsRaw
	return nil
}

// typedEndpointOverrides returns the scalar endpoint settings of the operation, re-applied after the root defaults so they take precedence
func typedEndpointOverrides(settings operationSettings) map[string]interface{} {
	out := map[string]interface{}{}
	if settings.Timeout != "" {
		out["timeout"] = settings.Timeout
	}
	if settings.CacheTTL != "" {
		out["cache_ttl"] = settings.CacheTTL
	}
	if settings.ConcurrentCalls > 0 {
		out["concurrent_calls"] = settings.ConcurrentCalls
	}
	if settings.OutputEncoding != "" {
		out["output_encoding"] = settings.OutputEncoding
	}
	return out
}

func mergeExtraConfig(target map[string]interface{}, extra map[string]interface{}) {
	if len(extra) == 0 {
		return
	}
	currentExtraConfig, _ := toMap(target["extra_config"])
	if currentExtraConfig == nil {
		currentExtraConfig = map[string]interface{}{}
	}
	deepMergeMap(currentExtraConfig, extra)
	target["extra_config"] = currentExtraConfig
}

func operationKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func toMap(value interface{}) (map[string]interface{}, bool) {
	switch v := value.(type) {
	case map[string]interface{}:
		return normalizeMap(v), true
	case map[interface{}]interface{}:
		return normalizeMapInterface(v), true
	default:
		return nil, false
	}
}

func normalizeMap(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{}, len(input))
	for k, v := range input {
		output[k] = normalizeValue(v)
	}
	return output
}

func normalizeMapInterface(input map[interface{}]interface{}) map[string]interface{} {
	output := make(map[string]interface{}, len(input))
	for k, v := range input {
		output[fmt.Sprintf("%v", k)] = normalizeValue(v)
	}
	return output
}

func normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return normalizeMap(v)
	case map[interface{}]interface{}:
		return normalizeMapInterface(v)
	case []interface{}:
		for i := range v {
			v[i] = normalizeValue(v[i])
		}
		return v
	default:
		return v
	}
}

func deepMergeMap(target map[string]interface{}, override map[string]interface{}) {
	for k, overrideValue := range override {
		if targetValue, exists := target[k]; exists {
			targetMap, targetIsMap := toMap(targetValue)
			overrideMap, overrideIsMap := toMap(overrideValue)
			if targetIsMap && overrideIsMap {
				deepMergeMap(targetMap, overrideMap)
				target[k] = targetMap
				continue
			}
		}
		target[k] = normalizeValue(overrideValue)
	}
}

func configToMap(config KrakenDConfig) (map[string]interface{}, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}
