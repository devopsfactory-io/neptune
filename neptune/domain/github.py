from typing import List, Dict, Optional
from dataclasses import dataclass
from neptune.domain.run import StepsOutput
from neptune.domain.lock import TerraformStacks

@dataclass
class PRRequirementsStatus:
    """Status of PR requirements check
    - is_compliant: Whether the PR is compliant with the requirements
    - failed_requirements: The requirements that failed
    - error_message: The error message if the requirements check failed
    """
    is_compliant: bool
    failed_requirements: List[str]
    error_message: str

@dataclass
class PRInfo:
    """Information about a PR
    - response: The response from the GitHub API
    - pr_number: The number of the PR
    - repo: The repository name
    - api_url: The API URL for the repository
    """
    response: Dict
    pr_number: int
    repo: str
    api_url: str

@dataclass
class PullRequestComment:
    """
    A definition of a comment on a pull request
    - steps_output: The StepsOutput that will be used to format the comment
    - stacks: The TerraformStacks that will be used to format the comment
    - simple_output: A single string that will be used as a primary comment for simple comments
    """
    steps_output: Optional[StepsOutput] = None
    stacks: Optional[TerraformStacks] = None
    simple_output: Optional[str] = None
    overall_status: Optional[int] = None
