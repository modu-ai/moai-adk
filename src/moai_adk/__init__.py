"""
@FEATURE:MOAI-ADK-001 🗿 MoAI-ADK: Agentic Development Toolkit

@TASK:AGENTIC-TOOLKIT-001 A comprehensive toolkit for integrating agentic development workflows
with Claude Code, featuring intelligent project initialization, automated
hooks, and collaborative agent systems.
"""

from ._version import __version__
__author__ = "MoAI Team"
__email__ = "contact@moai-adk.dev"
__description__ = "Agentic Development Toolkit for Claude Code Integration"

# Core imports
from .config import Config
from .utils.logger import get_logger

# Subpackage imports with backward compatibility
from .install.installer import SimplifiedInstaller
from .core.security import SecurityManager
from .core.config_manager import ConfigManager
from .core.template_engine import TemplateEngine
from .cli import CLICommands

# Backward compatibility aliases
Installer = SimplifiedInstaller

__all__ = [
    "__version__",
    "SimplifiedInstaller",
    "Installer",  # Backward compatibility
    "Config",
    "get_logger",
    "SecurityManager",
    "ConfigManager",
    "TemplateEngine",
    "CLICommands",
]