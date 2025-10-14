# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""CLI 명령어 모듈

4개 핵심 명령어:
- init: 프로젝트 초기화
- doctor: 시스템 진단
- status: 프로젝트 상태
- restore: 백업 복원
"""

from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.restore import restore
from moai_adk.cli.commands.status import status

__all__ = ["init", "doctor", "status", "restore"]
