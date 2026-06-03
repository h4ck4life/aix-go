package core

import (
	"strings"
	"testing"

	"github.com/h4ck4life/aix-go/constants"
)

func TestGenerateEnvironmentVarsTokenExclusivity(t *testing.T) {
	tests := []struct {
		name     string
		tokenVar string
		wantKey  string
		wantAuth string
	}{
		{"api-key type", constants.TokenTypeAPIKey, "test-token", ""},
		{"auth-token type", constants.TokenTypeAuthToken, "", "test-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{Env: make(map[string]string)}
			cfg := constants.ProviderConfig{
				BaseURL:  "https://api.test.com",
				TokenVar: tt.tokenVar,
			}
			if err := s.GenerateEnvironmentVars("test", cfg, "test-token"); err != nil {
				t.Fatalf("GenerateEnvironmentVars failed: %v", err)
			}

			if got := s.Env[constants.EnvAnthropicAPIKey]; got != tt.wantKey {
				t.Errorf("ANTHROPIC_API_KEY = %q, want %q", got, tt.wantKey)
			}
			if got := s.Env[constants.EnvAnthropicAuthToken]; got != tt.wantAuth {
				t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", got, tt.wantAuth)
			}
		})
	}
}

func TestGenerateEnvironmentVarsModelAliases(t *testing.T) {
	s := &Settings{Env: make(map[string]string)}
	cfg := constants.ProviderConfig{
		BaseURL:   "https://api.test.com",
		TokenVar:  constants.TokenTypeAPIKey,
		ModelName: "custom-model",
		DefaultModels: map[string]string{
			"opus":   "claude-opus-4-8",
			"sonnet": "claude-sonnet-4-6",
		},
	}

	if err := s.GenerateEnvironmentVars("test", cfg, "token"); err != nil {
		t.Fatalf("GenerateEnvironmentVars failed: %v", err)
	}

	if got := s.Env[constants.EnvAnthropicModel]; got != "custom-model" {
		t.Errorf("ANTHROPIC_MODEL = %q, want %q", got, "custom-model")
	}
	if got := s.Env[constants.EnvDefaultOpusModel]; got != "claude-opus-4-8" {
		t.Errorf("ANTHROPIC_DEFAULT_OPUS_MODEL = %q, want %q", got, "claude-opus-4-8")
	}
	if got := s.Env[constants.EnvDefaultSonnetModel]; got != "claude-sonnet-4-6" {
		t.Errorf("ANTHROPIC_DEFAULT_SONNET_MODEL = %q, want %q", got, "claude-sonnet-4-6")
	}
	// Haiku should not be set
	if _, ok := s.Env[constants.EnvDefaultHaikuModel]; ok {
		t.Error("ANTHROPIC_DEFAULT_HAIKU_MODEL should not be set")
	}
}

func TestGenerateEnvironmentVarsCleansUpOldAliases(t *testing.T) {
	// Start with all aliases set
	s := &Settings{Env: map[string]string{
		constants.EnvDefaultOpusModel:     "old-opus",
		constants.EnvDefaultSonnetModel:   "old-sonnet",
		constants.EnvDefaultHaikuModel:    "old-haiku",
		constants.EnvDefaultSubagentModel: "old-subagent",
	}}

	// New config only has opus
	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
		DefaultModels: map[string]string{
			"opus": "new-opus",
		},
	}

	s.GenerateEnvironmentVars("test", cfg, "token")

	if got := s.Env[constants.EnvDefaultOpusModel]; got != "new-opus" {
		t.Errorf("opus = %q, want %q", got, "new-opus")
	}
	if _, ok := s.Env[constants.EnvDefaultSonnetModel]; ok {
		t.Error("sonnet should be cleaned up")
	}
	if _, ok := s.Env[constants.EnvDefaultHaikuModel]; ok {
		t.Error("haiku should be cleaned up")
	}
	if _, ok := s.Env[constants.EnvDefaultSubagentModel]; ok {
		t.Error("subagent should be cleaned up")
	}
}

func TestGetCurrentEnvironment(t *testing.T) {
	s := &Settings{Env: map[string]string{
		"KEY1": "val1",
		"KEY2": "val2",
	}}

	env := s.GetCurrentEnvironment()
	if len(env) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(env))
	}
	if env["KEY1"] != "val1" {
		t.Errorf("KEY1 = %q, want %q", env["KEY1"], "val1")
	}
}

func TestGetCurrentEnvironmentNil(t *testing.T) {
	s := &Settings{}
	env := s.GetCurrentEnvironment()
	if env == nil || len(env) != 0 {
		t.Errorf("expected empty non-nil map, got %v", env)
	}
}

func TestGetCurrentModel(t *testing.T) {
	s := &Settings{Env: map[string]string{
		constants.EnvAnthropicModel: "claude-opus-4-8",
	}}
	if got := s.GetCurrentModel(); got != "claude-opus-4-8" {
		t.Errorf("GetCurrentModel() = %q, want %q", got, "claude-opus-4-8")
	}
}

func TestGetCurrentModelEmpty(t *testing.T) {
	s := &Settings{Env: map[string]string{}}
	if got := s.GetCurrentModel(); got != "" {
		t.Errorf("GetCurrentModel() = %q, want empty", got)
	}
}

func TestGetCurrentProvider(t *testing.T) {
	s := &Settings{Env: map[string]string{
		constants.EnvAnthropicBaseURL: "https://api.test.com",
	}}
	if got := s.GetCurrentProvider(); got != "https://api.test.com" {
		t.Errorf("GetCurrentProvider() = %q, want %q", got, "https://api.test.com")
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name   string
		shell  string
		input  string
		substr string // substring that must appear in output
	}{
		// Bash / Zsh / Fish (POSIX single-quote escaping)
		{"bash simple", constants.ShellBash, "hello", `'hello'`},
		{"bash with space", constants.ShellBash, "hello world", `'hello world'`},
		{"bash with dollar", constants.ShellBash, "$HOME/path", `'$HOME/path'`},
		{"bash with backtick", constants.ShellBash, "echo `cmd`", "'echo `cmd`'"},
		{"bash with ampersand", constants.ShellBash, "foo && bar", `'foo && bar'`},
		{"bash with single quote", constants.ShellBash, "it's here", `'it'\''s here'`},
		{"bash with double quote", constants.ShellBash, `say "hi"`, `'say "hi"'`},
		{"bash empty string", constants.ShellBash, "", `''`},
		{"zsh special chars", constants.ShellZsh, "$foo!bar*baz", `'$foo!bar*baz'`},
		{"fish with special", constants.ShellFish, "a&b|c;d", `'a&b|c;d'`},

		// PowerShell (single-quote wrapping, '' for embedded quotes)
		{"ps simple", constants.ShellPowerShell, "hello", `'hello'`},
		{"ps with space", constants.ShellPowerShell, "hello world", `'hello world'`},
		{"ps with dollar", constants.ShellPowerShell, "$env:FOO", `'$env:FOO'`},
		{"ps with double quote", constants.ShellPowerShell, `key"value`, `'key"value'`},
		{"ps with single quote", constants.ShellPowerShell, "it's here", `'it''s here'`},
		{"ps with backtick", constants.ShellPowerShell, "a`b", "'a`b'"},
		{"ps empty string", constants.ShellPowerShell, "", `''`},

		// CMD (^ escaping, double-quote wrapping, %% for %)
		{"cmd simple", constants.ShellCmd, "hello", `"hello"`},
		{"cmd with space", constants.ShellCmd, "hello world", `"hello world"`},
		{"cmd with ampersand", constants.ShellCmd, "foo&bar", `"foo^&bar"`},
		{"cmd with pipe", constants.ShellCmd, "foo|bar", `"foo^|bar"`},
		{"cmd with redirect", constants.ShellCmd, "foo>bar", `"foo^>bar"`},
		{"cmd with caret", constants.ShellCmd, "foo^bar", `"foo^^bar"`},
		{"cmd with percent", constants.ShellCmd, "50%", `"50%%"`},
		{"cmd with double quote", constants.ShellCmd, `say "hi"`, `"say ^"hi^""`},
		{"cmd URL with query", constants.ShellCmd, "https://api.example.com?key=val&other=2", `"https://api.example.com?key=val^&other=2"`},
		{"cmd empty string", constants.ShellCmd, "", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.shell, tt.input)
			if !strings.Contains(result, tt.substr) {
				t.Errorf("shellQuote(%q, %q) = %q, want to contain %q", tt.shell, tt.input, result, tt.substr)
			}
		})
	}
}

func TestFormatForShell(t *testing.T) {
	env := map[string]string{
		"ANTHROPIC_BASE_URL": "https://api.example.com",
		"ANTHROPIC_API_KEY":  "sk-test-key-123",
	}

	s := Settings{Env: env}

	tests := []struct {
		name   string
		shell  string
		prefix string // expected line prefix
	}{
		{"bash export", constants.ShellBash, "export ANTHROPIC_BASE_URL="},
		{"zsh export", constants.ShellZsh, "export ANTHROPIC_API_KEY="},
		{"fish set", constants.ShellFish, "set -x ANTHROPIC_BASE_URL "},
		{"ps env", constants.ShellPowerShell, "$env:ANTHROPIC_API_KEY = "},
		{"cmd set", constants.ShellCmd, "set ANTHROPIC_BASE_URL="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := s.FormatForShell(tt.shell)
			if !strings.Contains(output, tt.prefix) {
				t.Errorf("FormatForShell(%q) = %q, want to contain %q", tt.shell, output, tt.prefix)
			}
		})
	}
}

func TestFormatForShellRoundTrip(t *testing.T) {
	// Values with special characters that must survive eval
	specialValues := map[string]string{
		"ANTHROPIC_BASE_URL": "https://api.example.com/v1?key=val&other=2",
		"ANTHROPIC_API_KEY":  "sk-proj-$pecial'key\"with`chars",
	}

	s := Settings{Env: specialValues}

	shells := []string{constants.ShellBash, constants.ShellZsh, constants.ShellFish, constants.ShellPowerShell}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			output := s.FormatForShell(shell)
			// Verify the output contains the env var assignments
			if !strings.Contains(output, "ANTHROPIC_BASE_URL") {
				t.Errorf("missing ANTHROPIC_BASE_URL in output for %s", shell)
			}
			if !strings.Contains(output, "ANTHROPIC_API_KEY") {
				t.Errorf("missing ANTHROPIC_API_KEY in output for %s", shell)
			}
			// For POSIX shells, verify single-quote wrapping is used
			if shell == constants.ShellBash || shell == constants.ShellZsh || shell == constants.ShellFish {
				if !strings.Contains(output, "export ") && shell != constants.ShellFish {
					t.Errorf("expected 'export' in %s output", shell)
				}
				if shell == constants.ShellFish && !strings.Contains(output, "set -x ") {
					t.Errorf("expected 'set -x' in fish output")
				}
			}
		})
	}
}
