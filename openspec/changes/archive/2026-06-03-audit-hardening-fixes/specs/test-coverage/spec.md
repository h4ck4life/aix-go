## ADDED Requirements

### Requirement: Core package has unit tests
The `core/` package SHALL have test files covering `registry.go`, `token.go`, and `settings.go` with table-driven tests for all public methods.

#### Scenario: Registry CRUD operations
- **WHEN** `go test ./core/` is run
- **THEN** tests cover `Load`, `GetAll`, `GetOne`, `SetOne`, `RemoveOne`, `RenameOne`, `SetModelName`, `SetDefaultModel`, `ClearCache`, and cache TTL behavior

#### Scenario: Token manager storage operations
- **WHEN** `go test ./core/` is run
- **THEN** tests cover `GetToken`, `SetToken`, `DeleteToken`, `MoveToken`, `HasToken`, `GetStorageInfo` for both keychain and file fallback paths

#### Scenario: Settings environment generation
- **WHEN** `go test ./core/` is run
- **THEN** tests cover `GenerateEnvironmentVars` (token exclusivity, model aliases, cleanup), `FormatForShell` for all 5 shells, `GetCurrentEnvironment`, `GetCurrentModel`, `GetCurrentProvider`

### Requirement: Crypto package has round-trip tests
The `crypto/` package SHALL have tests verifying encrypt-decrypt round-trips, padding edge cases, and error handling.

#### Scenario: Encrypt-decrypt round-trip
- **WHEN** a plaintext string is encrypted and then decrypted
- **THEN** the decrypted value matches the original plaintext exactly

#### Scenario: Invalid ciphertext handling
- **WHEN** an invalid or corrupt ciphertext is passed to `Decrypt`
- **THEN** the function returns a "decryption failed" error without panicking

#### Scenario: Empty plaintext encryption
- **WHEN** an empty string is encrypted and decrypted
- **THEN** the round-trip produces an empty string

### Requirement: Validation package has edge-case tests
The `validation/` package SHALL have table-driven tests covering valid and invalid inputs for provider names, URLs, token types, and model aliases.

#### Scenario: Provider name validation
- **WHEN** `go test ./validation/` is run
- **THEN** tests cover valid names (`a`, `my-provider`, `abc123`), invalid names (empty, uppercase, starts with number, special characters, spaces)

#### Scenario: URL validation
- **WHEN** `go test ./validation/` is run
- **THEN** tests cover valid URLs (`https://api.example.com`), invalid URLs (no scheme, no host, empty string)

#### Scenario: Token type normalization
- **WHEN** `go test ./validation/` is run
- **THEN** tests cover short forms (`api-key`, `auth-token`) expanding to full forms

### Requirement: Constants package has shell detection tests
The `constants/` package SHALL have tests for `DetectShell` and `ModelAliasToEnvVar`.

#### Scenario: Shell detection from SHELL env var
- **WHEN** `SHELL=/bin/zsh` is set
- **THEN** `DetectShell()` returns `ShellZsh`

#### Scenario: Model alias mapping
- **WHEN** `ModelAliasToEnvVar("opus")` is called
- **THEN** it returns `EnvDefaultOpusModel`

### Requirement: Test infrastructure uses temp directories
All file-based tests SHALL use `t.TempDir()` or equivalent to avoid writing to the user's real config directory. Tests SHALL NOT require a real keychain to pass.

#### Scenario: Token manager tests in temp directory
- **WHEN** token file operations are tested
- **THEN** they operate on files in a temporary directory that is cleaned up after the test

#### Scenario: Registry tests in temp directory
- **WHEN** registry file operations are tested
- **THEN** they operate on a temporary `models.json` that is cleaned up after the test
