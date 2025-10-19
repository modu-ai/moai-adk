# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_foundation.py
"""MoAI Agentic Development Kit

SPEC-First TDD Framework with Alfred SuperAgent
"""

from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("moai-adk")
except PackageNotFoundError:
    # Development mode fallback
    __version__ = "0.4.0-dev"

__all__ = ["__version__"]
