"""GitHub notifications utilities for Neptune CLI."""
import logging
from os import getenv
import re
import requests # pylint: disable=import-error
from neptune.domain.config import NeptuneConfig
from neptune.domain.run import StepsOutput
from neptune.domain.lock import TerraformStacks
from neptune.domain.github import PullRequestComment

LOG_ERROR_PREFIX = "neptune notifications.github error -"
LOG_PREFIX = "neptune notifications.github -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

class GitHubAPI:
    """GitHub API common configuration for Neptune CLI."""
    def __init__(self, config: NeptuneConfig):
        """Initialize the GitHub API."""
        logger.info("%s Initializing GitHub API", LOG_PREFIX)
        self.config = config
        self.token = self.config.repository.github.token
        self.repo = self.config.repository.github.repository.replace("https://github.com/", "")
        self.pr_number = self.config.repository.github.pull_request_number
        self.pr_info = None
        self.api_url = "https://api.github.com"
        self.headers = {
            "Authorization": f"token {self.token}",
            "Accept": "application/vnd.github.v3+json"
        } if self.token else {}
        self.stacks = TerraformStacks(stacks=[])

class GitHubNotifications(GitHubAPI):
    """GitHub notifications utilities for Neptune CLI."""
    default_header = """### 🌊 Neptune Execution Results"""
    body_limit = 65536 - 2048 # 2048 is a guardrail to avoid the comment being too big

    @staticmethod
    def _strip_ansi_codes(text: str) -> str:
        """Remove ANSI escape codes from text."""
        ansi_escape = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')
        return ansi_escape.sub('', text)

    @staticmethod
    def _truncate_output(text: str, limit: int = 20000) -> str:
        """Truncate the stdout output to a certain number of characters."""
        if len(text) > limit:
            truncated_const = "...[truncated]"
            return text[:limit-len(truncated_const)] + truncated_const
        return text

    def _format_custom_output(self, comment: PullRequestComment) -> str:
        """Format custom output as a GitHub comment."""
        logger.info("%s Formatting custom output for PR %s", LOG_PREFIX, self.pr_number)
        logger.debug("%s Comment object for custom output: %s", LOG_PREFIX, comment)

        error_message = ""
        if comment.overall_status != 0 and comment.simple_output is not None:
            error_message = f"""
An error occurred:
- {comment.simple_output}
"""

        stacks_comment = ""
        if comment.stacks is not None:
            stacks_comment = f"""
**Terraform Stacks:** `{", ".join(comment.stacks.stacks)}`
"""

        overall_status_comment = ""
        if comment.overall_status is not None:
            overall_status_comment = f"""
**Neptune completed the {comment.steps_output.phase} workflow with status:** {"✅" if comment.overall_status == 0 else "❌"}
> For more details, see the [GitHub Actions run](https://github.com/{self.repo}/actions/runs/{self.config.repository.github.run_id})
"""

        steps_output_comment = ""
        if comment.steps_output is not None:
            for output in comment.steps_output.outputs:
                cleaned_error = self._truncate_output(
                    self._strip_ansi_codes(output.error),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment))
                cleaned_output = self._truncate_output(
                    self._strip_ansi_codes(output.output),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment) - len(cleaned_error))
                steps_output_comment += f"""
- **Command {"✅" if output.status == 0 else "❌"}** `{output.command}`
<details>
<summary>Click to see the command output</summary>

```
stderr:
{cleaned_error}

stdout:
{cleaned_output}
```

</details>

"""

        return f"""{self.default_header}
{error_message}{stacks_comment}{overall_status_comment}{steps_output_comment}
"""

    def _format_plan_output(self, comment: PullRequestComment) -> str:
        """Format terraform plan output as a GitHub comment."""
        logger.info("%s Formatting plan output for PR %s", LOG_PREFIX, self.pr_number)
        self.default_header = """### 🌊 Neptune Plan Results"""

        stacks_comment = ""
        if comment.stacks is not None:
            stacks_comment = f"""
**Terraform Stacks:** `{", ".join(comment.stacks.stacks)}`
"""

        overall_status_comment = ""
        if comment.overall_status is not None:
            overall_status_comment = f"""
**Neptune completed the plan with status:** {"✅" if comment.overall_status == 0 else "❌"}
> For more details, see the [GitHub Actions run](https://github.com/{self.repo}/actions/runs/{self.config.repository.github.run_id})
"""

        steps_output_comment = ""
        if comment.steps_output is not None:
            for output in comment.steps_output.outputs:
                cleaned_error = self._truncate_output(
                    self._strip_ansi_codes(output.error),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment))
                cleaned_output = self._truncate_output(
                    self._strip_ansi_codes(output.output),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment) - len(cleaned_error))
                steps_output_comment += f"""
- **Command {"✅" if output.status == 0 else "❌"}** `{output.command}`
<details>
<summary>Click to see the command output</summary>

```
stderr:
{cleaned_error}

stdout:
{cleaned_output}
```

</details>

"""

        return f"""{self.default_header}
{stacks_comment}{overall_status_comment}{steps_output_comment}


To apply these changes, comment:
```
/neptune apply
```
"""

    def _format_apply_output(self, comment: PullRequestComment) -> str:
        """Format terraform apply output as a GitHub comment."""
        logger.info("%s Formatting apply output for PR %s", LOG_PREFIX, self.pr_number)
        self.default_header = """### 🌊 Neptune Apply Results"""

        stacks_comment = ""
        if comment.stacks is not None:
            stacks_comment = f"""
**Terraform Stacks:** `{", ".join(comment.stacks.stacks)}`
"""

        overall_status_comment = ""
        if comment.overall_status is not None:
            overall_status_comment = f"""
**Neptune completed the apply with status:** {"✅" if comment.overall_status == 0 else "❌"}
> For more details, see the [GitHub Actions run](https://github.com/{self.repo}/actions/runs/{self.config.repository.github.run_id})
"""

        steps_output_comment = ""
        if comment.steps_output is not None:
            for output in comment.steps_output.outputs:
                cleaned_error = self._truncate_output(
                    self._strip_ansi_codes(output.error),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment))
                cleaned_output = self._truncate_output(
                    self._strip_ansi_codes(output.output),
                    limit=self.body_limit - len(stacks_comment) - len(overall_status_comment) - len(cleaned_error))
                steps_output_comment += f"""
- **Command {"✅" if output.status == 0 else "❌"}** `{output.command}`
<details>
<summary>Click to see the command output</summary>

```
stderr:
{cleaned_error}

stdout:
{cleaned_output}
```

</details>

"""

        return f"""{self.default_header}
{stacks_comment}{overall_status_comment}{steps_output_comment}"""

    def create_comment(self, comment: PullRequestComment) -> None:
        """Add a comment to the current PR."""
        logger.info("%s Creating comment on PR %s", LOG_PREFIX, self.pr_number)
        if comment.stacks is not None:
            self.stacks = comment.stacks.stacks

        if not all([self.token, self.repo, self.pr_number]):
            return

        if comment.steps_output is None:
            comment.steps_output = StepsOutput(phase="custom", overall_status=comment.overall_status, outputs=[])
        match comment.steps_output.phase:
            case "plan":
                formatted_comment = self._format_plan_output(comment = comment)
            case "apply":
                formatted_comment = self._format_apply_output(comment = comment)
            case _:
                formatted_comment = self._format_custom_output(comment = comment)

        url = f"{self.api_url}/repos/{self.repo}/issues/{self.pr_number}/comments"
        response = requests.post(url, headers=self.headers, json={"body": formatted_comment}, timeout=10)

        if response.status_code != 201:
            logger.error("%s Failed to create comment on PR %s", LOG_PREFIX, self.pr_number)
            logger.error("%s Response: %s", LOG_PREFIX, response.text)
            response.raise_for_status()
            return
        logger.info("%s Comment created on PR %s", LOG_PREFIX, self.pr_number)
