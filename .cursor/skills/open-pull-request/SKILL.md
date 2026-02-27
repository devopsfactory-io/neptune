---
name: open-pull-request
description: Commits current changes and opens a pull request via GitHub CLI (gh). Use only when the user explicitly says "open a pull request", "open an pull request", or "create a pull request". Do not run this workflow for other requests.
---

# Open a Pull Request

## When to Use

Use this skill **only** when the user explicitly requests to open a pull request (e.g. "open a pull request", "open an pull request", "create a pull request"). Do not run git commit, push, or `gh pr create` for any other request.

## Preconditions

1. **Working tree**: Confirm that current changes (staged or unstaged) are what the user wants to commit. If the state is unclear or already on a feature branch with unpushed commits, ask before proceeding.
2. **GitHub CLI**: Require `gh` installed and authenticated. Run `gh auth status`; if not OK, instruct the user to install [GitHub CLI](https://cli.github.com/) and run `gh auth login`, then stop.

## Workflow

### 1. Create a new branch

- Create a new branch from current HEAD: `git checkout -b <branch-name>`.
- **Branch name**: Use user-provided name if given. Otherwise derive from the commit message: e.g. `feat/scope-short-desc` or `fix/short-desc` (slug from the first line of the conventional commit). Prefer lowercase, hyphens; avoid spaces.

### 2. Stage and commit

- Run `git add .`
- **Commit message**: Use conventional style per project (e.g. `feat(scope): description`, `fix(scope): description`). If the user provided a message, use it; otherwise generate from `git status` and `git diff` (or `git diff --staged` after add).
- Run `git commit -m "<message>"`

### 3. Push and create PR

- Push: `git push -u origin <branch-name>`
- Create PR: `gh pr create` with:
  - **Title**: Human-readable, concise; conventional style preferred (e.g. `feat(lock): integrate Terramate SDK for change detection`).
  - **Body**: Use the dual-purpose template below (human summary + AI block).

### 4. Safety

- Do not run `git push`, `git commit`, or `gh pr create` unless the user has explicitly requested to open a pull request.
- If the repo is in an unexpected state (e.g. already on a feature branch with unpushed commits, or detached HEAD), ask the user before creating a new branch or pushing.

## PR body template

Use this structure so the PR is readable by both humans and AI:

```markdown
## Summary
[One or two sentences for humans: what this PR does.]

## Summary for AI
- **Scope**: [e.g. config, lock, run]
- **Type**: feat | fix | docs | refactor
- **Key files/areas**: [paths or areas]
```

Fill the placeholders from the commit message and the changed files (e.g. from `git diff --name-only` or the conversation context).
