"""
Installation and setup modules for MoAI-ADK.

This package contains installation and resource management:
- installer: Main installation logic
- resource_manager: Resource and template management
"""

from .installer import SimplifiedInstaller
from .resource_manager import ResourceManager

# Backward compatibility alias
Installer = SimplifiedInstaller

__all__ = [
    'SimplifiedInstaller',
    'Installer',  # Backward compatibility
    'ResourceManager'
]