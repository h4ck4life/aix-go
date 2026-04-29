# aix

A lightweight CLI tool for switching between Anthropic-compatible API providers and tokens for Claude Code. This is the Go rewrite of the original Node.js aix CLI.

## Installation

### via go install (Recommended)

Install directly from source. Make sure `$GOPATH/bin` or `$GOBIN` is in your `$PATH`.

```bash
# Install latest release
go install github.com/h4ck4life/aix-go@latest

# Verify installation
which aix
aix --version
```

**macOS / Linux system-wide** (requires sudo):

```bash
# Build and copy to /usr/local/bin
go build -o /usr/local/bin/aix .

# Or use the Makefile
make install
```

**Windows system-wide** (Administrator PowerShell):

```powershell
# Build for Windows
go build -o aix.exe .

# Move to a directory in your PATH
Move-Item aix.exe C:\Windows\System32\aix.exe

# Or to a user-local bin directory
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item aix.exe "$env:USERPROFILE\bin\aix.exe"

# Add to PATH (if using user-local bin)
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:USERPROFILE\bin",
    "User"
)

# Verify
aix --version
```

### via Homebrew (macOS/Linux)

```bash
brew tap h4ck4life/aix
brew install aix
```

### via Scoop (Windows)

```powershell
# Add the bucket
scoop bucket add aix https://github.com/h4ck4life/scoop-bucket.git

# Install
scoop install aix

# Update later
scoop update aix
```

### Pre-built binaries

Download the latest release for your platform from the [releases page](https://github.com/h4ck4life/aix-go/releases).

## Features

- **Provider Management**: Add, list, remove, rename, and switch between API providers
- **Token Management**: Securely store and test API tokens via OS keychain or encrypted file fallback
- **Model Aliases**: Set default models for opus, sonnet, haiku, and subagent
- **Interactive Wizards**: Bubble Tea-powered prompts for first-time setup and provider configuration
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Shell Auto-Detection**: Automatically detects bash, zsh, fish, PowerShell, and CMD
- **Zero Config Modification**: Sets env vars in your current shell session — no files touched
- **Optional Persistence**: Save active provider to `~/.claude/settings.json` with `--persist`
- **Provider Editing**: Modify existing providers without remove+re-add

## Usage

### Provider Commands

```bash
# List all providers
aix provider list
aix p ls

# Add a provider
aix provider add my-provider https://api.example.com -t api-key
aix provider add -i  # Interactive mode

# Remove a provider
aix provider remove my-provider
aix provider rm my-provider --yes

# Rename a provider
aix provider rename old-name new-name

# Set active provider (outputs shell exports)
eval $(aix provider use my-provider)

# Persist to ~/.claude/settings.json
aix provider use my-provider --persist

# Explicit shell syntax
eval $(aix provider use my-provider --shell zsh)
eval $(aix provider use my-provider --shell fish)

# PowerShell
Invoke-Expression (aix provider use my-provider --shell powershell | Out-String)

# CMD
for /f "tokens=*" %a in ('aix provider use my-provider --shell cmd') do @%a

# Edit an existing provider
aix provider edit my-provider --url https://api.new.com
aix provider edit my-provider --token-type auth-token
aix provider edit my-provider --model claude-sonnet-4-6
aix provider edit my-provider --default-model sonnet=claude-sonnet-4-6

# Set custom model
aix provider set-model my-provider claude-sonnet-4-6

# Set default model alias
aix provider set-default my-provider opus claude-opus-4-7
```

### Token Commands

```bash
# Set token for a provider
aix token set my-provider
aix token set my-provider -t sk-...

# Test token
aix token test my-provider

# Remove token
aix token remove my-provider
aix token rm my-provider --yes
```

### Config Commands

```bash
# Show current configuration
aix config current

# Export configuration
aix config export -o aix-backup.json

# Import configuration
aix config import aix-backup.json --merge
```

### Other Commands

```bash
# First-time setup wizard
aix init

# Run diagnostics
aix doctor

# Print version
aix version
aix --version
```

## How It Works

By default, `aix provider use <name>` outputs `export` commands to stdout:

1. You wrap it with `eval $(...)` to set env vars in your current shell session
2. Claude Code picks up the env vars when run in the same session

This is cleaner, more transparent, and follows the Unix philosophy.

To persist the active provider to `~/.claude/settings.json` (so Claude Code uses it automatically):

```bash
aix provider use my-provider --persist
```

### Shell Integration (Optional)

Add this to your `.zshrc` or `.bashrc` for seamless switching:

```bash
aix-use() {
    eval $(aix provider use "$1")
}
```

Then use `aix-use ccpro` instead of `eval $(aix provider use ccpro)`.

## Configuration Files

aix reads provider configurations from:

- `~/.anthropic-switch/models.json` — Provider registry (shared with Node.js version)
- `~/.aix/tokens.enc` + `~/.aix/key` — Encrypted token fallback (shared with Node.js version)

Tokens are stored in your OS-native keychain by default.

## Development

### Prerequisites

- Go 1.22+
- Make

### Build

```bash
# Build for current platform
make build

# Run in development mode
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint
make lint

# Install locally
make install
```

### Cross-Platform Build

Build binaries for all supported platforms (macOS, Linux, Windows on amd64 and arm64):

```bash
make build-all
```

This generates:

```
build/
├── aix-darwin-amd64
├── aix-darwin-arm64
├── aix-linux-amd64
├── aix-linux-arm64
├── aix-windows-amd64.exe
└── aix-windows-arm64.exe
```

You can also build for a specific platform manually:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o aix-linux-amd64 .

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -o aix-windows-arm64.exe .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o aix-darwin-arm64 .
```

### Release with GoReleaser

Releases are fully automated via GoReleaser. See [RELEASING.md](RELEASING.md) for the complete workflow.

Quick release:

```bash
# Install GoReleaser (once)
brew install goreleaser

# Tag and release
make tag VERSION=v1.0.1
make release VERSION=v1.0.1
```

This builds all 6 platforms, creates a GitHub release, and updates the Homebrew tap automatically.

## Token Storage

aix uses your OS-native keychain as the primary storage for tokens:

- **macOS**: Keychain Access (Security.framework)
- **Linux**: Secret Service / libsecret (via D-Bus)
- **Windows**: Credential Manager

If the keychain is unavailable (e.g., headless Linux), aix falls back to AES-256-CBC encrypted file storage at `~/.aix/tokens.enc`.

## License

MIT
