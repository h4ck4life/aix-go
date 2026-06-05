## Context

The aix-go CLI tool manages API provider configurations and tokens for Claude Code. An audit uncovered systemic issues where operations silently succeed even when they fail — token errors discarded with `_ =`, registry deletes no-op on missing keys, and settings writes destroying unrelated Claude Code config. The codebase has 0 test coverage on `cmd/` handlers, meaning these bugs have no regression safety net.

The change spans 7 source files and adds 3 new test files. No external dependencies are added.

## Goals / Non-Goals

**Goals:**
- Every user-facing operation returns honest success/failure status
- Settings file writes preserve non-aix configuration keys
- `config import --merge` actually merges instead of always overwriting
- Doctor diagnostics give accurate results (no false positives)
- cmd/ handlers have unit test coverage for error paths and happy paths
- Crypto key loading validates key integrity before use
- Map iteration produces deterministic output

**Non-Goals:**
- Migrating from AES-CBC to AES-GCM (separate change — threat model is local-only)
- Adding `--json` output flags or shell completion (UX feature, not a bug fix)
- Adding `--token` flag to non-interactive `provider add` (feature request)
- Fixing Unicode display width in table padding (cosmetic, low priority)

## Decisions

### 1. Error propagation: return vs warn-and-continue

**Decision**: All 3 silent-error sites in `cmd/provider.go` will return the error (not just warn).

**Rationale**: `SetToken` failure during `provider add` means the provider is useless — returning an error prevents the success message from misleading the user. Same for `DeleteToken` during remove and `MoveToken` during rename. A warning alone would let the operation complete with a lie.

**Alternative considered**: Log a warning and continue. Rejected because the user sees "success" and has no indication the operation was incomplete.

### 2. RemoveOne existence check: before delete, not after

**Decision**: Check `r.data[name]` existence before calling `delete()`, return `ValidationError` if missing.

**Rationale**: `delete()` on missing key is a no-op in Go, so the only way to detect "nothing happened" is to check first. This is consistent with how `RenameOne` already handles missing providers (lines 189-191 of `registry.go`).

### 3. DeleteToken: return first non-nil error

**Decision**: If keychain deletion fails, return that error regardless of file deletion result. Only return nil if keychain succeeds (or keychain is unavailable).

**Rationale**: The current `&&` logic treats "keychain failed but file succeeded" as success, which means the token is still accessible via keychain. The user believes the token is deleted. The keychain is the primary store — its errors take priority.

### 4. Settings preservation: raw JSON round-trip

**Decision**: `Settings.Read()` unmarshals into `map[string]interface{}` first, extracts only `env`, and stores the full map. `Settings.Write()` merges `env` back into the stored map and marshals. `Reset()` only clears the `env` key.

**Rationale**: Claude Code's `settings.json` may contain `permissions`, `hooks`, and other keys we don't control. The current `Settings` struct with only `Env` field causes `json.Unmarshal` to discard unknown fields, and `json.Marshal` to write only `Env`. We need to preserve unknown fields.

**Alternative considered**: Using `json.RawMessage` for unknown fields. More complex, same result. The raw map approach is simpler for a single-file settings format.

### 5. Config import merge: field-level merge on conflict

**Decision**: When `--merge` is true and a provider name exists in both current and imported configs, merge field-by-field: imported values overwrite existing values, but fields not present in the imported config preserve existing values.

**Rationale**: This is the expected merge behavior — imported config is the "source of truth" for fields it specifies, but doesn't wipe unspecified fields.

### 6. Test strategy for cmd/ handlers

**Decision**: Test cmd/ handlers by directly calling the `RunE` functions with mock `cobra.Command` and args. Use real `core.Registry` with temp directories (no interface mocking needed since registry writes to a known path).

**Rationale**: The `RunE` functions accept `*cobra.Command` and `[]string` — we can construct these directly. Using temp directories for `~/.anthropic-switch/` and `~/.aix/` gives realistic tests without mocking file I/O. This is simpler than introducing interfaces for Registry and TokenManager.

### 7. Deterministic map iteration: sort keys

**Decision**: Collect map keys into a slice, `sort.Strings()`, then iterate in sorted order. Apply this in `FormatForShell` and `config current`.

**Rationale**: Simple, idiomatic Go, zero dependencies. Alphabetical order is predictable and debuggable.

## Risks / Trade-offs

- **[Risk] Settings round-trip may reformat JSON**: Preserving unknown fields via `map[string]interface{}` may change whitespace or key ordering in `settings.json`. → **Mitigation**: Use `json.MarshalIndent` with same indent as Claude Code uses. Minor formatting differences are acceptable.

- **[Risk] RemoveOne breaking change for callers**: Any code that calls `RemoveOne` on a nonexistent provider now gets an error. → **Mitigation**: Only `runProviderRemove` calls it, and it always loads the registry first. The only scenario is "user types a wrong name" — returning an error is the correct behavior.

- **[Risk] cmd/ tests depend on temp directory setup**: Tests that create temp dirs for registry/token paths must clean up. → **Mitigation**: Use `t.TempDir()` which auto-cleans. Override path constants in test setup.

- **[Trade-off] MoveToken atomicity**: True atomicity would require a single locked transaction. Current fix (reorder to set-then-delete) still has a window if process crashes between operations. → **Acceptable**: This is a CLI tool, not a distributed system. Crash-safe atomicity would require a WAL or journal, which is overkill.

- **[Trade-off] Doctor network check may be slow**: Adding status code check doesn't change latency (already waits for response). But the 5-second timeout is the real cost. → **No change needed**: Already acceptable for a diagnostic command.
