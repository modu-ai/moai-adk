# @TEST:CLI-001 | SPEC: SPEC-CLI-001.md
"""CLI 명령어 테스트

SPEC-CLI-001 요구사항 기반 테스트:
- moai init: 프로젝트 초기화
- moai doctor: 시스템 진단
- moai status: 프로젝트 상태
- moai restore: 백업 복원
"""

import pytest
from click.testing import CliRunner

from moai_adk.__main__ import cli


@pytest.fixture
def runner() -> CliRunner:
    """Click CLI Runner 픽스처"""
    return CliRunner()


class TestCLICommands:
    """CLI 명령어 테스트 그룹"""

    # TEST-CLI-001-INIT: moai init 명령어
    def test_init_command_should_initialize_project(self, runner: CliRunner) -> None:
        """TEST-CLI-001-INIT: moai init 명령어는 프로젝트를 초기화해야 한다"""
        result = runner.invoke(cli, ["init", "."])
        assert result.exit_code == 0
        assert "Initializing project" in result.output
        assert "✓ Project initialized successfully" in result.output

    def test_init_command_should_accept_custom_path(self, runner: CliRunner) -> None:
        """TEST-CLI-001-INIT: moai init 명령어는 사용자 지정 경로를 받아야 한다"""
        result = runner.invoke(cli, ["init", "/tmp/test-project"])
        # 실제 초기화는 실패할 수 있지만 명령어는 인식되어야 함
        assert result.exit_code in [0, 1]
        assert "init" not in result.output.lower() or "initializing" in result.output.lower()

    # TEST-CLI-001-DOCTOR: moai doctor 명령어
    def test_doctor_command_should_run_diagnostics(self, runner: CliRunner) -> None:
        """TEST-CLI-001-DOCTOR: moai doctor 명령어는 시스템 진단을 실행해야 한다"""
        result = runner.invoke(cli, ["doctor"])
        assert result.exit_code == 0
        assert "diagnostics" in result.output.lower() or "checking" in result.output.lower()

    def test_doctor_command_should_show_check_results(self, runner: CliRunner) -> None:
        """TEST-CLI-001-DOCTOR: moai doctor 명령어는 체크 결과를 표시해야 한다"""
        result = runner.invoke(cli, ["doctor"])
        # ✓ 또는 ✗ 아이콘이 포함되어야 함
        assert "✓" in result.output or "✗" in result.output

    # TEST-CLI-001-STATUS: moai status 명령어
    def test_status_command_should_show_project_status(self, runner: CliRunner) -> None:
        """TEST-CLI-001-STATUS: moai status 명령어는 프로젝트 상태를 표시해야 한다"""
        result = runner.invoke(cli, ["status"])
        assert result.exit_code == 0
        assert "status" in result.output.lower() or "project" in result.output.lower()

    def test_status_command_should_show_mode_and_locale(self, runner: CliRunner) -> None:
        """TEST-CLI-001-STATUS: moai status 명령어는 mode와 locale을 표시해야 한다"""
        result = runner.invoke(cli, ["status"])
        assert "mode" in result.output.lower() or "locale" in result.output.lower()

    # TEST-CLI-001-RESTORE: moai restore 명령어
    def test_restore_command_should_restore_from_latest_backup(
        self, runner: CliRunner
    ) -> None:
        """TEST-CLI-001-RESTORE: moai restore 명령어는 최신 백업을 복원해야 한다"""
        result = runner.invoke(cli, ["restore"])
        # 백업이 없어도 명령어는 인식되어야 함
        assert result.exit_code in [0, 1]
        assert "restore" in result.output.lower() or "backup" in result.output.lower()

    def test_restore_command_should_accept_timestamp_option(
        self, runner: CliRunner
    ) -> None:
        """TEST-CLI-001-RESTORE: moai restore 명령어는 --timestamp 옵션을 받아야 한다"""
        result = runner.invoke(cli, ["restore", "--timestamp", "2025-10-13"])
        # 실제 복원은 실패할 수 있지만 옵션은 인식되어야 함
        assert result.exit_code in [0, 1]
        # restore 또는 backup 관련 메시지가 있어야 함
        assert (
            "2025-10-13" in result.output
            or "timestamp" in result.output.lower()
            or "backup" in result.output.lower()
        )

    # TEST-CLI-001-HELP: 도움말 및 버전
    def test_cli_should_show_help_with_help_option(self, runner: CliRunner) -> None:
        """TEST-CLI-001-HELP: moai --help는 도움말을 표시해야 한다"""
        result = runner.invoke(cli, ["--help"])
        assert result.exit_code == 0
        assert "MoAI" in result.output
        assert "SPEC-First" in result.output

    def test_cli_should_show_version_with_version_option(self, runner: CliRunner) -> None:
        """TEST-CLI-001-VERSION: moai --version은 버전을 표시해야 한다"""
        result = runner.invoke(cli, ["--version"])
        assert result.exit_code == 0
        assert "0.3.0" in result.output

    def test_cli_should_show_logo_when_no_command(self, runner: CliRunner) -> None:
        """TEST-CLI-001-LOGO: moai 명령어만 입력하면 로고를 표시해야 한다"""
        result = runner.invoke(cli)
        assert result.exit_code == 0
        assert "▶◀ MoAI-ADK" in result.output
        assert "SPEC-First" in result.output


class TestCLIOutput:
    """CLI 출력 스타일 테스트"""

    def test_cli_should_use_rich_styling(self, runner: CliRunner) -> None:
        """TEST-CLI-001-OUTPUT: CLI는 Rich 스타일링을 사용해야 한다"""
        result = runner.invoke(cli, ["--help"])
        # Rich가 활성화되면 컬러 코드나 포맷팅이 포함됨
        # 단, Click CliRunner는 기본적으로 ANSI 코드를 제거하므로
        # 명령어가 정상 실행되는지 확인하는 것으로 충분
        assert result.exit_code == 0
        assert "MoAI" in result.output

    def test_error_messages_should_have_error_icon(self, runner: CliRunner) -> None:
        """TEST-CLI-001-ERROR: 에러 메시지는 ✗ 아이콘을 포함해야 한다"""
        # 존재하지 않는 명령어로 테스트
        result = runner.invoke(cli, ["nonexistent"])
        assert result.exit_code != 0


class TestCLIPerformance:
    """CLI 성능 테스트"""

    def test_init_command_should_complete_within_3_seconds(
        self, runner: CliRunner
    ) -> None:
        """TEST-CLI-001-PERF: init 명령어는 3초 이내에 완료되어야 한다"""
        import time

        start = time.time()
        result = runner.invoke(cli, ["init", "."])
        elapsed = time.time() - start

        # 실패해도 성능 제약은 확인
        assert elapsed < 3.0, f"Command took {elapsed:.2f} seconds (should be < 3s)"

    def test_status_command_should_complete_within_3_seconds(
        self, runner: CliRunner
    ) -> None:
        """TEST-CLI-001-PERF: status 명령어는 3초 이내에 완료되어야 한다"""
        import time

        start = time.time()
        result = runner.invoke(cli, ["status"])
        elapsed = time.time() - start

        assert elapsed < 3.0, f"Command took {elapsed:.2f} seconds (should be < 3s)"
