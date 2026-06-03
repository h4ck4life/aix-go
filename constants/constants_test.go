package constants

import (
	"os"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name      string
		shellEnv  string
		psModPath string
		prompt    string
		goos      string
		want      string
	}{
		{"zsh", "/bin/zsh", "", "", "darwin", ShellZsh},
		{"bash", "/bin/bash", "", "", "darwin", ShellBash},
		{"fish", "/usr/bin/fish", "", "", "darwin", ShellFish},
		{"powershell via PSModulePath", "", "/some/path", "", "darwin", ShellPowerShell},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			origShell := os.Getenv("SHELL")
			origPS := os.Getenv("PSModulePath")
			t.Cleanup(func() {
				os.Setenv("SHELL", origShell)
				os.Setenv("PSModulePath", origPS)
			})

			os.Setenv("SHELL", tt.shellEnv)
			os.Setenv("PSModulePath", tt.psModPath)

			got := DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModelAliasToEnvVar(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"opus", EnvDefaultOpusModel},
		{"sonnet", EnvDefaultSonnetModel},
		{"haiku", EnvDefaultHaikuModel},
		{"subagent", EnvDefaultSubagentModel},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			got := ModelAliasToEnvVar(tt.alias)
			if got != tt.want {
				t.Errorf("ModelAliasToEnvVar(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

func TestPathsAreConsistent(t *testing.T) {
	// RegistryPath should end with .anthropic-switch/models.json
	reg := RegistryPath()
	if len(reg) < len("/.anthropic-switch/models.json") {
		t.Errorf("RegistryPath too short: %s", reg)
	}

	// TokenDir should end with .aix
	dir := TokenDir()
	if len(dir) < len("/.aix") {
		t.Errorf("TokenDir too short: %s", dir)
	}

	// TokenEncPath should be inside TokenDir
	enc := TokenEncPath()
	if len(enc) <= len(dir) {
		t.Errorf("TokenEncPath should be under TokenDir")
	}

	// TokenKeyPath should be inside TokenDir
	key := TokenKeyPath()
	if len(key) <= len(dir) {
		t.Errorf("TokenKeyPath should be under TokenDir")
	}
}
