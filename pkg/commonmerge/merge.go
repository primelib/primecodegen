package commonmerge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/primelib/primecodegen/pkg/util"
	"gopkg.in/yaml.v3"
)

// ReadAndMergeFiles reads all files (yaml and json) and merges them into a single document
// The output format is determined by the first file's format
func ReadAndMergeFiles(files []string) ([]byte, error) {
	if len(files) == 0 {
		return nil, util.ErrNoFilesSpecified
	}
	if len(files) == 1 {
		if _, err := os.Stat(files[0]); os.IsNotExist(err) {
			return nil, errors.Join(util.ErrFileMissing, err)
		}

		return os.ReadFile(files[0])
	}

	var result map[string]interface{}
	var outputFormat string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, errors.Join(util.ErrFileMissing, err)
		}

		var data map[string]interface{}
		var err error
		switch {
		case strings.HasSuffix(file, ".json"):
			data, err = readJSON(file)
			if outputFormat == "" {
				outputFormat = "json"
			}
		case strings.HasSuffix(file, ".yaml"), strings.HasSuffix(file, ".yml"):
			data, err = readYAML(file)
			if outputFormat == "" {
				outputFormat = "yaml"
			}
		default:
			return nil, fmt.Errorf("unsupported file format: %s", file)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		result = deepMerge(result, data)
	}

	var resultBytes []byte
	var err error
	if outputFormat == "json" {
		resultBytes, err = json.Marshal(result)
	} else if outputFormat == "yaml" {
		resultBytes, err = yaml.Marshal(result)
	}
	if err != nil {
		return nil, errors.Join(util.ErrJSONMarshal, err)
	}

	return resultBytes, nil
}

// readYAML reads a YAML file and unmarshals it into a map
func readYAML(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// readJSON reads a JSON file and unmarshals it into a map
func readJSON(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// deepMerge recursively merges two maps
func deepMerge(dst, src map[string]interface{}) map[string]interface{} {
	if dst == nil {
		dst = make(map[string]interface{})
	}

	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			switch dstValTyped := dstVal.(type) {
			case map[string]interface{}:
				if srcValTyped, ok := srcVal.(map[string]interface{}); ok {
					dst[key] = deepMerge(dstValTyped, srcValTyped)
				} else {
					dst[key] = srcVal
				}
			default:
				dst[key] = srcVal
			}
		} else {
			dst[key] = srcVal
		}
	}
	return dst
}
