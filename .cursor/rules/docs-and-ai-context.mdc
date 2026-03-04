---
description: After any code/config/CI change, delegate to the documentation-maintainer subagent to run the doc checklist. When in doubt, update.
alwaysApply: true
---

# Documentation and AI Context

After making changes to code, config, or CI:

1. **Delegate**: Delegate documentation updates to the **documentation-maintainer** subagent so it can run the full maintain-documentation checklist (README, docs/, examples/, AGENTS.md, .cursor/rules, .cursor/commands, .cursor/skills).
2. **Do not edit plan files** (e.g. `neptune_go_rewrite*.plan.md`, `ai_agent_config*.plan.md`) unless the user explicitly asks.

The **documentation-maintainer** subagent (`.cursor/agents/documentation-maintainer.md`) embodies the project skill **maintain-documentation** (`.cursor/skills/maintain-documentation/`); use the subagent for the full checklist and when to update each artifact.
