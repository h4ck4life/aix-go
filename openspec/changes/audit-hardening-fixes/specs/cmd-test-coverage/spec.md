## ADDED Requirements

### Requirement: cmd/provider.go has unit tests
The `cmd/provider.go` file SHALL have a corresponding `cmd/provider_test.go` with tests for all `RunE` functions covering both happy paths and error paths.

#### Scenario: provider add happy path
- **WHEN** `runProviderAdd` is called with valid name and URL args
- **THEN** the provider is stored in the registry and no error is returned

#### Scenario: provider add with invalid name
- **WHEN** `runProviderAdd` is called with an invalid provider name (uppercase, special chars)
- **THEN** a `ValidationError` is returned

#### Scenario: provider add missing args
- **WHEN** `runProviderAdd` is called with fewer than 2 non-interactive args
- **THEN** a `ValidationError` is returned with usage message

#### Scenario: provider remove nonexistent
- **WHEN** `runProviderRemove` is called for a provider that doesn't exist
- **THEN** a `ValidationError` is returned (after RemoveOne fix)

#### Scenario: provider rename nonexistent
- **WHEN** `runProviderRename` is called for a provider that doesn't exist
- **THEN** a `ValidationError` is returned

#### Scenario: provider use nonexistent
- **WHEN** `runProviderUse` is called for a provider that doesn't exist
- **THEN** a `ValidationError` is returned

#### Scenario: provider use with no token
- **WHEN** `runProviderUse` is called for a provider that exists but has no stored token
- **THEN** a `TokenError` is returned

#### Scenario: provider list empty registry
- **WHEN** `runProviderList` is called and the registry is empty
- **THEN** "No providers configured" is printed and no error is returned

#### Scenario: provider edit with no changes
- **WHEN** `runProviderEdit` is called with no flags set
- **THEN** a `ValidationError` is returned indicating no changes specified

### Requirement: cmd/config.go has unit tests
The `cmd/config.go` file SHALL have a corresponding `cmd/config_test.go` with tests for export, import, and import-merge flows.

#### Scenario: config export roundtrip
- **WHEN** providers are exported to a file then imported from that file
- **THEN** the imported providers match the original providers exactly

#### Scenario: config import invalid file
- **WHEN** `runConfigImport` is called with a file containing invalid JSON
- **THEN** a `ValidationError` is returned

#### Scenario: config import merge
- **WHEN** `runConfigImport` with `--merge` imports a file with one new provider and one existing provider
- **THEN** the new provider is added and the existing provider is updated (merged), not replaced

#### Scenario: config import nonexistent file
- **WHEN** `runConfigImport` is called with a path that doesn't exist
- **THEN** a `FileNotFoundError` is returned

### Requirement: cmd/doctor.go has unit tests
The `cmd/doctor.go` file SHALL have a corresponding `cmd/doctor_test.go` with tests for each check function.

#### Scenario: checkRegistry with valid registry
- **WHEN** `checkRegistry()` is called and the registry file exists with valid JSON
- **THEN** returns true with provider count

#### Scenario: checkNetwork returns failure on 500
- **WHEN** `checkNetwork()` receives an HTTP 500 response
- **THEN** returns false with status code information

#### Scenario: checkPermissions with writable directories
- **WHEN** `checkPermissions()` is called and all directories are writable
- **THEN** returns true

### Requirement: Tests use temp directories for file operations
All cmd/ tests SHALL use `t.TempDir()` to isolate file-based operations. Tests SHALL override path constants or inject paths to avoid writing to the user's real config directory.

#### Scenario: Test isolation
- **WHEN** cmd/ tests are run
- **THEN** no files are created or modified outside the test's temporary directory

#### Scenario: Tests run in parallel safely
- **WHEN** multiple cmd/ tests run concurrently (t.Parallel())
- **THEN** each test operates on its own isolated temp directory without interference

### Requirement: Crypto tests cover key validation
The crypto test suite SHALL include tests for the key file validation added in this change.

#### Scenario: Truncated key file
- **WHEN** a key file with fewer than 32 bytes exists
- **THEN** `loadOrCreateKey` returns an error and the error message includes the actual byte count

#### Scenario: Valid key file
- **WHEN** a key file with exactly 32 bytes exists
- **THEN** `loadOrCreateKey` returns the key bytes without error
