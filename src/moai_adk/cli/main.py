# @CODE:PY314-001 | SPEC: SPEC-CLI-001/spec.md | TEST: tests/unit/test_cli_backup.py
"""CLI Main Module

CLI entry module:
- Re-exports the cli function from __main__.py
- Click-based CLI framework
- Rich console terminal output
"""
# type: ignore

from moai_adk.__main__ import cli, show_logo  # type: ignore[attr-defined]

__all__ = ["cli", "show_logo"]


# @CODE:USER-EXPERIENCE-001
