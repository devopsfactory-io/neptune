# Changelog

All notable changes to Neptune are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Neptune adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [1.0.0] - 2026-03-16

First stable, production-ready release of Neptune. This release signals API and
configuration stability: the `.neptune.yaml` schema, CLI flags, Lambda webhook
contract, and GitHub Actions integration are considered stable going forward.
Patch and minor updates will not introduce breaking changes without a major
version bump.

### CI/CD

- Automate Lambda deploy and smoke test on release — the `release.yml` workflow
  now builds, uploads, and smoke-tests the Lambda function as part of every
  tagged release (#82)
- Bump `golangci/golangci-lint` to v2.11.2 (#83) and v2.11.3 (#87) in the
  `lint.yml` workflow

### Dependencies

- Bump `github.com/aws/aws-lambda-go` to v1.53.0 in `lambda/go.mod` (#86)

---

## [0.2.0] - 2026-03-12

This release concentrates on developer and operator experience: AI-assisted
workflows via Claude Code, documentation quality, and continuous dependency
maintenance.

### Added

- **AI agent tooling** — `issue-reviewer` and `pr-reviewer` Claude Code
  subagents for validating issue and PR drafts before upload (#65)
- **AI agent tooling** — `issue-writer` subagent powers the `/feature` and
  `/bug` slash commands, creating structured GitHub issues (#68)
- **AI agent tooling** — `documentation-maintainer` subagent that runs the full
  doc checklist (README, docs/, examples/, CLAUDE.md, skills) after any
  behavior-affecting change (#74)
- Validation rule: validate `/feature` and `/bug` drafts with `issue-reviewer`
  before calling `gh issue create` (#73)
- Go Report Card badge added to README (#67)

### Changed

- AI configuration migrated from Cursor to Claude Code; `.cursor/` rules
  replaced by `.claude/` agents, skills, and commands (#77)
- CLAUDE.md, Makefile, and GoReleaser config updated for clarity and
  correctness; CloudFormation parameter descriptions fixed (#81)
- Clarified `documentation-maintainer` delegation methods in CLAUDE.md (#79)

### Documentation

- Go Report Card badge temporarily removed then restored to README (#66, #67)
- AI configuration migration guide from Cursor to Claude Code (#77)

### Dependencies

- Bump `github.com/aws/aws-sdk-go-v2/service/s3` to v1.96.2 (#69) and
  v1.96.4 (#80)
- Bump `github.com/zclconf/go-cty` to v1.18.0 (#55)
- Bump `github.com/google/go-github/v83` to v84 (#75)
- Bump `aws-sdk-go-v2` monorepo in `lambda/go.mod` (#76)
- Bump Go toolchain to v1.26.1 in `lambda/go.mod` (#78)

---

## [0.1.0] - 2026-03-03

Initial release of Neptune — a Terraform and OpenTofu pull request automation
tool that runs entirely in GitHub Actions. No separate servers or self-hosted
runners required.

### Added

- **Core engine** — Refactored Neptune from Python to Go; idiomatic error
  handling, structured logging with `zerolog`, explicit context propagation
- **Stack locking** — GCS and S3-compatible object storage backends; prevents
  concurrent modifications to the same stack across PRs
- **Terramate integration** — Uses the Terramate Go SDK for change detection and
  ordered multi-stack plan/apply execution (#4)
- **Local stacks** — `neptune stacks` command and `local` mode using
  `stack.hcl` discovery and git-based change detection (#17)
- **GitHub integration** — Posts plan/apply output as PR comments; sets commit
  statuses for `neptune plan` and `neptune apply` (use as required status
  checks in branch protection)
- **Lambda webhook** — GitHub App webhook handler deployed as AWS Lambda
  with CloudFormation template; triggers Neptune on PR open and `@neptbot`
  mentions (#6)
- **neptbot GitHub App** — Lambda gates on a PR label (`NEPTUNE_PR_LABEL`) and
  adds an eyes reaction before triggering, preventing accidental runs (#10,
  #12)
- **Config security** — `.neptune.yaml` is always loaded from the default
  branch via git, not the PR branch, to prevent config injection attacks (#9)
- **Auto-merge** — `repository.automerge: true` enables automatic PR merge
  after a successful apply (#13)
- **`--log-level` flag** — Runtime log-level control via CLI flag or
  `NEPTUNE_LOG_LEVEL` environment variable; supports DEBUG, INFO, ERROR (#48)
- **E2E tests** — Integration test suite runnable against MinIO with
  `make e2e` or `./e2e/scripts/run-terramate.sh` (#3)
- **Integration tests** — CI workflow that runs against real GitHub (#5)
- **Examples** — `examples/` submodule with S3/GCS backend, automerge,
  Terramate stacks, and Terragrunt configurations (#14)

### CI/CD

- GitHub Actions workflows: test, lint (`golangci-lint`), integration, e2e,
  release (GoReleaser), labeler
- GoReleaser config: multi-platform binaries, Lambda zip artifact,
  checksums, release notes via `github-native` changelog
- GitHub-native release notes categories defined in `.github/release.yml`
  (Breaking Changes, Features, Bug Fixes, Documentation, Dependencies,
  Other) with PR label routing (#58)
- Automated PR labeling by branch name prefix (`feat/`, `fix/`, `enhance/`,
  `ci/`, `deps/`) via `actions/labeler` (#58)
- `label-old-prs` workflow to backfill labels on existing PRs (#62, #63, #64)
- Renovate bot configured for automated dependency updates to Go modules and
  GitHub Actions (#19, #20)
- Concurrency disabled in workflows to prevent race conditions on overlapping
  runs (#27)
- E2e and integration workflows trigger on `go.mod`/`go.sum` changes (#26)
- DCO sign-off enforced; all commits must include `Signed-off-by` trailer

### Documentation

- Neptune logo added to README (#18)
- Getting started guides and neptune PR label requirement documented (#47)
- `neptbot` GitHub App install URL published in README
- Open source and CNCF readiness statement (#16)
- `CONTRIBUTING.md`, PR/issue templates, and labeler workflow documented (#11)
- Workflow comparison between standard Terraform, Neptune, and Atlantis

### Dependencies

- Initial Go module dependencies: `terramate-io/terramate`, `aws-sdk-go-v2`,
  `google/go-github`, `rs/zerolog`, `hashicorp/hcl/v2`, `aws-lambda-go`
- Security patches: `sigstore/timestamp-authority`, `cloudflare/circl`,
  `golang-jwt/jwt`, `go.opentelemetry.io/otel/sdk` (#21–#24, #28, #29)

---

[Unreleased]: https://github.com/devopsfactory-io/neptune/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/devopsfactory-io/neptune/compare/v0.2.0...v1.0.0
[0.2.0]: https://github.com/devopsfactory-io/neptune/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/devopsfactory-io/neptune/releases/tag/v0.1.0
