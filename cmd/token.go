package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/utils"
)

var tokenCmd = &cobra.Command{
	Use:     "token",
	Aliases: []string{"t"},
	Short:   "Manage auth tokens",
	Long:    "Set, test, and remove API tokens for providers.",
}

var (
	tokenSetToken string
	tokenRemoveYes bool
)

func init() {
	tokenCmd.AddCommand(tokenSetCmd)
	tokenCmd.AddCommand(tokenTestCmd)
	tokenCmd.AddCommand(tokenRemoveCmd)

	tokenSetCmd.Flags().StringVarP(&tokenSetToken, "token", "t", "", "Token value")
	tokenRemoveCmd.Flags().BoolVar(&tokenRemoveYes, "yes", false, "Skip confirmation")
}

var tokenSetCmd = &cobra.Command{
	Use:   "set [provider]",
	Short: "Set token for a provider",
	RunE:  runTokenSet,
}

var tokenTestCmd = &cobra.Command{
	Use:   "test [provider]",
	Short: "Test token for a provider",
	RunE:  runTokenTest,
}

var tokenRemoveCmd = &cobra.Command{
	Use:     "remove [provider]",
	Aliases: []string{"rm"},
	Short:   "Remove token for a provider",
	RunE:    runTokenRemove,
}

func runTokenSet(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix token set <provider>")
	}
	name := args[0]

	registry := core.NewRegistry()
	if _, err := registry.GetOne(name); err != nil {
		return err
	}

	token := tokenSetToken
	if token == "" {
		tokenModel := ui.NewSecureInput(fmt.Sprintf("Enter token for '%s':", name))
		if _, err := ui.RunPrompt(tokenModel); err != nil {
			return err
		}
		if tokenModel.Cancelled() {
			fmt.Println("Cancelled")
			return nil
		}
		token = tokenModel.Value()
	}

	if token == "" {
		return utils.NewValidationError("token", "token is required")
	}

	tokenMgr := core.NewTokenManager()
	if err := tokenMgr.SetToken(name, token); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Token set for provider '%s'", name)))
	return nil
}

func runTokenTest(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix token test <provider>")
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

	fmt.Printf("Testing token for '%s'...\n", name)
	start := time.Now()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", cfg.BaseURL, nil)
	if err != nil {
		return err
	}

	if cfg.TokenVar == constants.TokenTypeAPIKey {
		req.Header.Set("x-api-key", token)
		req.Header.Set("anthropic-api-key", token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Failed: %v (%s)", err, elapsed)))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println(ui.Success(fmt.Sprintf("OK (%s)", elapsed)))
	} else {
		fmt.Println(ui.Warning(fmt.Sprintf("HTTP %d (%s)", resp.StatusCode, elapsed)))
	}

	return nil
}

func runTokenRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return utils.NewValidationError("args", "usage: aix token remove <provider>")
	}
	name := args[0]

	if !tokenRemoveYes {
		confirmModel := ui.NewConfirm(fmt.Sprintf("Remove token for provider '%s'?", name), false)
		if _, err := ui.RunPrompt(confirmModel); err != nil {
			return err
		}
		if confirmModel.Cancelled() || !confirmModel.Value() {
			fmt.Println("Cancelled")
			return nil
		}
	}

	tokenMgr := core.NewTokenManager()
	if err := tokenMgr.DeleteToken(name); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("Token removed for provider '%s'", name)))
	return nil
}
