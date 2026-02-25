"""Test GitHub module."""
import pytest
from unittest.mock import patch, Mock
import requests
from typer import Exit

from neptune.notifications.github import GitHubAPI, GitHubNotifications
from neptune.github import GitHubPullRequest
from neptune.domain.config import NeptuneConfig, RepositoryConfig, GitHubConfig
from neptune.domain.run import StepsOutput, RunOutput
from neptune.domain.lock import TerraformStacks
from neptune.domain.github import PullRequestComment, PRInfo, PRRequirementsStatus

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
def mock_requests():
    """Mock requests library"""
    with patch('requests.post') as mock_post:
        mock_response = Mock()
        mock_response.status_code = 201
        mock_post.return_value = mock_response
        yield mock_post

@pytest.fixture
def mock_cli_output():
    """Mock CliOutput class to prevent exits"""
    def mock_cli_output_factory(*args, **kwargs):
        if kwargs.get('status', 0) != 0:
            raise Exit(code=kwargs.get('status', 1))
        return Mock()

    with patch('neptune.github.CliOutput', side_effect=mock_cli_output_factory) as mock_cli:
        yield mock_cli

class TestGitHubAPI:
    """Test GitHubAPI class"""

    def test_init_with_token(self, mock_config):
        """Test initialization with token"""
        api = GitHubAPI(mock_config)
        assert api.token == "test-token"
        assert api.repo == "test/repo"
        assert api.pr_number == "123"
        assert api.api_url == "https://api.github.com"
        assert api.headers == {
            "Authorization": "token test-token",
            "Accept": "application/vnd.github.v3+json"
        }

    def test_init_without_token(self, mock_config):
        """Test initialization without token"""
        mock_config.repository.github.token = ""
        api = GitHubAPI(mock_config)
        assert api.token == ""
        assert api.headers == {}

class TestGitHubPullRequest:
    """Test GitHubPullRequest class"""

    def test_get_pr_info_success(self, mock_config):
        """Test getting PR info successfully"""
        with patch('requests.get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {
                "number": 123,
                "state": "open",
                "mergeable": True,
                "mergeable_state": "clean"
            }
            mock_get.return_value = mock_response

            pr = GitHubPullRequest(mock_config)
            pr_info = pr._get_pr_info()

            assert isinstance(pr_info, PRInfo)
            assert pr_info.pr_number == "123"
            assert pr_info.repo == "test/repo"
            assert pr_info.api_url == "https://api.github.com"
            assert pr_info.response["mergeable"] is True

    def test_get_pr_info_missing_credentials(self, mock_config, mock_cli_output):
        """Test getting PR info with missing credentials"""
        mock_config.repository.github.token = ""
        pr = GitHubPullRequest(mock_config)
        with pytest.raises(Exit) as exc_info:
            pr._get_pr_info()
        assert exc_info.value.exit_code == 1

    def test_get_pr_info_api_error(self, mock_config, mock_cli_output):
        """Test getting PR info with API error"""
        with patch('requests.get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 404
            mock_response.json.return_value = {"message": "Not Found"}
            mock_get.return_value = mock_response

            pr = GitHubPullRequest(mock_config)
            with pytest.raises(Exit) as exc_info:
                pr._get_pr_info()
            assert exc_info.value.exit_code == 1

    def test_is_pr_open_true(self, mock_config):
        """Test checking if PR is open when it is"""
        with patch('requests.get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {"state": "open"}
            mock_get.return_value = mock_response

            pr = GitHubPullRequest(mock_config)
            assert pr.is_pr_open("123") is True

    def test_is_pr_open_false(self, mock_config):
        """Test checking if PR is open when it's closed"""
        with patch('requests.get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 200
            mock_response.json.return_value = {"state": "closed"}
            mock_get.return_value = mock_response

            pr = GitHubPullRequest(mock_config)
            assert pr.is_pr_open("123") is False

    def test_is_pr_open_api_error(self, mock_config, mock_cli_output):
        """Test checking if PR is open with API error"""
        with patch('requests.get') as mock_get:
            mock_response = Mock()
            mock_response.status_code = 404
            mock_response.json.return_value = {"message": "Not Found"}
            mock_get.return_value = mock_response

            pr = GitHubPullRequest(mock_config)
            with pytest.raises(Exit) as exc_info:
                pr.is_pr_open("123")
            assert exc_info.value.exit_code == 1

    def test_check_requirements_no_requirements(self, mock_config):
        """Test checking requirements when none are specified"""
        pr = GitHubPullRequest(mock_config)
        result = pr.check_requirements([])

        assert isinstance(result, PRRequirementsStatus)
        assert result.is_compliant is True
        assert result.failed_requirements == []
        assert result.error_message == ""

    def test_check_requirements_missing_pr_info(self, mock_config, mock_cli_output):
        """Test checking requirements when PR info is missing"""
        mock_config.repository.github.token = ""
        pr = GitHubPullRequest(mock_config)
        with pytest.raises(Exit) as exc_info:
            pr.check_requirements(["approved", "mergeable"])
        assert exc_info.value.exit_code == 1

    def test_check_requirements_all_pass(self, mock_config):
        """Test checking requirements when all pass"""
        with patch('requests.get') as mock_get:
            # Mock PR info response
            pr_response = Mock()
            pr_response.status_code = 200
            pr_response.json.return_value = {
                "number": 123,
                "mergeable": True,
                "mergeable_state": "clean"
            }

            # Mock reviews response
            reviews_response = Mock()
            reviews_response.status_code = 200
            reviews_response.json.return_value = [{"state": "APPROVED"}]

            mock_get.side_effect = [pr_response, reviews_response]

            pr = GitHubPullRequest(mock_config)
            result = pr.check_requirements(["approved", "mergeable", "undiverged"])

            assert isinstance(result, PRRequirementsStatus)
            assert result.is_compliant is True
            assert result.failed_requirements == []
            assert result.error_message == ""

    def test_check_requirements_some_fail(self, mock_config):
        """Test checking requirements when some fail"""
        with patch('requests.get') as mock_get:
            # Mock PR info response
            pr_response = Mock()
            pr_response.status_code = 200
            pr_response.json.return_value = {
                "number": 123,
                "mergeable": False,
                "mergeable_state": "behind"
            }

            # Mock reviews response
            reviews_response = Mock()
            reviews_response.status_code = 200
            reviews_response.json.return_value = [{"state": "COMMENTED"}]

            mock_get.side_effect = [pr_response, reviews_response]

            pr = GitHubPullRequest(mock_config)
            result = pr.check_requirements(["approved", "mergeable", "undiverged"])

            assert isinstance(result, PRRequirementsStatus)
            assert result.is_compliant is False
            assert set(result.failed_requirements) == {"approved", "mergeable", "undiverged"}
            assert "PR does not meet the following requirements" in result.error_message

    def test_check_requirements_reviews_api_error(self, mock_config):
        """Test checking requirements when reviews API fails"""
        with patch('requests.get') as mock_get:
            # Mock PR info response
            pr_response = Mock()
            pr_response.status_code = 200
            pr_response.json.return_value = {
                "number": 123,
                "mergeable": True,
                "mergeable_state": "clean"
            }

            # Mock reviews response
            reviews_response = Mock()
            reviews_response.status_code = 404
            reviews_response.json.return_value = {"message": "Not Found"}

            mock_get.side_effect = [pr_response, reviews_response]

            pr = GitHubPullRequest(mock_config)
            result = pr.check_requirements(["approved"])

            assert isinstance(result, PRRequirementsStatus)
            assert result.is_compliant is False
            assert result.failed_requirements == ["approved"]
            assert "Could not fetch PR reviews" in result.error_message

class TestGitHubNotifications:
    """Test GitHubNotifications class"""

    def test_strip_ansi_codes(self):
        """Test stripping ANSI codes from text"""
        text = "\x1B[31mError\x1B[0m"
        assert GitHubNotifications._strip_ansi_codes(text) == "Error"

    def test_create_comment_missing_credentials(self, mock_config):
        """Test comment creation with missing credentials"""
        mock_config.repository.github.token = ""
        notifications = GitHubNotifications(mock_config)
        comment = PullRequestComment(
            simple_output="test",
            overall_status=0,
            steps_output=None
        )
        # Should return silently when credentials are missing
        notifications.create_comment(comment)

    def test_create_comment_success(self, mock_config, mock_requests):
        """Test successful comment creation"""
        notifications = GitHubNotifications(mock_config)
        comment = PullRequestComment(
            simple_output="test",
            overall_status=0,
            steps_output=StepsOutput(
                phase="plan",
                overall_status=0,
                outputs=[
                    RunOutput(
                        command="terraform plan",
                        output="No changes",
                        error="",
                        status=0
                    )
                ]
            )
        )
        notifications.create_comment(comment)
        mock_requests.assert_called_once()
        assert "Neptune Plan Results" in mock_requests.call_args[1]["json"]["body"]

    def test_create_comment_api_error(self, mock_config):
        """Test comment creation with API error"""
        with patch('requests.post') as mock_post:
            mock_response = Mock()
            mock_response.status_code = 401
            mock_response.text = "Unauthorized"
            mock_response.raise_for_status.side_effect = requests.exceptions.HTTPError("401 Client Error: Unauthorized")
            mock_post.return_value = mock_response

            notifications = GitHubNotifications(mock_config)
            comment = PullRequestComment(
                simple_output="test",
                overall_status=0,
                steps_output=StepsOutput(
                    phase="plan",
                    overall_status=0,
                    outputs=[
                        RunOutput(
                            command="terraform plan",
                            output="No changes",
                            error="",
                            status=0
                        )
                    ]
                )
            )
            with pytest.raises(requests.exceptions.HTTPError):
                notifications.create_comment(comment)

    def test_format_plan_output(self, mock_config):
        """Test formatting plan output"""
        notifications = GitHubNotifications(mock_config)
        comment = PullRequestComment(
            simple_output="test",
            overall_status=0,
            steps_output=StepsOutput(
                phase="plan",
                overall_status=0,
                outputs=[
                    RunOutput(
                        command="terraform plan",
                        output="No changes",
                        error="",
                        status=0
                    )
                ]
            ),
            stacks=TerraformStacks(stacks=["stack1", "stack2"])
        )
        output = notifications._format_plan_output(comment)
        assert "Neptune Plan Results" in output
        assert "stack1, stack2" in output
        assert "✅" in output
        assert "/neptune apply" in output

    def test_format_apply_output(self, mock_config):
        """Test formatting apply output"""
        notifications = GitHubNotifications(mock_config)
        comment = PullRequestComment(
            simple_output="test",
            overall_status=0,
            steps_output=StepsOutput(
                phase="apply",
                overall_status=0,
                outputs=[
                    RunOutput(
                        command="terraform apply",
                        output="Applied successfully",
                        error="",
                        status=0
                    )
                ]
            ),
            stacks=TerraformStacks(stacks=["stack1", "stack2"])
        )
        output = notifications._format_apply_output(comment)
        assert "Neptune Apply Results" in output
        assert "stack1, stack2" in output
        assert "✅" in output

    def test_format_custom_output(self, mock_config):
        """Test formatting custom output"""
        notifications = GitHubNotifications(mock_config)
        comment = PullRequestComment(
            simple_output="Custom error",
            overall_status=1,
            steps_output=StepsOutput(
                phase="custom",
                overall_status=1,
                outputs=[
                    RunOutput(
                        command="custom command",
                        output="output",
                        error="error",
                        status=1
                    )
                ]
            )
        )
        output = notifications._format_custom_output(comment)
        assert "Neptune Execution Results" in output
        assert "Custom error" in output
        assert "❌" in output 