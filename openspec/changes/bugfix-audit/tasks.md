## 1. Foundation Fixes

- [x] 1.1 Fix `--version` flag: set `rootCmd.Version = ver` inside `Execute` before `rootCmd.Execute()` (`cmd/root.go`)
- [x] 1.2 Add `version` subcommand that prints version info (`cmd/root.go`)
- [x] 1.3 Wire `debugFlag` to `utils.InitLogger(debugFlag)` in root command pre-run (`cmd/root.go`)
- [x] 1.4 Fix error exit codes: introduce `coder` interface, implement on all error types, rewrite `GetExitCode` with `errors.As` (`utils/errors.go`)

## 2. Core Stability

- [x] 2.1 Fix `DeleteToken` race: acquire `tm.mu` before calling `deleteFromFile` (`core/token.go`)
- [x] 2.2 Fix `MoveToken` to not error when old token is missing (`core/token.go`)
- [x] 2.3 Tighten registry file permissions from `0644` to `0600` (`core/registry.go`)
- [x] 2.4 Fix `checkPermissions` to check parent dirs and report missing paths (`cmd/doctor.go`)

## 3. UI and Shell Fixes

- [x] 3.1 Fix `RunSpinner` hang: use `tea.Quit` command instead of fake keypress (`ui/spinner.go`)
- [x] 3.2 Add bounds checking to `RenderSimpleTable` to prevent panic on short rows (`ui/table.go`)
- [x] 3.3 Escape shell values in `FormatForShell` for all supported shells (`core/settings.go`)
- [x] 3.4 Fix `token test` to send GET to `/v1/models` appended to base URL (`cmd/token.go`)

## 4. Provider Commands

- [x] 4.1 Remove `--description` flag and interactive wizard step (`cmd/provider.go`, `interactive/add_provider.go`)
- [x] 4.2 Implement `aix provider edit <name>` with `--url`, `--token-type`, `--model`, `--default-model` flags (`cmd/provider.go`)

## 5. Settings Persistence

- [x] 5.1 Add `--persist` flag to `aix provider use` (`cmd/provider.go`)
- [x] 5.2 When `--persist` is set, call `settings.Write()` after generating env vars (`cmd/provider.go`, `core/settings.go`)

## 6. Documentation and Release

- [x] 6.1 Update `CLAUDE.md` to reflect new commands and fixed behavior
- [x] 6.2 Update `README.md` with new `provider edit`, `version`, and `--persist` usage
- [x] 6.3 Run `make fmt`, `go vet ./...`, `go test ./...`
- [x] 6.4 Git commit and push
- [x] 6.5 Create and push release tag, publish via GoReleaser
