"""CLI Main Module

CLI entry module:
- Re-exports the cli function from __main__.py
- Click-based CLI framework
- Rich console terminal output
"""

# type: ignore

from moai_adk.__main__ import cli, show_logo  # type: ignore[attr-defined]

__all__ = ["cli", "show_logo"]
