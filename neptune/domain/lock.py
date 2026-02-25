import logging
from os import getenv
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum

class WorkflowStatus(str, Enum):
    """Workflow status
    - IN_PROGRESS: The workflow is in progress
    - PENDING: The workflow is pending
    - COMPLETED: The workflow is completed
    """
    IN_PROGRESS = "in_progress"
    PENDING = "pending"
    COMPLETED = "completed"

@dataclass
class WorkflowPhase:
    """Workflow phase
    - status: The status of the workflow phase
    
    Additional methods:
    - to_dict: Convert the workflow phase to a dictionary
    """
    status: WorkflowStatus

    def to_dict(self) -> Dict[str, Any]:
        """Convert the workflow phase to a dictionary"""
        return {
            "status": self.status.value
        }

@dataclass
class LockFile:
    """Lock file
    - locked_by_pr_id: The ID of the PR that locked the file
    - workflow_phases: The phases of the workflow

    Additional methods:
    - to_dict: Convert the lock file to a dictionary
    """
    locked_by_pr_id: str
    workflow_phases: Dict[str, WorkflowPhase]

    def to_dict(self) -> Dict[str, Any]:
        """Convert the lock file to a dictionary"""
        return {
            "locked_by_pr_id": self.locked_by_pr_id,
            "workflow_phases": {
                phase: workflow_phase.to_dict()
                for phase, workflow_phase in self.workflow_phases.items()
            }
        }

@dataclass
class LockedStacks:
    """Locked stacks
    - locked: Whether the stacks are locked
    - stack_path: The path to the stack
    - prs: The PRs that locked the stacks
    """
    locked: bool
    stack_path: List[str]
    prs: List[str]

@dataclass
class LockStacksDetails:
    """Lock stacks details
    - details: A list of dictionaries containing the path and lock file of the stack
    """
    details: List[
        Dict[
            "path": str,
            "lock_file": Optional[LockFile],
        ]
    ]

@dataclass
class TerraformStacks:
    """Terraform stacks
    - stacks: A list of strings containing the paths to the stacks
    """
    stacks: List[str]
