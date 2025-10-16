# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_commands.py
"""CLI Main Module

CLI entry module:
- Re-exports the cli function from __main__.py
- Click-based CLI framework
- Rich console terminal output
"""

from moai_adk.__main__ import cli, show_logo

__all__ = ["cli", "show_logo"]
