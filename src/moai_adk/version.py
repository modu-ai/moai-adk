"""Version information for MoAI-ADK.

Provides version constants for template and MoAI framework.
"""

from importlib.metadata import PackageNotFoundError
from importlib.metadata import version as pkg_version

# MoAI Framework Version
# Fallback version for development environment (only used when package not installed)
# This is automatically overwritten by importlib.metadata when package is installed via pip/uv
_FALLBACK_VERSION = "0.36.1"

try:
    MOAI_VERSION = pkg_version("moai-adk")
except PackageNotFoundError:
    MOAI_VERSION = _FALLBACK_VERSION

# Template Schema Version
TEMPLATE_VERSION = "3.0.0"

__all__ = ["MOAI_VERSION", "TEMPLATE_VERSION"]
