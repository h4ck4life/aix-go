package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	debugFlag bool
	version   string
)

// Execute runs the root command
func Execute(ver string) error {
	version = ver
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "aix",
	Short: "Tiny CLI to switch Anthropic-compatible endpoints and tokens",
	Long: `aix is a lightweight CLI tool for switching between Anthropic-compatible
API providers and tokens for Claude Code. It manages provider configurations
(endpoints, auth, model overrides) and writes environment variables to
~/.claude/settings.json.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug logging")

	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(initCmd)
}

func showVersion() {
	fmt.Printf("aix version %s\n", version)
}
