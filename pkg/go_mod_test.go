package pkg

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Module struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
}

func TestLibopenapiModuleVersion(t *testing.T) {
	// github.com/pb33f/libopenapi v0.18.5 introduces a bug resulting in too many properties merged in InlineAllOfHierarchies
	moduleName := "github.com/pb33f/libopenapi"
	maxAllowedVersion := "v0.18.4"

	// Get the module version information
	cmd := exec.Command("go", "list", "-m", "-json", moduleName)
	output, err := cmd.Output()
	assert.NoError(t, err)

	// Parse the JSON output
	var module Module
	err = json.Unmarshal(output, &module)
	assert.NoError(t, err)
	// Compare the module version with the maximum allowed version
	if compareVersions(module.Version, maxAllowedVersion) > 0 {
		t.Fatalf("Module %s version %s exceeds the maximum allowed version %s", moduleName, module.Version, maxAllowedVersion)
	}
}

// compareVersions compares two semantic version strings.
// Returns 1 if v1 > v2, -1 if v1 < v2, and 0 if v1 == v2.
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		if v1Parts[i] > v2Parts[i] {
			return 1
		} else if v1Parts[i] < v2Parts[i] {
			return -1
		}
	}

	if len(v1Parts) > len(v2Parts) {
		return 1
	} else if len(v1Parts) < len(v2Parts) {
		return -1
	}

	return 0
}
