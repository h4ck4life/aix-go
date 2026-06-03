package validation

import "testing"

func TestValidateProviderName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid single char", "a", false},
		{"valid multi char", "myprovider", false},
		{"valid with hyphen", "my-provider", false},
		{"valid with numbers", "abc123", false},
		{"valid complex", "my-provider-v2", false},
		{"empty", "", true},
		{"uppercase", "MyProvider", true},
		{"starts with number", "1provider", true},
		{"starts with hyphen", "-provider", true},
		{"special chars", "my_provider", true},
		{"spaces", "my provider", true},
		{"dots", "my.provider", true},
		{"slashes", "my/provider", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviderName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid https", "https://api.example.com", false},
		{"valid http", "http://localhost:8080", false},
		{"valid with path", "https://api.example.com/v1/chat", false},
		{"valid with query", "https://api.example.com?key=val", false},
		{"valid with port", "https://api.example.com:443", false},
		{"empty", "", true},
		{"no scheme", "api.example.com", true},
		{"ftp scheme", "ftp://files.example.com", true},
		{"no host", "https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeTokenVar(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"api-key", "ANTHROPIC_API_KEY"},
		{"API-KEY", "ANTHROPIC_API_KEY"},
		{"apikey", "ANTHROPIC_API_KEY"},
		{"ApiKey", "ANTHROPIC_API_KEY"},
		{"api_key", "ANTHROPIC_API_KEY"},
		{"auth-token", "ANTHROPIC_AUTH_TOKEN"},
		{"AUTH-TOKEN", "ANTHROPIC_AUTH_TOKEN"},
		{"authtoken", "ANTHROPIC_AUTH_TOKEN"},
		{"AuthToken", "ANTHROPIC_AUTH_TOKEN"},
		{"auth_token", "ANTHROPIC_AUTH_TOKEN"},
		{"ANTHROPIC_API_KEY", "ANTHROPIC_API_KEY"},
		{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_AUTH_TOKEN"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeTokenVar(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeTokenVar(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateTokenVar(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"api-key", "api-key", false},
		{"auth-token", "auth-token", false},
		{"full api key", "ANTHROPIC_API_KEY", false},
		{"full auth token", "ANTHROPIC_AUTH_TOKEN", false},
		{"invalid", "unknown-type", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenVar(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTokenVar(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateModelAlias(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"opus", "opus", false},
		{"sonnet", "sonnet", false},
		{"haiku", "haiku", false},
		{"subagent", "subagent", false},
		{"invalid", "unknown", true},
		{"empty", "", true},
		{"typo", "opuss", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelAlias(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModelAlias(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateShell(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"bash", "bash", false},
		{"zsh", "zsh", false},
		{"fish", "fish", false},
		{"powershell", "powershell", false},
		{"cmd", "cmd", false},
		{"invalid", "csh", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShell(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShell(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
