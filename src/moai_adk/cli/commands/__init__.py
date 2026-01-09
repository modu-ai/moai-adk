"""CLI command module

Core commands:
- init: initialize the project
- doctor: run system diagnostics
- status: show project status
- update: update templates to latest version
- web: start the web backend server

Note: restore functionality is handled by checkpoint system in core.git.checkpoint
"""

from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update
from moai_adk.cli.commands.web import web

__all__ = ["init", "doctor", "status", "update", "web"]
