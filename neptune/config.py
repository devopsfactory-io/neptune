'''
Neptune CLI Configuration

This module is responsible for loading the configuration for the Neptune CLI, including:
- Environment variables from CI environment
- Configuration from .neptune.yaml file provided by the ENV variable: NEPTUNE_CONFIG_PATH
'''

import logging
from os import getenv
import os
from pathlib import Path
from typing import Dict
import yaml # pylint: disable=import-error
from neptune.utils import CliOutput
from neptune.domain.config import (
    RepositoryConfig,
    Workflows,
    WorkflowStatement,
    WorkflowPhase,
    WorkflowStep,
    GitHubConfig,
    NeptuneConfig
)

LOG_ERROR_PREFIX = "neptune config error -"
LOG_PREFIX = "neptune config -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

class Config:
    """Configuration for the Neptune CLI"""
    def __init__(self):
        self.repository = RepositoryConfig(
            object_storage=None,
            branch=None,
            plan_requirements=None,
            apply_requirements=None,
            allowed_workflow=None,
            github=GitHubConfig(
                repository=None,
                pull_request_branch=None,
                pull_request_number=None,
                pull_request_comment_id=None,
                run_id=None,
                token=None)
        )
        self.workflows = Workflows(
            workflows=None
        )

    @classmethod
    def _load_env_vars(cls) -> Dict[str, str]:
        """
        Load environment variables from the environment.
        """
        logger.info("%s Loading environment variables", LOG_PREFIX)
        env_vars = {
            'NEPTUNE_CONFIG_PATH': os.getenv('NEPTUNE_CONFIG_PATH', '.neptune.yaml'),
            'GITHUB_REPOSITORY': os.getenv('GITHUB_REPOSITORY', ''),
            'GITHUB_PULL_REQUEST_BRANCH': os.getenv('GITHUB_PULL_REQUEST_BRANCH', ''),
            'GITHUB_PULL_REQUEST_NUMBER': os.getenv('GITHUB_PULL_REQUEST_NUMBER', ''),
            'GITHUB_PULL_REQUEST_COMMENT_ID': os.getenv('GITHUB_PULL_REQUEST_COMMENT_ID', ''),
            'GITHUB_RUN_ID': os.getenv('GITHUB_RUN_ID', ''),
            'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN', ''),
        }
        required_env_vars = [
            'NEPTUNE_CONFIG_PATH',
            'GITHUB_REPOSITORY',
            'GITHUB_PULL_REQUEST_BRANCH',
            'GITHUB_PULL_REQUEST_NUMBER',
            'GITHUB_PULL_REQUEST_COMMENT_ID',
            'GITHUB_RUN_ID',
            'GITHUB_TOKEN',
        ]
        
        missing_vars = [var for var in required_env_vars if not env_vars[var]]
        if missing_vars:
            raise ValueError(f"Environment variables {', '.join(missing_vars)} are required")

        return env_vars

    @classmethod
    def _check_config_options(cls) -> None:
        """Check if the NeptuneConfig options are valid"""
        logger.info("%s Checking config options", LOG_PREFIX)
        plan_apply_requirements = [
            "undiverged",
            "approved",
            "mergeable",
            "rebased",
        ]

        # check if object_storage is not null and if it is a valid URL
        if cls.repository.object_storage is None or cls.repository.object_storage == '':
            raise ValueError("Repository object storage locking is required")
        if cls.repository.object_storage is not None:
            if not cls.repository.object_storage.startswith('gs://'):
                raise ValueError("Repository object storage must be a valid GCS URL")
        # if plan_requirements is not null, check if it is in plan_apply_requirements
        if cls.repository.plan_requirements is not None:
            for requirement in cls.repository.plan_requirements:
                if requirement not in plan_apply_requirements:
                    raise ValueError(f"Repository plan requirements must be one of: {plan_apply_requirements}")
        # if apply_requirements is not null, check if it is in plan_apply_requirements
        if cls.repository.apply_requirements is not None:
            for requirement in cls.repository.apply_requirements:
                if requirement not in plan_apply_requirements:
                    raise ValueError(f"Repository apply requirements must be one of: {plan_apply_requirements}")

        # check if allowed_workflow is not null and if it is has at least one workflow in cls.workflows.workflows.keys()
        if cls.repository.allowed_workflow is None or cls.repository.allowed_workflow == '':
            raise ValueError("Repository allowed workflow is required")
        if cls.repository.allowed_workflow is not None:
            if cls.repository.allowed_workflow == "":
                raise ValueError("Repository allowed workflows must have at least one workflow")
            if cls.repository.allowed_workflow not in cls.workflows.workflows.keys():
                raise ValueError(f"Repository allowed workflows must be one of: {cls.workflows.workflows.keys()}")

        # check if the repository.branch is not None and if it is not equals to the pull request branch (the workflows should be executed in the pull request branch)
        if cls.repository.branch is None or cls.repository.branch == '':
            raise ValueError("Repository branch is required, check the GitHub Action configuration")
        if cls.repository.branch is not None:
            if cls.repository.branch == cls.repository.github.pull_request_branch:
                raise ValueError("The `repository.branch` (default branch) should not be used to execute the workflows, check the GitHub Action configuration, the workflows should be executed in the pull request branch")

        for workflow in cls.workflows.workflows.values():
            for phase_name, phase in workflow.phases.items():
                # if phase has depends_on, check if the depends_on is in the phases.keys()
                if phase.depends_on is not None:
                    for depends_on in phase.depends_on:
                        if depends_on not in workflow.phases.keys():
                            raise ValueError(f"Phase {phase_name} depends on {depends_on}, but {depends_on} is not a valid workflow phase")
                # phases.keys() should include at least plan and apply
                if "plan" not in workflow.phases.keys() or "apply" not in workflow.phases.keys():
                    raise ValueError("Phases should include at least plan and apply phases, check the workflow configuration")
                for step in phase.steps:
                    # steps should have at least one step
                    if step.run is None or step.run == '':
                        raise ValueError("At least one step is required in each phase")
                    if step.run is not None:
                        # if step.run contains terragrunt or terraform, check if the step.run has both terramate and --changed flag
                        if "terragrunt" in step.run or "terraform" in step.run:
                            if not ("terramate" in step.run and "--changed" in step.run):
                                raise ValueError("The step run must use both the `terramate` command AND the `--changed` flag when using `terragrunt` or `terraform`")

    @classmethod
    def load(cls) -> NeptuneConfig | CliOutput:
        """
        Load Neptune configuration from environment variables and config file.
        
        Returns:
            NeptuneConfig: The loaded configuration object
        
        Raises:
            ValueError: If the config file is invalid or required env vars are missing
        """
        try:
            env_vars = cls._load_env_vars()
        except ValueError as e:
            return CliOutput(output=str(e), status=1)

        config_file = Path(env_vars['NEPTUNE_CONFIG_PATH'])
        if not config_file.exists():
            return CliOutput(output=f"Config file not found: {env_vars['NEPTUNE_CONFIG_PATH']}", status=1)
        
        with open(config_file, 'r', encoding='utf-8') as f:
            logger.info("%s Loading config file", LOG_PREFIX)
            config_data = yaml.safe_load(f)
        
        #Parse GitHub config
        try:
            logger.debug("%s Parsing GitHub config from environment variables", LOG_PREFIX)
            github_config = GitHubConfig(
                repository=env_vars.get('GITHUB_REPOSITORY', ''),
                pull_request_branch=env_vars.get('GITHUB_PULL_REQUEST_BRANCH', ''),
                pull_request_number=env_vars.get('GITHUB_PULL_REQUEST_NUMBER', ''),
                pull_request_comment_id=env_vars.get('GITHUB_PULL_REQUEST_COMMENT_ID', ''),
                run_id=env_vars.get('GITHUB_RUN_ID', ''),
                token=env_vars.get('GITHUB_TOKEN', ''))
        except Exception as e:
            # Should not comment on PR, because the config is not valid
            return CliOutput(output=f"Error parsing GitHub config: {e}", status=1, github_comment=False)
        
        try:
            # Parse repository config
            logger.debug("%s Parsing repository config", LOG_PREFIX)
            repo_data = config_data.get('repository', {})
            cls.repository = RepositoryConfig(
                object_storage=repo_data.get('object_storage', ''),
                branch=repo_data.get('branch', 'master'),
                plan_requirements=repo_data.get('plan_requirements', []),
                apply_requirements=repo_data.get('apply_requirements', []),
                allowed_workflow=repo_data.get('allowed_workflow', ''),
                github=github_config
            )
        except Exception as e:
            return CliOutput(
                output=f"Error parsing repository config: {e}",
                status=1, github_comment=True,
                config=NeptuneConfig(
                    repository=RepositoryConfig(
                        object_storage=repo_data.get('object_storage', ''),
                        branch=repo_data.get('branch', 'master'),
                        plan_requirements=repo_data.get('plan_requirements', []),
                        apply_requirements=repo_data.get('apply_requirements', []),
                        allowed_workflow=repo_data.get('allowed_workflow', ''),
                        github=github_config),
                    workflows=cls.workflows))

        try:
            # Parse workflows
            logger.debug("%s Parsing workflows", LOG_PREFIX)
            workflow_statements = {}
            for workflow_name, workflow_data in config_data.get('workflows', {}).items():
                phases = {}
                for phase_name, phase_data in workflow_data.items():
                    steps = [WorkflowStep(run=step['run']) for step in phase_data.get('steps', [])]
                    phases[phase_name] = WorkflowPhase(
                        steps=steps,
                        depends_on=phase_data.get('depends_on')
                    )
                workflow_statements[workflow_name] = WorkflowStatement(
                    name=workflow_name,
                    phases=phases
                )
        except Exception as e:
            return CliOutput(output=f"Error parsing workflows: {e}, make sure the workflows are valid", status=1, github_comment=True, config=NeptuneConfig(repository=cls.repository, workflows=cls.workflows))

        cls.workflows = Workflows(workflows=workflow_statements)
        config = NeptuneConfig(repository=cls.repository, workflows=cls.workflows)

        try:
            cls._check_config_options()
        except ValueError as e:
            return CliOutput(output=str(e), status=1, github_comment=True, config=config)

        return config
