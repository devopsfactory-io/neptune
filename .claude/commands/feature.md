# Create a feature request

Use the **issue-writer** subagent (`.claude/agents/issue-writer.md`) to open a GitHub feature request for this repo.

The user's message after `/feature` is the initial idea or title. Follow the issue-writer workflow: map it to the template in `.github/ISSUE_TEMPLATE/1-feature-request.md` (Why is this needed?, What would you like to be added?, Who is this feature for?), ask for more details if the description is too short. Before creating the issue, validate the draft with the **issue-reviewer** subagent; only then run `gh issue create`.
