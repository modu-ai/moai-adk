"""
@FEATURE:INSTALL-001 Installation and setup modules for MoAI-ADK

@REQ:INSTALL-SYSTEM-001 → @DESIGN:INSTALL-ARCHITECTURE-001 → @TASK:INSTALL-INTEGRATION-001 → @TEST:INSTALL-PROCESS-001

@DESIGN:INSTALL-ARCHITECTURE-001 Modular installation system with clean separation
@TASK:INSTALL-INTEGRATION-001 Installation process orchestration

This package contains installation and resource management:
- @TASK:INSTALLER-MAIN-001 Main installation logic
- @TASK:RESOURCE-MGMT-001 Resource and template management
- @TASK:INSTALL-RESULT-001 Installation result data structures
- @TASK:POST-INSTALL-001 Post-installation setup and hooks
"""

from .installation_result import InstallationResult
from .installer import SimplifiedInstaller
from .post_install import main as post_install_main
from .resource_manager import ResourceManager

# Backward compatibility alias
Installer = SimplifiedInstaller

__all__ = [
    "InstallationResult",
    "Installer",  # Backward compatibility
    "ResourceManager",
    "SimplifiedInstaller",
    "post_install_main",
]
