"""
@FEATURE:RESOURCE-VALIDATOR-001 Resource Validation for MoAI-ADK

@TASK:VALIDATOR-001 Handles path validation, security checks, and resource verification
@DESIGN:MODULE-SPLIT-001 Extracted from resource_manager.py for TRUST principle compliance
"""

import logging
from pathlib import Path

logger = logging.getLogger(__name__)


class ResourceValidator:
    """Resource validation and security manager."""

    def __init__(self):
        """Initialize resource validator."""
        self.dangerous_paths = {
            "/",
            "/bin",
            "/sbin",
            "/usr",
            "/etc",
            "/var",
            "/tmp",
            "/boot",
            "/dev",
            "/proc",
            "/sys",
            Path.home(),
        }

    def validate_safe_path(self, target_path: Path) -> bool:
        """
        Validate that target path is safe for operations.

        Args:
            target_path: Path to validate

        Returns:
            True if safe, False if dangerous
        """
        try:
            resolved_path = target_path.resolve()

            # Check against dangerous paths
            for dangerous_path in self.dangerous_paths:
                dangerous_resolved = Path(dangerous_path).resolve()

                # Prevent operations in system directories
                if resolved_path == dangerous_resolved:
                    logger.warning(f"Dangerous path detected: {resolved_path}")
                    return False

                # Prevent operations in parent system directories
                try:
                    resolved_path.relative_to(dangerous_resolved)
                    logger.warning(f"Path in dangerous directory: {resolved_path}")
                    return False
                except ValueError:
                    # Not a subpath, which is good
                    continue

            # Check for path traversal attempts
            path_str = str(resolved_path)
            if ".." in target_path.parts or ".." in path_str:
                logger.warning(f"Path traversal detected: {target_path}")
                return False

            return True

        except Exception as e:
            logger.error(f"Error validating path {target_path}: {e}")
            return False

    def validate_project_resources(self, project_path: Path) -> bool:
        """
        Validate project has required resources.

        Args:
            project_path: Project root path

        Returns:
            True if valid, False otherwise
        """
        try:
            required_items = [
                project_path / ".claude",
                project_path / ".moai",
                project_path / "CLAUDE.md",
            ]

            for item in required_items:
                if not item.exists():
                    logger.warning(f"Required resource missing: {item}")
                    return False

            logger.info("All required project resources validated")
            return True

        except Exception as e:
            logger.error(f"Failed to validate project resources: {e}")
            return False

    def validate_required_resources(self, project_path: Path) -> bool:
        """
        Validate all required MoAI-ADK resources exist.

        Args:
            project_path: Project root path

        Returns:
            True if all required resources exist, False otherwise
        """
        try:
            required_paths = [
                project_path / ".claude",
                project_path / ".moai",
                project_path / "CLAUDE.md",
            ]

            missing_paths = []
            for path in required_paths:
                if not path.exists():
                    missing_paths.append(path.name)

            if missing_paths:
                logger.error(f"Missing required resources: {', '.join(missing_paths)}")
                return False

            return True

        except Exception as e:
            logger.error(f"Failed to validate required resources: {e}")
            return False

    def validate_clean_installation(self, target_path: Path) -> bool:
        """
        Validate the target path is suitable for clean installation.

        Args:
            target_path: Target installation path

        Returns:
            True if suitable for clean installation, False otherwise
        """
        try:
            # Check if path is safe
            if not self.validate_safe_path(target_path):
                return False

            # For clean installation, check if MoAI-ADK files already exist
            moai_files = [
                target_path / ".moai",
                target_path / ".claude",
                target_path / "CLAUDE.md",
            ]

            existing_files = [f for f in moai_files if f.exists()]

            if existing_files:
                logger.warning(f"Existing MoAI-ADK files found: {[f.name for f in existing_files]}")
                logger.warning("Use --force to overwrite or --backup to backup first")
                return False

            return True

        except Exception as e:
            logger.error(f"Failed to validate clean installation path: {e}")
            return False

    def check_path_conflicts(self, project_path: Path) -> list[str]:
        """
        Check for potential path conflicts.

        Args:
            project_path: Project path to check

        Returns:
            List of conflict descriptions
        """
        conflicts = []

        try:
            # Check for existing MoAI-ADK files
            if (project_path / ".moai").exists():
                conflicts.append("Existing .moai directory")

            if (project_path / ".claude").exists():
                conflicts.append("Existing .claude directory")

            if (project_path / "CLAUDE.md").exists():
                conflicts.append("Existing CLAUDE.md file")

            # Check for Git repository
            if (project_path / ".git").exists():
                conflicts.append("Git repository detected")

            # Check for common project files
            common_files = ["package.json", "requirements.txt", "Cargo.toml", "pom.xml"]
            for file_name in common_files:
                if (project_path / file_name).exists():
                    conflicts.append(f"Existing {file_name} project file")
                    break  # Only report one project file type

        except Exception as e:
            logger.error(f"Failed to check path conflicts: {e}")
            conflicts.append(f"Error checking conflicts: {e}")

        return conflicts

    def get_project_status(self, project_path: Path) -> dict[str, bool]:
        """
        Get comprehensive project status.

        Args:
            project_path: Project path to analyze

        Returns:
            Dictionary with status information
        """
        try:
            return {
                "moai_initialized": (project_path / ".moai").exists(),
                "claude_initialized": (project_path / ".claude").exists(),
                "memory_file": (project_path / "CLAUDE.md").exists(),
                "git_repository": (project_path / ".git").exists(),
                "has_conflicts": bool(self.check_path_conflicts(project_path)),
                "safe_path": self.validate_safe_path(project_path),
            }
        except Exception as e:
            logger.error(f"Failed to get project status: {e}")
            return {
                "moai_initialized": False,
                "claude_initialized": False,
                "memory_file": False,
                "git_repository": False,
                "has_conflicts": True,
                "safe_path": False,
            }