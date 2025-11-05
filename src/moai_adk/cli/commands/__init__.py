# @CODE:CLI-001 | SPEC: SPEC-CLI-001/spec.md | TEST: tests/unit/test_cli_commands.py
"""CLI command module

Core commands:
- init: initialize the project
- doctor: run system diagnostics
- status: show project status
- update: update templates to latest version
- backup: create project backups

Note: restore functionality is handled by checkpoint system in core.git.checkpoint
"""

from moai_adk.cli.commands.backup import backup
from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update

__all__ = ["init", "doctor", "status", "update", "backup"]
