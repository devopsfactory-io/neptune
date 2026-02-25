from dataclasses import dataclass
from typing import List, Optional, Dict

@dataclass
class WorkflowStep:
    """A single step in a workflow
    - run: The command to run in the step
    """
    run: str

@dataclass
class WorkflowPhase:
    """A phase in a workflow, e.g. plan or apply
    - steps: The steps to run in the phase(WorkflowStep)
    - depends_on: The phases that must be completed before this phase can run
    """
    steps: List[WorkflowStep]
    depends_on: Optional[List[str]] = None

@dataclass
class WorkflowStatement:
    """Full definition of a single workflow with a dictionary of phases
    - name: The name of the workflow
    - phases: A dictionary of phases(WorkflowPhase)
    """
    name: str
    phases: Dict[str, WorkflowPhase]

@dataclass
class Workflows:
    """A dictionary of workflows
    - workflows: A dictionary of workflows(WorkflowStatement)
    """
    workflows: Dict[str, WorkflowStatement]

@dataclass
class GitHubConfig:
    """Configuration for the GitHub
    - repository: The repository name
    - pull_request_branch: The branch name of the pull request
    - pull_request_number: The number of the pull request
    - pull_request_comment_id: The ID of the pull request comment
    - token: The token to use for the GitHub API
    - run_id: The ID of the GitHub run
    """
    repository: str
    pull_request_branch: str
    pull_request_number: str
    pull_request_comment_id: str
    run_id: str
    token: str

@dataclass
class RepositoryConfig:
    """Configuration for the repository
    - object_storage: The object storage bucket name
    - branch: The branch name of the repository
    - plan_requirements: The requirements for the plan workflow
    - apply_requirements: The requirements for the apply workflow
    - allowed_workflow: The allowed workflow to run
    - github: The GitHub configuration(GitHubConfig)
    """
    object_storage: str
    branch: str
    plan_requirements: List[str]
    apply_requirements: List[str]
    allowed_workflow: str
    github: GitHubConfig

@dataclass
class NeptuneConfig:
    """Configuration for the Neptune CLI
    - repository: The repository configuration(RepositoryConfig)
    - workflows: The workflows configuration(Workflows)
    """
    repository: RepositoryConfig
    workflows: Workflows
