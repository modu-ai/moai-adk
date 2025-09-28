"""
@FEATURE:INSTALLER-001 Modularized MoAI-ADK Project Installer

@TASK:INSTALL-REFACTOR-001 Refactored installer following TRUST principles with clear separation of concerns.
@TASK:BACKWARD-COMPATIBILITY-001 Maintains 100% backward compatibility while using modular architecture.

Legacy support: SimplifiedInstaller class delegates to InstallationOrchestrator for actual work.
"""

from collections.abc import Callable

from ..config import Config
from ..utils.logger import get_logger
from .installation_orchestrator import InstallationOrchestrator
from .installation_result import InstallationResult

logger = get_logger(__name__)


class SimplifiedInstaller:
    """
    @TASK:INSTALLER-LEGACY-001 Legacy installer maintaining backward compatibility

    This class now delegates all installation work to InstallationOrchestrator
    while maintaining the exact same public API for backward compatibility.

    Migration to modular architecture:
    - InstallationOrchestrator: Overall process coordination (≤300 LOC)
    - ResourceInstaller: Resource installation management (≤300 LOC)
    - ConfigurationGenerator: Configuration and Git setup (≤300 LOC)
    - InstallationValidator: Verification and validation (≤300 LOC)

    TRUST Principles Applied:
    - T: Single responsibility per module
    - R: Clear, readable delegation pattern
    - U: Unified architecture with specialized components
    - S: Security handled by specialized managers
    - T: Fully traceable installation process
    """

    def __init__(self, config: Config):
        """
        Initialize installation manager

        Args:
            config: Project configuration
        """
        self.config = config
        self.orchestrator = InstallationOrchestrator(config)

        logger.info("SimplifiedInstaller (legacy wrapper) initialized for: %s", config.project_path)

    def install(
        self, progress_callback: Callable[[str, int, int], None] | None = None
    ) -> InstallationResult:
        """
        Execute MoAI-ADK project installation

        This method now delegates to InstallationOrchestrator while maintaining
        the exact same public API for backward compatibility.

        Args:
            progress_callback: Progress callback function

        Returns:
            InstallationResult: Installation result
        """
        logger.info("Starting installation using modular architecture")
        return self.orchestrator.execute_installation(progress_callback)


# Backward compatibility alias
ProjectInstaller = SimplifiedInstaller