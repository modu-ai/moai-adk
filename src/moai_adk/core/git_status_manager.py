"""
@FEATURE:GIT-003 Git status management for MoAI-ADK

Handles Git repository status checking, remote information,
and repository state analysis with security validation.
"""

import subprocess
from pathlib import Path

from ..utils.logger import get_logger
from .security import SecurityManager

logger = get_logger(__name__)


class GitStatusManager:
    """@TASK:GIT-STATUS-001 Manages Git repository status and information."""

    def __init__(self, security_manager: SecurityManager = None):
        """
        Initialize Git status manager.

        Args:
            security_manager: Security manager instance for validation
        """
        self.security_manager = security_manager or SecurityManager()

    def check_git_status(self, project_path: Path) -> dict:
        """
        Check the status of Git repository.

        Args:
            project_path: Project root path

        Returns:
            dict: Git status information
        """
        try:
            if not (project_path / ".git").exists():
                return {"is_git_repo": False, "error": "Not a git repository"}

            # Security validation
            if not self.security_manager.validate_subprocess_path(
                project_path, project_path
            ):
                return {
                    "is_git_repo": True,
                    "error": "Security validation failed for git status",
                }

            # Get git status
            result = subprocess.run(
                ["git", "status", "--porcelain"],
                cwd=project_path,
                capture_output=True,
                text=True,
                check=True,
            )

            # Parse status output
            status_lines = (
                result.stdout.strip().split("\n") if result.stdout.strip() else []
            )

            modified_files = []
            untracked_files = []
            staged_files = []

            for line in status_lines:
                if len(line) >= 3:
                    status_code = line[:2]
                    filename = line[3:]

                    if status_code[0] in ["M", "A", "D", "R", "C"]:
                        staged_files.append(filename)
                    if status_code[1] in ["M", "D"]:
                        modified_files.append(filename)
                    if status_code == "??":
                        untracked_files.append(filename)

            return {
                "is_git_repo": True,
                "is_clean": len(status_lines) == 0,
                "modified_files": modified_files,
                "untracked_files": untracked_files,
                "staged_files": staged_files,
                "total_changes": len(status_lines),
            }

        except subprocess.CalledProcessError as e:
            logger.error("Git status command failed: %s", e.stderr)
            return {"is_git_repo": True, "error": f"Git command failed: {e.stderr}"}
        except Exception as e:
            logger.error("Error checking git status: %s", e)
            return {"is_git_repo": False, "error": str(e)}

    def get_remote_info(self, project_path: Path) -> dict:
        """Get Git remote information."""
        try:
            # Security validation
            if not self.security_manager.validate_subprocess_path(
                project_path, project_path
            ):
                return {"error": "Security validation failed"}

            result = subprocess.run(
                ["git", "remote", "-v"],
                cwd=project_path,
                capture_output=True,
                text=True,
                check=True,
            )

            remotes = {}
            for line in result.stdout.strip().split("\n"):
                if line:
                    parts = line.split()
                    if len(parts) >= 2:
                        remote_name = parts[0]
                        remote_url = parts[1]
                        remote_type = parts[2] if len(parts) > 2 else ""

                        if remote_name not in remotes:
                            remotes[remote_name] = {}

                        if "(fetch)" in remote_type:
                            remotes[remote_name]["fetch"] = remote_url
                        elif "(push)" in remote_type:
                            remotes[remote_name]["push"] = remote_url

            return {"remotes": remotes}

        except subprocess.CalledProcessError as e:
            return {"error": f"Git remote command failed: {e.stderr}"}
        except Exception as e:
            return {"error": str(e)}

    def get_comprehensive_git_info(self, project_path: Path) -> dict:
        """
        Get comprehensive Git repository information.

        Args:
            project_path: Project root path

        Returns:
            dict: Git repository information
        """
        git_info = {
            "git_available": self._check_git_available(),
            "is_git_repo": (project_path / ".git").exists(),
            "status": {},
            "remote_info": {},
        }

        if git_info["git_available"] and git_info["is_git_repo"]:
            git_info["status"] = self.check_git_status(project_path)
            git_info["remote_info"] = self.get_remote_info(project_path)

        return git_info

    def _check_git_available(self) -> bool:
        """Check if git is available in the system."""
        try:
            subprocess.run(
                ["git", "--version"], capture_output=True, text=True, check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            return False