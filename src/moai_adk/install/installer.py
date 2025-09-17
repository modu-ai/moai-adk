"""
Simplified MoAI-ADK Project Installer

Simplified installation system using embedded package resources instead of symbolic links.
Ensures perfect compatibility with Claude Code by directly copying all resources through ResourceManager.
"""

from pathlib import Path
from typing import List, Optional
from collections.abc import Callable

from ..utils.logger import get_logger
from ..config import Config
from .installation_result import InstallationResult
from ..utils.progress_tracker import ProgressTracker
from ..core.security import SecurityManager
from ..core.directory_manager import DirectoryManager
from ..core.config_manager import ConfigManager
from ..core.git_manager import GitManager
from .resource_manager import ResourceManager
from ..core.resource_version import ResourceVersionManager
from .._version import __version__

logger = get_logger(__name__)


class SimplifiedInstaller:
    """
    Simplified MoAI-ADK project installation manager

    Installation system that directly copies embedded package resources
    instead of symbolic links for stable operation across all platforms.
    """

    def __init__(self, config: Config):
        """
        Initialize installation manager

        Args:
            config: Project configuration
        """
        self.config = config
        self.progress = ProgressTracker()

        # Initialize core managers
        self.security_manager = SecurityManager()
        self.directory_manager = DirectoryManager(self.security_manager)
        self.config_manager = ConfigManager()
        self.git_manager = GitManager()
        self.resource_manager = ResourceManager()

        logger.info("SimplifiedInstaller initialized for: %s", config.project_path)

    def install(self, progress_callback: Optional[Callable[[str, int, int], None]] = None) -> InstallationResult:
        """
        Execute MoAI-ADK project installation

        Args:
            progress_callback: Progress callback function

        Returns:
            InstallationResult: Installation result
        """
        files_created: List[str] = []
        errors: List[str] = []

        try:
            # Step 1: Creating project directory
            self.progress.update_progress("Creating project directory...", progress_callback)
            self.directory_manager.create_project_directory(self.config)
            files_created.append(str(self.config.project_path))

            # Step 2: Installing Claude Code resources (creates .claude/ with content)
            self.progress.update_progress("Installing Claude Code resources...", progress_callback)
            claude_files = self._install_claude_resources()
            files_created.extend([str(f) for f in claude_files])

            # Step 3: Installing MoAI resources (creates .moai/ with content)
            self.progress.update_progress("Installing MoAI resources...", progress_callback)
            moai_files = self._install_moai_resources()
            files_created.extend([str(f) for f in moai_files])

            # Step 3b: Record installed template version
            self._write_resource_version_info()

            # Step 4: Creating auxiliary directories (logs, empty directories)
            self.progress.update_progress("Setting up auxiliary directories...", progress_callback)
            directories = self._create_basic_structure()
            files_created.extend([str(d) for d in directories])

            # Step 5: Setting up GitHub workflows (optional)
            if self.config.include_github:
                self.progress.update_progress("Setting up GitHub workflows...", progress_callback)
                github_files = self._install_github_workflows()
                files_created.extend([str(f) for f in github_files])

            # Step 6: Creating project memory
            self.progress.update_progress("Creating project memory...", progress_callback)
            if self._install_project_memory():
                files_created.append(str(self.config.project_path / 'CLAUDE.md'))

            # Step 7: Generating configuration files
            self.progress.update_progress("Generating configuration files...", progress_callback)
            config_files = self._create_configuration_files()
            files_created.extend([str(f) for f in config_files])

            # Step 8: Initializing Git repository (optional)
            if self.config.initialize_git:
                self.progress.update_progress("Initializing Git repository...", progress_callback)
                git_files = self._initialize_git_repository()
                files_created.extend([str(f) for f in git_files])

            # Step 9: Verifying installation
            self.progress.update_progress("Verifying installation...", progress_callback)
            if not self._verify_installation():
                errors.append("Installation verification failed")

            self.progress.update_progress("Installation complete!", progress_callback)

            return InstallationResult(
                success=len(errors) == 0,
                project_path=str(self.config.project_path),
                files_created=files_created,
                errors=errors,
                next_steps=self._generate_next_steps(),
                config=self.config
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
                config=self.config
            )

    def _create_basic_structure(self) -> List[Path]:
        """Create only auxiliary directories that won't be populated by ResourceManager"""
        directories = [
            # Only create logs directory - ResourceManager will handle .claude/ with content
            self.config.project_path / ".claude" / "logs",

            # Only create empty directories that ResourceManager doesn't populate
            self.config.project_path / ".moai" / "steering",
            self.config.project_path / ".moai" / "specs",
            self.config.project_path / ".moai" / "reports",
        ]

        # Note: .claude/ and .moai/ will be created by ResourceManager with content
        # Note: .moai/memory/ and .moai/indexes/ will be created by ResourceManager

        created_dirs = []
        for directory in directories:
            try:
                if not self.security_manager.validate_path_safety(directory, self.config.project_path):
                    logger.error("Security validation failed for directory: %s", directory)
                    continue

                directory.mkdir(parents=True, exist_ok=True)
                created_dirs.append(directory)
                logger.debug("Created directory: %s", directory)

            except Exception as e:
                logger.error("Failed to create directory %s: %s", directory, e)

        return created_dirs

    def _install_claude_resources(self) -> List[Path]:
        """Claude Code 리소스 설치"""
        try:
            return self.resource_manager.copy_claude_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to install Claude resources: %s", e)
            return []

    def _install_moai_resources(self) -> List[Path]:
        """MoAI 리소스 설치"""
        try:
            # templates_mode: 'copy' | 'package'
            exclude_templates = False
            if hasattr(self.config, 'templates_mode') and str(getattr(self.config, 'templates_mode') or '').lower() == 'package':
                exclude_templates = True

            return self.resource_manager.copy_moai_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite,
                exclude_templates=exclude_templates,
            )
        except Exception as e:
            logger.error("Failed to install MoAI resources: %s", e)
            return []

    def _install_github_workflows(self) -> List[Path]:
        """GitHub 워크플로우 설치"""
        try:
            return self.resource_manager.copy_github_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to install GitHub workflows: %s", e)
            return []

    def _install_project_memory(self) -> bool:
        """프로젝트 메모리 파일 생성"""
        try:
            result = self.resource_manager.copy_project_memory(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
            context = self.config.get_template_context()
            joined_stack = ", ".join(self.config.tech_stack) if self.config.tech_stack else "미지정"
            memory_context = {
                **{k.upper(): str(v) for k, v in context.items()},
                "PROJECT_NAME": self.config.name,
                "TECH_STACK": joined_stack,
                "TECH_STACK_LIST": joined_stack,
            }

            self.resource_manager.copy_memory_templates(
                self.config.project_path,
                self.config.tech_stack,
                memory_context,
                overwrite=self.config.force_overwrite,
            )

            return result
        except Exception as e:
            logger.error("Failed to create project memory: %s", e)
            return False

    def _write_resource_version_info(self) -> None:
        """Record the installed template/resource version metadata."""
        try:
            version_manager = ResourceVersionManager(self.config.project_path)
            template_version = self.resource_manager.get_version()
            version_manager.write(template_version, __version__)
        except Exception as exc:
            logger.warning("Failed to write resource version metadata: %s", exc)

    def _create_configuration_files(self) -> List[Path]:
        """설정 파일 생성"""
        config_files = []

        try:
            # Claude Code 설정
            claude_settings = self.config.project_path / ".claude" / "settings.json"
            if self.config_manager.create_claude_settings(claude_settings, self.config):
                config_files.append(claude_settings)

            # MoAI 설정
            moai_config = self.config.project_path / ".moai" / "config.json"
            if self.config_manager.create_moai_config(moai_config, self.config):
                config_files.append(moai_config)

            logger.info("Created %d configuration files", len(config_files))
            return config_files

        except Exception as e:
            logger.error("Failed to create configuration files: %s", e)
            return []

    def _initialize_git_repository(self) -> List[Path]:
        """Git 저장소 초기화"""
        git_files = []

        try:
            if self.git_manager.initialize_repository(self.config.project_path):
                git_files.append(self.config.project_path / ".git")

                # .gitignore 생성
                gitignore_path = self.config.project_path / ".gitignore"
                if self.git_manager.create_gitignore(gitignore_path):
                    git_files.append(gitignore_path)

            logger.info("Initialized Git repository with %d files", len(git_files))
            return git_files

        except Exception as e:
            logger.error("Failed to initialize Git repository: %s", e)
            return []

    def _verify_installation(self) -> bool:
        """설치 검증"""
        try:
            return self.resource_manager.validate_project_resources(self.config.project_path)
        except Exception as e:
            logger.error("Installation verification failed: %s", e)
            return False

    def _generate_next_steps(self) -> List[str]:
        """Generate next steps guidance"""
        next_steps = [
            "Project successfully initialized!",
            "",
            "Next steps:",
            "1. Navigate to your project:",
            f"   cd {self.config.project_path}",
            "",
            "2. Initialize project setup:",
            "   /moai:1-project init",
            "",
            "3. Create your first feature:",
            "   /moai:2-spec <feature-name> \"Feature description\"",
            "",
            "4. Get help:",
            "   /moai:help",
        ]

        # Claude Code integration guidance
        if hasattr(self.config, 'claude_version'):
            next_steps.extend([
                "",
                "Claude Code Integration:",
                "- All MoAI commands are installed as slash commands",
                "- Start with /moai:1-project to begin",
            ])

        return next_steps


# 하위 호환성을 위한 별칭
ProjectInstaller = SimplifiedInstaller
