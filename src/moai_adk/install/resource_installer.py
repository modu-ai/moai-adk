"""
@FEATURE:RESOURCE-INSTALLER-001 Specialized Resource Installation System

@TASK:RESOURCE-001 Handles installation of Claude Code, MoAI, and GitHub resources
with proper error handling and version tracking.
"""

from importlib import resources
from pathlib import Path

from .._version import __version__
from ..config import Config
from ..core.file_manager import FileManager
from ..core.resource_version import ResourceVersionManager
from ..core.security import SecurityManager
from ..core.system_manager import SystemManager
from ..utils.logger import get_logger
from .resource_manager import ResourceManager

logger = get_logger(__name__)


class ResourceInstaller:
    """
    @TASK:RESOURCE-MAIN-001 Specialized resource installation component

    Handles installation of all project resources including Claude Code
    configurations, MoAI templates, and GitHub workflows.

    Responsibilities:
    - Claude Code resource installation with Python command detection
    - MoAI resource installation with template context substitution
    - GitHub workflows installation
    - Project memory creation
    - Resource version tracking
    """

    def __init__(self, config: Config, security_manager: SecurityManager):
        """
        Initialize resource installer

        Args:
            config: Project configuration
            security_manager: Security validation manager
        """
        self.config = config
        self.security_manager = security_manager
        self.system_manager = SystemManager()
        self.resource_manager = ResourceManager()

        # Initialize FileManager with templates from ResourceManager
        with resources.as_file(self.resource_manager.templates_root) as templates_path:
            self.file_manager = FileManager(templates_path, security_manager)

        logger.info("ResourceInstaller initialized for: %s", config.project_path)

    def install_claude_resources(self) -> list[Path]:
        """
        @TASK:CLAUDE-INSTALL-001 Install Claude Code resources with Python command detection

        Automatically detects system Python command for Windows compatibility
        and generates dynamic settings.json configuration.

        Returns:
            List of created file paths

        Raises:
            RuntimeError: If no Claude resources were copied
        """
        try:
            # Detect available Python command for the system
            python_command = self.system_manager.detect_python_command()
            logger.info("Detected Python command for Claude hooks: %s", python_command)

            result = self.resource_manager.copy_claude_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite,
                python_command=python_command,
            )

            if not result:
                raise RuntimeError(
                    "No Claude resources were copied - installation may have failed"
                )

            # Create a list with Claude root directory to maintain API compatibility
            claude_dir = self.config.project_path / ".claude"
            result_list = [claude_dir] if claude_dir.exists() else []
            logger.info("Successfully installed Claude resources")
            return result_list

        except Exception as e:
            logger.error("Failed to install Claude resources: %s", e)
            raise  # Propagate exception to ensure installation failure is clear

    def install_moai_resources(self) -> list[Path]:
        """
        @TASK:MOAI-INSTALL-001 Install MoAI resources with template context substitution

        Installs MoAI system resources including configuration templates,
        memory systems, and project structure with proper context substitution.

        Returns:
            List of created file paths

        Raises:
            RuntimeError: If no MoAI resources were copied
        """
        try:
            # Determine template exclusion based on mode
            exclude_templates = False
            if (
                hasattr(self.config, "templates_mode")
                and str(self.config.templates_mode or "").lower() == "package"
            ):
                exclude_templates = True

            # Build comprehensive template context
            context = self._build_moai_context()

            result = self.resource_manager.copy_moai_resources_with_context(
                self.config.project_path,
                context,
                overwrite=self.config.force_overwrite,
                exclude_templates=exclude_templates,
            )

            if not result:
                raise RuntimeError(
                    "No MoAI resources were copied - installation may have failed"
                )

            # Create a list with MoAI root directory to maintain API compatibility
            moai_dir = self.config.project_path / ".moai"
            result_list = [moai_dir] if moai_dir.exists() else []
            logger.info("Successfully installed MoAI resources")
            return result_list

        except Exception as e:
            logger.error("Failed to install MoAI resources: %s", e)
            raise  # Propagate exception for proper error handling

    def install_github_workflows(self) -> list[Path]:
        """
        @TASK:GITHUB-INSTALL-001 Install GitHub workflows and CI/CD resources

        Installs GitHub Actions workflows and related CI/CD configurations
        for automated project management.

        Returns:
            List of created file paths (empty list if failed, no exception)
        """
        try:
            result = self.resource_manager.copy_github_resources(
                self.config.project_path, overwrite=self.config.force_overwrite
            )

            if result:
                # Create a list with GitHub workflows directory to maintain API compatibility
                github_dir = self.config.project_path / ".github" / "workflows"
                result_list = [github_dir] if github_dir.exists() else []
                logger.info("Successfully installed GitHub workflows")
                return result_list
            else:
                logger.warning("No GitHub workflows were installed")
                return []

        except Exception as e:
            logger.error("Failed to install GitHub workflows: %s", e)
            return []  # Return empty list instead of raising (GitHub is optional)

    def install_project_memory(self) -> list[Path]:
        """
        @TASK:MEMORY-INSTALL-001 Install project memory and documentation templates

        Creates project memory files including CLAUDE.md and memory templates
        with proper context substitution for project-specific information.

        Returns:
            List of created file paths
        """
        created_files = []

        try:
            # Build memory-specific context
            context = self._build_memory_context()

            # Install CLAUDE.md with template substitution
            claude_md_result = self.resource_manager.copy_project_memory_with_context(
                self.config.project_path,
                context,
                overwrite=self.config.force_overwrite,
            )

            if claude_md_result:
                created_files.append(self.config.project_path / "CLAUDE.md")

            # Install memory templates
            self.resource_manager.copy_memory_templates(
                self.config.project_path,
                self.config.tech_stack,
                context,
                overwrite=self.config.force_overwrite,
            )

            # Add memory directory to created files
            memory_dir = self.config.project_path / ".moai" / "memory"
            if memory_dir.exists():
                created_files.append(memory_dir)

            if created_files:
                logger.info("Successfully installed %d memory components", len(created_files))
            else:
                logger.warning("No project memory components were created")

            return created_files

        except Exception as e:
            logger.error("Failed to create project memory: %s", e)
            return []

    def write_version_info(self) -> None:
        """
        @TASK:VERSION-TRACKING-001 Record installed template/resource version metadata

        Writes version information for installed templates and resources
        to enable future upgrade and compatibility checks.
        """
        try:
            version_manager = ResourceVersionManager(self.config.project_path)
            template_version = self.resource_manager.get_version()
            version_manager.write(template_version, __version__)

            logger.info(
                "Recorded version info - Template: %s, Package: %s",
                template_version,
                __version__,
            )

        except Exception as exc:
            logger.warning("Failed to write resource version metadata: %s", exc)
            # Don't fail installation for version tracking issues

    def _build_moai_context(self) -> dict[str, str]:
        """
        Build comprehensive context for MoAI resource template substitution

        Returns:
            Dictionary containing all template substitution variables
        """
        # Get base template context from config
        base_context = self.config.get_template_context()

        # Build tech stack representation
        joined_stack = (
            ", ".join(self.config.tech_stack)
            if self.config.tech_stack
            else "미지정"
        )

        # Merge and transform context for MoAI templates
        moai_context = {
            **{k.upper(): str(v) for k, v in base_context.items()},
            "PROJECT_NAME": self.config.name,
            "TECH_STACK": joined_stack,
            "TECH_STACK_LIST": joined_stack,
        }

        return moai_context

    def _build_memory_context(self) -> dict[str, str]:
        """
        Build context specifically for memory template substitution

        Returns:
            Dictionary containing memory-specific template variables
        """
        # Get base template context from config
        base_context = self.config.get_template_context()

        # Build tech stack representation
        joined_stack = (
            ", ".join(self.config.tech_stack)
            if self.config.tech_stack
            else "미지정"
        )

        # Create memory-specific context
        memory_context = {
            **{k.upper(): str(v) for k, v in base_context.items()},
            "PROJECT_NAME": self.config.name,
            "TECH_STACK": joined_stack,
            "TECH_STACK_LIST": joined_stack,
        }

        return memory_context