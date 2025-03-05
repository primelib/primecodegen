package specutil

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
)

type Diff struct {
	OpenAPI []OpenAPIDiff
}

func DiffSpec(format string, file1 string, file2 string) (Diff, error) {
	var diff = Diff{
		OpenAPI: []OpenAPIDiff{},
	}

	// check of files exist
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		return diff, fmt.Errorf("file %s does not exist", file1)
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		return diff, fmt.Errorf("file %s does not exist", file2)
	}

	// diff openapi
	if format == "openapi" || format == "" {
		d, err := DiffOpenAPI(file1, file2)
		if err != nil {
			return diff, fmt.Errorf("failed to diff openapi: %w", err)
		}

		// sort by level
		sort.Slice(d, func(i, j int) bool {
			return d[i].Level > d[j].Level
		})

		diff.OpenAPI = d
	}

	return diff, nil
}

func BumpVersion(format string, file1 string, file2 string, currentVersion string) (string, error) {
	// parse current version
	v, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", fmt.Errorf("failed to parse current version: %w", err)
	}

	// set initial version to 0.1.0 if no old spec is available
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		return "0.1.0", nil
	}

	// require spec
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", file2)
	}

	// diff openapi
	if format == "openapi" || format == "" {
		d, err := DiffOpenAPI(file1, file2)
		if err != nil {
			return "", fmt.Errorf("failed to diff openapi: %w", err)
		}

		maxLevel := 0
		for _, r := range d {
			if r.Level > maxLevel {
				maxLevel = r.Level
			}
		}

		if maxLevel == 3 {
			*v = v.IncMajor()
		} else if maxLevel == 2 {
			*v = v.IncMinor()
		} else {
			*v = v.IncPatch()
		}
	}

	return v.String(), nil
}
