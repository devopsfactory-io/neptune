---
name: release-and-versioning
description: Cuts a new release with semantic versioning and GoReleaser. Use when the user wants to release a new version, tag a release, or understand release workflow.
---

# Release and Versioning

## Semantic Versioning

- Tags use the form **vMAJOR.MINOR.PATCH** (e.g. `v0.2.0`, `v1.0.0`).
- The `v` prefix is required for the release workflow. Do not add a version in `go.mod` for release; the binary gets version from ldflags at build time.

## Release Steps

1. **Ensure main is green** – CI (test, lint) must pass on the branch you will tag.
2. **Create and push the tag**:
   - `git tag vX.Y.Z`
   - `git push origin vX.Y.Z`
3. **Let CI run** – `.github/workflows/release.yml` runs on tag push. GoReleaser creates the GitHub Release with neptune binaries (lowercase archives e.g. `neptune_linux_amd64.tar.gz` and raw binaries), `neptune-webhook.zip` and raw `neptune-webhook_linux_amd64` (Lambda binary is `neptune-webhook`), checksums, and changelog in sections (see `.goreleaser.yml`).
4. **Do not run destructive commands** (e.g. `git push` or creating tags) without user confirmation.

## GoReleaser

- **.goreleaser.yml**: Builds `neptune` from the repo root (ldflags set `main.version`, `main.commit`, `main.date`) and Lambda `neptune-webhook` from `lambda/` (released as `neptune-webhook.zip` and raw `neptune-webhook_linux_amd64`). Produces lowercase-named archives (e.g. `neptune_linux_amd64.tar.gz`), raw binaries, and checksums; changelog is grouped (Breaking Changes, Features, Bug fixes, etc.) with release footer linking to full changelog.
- **release.yml**: Uses `goreleaser/goreleaser-action` with `release --clean` and `contents: write` permission.

## Breaking changes and changelog

When a release includes **breaking changes**, use the conventional `!` before `:` in the commit subject so those commits appear under the "Breaking Changes" section (e.g. `feat!: remove old API`, `fix(config)!: change default`). The changelog groups by the first line of the commit; only subjects containing `!:` are classified as Breaking Changes.

If the user only wants to understand the workflow, explain the steps and point to `.goreleaser.yml` and `.github/workflows/release.yml` without running any commands.
