"""
@FEATURE:COMMAND-001 Command availability checker for MoAI-ADK

Handles system command detection and package manager availability checks.
Extracted from system_manager.py for TRUST compliance (≤300 LOC).
"""

import subprocess
from typing import Any

from ..utils.logger import get_logger

logger = get_logger(__name__)


class CommandChecker:
    """@TASK:COMMAND-CHECKER-001 Checks command and tool availability across platforms."""

    def check_command_exists(self, command: str) -> bool:
        """
        Check if a command exists in the system.

        Args:
            command: Command name to check

        Returns:
            bool: True if command is available
        """
        try:
            subprocess.run(
                [command, "--version"], capture_output=True, text=True, check=True
            )
            logger.debug(f"Command '{command}' is available")
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            logger.debug(f"Command '{command}' is not available")
            return False

    def detect_python_command(self) -> str:
        """
        @TASK:PYTHON-DETECT-001 Detect available Python command for cross-platform hook execution

        Windows environments often don't have python3 command, causing hook execution failures.
        This method detects the best available Python command to ensure compatibility.

        Priority order: python3 > python > py

        Returns:
            str: Best available Python command (fallback to 'python' if none found)
        """
        python_commands = ["python3", "python", "py"]

        for cmd in python_commands:
            if self.check_command_exists(cmd):
                logger.info(f"Detected Python command: {cmd}")
                return cmd

        # Fallback to 'python' even if not found (most common case)
        logger.warning("No Python command detected, falling back to 'python'")
        return "python"

    def get_python_commands_info(self) -> dict[str, Any]:
        """
        Get comprehensive Python command availability information.

        Returns:
            dict: Python commands details including detected command
        """
        detected_command = self.detect_python_command()
        available_commands = []

        # Check which Python commands are available
        for cmd in ["python", "python3", "py"]:
            if self.check_command_exists(cmd):
                available_commands.append(cmd)

        logger.info(f"Python commands available: {available_commands}")
        logger.info(f"Detected best Python command: {detected_command}")

        return {
            "detected_command": detected_command,
            "available_commands": available_commands,
        }

    def get_package_managers_info(self) -> dict[str, bool]:
        """
        Get information about available package managers across platforms.

        Returns:
            dict: Package manager availability mapping
        """
        logger.info("Scanning for available package managers")

        package_managers = {
            "pip": self.check_command_exists("pip"),
            "pip3": self.check_command_exists("pip3"),
            "conda": self.check_command_exists("conda"),
            "brew": self.check_command_exists("brew"),  # macOS
            "apt": self.check_command_exists("apt"),  # Ubuntu/Debian
            "yum": self.check_command_exists("yum"),  # CentOS/RHEL
            "dnf": self.check_command_exists("dnf"),  # Fedora
            "choco": self.check_command_exists("choco"),  # Windows
            "winget": self.check_command_exists("winget"),  # Windows 10+
        }

        available_managers = [name for name, available in package_managers.items() if available]
        logger.info(f"Available package managers: {available_managers}")

        return package_managers

    def check_git_availability(self) -> bool:
        """
        Check if Git is available in the system.

        Returns:
            bool: True if Git is available
        """
        is_available = self.check_command_exists("git")

        if is_available:
            logger.info("Git is available")
        else:
            logger.warning("Git is not available")

        return is_available