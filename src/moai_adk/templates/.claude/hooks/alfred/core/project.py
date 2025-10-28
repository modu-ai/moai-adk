#!/usr/bin/env python3
"""Project metadata utilities

Project information inquiry (language, Git, SPEC progress, etc.)
"""

import json
import signal
import subprocess
from contextlib import contextmanager
from pathlib import Path
from typing import Any


class TimeoutError(Exception):
    """Signal-based timeout exception"""
    pass


@contextmanager
def timeout_handler(seconds: int):
    """Hard timeout using SIGALRM (works on Unix systems including macOS)

    This uses kernel-level signal to interrupt ANY blocking operation,
    even if subprocess.run() timeout fails on macOS.

    Args:
        seconds: Timeout duration in seconds

    Raises:
        TimeoutError: If operation exceeds timeout
    """
    def _handle_timeout(signum, frame):
        raise TimeoutError(f"Operation timed out after {seconds} seconds")

    # Set the signal handler
    old_handler = signal.signal(signal.SIGALRM, _handle_timeout)
    signal.alarm(seconds)
    try:
        yield
    finally:
        signal.alarm(0)  # Disable alarm
        signal.signal(signal.SIGALRM, old_handler)


def detect_language(cwd: str) -> str:
    """Detect project language (supports 20 items languages)

    Browse the File system to detect your project's main development language.
    First, check configuration files such as pyproject.toml and tsconfig.json.
    Apply TypeScript first principles (if tsconfig.json exists).

    Args:
        cwd: Project root directory path (both absolute and relative paths are possible)

    Returns:
        Detected language name (lowercase). If detection fails, "Unknown Language" is returned.
        Supported languages: python, typescript, javascript, java, go, rust,
                  dart, swift, kotlin, php, ruby, elixir, scala,
                  clojure, cpp, c, csharp, haskell, shell, lua

    Examples:
        >>> detect_language("/path/to/python/project")
        'python'
        >>> detect_language("/path/to/typescript/project")
        'typescript'
        >>> detect_language("/path/to/unknown/project")
        'Unknown Language'

    TDD History:
        - RED: Write a 21 items language detection test (20 items language + 1 items unknown)
        - GREEN: 20 items language + unknown implementation, all tests passed
        - REFACTOR: Optimize file inspection order, apply TypeScript priority principle
    """
    cwd_path = Path(cwd)

    # Language detection mapping
    language_files = {
        "pyproject.toml": "python",
        "tsconfig.json": "typescript",
        "package.json": "javascript",
        "pom.xml": "java",
        "go.mod": "go",
        "Cargo.toml": "rust",
        "pubspec.yaml": "dart",
        "Package.swift": "swift",
        "build.gradle.kts": "kotlin",
        "composer.json": "php",
        "Gemfile": "ruby",
        "mix.exs": "elixir",
        "build.sbt": "scala",
        "project.clj": "clojure",
        "CMakeLists.txt": "cpp",
        "Makefile": "c",
    }

    # Check standard language files
    for file_name, language in language_files.items():
        if (cwd_path / file_name).exists():
            # Special handling for package.json - prefer typescript if tsconfig exists
            if file_name == "package.json" and (cwd_path / "tsconfig.json").exists():
                return "typescript"
            return language

    # Check for C# project files (*.csproj)
    if any(cwd_path.glob("*.csproj")):
        return "csharp"

    # Check for Haskell project files (*.cabal)
    if any(cwd_path.glob("*.cabal")):
        return "haskell"

    # Check for Shell scripts (*.sh)
    if any(cwd_path.glob("*.sh")):
        return "shell"

    # Check for Lua files (*.lua)
    if any(cwd_path.glob("*.lua")):
        return "lua"

    return "Unknown Language"


def _run_git_command(args: list[str], cwd: str, timeout: int = 2) -> str:
    """Git command execution with HARD timeout protection

    Safely execute Git commands and return output.
    Uses SIGALRM (kernel-level interrupt) to handle macOS subprocess timeout bug.
    Eliminates code duplication and provides consistent error handling.

    Args:
        args: Git command argument list (git adds automatically)
        cwd: Execution directory path
        timeout: Timeout (seconds, default 2 seconds)

    Returns:
        Git command output (stdout, removing leading and trailing spaces)

    Raises:
        subprocess.TimeoutExpired: Timeout exceeded (via TimeoutError)
        subprocess.CalledProcessError: Git command failed

    Examples:
        >>> _run_git_command(["branch", "--show-current"], ".")
        'main'

    TDD History:
        - RED: Git command hang scenario test
        - GREEN: SIGALRM-based timeout implementation
        - REFACTOR: Exception conversion to subprocess.TimeoutExpired
    """
    try:
        with timeout_handler(timeout):
            result = subprocess.run(
                ["git"] + args,
                cwd=cwd,
                capture_output=True,
                text=True,
                check=False,  # Don't raise on non-zero exit - we'll check manually
            )

            # Check exit code manually
            if result.returncode != 0:
                raise subprocess.CalledProcessError(
                    result.returncode, ["git"] + args, result.stdout, result.stderr
                )

            return result.stdout.strip()

    except TimeoutError:
        # Convert to subprocess.TimeoutExpired for consistent error handling
        raise subprocess.TimeoutExpired(["git"] + args, timeout)


def get_git_info(cwd: str) -> dict[str, Any]:
    """Gather Git repository information

    View the current status of a Git repository.
    Returns the branch name, commit hash, number of changes, and last commit message.
    If it is not a Git repository, it returns an empty dictionary.

    Args:
        cwd: Project root directory path

    Returns:
        Git information dictionary. Includes the following keys:
        - branch: Current branch name (str)
        - commit: Current commit hash (str, full hash)
        - changes: Number of changed files (int, staged + unstaged)
        - last_commit: Last commit message (str, subject only)

        Empty dictionary {} if it is not a Git repository or the query fails.

    Examples:
        >>> get_git_info("/path/to/git/repo")
        {'branch': 'main', 'commit': 'abc123...', 'changes': 3, 'last_commit': 'Fix bug'}
        >>> get_git_info("/path/to/non-git")
        {}

    Notes:
        - Timeout: 2 seconds for each Git command
        - Security: Safe execution with subprocess.run(shell=False)
        - Error handling: Returns an empty dictionary in case of all exceptions
        - Commit message limited to 50 characters for display purposes

    TDD History:
        - RED: 3 items scenario test (Git repo, non-Git, error)
        - GREEN: Implementation of subprocess-based Git command execution
        - REFACTOR: Add timeout (2 seconds), strengthen exception handling, remove duplicates with helper function
        - UPDATE: Added last_commit message field for SessionStart display
    """
    try:
        # Check if it's a git repository
        _run_git_command(["rev-parse", "--git-dir"], cwd)

        # Get branch name, commit hash, and changes
        branch = _run_git_command(["branch", "--show-current"], cwd)
        commit = _run_git_command(["rev-parse", "HEAD"], cwd)
        status_output = _run_git_command(["status", "--short"], cwd)
        changes = len([line for line in status_output.splitlines() if line])

        # Get last commit message (subject only, limited to 50 chars)
        last_commit = _run_git_command(["log", "-1", "--format=%s"], cwd)
        if len(last_commit) > 50:
            last_commit = last_commit[:47] + "..."

        return {
            "branch": branch,
            "commit": commit,
            "changes": changes,
            "last_commit": last_commit,
        }

    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        return {}


def count_specs(cwd: str) -> dict[str, int]:
    """SPEC File count and progress calculation

    Browse the .moai/specs/ directory to find the number of SPEC Files and
    Counts the number of SPECs with status: completed.

    Args:
        cwd: Project root directory path

    Returns:
        SPEC progress dictionary. Includes the following keys:
        - completed: Number of completed SPECs (int)
        - total: total number of SPECs (int)
        - percentage: completion percentage (int, 0~100)

        All 0 if .moai/specs/ directory does not exist

    Examples:
        >>> count_specs("/path/to/project")
        {'completed': 2, 'total': 5, 'percentage': 40}
        >>> count_specs("/path/to/no-specs")
        {'completed': 0, 'total': 0, 'percentage': 0}

    Notes:
        - SPEC File Location: .moai/specs/SPEC-{ID}/spec.md
        - Completion condition: Include “status: completed” in YAML front matter
        - If parsing fails, the SPEC is considered incomplete.

    TDD History:
        - RED: 5 items scenario test (0/0, 2/5, 5/5, no directory, parsing error)
        - GREEN: SPEC search with Path.iterdir(), YAML parsing implementation
        - REFACTOR: Strengthened exception handling, improved percentage calculation safety
    """
    specs_dir = Path(cwd) / ".moai" / "specs"

    if not specs_dir.exists():
        return {"completed": 0, "total": 0, "percentage": 0}

    completed = 0
    total = 0

    for spec_dir in specs_dir.iterdir():
        if not spec_dir.is_dir() or not spec_dir.name.startswith("SPEC-"):
            continue

        spec_file = spec_dir / "spec.md"
        if not spec_file.exists():
            continue

        total += 1

        # Parse YAML front matter
        try:
            content = spec_file.read_text()
            if content.startswith("---"):
                yaml_end = content.find("---", 3)
                if yaml_end > 0:
                    yaml_content = content[3:yaml_end]
                    if "status: completed" in yaml_content:
                        completed += 1
        except (OSError, UnicodeDecodeError):
            # File read failure or encoding error - considered incomplete
            pass

    percentage = int(completed / total * 100) if total > 0 else 0

    return {
        "completed": completed,
        "total": total,
        "percentage": percentage,
    }


def get_project_language(cwd: str) -> str:
    """Determine the primary project language (prefers config.json).

    Args:
        cwd: Project root directory.

    Returns:
        Language string in lower-case.

    Notes:
        - Reads ``.moai/config.json`` first for a quick answer.
        - Falls back to ``detect_language`` if configuration is missing.
    """
    config_path = Path(cwd) / ".moai" / "config.json"
    if config_path.exists():
        try:
            config = json.loads(config_path.read_text())
            lang = config.get("language", "")
            if lang:
                return lang
        except (OSError, json.JSONDecodeError):
            # Fall back to detection on parse errors
            pass

    # Fall back to the original language detection routine
    return detect_language(cwd)


def get_package_version_info() -> dict[str, Any]:
    """Check MoAI-ADK current and latest version from PyPI

    Compares the installed version with the latest version available on PyPI.
    Returns version information for SessionStart hook to display update recommendations.

    Returns:
        dict with keys:
            - "current": Current installed version
            - "latest": Latest version available on PyPI
            - "update_available": Boolean indicating if update is available
            - "upgrade_command": Recommended upgrade command (if update available)

    Note:
        - Has 1-second timeout to avoid blocking SessionStart
        - Returns graceful fallback if PyPI check fails
        - Handles version parsing gracefully
    """
    import urllib.error
    import urllib.request
    from importlib.metadata import PackageNotFoundError, version

    result = {
        "current": "unknown",
        "latest": "unknown",
        "update_available": False,
        "upgrade_command": ""
    }

    # Get current version
    try:
        result["current"] = version("moai-adk")
    except PackageNotFoundError:
        result["current"] = "dev"
        return result

    # Get latest version from PyPI (with 1-second timeout)
    try:
        with timeout_handler(1):
            url = "https://pypi.org/pypi/moai-adk/json"
            headers = {"Accept": "application/json"}
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=0.8) as response:
                data = json.load(response)
                result["latest"] = data.get("info", {}).get("version", "unknown")
    except (urllib.error.URLError, TimeoutError, Exception):
        # Network error or timeout - return with unknown latest version
        return result

    # Compare versions (simple comparison)
    if result["current"] != "unknown" and result["latest"] != "unknown":
        try:
            # Parse versions for comparison
            current_parts = [int(x) for x in result["current"].split(".")]
            latest_parts = [int(x) for x in result["latest"].split(".")]

            # Pad shorter version with zeros
            max_len = max(len(current_parts), len(latest_parts))
            current_parts.extend([0] * (max_len - len(current_parts)))
            latest_parts.extend([0] * (max_len - len(latest_parts)))

            if latest_parts > current_parts:
                result["update_available"] = True
                result["upgrade_command"] = f"uv pip install --upgrade moai-adk>={result['latest']}"
        except (ValueError, AttributeError):
            # Version parsing failed - skip comparison
            pass

    return result


__all__ = [
    "detect_language",
    "get_git_info",
    "count_specs",
    "get_project_language",
    "get_package_version_info",
]
