"""
@FEATURE:ENVIRONMENT-001 Environment and platform detection for MoAI-ADK

Handles OS, platform, and Python environment detection with cross-platform compatibility.
Extracted from system_manager.py for TRUST compliance (≤300 LOC).
"""

import platform
import sys
from typing import Any

from ..utils.logger import get_logger

logger = get_logger(__name__)


class EnvironmentDetector:
    """@TASK:ENV-DETECTOR-001 Detects environment and platform information."""

    def get_platform_info(self) -> dict[str, Any]:
        """
        Get comprehensive platform information.

        Returns:
            dict: Platform details including OS, architecture, processor
        """
        return {
            "system": platform.system(),
            "release": platform.release(),
            "machine": platform.machine(),
            "processor": platform.processor(),
        }

    def get_python_info(self) -> dict[str, Any]:
        """
        Get comprehensive Python environment information.

        Returns:
            dict: Python environment details including version and executable
        """
        return {
            "version": sys.version,
            "version_info": {
                "major": sys.version_info.major,
                "minor": sys.version_info.minor,
                "micro": sys.version_info.micro,
            },
            "executable": sys.executable,
        }

    def check_python_version(self, min_version: tuple = (3, 8)) -> bool:
        """
        Check if Python version meets minimum requirements.

        Args:
            min_version: Minimum required version as tuple (major, minor)

        Returns:
            bool: True if version is sufficient
        """
        current_version = (sys.version_info.major, sys.version_info.minor)
        is_compatible = current_version >= min_version

        if is_compatible:
            logger.info(f"Python version {current_version} meets requirements (>={min_version})")
        else:
            logger.warning(f"Python version {current_version} below requirements (>={min_version})")

        return is_compatible

    def get_system_info(self) -> dict[str, Any]:
        """
        Get comprehensive system information.

        Returns:
            dict: Complete system information including platform and Python
        """
        logger.info("Collecting system environment information")

        return {
            "platform": self.get_platform_info(),
            "python": self.get_python_info(),
        }