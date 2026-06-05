package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	Short:   "View and manage configuration",
	Long:    "View current config, export, or import provider settings.",
}

var (
	configExportFormat string
	configExportOutput string
	configImportMerge  bool
)

func init() {
	configCmd.AddCommand(configCurrentCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)

	configExportCmd.Flags().StringVarP(&configExportFormat, "format", "f", "json", "Export format (json)")
	configExportCmd.Flags().StringVarP(&configExportOutput, "output", "o", "", "Output file")
	configImportCmd.Flags().BoolVar(&configImportMerge, "merge", false, "Merge with existing providers")
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current configuration",
	RunE:  runConfigCurrent,
}

var configExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export configuration",
	RunE:  runConfigExport,
}

var configImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import configuration",
	RunE:  runConfigImport,
}

func runConfigCurrent(cmd *cobra.Command, args []string) error {
	env := readAnthropicEnv()
	if len(env) == 0 {
		fmt.Println("No active provider configured")
		return nil
	}

	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := []string{"Key", "Value"}
	rows := make([][]string, 0, len(env))
	for _, k := range keys {
		rows = append(rows, []string{k, env[k]})
	}

	fmt.Println(ui.RenderSimpleTable(headers, rows))
	return nil
}

func readAnthropicEnv() map[string]string {
	keys := []string{
		constants.EnvAnthropicBaseURL,
		constants.EnvAnthropicAPIKey,
		constants.EnvAnthropicAuthToken,
		constants.EnvAnthropicModel,
		constants.EnvDefaultOpusModel,
		constants.EnvDefaultSonnetModel,
		constants.EnvDefaultHaikuModel,
		constants.EnvDefaultSubagentModel,
	}
	env := make(map[string]string)
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			env[k] = v
		}
	}
	return env
}

func runConfigExport(cmd *cobra.Command, args []string) error {
	registry := core.NewRegistry()
	providers, err := registry.GetAll()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return err
	}

	if configExportOutput != "" {
		if err := os.WriteFile(configExportOutput, data, 0600); err != nil {
			return err
		}
		fmt.Println(ui.Success(fmt.Sprintf("Configuration exported to %s", configExportOutput)))
	} else {
		fmt.Println(string(data))
	}

	return nil
}

func runConfigImport(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix config import <file>")
	}
	file := args[0]

	data, err := os.ReadFile(file)
	if err != nil {
		return utils.NewFileNotFoundError(file)
	}

	var imported map[string]constants.ProviderConfig
	if err := json.Unmarshal(data, &imported); err != nil {
		return utils.NewValidationError("import", fmt.Sprintf("failed to parse import file: %v", err))
	}

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	existing, err := registry.GetAll()
	if err != nil {
		return utils.NewValidationError("registry", fmt.Sprintf("failed to read existing providers: %v", err))
	}

	for name, cfg := range imported {
		if configImportMerge {
			if _, ok := existing[name]; ok {
				fmt.Printf("Merging provider '%s'...\n", name)
				if err := registry.MergeOne(name, cfg); err != nil {
					return err
				}
				continue
			}
		}
		if err := registry.SetOne(name, cfg); err != nil {
			return err
		}
	}

	fmt.Println(ui.Success(fmt.Sprintf("Imported %d provider(s)", len(imported))))
	return nil
}
