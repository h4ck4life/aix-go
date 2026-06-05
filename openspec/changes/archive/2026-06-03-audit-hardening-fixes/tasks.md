## 1. Concurrency Fixes

- [x] 1.1 Rewrite `setInFile` in `core/token.go` to use single read-modify-write cycle: read all lines, replace or append target entry, write atomically via temp-file-then-rename (matching `registry.saveLocked` pattern)
- [x] 1.2 Remove the `_ = tm.deleteFromFile(account)` call in `setInFile` and propagate errors from file operations
- [x] 1.3 Update `deleteFromFile` to also use atomic writes (temp-file-then-rename) instead of direct `os.WriteFile`
- [x] 1.4 Fix `IsAvailable` in `keychain/keychain.go` to use process-unique test account name: `aix-test-{pid}` using `os.Getpid()`

## 2. Crypto Hardening

- [x] 2.1 Replace PKCS7 padding validation in `crypto/encrypt.go` with constant-time check: iterate all padding bytes unconditionally, accumulate match result, then check once
- [x] 2.2 Replace specific error messages ("ciphertext too short", "invalid padding") with a single generic "decryption failed" error
- [x] 2.3 Add encrypt-decrypt round-trip test to `crypto/encrypt_test.go` (happy path, empty string, special characters)
- [x] 2.4 Add decryption failure tests: ciphertext too short, corrupt base64, invalid padding

## 3. Shell Escaping

- [x] 3.1 Fix `shellQuote` for CMD: escape metacharacters `& | < > ^ "` using `^` prefix, wrap values with spaces in double quotes
- [x] 3.2 Fix `shellQuote` for PowerShell: switch from backtick-escaped double quotes to single-quote wrapping with `''` escaping for embedded single quotes
- [x] 3.3 Add `core/settings_test.go` with table-driven tests for `shellQuote` across all 5 shells (bash, zsh, fish, powershell, cmd)
- [x] 3.4 Add `FormatForShell` tests verifying correct output syntax for each shell type
- [x] 3.5 Test round-trip: `GenerateEnvironmentVars` → `FormatForShell` produces eval-safe output for values containing special characters (`$`, `&`, `'`, `"`, spaces, backticks)

## 4. Error Consistency

- [x] 4.1 Add `WrapError(message string, err error) error` to `utils/errors.go` — wraps any error in `AixError` with `ExitGeneralError` and the original error as `Cause`
- [x] 4.2 Update `cmd/doctor.go` to return a wrapped error when diagnostic checks fail (currently always returns `nil`)
- [x] 4.3 Audit all `cmd/*.go` files for bare `error` returns that bypass the custom hierarchy and wrap them appropriately
- [x] 4.4 Add test for `WrapError` in `utils/errors_test.go` verifying exit code is 1 and original error is preserved as `Cause`

## 5. Test Coverage — Core

- [x] 5.1 Add `core/registry_test.go`: table-driven tests for `Load`, `GetAll`, `GetOne`, `SetOne`, `RemoveOne`, `RenameOne`, `SetModelName`, `SetDefaultModel`, `ClearCache` using temp directories
- [x] 5.2 Add cache TTL test: verify `ensureLoaded` uses cache within TTL and reloads after expiry
- [x] 5.3 Add `core/token_test.go`: tests for `GetToken`, `SetToken`, `DeleteToken`, `MoveToken`, `HasToken` with file-based storage in temp directories (keychain mocked via interface or skipped)
- [x] 5.4 Add `core/settings_test.go`: tests for `GenerateEnvironmentVars` (token exclusivity, model aliases, cleanup of removed aliases), `GetCurrentEnvironment`, `GetCurrentModel`, `GetCurrentProvider`

## 6. Test Coverage — Validation & Constants

- [x] 6.1 Add `validation/validation_test.go`: table-driven tests for `ValidateProviderName` (valid: `a`, `my-prov`, `abc123`; invalid: empty, uppercase, numbers-first, special chars, spaces)
- [x] 6.2 Add URL validation tests (valid: `https://api.example.com`; invalid: no scheme, no host, empty)
- [x] 6.3 Add token type normalization tests (`api-key` → `ANTHROPIC_API_KEY`, `auth-token` → `ANTHROPIC_AUTH_TOKEN`)
- [x] 6.4 Add model alias validation tests (valid: opus/sonnet/haiku/subagent; invalid: unknown)
- [x] 6.5 Add `constants/constants_test.go`: tests for `DetectShell` (mock `SHELL` env var), `ModelAliasToEnvVar`, path functions

## 7. Cleanup

- [x] 7.1 Add `dist/` to `.gitignore`
- [x] 7.2 Run `make fmt && make lint && go test ./...` and fix any issues
