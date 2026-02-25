---
name: maintain-documentation
description: Ensures human and AI documentation stay in sync with code and config. Use when changing behavior, adding features, refactoring, or when the user asks to update docs. Always consider whether README, AGENTS.md, .cursor/rules, or .cursor/skills need updates.
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

1. **README.md** – Update if install steps, usage, or config schema changed. Keep examples and env var list accurate.
2. **AGENTS.md** – Update if project structure, setup commands, code style, testing, or CI changed. Keep repo layout and "Documentation and AI context" section accurate.
3. **.cursor/rules/*.mdc** – Update if coding conventions, workflow rules, or file-scoped guidance changed. Match globs to the files the rule applies to.
4. **.cursor/skills/*/SKILL.md** – Update if a documented workflow (e.g. release steps, test commands) or checklist changed.

## Where Things Live

| Artifact | Audience | Purpose |
|----------|----------|---------|
| README.md | Humans | Install, usage, config overview, development commands |
| AGENTS.md | AI agents | Project overview, structure, setup, style, testing, CI, doc-update requirement |
| .cursor/rules/*.mdc | AI agents | File-specific or always-applied rules (Go, CI, config, docs) |
| .cursor/skills/*/SKILL.md | AI agents | Step-by-step workflows (e.g. maintain-documentation, release, testing) |

## When to Update AI Docs

- **New commands or flags** → README and AGENTS.md (structure / usage).
- **New env vars or config keys** → README, .neptune.example.yaml, config-and-yaml rule.
- **New Make targets or CI jobs** → AGENTS.md, ci-and-release rule, testing-and-ci skill.
- **New patterns agents should follow** → Add or update a rule or skill; mention in AGENTS.md if central.

Do not edit plan files (e.g. in .cursor/plans or *.plan.md) unless the user explicitly asks.
