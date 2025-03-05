package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/primelib/primecodegen/pkg/app/appcmd"
	"github.com/primelib/primecodegen/pkg/openapi/openapicmd"
	"github.com/spf13/cobra"
)

var cfg zerologconfig.LogConfig

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `primecodegen`,
		Short: `PrimeCodeGen is a code generator for API specifications.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerologconfig.Configure(cfg)
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(zerologconfig.ValidLogLevels, ","))
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(zerologconfig.ValidLogFormats, ","))
	cmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")

	cmd.AddCommand(versionCmd())

	// openapi
	cmd.AddGroup(&cobra.Group{ID: "openapi", Title: "OpenAPI Generation"})
	cmd.AddCommand(openapicmd.ConvertCmd())
	cmd.AddCommand(openapicmd.MergeCmd())
	cmd.AddCommand(openapicmd.PatchCmd())
	cmd.AddCommand(openapicmd.GenerateCmd())
	cmd.AddCommand(openapicmd.GenerateTemplateCmd())

	// app
	cmd.AddGroup(&cobra.Group{ID: "vcsapp", Title: "VCS App"})
	cmd.AddCommand(appcmd.UpdateCmd())
	cmd.AddCommand(appcmd.GenerateCmd())
	cmd.AddCommand(appcmd.ReleaseCmd())
	cmd.AddCommand(appcmd.ListCmd())

	return cmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd().Execute()
}
