"""Test utils module."""
import pytest
from unittest.mock import patch, Mock
from typer import Exit

from neptune.utils import CliOutput
from neptune.domain.github import PullRequestComment
from neptune.domain.config import NeptuneConfig, RepositoryConfig, GitHubConfig

@pytest.fixture
def mock_config():
    """Create a mock config"""
    return NeptuneConfig(
        repository=RepositoryConfig(
            object_storage="gs://test-bucket",
            branch="main",
            plan_requirements=["undiverged"],
            apply_requirements=["approved", "mergeable"],
            allowed_workflow="default",
            github=GitHubConfig(
                repository="test/repo",
                pull_request_branch="feature/test",
                pull_request_number="123",
                pull_request_comment_id="456",
                token="test-token"
            )
        ),
        workflows=None
    )

@pytest.fixture
def mock_github():
    """Mock GitHub notifications"""
    with patch('neptune.utils.GitHubNotifications') as mock:
        yield mock

@pytest.fixture
def mock_print():
    """Mock print function"""
    with patch('neptune.utils.print') as mock:
        yield mock

class TestCliOutput:
    """Test CliOutput class"""

    def test_success_output(self, mock_print):
        """Test successful output"""
        with pytest.raises(Exit) as exc_info:
            CliOutput(output="Success", status=0)
        
        # In Typer, Exit(0) is used for success
        assert exc_info.value.exit_code == 0
        mock_print.assert_called_once_with("[bold green]Success:[/bold green] Success")

    def test_error_output(self, mock_print):
        """Test error output"""
        with pytest.raises(Exit) as exc_info:
            CliOutput(output="Error", status=1)
        
        # In Typer, Exit(1) is used for errors
        assert exc_info.value.exit_code == 1
        mock_print.assert_called_once_with("[bold red]Error:[/bold red] Error")

    def test_github_comment_without_output(self, mock_print):
        """Test GitHub comment without output"""
        with pytest.raises(Exit) as exc_info:
            CliOutput(output="", status=0, github_comment=True)
        
        # Should exit with error code 1
        assert exc_info.value.exit_code == 1
        mock_print.assert_called_once_with(
            "[bold red]Error:[/bold red] Comment is required when github_comment is True, this is usually a bug in the code"
        )

    def test_github_comment_with_output(self, mock_config, mock_github, mock_print):
        """Test GitHub comment with output"""
        with pytest.raises(Exit) as exc_info:
            CliOutput(
                output="Test message",
                status=0,
                github_comment=True,
                config=mock_config
            )
        
        # Should exit with success code 0
        assert exc_info.value.exit_code == 0
        mock_github.return_value.create_comment.assert_called_once()
        mock_print.assert_called_with("[bold green]Success:[/bold green] Test message")

    def test_custom_comment(self, mock_config, mock_github, mock_print):
        """Test custom comment"""
        custom_comment = PullRequestComment(
            simple_output="Custom message",
            steps_output=None,
            overall_status=0
        )

        with pytest.raises(Exit) as exc_info:
            CliOutput(
                output="Success",
                status=0,
                github_comment=True,
                config=mock_config,
                comment=custom_comment
            )
        
        # Should exit with success code 0
        assert exc_info.value.exit_code == 0
        mock_github.return_value.create_comment.assert_called_once_with(comment=custom_comment)
        mock_print.assert_called_with("[bold green]Success:[/bold green] Success") 