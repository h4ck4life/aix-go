package constants

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	AppName    = "aix"
	AppVersion = "1.0.0"

	// Service name for keychain
	KeychainService = "aix"

	// Cache TTL in seconds
	RegistryCacheTTL = 5

	// Shell types
	ShellBash      = "bash"
	ShellZsh       = "zsh"
	ShellFish      = "fish"
	ShellPowerShell = "powershell"
	ShellCmd       = "cmd"
)

// Model alias keys
const (
	ModelOpus     = "opus"
	ModelSonnet   = "sonnet"
	ModelHaiku    = "haiku"
	ModelSubagent = "subagent"
)

// Environment variable names
const (
	EnvAnthropicBaseURL       = "ANTHROPIC_BASE_URL"
	EnvAnthropicAPIKey        = "ANTHROPIC_API_KEY"
	EnvAnthropicAuthToken     = "ANTHROPIC_AUTH_TOKEN"
	EnvAnthropicModel         = "ANTHROPIC_MODEL"
	EnvDefaultOpusModel       = "ANTHROPIC_DEFAULT_OPUS_MODEL"
	EnvDefaultSonnetModel     = "ANTHROPIC_DEFAULT_SONNET_MODEL"
	EnvDefaultHaikuModel      = "ANTHROPIC_DEFAULT_HAIKU_MODEL"
	EnvDefaultSubagentModel   = "CLAUDE_CODE_SUBAGENT_MODEL"
)

// Token types
const (
	TokenTypeAPIKey    = "ANTHROPIC_API_KEY"
	TokenTypeAuthToken = "ANTHROPIC_AUTH_TOKEN"
)

// Validation patterns
var (
	ProviderNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// Pre-configured providers
var PreconfiguredProviders = map[string]ProviderConfig{
	"ccpro": {
		BaseURL:   "https://cc.malif.dev/ccpro",
		TokenVar:  TokenTypeAPIKey,
	},
	"ccprotwo": {
		BaseURL:   "https://cc.malif.dev/ccprotwo",
		TokenVar:  TokenTypeAuthToken,
	},
	"kimi": {
		BaseURL:   "https://cc.malif.dev/moonshot",
		TokenVar:  TokenTypeAuthToken,
	},
	"minimax": {
		BaseURL:   "https://api.minimax.io/anthropic",
		TokenVar:  TokenTypeAuthToken,
	},
	"zai": {
		BaseURL:   "https://cc.malif.dev/zai",
		TokenVar:  TokenTypeAuthToken,
	},
}

// ProviderConfig represents a provider's configuration
type ProviderConfig struct {
	BaseURL       string            `json:"baseUrl"`
	TokenVar      string            `json:"tokenVar"`
	ModelName     string            `json:"modelName,omitempty"`
	DefaultModels map[string]string `json:"defaultModels,omitempty"`
}

// RegistryPath returns the path to the models.json file
func RegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".anthropic-switch", "models.json")
}

// TokenDir returns the path to the token storage directory
func TokenDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aix")
}

// TokenEncPath returns the path to the encrypted tokens file
func TokenEncPath() string {
	return filepath.Join(TokenDir(), "tokens.enc")
}

// TokenKeyPath returns the path to the encryption key file
func TokenKeyPath() string {
	return filepath.Join(TokenDir(), "key")
}

// SettingsPath returns the path to Claude's settings.json
func SettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

// DetectShell attempts to auto-detect the current shell
func DetectShell() string {
	// Check parent process name via SHELL env var
	if shell := os.Getenv("SHELL"); shell != "" {
		switch {
		case strings.Contains(shell, "fish"):
			return ShellFish
		case strings.Contains(shell, "zsh"):
			return ShellZsh
		case strings.Contains(shell, "bash"):
			return ShellBash
		}
	}

	// Check PSModulePath for PowerShell
	if os.Getenv("PSModulePath") != "" {
		return ShellPowerShell
	}

	// Check Windows-specific indicators
	if runtime.GOOS == "windows" {
		if os.Getenv("PROMPT") != "" {
			return ShellCmd
		}
		return ShellPowerShell
	}

	// Default fallback
	return ShellBash
}

// ModelAliasToEnvVar maps model alias to environment variable name
func ModelAliasToEnvVar(alias string) string {
	switch alias {
	case ModelOpus:
		return EnvDefaultOpusModel
	case ModelSonnet:
		return EnvDefaultSonnetModel
	case ModelHaiku:
		return EnvDefaultHaikuModel
	case ModelSubagent:
		return EnvDefaultSubagentModel
	default:
		return ""
	}
}
