"""Test config module."""
import os
from pathlib import Path
import pytest
from unittest.mock import patch, mock_open, Mock
from typer import Exit

from neptune.config import Config
from neptune.utils import CliOutput
from neptune.domain.config import NeptuneConfig, RepositoryConfig, GitHubConfig, Workflows, WorkflowStatement, WorkflowPhase, WorkflowStep

@pytest.fixture
def mock_env_vars():
    """Mock environment variables"""
    env_vars = {
        'NEPTUNE_CONFIG_PATH': '.neptune.yaml',
        'GITHUB_REPOSITORY': 'test/repo',
        'GITHUB_PULL_REQUEST_BRANCH': 'feature/test',
        'GITHUB_PULL_REQUEST_NUMBER': '123',
        'GITHUB_PULL_REQUEST_COMMENT_ID': '456',
        'GITHUB_TOKEN': 'test-token'
    }
    with patch.dict(os.environ, env_vars):
        yield env_vars

@pytest.fixture
def mock_config_file():
    """Mock config file content"""
    return """
repository:
  object_storage: gs://test-bucket
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
    - mergeable
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on:
        - plan
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
"""

@pytest.fixture
def mock_github():
    """Mock GitHub API"""
    with patch('neptune.notifications.github.GitHubNotifications') as mock:
        instance = Mock()
        instance.create_comment.return_value = None
        mock.return_value = instance
        yield mock

@pytest.fixture
def mock_cli_output():
    """Mock CliOutput to prevent GitHub API calls"""
    with patch('neptune.utils.CliOutput._notify') as mock:
        yield mock

class TestConfig:
    """Test Config class"""

    def test_load_env_vars_success(self, mock_env_vars):
        """Test loading environment variables successfully"""
        env_vars = Config._load_env_vars()
        assert env_vars == mock_env_vars

    def test_load_env_vars_missing(self):
        """Test loading environment variables with missing required vars"""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError) as exc_info:
                Config._load_env_vars()
            assert "Environment variables" in str(exc_info.value)
            assert "are required" in str(exc_info.value)

    def test_load_config_success(self, mock_env_vars, mock_config_file, mock_cli_output):
        """Test loading config file successfully"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=mock_config_file)):
            config = Config.load()
            assert isinstance(config, NeptuneConfig)
            assert config.repository.object_storage == "gs://test-bucket"
            assert config.repository.branch == "master"
            assert config.repository.plan_requirements == ["undiverged"]
            assert config.repository.apply_requirements == ["approved", "mergeable"]
            assert config.repository.allowed_workflow == "default"
            assert config.repository.github.repository == "test/repo"
            assert config.repository.github.pull_request_branch == "feature/test"
            assert config.repository.github.pull_request_number == "123"
            assert config.repository.github.pull_request_comment_id == "456"
            assert config.repository.github.token == "test-token"
            assert "default" in config.workflows.workflows
            assert "plan" in config.workflows.workflows["default"].phases
            assert "apply" in config.workflows.workflows["default"].phases

    def test_invalid_github_config(self, mock_env_vars, mock_config_file, mock_cli_output):
        """Test handling invalid GitHub config"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=mock_config_file)), \
             patch('neptune.config.GitHubConfig', side_effect=Exception("Invalid GitHub config")), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1

    def test_invalid_repository_config(self, mock_env_vars, mock_cli_output):
        """Test handling invalid repository config"""
        invalid_config = """
repository:
  object_storage: invalid-url
  branch: master
  plan_requirements:
    - invalid-requirement
  apply_requirements:
    - approved
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=invalid_config)), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1

    def test_invalid_workflow_config(self, mock_env_vars, mock_cli_output):
        """Test handling invalid workflow config"""
        invalid_config = """
repository:
  object_storage: gs://test-bucket
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: terraform plan  # Missing terramate --changed flag
"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=invalid_config)), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1

    def test_missing_required_phases(self, mock_env_vars, mock_cli_output):
        """Test handling missing required phases"""
        invalid_config = """
repository:
  object_storage: gs://test-bucket
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
  allowed_workflow: default

workflows:
  default:
    test:  # Missing plan and apply phases
      steps:
        - run: echo "test"
"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=invalid_config)), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1

    def test_invalid_phase_dependency(self, mock_env_vars, mock_cli_output):
        """Test handling invalid phase dependency"""
        invalid_config = """
repository:
  object_storage: gs://test-bucket
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on:
        - invalid_phase  # Depends on non-existent phase
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=invalid_config)), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1

    def test_branch_validation(self, mock_env_vars, mock_cli_output):
        """Test branch validation"""
        invalid_config = """
repository:
  object_storage: gs://test-bucket
  branch: feature/test  # Same as pull request branch
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
"""
        with patch('pathlib.Path.exists', return_value=True), \
             patch('builtins.open', mock_open(read_data=invalid_config)), \
             pytest.raises(Exit) as exc_info:
            Config.load()
        assert exc_info.value.exit_code == 1 