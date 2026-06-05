## ADDED Requirements

### Requirement: RemoveOne rejects nonexistent providers
The `RemoveOne` method in `core/registry.go` SHALL check whether the provider exists in the map before calling `delete()`. If the provider does not exist, it SHALL return a `ValidationError` with the message "provider 'X' not found".

#### Scenario: Remove an existing provider
- **WHEN** `RemoveOne` is called with a provider name that exists in the registry
- **THEN** the provider is removed from the map, the file is saved, and nil is returned

#### Scenario: Remove a nonexistent provider
- **WHEN** `RemoveOne` is called with a provider name that does not exist in the registry
- **THEN** a `ValidationError` is returned and the registry file is NOT rewritten

### Requirement: Settings file preserves non-env keys
The `Settings.Read()` method SHALL preserve all keys from `~/.claude/settings.json`, not just `env`. The `Settings.Write()` method SHALL merge the `env` field back into the preserved data and write the complete object.

#### Scenario: Settings file contains permissions key
- **WHEN** `~/.claude/settings.json` contains `{"env": {...}, "permissions": {...}}`
- **THEN** after `Read()` then `Write()`, the `permissions` key is preserved unchanged

#### Scenario: Settings file is empty or missing
- **WHEN** `~/.claude/settings.json` does not exist
- **THEN** `Read()` initializes `Env` to an empty map; `Write()` creates the file with `{"env":{}}`

#### Scenario: Reset clears only env
- **WHEN** `Reset()` is called on settings that contain both `env` and other keys
- **THEN** only the `env` key is cleared; other keys are preserved

### Requirement: Config import merge actually merges
When the `--merge` flag is true, `runConfigImport` SHALL merge imported providers with existing providers. For providers that exist in both current and imported configs, imported fields SHALL overwrite existing fields. Fields not present in the imported config SHALL preserve their existing values.

#### Scenario: Merge with new provider
- **WHEN** `config import --merge` is run and the imported file contains a provider not in the current registry
- **THEN** the new provider is added to the registry

#### Scenario: Merge with existing provider
- **WHEN** `config import --merge` is run and the imported file contains a provider that already exists
- **THEN** only the fields specified in the imported config are overwritten; unspecified fields preserve existing values

#### Scenario: Import without merge flag
- **WHEN** `config import` is run without `--merge`
- **THEN** imported providers overwrite existing providers entirely (current behavior, unchanged)

### Requirement: Crypto key file is validated for length
The `loadOrCreateKey` function in `crypto/encrypt.go` SHALL validate that the loaded key file is exactly 32 bytes. If the file contains fewer or more than 32 bytes, the function SHALL return a clear error message.

#### Scenario: Valid key file
- **WHEN** `~/.aix/key` contains exactly 32 bytes
- **THEN** the key is loaded successfully

#### Scenario: Truncated key file
- **WHEN** `~/.aix/key` contains fewer than 32 bytes (e.g., 16 bytes)
- **THEN** an error is returned: "invalid key file: expected 32 bytes, got N"

#### Scenario: Oversized key file
- **WHEN** `~/.aix/key` contains more than 32 bytes (e.g., 64 bytes)
- **THEN** an error is returned: "invalid key file: expected 32 bytes, got N"

#### Scenario: Key file does not exist
- **WHEN** `~/.aix/key` does not exist
- **THEN** a new 32-byte key is generated and saved (existing behavior, unchanged)

### Requirement: Export file uses restrictive permissions
The `config export --output` command SHALL write the exported file with 0600 permissions (owner read/write only), matching the token file permission.

#### Scenario: Export to file
- **WHEN** `config export --output backup.json` is executed
- **THEN** the file is created with permission bits 0600

### Requirement: Map iteration is deterministic
The `FormatForShell` method and `config current` command SHALL iterate map keys in sorted (alphabetical) order to produce deterministic output.

#### Scenario: FormatForShell output order
- **WHEN** `FormatForShell` is called with a settings object containing multiple env vars
- **THEN** the export lines are in alphabetical order by env var name

#### Scenario: Config current table order
- **WHEN** `config current` is run with multiple active env vars
- **THEN** the table rows appear in alphabetical order by key name
