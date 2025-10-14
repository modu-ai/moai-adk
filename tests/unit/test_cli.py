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


class TestCLICommands:
    """CLI 명령어 테스트 (init, doctor, status, restore)"""

    def test_init_command_with_path(self, cli_runner: CliRunner):
        """moai init . 테스트

        Given: CLI 진입점
        When: init 명령어 실행
        Then: 프로젝트 초기화 메시지 출력
        """
        result = cli_runner.invoke(cli, ["init", "."])
        assert result.exit_code == 0
        assert "Initializing" in result.output
        assert "successfully" in result.output or "completed" in result.output

    def test_init_command_default_path(self, cli_runner: CliRunner):
        """moai init (경로 없음) 테스트

        Given: CLI 진입점
        When: init 명령어만 실행 (기본 경로 '.')
        Then: 프로젝트 초기화 메시지 출력
        """
        result = cli_runner.invoke(cli, ["init"])
        assert result.exit_code == 0
        assert "Initializing" in result.output

    def test_doctor_command(self, cli_runner: CliRunner):
        """moai doctor 테스트

        Given: CLI 진입점
        When: doctor 명령어 실행
        Then: 시스템 진단 결과 출력
        """
        result = cli_runner.invoke(cli, ["doctor"])
        assert result.exit_code == 0
        assert "diagnostics" in result.output or "Running" in result.output
        # 최소 하나의 체크 마크가 있어야 함
        assert "✓" in result.output or "✗" in result.output

    def test_status_command(self, cli_runner: CliRunner):
        """moai status 테스트

        Given: CLI 진입점
        When: status 명령어 실행
        Then: 프로젝트 상태 정보 출력
        """
        result = cli_runner.invoke(cli, ["status"])
        assert result.exit_code == 0
        assert "Project Status" in result.output or "Status" in result.output
        # 최소한 mode, locale 정보 중 하나는 있어야 함 (새로운 테이블 형식은 콜론 없음)
        assert "Mode" in result.output or "Locale" in result.output

    def test_restore_command_default(self, cli_runner: CliRunner):
        """moai restore (옵션 없음) 테스트

        Given: CLI 진입점
        When: restore 명령어 실행 (기본: 최신 백업)
        Then: 백업 복원 메시지 출력
        """
        result = cli_runner.invoke(cli, ["restore"])
        # 백업이 없어도 명령어는 인식되어야 함
        assert result.exit_code in [0, 1]
        assert "restor" in result.output.lower() or "backup" in result.output.lower()

    def test_restore_command_with_timestamp(self, cli_runner: CliRunner):
        """moai restore --timestamp 테스트

        Given: CLI 진입점
        When: restore 명령어 + timestamp 옵션 실행
        Then: 특정 시점 백업 복원 메시지 출력
        """
        timestamp = "2025-10-14-120000"
        result = cli_runner.invoke(cli, ["restore", "--timestamp", timestamp])
        # 백업이 없어도 옵션은 인식되어야 함
        assert result.exit_code in [0, 1]
        assert (
            "restor" in result.output.lower()
            or timestamp in result.output
            or "backup" in result.output.lower()
        )
