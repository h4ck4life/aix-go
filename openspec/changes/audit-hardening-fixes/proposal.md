## Why

A comprehensive audit of the aix-go codebase found 8 critical/high bugs, 6 medium bugs, 4 security concerns, and 12 functionality gaps. The most impactful issues are: token errors silently discarded (leaving users thinking an operation succeeded when it didn't), registry operations that silently succeed on nonexistent providers, destructive settings writes that can wipe unrelated Claude Code configuration, and a `--merge` flag that doesn't actually merge. These erode user trust and can cause data loss.

## What Changes

- **Fix silent token errors**: Replace `_ = tokenMgr.X(...)` with proper error checking in `provider add` (line 129), `provider remove` (line 241), and `provider rename` (line 274) in `cmd/provider.go`
- **Registry existence checks**: `RemoveOne` returns error when provider doesn't exist instead of silently succeeding
- **DeleteToken keychain error handling**: Return keychain deletion error even when file deletion succeeds (currently requires BOTH to fail)
- **Settings file preservation**: `Settings.Read()` preserves non-`env` keys from `~/.claude/settings.json`; `Reset()` only clears `env`, not the entire file
- **Config import merge**: `--merge` flag actually merges imported providers with existing ones instead of always overwriting
- **Doctor accuracy**: Network check validates HTTP status code; token storage check validates key file integrity
- **Model name validation**: `provider edit --model` and `provider set-model` validate model names
- **Map iteration determinism**: Sort map keys before iterating in `FormatForShell` and `config current`
- **Crypto key validation**: Verify key file is exactly 32 bytes after loading
- **Export file permissions**: Use 0600 instead of 0644 for exported config files
- **Atomic rename fallback**: Handle cross-device rename failures gracefully
- **Unit tests for cmd/**: Add test coverage for `cmd/provider.go`, `cmd/config.go`, `cmd/doctor.go` error paths and edge cases
- **Unit tests for crypto edge cases**: Corrupted ciphertext, truncated key, empty plaintext

## Capabilities

### New Capabilities
- `error-handling-safety`: All operations return accurate success/failure status to the user — no silent error swallowing
- `data-integrity`: Settings file and registry operations preserve data integrity — no destructive overwrites, no no-op successes
- `cmd-test-coverage`: Unit tests for all `cmd/` command handlers covering error paths and edge cases

### Modified Capabilities
- `error-consistency`: Extend to cover the 3 specific silent-error sites in `cmd/provider.go`
- `crypto-hardening`: Add key file length validation (32 bytes) and key integrity check after load
- `test-coverage`: Expand scope from core-only to include cmd/ test coverage and crypto edge cases
- `concurrent-safety`: Add MoveToken atomicity requirement (delete-then-set ordering)

## Impact

- **Files modified**: `cmd/provider.go`, `cmd/config.go`, `cmd/doctor.go`, `core/registry.go`, `core/token.go`, `core/settings.go`, `crypto/encrypt.go`
- **Files added**: `cmd/provider_test.go`, `cmd/config_test.go`, `cmd/doctor_test.go`, `crypto/encrypt_test.go` (expanded)
- **Breaking changes**: `RemoveOne` will now return an error for nonexistent providers — callers that relied on silent success will need to handle the error (only affects `runProviderRemove` which already checks `registry.Load()` error)
- **Dependencies**: None added
