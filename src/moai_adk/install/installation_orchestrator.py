"""
@FEATURE:INSTALL-ORCHESTRATOR-001 Installation Process Orchestrator

@TASK:ORCHESTRATE-001 Coordinates the entire MoAI-ADK installation process
following TRUST principles with clear separation of concerns.
"""

from collections.abc import Callable
from pathlib import Path

from .._version import __version__
from ..config import Config
from ..core.directory_manager import DirectoryManager
from ..core.security import SecurityManager
from ..utils.logger import get_logger
from ..utils.progress_tracker import ProgressTracker
from .configuration_generator import ConfigurationGenerator
from .installation_result import InstallationResult
from .installation_validator import InstallationValidator
from .resource_installer import ResourceInstaller

logger = get_logger(__name__)


class InstallationOrchestrator:
    """
    @TASK:ORCHESTRATOR-MAIN-001 Central coordinator for MoAI-ADK installation

    Orchestrates the installation process by delegating specific tasks to
    specialized components while maintaining overall process integrity.

    Responsibilities:
    - Overall installation flow coordination
    - Progress tracking and reporting
    - Error collection and handling
    - Backup management
    - Result compilation
    """

    def __init__(self, config: Config):
        """
        Initialize installation orchestrator

        Args:
            config: Project configuration
        """
        self.config = config
        self.progress = ProgressTracker()

        # Initialize security and directory management
        self.security_manager = SecurityManager()
        self.directory_manager = DirectoryManager(self.security_manager)

        # Initialize specialized components
        self.resource_installer = ResourceInstaller(config, self.security_manager)
        self.config_generator = ConfigurationGenerator(config, self.security_manager)
        self.validator = InstallationValidator(config, self.security_manager)

        logger.info("InstallationOrchestrator initialized for: %s", config.project_path)

    def execute_installation(
        self, progress_callback: Callable[[str, int, int], None] | None = None
    ) -> InstallationResult:
        """
        Execute complete MoAI-ADK project installation

        Args:
            progress_callback: Progress callback function

        Returns:
            InstallationResult: Complete installation result
        """
        files_created: list[str] = []
        errors: list[str] = []

        try:
            # Phase 1: Preparation and backup
            self._execute_preparation_phase(progress_callback, files_created, errors)

            # Phase 2: Directory structure creation
            self._execute_directory_phase(progress_callback, files_created, errors)

            # Phase 3: Resource installation
            self._execute_resource_phase(progress_callback, files_created, errors)

            # Phase 4: Configuration generation
            self._execute_configuration_phase(progress_callback, files_created, errors)

            # Phase 5: Validation and finalization
            self._execute_validation_phase(progress_callback, errors)

            self.progress.update_progress("Installation complete!", progress_callback)

            return InstallationResult(
                success=len(errors) == 0,
                project_path=str(self.config.project_path),
                files_created=files_created,
                errors=errors,
                next_steps=self._generate_next_steps(),
                config=self.config,
            )

        except Exception as e:
            error_msg = f"Installation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

            return InstallationResult(
                success=False,
                project_path=str(self.config.project_path),
                files_created=files_created,
                errors=errors,
                next_steps=["Fix the errors above and retry installation"],
                config=self.config,
            )

    def _execute_preparation_phase(
        self,
        progress_callback: Callable[[str, int, int], None] | None,
        files_created: list[str],
        errors: list[str],
    ) -> None:
        """Execute preparation phase including backup creation"""
        if self.config.backup_enabled:
            self.progress.update_progress("Creating backup...", progress_callback)

            try:
                from ..cli.helpers import create_installation_backup
                backup_success = create_installation_backup(self.config.project_path)
                if backup_success:
                    logger.info("Backup created successfully")
                else:
                    logger.warning("Failed to create backup, but continuing installation")
            except Exception as e:
                logger.warning("Backup creation failed: %s", e)
                # Don't fail installation for backup issues

    def _execute_directory_phase(
        self,
        progress_callback: Callable[[str, int, int], None] | None,
        files_created: list[str],
        errors: list[str],
    ) -> None:
        """Execute directory structure creation phase"""
        self.progress.update_progress("Creating project directory...", progress_callback)

        try:
            self.directory_manager.create_project_directory(self.config)
            files_created.append(str(self.config.project_path))

            # Create auxiliary directories
            self.progress.update_progress(
                "Setting up auxiliary directories...", progress_callback
            )
            directories = self._create_auxiliary_directories()
            files_created.extend([str(d) for d in directories])

        except Exception as e:
            error_msg = f"Directory creation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

    def _execute_resource_phase(
        self,
        progress_callback: Callable[[str, int, int], None] | None,
        files_created: list[str],
        errors: list[str],
    ) -> None:
        """Execute resource installation phase"""
        try:
            # Install Claude Code resources
            self.progress.update_progress(
                "Installing Claude Code resources...", progress_callback
            )
            claude_files = self.resource_installer.install_claude_resources()
            if claude_files:
                files_created.extend([str(f) for f in claude_files])

            # Install MoAI resources
            self.progress.update_progress(
                "Installing MoAI resources...", progress_callback
            )
            moai_files = self.resource_installer.install_moai_resources()
            if moai_files:
                files_created.extend([str(f) for f in moai_files])

            # Install GitHub workflows (optional)
            if self.config.include_github:
                self.progress.update_progress(
                    "Setting up GitHub workflows...", progress_callback
                )
                github_files = self.resource_installer.install_github_workflows()
                if github_files:
                    files_created.extend([str(f) for f in github_files])

            # Install project memory
            self.progress.update_progress(
                "Creating project memory...", progress_callback
            )
            memory_files = self.resource_installer.install_project_memory()
            if memory_files:
                files_created.extend([str(f) for f in memory_files])

            # Record version information
            self.resource_installer.write_version_info()

        except Exception as e:
            error_msg = f"Resource installation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

    def _execute_configuration_phase(
        self,
        progress_callback: Callable[[str, int, int], None] | None,
        files_created: list[str],
        errors: list[str],
    ) -> None:
        """Execute configuration generation phase"""
        try:
            # Generate configuration files
            self.progress.update_progress(
                "Generating configuration files...", progress_callback
            )
            config_files = self.config_generator.create_configuration_files()
            if config_files:
                files_created.extend([str(f) for f in config_files])

            # Create initial indexes
            self.progress.update_progress(
                "Creating initial indexes...", progress_callback
            )
            index_files = self.config_generator.create_initial_indexes()
            if index_files:
                files_created.extend([str(f) for f in index_files])

            # Initialize Git repository (optional)
            if self.config.initialize_git:
                self.progress.update_progress(
                    "Initializing Git repository...", progress_callback
                )
                git_files = self.config_generator.initialize_git_repository()
                if git_files:
                    files_created.extend([str(f) for f in git_files])

        except Exception as e:
            error_msg = f"Configuration generation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

    def _execute_validation_phase(
        self,
        progress_callback: Callable[[str, int, int], None] | None,
        errors: list[str],
    ) -> None:
        """Execute validation phase"""
        self.progress.update_progress("Verifying installation...", progress_callback)

        try:
            if not self.validator.verify_installation():
                errors.append("Installation verification failed")
        except Exception as e:
            error_msg = f"Validation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

    def _create_auxiliary_directories(self) -> list[Path]:
        """Create auxiliary directories not handled by resource installation"""
        directories = [
            self.config.project_path / ".claude" / "logs",
            self.config.project_path / ".moai" / "project",
            self.config.project_path / ".moai" / "specs",
            self.config.project_path / ".moai" / "reports",
        ]

        created_dirs = []
        for directory in directories:
            try:
                if not self.security_manager.validate_path_safety(
                    directory, self.config.project_path
                ):
                    logger.error("Security validation failed for directory: %s", directory)
                    continue

                directory.mkdir(parents=True, exist_ok=True)
                created_dirs.append(directory)
                logger.debug("Created auxiliary directory: %s", directory)

            except Exception as e:
                logger.error("Failed to create directory %s: %s", directory, e)

        return created_dirs

    def _generate_next_steps(self) -> list[str]:
        """Generate next steps guidance for user"""
        next_steps = [
            "Project successfully initialized!",
            "",
            "Next steps:",
            "1. Navigate to your project:",
            f"   cd {self.config.project_path}",
            "",
            "2. Initialize project setup:",
            "   /moai:1-project",
            "",
            "3. Create your first feature:",
            '   /moai:2-spec <feature-name> "Feature description"',
            "",
            "4. Get help:",
            "   /moai:help",
        ]

        # Claude Code integration guidance
        if hasattr(self.config, "claude_version"):
            next_steps.extend([
                "",
                "Claude Code Integration:",
                "- All MoAI commands are installed as slash commands",
                "- Start with /moai:1-project to begin",
            ])

        return next_steps