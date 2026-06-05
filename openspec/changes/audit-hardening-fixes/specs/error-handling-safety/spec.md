## ADDED Requirements

### Requirement: Interactive provider add checks token storage result
The `runProviderAdd` function SHALL check the error returned by `tokenMgr.SetToken()` and propagate it to the caller instead of discarding it with `_ =`.

#### Scenario: Token storage succeeds during interactive add
- **WHEN** a provider is added interactively with a valid token
- **THEN** the token is stored and the command prints "Provider added successfully"

#### Scenario: Token storage fails during interactive add
- **WHEN** `tokenMgr.SetToken()` returns an error (keychain unavailable, file permission denied)
- **THEN** the command returns the error and does NOT print "Provider added successfully"

### Requirement: Provider remove checks token deletion result
The `runProviderRemove` function SHALL check the error returned by `tokenMgr.DeleteToken()` and warn the user if token deletion fails, while still proceeding to remove the provider from the registry.

#### Scenario: Token deletion succeeds during provider remove
- **WHEN** a provider is removed and its token is successfully deleted
- **THEN** the command prints "Provider removed" with no warnings

#### Scenario: Token deletion fails during provider remove
- **WHEN** `tokenMgr.DeleteToken()` returns an error
- **THEN** the command still removes the provider from the registry but prints a warning about the orphaned token

### Requirement: Provider rename checks token move result
The `runProviderRename` function SHALL check the error returned by `tokenMgr.MoveToken()` and propagate it to the caller.

#### Scenario: Token move succeeds during rename
- **WHEN** a provider is renamed and the token is successfully moved
- **THEN** the command prints "Provider renamed" with no warnings

#### Scenario: Token move fails during rename
- **WHEN** `tokenMgr.MoveToken()` returns an error
- **THEN** the command returns the error and does NOT proceed with the registry rename (to avoid a broken renamed provider with no token)

### Requirement: DeleteToken returns keychain error even when file deletion succeeds
The `DeleteToken` method SHALL return the keychain deletion error when the keychain backend is active, even if the file fallback deletion succeeds. Only return nil when the primary storage backend succeeds.

#### Scenario: Keychain deletion fails, file deletion succeeds
- **WHEN** keychain is the primary backend and `keychain.Delete()` returns an error but the file copy is successfully deleted
- **THEN** `DeleteToken` returns the keychain error

#### Scenario: Both deletions succeed
- **WHEN** both keychain and file deletion succeed
- **THEN** `DeleteToken` returns nil

#### Scenario: Keychain is unavailable
- **WHEN** `useKeychain` is false and file deletion succeeds
- **THEN** `DeleteToken` returns nil (no keychain error expected)

### Requirement: Doctor network check validates HTTP status
The `checkNetwork` function in `cmd/doctor.go` SHALL check the HTTP response status code and report failure for any status code >= 400.

#### Scenario: API returns 200
- **WHEN** `https://api.anthropic.com` returns HTTP 200
- **THEN** check reports "OK"

#### Scenario: API returns 403 or 500
- **WHEN** `https://api.anthropic.com` returns HTTP 403, 500, or any code >= 400
- **THEN** check reports failure with the status code

#### Scenario: Network unreachable
- **WHEN** the HTTP request fails with a connection error
- **THEN** check reports failure with the error message (existing behavior, unchanged)

### Requirement: Model names are validated before storage
The `provider edit --model` flag and `provider set-model` command SHALL validate the model name is non-empty and contains only alphanumeric characters, hyphens, dots, and underscores before storing it.

#### Scenario: Valid model name
- **WHEN** `provider edit myprovider --model claude-sonnet-4-6` is executed
- **THEN** the model name is stored successfully

#### Scenario: Empty model name
- **WHEN** `provider edit myprovider --model ""` is executed
- **THEN** a ValidationError is returned indicating model name cannot be empty

#### Scenario: Model name with special characters
- **WHEN** `provider edit myprovider --model "my model with spaces"` is executed
- **THEN** a ValidationError is returned indicating invalid characters
