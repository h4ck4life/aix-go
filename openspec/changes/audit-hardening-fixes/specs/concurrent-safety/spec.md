## ADDED Requirements

### Requirement: Token file writes are atomic
The `setInFile` method SHALL perform a single read-modify-write cycle under the existing `sync.RWMutex`. It SHALL read all lines from the encrypted token file, replace or append the target entry, and write the complete file atomically using a temp-file-then-rename pattern. The method SHALL NOT use a separate delete-then-append sequence.

#### Scenario: New token stored when file does not exist
- **WHEN** `setInFile` is called for a new account and the encrypted file does not exist
- **THEN** a new file is created with the single encrypted entry, written via temp-file-then-rename

#### Scenario: Existing token updated atomically
- **WHEN** `setInFile` is called for an account that already has a token in the file
- **THEN** the file is rewritten with the old entry replaced by the new encrypted token, with no intermediate state where neither old nor new token exists

#### Scenario: Process crash during write does not corrupt file
- **WHEN** the process crashes during `setInFile`
- **THEN** the original file remains intact (the temp file is abandoned, original is untouched until rename)

### Requirement: Token file delete errors are not silently swallowed
The `setInFile` method SHALL propagate errors from file operations instead of ignoring them with `_ =`.

#### Scenario: File read failure during token update
- **WHEN** `setInFile` cannot read the existing token file
- **THEN** the error is returned to the caller, not silently ignored

### Requirement: Keychain availability check uses unique account names
The `IsAvailable` function in the keychain package SHALL use a process-unique test account name that includes the process ID, preventing interference between concurrent `aix` invocations.

#### Scenario: Two concurrent processes check keychain availability
- **WHEN** two `aix` processes call `IsAvailable` simultaneously
- **THEN** each uses a distinct test account name and neither interferes with the other's test token

### Requirement: Registry RenameOne is safe (verified)
The `RenameOne` method performs its check-delete-set sequence under a write lock (`r.mu.Lock()`). This is confirmed safe — no code change required.

#### Scenario: Concurrent rename operations on different providers
- **WHEN** two goroutines call `RenameOne` with different old names simultaneously
- **THEN** both operations complete correctly, serialized by the write lock
