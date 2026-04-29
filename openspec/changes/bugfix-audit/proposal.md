## Why

A code audit of the aix-go CLI revealed 16 bugs, broken features, and design gaps — including non-working `--version`, broken error exit codes, a hung spinner, data races, no-op flags, and stale config reads. These issues degrade reliability, confuse users, and break shell integrations. This change fixes all identified defects to bring the CLI to production quality.

## What Changes

- **Fix `--version` flag** (`cmd/root.go`): Set `rootCmd.Version` inside `Execute` so Cobra registers the flag correctly.
- **Fix error exit codes** (`utils/errors.go`): Rewrite `IsAixError`, `AsAixError`, and `GetExitCode` to use `errors.As` with an interface, so `ValidationError`, `FileNotFoundError`, etc. report correct exit codes.
- **Fix `RunSpinner` hang** (`ui/spinner.go`): Quit the Bubble Tea program via `tea.Quit` command instead of an unhandled fake keypress.
- **Fix `DeleteToken` race** (`core/token.go`): Hold `tm.mu` around `deleteFromFile` call.
- **Wire up `debugFlag`** (`cmd/root.go`): Call `utils.InitLogger(debugFlag)` before command execution.
- **Remove or implement `--description` flag** (`cmd/provider.go`, `interactive/add_provider.go`): Either add `Description` to `ProviderConfig` or remove the flag/wizard step.
- **Fix `token test` endpoint** (`cmd/token.go`): Send GET to `/v1/models` (or append `/v1/models` to base URL) instead of bare base URL.
- **Connect `provider use` to `config current`** (`cmd/config.go`, `cmd/provider.go`, `core/settings.go`): Add `--persist` flag to `provider use` that writes env vars to `~/.claude/settings.json`.
- **Escape shell values** (`core/settings.go`): Quote/escape token and URL values in `FormatForShell`.
- **Add `version` subcommand** (`cmd/root.go`): Register a `version` command that prints version info.
- **Add `provider edit` command** (`cmd/provider.go`): Allow modifying existing provider URL, token type, and model without remove+re-add.
- **Fix `MoveToken` blocking rename** (`core/token.go`): Return success when old token doesn't exist; rename should not require a token.
- **Add bounds checking to `RenderSimpleTable`** (`ui/table.go`): Skip missing cells instead of panicking.
- **Fix `checkPermissions` gaps** (`cmd/doctor.go`): Check parent directory writability when target path doesn't exist; report missing paths.
- **Tighten registry file permissions** (`core/registry.go`): Use `0600` instead of `0644`.

## Capabilities

### New Capabilities
- `provider-edit`: Modify existing provider properties (URL, token type, model) without deleting and recreating.
- `provider-use-persist`: Persist activated provider environment variables to `~/.claude/settings.json` via an optional flag.

### Modified Capabilities

## Impact

- `cmd/*`: All command files touched for bug fixes and new commands.
- `core/*`: Token manager locking, settings persistence, registry permissions.
- `ui/*`: Spinner lifecycle, table bounds checking.
- `utils/*`: Error type hierarchy and exit code logic.
- `interactive/*`: Remove unused description step or wire it through.
- No new dependencies; changes use existing Go stdlib, Cobra, and Bubble Tea APIs.
