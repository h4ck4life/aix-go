package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/interactive"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/utils"
	"github.com/h4ck4life/aix-go/validation"
)

var providerCmd = &cobra.Command{
	Use:     "provider",
	Aliases: []string{"p"},
	Short:   "Manage Anthropic-compatible API providers",
	Long:    "Add, list, remove, rename, and switch between API providers.",
}

var (
	providerAddInteractive bool
	providerAddTokenType   string
	providerAddDesc        string
	providerRemoveYes      bool
	providerRenameYes      bool
	providerUseShell       string
)

func init() {
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerRenameCmd)
	providerCmd.AddCommand(providerUseCmd)
	providerCmd.AddCommand(providerSetModelCmd)
	providerCmd.AddCommand(providerSetDefaultCmd)

	providerAddCmd.Flags().BoolVarP(&providerAddInteractive, "interactive", "i", false, "Interactive mode")
	providerAddCmd.Flags().StringVarP(&providerAddTokenType, "token-type", "t", "", "Token type (api-key or auth-token)")
	providerAddCmd.Flags().StringVarP(&providerAddDesc, "description", "d", "", "Description")

	providerRemoveCmd.Flags().BoolVar(&providerRemoveYes, "yes", false, "Skip confirmation")
	providerRenameCmd.Flags().BoolVar(&providerRenameYes, "yes", false, "Skip confirmation")
	providerUseCmd.Flags().StringVar(&providerUseShell, "shell", "", "Output shell export commands")
}

var providerAddCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a new provider",
	RunE:  runProviderAdd,
}

var providerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all providers",
	RunE:    runProviderList,
}

var providerRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a provider",
	RunE:    runProviderRemove,
}

var providerRenameCmd = &cobra.Command{
	Use:     "rename [old-name] [new-name]",
	Aliases: []string{"mv"},
	Short:   "Rename a provider",
	RunE:    runProviderRename,
}

var providerUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Set active provider",
	RunE:  runProviderUse,
}

var providerSetModelCmd = &cobra.Command{
	Use:   "set-model [name] [model]",
	Short: "Set custom model for a provider",
	RunE:  runProviderSetModel,
}

var providerSetDefaultCmd = &cobra.Command{
	Use:   "set-default [name] [alias] [model]",
	Short: "Set default model alias (opus, sonnet, haiku, subagent)",
	RunE:  runProviderSetDefault,
}

func runProviderAdd(cmd *cobra.Command, args []string) error {
	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	var cfg *constants.ProviderConfig
	var name string

	if providerAddInteractive {
		var err error
		cfg, _, err = interactive.RunAddProviderWizard()
		if err != nil {
			return err
		}
		name = cfg.BaseURL // This will be overwritten below
	} else {
		if len(args) < 2 {
			return utils.NewValidationError("args", "usage: aix provider add <name> <url>")
		}
		name = args[0]
		baseURL := args[1]

		if err := validation.ValidateProviderName(name); err != nil {
			return err
		}
		if err := validation.ValidateURL(baseURL); err != nil {
			return err
		}

		tokenVar := constants.TokenTypeAPIKey
		if providerAddTokenType != "" {
			tokenVar = validation.NormalizeTokenVar(providerAddTokenType)
			if err := validation.ValidateTokenVar(tokenVar); err != nil {
				return err
			}
		} else {
			tokenVar = interactive.DetectTokenType(baseURL)
		}

		cfg = &constants.ProviderConfig{
			BaseURL:  baseURL,
			TokenVar: tokenVar,
		}
	}

	// For interactive mode, we need to get the name from the wizard
	if providerAddInteractive {
		// The wizard doesn't return the name directly, we need to handle this differently
		// For now, let's just use the first argument or prompt
		fmt.Println("Interactive mode not fully implemented yet")
		return nil
	}

	if err := registry.SetOne(name, *cfg); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Provider '%s' added successfully", name)))
	return nil
}

func runProviderList(cmd *cobra.Command, args []string) error {
	registry := core.NewRegistry()
	providers, err := registry.GetAll()
	if err != nil {
		return err
	}

	tokenMgr := core.NewTokenManager()

	if len(providers) == 0 {
		fmt.Println("No providers configured")
		return nil
	}

	headers := []string{"Name", "Base URL", "Token Type", "Custom Model", "Token"}
	rows := make([][]string, 0, len(providers))

	for name, cfg := range providers {
		tokenStatus := ui.Cross()
		if tokenMgr.HasToken(name) {
			tokenStatus = ui.Checkmark()
		}

		model := cfg.ModelName
		if model == "" {
			model = "-"
		}

		rows = append(rows, []string{
			name,
			cfg.BaseURL,
			cfg.TokenVar,
			model,
			tokenStatus,
		})
	}

	fmt.Println(ui.RenderSimpleTable(headers, rows))
	return nil
}

func runProviderRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix provider remove <name>")
	}
	name := args[0]

	if !providerRemoveYes {
		confirmModel := ui.NewConfirm(fmt.Sprintf("Remove provider '%s' and its token?", name), false)
		if _, err := ui.RunPrompt(confirmModel); err != nil {
			return err
		}
		if confirmModel.Cancelled() || !confirmModel.Value() {
			fmt.Println("Cancelled")
			return nil
		}
	}

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	tokenMgr := core.NewTokenManager()
	_ = tokenMgr.DeleteToken(name)

	if err := registry.RemoveOne(name); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Provider '%s' removed", name)))
	return nil
}

func runProviderRename(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return utils.NewValidationError("args", "usage: aix provider rename <old-name> <new-name>")
	}
	oldName, newName := args[0], args[1]

	if !providerRenameYes {
		confirmModel := ui.NewConfirm(fmt.Sprintf("Rename provider '%s' to '%s'?", oldName, newName), false)
		if _, err := ui.RunPrompt(confirmModel); err != nil {
			return err
		}
		if confirmModel.Cancelled() || !confirmModel.Value() {
			fmt.Println("Cancelled")
			return nil
		}
	}

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	tokenMgr := core.NewTokenManager()
	_ = tokenMgr.MoveToken(oldName, newName)

	if err := registry.RenameOne(oldName, newName); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Provider renamed from '%s' to '%s'", oldName, newName)))
	return nil
}

func runProviderUse(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix provider use <name>")
	}
	name := args[0]

	registry := core.NewRegistry()
	cfg, err := registry.GetOne(name)
	if err != nil {
		return err
	}

	tokenMgr := core.NewTokenManager()
	token, err := tokenMgr.GetToken(name)
	if err != nil {
		return utils.NewTokenError(fmt.Sprintf("no token found for provider '%s': %v", name, err))
	}

	settings := &core.Settings{}
	if err := settings.GenerateEnvironmentVars(name, cfg, token); err != nil {
		return err
	}

	shell := providerUseShell
	if shell == "" {
		shell = constants.DetectShell()
	}
	if err := validation.ValidateShell(shell); err != nil {
		return err
	}

	fmt.Print(settings.FormatForShell(shell))
	return nil
}

func runProviderSetModel(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return utils.NewValidationError("args", "usage: aix provider set-model <name> <model>")
	}
	name, model := args[0], args[1]

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	if err := registry.SetModelName(name, model); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Set model '%s' for provider '%s'", model, name)))
	return nil
}

func runProviderSetDefault(cmd *cobra.Command, args []string) error {
	if len(args) < 3 {
		return utils.NewValidationError("args", "usage: aix provider set-default <name> <alias> <model>")
	}
	name, alias, model := args[0], args[1], args[2]

	if err := validation.ValidateModelAlias(alias); err != nil {
		return err
	}

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	if err := registry.SetDefaultModel(name, alias, model); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Set default %s model '%s' for provider '%s'", alias, model, name)))
	return nil
}
