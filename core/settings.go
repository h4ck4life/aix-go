package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
// Preserves non-env keys (permissions, hooks, etc.) from the original file
type Settings struct {
	Env map[string]string      `json:"env,omitempty"`
	Raw map[string]interface{} `json:"-"` // full JSON object; preserved keys go here
}

// Read reads the settings file
func (s *Settings) Read() error {
	path := constants.SettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.Env = make(map[string]string)
			s.Raw = make(map[string]interface{})
			return nil
		}
		return utils.NewFileNotFoundError(path)
	}

	// First unmarshal into raw map to preserve all keys
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return utils.NewValidationError("settings", fmt.Sprintf("failed to parse settings: %v", err))
	}
	s.Raw = raw

	// Extract env key into typed map
	s.Env = make(map[string]string)
	if envRaw, ok := raw["env"]; ok {
		if envMap, ok := envRaw.(map[string]interface{}); ok {
			for k, v := range envMap {
				if str, ok := v.(string); ok {
					s.Env[k] = str
				}
			}
		}
	}

	return nil
}

// Write writes the settings file, preserving non-env keys from the original
func (s *Settings) Write() error {
	path := constants.SettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Merge current Env back into Raw
	if s.Raw == nil {
		s.Raw = make(map[string]interface{})
	}
	envMap := make(map[string]interface{}, len(s.Env))
	for k, v := range s.Env {
		envMap[k] = v
	}
	s.Raw["env"] = envMap

	tmpPath := path + ".tmp"
	data, err := json.MarshalIndent(s.Raw, "", "  ")
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

// FormatForShell outputs shell export commands with deterministic key ordering
func (s *Settings) FormatForShell(shell string) string {
	keys := make([]string, 0, len(s.Env))
	for key := range s.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		value := s.Env[key]
		quoted := shellQuote(shell, value)
		switch shell {
		case constants.ShellFish:
			fmt.Fprintf(&sb, "set -x %s %s\n", key, quoted)
		case constants.ShellPowerShell:
			fmt.Fprintf(&sb, "$env:%s = %s\n", key, quoted)
		case constants.ShellCmd:
			fmt.Fprintf(&sb, "set %s=%s\n", key, quoted)
		default:
			fmt.Fprintf(&sb, "export %s=%s\n", key, quoted)
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

// Reset clears only the env key, preserving other settings
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
