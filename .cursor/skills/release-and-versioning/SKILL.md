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
3. **Let CI run** – `.github/workflows/release.yml` runs on tag push. GoReleaser creates the GitHub Release with neptune binaries, `neptune-webhook.zip` (Lambda), and auto-generated changelog (see `.goreleaser.yml`).
4. **Do not run destructive commands** (e.g. `git push` or creating tags) without user confirmation.

## GoReleaser

- **.goreleaser.yml**: Builds `neptune` from the repo root (ldflags set `main.version`, `main.commit`, `main.date`) and Lambda `bootstrap` from `lambda/` (released as `neptune-webhook.zip`). Produces archives (tar.gz, zip) and checksums; changelog is auto-generated (e.g. `use: github`).
- **release.yml**: Uses `goreleaser/goreleaser-action` with `release --clean` and `contents: write` permission.

If the user only wants to understand the workflow, explain the steps and point to `.goreleaser.yml` and `.github/workflows/release.yml` without running any commands.
