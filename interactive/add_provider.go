package interactive

import (
	"fmt"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/h4ck4life/aix-go/validation"
)

// AddProviderWizard guides through adding a new provider
type AddProviderWizard struct {
	step        int
	name        string
	url         string
	tokenVar    string
	description string
	modelName   string
	defaultModels map[string]string
	err         error
}

// RunAddProviderWizard runs the interactive add provider wizard
func RunAddProviderWizard() (*constants.ProviderConfig, string, error) {
	wizard := &AddProviderWizard{}

	// Step 1: Name
	nameModel := ui.NewInput("Provider name:", "e.g., my-provider")
	if _, err := ui.RunPrompt(nameModel); err != nil {
		return nil, "", err
	}
	if nameModel.Cancelled() {
		return nil, "", fmt.Errorf("cancelled")
	}
	wizard.name = nameModel.Value()
	if err := validation.ValidateProviderName(wizard.name); err != nil {
		return nil, "", err
	}

	// Step 2: URL
	urlModel := ui.NewInput("Base URL:", "https://api.example.com")
	if _, err := ui.RunPrompt(urlModel); err != nil {
		return nil, "", err
	}
	if urlModel.Cancelled() {
		return nil, "", fmt.Errorf("cancelled")
	}
	wizard.url = urlModel.Value()
	if err := validation.ValidateURL(wizard.url); err != nil {
		return nil, "", err
	}

	// Step 3: Token type
	tokenType, err := ui.RunSelect("Token type:", []string{"api-key", "auth-token"})
	if err != nil {
		return nil, "", err
	}
	wizard.tokenVar = validation.NormalizeTokenVar(tokenType)

	// Step 4: Description (optional)
	descModel := ui.NewInput("Description (optional):", "")
	if _, err := ui.RunPrompt(descModel); err != nil {
		return nil, "", err
	}
	if !descModel.Cancelled() {
		wizard.description = descModel.Value()
	}

	// Step 5: Custom model (optional)
	modelModel := ui.NewInput("Custom model (optional):", "e.g., claude-sonnet-4-6")
	if _, err := ui.RunPrompt(modelModel); err != nil {
		return nil, "", err
	}
	if !modelModel.Cancelled() {
		wizard.modelName = modelModel.Value()
	}

	// Step 6: Default model aliases (optional)
	aliases := map[string]string{}
	for _, alias := range []string{"opus", "sonnet", "haiku", "subagent"} {
		aliasModel := ui.NewInput(
			fmt.Sprintf("Default %s model (optional):", alias),
			fmt.Sprintf("e.g., claude-%s-4-x", alias),
		)
		if _, err := ui.RunPrompt(aliasModel); err != nil {
			return nil, "", err
		}
		if !aliasModel.Cancelled() && aliasModel.Value() != "" {
			aliases[alias] = aliasModel.Value()
		}
	}
	if len(aliases) > 0 {
		wizard.defaultModels = aliases
	}

	// Step 7: Token (optional)
	tokenModel := ui.NewSecureInput("Token (optional, press Enter to skip):")
	if _, err := ui.RunPrompt(tokenModel); err != nil {
		return nil, "", err
	}

	cfg := constants.ProviderConfig{
		BaseURL:       wizard.url,
		TokenVar:      wizard.tokenVar,
		ModelName:     wizard.modelName,
		DefaultModels: wizard.defaultModels,
	}

	return &cfg, tokenModel.Value(), nil
}

// DetectTokenType tries to auto-detect token type from URL patterns
func DetectTokenType(url string) string {
	lower := strings.ToLower(url)
	if strings.Contains(lower, "minimax") {
		return constants.TokenTypeAuthToken
	}
	return constants.TokenTypeAPIKey
}
