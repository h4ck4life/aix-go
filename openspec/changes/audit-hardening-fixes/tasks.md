## 1. Silent Error Fixes in cmd/provider.go

- [x] 1.1 Fix `runProviderAdd` line 129: replace `_ = tokenMgr.SetToken(name, token)` with error check — if error, return it and do not print success message
- [x] 1.2 Fix `runProviderRemove` line 241: replace `_ = tokenMgr.DeleteToken(name)` with error check — if error, print warning but still proceed with registry removal
- [x] 1.3 Fix `runProviderRename` line 274: replace `_ = tokenMgr.MoveToken(oldName, newName)` with error check — if error, return it and skip registry rename

## 2. Registry and Token Safety Fixes

- [x] 2.1 Fix `RemoveOne` in `core/registry.go`: add existence check `if _, ok := r.data[name]; !ok` before `delete()` — return `ValidationError` if not found
- [x] 2.2 Fix `DeleteToken` in `core/token.go`: change condition from `keychainErr != nil && fileErr != nil` to `keychainErr != nil` — return keychain error when keychain is the active backend, regardless of file result
- [x] 2.3 Fix `MoveToken` in `core/token.go`: reorder to set-then-delete (currently delete-then-set is not the issue, but verify order is `GetToken → SetToken(new) → DeleteToken(old)`)

## 3. Settings File Preservation

- [x] 3.1 Refactor `Settings` struct in `core/settings.go`: add a `Raw map[string]interface{}` field to preserve non-env keys from settings.json
- [x] 3.2 Update `Settings.Read()`: unmarshal into `map[string]interface{}` first, extract `env` into `s.Env`, store full map in `s.Raw`
- [x] 3.3 Update `Settings.Write()`: merge `s.Env` back into `s.Raw["env"]`, marshal the full `s.Raw` map
- [x] 3.4 Update `Settings.Reset()`: only clear `s.Env` and `s.Raw["env"]`, preserve all other keys

## 4. Config Import Merge

- [x] 4.1 Implement actual merge logic in `runConfigImport`: when `--merge` is true and provider exists, merge imported fields into existing config instead of full overwrite
- [x] 4.2 Add merge function to `core/registry.go`: `MergeOne(name string, incoming constants.ProviderConfig)` that updates only non-zero fields from incoming config

## 5. Doctor Accuracy

- [x] 5.1 Update `checkNetwork` in `cmd/doctor.go`: check `resp.StatusCode >= 400` and return failure with status code

## 6. Model Name Validation

- [x] 6.1 Add `ValidateModelName(name string) error` to `validation/validation.go`: model names must be non-empty and match `/^[a-zA-Z0-9._-]+$/`
- [x] 6.2 Add validation call in `provider edit --model` path in `cmd/provider.go` line 396
- [x] 6.3 Add validation call in `runProviderSetModel` in `cmd/provider.go` line 322

## 7. Deterministic Map Iteration

- [x] 7.1 Update `FormatForShell` in `core/settings.go`: sort `s.Env` keys alphabetically before iterating
- [x] 7.2 Update `runConfigCurrent` in `cmd/config.go`: sort env map keys before building table rows

## 8. Crypto Key Validation

- [x] 8.1 Update `loadOrCreateKey` in `crypto/encrypt.go`: after reading key file, validate `len(data) == 32` — return error "invalid key file: expected 32 bytes, got N" if mismatch

## 9. Export File Permissions

- [x] 9.1 Update `runConfigExport` in `cmd/config.go`: change `os.WriteFile(configExportOutput, data, 0644)` to `0600`

## 10. Unit Tests — cmd/provider_test.go

- [x] 10.1 Create `cmd/provider_test.go` with test setup using `t.TempDir()` for registry and token paths
- [x] 10.2 Add test: `runProviderAdd` happy path — valid name + URL, verify provider in registry
- [x] 10.3 Add test: `runProviderAdd` missing args — verify ValidationError with usage message
- [x] 10.4 Add test: `runProviderAdd` invalid name — verify ValidationError
- [x] 10.5 Add test: `runProviderAdd` token storage failure — verify error returned (not silently discarded)
- [x] 10.6 Add test: `runProviderRemove` nonexistent provider — verify ValidationError (after RemoveOne fix)
- [x] 10.7 Add test: `runProviderRename` nonexistent provider — verify ValidationError
- [x] 10.8 Add test: `runProviderUse` nonexistent provider — verify error
- [x] 10.9 Add test: `runProviderUse` no token — verify TokenError
- [x] 10.10 Add test: `runProviderList` empty registry — verify "No providers configured" output
- [x] 10.11 Add test: `runProviderEdit` no flags — verify ValidationError

## 11. Unit Tests — cmd/config_test.go

- [x] 11.1 Create `cmd/config_test.go` with test setup
- [x] 11.2 Add test: `config export` roundtrip — export then import, verify providers match
- [x] 11.3 Add test: `config import` with invalid JSON — verify ValidationError
- [x] 11.4 Add test: `config import --merge` — verify merge behavior (new provider added, existing provider updated)
- [x] 11.5 Add test: `config import` nonexistent file — verify FileNotFoundError

## 12. Unit Tests — cmd/doctor_test.go

- [x] 12.1 Create `cmd/doctor_test.go` with test setup
- [x] 12.2 Add test: `checkRegistry` with valid registry — verify returns true
- [x] 12.3 Add test: `checkNetwork` HTTP 500 — verify returns false with status code (after fix)
- [x] 12.4 Add test: `checkPermissions` writable dirs — verify returns true

## 13. Unit Tests — crypto edge cases

- [x] 13.1 Add test to `crypto/encrypt_test.go`: truncated key file — verify `loadOrCreateKey` returns error with byte count
- [x] 13.2 Add test: oversized key file — verify error
- [x] 13.3 Add test: valid 32-byte key file — verify loads successfully

## 14. Unit Tests — core RemoveOne edge case

- [x] 14.1 Add test to `core/registry_test.go`: `RemoveOne` on nonexistent provider — verify returns error
- [x] 14.2 Add test: `RemoveOne` on existing provider — verify returns nil and provider is gone

## 15. Verify and Lint

- [x] 15.1 Run `go test ./...` — all tests pass
- [x] 15.2 Run `make lint` — no lint errors (golangci-lint not installed, go vet clean)
- [x] 15.3 Run `make fmt` — code formatted
