"""Unit tests for moai_adk.core.issue_creator module."""

from unittest import mock

import pytest

from moai_adk.core.issue_creator import (
    GitHubIssueCreator,
    IssueConfig,
    IssueCreatorFactory,
    IssuePriority,
    IssueType,
)


class TestIssueType:
    """Test IssueType enumeration."""

    def test_issue_types(self):
        """Test all issue types exist."""
        assert IssueType.BUG.value == "bug"
        assert IssueType.FEATURE.value == "feature"
        assert IssueType.IMPROVEMENT.value == "improvement"
        assert IssueType.QUESTION.value == "question"


class TestIssuePriority:
    """Test IssuePriority enumeration."""

    def test_priority_levels(self):
        """Test all priority levels exist."""
        assert IssuePriority.CRITICAL.value == "critical"
        assert IssuePriority.HIGH.value == "high"
        assert IssuePriority.MEDIUM.value == "medium"
        assert IssuePriority.LOW.value == "low"


class TestIssueConfig:
    """Test IssueConfig dataclass."""

    def test_create_minimal_config(self):
        """Test creating minimal config."""
        config = IssueConfig(issue_type=IssueType.BUG, title="Test bug", description="Bug description")
        assert config.issue_type == IssueType.BUG
        assert config.priority == IssuePriority.MEDIUM

    def test_create_full_config(self):
        """Test creating full config."""
        config = IssueConfig(
            issue_type=IssueType.FEATURE,
            title="New feature",
            description="Feature description",
            priority=IssuePriority.HIGH,
            category="Enhancement",
            assignees=["user1"],
            custom_labels=["label1"],
        )
        assert config.priority == IssuePriority.HIGH
        assert config.category == "Enhancement"


class TestGitHubIssueCreator:
    """Test GitHubIssueCreator class."""

    def test_init_success(self):
        """Test initialization with successful gh check."""
        with mock.patch("moai_adk.core.issue_creator.subprocess.run") as mock_run:
            mock_run.return_value.returncode = 0
            creator = GitHubIssueCreator()
            assert creator is not None

    def test_init_gh_not_installed(self):
        """Test initialization when gh CLI not installed."""
        with mock.patch("moai_adk.core.issue_creator.subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError()
            with pytest.raises(RuntimeError):
                GitHubIssueCreator()

    def test_init_gh_not_authenticated(self):
        """Test initialization when gh CLI not authenticated."""
        with mock.patch("moai_adk.core.issue_creator.subprocess.run") as mock_run:
            mock_run.return_value.returncode = 1
            with pytest.raises(RuntimeError):
                GitHubIssueCreator()

    def test_label_map(self):
        """Test label mapping."""
        assert "bug" in GitHubIssueCreator.LABEL_MAP[IssueType.BUG]
        assert "feature-request" in GitHubIssueCreator.LABEL_MAP[IssueType.FEATURE]

    def test_priority_emoji(self):
        """Test priority emoji mapping."""
        assert GitHubIssueCreator.PRIORITY_EMOJI[IssuePriority.CRITICAL] == "🔴"
        assert GitHubIssueCreator.PRIORITY_EMOJI[IssuePriority.HIGH] == "🟠"

    def test_type_emoji(self):
        """Test type emoji mapping."""
        assert GitHubIssueCreator.TYPE_EMOJI[IssueType.BUG] == "🐛"
        assert GitHubIssueCreator.TYPE_EMOJI[IssueType.FEATURE] == "✨"

    def test_create_issue_success(self):
        """Test successful issue creation."""
        with mock.patch("moai_adk.core.issue_creator.subprocess.run") as mock_run:
            mock_run.return_value.returncode = 0
            mock_run.return_value.stdout = "https://github.com/owner/repo/issues/123"

            # Mock both the auth check and create calls
            def run_side_effect(*args, **kwargs):
                if "auth" in args[0]:
                    result = mock.Mock()
                    result.returncode = 0
                    return result
                else:
                    result = mock.Mock()
                    result.returncode = 0
                    result.stdout = "https://github.com/owner/repo/issues/123"
                    return result

            mock_run.side_effect = run_side_effect

            creator = GitHubIssueCreator()
            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Bug description",
            )

            result = creator.create_issue(config)
            assert result["success"] is True
            assert result["issue_number"] == 123

    def test_extract_issue_number(self):
        """Test extracting issue number from URL."""
        url = "https://github.com/owner/repo/issues/456"
        number = GitHubIssueCreator._extract_issue_number(url)
        assert number == 456

    def test_extract_issue_number_invalid(self):
        """Test extracting issue number from invalid URL."""
        url = "https://github.com/owner/repo"
        with pytest.raises(ValueError):
            GitHubIssueCreator._extract_issue_number(url)

    def test_build_body(self):
        """Test building issue body."""
        with mock.patch("moai_adk.core.issue_creator.subprocess.run") as mock_run:
            mock_run.return_value.returncode = 0
            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test",
                description="Description",
                priority=IssuePriority.HIGH,
                category="Testing",
            )

            body = creator._build_body(config)
            assert "Description" in body
            assert "Type" in body
            assert "Priority" in body


class TestIssueCreatorFactory:
    """Test IssueCreatorFactory class."""

    def test_create_bug_issue(self):
        """Test creating bug issue."""
        config = IssueCreatorFactory.create_bug_issue("Bug title", "Bug details")
        assert config.issue_type == IssueType.BUG
        assert config.title == "Bug title"

    def test_create_feature_issue(self):
        """Test creating feature issue."""
        config = IssueCreatorFactory.create_feature_issue("Feature", "Details")
        assert config.issue_type == IssueType.FEATURE

    def test_create_improvement_issue(self):
        """Test creating improvement issue."""
        config = IssueCreatorFactory.create_improvement_issue("Improve", "Details")
        assert config.issue_type == IssueType.IMPROVEMENT

    def test_create_question_issue(self):
        """Test creating question issue."""
        config = IssueCreatorFactory.create_question_issue("Question", "Details")
        assert config.issue_type == IssueType.QUESTION
