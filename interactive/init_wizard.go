package interactive

import (
	"fmt"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/validation"
)

// RunInitWizard runs the first-time setup wizard
func RunInitWizard() error {
	fmt.Println(ui.TitleStyle.Render("Welcome to aix!"))
	fmt.Println("Let's set up your first provider.")
	fmt.Println()

	// Step 1: Provider name
	nameModel := ui.NewInput("Provider name:", "e.g., my-provider")
	if _, err := ui.RunPrompt(nameModel); err != nil {
		return err
	}
	if nameModel.Cancelled() {
		return fmt.Errorf("cancelled")
	}
	name := nameModel.Value()
	if err := validation.ValidateProviderName(name); err != nil {
		return err
	}

	// Step 2: URL
	urlModel := ui.NewInput("Base URL:", "https://api.example.com")
	if _, err := ui.RunPrompt(urlModel); err != nil {
		return err
	}
	if urlModel.Cancelled() {
		return fmt.Errorf("cancelled")
	}
	baseURL := urlModel.Value()
	if err := validation.ValidateURL(baseURL); err != nil {
		return err
	}

	// Step 3: Token type (auto-detect or select)
	detected := DetectTokenType(baseURL)
	var tokenVar string
	if detected == constants.TokenTypeAuthToken {
		fmt.Println(ui.Info("Detected auth-token based on URL pattern"))
		tokenVar = constants.TokenTypeAuthToken
	} else {
		choice, err := ui.RunSelect("Token type:", []string{"api-key", "auth-token"})
		if err != nil {
			return err
		}
		tokenVar = validation.NormalizeTokenVar(choice)
	}

	// Step 4: Token
	tokenModel := ui.NewSecureInput("API token:")
	if _, err := ui.RunPrompt(tokenModel); err != nil {
		return err
	}
	if tokenModel.Cancelled() {
		return fmt.Errorf("cancelled")
	}
	token := tokenModel.Value()
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Step 5: Model override (optional)
	modelModel := ui.NewInput("Custom model (optional):", "e.g., claude-sonnet-4-6")
	if _, err := ui.RunPrompt(modelModel); err != nil {
		return err
	}
	modelName := ""
	if !modelModel.Cancelled() {
		modelName = modelModel.Value()
	}

	// Save provider
	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return err
	}

	cfg := constants.ProviderConfig{
		BaseURL:   baseURL,
		TokenVar:  tokenVar,
		ModelName: modelName,
	}

	if err := registry.SetOne(name, cfg); err != nil {
		return err
	}

	// Save token
	tokenMgr := core.NewTokenManager()
	if err := tokenMgr.SetToken(name, token); err != nil {
		return err
	}

	// Activate provider
	settings := &core.Settings{}
	if err := settings.Read(); err != nil {
		return err
	}
	if err := settings.GenerateEnvironmentVars(name, cfg, token); err != nil {
		return err
	}
	if err := settings.Write(); err != nil {
		return err
	}

	fmt.Println(ui.Success(fmt.Sprintf("\nProvider '%s' configured and activated!", name)))
	return nil
}
