# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli.py
"""MoAI-ADK CLI Entry Point

CLI 진입점 구현:
- Click 기반 CLI 프레임워크
- Rich console 터미널 출력
- ASCII 로고 출력
- --version, --help 옵션
"""

import sys

from moai_adk.cli.main import cli


def main() -> int:
    """CLI 진입점

    Returns:
        exit code (0: 성공, 1: 실패)
    """
    try:
        cli()
        return 0
    except Exception:
        # Exception은 이미 명령어에서 처리되었으므로
        # 여기서는 exit code만 반환
        return 1


if __name__ == "__main__":
    sys.exit(main())
