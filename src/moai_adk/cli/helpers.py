#!/usr/bin/env python3
"""
@FEATURE:CLI-HELPERS-001 CLI Helper Functions for MoAI-ADK

@TASK:CLI-UTILS-001 Contains utility functions used by the CLI commands including backup,
conflict detection, environment validation, and project analysis.
"""

import datetime
import shutil
from pathlib import Path

import click

from .._version import __version__, get_version_format
from ..core.resource_version import ResourceVersionManager
from ..core.validator import validate_python_version
from ..install.resource_manager import ResourceManager
from ..utils.logger import get_logger

logger = get_logger(__name__)


def create_installation_backup(project_path: Path) -> bool:
    """@TASK:BACKUP-001 Create a backup of existing MoAI-ADK installation.

    Args:
        project_path: Path to the project directory

    Returns:
        bool: True if backup was created successfully, False otherwise
    """
    try:
        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = project_path / f".moai_backup_{timestamp}"

        # Create backup directory
        backup_dir.mkdir(exist_ok=True)

        # Backup .moai directory
        if (project_path / ".moai").exists():
            shutil.copytree(project_path / ".moai", backup_dir / ".moai")

        # Backup .claude directory
        if (project_path / ".claude").exists():
            shutil.copytree(project_path / ".claude", backup_dir / ".claude")

        # Backup CLAUDE.md if exists
        if (project_path / "CLAUDE.md").exists():
            shutil.copy2(project_path / "CLAUDE.md", backup_dir / "CLAUDE.md")

        # Create backup info file
        backup_info = backup_dir / "backup_info.txt"
        backup_info.write_text(f"""MoAI-ADK Backup Information
Created: {datetime.datetime.now().isoformat()}
Original Path: {project_path.absolute()}
Backup Contents:
- .moai/ directory configuration
- .claude/ directory configuration
- CLAUDE.md project memory file

To restore this backup:
1. Remove current .moai/ and .claude/ directories
2. Copy contents from this backup directory back to original location
3. Restart Claude Code if running
""")

        logger.info(f"Backup created at: {backup_dir}")
        return True

    except Exception as e:
        logger.error(f"Failed to create backup: {e}")
        return False


def detect_potential_conflicts(project_path: Path) -> list[str]:
    """Detect potential conflicts with existing files/directories.

    Args:
        project_path: Path to check for conflicts

    Returns:
        list: List of potential conflict descriptions
    """
    conflicts = []

    # Check for existing .moai directory
    if (project_path / ".moai").exists():
        moai_files = list((project_path / ".moai").rglob("*"))
        if moai_files:
            conflicts.append(f"Existing .moai directory with {len(moai_files)} files")

    # Check for existing .claude directory
    if (project_path / ".claude").exists():
        claude_files = list((project_path / ".claude").rglob("*"))
        if claude_files:
            conflicts.append(
                f"Existing .claude directory with {len(claude_files)} files"
            )

    # Check for existing CLAUDE.md
    if (project_path / "CLAUDE.md").exists():
        conflicts.append("Existing CLAUDE.md file")

    # Check for package.json (might interfere with Node.js projects)
    if (project_path / "package.json").exists():
        conflicts.append("Existing package.json (Node.js project detected)")

    # Check for pyproject.toml (Python project)
    if (project_path / "pyproject.toml").exists():
        conflicts.append("Existing pyproject.toml (Python project detected)")

    return conflicts


def analyze_existing_project(project_path: Path) -> dict:
    """Analyze existing project structure and provide recommendations.

    Args:
        project_path: Path to analyze

    Returns:
        dict: Analysis results with recommendations
    """
    analysis = {
        "project_type": "unknown",
        "existing_tools": [],
        "recommendations": [],
        "compatibility": "unknown",
    }

    try:
        # Detect project type
        if (project_path / "package.json").exists():
            analysis["project_type"] = "nodejs"
            analysis["existing_tools"].append("npm/yarn")

        if (project_path / "pyproject.toml").exists():
            analysis["project_type"] = "python"
            analysis["existing_tools"].append("poetry/pip")

        if (project_path / "Cargo.toml").exists():
            analysis["project_type"] = "rust"
            analysis["existing_tools"].append("cargo")

        if (project_path / ".git").exists():
            analysis["existing_tools"].append("git")

        # Check for existing AI tools
        ai_tools = []
        if (project_path / ".cursor").exists():
            ai_tools.append("Cursor")

        if (project_path / ".github" / "copilot.yml").exists():
            ai_tools.append("GitHub Copilot")

        if any((project_path / ".vscode").glob("*.json")):
            ai_tools.append("VS Code")

        analysis["existing_tools"].extend(ai_tools)

        # Generate recommendations
        if analysis["project_type"] == "unknown":
            analysis["recommendations"].append(
                "Consider initializing as a standard project type"
            )

        if "git" not in analysis["existing_tools"]:
            analysis["recommendations"].append(
                "Initialize Git repository for version control"
            )

        if ai_tools:
            analysis["recommendations"].append(
                f"Existing AI tools detected: {', '.join(ai_tools)}. MoAI-ADK can work alongside them."
            )

        # Determine compatibility
        if analysis["project_type"] in ["nodejs", "python", "rust"]:
            analysis["compatibility"] = "high"
        elif analysis["project_type"] == "unknown":
            analysis["compatibility"] = "medium"
        else:
            analysis["compatibility"] = "low"

    except Exception as e:
        logger.error(f"Error analyzing project: {e}")
        analysis["recommendations"].append("Failed to analyze project structure")

    return analysis


def print_banner(show_usage: bool = False) -> None:
    """Print MoAI-ADK banner with optional usage information."""
    from .banner import print_banner as print_ascii_banner

    print_ascii_banner()

    if show_usage:
        click.echo(f"\n{get_version_format('full')}")
        click.echo("Usage: moai [COMMAND] [OPTIONS]")
        click.echo("Run 'moai help' for more information.")


def validate_environment() -> bool:
    """Validate the current environment for MoAI-ADK installation.

    Returns:
        bool: True if environment is suitable
    """
    try:
        # Check Python version
        if not validate_python_version():
            return False

        # Check for Claude Code (optional but recommended)
        # This is a basic check - more sophisticated detection could be added
        logger.info("Environment validation passed")
        return True

    except Exception as e:
        logger.error(f"Environment validation failed: {e}")
        return False


def format_project_status(project_path: Path, config_data: dict | None = None) -> dict:
    """Format project status information for display.

    Args:
        project_path: Path to the project
        config_data: Optional configuration data

    Returns:
        dict: Formatted status information
    """
    status = {
        "path": str(project_path.absolute()),
        "moai_initialized": (project_path / ".moai").exists(),
        "claude_initialized": (project_path / ".claude").exists(),
        "memory_file": (project_path / "CLAUDE.md").exists(),
        "git_repository": (project_path / ".git").exists(),
        "project_type": "unknown",
        "file_counts": {},
    }

    # Count files in key directories
    if status["moai_initialized"]:
        moai_files = list((project_path / ".moai").rglob("*"))
        status["file_counts"]["moai"] = len([f for f in moai_files if f.is_file()])

    if status["claude_initialized"]:
        claude_files = list((project_path / ".claude").rglob("*"))
        status["file_counts"]["claude"] = len([f for f in claude_files if f.is_file()])

    # Determine project type
    if (project_path / "package.json").exists():
        status["project_type"] = "Node.js"
    elif (project_path / "pyproject.toml").exists():
        status["project_type"] = "Python"
    elif (project_path / "Cargo.toml").exists():
        status["project_type"] = "Rust"

    # Add configuration data if provided
    if config_data:
        status["config"] = config_data

    try:
        resource_manager = ResourceManager()
        version_manager = ResourceVersionManager(project_path)
        version_info = version_manager.read()
        template_version = version_info.get("template_version") or "unknown"
        available_template_version = resource_manager.get_version()

        status["versions"] = {
            "package": __version__,
            "resources": template_version,
            "available_resources": available_template_version,
            "last_updated": version_info.get("last_updated"),
            "outdated": (
                template_version != "unknown"
                and available_template_version not in (None, "unknown")
                and template_version != available_template_version
            ),
        }
    except Exception as exc:
        logger.warning("Failed to read resource version info: %s", exc)

    return status
