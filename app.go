package main

import (
	"log/slog"
	"os"

	"github.com/primelib/primecodegen/pkg/cmd"
	"github.com/primelib/primecodegen/pkg/constants"
)

var (
	version = constants.Version
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	// Set Version Information
	constants.Version = version
	constants.CommitHash = commit
	constants.BuildAt = date
	constants.RepositoryStatus = status
}

// CLI Main Entrypoint
func main() {
	cmdErr := cmd.Execute()
	if cmdErr != nil {
		slog.Error("cli error", "err", cmdErr)
		os.Exit(1)
	}
}
