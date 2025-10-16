# @CODE:CORE-PROJECT-003 | SPEC: SPEC-CORE-PROJECT-001.md
"""Project initialization validation module.

Validates system requirements and installation results.
"""

import shutil
from pathlib import Path


class ValidationError(Exception):
    """Raised when validation fails."""

    pass


class ProjectValidator:
    """Validate project initialization."""

    # Required directory structure
    REQUIRED_DIRECTORIES = [
        ".moai/",
        ".moai/project/",
        ".moai/specs/",
        ".moai/memory/",
        ".claude/",
    ]

    # Required files
    REQUIRED_FILES = [
        ".moai/config.json",
        "CLAUDE.md",
    ]

    def validate_system_requirements(self) -> None:
        """Verify system requirements.

        Raises:
            ValidationError: Raised when requirements are not satisfied.
        """
        # Ensure Git is installed
        if not shutil.which("git"):
            raise ValidationError("Git is not installed")

        # Check Python version (3.10+)
        import sys

        if sys.version_info < (3, 10):
            raise ValidationError(
                f"Python 3.10+ required (current: {sys.version_info.major}.{sys.version_info.minor})"
            )

    def validate_project_path(self, project_path: Path) -> None:
        """Verify the project path.

        Args:
            project_path: Project path.

        Raises:
            ValidationError: Raised when the path is invalid.
        """
        # Must be an absolute path
        if not project_path.is_absolute():
            raise ValidationError(f"Project path must be absolute: {project_path}")

        # Parent directory must exist
        if not project_path.parent.exists():
            raise ValidationError(f"Parent directory does not exist: {project_path.parent}")

        # Prevent initialization inside the MoAI-ADK package
        if self._is_inside_moai_package(project_path):
            raise ValidationError(
                "Cannot initialize inside MoAI-ADK package directory"
            )

    def validate_installation(self, project_path: Path) -> None:
        """Validate installation results.

        Args:
            project_path: Project path.

        Raises:
            ValidationError: Raised when installation was incomplete.
        """
        # Verify required directories
        for directory in self.REQUIRED_DIRECTORIES:
            dir_path = project_path / directory
            if not dir_path.exists():
                raise ValidationError(f"Required directory not found: {directory}")

        # Verify required files
        for file in self.REQUIRED_FILES:
            file_path = project_path / file
            if not file_path.exists():
                raise ValidationError(f"Required file not found: {file}")

    def _is_inside_moai_package(self, project_path: Path) -> bool:
        """Determine whether the path is inside the MoAI-ADK package.

        Args:
            project_path: Path to check.

        Returns:
            True when the path resides within the package.
        """
        # The package root contains a pyproject.toml referencing moai-adk
        current = project_path.resolve()
        while current != current.parent:
            pyproject = current / "pyproject.toml"
            if pyproject.exists():
                try:
                    content = pyproject.read_text(encoding="utf-8")
                    if "name = \"moai-adk\"" in content or 'name = "moai-adk"' in content:
                        return True
                except Exception:
                    pass
            current = current.parent
        return False
