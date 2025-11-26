"""Version information for MoAI-ADK.

Provides version constants for template and MoAI framework.
"""

from importlib.metadata import PackageNotFoundError
from importlib.metadata import version as pkg_version

# MoAI Framework Version
try:
    MOAI_VERSION = pkg_version("moai-adk")
except PackageNotFoundError:
    MOAI_VERSION = "0.30.0"

# Template Schema Version
TEMPLATE_VERSION = "3.0.0"

__all__ = ["MOAI_VERSION", "TEMPLATE_VERSION"]
