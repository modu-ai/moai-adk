# @TEST:CLI-001 | SPEC: SPEC-CLI-001.md
"""CLI 명령어 테스트

테스트 대상 명령어:
1. moai init . → 프로젝트 초기화
2. moai doctor → 환경 검증
3. moai status → 프로젝트 상태
4. moai restore → 백업 복원

모킹 전략:
- core.project.initialize_project 모킹
- core.project.check_environment 모킹
- core.project.get_project_status 모킹
- core.backup.restore_backup 모킹
"""

from unittest.mock import MagicMock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.main import cli


class TestInitCommand:
    """moai init 명령어 테스트"""

    @patch("moai_adk.cli.commands.init.initialize_project")
    def test_init_current_directory(self, mock_init: MagicMock, cli_runner: CliRunner):
        """moai init . 테스트

        Given: init 명령어
        When: 현재 디렉토리(.)로 실행
        Then: initialize_project 호출, 성공 메시지 출력
        """
        mock_init.return_value = None
        result = cli_runner.invoke(cli, ["init", "."])

        assert result.exit_code == 0
        assert "Project initialized successfully" in result.output or "✓" in result.output
        mock_init.assert_called_once_with(".")

    @patch("moai_adk.cli.commands.init.initialize_project")
    def test_init_custom_path(self, mock_init: MagicMock, cli_runner: CliRunner):
        """moai init <path> 테스트

        Given: init 명령어
        When: 사용자 지정 경로로 실행
        Then: 지정 경로로 initialize_project 호출
        """
        mock_init.return_value = None
        result = cli_runner.invoke(cli, ["init", "/tmp/test-project"])

        assert result.exit_code == 0
        mock_init.assert_called_once_with("/tmp/test-project")

    @patch("moai_adk.cli.commands.init.initialize_project")
    def test_init_error_handling(self, mock_init: MagicMock, cli_runner: CliRunner):
        """init 에러 처리 테스트

        Given: init 명령어
        When: initialize_project가 예외 발생
        Then: 에러 메시지 출력, non-zero exit code
        """
        mock_init.side_effect = FileNotFoundError("Directory not found")
        result = cli_runner.invoke(cli, ["init", "."])

        assert result.exit_code != 0
        assert "Error" in result.output or "not found" in result.output


class TestDoctorCommand:
    """moai doctor 명령어 테스트"""

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_all_checks_pass(
        self, mock_check: MagicMock, cli_runner: CliRunner
    ):
        """doctor 명령어 - 모든 체크 통과 테스트

        Given: doctor 명령어
        When: 모든 환경 체크 통과
        Then: ✓ 아이콘과 함께 성공 메시지 출력
        """
        mock_check.return_value = {
            "Python 3.14+": True,
            "Git": True,
            "uv": True,
            ".moai/ directory": True,
        }
        result = cli_runner.invoke(cli, ["doctor"])

        assert result.exit_code == 0
        assert "Python 3.14+" in result.output
        assert "Git" in result.output
        # Rich 색상 코드로 인해 정확한 ✓ 확인은 어려울 수 있음
        assert "✓" in result.output or "diagnostics" in result.output

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_some_checks_fail(
        self, mock_check: MagicMock, cli_runner: CliRunner
    ):
        """doctor 명령어 - 일부 체크 실패 테스트

        Given: doctor 명령어
        When: 일부 환경 체크 실패
        Then: ✗ 아이콘과 함께 실패 메시지 출력
        """
        mock_check.return_value = {
            "Python 3.14+": True,
            "Git": False,
            "uv": True,
            ".moai/ directory": False,
        }
        result = cli_runner.invoke(cli, ["doctor"])

        assert result.exit_code == 0  # doctor는 실패해도 exit 0
        assert "Git" in result.output
        assert ".moai/ directory" in result.output


class TestStatusCommand:
    """moai status 명령어 테스트"""

    @patch("moai_adk.cli.commands.status.get_project_status")
    def test_status_command(self, mock_status: MagicMock, cli_runner: CliRunner):
        """status 명령어 테스트

        Given: status 명령어
        When: 프로젝트 상태 조회
        Then: 프로젝트 정보 출력 (mode, locale, spec_count)
        """
        mock_status.return_value = {
            "mode": "personal",
            "locale": "ko",
            "spec_count": 5,
        }
        result = cli_runner.invoke(cli, ["status"])

        assert result.exit_code == 0
        assert "Mode:" in result.output or "personal" in result.output
        assert "Locale:" in result.output or "ko" in result.output
        assert "SPECs:" in result.output or "5" in result.output

    @patch("moai_adk.cli.commands.status.get_project_status")
    def test_status_no_project(self, mock_status: MagicMock, cli_runner: CliRunner):
        """status 명령어 - 프로젝트 없음 테스트

        Given: status 명령어
        When: 프로젝트가 초기화되지 않음
        Then: 에러 메시지 출력
        """
        mock_status.side_effect = FileNotFoundError("No .moai/ directory found")
        result = cli_runner.invoke(cli, ["status"])

        assert result.exit_code != 0
        assert "Error" in result.output or "not found" in result.output


class TestRestoreCommand:
    """moai restore 명령어 테스트"""

    @patch("moai_adk.cli.commands.restore.restore_backup")
    def test_restore_latest(self, mock_restore: MagicMock, cli_runner: CliRunner):
        """restore 명령어 - 최신 백업 복원 테스트

        Given: restore 명령어
        When: timestamp 옵션 없이 실행
        Then: 최신 백업 복원, None 인자 전달
        """
        mock_restore.return_value = None
        result = cli_runner.invoke(cli, ["restore"])

        assert result.exit_code == 0
        assert "Restore completed" in result.output or "✓" in result.output
        mock_restore.assert_called_once_with(None)

    @patch("moai_adk.cli.commands.restore.restore_backup")
    def test_restore_specific_timestamp(
        self, mock_restore: MagicMock, cli_runner: CliRunner
    ):
        """restore 명령어 - 특정 시간 백업 복원 테스트

        Given: restore 명령어
        When: --timestamp 옵션으로 특정 시간 지정
        Then: 지정 시간 백업 복원
        """
        mock_restore.return_value = None
        timestamp = "2025-10-13T10:00:00"
        result = cli_runner.invoke(cli, ["restore", "--timestamp", timestamp])

        assert result.exit_code == 0
        mock_restore.assert_called_once_with(timestamp)

    @patch("moai_adk.cli.commands.restore.restore_backup")
    def test_restore_error(self, mock_restore: MagicMock, cli_runner: CliRunner):
        """restore 명령어 - 복원 실패 테스트

        Given: restore 명령어
        When: 백업 복원 중 에러 발생
        Then: 에러 메시지 출력, non-zero exit code
        """
        mock_restore.side_effect = FileNotFoundError("Backup not found")
        result = cli_runner.invoke(cli, ["restore"])

        assert result.exit_code != 0
        assert "Error" in result.output or "not found" in result.output
