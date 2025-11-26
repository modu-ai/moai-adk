"""
GitHub Issue Creator for MoAI-ADK quick issue reporting.

Enables users to quickly create GitHub Issues with standardized templates
using `/moai:9-feedback` interactive dialog.

"""

import subprocess
from dataclasses import dataclass
from enum import Enum
from typing import Any, Dict, List, Optional


class IssueType(Enum):
    """Supported GitHub issue types."""

    BUG = "bug"
    FEATURE = "feature"
    IMPROVEMENT = "improvement"
    QUESTION = "question"


class IssuePriority(Enum):
    """Issue priority levels."""

    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


@dataclass
class IssueConfig:
    """Configuration for issue creation."""

    issue_type: IssueType
    title: str
    description: str
    priority: IssuePriority = IssuePriority.MEDIUM
    category: Optional[str] = None
    assignees: Optional[List[str]] = None
    custom_labels: Optional[List[str]] = None


class GitHubIssueCreator:
    """
    Creates GitHub Issues using the `gh` CLI.

    Supports:
    - Multiple issue types (bug, feature, improvement, question)
    - Priority levels and categories
    - Standard templates for each type
    - Label automation
    - Priority emoji indicators
    """

    # Label mapping for issue types
    LABEL_MAP = {
        IssueType.BUG: ["bug", "reported"],
        IssueType.FEATURE: ["feature-request", "enhancement"],
        IssueType.IMPROVEMENT: ["improvement", "enhancement"],
        IssueType.QUESTION: [
            "question",
            "help wanted",
        ],  # Fixed: "help-wanted" → "help wanted" (GitHub standard)
    }

    # Priority emoji
    PRIORITY_EMOJI = {
        IssuePriority.CRITICAL: "🔴",
        IssuePriority.HIGH: "🟠",
        IssuePriority.MEDIUM: "🟡",
        IssuePriority.LOW: "🟢",
    }

    # Issue type emoji
    TYPE_EMOJI = {
        IssueType.BUG: "🐛",
        IssueType.FEATURE: "✨",
        IssueType.IMPROVEMENT: "⚡",
        IssueType.QUESTION: "❓",
    }

    def __init__(self, github_token: Optional[str] = None):
        """
        Initialize the GitHub Issue Creator.

        Args:
            github_token: GitHub API token. If not provided, uses GITHUB_TOKEN env var.
        """
        self.github_token = github_token
        self._check_gh_cli()

    def _check_gh_cli(self) -> None:
        """
        Check if `gh` CLI is installed and accessible.

        Raises:
            RuntimeError: If `gh` CLI is not found or not authenticated.
        """
        try:
            result = subprocess.run(["gh", "auth", "status"], capture_output=True, text=True, timeout=5)
            if result.returncode != 0:
                raise RuntimeError("GitHub CLI (gh) is not authenticated. Run `gh auth login` to authenticate.")
        except FileNotFoundError:
            raise RuntimeError("GitHub CLI (gh) is not installed. Please install it: https://cli.github.com")

    def create_issue(self, config: IssueConfig) -> Dict[str, Any]:
        """
        Create a GitHub issue with the given configuration.

        Args:
            config: Issue configuration

        Returns:
            Dictionary containing issue creation result:
            {
                "success": bool,
                "issue_number": int,
                "issue_url": str,
                "message": str
            }

        Raises:
            RuntimeError: If issue creation fails
        """
        # Build title with emoji and priority
        emoji = self.TYPE_EMOJI.get(config.issue_type, "📋")
        priority_emoji = self.PRIORITY_EMOJI.get(config.priority, "")
        full_title = f"{emoji} [{config.issue_type.value.upper()}] {config.title}"
        if priority_emoji:
            full_title = f"{priority_emoji} {full_title}"

        # Build body with template
        body = self._build_body(config)

        # Collect labels
        labels = self.LABEL_MAP.get(config.issue_type, []).copy()
        if config.priority:
            labels.append(config.priority.value)  # Fixed: removed "priority-" prefix (use direct label names)
        if config.category:
            labels.append(f"category-{config.category.lower().replace(' ', '-')}")
        if config.custom_labels:
            labels.extend(config.custom_labels)

        # Build gh command
        gh_command = [
            "gh",
            "issue",
            "create",
            "--title",
            full_title,
            "--body",
            body,
        ]

        # Add labels
        if labels:
            gh_command.extend(["--label", ",".join(set(labels))])

        # Add assignees if provided
        if config.assignees:
            gh_command.extend(["--assignee", ",".join(config.assignees)])

        try:
            result = subprocess.run(gh_command, capture_output=True, text=True, timeout=30)

            if result.returncode != 0:
                error_msg = result.stderr or result.stdout
                raise RuntimeError(f"Failed to create GitHub issue: {error_msg}")

            # Parse issue URL from output
            issue_url = result.stdout.strip()
            issue_number = self._extract_issue_number(issue_url)

            return {
                "success": True,
                "issue_number": issue_number,
                "issue_url": issue_url,
                "message": f"✅ GitHub Issue #{issue_number} created successfully",
                "title": full_title,
                "labels": labels,
            }

        except subprocess.TimeoutExpired:
            raise RuntimeError("GitHub issue creation timed out")
        except Exception as e:
            raise RuntimeError(f"Error creating GitHub issue: {e}")

    def _build_body(self, config: IssueConfig) -> str:
        """
        Build the issue body based on issue type.

        Args:
            config: Issue configuration

        Returns:
            Formatted issue body
        """
        body = config.description

        # Add metadata footer
        footer = "\n\n---\n\n"
        footer += f"**Type**: {config.issue_type.value}  \n"
        footer += f"**Priority**: {config.priority.value}  \n"
        if config.category:
            footer += f"**Category**: {config.category}  \n"
        footer += "**Created via**: `/moai:9-feedback`"

        return body + footer

    @staticmethod
    def _extract_issue_number(url: str) -> int:
        """
        Extract issue number from GitHub URL.

        Args:
            url: GitHub issue URL

        Returns:
            Issue number

        Raises:
            ValueError: If unable to extract issue number
        """
        try:
            # URL format: https://github.com/owner/repo/issues/123
            return int(url.strip().split("/")[-1])
        except (ValueError, IndexError):
            raise ValueError(f"Unable to extract issue number from URL: {url}")

    def format_result(self, result: Dict[str, Any]) -> str:
        """
        Format the issue creation result for display.

        Args:
            result: Issue creation result

        Returns:
            Formatted result string
        """
        if result["success"]:
            output = f"{result['message']}\n"
            output += f"📋 Title: {result['title']}\n"
            output += f"🔗 URL: {result['issue_url']}\n"
            if result.get("labels"):
                output += f"🏷️  Labels: {', '.join(result['labels'])}\n"
            return output
        else:
            return f"❌ Failed to create issue: {result.get('message', 'Unknown error')}"


class IssueCreatorFactory:
    """
    Factory for creating issue creators with predefined configurations.
    """

    @staticmethod
    def create_bug_issue(title: str, description: str, priority: IssuePriority = IssuePriority.HIGH) -> IssueConfig:
        """Create a bug report issue configuration."""
        return IssueConfig(
            issue_type=IssueType.BUG,
            title=title,
            description=description,
            priority=priority,
            category="Bug Report",
        )

    @staticmethod
    def create_feature_issue(
        title: str, description: str, priority: IssuePriority = IssuePriority.MEDIUM
    ) -> IssueConfig:
        """Create a feature request issue configuration."""
        return IssueConfig(
            issue_type=IssueType.FEATURE,
            title=title,
            description=description,
            priority=priority,
            category="Feature Request",
        )

    @staticmethod
    def create_improvement_issue(
        title: str, description: str, priority: IssuePriority = IssuePriority.MEDIUM
    ) -> IssueConfig:
        """Create an improvement issue configuration."""
        return IssueConfig(
            issue_type=IssueType.IMPROVEMENT,
            title=title,
            description=description,
            priority=priority,
            category="Improvement",
        )

    @staticmethod
    def create_question_issue(title: str, description: str, priority: IssuePriority = IssuePriority.LOW) -> IssueConfig:
        """Create a question/discussion issue configuration."""
        return IssueConfig(
            issue_type=IssueType.QUESTION,
            title=title,
            description=description,
            priority=priority,
            category="Question",
        )
