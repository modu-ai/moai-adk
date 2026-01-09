"""CLI command module

Core commands:
- init: initialize the project
- doctor: run system diagnostics
- status: show project status
- update: update templates to latest version

Note: restore functionality is handled by checkpoint system in core.git.checkpoint
"""

from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update

__all__ = ["init", "doctor", "status", "update"]
