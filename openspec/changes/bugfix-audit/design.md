## Context

aix-go is a Go CLI using Cobra for commands, Bubble Tea for interactive UI, and an encrypted file/keychain for token storage. A full-code read surfaced 16 defects across error handling, concurrency, UI lifecycle, shell output, and missing commands. Most fixes are localized, but a few (error exit codes, settings persistence, provider edit) require cross-package coordination.

## Goals / Non-Goals

**Goals:**
- Fix all identified runtime bugs and broken features.
- Add `provider edit` and `provider use --persist`.
- Ensure error exit codes match the documented hierarchy.
- Eliminate the `DeleteToken` race and `RunSpinner` hang.
- Make shell output safe for special characters.

**Non-Goals:**
- No new external dependencies.
- No rewrite of the config format or registry schema.
- No test suite expansion (out of scope per user request).

## Decisions

### Error exit codes: interface-based dispatch
Instead of fixing brittle type assertions against `*AixError`, introduce a private interface (`coder`) with a `code() int` method. Implement it on all error types via pointer receiver. `GetExitCode` uses `errors.As` against the interface. This is idiomatic Go and survives future error subtypes.

### Settings persistence: explicit `--persist` flag
`aix provider use` currently prints to stdout by design. To avoid surprising users who eval the output, persistence is opt-in via `--persist`. When set, `GenerateEnvironmentVars` is called, then `settings.Write()` saves to `~/.claude/settings.json`. This keeps backward compatibility.

### Provider edit: in-place mutation via registry
`provider edit <name>` will support `--url`, `--token-type`, `--model`, and `--default-model` flags. It loads the registry, mutates the existing `ProviderConfig`, and saves. No token migration is needed because provider name stays the same.

### Description flag: remove it
`ProviderConfig` has no `Description` field and the Node.js original does not store descriptions. Rather than expanding the data model, remove the `--description` flag and the interactive wizard step. This is the minimal fix.

### Token test endpoint: append `/v1/models`
Anthropic-compatible APIs expose `GET /v1/models` for listing available models. Append this path to `cfg.BaseURL` if it does not already end with it. This yields a valid test request.

### Shell escaping: Go `strconv.Quote` for POSIX, manual for fish/PowerShell/CMD
`FormatForShell` will quote values. For bash/zsh/fish, use single-quote wrapping with `'` escape. For PowerShell, escape `"` as `""`. For CMD, no additional escaping beyond the existing `set KEY=VALUE`.

## Risks / Trade-offs

- **[Risk]** Changing `--description` from no-op to removed could break scripts that pass it.  
  **Mitigation:** The flag never did anything; removing it is a no-op for behavior, but Cobra will error if scripts pass it. Since it was already non-functional, this is acceptable breakage.

- **[Risk]** `provider use --persist` writes to `~/.claude/settings.json`, which is also managed by Claude Code itself.  
  **Mitigation:** Only write the `env` key; preserve all other keys in the JSON file. Read → modify `env` → write.

- **[Risk]** Appending `/v1/models` to base URLs that already include it could double the path.  
  **Mitigation:** Check if `BaseURL` already ends with `/v1/models` before appending.
