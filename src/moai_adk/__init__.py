# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_foundation.py
"""MoAI Agentic Development Kit

SPEC-First TDD Framework with Alfred SuperAgent
"""

from importlib.metadata import PackageNotFoundError, version

try:
    __version__ = version("moai-adk")
except PackageNotFoundError:
    __version__ = "0.4.6-dev"

__all__ = ["__version__"]
