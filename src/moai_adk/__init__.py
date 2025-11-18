"""MoAI Agentic Development Kit

SPEC-First TDD Framework with Alfred SuperAgent
"""

from importlib.metadata import PackageNotFoundError, version

try:
    __version__ = version("moai-adk")
except PackageNotFoundError:
    __version__ = "0.25.11"

__all__ = ["__version__"]
