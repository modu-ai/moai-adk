"""
@FEATURE:INSTALLATION-VALIDATOR-001 Installation Verification and Validation System

@TASK:VALIDATION-001 Handles comprehensive validation of installation results
and verification of all components for proper MoAI-ADK operation.
"""

from pathlib import Path

from ..config import Config
from ..core.security import SecurityManager
from ..utils.logger import get_logger
from .resource_manager import ResourceManager

logger = get_logger(__name__)


class InstallationValidator:
    """
    @TASK:VALIDATOR-MAIN-001 Specialized installation validation component

    Handles comprehensive validation of installation results including
    resource verification, configuration validation, and system integrity checks.

    Responsibilities:
    - Installation completeness verification
    - Resource integrity validation
    - Configuration file validation
    - System integration checks
    - Error detection and reporting
    """

    def __init__(self, config: Config, security_manager: SecurityManager):
        """
        Initialize installation validator

        Args:
            config: Project configuration
            security_manager: Security validation manager
        """
        self.config = config
        self.security_manager = security_manager
        self.resource_manager = ResourceManager()

        logger.info("InstallationValidator initialized for: %s", config.project_path)

    def verify_installation(self) -> bool:
        """
        @TASK:VERIFY-MAIN-001 Perform comprehensive installation verification

        Validates all aspects of the installation including resource presence,
        configuration integrity, and system compatibility.

        Returns:
            True if installation is valid and complete, False otherwise
        """
        try:
            # Validate core project structure
            if not self._validate_project_structure():
                logger.error("Project structure validation failed")
                return False

            # Validate installed resources
            if not self._validate_installed_resources():
                logger.error("Resource validation failed")
                return False

            # Validate configuration files
            if not self._validate_configuration_files():
                logger.error("Configuration validation failed")
                return False

            # Validate security settings
            if not self._validate_security_settings():
                logger.error("Security validation failed")
                return False

            # Validate optional components if they were requested
            if not self._validate_optional_components():
                logger.error("Optional component validation failed")
                return False

            logger.info("Installation verification completed successfully")
            return True

        except Exception as e:
            logger.error("Installation verification failed with exception: %s", e)
            return False

    def _validate_project_structure(self) -> bool:
        """
        Validate basic project directory structure

        Returns:
            True if project structure is valid
        """
        try:
            required_dirs = [
                self.config.project_path,
                self.config.project_path / ".claude",
                self.config.project_path / ".moai",
            ]

            for directory in required_dirs:
                if not directory.exists():
                    logger.error("Required directory missing: %s", directory)
                    return False

                if not directory.is_dir():
                    logger.error("Path exists but is not a directory: %s", directory)
                    return False

                # Validate directory security
                if not self.security_manager.validate_path_safety(
                    directory, self.config.project_path
                ):
                    logger.error("Directory security validation failed: %s", directory)
                    return False

            logger.debug("Project structure validation passed")
            return True

        except Exception as e:
            logger.error("Project structure validation error: %s", e)
            return False

    def _validate_installed_resources(self) -> bool:
        """
        Validate that all required resources were installed correctly

        Returns:
            True if all resources are properly installed
        """
        try:
            # Use ResourceManager's validation method
            is_valid = self.resource_manager.validate_project_resources(
                self.config.project_path
            )

            if not is_valid:
                logger.error("Resource manager validation failed")
                return False

            # Additional resource-specific validations
            if not self._validate_claude_resources():
                return False

            if not self._validate_moai_resources():
                return False

            logger.debug("Resource validation passed")
            return True

        except Exception as e:
            logger.error("Resource validation error: %s", e)
            return False

    def _validate_claude_resources(self) -> bool:
        """Validate Claude Code specific resources"""
        try:
            claude_dir = self.config.project_path / ".claude"
            settings_file = claude_dir / "settings.json"

            # Validate settings.json if it exists
            if settings_file.exists() and (
                not settings_file.is_file() or settings_file.stat().st_size == 0
            ):
                logger.error("Invalid Claude settings.json")
                return False

            logger.debug("Claude resources validation passed")
            return True
        except Exception as e:
            logger.error("Claude resources validation error: %s", e)
            return False

    def _validate_moai_resources(self) -> bool:
        """Validate MoAI specific resources"""
        try:
            moai_dir = self.config.project_path / ".moai"
            required_dirs = ["memory", "indexes", "project", "specs"]

            # Create missing directories
            for dir_name in required_dirs:
                directory = moai_dir / dir_name
                if not directory.exists():
                    directory.mkdir(parents=True, exist_ok=True)
                    logger.info("Created missing MoAI directory: %s", directory)

            # Validate config.json if it exists
            config_file = moai_dir / "config.json"
            if config_file.exists() and (
                not config_file.is_file() or config_file.stat().st_size == 0
            ):
                logger.error("Invalid MoAI config.json")
                return False

            logger.debug("MoAI resources validation passed")
            return True
        except Exception as e:
            logger.error("MoAI resources validation error: %s", e)
            return False

    def _validate_configuration_files(self) -> bool:
        """
        Validate configuration files integrity

        Returns:
            True if configuration files are valid
        """
        try:
            # Validate Claude settings.json
            claude_settings = self.config.project_path / ".claude" / "settings.json"
            if claude_settings.exists():
                if not self._validate_json_file(claude_settings):
                    logger.error("Invalid Claude settings.json format")
                    return False

            # Validate MoAI config.json
            moai_config = self.config.project_path / ".moai" / "config.json"
            if moai_config.exists():
                if not self._validate_json_file(moai_config):
                    logger.error("Invalid MoAI config.json format")
                    return False

            logger.debug("Configuration files validation passed")
            return True

        except Exception as e:
            logger.error("Configuration files validation error: %s", e)
            return False

    def _validate_security_settings(self) -> bool:
        """Validate security-related settings and permissions"""
        try:
            # Validate project path security
            if not self.security_manager.validate_path_safety(
                self.config.project_path, self.config.project_path.parent
            ):
                logger.error("Project path security validation failed")
                return False

            logger.debug("Security settings validation passed")
            return True
        except Exception as e:
            logger.error("Security settings validation error: %s", e)
            return False

    def _validate_optional_components(self) -> bool:
        """Validate optional components if they were requested"""
        try:
            # Validate Git repository if it was requested
            if self.config.initialize_git and not (self.config.project_path / ".git").exists():
                logger.error("Git repository was requested but .git directory not found")
                return False

            logger.debug("Optional components validation passed")
            return True
        except Exception as e:
            logger.error("Optional components validation error: %s", e)
            return False

    def _validate_json_file(self, file_path: Path) -> bool:
        """
        Validate that a file contains valid JSON

        Args:
            file_path: Path to JSON file to validate

        Returns:
            True if file contains valid JSON
        """
        try:
            import json

            with open(file_path, "r", encoding="utf-8") as f:
                json.load(f)

            return True

        except json.JSONDecodeError as e:
            logger.error("Invalid JSON in file %s: %s", file_path, e)
            return False
        except Exception as e:
            logger.error("Error reading JSON file %s: %s", file_path, e)
            return False