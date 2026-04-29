# Releasing aix

This document describes how to publish new versions of aix, including binaries, GitHub releases, Homebrew formula updates, and Scoop bucket updates.

## Overview

Releases are fully automated via [GoReleaser](https://goreleaser.com). A single command builds binaries for all 6 platforms, creates a GitHub release with archives and checksums, updates the Homebrew tap formula, and updates the Scoop bucket manifest.

## Prerequisites

- [GoReleaser](https://goreleaser.com/install) installed (`brew install goreleaser`)
- `gh` CLI authenticated (`gh auth status`)
- Push access to `github.com/h4ck4life/aix-go`
- Push access to `github.com/h4ck4life/homebrew-aix`
- Push access to `github.com/h4ck4life/scoop-bucket`

## Release Workflow

### 1. Prepare changes

Make sure everything you want in the release is committed on the main branch:

```bash
git status    # should be clean
git log       # verify commits look right
```

### 2. Tag the release

Use semantic versioning (`vMAJOR.MINOR.PATCH`):

```bash
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1
```

Or use the Makefile:

```bash
make tag VERSION=v1.0.1
```

### 3. Run GoReleaser

GoReleaser uses the `gh` auth token automatically. No manual `GITHUB_TOKEN` export needed if `gh` is logged in:

```bash
goreleaser release --clean
```

Or use the Makefile (which exports the token for you):

```bash
make release VERSION=v1.0.1
```

### 4. What GoReleaser does

1. **Builds** binaries for:
   - `darwin/amd64`, `darwin/arm64`
   - `linux/amd64`, `linux/arm64`
   - `windows/amd64`, `windows/arm64`
2. **Packages** them into `.tar.gz` (Unix) and `.zip` (Windows)
3. **Generates** `checksums.txt`
4. **Creates** a GitHub Release at `github.com/h4ck4life/aix-go/releases`
5. **Pushes** the Homebrew formula to `github.com/h4ck4life/homebrew-aix/Formula/aix.rb`
6. **Pushes** the Scoop manifest to `github.com/h4ck4life/scoop-bucket/aix.json`

### 5. Verify the release

- Check the GitHub release: https://github.com/h4ck4life/aix-go/releases
- Check the Homebrew formula: https://github.com/h4ck4life/homebrew-aix/blob/main/Formula/aix.rb
- Check the Scoop manifest: https://github.com/h4ck4life/scoop-bucket/blob/main/aix.json
- Test Homebrew installation:
  ```bash
  brew update
  brew upgrade aix
  aix --version
  ```
- Test Scoop installation:
  ```powershell
  scoop update
  scoop update aix
  aix --version
  ```

## Homebrew Tap

The tap lives at `github.com/h4ck4life/homebrew-aix`. GoReleaser manages this automatically — you never edit the formula by hand. If the tap repo is ever deleted or recreated, GoReleaser will re-initialize it on the next release.

## Scoop Bucket

The Scoop bucket lives at `github.com/h4ck4life/scoop-bucket`. GoReleaser generates the `aix.json` manifest and pushes it to the root of this repo. The manifest stays in the root directory (not a subfolder) so `scoop bucket list` shows it correctly.

Windows users install via:

```powershell
scoop bucket add aix https://github.com/h4ck4life/scoop-bucket.git
scoop install aix
```

## Troubleshooting

### "git is in a dirty state"

Commit all changes before running GoReleaser. It refuses to release from a dirty working tree.

### "git tag was not made against commit"

The tag points to a different commit than HEAD. Either:
- Move the tag: `git tag -fa v1.0.1 -m "Release v1.0.1" && git push origin v1.0.1 --force`
- Or create a new version tag

### "could not read Username for https://github.com"

If `go install` from GitHub fails for users, it's because the repo requires authentication. For public repos, this is usually a local git/Go proxy issue. Users should set:

```bash
export GOPRIVATE=github.com/h4ck4life/aix-go
```

Or configure git to use SSH:

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### Version shows "dev" instead of the tag

The `main.version` variable must be a `var`, not a `const`, for `-ldflags` to override it at link time. This is already configured in `main.go` and `.goreleaser.yml`.
