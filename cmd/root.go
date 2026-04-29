package cmd

import (
	"fmt"

	"github.com/h4ck4life/aix-go/utils"
	"github.com/spf13/cobra"
)

var (
	debugFlag bool
	version   string
)

// Execute runs the root command
func Execute(ver string) error {
	version = ver
	rootCmd.Version = ver
	utils.InitLogger(debugFlag)
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aix version %s\n", version)
	},
}

var rootCmd = &cobra.Command{
	Use:   "aix",
	Short: "Tiny CLI to switch Anthropic-compatible endpoints and tokens",
	Long: `aix is a lightweight CLI tool for switching between Anthropic-compatible
API providers and tokens for Claude Code. It manages provider configurations
(endpoints, auth, model overrides) and outputs shell export commands to stdout.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug logging")
	rootCmd.SetVersionTemplate("aix version {{.Version}}\n")

	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}
