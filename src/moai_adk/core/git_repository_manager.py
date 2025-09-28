"""
@FEATURE:GIT-004 Git repository initialization management for MoAI-ADK

Handles Git repository creation, initialization, and gitignore management
with security validation and comprehensive error handling.
"""

import subprocess
from pathlib import Path

from ..utils.logger import get_logger
from .file_manager import FileManager
from .git_installation_manager import GitInstallationManager
from .security import SecurityManager

logger = get_logger(__name__)


class GitRepositoryManager:
    """@TASK:GIT-REPO-001 Manages Git repository initialization and setup."""

    def __init__(
        self,
        security_manager: SecurityManager = None,
        file_manager: FileManager = None,
    ):
        """
        Initialize Git repository manager.

        Args:
            security_manager: Security manager instance for validation
            file_manager: File manager for .gitignore creation
        """
        self.security_manager = security_manager or SecurityManager()
        self.file_manager = file_manager
        self.installation_manager = GitInstallationManager()

    def initialize_repository(self, project_path: Path) -> bool:
        """
        Initialize git repository if not already initialized.

        Args:
            project_path: Project root path

        Returns:
            bool: True if git repo exists or was successfully initialized
        """
        try:
            # Check if already a git repository
            git_dir = project_path / ".git"
            if git_dir.exists():
                logger.info("Git repository already exists - skipping initialization")
                return True

            # Check if git is available
            if not self.installation_manager.check_git_available():
                if self.installation_manager.offer_git_installation():
                    # Try again after installation
                    if not self.installation_manager.check_git_available():
                        logger.error(
                            "Git still not available after installation attempt"
                        )
                        return False
                else:
                    logger.info(
                        "Git installation declined - skipping git initialization"
                    )
                    return False

            # Initialize git repository
            if not self._initialize_repository(project_path):
                return False

            # Create initial .gitignore if it doesn't exist
            gitignore_path = project_path / ".gitignore"
            if not gitignore_path.exists() and self.file_manager:
                self.file_manager.create_gitignore(gitignore_path)

            logger.info("Git repository initialized successfully")
            return True

        except Exception as e:
            logger.error("Error initializing git repository: %s", e)
            return False

    def _initialize_repository(self, project_path: Path) -> bool:
        """Initialize git repository with security validation."""
        try:
            # Security validation
            if not self.security_manager.validate_subprocess_path(
                project_path, project_path
            ):
                logger.error(
                    "Security: Invalid path for git initialization: %s", project_path
                )
                return False

            result = subprocess.run(
                ["git", "init"], cwd=project_path, capture_output=True, text=True
            )

            if result.returncode == 0:
                logger.debug("Git init completed successfully")
                return True
            else:
                logger.error("Failed to initialize git: %s", result.stderr)
                return False

        except Exception as e:
            logger.error("Error during git initialization: %s", e)
            return False

    def create_gitignore(self, gitignore_path: Path) -> bool:
        """
        Create a comprehensive .gitignore file.

        Args:
            gitignore_path: Path where .gitignore should be created

        Returns:
            bool: True if successful
        """
        if self.file_manager:
            return self.file_manager.create_gitignore(gitignore_path)
        else:
            logger.warning("FileManager not available for gitignore creation")
            return False