package main

import (
	"log/slog"
	"os"

	"github.com/primelib/primecodegen/pkg/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	// pass version info the version cmd
	cmd.Version = version
	cmd.CommitHash = commit
	cmd.BuildAt = date
	cmd.RepositoryStatus = status
}

// CLI Main Entrypoint
func main() {
	cmdErr := cmd.Execute()
	if cmdErr != nil {
		slog.Error("cli error", "err", cmdErr)
		os.Exit(1)
	}
}
