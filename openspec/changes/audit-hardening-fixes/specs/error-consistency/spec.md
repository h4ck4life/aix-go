## ADDED Requirements

### Requirement: All commands use custom error types for exit codes
Every command in `cmd/` SHALL return errors wrapped in types from the `utils` error hierarchy (`AixError`, `ValidationError`, `TokenError`, `FileNotFoundError`, `PermissionDeniedError`). Raw `error` values SHALL be wrapped using a generic wrapper that assigns `ExitGeneralError` (exit code 1).

#### Scenario: Doctor check fails
- **WHEN** a doctor diagnostic check encounters an error
- **THEN** the error is wrapped in a custom error type and `main.go` returns the correct non-zero exit code

#### Scenario: Registry file is missing
- **WHEN** the registry file cannot be found during any command
- **THEN** a `FileNotFoundError` is returned, producing exit code 3

#### Scenario: Provider name is invalid
- **WHEN** a provider name fails validation
- **THEN** a `ValidationError` is returned, producing exit code 2

### Requirement: Generic error wrapper available
A `utils.WrapError(message string, err error) error` function SHALL be available for wrapping arbitrary errors that don't fit a specific error category. The wrapper SHALL produce an `AixError` with `ExitGeneralError` and include the original error as the `Cause`.

#### Scenario: Wrapping a network error in doctor
- **WHEN** an HTTP request fails during `checkNetwork`
- **THEN** the error is wrapped with `WrapError` and has exit code 1

#### Scenario: Wrapping a filesystem error
- **WHEN** an `os.ReadFile` call fails with a permission error
- **THEN** the error is wrapped with `WrapError` and has exit code 1

### Requirement: Doctor command returns non-zero on failure
The `runDoctor` function SHALL return an error (not `nil`) when any diagnostic check fails. Currently it always returns `nil`.

#### Scenario: All checks pass
- **WHEN** all doctor checks succeed
- **THEN** the command exits with code 0

#### Scenario: One or more checks fail
- **WHEN** any doctor check fails
- **THEN** the command returns a wrapped error and exits with code 1
