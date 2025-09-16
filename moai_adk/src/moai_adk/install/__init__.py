"""
Installation and setup modules for MoAI-ADK.

This package contains installation and resource management:
- installer: Main installation logic
- resource_manager: Resource and template management
- installation_result: Installation result data structures
- post_install: Post-installation setup and hooks
"""

from .installer import SimplifiedInstaller
from .resource_manager import ResourceManager
from .installation_result import InstallationResult
from .post_install import main as post_install_main

# Backward compatibility alias
Installer = SimplifiedInstaller

__all__ = [
    'SimplifiedInstaller',
    'Installer',  # Backward compatibility
    'ResourceManager',
    'InstallationResult',
    'post_install_main'
]