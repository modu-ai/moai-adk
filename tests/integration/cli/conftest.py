"""Pytest configuration and fixtures for CLI integration tests.

Provides fixtures for end-to-end CLI command testing with real file system
operations and temporary project directories.
"""

import json
import subprocess
import sys
from pathlib import Path
from typing import Generator

import pytest
import yaml
from click.testing import CliRunner

# ===== PATH SETUP =====
# Add project root to sys.path for imports
PROJECT_ROOT = Path(__file__).parent.parent.parent.parent
SRC_DIR = PROJECT_ROOT / "src"

if str(SRC_DIR) not in sys.path:
    sys.path.insert(0, str(SRC_DIR))


# ===== MARKERS =====
def pytest_configure(config):
    """Configure pytest markers."""
    config.addinivalue_line("markers", "integration: mark test as integration test (slow, real I/O)")


# ===== FIXTURES =====


@pytest.fixture
def cli_runner() -> CliRunner:
    """Provide a Click CLI runner for testing commands.

    Returns:
        CliRunner instance for invoking Click commands
    """
    return CliRunner()


@pytest.fixture
def temp_project_dir(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary project directory with .moai structure.

    This fixture creates a minimal MoAI-ADK project structure for testing
    CLI commands that require a valid project context.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Yields:
        Path to temporary project directory with .moai structure
    """
    project_dir = tmp_path / "test_project"
    project_dir.mkdir(parents=True, exist_ok=True)

    # Create .moai directory structure
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir(parents=True, exist_ok=True)

    # Create config directory
    config_dir = moai_dir / "config"
    config_dir.mkdir(parents=True, exist_ok=True)

    # Create sections directory
    sections_dir = config_dir / "sections"
    sections_dir.mkdir(parents=True, exist_ok=True)

    # Create minimal config.yaml
    config_data = {
        "moai": {
            "version": "0.37.0",
        },
        "project": {
            "name": "test-project",
            "optimized": False,
        },
    }

    config_yaml = config_dir / "config.yaml"
    with open(config_yaml, "w", encoding="utf-8") as f:
        yaml.safe_dump(config_data, f, default_flow_style=False, allow_unicode=True)

    # Create minimal language.yaml section
    language_data = {
        "language": {
            "conversation_language": "en",
            "conversation_language_name": "English",
            "git_commit_messages": "en",
            "code_comments": "en",
            "documentation": "en",
        }
    }

    language_yaml = sections_dir / "language.yaml"
    with open(language_yaml, "w", encoding="utf-8") as f:
        yaml.safe_dump(language_data, f, default_flow_style=False, allow_unicode=True)

    # Create minimal project.yaml section
    project_data = {
        "project": {
            "name": "test-project",
        }
    }

    project_yaml = sections_dir / "project.yaml"
    with open(project_yaml, "w", encoding="utf-8") as f:
        yaml.safe_dump(project_data, f, default_flow_style=False, allow_unicode=True)

    # Create logs directory for session data
    logs_dir = moai_dir / "logs"
    logs_dir.mkdir(parents=True, exist_ok=True)

    yield project_dir

    # Cleanup is automatic with tmp_path


@pytest.fixture
def temp_git_repo(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary Git repository for testing.

    Initializes a minimal Git repository with test configuration.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Yields:
        Path to temporary Git repository
    """
    repo_dir = tmp_path / "test_repo"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository
    subprocess.run(
        ["git", "init"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Configure Git user (required for commits)
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.email", "test@example.com"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    yield repo_dir


@pytest.fixture
def temp_session_file(temp_project_dir: Path) -> Path:
    """Create a temporary session file for analyze command testing.

    Creates a mock session log file with sample data for testing
    the analyze command.

    Args:
        temp_project_dir: Temporary project directory fixture

    Returns:
        Path to created session file
    """
    logs_dir = temp_project_dir / ".moai" / "logs"

    # Create a mock session file
    session_file = logs_dir / "session_20250113_120000.jsonl"
    session_data = [
        {
            "timestamp": "2025-01-13T12:00:00",
            "role": "user",
            "type": "user_message",
            "content": "Hello, help me with my project",
        },
        {
            "timestamp": "2025-01-13T12:00:01",
            "role": "assistant",
            "type": "tool_use",
            "tool": "Read",
            "content": {"file_path": "README.md"},
        },
        {
            "timestamp": "2025-01-13T12:00:02",
            "role": "assistant",
            "type": "tool_use",
            "tool": "Bash",
            "content": {"command": "ls -la"},
        },
    ]

    # Write session data as JSONL
    with open(session_file, "w", encoding="utf-8") as f:
        for entry in session_data:
            f.write(json.dumps(entry) + "\n")

    return session_file


@pytest.fixture
def temp_moai_commands_dir(temp_project_dir: Path) -> Path:
    """Create temporary slash command files for testing.

    Creates mock slash command files in the .claude/commands directory
    for testing the doctor command's --check-commands flag.

    Args:
        temp_project_dir: Temporary project directory fixture

    Returns:
        Path to commands directory
    """
    claude_dir = temp_project_dir / ".claude"
    claude_dir.mkdir(parents=True, exist_ok=True)

    commands_dir = claude_dir / "commands"
    commands_dir.mkdir(parents=True, exist_ok=True)

    # Create a valid command file
    valid_command = commands_dir / "test-command.md"
    valid_command.write_text(
        """
---
name: test-command
description: A test command
---

# Test Command

This is a test command for doctor verification.
"""
    )

    # Create an invalid command file (missing frontmatter)
    invalid_command = commands_dir / "invalid-command.md"
    invalid_command.write_text(
        """
# Invalid Command

This command is missing frontmatter.
"""
    )

    return commands_dir


# ===== HELPER FUNCTIONS =====


def assert_project_structure(project_dir: Path, expected_files: list[str] | None = None) -> None:
    """Assert that project structure contains expected files.

    Args:
        project_dir: Path to project directory
        expected_files: List of expected file paths (relative to project root)

    Raises:
        AssertionError: If expected files are missing
    """
    if expected_files is None:
        expected_files = [
            ".moai/config/config.yaml",
            ".moai/config/sections/language.yaml",
            ".moai/config/sections/project.yaml",
        ]

    for file_path in expected_files:
        full_path = project_dir / file_path
        assert full_path.exists(), f"Expected file not found: {file_path}"


def count_command_files(commands_dir: Path) -> tuple[int, int]:
    """Count valid and invalid command files.

    Args:
        commands_dir: Path to commands directory

    Returns:
        Tuple of (total_files, valid_files)
    """
    if not commands_dir.exists():
        return 0, 0

    total_files = 0
    valid_files = 0

    for file_path in commands_dir.rglob("*.md"):
        total_files += 1
        content = file_path.read_text(encoding="utf-8")
        if content.startswith("---"):
            valid_files += 1

    return total_files, valid_files


@pytest.fixture
def git_repo_with_commit(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a Git repository with initial commit for testing.

    This fixture creates a complete Git repository with commits
    for testing commands that interact with Git.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Yields:
        Path to Git repository with initial commit
    """
    repo_dir = tmp_path / "git_repo_with_commit"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository
    subprocess.run(
        ["git", "init"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Configure Git user
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.email", "test@example.com"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Create README and initial commit
    readme = repo_dir / "README.md"
    readme.write_text("# Test Repository\n\nFor testing Git integration.\n", encoding="utf-8")
    subprocess.run(["git", "add", "."], cwd=repo_dir, check=True, capture_output=True)
    subprocess.run(
        ["git", "commit", "-m", "Initial commit"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Rename to main if needed
    try:
        subprocess.run(
            ["git", "branch", "-M", "main"],
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )
    except subprocess.CalledProcessError:
        pass  # Already on main

    yield repo_dir


@pytest.fixture
def cli_runner_with_cwd():
    """Provide a CliRunner that can run commands from a specific directory.

    Returns:
        CliRunner instance that can accept cwd parameter
    """
    return CliRunner()
