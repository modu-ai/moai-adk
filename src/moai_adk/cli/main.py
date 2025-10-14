# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_commands.py
"""CLI Main Module

CLI 진입점 모듈:
- __main__.py의 cli 함수 재export
- Click 기반 CLI 프레임워크
- Rich console 터미널 출력
"""

from moai_adk.__main__ import cli, show_logo

__all__ = ["cli", "show_logo"]
