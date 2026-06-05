package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/interactive"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/utils"
	"github.com/h4ck4life/aix-go/validation"
	"github.com/spf13/cobra"
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
	providerRemoveYes      bool
	providerRenameYes      bool
	providerUseShell       string

	providerEditURL          string
	providerEditTokenType    string
	providerEditModel        string
	providerEditDefaultModel string
)

func init() {
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerRenameCmd)
	providerCmd.AddCommand(providerUseCmd)
	providerCmd.AddCommand(providerSetModelCmd)
	providerCmd.AddCommand(providerSetDefaultCmd)
	providerCmd.AddCommand(providerEditCmd)

	providerAddCmd.Flags().BoolVarP(&providerAddInteractive, "interactive", "i", false, "Interactive mode")
	providerAddCmd.Flags().StringVarP(&providerAddTokenType, "token-type", "t", "", "Token type (api-key or auth-token)")

	providerRemoveCmd.Flags().BoolVar(&providerRemoveYes, "yes", false, "Skip confirmation")
	providerRenameCmd.Flags().BoolVar(&providerRenameYes, "yes", false, "Skip confirmation")
	providerUseCmd.Flags().StringVar(&providerUseShell, "shell", "", "Output shell export commands")

	providerEditCmd.Flags().StringVar(&providerEditURL, "url", "", "New base URL")
	providerEditCmd.Flags().StringVar(&providerEditTokenType, "token-type", "", "New token type (api-key or auth-token)")
	providerEditCmd.Flags().StringVar(&providerEditModel, "model", "", "New custom model")
	providerEditCmd.Flags().StringVar(&providerEditDefaultModel, "default-model", "", "Set default model alias (format: alias=model)")
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

var providerEditCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an existing provider",
	RunE:  runProviderEdit,
}

func runProviderAdd(cmd *cobra.Command, args []string) error {
	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	var cfg *constants.ProviderConfig
	var name, token string

	if providerAddInteractive {
		var err error
		name, cfg, token, err = interactive.RunAddProviderWizard()
		if err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return utils.NewValidationError("args", "usage: aix provider add <name> <url>")
		}
		name = args[0]
		baseURL := args[1]

		if err := validation.ValidateProviderName(name); err != nil {
			return utils.NewValidationError("name", err.Error())
		}
		if err := validation.ValidateURL(baseURL); err != nil {
			return utils.NewValidationError("url", err.Error())
		}

		tokenVar := constants.TokenTypeAPIKey
		if providerAddTokenType != "" {
			tokenVar = validation.NormalizeTokenVar(providerAddTokenType)
			if err := validation.ValidateTokenVar(tokenVar); err != nil {
				return utils.NewValidationError("token-type", err.Error())
			}
		} else {
			tokenVar = interactive.DetectTokenType(baseURL)
		}

		cfg = &constants.ProviderConfig{
			BaseURL:  baseURL,
			TokenVar: tokenVar,
		}
	}

	// Registry write first — only store the token after the provider entry exists.
	// This avoids orphaned tokens under a name that is not in the registry.
	if err := registry.SetOne(name, *cfg); err != nil {
		return err
	}

	if token != "" {
		tokenMgr := core.NewTokenManager()
		if err := tokenMgr.SetToken(name, token); err != nil {
			return utils.NewTokenError(fmt.Sprintf("provider '%s' added but failed to store token: %v", name, err))
		}
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

	// Sort providers by name for consistent output
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cfg := providers[name]
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
	if err := tokenMgr.DeleteToken(name); err != nil {
		fmt.Println(ui.Warning(fmt.Sprintf("Warning: could not delete token for '%s': %v", name, err)))
	}

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

	// Rename registry first. If this fails, the token is still under oldName
	// and the user can retry. If we moved the token first and the registry
	// rename failed, the token would be orphaned under newName.
	if err := registry.RenameOne(oldName, newName); err != nil {
		return err
	}

	tokenMgr := core.NewTokenManager()
	if err := tokenMgr.MoveToken(oldName, newName); err != nil {
		// Roll back the registry rename so the provider and token stay aligned.
		if rbErr := registry.RenameOne(newName, oldName); rbErr != nil {
			return utils.NewTokenError(fmt.Sprintf("failed to move token from '%s' to '%s': %v (rollback also failed: %v)", oldName, newName, err, rbErr))
		}
		return utils.NewTokenError(fmt.Sprintf("failed to move token from '%s' to '%s': %v (rolled back registry rename)", oldName, newName, err))
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
		return utils.NewValidationError("shell", err.Error())
	}

	fmt.Print(settings.FormatForShell(shell))
	return nil
}

func runProviderSetModel(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return utils.NewValidationError("args", "usage: aix provider set-model <name> <model>")
	}
	name, model := args[0], args[1]

	if err := validation.ValidateModelName(model); err != nil {
		return utils.NewValidationError("model", err.Error())
	}

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
		return utils.NewValidationError("alias", err.Error())
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

func runProviderEdit(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix provider edit <name>")
	}
	name := args[0]

	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	cfg, err := registry.GetOne(name)
	if err != nil {
		return err
	}

	modified := false

	if providerEditURL != "" {
		if err := validation.ValidateURL(providerEditURL); err != nil {
			return utils.NewValidationError("url", err.Error())
		}
		cfg.BaseURL = providerEditURL
		modified = true
	}

	if providerEditTokenType != "" {
		tokenVar := validation.NormalizeTokenVar(providerEditTokenType)
		if err := validation.ValidateTokenVar(tokenVar); err != nil {
			return utils.NewValidationError("token-type", err.Error())
		}
		cfg.TokenVar = tokenVar
		modified = true
	}

	if providerEditModel != "" {
		if err := validation.ValidateModelName(providerEditModel); err != nil {
			return utils.NewValidationError("model", err.Error())
		}
		cfg.ModelName = providerEditModel
		modified = true
	}

	if providerEditDefaultModel != "" {
		parts := strings.SplitN(providerEditDefaultModel, "=", 2)
		if len(parts) != 2 {
			return utils.NewValidationError("default-model", "format must be alias=model")
		}
		alias, model := parts[0], parts[1]
		if err := validation.ValidateModelAlias(alias); err != nil {
			return utils.NewValidationError("alias", err.Error())
		}
		if cfg.DefaultModels == nil {
			cfg.DefaultModels = make(map[string]string)
		}
		cfg.DefaultModels[alias] = model
		modified = true
	}

	if !modified {
		return utils.NewValidationError("args", "no changes specified; use --url, --token-type, --model, or --default-model")
	}

	if err := registry.SetOne(name, cfg); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Provider '%s' updated", name)))
	return nil
}
