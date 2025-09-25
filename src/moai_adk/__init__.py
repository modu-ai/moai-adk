"""
@FEATURE:MOAI-ADK-001 MoAI-ADK: Agentic Development Toolkit

@REQ:TOOLKIT-CORE-001 → @DESIGN:PACKAGE-ARCHITECTURE-001 → @TASK:AGENTIC-TOOLKIT-001 → @TEST:INTEGRATION-001

@TASK:AGENTIC-TOOLKIT-001 A comprehensive toolkit for integrating agentic development workflows
with Claude Code, featuring intelligent project initialization, automated
hooks, and collaborative agent systems.

@DESIGN:PACKAGE-ARCHITECTURE-001 Main package entry point with clean public API
"""

from ._version import __version__

__author__ = "MoAI Team"
__email__ = "contact@moai-adk.dev"
__description__ = "Modu-AI's Agentic Development Kit for streamlined development workflow"

# Core imports
from .cli import CLICommands
from .config import Config
from .core.config_manager import ConfigManager
from .core.security import SecurityManager
from .core.template_engine import TemplateEngine

# Subpackage imports with backward compatibility
from .install.installer import SimplifiedInstaller
from .utils.logger import get_logger

# Backward compatibility aliases
Installer = SimplifiedInstaller

__all__ = [
    "CLICommands",
    "Config",
    "ConfigManager",
    "Installer",  # Backward compatibility
    "SecurityManager",
    "SimplifiedInstaller",
    "TemplateEngine",
    "__version__",
    "get_logger",
]
