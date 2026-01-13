"""Additional coverage tests for issue creator.

Tests for lines not covered by existing tests.
"""

from unittest.mock import MagicMock, patch

from moai_adk.core.issue_creator import GitHubIssueCreator, IssueConfig, IssueType


class TestGitHubIssueCreatorCategoryLabel:
    """Test category label handling."""

    def test_create_issue_with_category_adds_label(self):
        """Should add category label when category is provided."""
        # Mock subprocess to succeed for both init check and issue creation
        with patch("subprocess.run") as mock_run:
            # Init check succeeds
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            # Issue creation succeeds
            mock_issue_result = MagicMock()
            mock_issue_result.returncode = 0
            mock_issue_result.stdout = "https://github.com/user/repo/issues/1"

            mock_run.side_effect = [mock_init_result, mock_issue_result]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
                category="Feature Request",
            )

            creator.create_issue(config)

            # Verify category label was added
            call_args = mock_run.call_args_list
            gh_command = call_args[1][0][0]  # Second call (issue creation)
            assert any("category-feature-request" in str(arg) for arg in gh_command)


class TestGitHubIssueCreatorCustomLabels:
    """Test custom labels handling."""

    def test_create_issue_with_custom_labels(self):
        """Should extend labels with custom_labels."""
        # Mock subprocess to succeed
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            mock_issue_result = MagicMock()
            mock_issue_result.returncode = 0
            mock_issue_result.stdout = "https://github.com/user/repo/issues/1"

            mock_run.side_effect = [mock_init_result, mock_issue_result]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
                custom_labels=["needs-review", "high-priority"],
            )

            creator.create_issue(config)

            # Verify custom labels were added
            call_args = mock_run.call_args_list
            gh_command = call_args[1][0][0]
            assert "needs-review" in str(gh_command)
            assert "high-priority" in str(gh_command)


class TestGitHubIssueCreatorAssignees:
    """Test assignee handling."""

    def test_create_issue_with_assignees(self):
        """Should add assignees to gh command."""
        # Mock subprocess to succeed
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            mock_issue_result = MagicMock()
            mock_issue_result.returncode = 0
            mock_issue_result.stdout = "https://github.com/user/repo/issues/1"

            mock_run.side_effect = [mock_init_result, mock_issue_result]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
                assignees=["user1", "user2"],
            )

            creator.create_issue(config)

            # Verify assignees were added
            call_args = mock_run.call_args_list
            gh_command = call_args[1][0][0]
            assert "--assignee" in gh_command
            assignee_idx = gh_command.index("--assignee")
            assert gh_command[assignee_idx + 1] == "user1,user2"


class TestGitHubIssueCreatorErrorHandling:
    """Test error handling in issue creation."""

    def test_create_issue_gh_command_failure(self):
        """Should raise RuntimeError when gh command fails."""
        # Mock init to succeed, issue creation to fail
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            mock_issue_result = MagicMock()
            mock_issue_result.returncode = 1
            mock_issue_result.stderr = "Authentication failed"

            mock_run.side_effect = [mock_init_result, mock_issue_result]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
            )

            try:
                creator.create_issue(config)
                assert False, "Should have raised RuntimeError"
            except RuntimeError as e:
                assert "Failed to create GitHub issue" in str(e)
                assert "Authentication failed" in str(e)

    def test_create_issue_timeout(self):
        """Should raise RuntimeError when gh command times out."""
        import subprocess

        # Mock init to succeed, issue creation to timeout
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            mock_run.side_effect = [mock_init_result, subprocess.TimeoutExpired("gh", timeout=30)]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
            )

            try:
                creator.create_issue(config)
                assert False, "Should have raised RuntimeError"
            except RuntimeError as e:
                assert "timed out" in str(e)

    def test_create_issue_generic_exception(self):
        """Should raise RuntimeError for generic exceptions."""
        # Mock init to succeed, issue creation to raise exception
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""

            mock_run.side_effect = [mock_init_result, OSError("Network error")]

            creator = GitHubIssueCreator()

            config = IssueConfig(
                issue_type=IssueType.BUG,
                title="Test bug",
                description="Test description",
            )

            try:
                creator.create_issue(config)
                assert False, "Should have raised RuntimeError"
            except RuntimeError as e:
                assert "Error creating GitHub issue" in str(e)


class TestFormatResult:
    """Test format_result method."""

    def test_format_result_with_labels(self):
        """Should format result with labels."""
        # Mock subprocess for init
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""
            mock_run.return_value = mock_init_result

            creator = GitHubIssueCreator()

            result = {
                "success": True,
                "issue_number": 1,
                "issue_url": "https://github.com/user/repo/issues/1",
                "message": "✅ Issue created",
                "title": "Test bug",
                "labels": ["bug", "high-priority"],
            }

            formatted = creator.format_result(result)

            assert "✅ Issue created" in formatted
            assert "Test bug" in formatted
            assert "https://github.com/user/repo/issues/1" in formatted
            assert "Labels: bug, high-priority" in formatted

    def test_format_result_failure(self):
        """Should format failure result."""
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""
            mock_run.return_value = mock_init_result

            creator = GitHubIssueCreator()

            result = {"success": False, "message": "Authentication failed"}

            formatted = creator.format_result(result)

            assert "❌ Failed to create issue" in formatted
            assert "Authentication failed" in formatted

    def test_format_result_without_labels(self):
        """Should format result without labels."""
        with patch("subprocess.run") as mock_run:
            mock_init_result = MagicMock()
            mock_init_result.returncode = 0
            mock_init_result.stdout = ""
            mock_run.return_value = mock_init_result

            creator = GitHubIssueCreator()

            result = {
                "success": True,
                "issue_number": 1,
                "issue_url": "https://github.com/user/repo/issues/1",
                "message": "✅ Issue created",
                "title": "Test bug",
            }

            formatted = creator.format_result(result)

            assert "✅ Issue created" in formatted
            assert "Test bug" in formatted
            assert "https://github.com/user/repo/issues/1" in formatted
            assert "Labels:" not in formatted
