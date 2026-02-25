import logging
from os import getenv
import json
from typing import Optional, Dict, Any, List
from subprocess import Popen, PIPE
from concurrent.futures import ThreadPoolExecutor
from google.cloud import storage # pylint: disable=import-error
from neptune.domain.config import NeptuneConfig
from neptune.domain.lock import (
    LockFile,
    LockStacksDetails,
    LockedStacks,
    WorkflowStatus,
    WorkflowPhase,
    TerraformStacks
)
from neptune.github import GitHubPullRequest
from neptune.utils import CliOutput

LOG_ERROR_PREFIX = "neptune lock error -"
LOG_PREFIX = "neptune lock -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)
logger = logging.getLogger(__name__)

def changed_stacks(config: NeptuneConfig) -> TerraformStacks | CliOutput:
    """Get changed Terraform stacks with Terramate"""
    logger.info("%s Getting changed Terraform stacks with Terramate", LOG_PREFIX)
    with Popen(['terramate', 'list', '--changed', '--run-order'], stdout=PIPE, stderr=PIPE) as proc:
        output, error = proc.communicate()
        return_code = proc.returncode
        if return_code != 0:
            return CliOutput(output=f"Failed to get changed Terraform stacks: {error.decode('utf-8')}", status=1, github_comment=True, config=config)
        # Remove empty lines
        stacks = [stack for stack in output.decode('utf-8').split('\n') if stack]
        return TerraformStacks(stacks=stacks)

class GCSStorage:
    """GCS storage client"""
    bucket_name: str
    parent_folder: str

    def __init__(self, bucket_url: str, parent_folder: str, config: NeptuneConfig): # pylint: disable=return-in-init
        """Initialize the GCS storage client
        Args:
            bucket_url: The GCS bucket URL (e.g. gs://bucket-name)
            parent_folder: The parent folder of the lock files path
        """
        self.config = config
        self.bucket_url = bucket_url
        self.parent_folder = parent_folder
        self.bucket_name = ""
        self.client = None
        self.bucket = None
        self._initialize_client()

    def _initialize_client(self) -> CliOutput | None:
        """Initialize the GCS storage client"""
        if not self.bucket_url.startswith('gs://'):
            return CliOutput(output="Bucket URL must start with gs://", status=1, github_comment=True, config=self.config)
        self.bucket_name = self.bucket_url.replace('gs://', '')
        try:
            logger.info("%s Initializing GCS storage client", LOG_PREFIX)
            self.client = storage.Client()
            self.bucket = self.client.bucket(self.bucket_name)
            self.parent_folder = self.parent_folder.replace('/', '-').replace(':', '').replace('.', '-')
        except Exception as e:
            return CliOutput(output=f"Failed to initialize GCS storage client: {e}", status=1, github_comment=True, config=self.config)

    def get_lock_file(self, stack_path: str) -> Optional[Dict[str, Any]] | CliOutput:
        """Get the lock file for a PR
        Args:
            stack_path: The repository path + the path of the stack
        Returns:
            The lock file contents as a dictionary, or None if not found or a CliOutput if got an error fetching the lock file
        """
        logger.info("%s Getting lock file for stack %s", LOG_PREFIX, stack_path)
        try:
            blob = self.bucket.blob(f"{self.parent_folder}/{stack_path}/lock.json")
            if not blob.exists():
                return None
            content = blob.download_as_string()
            return json.loads(content)
        except Exception as e:
            return CliOutput(output=f"Failed to get lock file for stack {stack_path}: {e}", status=1, github_comment=True, config=self.config)

    def create_or_update_lock_file(self, stack_path: str, lock_data: LockFile) -> None:
        """Create or update a lock file
        Args:
            stack_path: The PR ID
            lock_data: The lock file contents
        """
        logger.info("%s Creating or updating lock file for stack %s", LOG_PREFIX, stack_path)
        try:
            blob = self.bucket.blob(f"{self.parent_folder}/{stack_path}/lock.json")
            blob.upload_from_string(
                json.dumps(lock_data.to_dict(), indent=2),
                content_type='application/json'
            )
        except Exception as e:
            return CliOutput(output=f"Failed to create or update lock file for stack {stack_path}: {e}", status=1, github_comment=True, config=self.config)

    def delete_lock_file(self, stack_path: str) -> None:
        """Delete a lock file
        Args:
            stack_path: The PR ID
        """
        logger.info("%s Deleting lock file for stack %s", LOG_PREFIX, stack_path)
        try:
            blob = self.bucket.blob(f"{self.parent_folder}/{stack_path}/lock.json")
            if blob.exists():
                blob.delete()
        except Exception as e:
            return CliOutput(output=f"Failed to delete lock file for stack {stack_path}: {e}", status=1, github_comment=True, config=self.config)

class LockFileInterface:
    """Lock file interface"""
    def __init__(self, config: NeptuneConfig):
        self.config = config
        self.storage = GCSStorage(config.repository.object_storage, config.repository.github.repository, config)
        self.terraform_stacks = changed_stacks(self.config)  # pylint: disable=no-value-for-parameter
        self.lock_stacks_details = self._get_lock_details()

    def _get_lock_details(self) -> LockStacksDetails | CliOutput:
        logger.info("%s Getting lock details for stacks", LOG_PREFIX)
        if self.terraform_stacks.stacks is None or not self.terraform_stacks.stacks:
            return CliOutput(output="No Terraform stacks found", status=0, github_comment=True, config=self.config)

        lock_stacks_details = [
            {
                "path": stack,
                "lock_file": None,
            }
            for stack in self.terraform_stacks.stacks
        ]

        def _get_lock_file_for_stack(stack_detail):
            stack_detail["lock_file"] = self.storage.get_lock_file(stack_detail["path"])
            return stack_detail

        with ThreadPoolExecutor(max_workers=min(len(lock_stacks_details), 10)) as executor:
            lock_stacks_details = list(executor.map(_get_lock_file_for_stack, lock_stacks_details))

        return LockStacksDetails(details=lock_stacks_details)

    def stacks_locked(self) -> LockedStacks:
        """Check if any stack is locked by a different PR
        Returns:
            LockedStacks: The status of the stacks lock
        """
        logger.info("%s Checking locked status of stacks...", LOG_PREFIX)
        current_pr = self.config.repository.github.pull_request_number

        stacks_lock_status = LockedStacks(locked=False, stack_path=[], prs=[])
        for stack_detail in self.lock_stacks_details.details:
            if stack_detail["lock_file"] is not None:
                locked_by = stack_detail["lock_file"].get("locked_by_pr_id")
                # Check if PR in "locked_by" is not closed or merged, if it is, unlock the stack
                github = GitHubPullRequest(self.config)
                if not github.is_pr_open(locked_by):
                    logger.info("%s PR %s is not open, unlocking stack %s", LOG_PREFIX, locked_by, stack_detail["path"])
                    self.storage.delete_lock_file(stack_detail["path"])
                    locked_by = None
                    continue
                # Check if the PR that locked the stack is the current PR, if it is not, the stack is locked by a different PR and should stop the execution
                if locked_by != current_pr:
                    logger.warning(
                        "%s Stack %s is locked by PR %s",
                        LOG_ERROR_PREFIX,
                        stack_detail["path"],
                        locked_by
                    )
                    stacks_lock_status.locked = True
                    stacks_lock_status.stack_path.append(stack_detail["path"])
                    stacks_lock_status.prs.append(locked_by)
        logger.info("%s Stacks lock status: %s", LOG_PREFIX, stacks_lock_status)
        return stacks_lock_status

    def unlock_all_stacks(self) -> None:
        """Unlock all stacks"""
        logger.info("%s Unlocking all stacks", LOG_PREFIX)
        for stack_detail in self.lock_stacks_details.details:
            self.storage.delete_lock_file(stack_detail["path"])

    def depends_on_completed(self, phase: str) -> bool:
        """Check if the depends_on is completed for a phase"""
        logger.info("%s Checking depends_on for phase %s", LOG_PREFIX, phase)
        depends_on = self.config.workflows.workflows[self.config.repository.allowed_workflow].phases[phase].depends_on
        logger.info("%s Depends on: %s", LOG_PREFIX, depends_on)

        if not depends_on:
            return True
        not_completed = []
        for dependency in depends_on:
            logger.info("%s Checking dependency: %s", LOG_PREFIX, dependency)
            for stack_detail in self.lock_stacks_details.details:
                if stack_detail["lock_file"] is None:
                    not_completed.append(dependency)
                elif dependency not in stack_detail["lock_file"]["workflow_phases"]:
                    not_completed.append(dependency)
                elif stack_detail["lock_file"]["workflow_phases"][dependency]["status"] != WorkflowStatus.COMPLETED:
                    not_completed.append(dependency)

        return len(not_completed) == 0

    def _define_lock_file(self, phase: str, stack_path: str, status: WorkflowStatus) -> LockFile:
        """Define the lock file for a stack, a phase and the status of the phase, this method will consider the current lock file and only update the status of the phase if it exists, otherwide, it will create a new LockFile structure"""
        logger.info("%s Defining lock file for stack %s and phase %s", LOG_PREFIX, stack_path, phase)
        # Check if the lock file exists
        lock_file = self.storage.get_lock_file(stack_path)
        if lock_file is None:
            lock_file = LockFile(
                locked_by_pr_id=self.config.repository.github.pull_request_number,
                workflow_phases={
                    phase: WorkflowPhase(status=status)
                }
            )
        else:
            # if there is other phases, transform it in the WorkflowPhase structure
            for existing_phase, existing_status in lock_file["workflow_phases"].items():
                lock_file["workflow_phases"][existing_phase] = WorkflowPhase(status=WorkflowStatus(existing_status["status"]))
            lock_file["workflow_phases"][phase] = WorkflowPhase(status=status)
            lock_file = LockFile(**lock_file)
        return lock_file

    def lock_stacks(self, phase: str, stacks: List[str], status: WorkflowStatus) -> None:
        """Lock the stacks for a phase and a status in parallel"""
        logger.info("%s Locking stacks for phase %s and status %s", LOG_PREFIX, phase, status)
        with ThreadPoolExecutor(max_workers=min(len(stacks), 10)) as executor:
            futures = [
                executor.submit(
                    lambda s: self.storage.create_or_update_lock_file(
                        stack_path=s,
                        lock_data=self._define_lock_file(phase=phase, stack_path=s, status=status)
                    ),
                    stack
                )
                for stack in stacks
            ]
            # Wait for all futures to complete
            for future in futures:
                future.result()  # This will raise any exceptions that occurred

    def _update_lock_status(self, phase: str, stack_path: str, status: WorkflowStatus) -> None:
        """Update the status of a phase for a stack"""
        logger.info("%s Updating lock status for stack %s and phase %s", LOG_PREFIX, stack_path, phase)
        lock_file = self._define_lock_file(phase=phase, stack_path=stack_path, status=status)
        self.storage.create_or_update_lock_file(stack_path=stack_path, lock_data=lock_file)

    def update_stacks(self, phase: str, stacks: List[str], status: WorkflowStatus) -> None:
        """Update the status of a phase for a list of stacks in parallel"""
        logger.info("%s Updating lock status for stacks %s and phase %s", LOG_PREFIX, stacks, phase)
        with ThreadPoolExecutor(max_workers=min(len(stacks), 10)) as executor:
            futures = [
                executor.submit(self._update_lock_status, phase=phase, stack_path=stack, status=status)
                for stack in stacks
            ]
            # Wait for all futures to complete
            for future in futures:
                future.result()  # This will raise any exceptions that occurred
