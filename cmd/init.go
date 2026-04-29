package cmd

import (
	"github.com/spf13/cobra"
	"github.com/h4ck4life/aix-go/interactive"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "First-time setup wizard",
	Long:  "Interactive wizard to set up your first provider and token.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	return interactive.RunInitWizard()
}
