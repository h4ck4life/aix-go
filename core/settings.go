package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/utils"
)

// shellQuote returns a shell-safe quoted string for the given shell.
func shellQuote(shell, value string) string {
	switch shell {
	case constants.ShellPowerShell:
		// Single-quote wrapping: everything is literal.
		// Escape embedded single quotes as '' (double-single-quote).
		escaped := strings.ReplaceAll(value, "'", "''")
		return "'" + escaped + "'"
	case constants.ShellCmd:
		// CMD has limited escaping. Use ^ for metacharacters and %% for %.
		// Wrap in double quotes to handle spaces and special characters.
		escaped := value
		// Escape CMD metacharacters with ^
		for _, ch := range []string{"^", "&", "|", "<", ">"} {
			escaped = strings.ReplaceAll(escaped, ch, "^"+ch)
		}
		// Escape % as %%
		escaped = strings.ReplaceAll(escaped, "%", "%%")
		// Escape " as ^"
		escaped = strings.ReplaceAll(escaped, `"`, `^"`)
		return `"` + escaped + `"`
	default:
		// bash, zsh, fish: use single quotes, escape single quotes
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
}

// Settings represents Claude Code's settings.json
type Settings struct {
	Env map[string]string `json:"env,omitempty"`
}

// Read reads the settings file
func (s *Settings) Read() error {
	path := constants.SettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.Env = make(map[string]string)
			return nil
		}
		return utils.NewFileNotFoundError(path)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return utils.NewValidationError("settings", fmt.Sprintf("failed to parse settings: %v", err))
	}

	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}
	*s = settings
	return nil
}

// Write writes the settings file
func (s *Settings) Write() error {
	path := constants.SettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// GenerateEnvironmentVars builds env vars from provider config
func (s *Settings) GenerateEnvironmentVars(providerName string, provider constants.ProviderConfig, token string) error {
	if s.Env == nil {
		s.Env = make(map[string]string)
	}

	// Set base URL
	s.Env[constants.EnvAnthropicBaseURL] = provider.BaseURL

	// Set token with exclusivity
	if provider.TokenVar == constants.TokenTypeAPIKey {
		s.Env[constants.EnvAnthropicAPIKey] = token
		delete(s.Env, constants.EnvAnthropicAuthToken)
	} else {
		s.Env[constants.EnvAnthropicAuthToken] = token
		delete(s.Env, constants.EnvAnthropicAPIKey)
	}

	// Set custom model if present
	if provider.ModelName != "" {
		s.Env[constants.EnvAnthropicModel] = provider.ModelName
	} else {
		delete(s.Env, constants.EnvAnthropicModel)
	}

	// Set default model aliases
	for alias, model := range provider.DefaultModels {
		if envVar := constants.ModelAliasToEnvVar(alias); envVar != "" {
			s.Env[envVar] = model
		}
	}

	// Clean up aliases that are no longer present
	allAliases := []string{constants.ModelOpus, constants.ModelSonnet, constants.ModelHaiku, constants.ModelSubagent}
	for _, alias := range allAliases {
		envVar := constants.ModelAliasToEnvVar(alias)
		if _, ok := provider.DefaultModels[alias]; !ok {
			delete(s.Env, envVar)
		}
	}

	return nil
}

// FormatForShell outputs shell export commands
func (s *Settings) FormatForShell(shell string) string {
	var sb strings.Builder
	for key, value := range s.Env {
		quoted := shellQuote(shell, value)
		switch shell {
		case constants.ShellFish:
			sb.WriteString(fmt.Sprintf("set -x %s %s\n", key, quoted))
		case constants.ShellPowerShell:
			sb.WriteString(fmt.Sprintf("$env:%s = %s\n", key, quoted))
		case constants.ShellCmd:
			sb.WriteString(fmt.Sprintf("set %s=%s\n", key, quoted))
		default:
			sb.WriteString(fmt.Sprintf("export %s=%s\n", key, quoted))
		}
	}
	return sb.String()
}

// GetCurrentEnvironment returns the current env vars
func (s *Settings) GetCurrentEnvironment() map[string]string {
	if s.Env == nil {
		return make(map[string]string)
	}
	return copyMap(s.Env)
}

// GetCurrentModel returns the current model name
func (s *Settings) GetCurrentModel() string {
	return s.Env[constants.EnvAnthropicModel]
}

// GetCurrentProvider returns the current provider base URL
func (s *Settings) GetCurrentProvider() string {
	return s.Env[constants.EnvAnthropicBaseURL]
}

// Reset clears all settings
func (s *Settings) Reset() error {
	s.Env = make(map[string]string)
	return s.Write()
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
