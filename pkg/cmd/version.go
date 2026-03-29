package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/primelib/primecodegen/pkg/constants"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "print version information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(os.Stdout, "GitVersion:    %s\n", constants.Version)
			_, _ = fmt.Fprintf(os.Stdout, "GitCommit:     %s\n", constants.CommitHash)
			_, _ = fmt.Fprintf(os.Stdout, "GitTreeState:  %s\n", constants.RepositoryStatus)
			_, _ = fmt.Fprintf(os.Stdout, "BuildDate:     %s\n", constants.BuildAt)
			_, _ = fmt.Fprintf(os.Stdout, "GoVersion:     %s\n", runtime.Version())
			_, _ = fmt.Fprintf(os.Stdout, "Compiler:      %s\n", runtime.Compiler)
			_, _ = fmt.Fprintf(os.Stdout, "Platform:      %s\n", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}

	return cmd
}
