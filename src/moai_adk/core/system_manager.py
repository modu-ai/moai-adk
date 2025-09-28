"""
@FEATURE:SYSTEM-001 System orchestration and management for MoAI-ADK

Orchestrates specialized system detection modules following TRUST principles.
Refactored from monolithic design to modular architecture (≤300 LOC).
"""

from typing import Any

from ..utils.logger import get_logger
from .command_checker import CommandChecker
from .environment_detector import EnvironmentDetector
from .nodejs_validator import NodejsValidator
from .project_detector import ProjectDetector

logger = get_logger(__name__)


class SystemManager:
    """@TASK:SYSTEM-MANAGER-001 Orchestrates system-level checks and validations."""

    def __init__(self):
        """Initialize system manager with specialized modules."""
        self.environment_detector = EnvironmentDetector()
        self.command_checker = CommandChecker()
        self.nodejs_validator = NodejsValidator()
        self.project_detector = ProjectDetector()

        logger.info("SystemManager initialized with specialized modules")

    def check_nodejs_and_npm(self) -> bool:
        """
        Check if Node.js and npm are installed, and verify ccusage can be used.

        Returns:
            bool: True if Node.js environment is properly set up
        """
        return self.nodejs_validator.check_nodejs_and_npm()

    def _check_command_exists(self, command: str) -> bool:
        """
        Check if a command exists in the system.

        @deprecated: Use command_checker.check_command_exists() directly
        """
        return self.command_checker.check_command_exists(command)

    def get_system_info(self) -> dict[str, Any]:
        """
        Get comprehensive system information.

        Returns:
            dict: System information including OS, Python, Node.js, etc.
        """
        logger.info("Collecting comprehensive system information")

        system_info = {
            "platform": self.environment_detector.get_platform_info(),
            "python": {
                **self.environment_detector.get_python_info(),
                **self.command_checker.get_python_commands_info(),
            },
            "nodejs": self.nodejs_validator.get_nodejs_info(),
            "git": {"available": self.command_checker.check_git_availability()},
            "package_managers": self.command_checker.get_package_managers_info(),
        }

        logger.info("System information collection completed")
        return system_info

    def check_python_version(self, min_version: tuple = (3, 8)) -> bool:
        """
        Check if Python version meets minimum requirements.

        Args:
            min_version: Minimum required version as tuple (major, minor)

        Returns:
            bool: True if version is sufficient
        """
        return self.environment_detector.check_python_version(min_version)

    def detect_python_command(self) -> str:
        """
        @TASK:PYTHON-DETECT-001 Detect available Python command for cross-platform hook execution

        Returns:
            str: Best available Python command (fallback to 'python' if none found)
        """
        return self.command_checker.detect_python_command()

    def get_python_info(self) -> dict[str, Any]:
        """
        @TASK:PYTHON-INFO-001 Get comprehensive Python environment information

        Returns:
            dict: Python environment details including available commands
        """
        return {
            **self.environment_detector.get_python_info(),
            **self.command_checker.get_python_commands_info(),
        }

    def detect_project_type(self, project_path) -> dict[str, Any]:
        """
        Detect project type based on existing files.

        Args:
            project_path: Path to project directory

        Returns:
            dict: Detected project information
        """
        return self.project_detector.detect_project_type(project_path)

    def should_create_package_json(self, config) -> bool:
        """
        Check if package.json should be created based on project configuration.

        Args:
            config: Project configuration

        Returns:
            bool: True if package.json should be created
        """
        return self.project_detector.should_create_package_json(config)
