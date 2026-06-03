## Why

A multi-agent audit of the entire codebase uncovered 30 findings across correctness, consistency, security, and test coverage dimensions. The most critical are: race conditions in token file writes and registry map operations, a padding oracle vulnerability in AES-256-CBC decryption, zero shell escaping for Windows CMD, and near-zero test coverage (14 of 15 packages untested). These issues risk data corruption, token leakage, and command injection. Fixing them now — before the tool sees wider use — is far cheaper than after.

## What Changes

- **Fix race conditions** in `core/token.go` (setInFile delete-then-append is not atomic) and `core/registry.go` (RenameOne check-then-delete under concurrent access)
- **Fix padding oracle** in `crypto/encrypt.go` by using constant-time padding validation
- **Fix CMD shell escaping** in `core/settings.go` — currently a no-op that allows command injection on Windows
- **Fix PowerShell shell escaping** — backtick-based escaping is incomplete for special characters
- **Fix error swallowing** — `deleteFromFile` errors ignored in `setInFile`, `doctor.go` returns raw errors bypassing the custom error hierarchy
- **Add consistent error wrapping** across all commands so exit codes are meaningful
- **Fix keychain availability check race** — fixed test account name allows concurrent processes to clobber each other
- **Add test coverage** for all core packages: `core/`, `crypto/`, `validation/`, `constants/`, with table-driven tests and mock interfaces
- **Fix `dist/` not in `.gitignore`** — build artifacts are showing in git status

## Capabilities

### New Capabilities
- `concurrent-safety`: Mutex and atomicity fixes for token file operations and registry map mutations
- `crypto-hardening`: Padding oracle fix and constant-time validation in AES-256-CBC encryption
- `shell-escaping`: Complete escaping for all 5 supported shells (bash, zsh, fish, powershell, cmd)
- `error-consistency`: Uniform error wrapping across all commands using the custom error hierarchy
- `test-coverage`: Unit tests for core, crypto, validation, constants, and settings packages

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **core/token.go**: Rewrite `setInFile` for atomicity, fix `MoveToken` fallback path
- **core/registry.go**: Harden `RenameOne` against concurrent access
- **core/settings.go**: Fix `shellQuote` for CMD and PowerShell
- **crypto/encrypt.go**: Fix padding validation to be constant-time
- **keychain/keychain.go**: Use unique test account names in `IsAvailable`
- **cmd/doctor.go**: Wrap errors with custom error types
- **cmd/*.go**: Audit all commands for consistent error wrapping
- **New test files**: `core/registry_test.go`, `core/token_test.go`, `core/settings_test.go`, `crypto/encrypt_test.go`, `validation/validation_test.go`, `constants/constants_test.go`
- **`.gitignore`**: Add `dist/` directory
- No breaking changes to CLI interface or config file format
