import logging
from os import getenv
import typer # pylint: disable=import-error
from rich.console import Console # pylint: disable=import-error
from rich.panel import Panel # pylint: disable=import-error
from rich import print # pylint: disable=redefined-builtin, import-error
from typing_extensions import Annotated
from neptune.config import Config
from neptune.utils import CliOutput
from neptune.lock import LockFileInterface
from neptune.github import GitHubPullRequest
from neptune.run import RunSteps
from neptune.domain.lock import WorkflowStatus
from neptune.domain.github import PullRequestComment

LOG_ERROR_PREFIX = "neptune cli error -"
LOG_PREFIX = "neptune cli -"

VERSION = "0.1.0"
logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

app = typer.Typer(
    name="neptune",
    help="Neptune CLI - Terraform pull request automation tool inspired by Atlantis",
    add_completion=True,
    no_args_is_help=True
)
console = Console()

@app.command()
def version():
    """Print the version of the tool."""
    print("[bold white]neptune version: " + VERSION + "[bold white]")
    raise typer.Exit(code=0)

@app.command()
def command(
    workflow: Annotated[str, typer.Argument(help="The workflow phase to run")]
) -> CliOutput:
    """Run a workflow phase"""
    console.print(Panel.fit(
        f"🌊 [bold]Neptune is running:[/bold] {workflow}",
        title="Neptune Command",
        border_style="blue"
    ))
    config = Config().load()
    config_summary = f"""
- [bold]Repository:[/bold] {config.repository.github.repository}
- [bold]Pull request branch:[/bold] {config.repository.github.pull_request_branch}
- [bold]Pull request number:[/bold] {config.repository.github.pull_request_number}
- [bold]Object storage:[/bold] {config.repository.object_storage}
- [bold]Plan requirements:[/bold] {', '.join(config.repository.plan_requirements)}
- [bold]Apply requirements:[/bold] {', '.join(config.repository.apply_requirements)}
- [bold]Allowed workflow:[/bold] {config.repository.allowed_workflow}
- [bold]Workflow phases:[/bold] {', '.join(config.workflows.workflows[config.repository.allowed_workflow].phases.keys())}
"""
    console.print(Panel.fit(config_summary, title="🌊 Neptune Config Summary", border_style="blue"))

    # Check if the workflow phase is valid in the allowed workflow
    if workflow not in config.workflows.workflows[config.repository.allowed_workflow].phases.keys():
        return CliOutput(output=f"Workflow {workflow} is not valid, check the allowed workflow in the Neptune config", status=1, github_comment=True, config=config)
    logger.info("%s Workflow %s is valid", LOG_PREFIX, workflow)

    # Check PR requirements based on workflow phase
    match workflow:
        case "plan":
            requirements = config.repository.plan_requirements
        case "apply":
            requirements = config.repository.apply_requirements
        case _:
            requirements = []
    github = GitHubPullRequest(config)
    requirements_status = github.check_requirements(requirements)
    if not requirements_status.is_compliant:
        return CliOutput(
            output=f"Cannot run {workflow} workflow: {requirements_status.error_message}",
            status=1,
            github_comment=True,
            config=config
        )
    logger.info("%s PR requirements check passed for workflow %s", LOG_PREFIX, workflow)
    console.print(Panel.fit(
        f"🌊 [bold]PR requirements ({', '.join(requirements)}) check passed for workflow:[/bold] {workflow}",
        title="🌊 Neptune Plan/Apply Requirements Check",
        border_style="blue"
    ))

    # Check if the stacks are locked by other PRs
    locks = LockFileInterface(config)
    console.print(Panel.fit(
        f"🔒 [bold]Neptune is considering the following stacks in the current PR:[/bold] {', '.join(locks.terraform_stacks.stacks)}",
        title="🌊 Neptune Lock",
        border_style="blue"
    ))

    # If any stack is locked by other PRs, return an error
    locked_stacks = locks.stacks_locked()
    if locked_stacks.locked:
        return CliOutput(output=f"Some stacks ({', '.join(locked_stacks.stack_path)}) are locked by other PRs: {', '.join(set(locked_stacks.prs))}", status=1, github_comment=True, config=config)

    # Check if the workflow depends_on is met
    if not locks.depends_on_completed(phase=workflow):
        return CliOutput(output=f"Dependency is not met for phase {workflow}", status=1, github_comment=True, config=config)

    # Lock the stacks for the workflow phase
    locks.lock_stacks(phase=workflow, stacks=locks.terraform_stacks.stacks, status=WorkflowStatus.PENDING)

    # Execute the workflow steps
    runner = RunSteps(config=config, phase=workflow, locks=locks)
    steps_output = runner.execute()

    # Create a Summary of the steps output
    string_steps_outputs = [output.to_string() for output in steps_output.outputs]
    steps_summary = f"""
- [bold]Phase:[/bold] {steps_output.phase}
- [bold]Steps:[/bold] {chr(10).join(string_steps_outputs)}
"""
    console.print(Panel.fit(steps_summary, title="🌊 Neptune Steps Summary", border_style="blue"))

    result = CliOutput(
        output=f"Workflow {workflow} completed with status {steps_output.overall_status}",
        status=steps_output.overall_status,
        github_comment=True,
        config=config,
        comment=PullRequestComment(
            overall_status=steps_output.overall_status,
            steps_output=steps_output,
            stacks=locks.terraform_stacks,
        )
    )

    return result

@app.command()
def unlock(
    all_stacks: bool = typer.Option(False, "--all", "-a", help="Unlock all stacks (default: False)")
) -> CliOutput:
    """Unlock all stacks"""
    if not all_stacks:
        return CliOutput(output="You need to use the flag --all to run this command", status=1)
    config = Config().load()
    locks = LockFileInterface(config)
    locks.unlock_all_stacks()
    return CliOutput(output="All changed stacks unlocked", status=0)

if __name__ == "__main__":
    app()
