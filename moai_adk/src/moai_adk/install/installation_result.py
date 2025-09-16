"""
Installation result data structure for MoAI-ADK.

Contains the result of project installation operations with success status,
created files, and next steps for the user.
"""

from dataclasses import dataclass
from typing import List
from pathlib import Path

from ..config import Config


@dataclass
class InstallationResult:
    """Result of project installation with comprehensive status information."""

    success: bool
    project_path: str
    files_created: List[str]
    next_steps: List[str]
    config: Config
    errors: List[str] | None = None
    warnings: List[str] | None = None
    git_initialized: bool = False
    backup_created: str | None = None

    def __post_init__(self):
        """Initialize optional fields if not provided."""
        if self.errors is None:
            self.errors = []
        if self.warnings is None:
            self.warnings = []

    def has_errors(self) -> bool:
        """Check if installation had any errors."""
        return bool(self.errors)

    def has_warnings(self) -> bool:
        """Check if installation had any warnings."""
        return bool(self.warnings)

    def add_error(self, error: str) -> None:
        """Add an error message to the result."""
        if self.errors is None:
            self.errors = []
        self.errors.append(error)

    def add_warning(self, warning: str) -> None:
        """Add a warning message to the result."""
        if self.warnings is None:
            self.warnings = []
        self.warnings.append(warning)

    def get_summary(self) -> str:
        """Generate a summary of the installation result."""
        if self.success:
            summary = f"✅ Installation successful in {self.project_path}"
            summary += f"\n📁 Created {len(self.files_created)} files"

            if self.git_initialized:
                summary += "\n🚀 Git repository initialized"

            if self.backup_created:
                summary += f"\n💾 Backup created at: {self.backup_created}"

            if self.warnings:
                summary += f"\n⚠️ {len(self.warnings)} warning(s) occurred"

        else:
            summary = f"❌ Installation failed in {self.project_path}"
            if self.errors:
                summary += f"\n🔥 {len(self.errors)} error(s) occurred"

        return summary

    def get_file_count_by_type(self) -> dict:
        """Count files by their extensions."""
        counts = {}
        for file_path in self.files_created:
            path = Path(file_path)
            ext = path.suffix.lower() if path.suffix else 'no_extension'
            counts[ext] = counts.get(ext, 0) + 1
        return counts

    def to_dict(self) -> dict:
        """Convert result to dictionary for serialization."""
        return {
            'success': self.success,
            'project_path': self.project_path,
            'files_created': self.files_created,
            'next_steps': self.next_steps,
            'errors': self.errors,
            'warnings': self.warnings,
            'git_initialized': self.git_initialized,
            'backup_created': self.backup_created,
            'file_count': len(self.files_created),
            'file_types': self.get_file_count_by_type()
        }

    @classmethod
    def create_success(
        cls,
        project_path: str,
        config: Config,
        files_created: List[str] = None,
        next_steps: List[str] = None,
        git_initialized: bool = False,
        backup_created: str | None = None
    ) -> 'InstallationResult':
        """Create a successful installation result."""
        return cls(
            success=True,
            project_path=project_path,
            files_created=files_created or [],
            next_steps=next_steps or [],
            config=config,
            git_initialized=git_initialized,
            backup_created=backup_created
        )

    @classmethod
    def create_failure(
        cls,
        project_path: str,
        config: Config,
        error: str,
        files_created: List[str] = None
    ) -> 'InstallationResult':
        """Create a failed installation result."""
        return cls(
            success=False,
            project_path=project_path,
            files_created=files_created or [],
            next_steps=[],
            config=config,
            errors=[error]
        )