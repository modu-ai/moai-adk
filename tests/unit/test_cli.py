# @TEST:CLI-001 | SPEC: SPEC-CLI-001.md
"""CLI 기본 동작 테스트

테스트 대상:
1. moai --version → "MoAI-ADK v0.3.0" 출력
2. moai --help → 도움말 출력
3. moai (인자 없음) → 로고 + Tip 출력
4. CLI 로딩 시간 < 500ms
"""

import time

import pytest
from click.testing import CliRunner

from moai_adk.__main__ import cli


class TestCLIBasics:
    """CLI 기본 동작 테스트"""

    def test_version_command(self, cli_runner: CliRunner):
        """moai --version 테스트

        Given: CLI 진입점
        When: --version 옵션 실행
        Then: "MoAI-ADK v0.3.0" 출력
        """
        result = cli_runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "0.3.0" in result.output

    def test_help_command(self, cli_runner: CliRunner):
        """moai --help 테스트

        Given: CLI 진입점
        When: --help 옵션 실행
        Then: 도움말 텍스트 출력
        """
        result = cli_runner.invoke(cli, ["--help"])
        assert result.exit_code == 0
        assert "MoAI Agentic Development Kit" in result.output
        assert "Options:" in result.output

    def test_no_arguments(self, cli_runner: CliRunner):
        """moai (인자 없음) 테스트

        Given: CLI 진입점
        When: 인자 없이 실행
        Then: 로고와 Tip 출력
        """
        result = cli_runner.invoke(cli, [])
        assert result.exit_code == 0
        assert "MoAI-ADK" in result.output
        assert "Tip:" in result.output

    def test_cli_loading_performance(self, cli_runner: CliRunner):
        """CLI 로딩 성능 테스트

        Given: CLI 진입점
        When: --version 실행하여 로딩 시간 측정
        Then: 500ms 이내 완료
        """
        start_time = time.time()
        result = cli_runner.invoke(cli, ["--version"])
        elapsed_time = (time.time() - start_time) * 1000  # ms 단위

        assert result.exit_code == 0
        assert elapsed_time < 500, f"CLI loading took {elapsed_time:.2f}ms (> 500ms)"

    def test_invalid_command(self, cli_runner: CliRunner):
        """존재하지 않는 명령어 테스트

        Given: CLI 진입점
        When: 잘못된 명령어 실행
        Then: 에러 메시지 출력 및 non-zero exit code
        """
        result = cli_runner.invoke(cli, ["invalid-command"])
        assert result.exit_code != 0
        assert "Error" in result.output or "No such command" in result.output
