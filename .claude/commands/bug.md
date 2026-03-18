# Create a bug report

Use the **issue-writer** agent (defined in the hub at `.claude/agents/neptune/issue-writer/`) to open a GitHub bug report for this repo.

The user's message after `/bug` is the initial description. Follow the issue-writer workflow: use title format "Area: Short description", gather What happened?, What did you expect?, Did this work before?, How do we reproduce?, and optional environment/platform per `.github/ISSUE_TEMPLATE/0-bug-report.yaml`, ask for more details if needed. Before creating the issue, validate the draft with the **issue-reviewer** subagent; only then run `gh issue create`.
