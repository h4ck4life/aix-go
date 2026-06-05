## MODIFIED Requirements

### Requirement: All commands use custom error types for exit codes
Every command in `cmd/` SHALL return errors wrapped in types from the `utils` error hierarchy (`AixError`, `ValidationError`, `TokenError`, `FileNotFoundError`, `PermissionDeniedError`). Raw `error` values SHALL be wrapped using a generic wrapper that assigns `ExitGeneralError` (exit code 1). Token operations in `cmd/provider.go` SHALL NOT discard errors with `_ =` — all token storage, deletion, and move errors SHALL be propagated to the caller.

#### Scenario: Doctor check fails
- **WHEN** a doctor diagnostic check encounters an error
- **THEN** the error is wrapped in a custom error type and `main.go` returns the correct non-zero exit code

#### Scenario: Registry file is missing
- **WHEN** the registry file cannot be found during any command
- **THEN** a `FileNotFoundError` is returned, producing exit code 3

#### Scenario: Provider name is invalid
- **WHEN** a provider name fails validation
- **THEN** a `ValidationError` is returned, producing exit code 2

#### Scenario: Token storage fails during interactive provider add
- **WHEN** `tokenMgr.SetToken()` returns an error in `runProviderAdd`
- **THEN** a `TokenError` is returned and the success message is NOT printed

#### Scenario: Token deletion fails during provider remove
- **WHEN** `tokenMgr.DeleteToken()` returns an error in `runProviderRemove`
- **THEN** a warning is printed and the provider is still removed from the registry

#### Scenario: Token move fails during provider rename
- **WHEN** `tokenMgr.MoveToken()` returns an error in `runProviderRename`
- **THEN** a `TokenError` is returned and the registry rename is NOT performed
