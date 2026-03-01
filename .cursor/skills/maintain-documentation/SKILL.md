---
name: maintain-documentation
description: Ensures human and AI documentation stay in sync with code and config. Use when changing behavior, adding features, refactoring, or when the user asks to update docs. Always consider whether README, docs/, examples/, AGENTS.md, .cursor/rules, or .cursor/skills need updates.
---

# Maintain Documentation

## When to Use

Use this skill when:
- Changing user-facing behavior, CLI commands, or flags
- Adding or changing config or env vars
- Refactoring project structure or conventions
- Modifying CI or release workflows
- The user asks to update or review documentation

## Checklist After a Change

1. **README.md** – Update if install steps, usage, or config schema changed. Keep high-level content and links to docs accurate.
2. **docs/*.md** – Update if install, config, object storage, usage, or development steps changed.
3. **examples/** – Treat as part of the documentation. Add or update examples when config, workflows, or usage change; keep [examples/README.md](examples/README.md) and per-example READMEs accurate.
4. **CONTRIBUTING.md** – Update if contributing workflow, issue/PR process, or checklist for contributors changed.
5. **AGENTS.md** – Update if project structure, setup commands, code style, testing, or CI changed. Keep repo layout and "Documentation and AI context" section accurate.
6. **.github/** – Update [pull_request_template.md](.github/pull_request_template.md) or [ISSUE_TEMPLATE/](.github/ISSUE_TEMPLATE/) if PR/issue structure or required sections change.
7. **.cursor/rules/*.mdc** – Update if coding conventions, workflow rules, or file-scoped guidance changed. Match globs to the files the rule applies to.
8. **.cursor/skills/*/SKILL.md** – Update if a documented workflow (e.g. release steps, test commands, open-pull-request) or checklist changed.

## Where Things Live

| Artifact | Audience | Purpose |
|----------|----------|---------|
| README.md | Humans | High-level entry point; links to docs and releases |
| docs/ | Humans | Configuration, object storage, installation, usage, development |
| examples/ | Humans | Copy-pasteable infra examples (S3/GCS, automerge, Terramate, Terragrunt); part of docs |
| CONTRIBUTING.md | Humans | How to contribute; issue/PR templates, checklist, CI |
| .github/pull_request_template.md | Humans + AI | PR description structure; includes AI Summary for reviewers |
| .github/ISSUE_TEMPLATE/ | Humans | Bug report, feature request templates |
| AGENTS.md | AI agents | Project overview, structure, setup, style, testing, CI, doc-update requirement |
| .cursor/rules/*.mdc | AI agents | File-specific or always-applied rules (Go, CI, config, docs) |
| .cursor/skills/*/SKILL.md | AI agents | Step-by-step workflows (e.g. maintain-documentation, release, open-pull-request) |

## When to Update AI Docs

- **New commands or flags** → README and AGENTS.md (structure / usage).
- **New env vars or config keys** → README, docs/ (configuration.md, object-storage.md), .neptune.example.yaml, examples/ (if an example should show the new config), config-and-yaml rule.
- **New Make targets or CI jobs** → AGENTS.md, ci-and-release rule, testing-and-ci skill.
- **New features or workflow options users can try** → Consider adding or updating an example in **examples/**.
- **New patterns agents should follow** → Add or update a rule or skill; mention in AGENTS.md if central.

Do not edit plan files (e.g. in .cursor/plans or *.plan.md) unless the user explicitly asks.
