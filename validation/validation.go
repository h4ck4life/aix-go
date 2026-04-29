package validation

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
)

// ValidateProviderName checks if a provider name is valid
func ValidateProviderName(name string) error {
	if name == "" {
		return errors.New("provider name is required")
	}
	if !constants.ProviderNamePattern.MatchString(name) {
		return fmt.Errorf("invalid provider name '%s': must start with lowercase letter and contain only lowercase letters, numbers, and hyphens", name)
	}
	return nil
}

// ValidateURL checks if a URL is valid
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("URL is required")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL '%s': %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL '%s': must use http or https scheme", rawURL)
	}
	if u.Host == "" {
		return fmt.Errorf("invalid URL '%s': host is required", rawURL)
	}
	return nil
}

// ValidateTokenVar checks if a token variable type is valid
func ValidateTokenVar(tokenVar string) error {
	normalized := NormalizeTokenVar(tokenVar)
	if normalized != constants.TokenTypeAPIKey && normalized != constants.TokenTypeAuthToken {
		return fmt.Errorf("invalid token type '%s': must be api-key or auth-token", tokenVar)
	}
	return nil
}

// NormalizeTokenVar converts short forms to full env var names
func NormalizeTokenVar(tokenVar string) string {
	switch strings.ToLower(tokenVar) {
	case "api-key", "apikey", "api_key":
		return constants.TokenTypeAPIKey
	case "auth-token", "authtoken", "auth_token":
		return constants.TokenTypeAuthToken
	default:
		return tokenVar
	}
}

// ValidateShell checks if a shell type is valid
func ValidateShell(shell string) error {
	validShells := []string{constants.ShellBash, constants.ShellZsh, constants.ShellFish, constants.ShellPowerShell, constants.ShellCmd}
	for _, s := range validShells {
		if s == shell {
			return nil
		}
	}
	return fmt.Errorf("invalid shell '%s': must be one of %v", shell, validShells)
}

// ValidateModelName checks if a model name is valid
func ValidateModelName(model string) error {
	if model == "" {
		return errors.New("model name is required")
	}
	return nil
}

// ValidateModelAlias checks if a model alias is valid
func ValidateModelAlias(alias string) error {
	validAliases := []string{constants.ModelOpus, constants.ModelSonnet, constants.ModelHaiku, constants.ModelSubagent}
	for _, a := range validAliases {
		if a == alias {
			return nil
		}
	}
	return fmt.Errorf("invalid model alias '%s': must be one of %v", alias, validAliases)
}
