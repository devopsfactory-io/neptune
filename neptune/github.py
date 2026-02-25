"""GitHub interaction utilities for Neptune CLI."""
import logging
from os import getenv
from typing import Optional, List
import requests # pylint: disable=import-error
import neptune.git as neptune_git
from neptune.utils import CliOutput
from neptune.domain.github import PRRequirementsStatus, PRInfo
from neptune.notifications.github import GitHubAPI

LOG_ERROR_PREFIX = "neptune github error -"
LOG_PREFIX = "neptune github -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

class GitHubPullRequest(GitHubAPI):
    """GitHub pull request utilities for Neptune CLI."""
    def _get_pr_info(self) -> Optional[PRInfo] | CliOutput:
        """Get PR information from GitHub API."""
        logger.info("%s Getting PR info", LOG_PREFIX)
        if not all([self.token, self.repo, self.pr_number]):
            return CliOutput(output="GITHUB_TOKEN, GITHUB_REPOSITORY, and GITHUB_PULL_REQUEST_NUMBER are required", status=1)

        url = f"{self.api_url}/repos/{self.repo}/pulls/{self.pr_number}"
        response = requests.get(url, headers=self.headers, timeout=10)
        if response.status_code != 200:
            return CliOutput(output=f"Failed to get PR info: {response.json()}", status=1, github_comment=True, config=self.config)
        self.pr_info = PRInfo(
            response=response.json(),
            pr_number=self.pr_number,
            repo=self.repo,
            api_url=self.api_url
        )
        return self.pr_info

    def is_pr_open(self, pr_number: str) -> bool | CliOutput:
        """Check if a PR is open
        Args:
            pr_number: The PR number
        Returns:
            bool | CliOutput: True if the PR is open, False otherwise or a CliOutput if got an error fetching the PR info
        """
        url = f"{self.api_url}/repos/{self.repo}/pulls/{pr_number}"
        response = requests.get(url, headers=self.headers, timeout=10)
        if response.status_code != 200:
            return CliOutput(output=f"Failed to get PR info: {response.json()}", status=1, github_comment=True, config=self.config)
        pr_data = response.json()
        return pr_data["state"] == "open"

    def check_requirements(self, requirements: List[str]) -> PRRequirementsStatus:
        """
        Check if PR meets the specified requirements.

        Args:
            requirements: List of requirements to check ('approved', 'mergeable', 'undiverged')

        Returns:
            PRRequirementsStatus with check results
        """
        logger.info("%s Checking requirements for PR %s", LOG_PREFIX, self.pr_number)
        if not requirements:
            return PRRequirementsStatus(True, [], "")

        pr_info = self._get_pr_info()
        if not pr_info:
            return PRRequirementsStatus(
                False, 
                requirements,
                "Could not fetch PR information. Make sure GITHUB_TOKEN is set and has access to the repository."
            )

        return self._get_pr_requirements_info(requirements, pr_info)

    def _get_pr_requirements_info(self, requirements: List[str], pr_info: PRInfo) -> PRRequirementsStatus:
        """Get PR requirements information from PR info and GitHub API"""
        logger.info("%s Getting PR requirements information", LOG_PREFIX)
        failed = []
        for req in requirements:
            logger.info("%s Checking requirement: %s", LOG_PREFIX, req)
            if req == "approved":
                # Check PR approval status
                logger.info("%s Checking PR approval status...", LOG_PREFIX)
                reviews_url = f"{self.api_url}/repos/{self.repo}/pulls/{self.pr_number}/reviews"
                reviews = requests.get(reviews_url, headers=self.headers, timeout=10)
                if reviews.status_code != 200:
                    return PRRequirementsStatus(
                        False,
                        requirements,
                        f"Could not fetch PR reviews: {reviews.json()}"
                    )
                reviews_data = reviews.json()
                logger.info("%s Reviews: %s", LOG_PREFIX, reviews_data)
                approved = any(
                    review["state"].lower() == "approved"
                    for review in reviews_data
                )
                if not approved:
                    failed.append(req)
                    logger.warning("%s PR is not approved...", LOG_PREFIX)
            elif req == "mergeable":
                # Check if PR is mergeable
                logger.info("%s Checking PR mergeability", LOG_PREFIX)
                if not pr_info.response.get("mergeable"):
                    failed.append(req)
                    logger.warning("%s PR is not mergeable...", LOG_PREFIX)
            elif req == "undiverged":
                # Check if PR branch is up to date with base
                logger.info("%s Checking PR branch is up to date with base", LOG_PREFIX)
                if pr_info.response.get("mergeable_state") == "behind":
                    failed.append(req)
                    logger.warning("%s PR branch is not up to date with base...", LOG_PREFIX)
            elif req == "rebased":
                # Check if PR branch is rebased
                logger.info("%s Checking PR branch is rebased...", LOG_PREFIX)
                if not neptune_git.is_branch_rebased(self.config):
                    failed.append(req)
                    logger.warning("%s PR branch is not rebased...", LOG_PREFIX)

        logger.info("%s PR requirements collected", LOG_PREFIX)
        return PRRequirementsStatus(
            is_compliant=len(failed) == 0,
            failed_requirements=failed,
            error_message="" if len(failed) == 0 else "PR does not meet the following requirements: " + ", ".join(failed)
        )
