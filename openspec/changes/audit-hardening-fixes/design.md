## Context

`aix` is a Go CLI tool (~3K lines across 24 files) that switches between Anthropic-compatible API providers. It was ported from a Node.js version and has not had a hardening pass. A multi-agent audit found 30 issues across correctness, consistency, security, and test coverage. The codebase is well-structured (Cobra commands, Bubble Tea UI, sync.RWMutex-protected registry) but has specific weak points in concurrency, crypto, shell escaping, and error handling. 14 of 15 packages have zero tests.

The tool handles sensitive data (API tokens) and generates shell commands evaluated via `eval $(aix provider use ...)`, making correctness in concurrency, encryption, and shell escaping critical.

## Goals / Non-Goals

**Goals:**
- Eliminate all high-severity correctness bugs (race conditions, error swallowing)
- Fix the padding oracle vulnerability in AES-256-CBC decryption
- Complete shell escaping for all 5 supported shells, especially CMD and PowerShell
- Achieve consistent error wrapping across all commands using the existing `utils` error hierarchy
- Add unit test coverage for `core/`, `crypto/`, `validation/`, `constants/`, and `core/settings.go`
- Add `dist/` to `.gitignore`

**Non-Goals:**
- Rewriting the encryption scheme (e.g., switching from CBC to GCM) — fix the oracle, don't redesign
- Adding integration or end-to-end tests in this change
- Changing the CLI interface or config file format
- Adding new features or commands
- Fixing low-severity nits (doctor network check granularity, table bounds that can't be hit in practice)

## Decisions

### 1. Token file atomicity: read-modify-write under write lock

**Current:** `setInFile` calls `deleteFromFile` (reads entire file, rewrites without the account) then opens the file again in append mode. The delete error is swallowed.

**Decision:** Replace the delete-then-append pattern with a single read-modify-write under the existing mutex. Read all lines, replace or append the target entry, write the entire file atomically (temp + rename, same pattern as `registry.saveLocked`).

**Rationale:** The current two-step approach can lose the token if the process crashes between delete and append. A single atomic write eliminates this window. The mutex already serializes access, so there's no performance concern.

**Alternative considered:** Using `fcntl`/`flock` file locking — rejected because it adds OS-specific complexity for a single-user CLI tool where the in-process mutex is sufficient.

### 2. Registry RenameOne: already safe under write lock

**Current:** `RenameOne` acquires write lock, checks existence, deletes old key, sets new key. All under `r.mu.Lock()`.

**Decision:** No code change needed. The audit flagged this as a concern, but re-reading the code: the entire operation is within a single `Lock()/Unlock()` block at `registry.go:181-196`. Concurrent calls are properly serialized. Mark as verified-safe.

### 3. Padding oracle: constant-time validation

**Current:** `crypto/encrypt.go:73-81` validates PKCS7 padding by checking each byte. If any byte is wrong, it returns immediately, creating a timing side-channel.

**Decision:** Use `subtle.ConstantTimeCompare` or a constant-time byte-by-byte check that always compares all padding bytes regardless of mismatches. Return a single generic "decryption failed" error (no distinction between "bad padding" and "ciphertext too short").

**Rationale:** While a local attacker with write access to `tokens.enc` is an unlikely threat model for a CLI tool, the fix is trivial and follows defense-in-depth. The Go `crypto/subtle` package is stdlib — no new dependency.

### 4. Shell escaping: use `shellescape`-style for POSIX, proper escaping for CMD and PowerShell

**Current:** POSIX shells get single-quote escaping (correct). PowerShell escapes only `"`. CMD does no escaping at all.

**Decision:**
- **bash/zsh/fish:** Keep existing single-quote escaping — it's correct for POSIX shells.
- **CMD:** Use `^` escaping for special characters (`& | < > ^ "`) and wrap in double quotes.
- **PowerShell:** Use single-quote wrapping (like POSIX) since PowerShell single quotes are literal strings. For values containing single quotes, use the `''` escape (double-single-quote) pattern.

**Rationale:** PowerShell's single-quote strings are simpler and safer than backtick-escaping inside double quotes. CMD has limited escaping capability, but `^` before metacharacters covers the attack surface.

**Alternative considered:** Using a third-party shell-escaping library — rejected to avoid new dependencies for a bounded problem.

### 5. Error consistency: wrap all raw errors in cmd/ through utils types

**Current:** `cmd/doctor.go` returns raw errors. Other commands mostly use the custom types. `main.go` calls `utils.GetExitCode(err)` which defaults to exit code 1 for non-`AixError` types.

**Decision:** Add a utility function `utils.WrapError(message string, err error) error` that wraps any error in a generic `AixError` with `ExitGeneralError`. Use it in `doctor.go` and audit all other `cmd/` files for bare error returns.

**Rationale:** The custom error hierarchy already exists and works. We just need to use it consistently. Adding a generic wrapper is cleaner than choosing specific error types for diagnostic failures.

### 6. Test infrastructure: interfaces + table-driven tests, no external deps

**Decision:** Use Go's built-in `testing` package with table-driven tests. For keychain/file dependencies, use interfaces + mocks (hand-written, no mockgen). Create a `testutil` package with temp-directory helpers.

**Rationale:** The codebase currently has no test dependencies. Adding `testify` or `mockgen` would be a philosophical shift. Go stdlib is sufficient for this codebase size. Interfaces are already partially there (keychain wraps `go-keyring`).

### 7. Keychain availability check: use random-ish account name

**Current:** `IsAvailable()` uses a fixed test account `aix-test-0`. Two concurrent `aix` processes would clobber each other.

**Decision:** Include the PID in the test account name: `aix-test-{pid}`. This is deterministic per-process and unique across concurrent invocations.

**Rationale:** PID is available, unique per process, and doesn't require importing `crypto/rand` into the keychain package. Process collision (PID reuse) within the microseconds of the test is negligible.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Rewriting `setInFile` could introduce new bugs in token storage | Encrypt/decrypt round-trip tests before and after; test with existing `tokens.enc` files |
| CMD escaping is inherently limited — some characters cannot be escaped in all contexts | Document the limitation; the primary use case (URLs and API keys) doesn't typically contain CMD metacharacters |
| Changing error types in `doctor.go` changes its exit code from 0-on-failure to 1-on-failure | This is actually a bug fix — doctor currently returns `nil` even when checks fail |
| Tests using interfaces require refactoring `keychain` and file ops behind interfaces | Minimal refactoring — only what's needed for testability |
| Constant-time padding check has micro-performance cost | Negligible for a CLI tool that decrypts a handful of tokens |

## Open Questions

- Should `dist/` be added to `.gitignore` or should the entire directory be removed from tracking? (Currently untracked per git status, but worth confirming intent)
- Should we bump the module version after hardening, or is this a patch-level change?
