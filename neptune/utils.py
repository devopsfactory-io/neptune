from dataclasses import dataclass
from typing import Optional
import typer # pylint: disable=import-error
from rich import print # pylint: disable=redefined-builtin, import-error
from neptune.notifications.github import GitHubNotifications
from neptune.domain.config import NeptuneConfig
from neptune.domain.github import PullRequestComment

@dataclass
class CliOutput:
    """CLI output
    - output: The output of the command (stderr or stdout), if comment is not provided, this will be used to create the comment
    - status: The status of the command (0 for success, >=1 for error)
    - github_comment: Whether to create a comment in the PR (default: False)
    - config: The configuration of the Neptune CLI (default: None)
    - comment: The comment to be created in the PR using the struct.github.PullRequestComment (default: None)
    """
    output: str
    status: int
    github_comment: Optional[bool] = False
    config: Optional[NeptuneConfig] = None
    comment: Optional[PullRequestComment] = None

    def __init__(
        self,
        output: str,
        status: int,
        github_comment: bool = False,
        config: Optional[NeptuneConfig] = None,
        comment: Optional[PullRequestComment] = None
    ):
        self.output = output
        self.status = status
        self.config = config
        self.github_comment = github_comment
        if self.github_comment and not (self.comment or self.output):
            print("[bold red]Error:[/bold red] Comment is required when github_comment is True, this is usually a bug in the code")
            raise typer.Exit(code=1)
        # If no comment is provided, create a default one with the output
        if comment:
            self.comment = comment
        else:
            self.comment = PullRequestComment(
                simple_output=self.output,
                steps_output=None,
                overall_status=self.status,
            )
        self._notify()
        self._exit()

    def _notify(self):
        if self.github_comment:
            github = GitHubNotifications(self.config)
            github.create_comment(comment=self.comment)

    def _exit(self):
        if self.status != 0:
            print(f"[bold red]Error:[/bold red] {self.output}")
            raise typer.Exit(code=self.status)

        print(f"[bold green]Success:[/bold green] {self.output}")
        raise typer.Exit(code=0)
