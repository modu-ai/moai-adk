"""
@FEATURE:CONFIG-GENERATOR-001 Configuration and Git Setup System

@TASK:CONFIG-001 Handles generation of configuration files and Git repository
initialization with proper error handling and validation.
"""

from pathlib import Path

from ..config import Config
from ..core.config_manager import ConfigManager
from ..core.file_manager import FileManager
from ..core.git_manager import GitManager
from ..core.security import SecurityManager
from ..utils.logger import get_logger

logger = get_logger(__name__)


class ConfigurationGenerator:
    """
    @TASK:CONFIG-MAIN-001 Specialized configuration and Git setup component

    Handles generation of all configuration files including Claude Code settings,
    MoAI configuration, initial indexes, and Git repository initialization.

    Responsibilities:
    - Claude Code settings.json generation
    - MoAI config.json creation
    - Initial index creation (SQLite database)
    - Git repository initialization
    - .gitignore file creation
    """

    def __init__(self, config: Config, security_manager: SecurityManager):
        """
        Initialize configuration generator

        Args:
            config: Project configuration
            security_manager: Security validation manager
        """
        self.config = config
        self.security_manager = security_manager
        self.config_manager = ConfigManager()

        # Initialize Git manager with temporary file manager
        # We'll create a minimal file manager just for Git operations
        self.git_manager = GitManager(
            project_dir=config.project_path,
            security_manager=security_manager,
            file_manager=None,  # Git manager can work without file manager for basic ops
        )

        logger.info("ConfigurationGenerator initialized for: %s", config.project_path)

    def create_configuration_files(self) -> list[Path]:
        """
        @TASK:CONFIG-FILES-001 Create all project configuration files

        Generates Claude Code settings.json and MoAI config.json with
        project-specific configurations and security settings.

        Returns:
            List of created configuration file paths
        """
        config_files = []

        try:
            # Create Claude Code settings
            claude_settings = self._create_claude_settings()
            if claude_settings:
                config_files.append(claude_settings)

            # Create MoAI configuration
            moai_config = self._create_moai_config()
            if moai_config:
                config_files.append(moai_config)

            logger.info("Created %d configuration files", len(config_files))
            return config_files

        except Exception as e:
            logger.error("Failed to create configuration files: %s", e)
            return []

    def create_initial_indexes(self) -> list[Path]:
        """
        @TASK:INDEX-CREATION-001 Create initial project indexes and databases

        Creates SQLite database and other index files required for MoAI system
        operation including TAG tracking and project metadata.

        Returns:
            List of created index file paths

        Raises:
            Exception: If critical index creation fails
        """
        try:
            index_files = self.config_manager.create_initial_indexes(
                self.config.project_path, self.config
            )

            logger.info("Successfully created %d index files", len(index_files))
            return index_files

        except Exception as e:
            error_msg = f"Failed to create initial indexes: {e}"
            logger.error(error_msg)
            # Index creation is critical for MoAI system functionality
            raise Exception(error_msg) from e

    def initialize_git_repository(self) -> list[Path]:
        """
        @TASK:GIT-INIT-001 Initialize Git repository with proper configuration

        Creates Git repository, generates .gitignore file, and sets up
        initial repository structure for project version control.

        Returns:
            List of created Git-related file paths
        """
        git_files = []

        try:
            # Initialize Git repository
            if self.git_manager.initialize_repository(self.config.project_path):
                git_files.append(self.config.project_path / ".git")
                logger.info("Successfully initialized Git repository")

                # Create .gitignore file
                gitignore_path = self._create_gitignore()
                if gitignore_path:
                    git_files.append(gitignore_path)

            logger.info("Initialized Git repository with %d components", len(git_files))
            return git_files

        except Exception as e:
            logger.error("Failed to initialize Git repository: %s", e)
            return []  # Git initialization is optional, don't fail the installation

    def _create_claude_settings(self) -> Path | None:
        """
        Create Claude Code settings.json configuration

        Returns:
            Path to created settings file or None if failed
        """
        try:
            claude_settings_path = self.config.project_path / ".claude" / "settings.json"

            # Ensure .claude directory exists
            claude_settings_path.parent.mkdir(parents=True, exist_ok=True)

            # Validate path security
            if not self.security_manager.validate_path_safety(
                claude_settings_path, self.config.project_path
            ):
                logger.error("Security validation failed for Claude settings path")
                return None

            success = self.config_manager.create_claude_settings(
                claude_settings_path, self.config
            )

            if success:
                logger.info("Created Claude Code settings: %s", claude_settings_path)
                return claude_settings_path
            else:
                logger.warning("Failed to create Claude Code settings")
                return None

        except Exception as e:
            logger.error("Error creating Claude settings: %s", e)
            return None

    def _create_moai_config(self) -> Path | None:
        """
        Create MoAI config.json configuration

        Returns:
            Path to created config file or None if failed
        """
        try:
            moai_config_path = self.config.project_path / ".moai" / "config.json"

            # Ensure .moai directory exists
            moai_config_path.parent.mkdir(parents=True, exist_ok=True)

            # Validate path security
            if not self.security_manager.validate_path_safety(
                moai_config_path, self.config.project_path
            ):
                logger.error("Security validation failed for MoAI config path")
                return None

            success = self.config_manager.create_moai_config(moai_config_path, self.config)

            if success:
                logger.info("Created MoAI configuration: %s", moai_config_path)
                return moai_config_path
            else:
                logger.warning("Failed to create MoAI configuration")
                return None

        except Exception as e:
            logger.error("Error creating MoAI config: %s", e)
            return None

    def _create_gitignore(self) -> Path | None:
        """
        Create .gitignore file with appropriate patterns

        Returns:
            Path to created .gitignore file or None if failed
        """
        try:
            gitignore_path = self.config.project_path / ".gitignore"

            # Validate path security
            if not self.security_manager.validate_path_safety(
                gitignore_path, self.config.project_path
            ):
                logger.error("Security validation failed for .gitignore path")
                return None

            success = self.git_manager.create_gitignore(gitignore_path)

            if success:
                logger.info("Created .gitignore file: %s", gitignore_path)
                return gitignore_path
            else:
                logger.warning("Failed to create .gitignore file")
                return None

        except Exception as e:
            logger.error("Error creating .gitignore: %s", e)
            return None