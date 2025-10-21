# @CODE:INIT-003:PHASE | SPEC: .moai/specs/SPEC-INIT-003/spec.md | TEST: tests/unit/test_init_reinit.py
"""Phase-based installation executor (SPEC-INIT-003 v0.4.2)

Runs the project initialization across five phases:
- Phase 1: Preparation (create single backup at .moai-backups/backup/)
- Phase 2: Directory (build directory structure)
- Phase 3: Resource (copy templates while preserving user content)
- Phase 4: Configuration (generate configuration files)
- Phase 5: Validation (verify and finalize)
"""

import json
import shutil
import subprocess
from collections.abc import Callable
from datetime import datetime
from pathlib import Path

from rich.console import Console

from moai_adk import __version__
from moai_adk.core.project.backup_utils import (
    get_backup_targets,
    has_any_moai_files,
    is_protected_path,
)
from moai_adk.core.project.validator import ProjectValidator

console = Console()

# Progress callback type alias
ProgressCallback = Callable[[str, int, int], None]


class PhaseExecutor:
    """Execute the installation across the five phases.

    Phases:
    1. Preparation: Back up and verify the system.
    2. Directory: Create the directory structure.
    3. Resource: Copy template resources.
    4. Configuration: Generate configuration files.
    5. Validation: Perform final checks.
    """

    # Required directory structure
    REQUIRED_DIRECTORIES = [
        ".moai/",
        ".moai/project/",
        ".moai/specs/",
        ".moai/reports/",
        ".moai/memory/",
        ".claude/",
        ".claude/logs/",
        ".github/",
    ]

    def __init__(self, validator: ProjectValidator) -> None:
        """Initialize the executor.

        Args:
            validator: Project validation helper.
        """
        self.validator = validator
        self.total_phases = 5
        self.current_phase = 0

    def execute_preparation_phase(
        self,
        project_path: Path,
        backup_enabled: bool = True,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 1: preparation and backup.

        Args:
            project_path: Project path.
            backup_enabled: Whether backups are enabled.
            progress_callback: Optional progress callback.
        """
        self.current_phase = 1
        self._report_progress(
            "Phase 1: Preparation and backup...", progress_callback
        )

        # Validate system requirements
        self.validator.validate_system_requirements()

        # Verify the project path
        self.validator.validate_project_path(project_path)

        # Create a backup when needed
        if backup_enabled and has_any_moai_files(project_path):
            self._create_backup(project_path)

    def execute_directory_phase(
        self,
        project_path: Path,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 2: create directories.

        Args:
            project_path: Project path.
            progress_callback: Optional progress callback.
        """
        self.current_phase = 2
        self._report_progress(
            "Phase 2: Creating directory structure...", progress_callback
        )

        for directory in self.REQUIRED_DIRECTORIES:
            dir_path = project_path / directory
            dir_path.mkdir(parents=True, exist_ok=True)

    def execute_resource_phase(
        self,
        project_path: Path,
        config: dict[str, str] | None = None,
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 3: install resources with variable substitution.

        Args:
            project_path: Project path.
            config: Configuration dictionary for template variable substitution.
            progress_callback: Optional progress callback.

        Returns:
            List of created files or directories.
        """
        self.current_phase = 3
        self._report_progress(
            "Phase 3: Installing resources...", progress_callback
        )

        # Copy resources via TemplateProcessor in silent mode
        from moai_adk.core.template import TemplateProcessor

        processor = TemplateProcessor(project_path)

        # Set template variable context (if provided)
        if config:
            context = {
                "MOAI_VERSION": __version__,
                "CREATION_TIMESTAMP": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "PROJECT_NAME": config.get("name", "unknown"),
                "PROJECT_DESCRIPTION": config.get("description", ""),
                "PROJECT_MODE": config.get("mode", "personal"),
                "PROJECT_VERSION": config.get("version", "0.1.0"),
                "AUTHOR": config.get("author", "@user"),
            }
            processor.set_context(context)

        processor.copy_templates(backup=False, silent=True)  # Avoid progress bar conflicts

        # Return a simplified list of generated assets
        return [
            ".claude/",
            ".moai/",
            ".github/",
            "CLAUDE.md",
            ".gitignore",
        ]

    def execute_configuration_phase(
        self,
        project_path: Path,
        config: dict[str, str | bool],
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 4: generate configuration.

        Args:
            project_path: Project path.
            config: Configuration dictionary.
            progress_callback: Optional progress callback.

        Returns:
            List of created files.
        """
        self.current_phase = 4
        self._report_progress(
            "Phase 4: Generating configurations...", progress_callback
        )

        # Attach version metadata (v0.3.1+)
        config["moai_adk_version"] = __version__
        config["optimized"] = False  # Default value

        # Write config.json
        config_path = project_path / ".moai" / "config.json"
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

        return [str(config_path)]

    def execute_validation_phase(
        self,
        project_path: Path,
        mode: str = "personal",
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 5: validation and wrap-up.

        @CODE:INIT-004:PHASE5 | Phase 5 verification logic
        @REQ:VALIDATION-001 | SPEC-INIT-004: Verify required files after initialization completion
        @CODE:INIT-004:PHASE5-INTEGRATION | Integration of validation in Phase 5

        Args:
            project_path: Project path.
            mode: Project mode (personal/team).
            progress_callback: Optional progress callback.
        """
        self.current_phase = 5
        self._report_progress(
            "Phase 5: Validation and finalization...", progress_callback
        )

        # @CODE:INIT-004:VERIFY-001 | Validate installation results
        # @CODE:INIT-004:VALIDATION-CHECK | Comprehensive installation validation
        # Verifies all required files including 4 Alfred command files:
        # - 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
        self.validator.validate_installation(project_path)

        # Initialize Git for team mode
        if mode == "team":
            self._initialize_git(project_path)

    def _create_backup(self, project_path: Path) -> None:
        """Create a single backup (v0.4.2).

        Maintains only one backup at .moai-backups/backup/.

        Args:
            project_path: Project path.
        """
        # Define backup directory
        backups_dir = project_path / ".moai-backups"
        backup_path = backups_dir / "backup"

        # Remove existing backup if present
        if backup_path.exists():
            shutil.rmtree(backup_path)

        backup_path.mkdir(parents=True, exist_ok=True)

        # Collect backup targets
        targets = get_backup_targets(project_path)
        backed_up_files: list[str] = []

        # Execute the backup
        for target in targets:
            src_path = project_path / target
            dst_path = backup_path / target

            if src_path.is_dir():
                self._copy_directory_selective(src_path, dst_path)
                backed_up_files.append(f"{target}/")
            else:
                dst_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src_path, dst_path)
                backed_up_files.append(target)

        # Avoid additional console messages to prevent progress bar conflicts

    def _copy_directory_selective(self, src: Path, dst: Path) -> None:
        """Copy a directory while skipping protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)

            # Skip protected paths
            if is_protected_path(rel_path):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _initialize_git(self, project_path: Path) -> None:
        """Initialize a Git repository.

        Args:
            project_path: Project path.
        """
        try:
            subprocess.run(
                ["git", "init"],
                cwd=project_path,
                check=True,
                capture_output=True,
            )
            # Intentionally avoid printing to keep progress output clean
        except subprocess.CalledProcessError:
            # Only log on error; failures are non-fatal
            pass

    def _report_progress(
        self, message: str, callback: ProgressCallback | None
    ) -> None:
        """Report progress.

        Args:
            message: Progress message.
            callback: Callback function.
        """
        if callback:
            callback(message, self.current_phase, self.total_phases)
