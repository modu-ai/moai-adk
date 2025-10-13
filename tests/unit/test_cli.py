# @TEST:PY314-001 | SPEC: SPEC-PY314-001.md
"""Test PY314-001: CLI Entry Point

Phase 3 Tests: CLI 진입점 검증
- moai 명령어 실행 가능
- --version 옵션 출력
- --help 옵션 출력
- Rich console 색상 출력
"""

from click.testing import CliRunner
from moai_adk.__main__ import cli


class TestCLI:
    """CLI 진입점 테스트"""

    def test_cli_runs(self):
        """moai 명령어가 실행되어야 한다"""
        runner = CliRunner()
        result = runner.invoke(cli)
        assert result.exit_code == 0, f"CLI 실행 실패: {result.output}"

    def test_version_flag(self):
        """--version 플래그가 버전 정보를 출력해야 한다"""
        runner = CliRunner()
        result = runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "0.3.0" in result.output, "버전 정보가 없습니다"

    def test_help_flag(self):
        """--help 플래그가 도움말을 출력해야 한다"""
        runner = CliRunner()
        result = runner.invoke(cli, ["--help"])
        assert result.exit_code == 0
        assert "MoAI-ADK" in result.output or "Usage" in result.output, "도움말이 출력되지 않습니다"

    def test_cli_shows_logo(self):
        """CLI 실행 시 로고가 출력되어야 한다"""
        runner = CliRunner()
        result = runner.invoke(cli)
        # Rich console은 색상 코드를 포함하므로 기본 텍스트만 확인
        assert "MoAI" in result.output or "▶◀" in result.output, "로고가 출력되지 않습니다"
