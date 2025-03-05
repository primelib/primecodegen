package specutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/primelib/primecodegen/pkg/util"
)

type OpenAPIDiff struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Level       int    `json:"level"`
	Operation   string `json:"operation"`
	OperationID string `json:"operationId"`
	Path        string `json:"path"`
	Source      string `json:"source"`
}

// DiffOpenAPI compares two OAS files and returns the differences, calls the oasdiff cli tool to retrieve the json
func DiffOpenAPI(file1 string, file2 string) ([]OpenAPIDiff, error) {
	// call cli
	cmd := exec.Command("oasdiff", "changelog", file1, file2, "-f", "json", "--exclude-elements", "examples")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to execute oasdiff: %w", err)
	}

	// parse json
	var diffs []OpenAPIDiff
	err = json.Unmarshal(stdout.Bytes(), &diffs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse oasdiff json output: %w", err)
	}

	// fix text
	for i, diff := range diffs {
		diffs[i].Text = util.StripANSI(diff.Text)
	}

	return diffs, nil
}
