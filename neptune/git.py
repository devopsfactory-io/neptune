"""Git interaction utilities for Neptune CLI."""
import logging
from os import getenv
from git import Repo
from neptune.domain.config import NeptuneConfig

LOG_ERROR_PREFIX = "neptune git error -"
LOG_PREFIX = "neptune git -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

def is_branch_rebased(config: NeptuneConfig) -> bool:
    """Check if the current branch is rebased on the default branch"""
    # git rev-list --count HEAD..origin/master
    try:
        logger.debug("%s Checking if branch is rebased...", LOG_PREFIX)
        repo = Repo('.')
        head_commit = repo.head.commit
        default_branch = repo.refs[f"origin/{config.repository.branch}"]

        # Count commits that are in default branch but not in current branch (commits we're missing)
        commits_behind = list(repo.iter_commits(f"{head_commit}..{default_branch.commit}"))
        logger.debug("%s Commits behind: %s", LOG_PREFIX, commits_behind)
        return len(commits_behind) == 0
    except Exception as e:
        logger.error("%s Failed to check if branch is rebased: %s", LOG_ERROR_PREFIX, str(e))
        return False
