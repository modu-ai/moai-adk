# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""CLI command module

Four core commands:
- init: initialize the project
- doctor: run system diagnostics
- status: show project status
- restore: restore backups
"""

from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.restore import restore
from moai_adk.cli.commands.status import status

__all__ = ["init", "doctor", "status", "restore"]
