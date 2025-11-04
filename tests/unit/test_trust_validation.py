"""
@TEST:HOOKS-003 - TRUST 원칙 자동 검증 시스템 단위 테스트
@CODE:HOOKS-003-HANDLER - PostToolUse Hook 핸들러 테스트

이 테스트는 PostToolUse Hook에서 TDD 완료를 감지하고
TRUST 5 원칙 자동 검증을 수행하는 기능을 테스트합니다.
"""

import json
import subprocess
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock
import pytest


# ==============================================================================
# Unit Tests for TDD Completion Detection
# ==============================================================================


class TestTddCompletionDetection:
    """TDD 완료 감지 로직 테스트"""

    @patch("subprocess.run")
    def test_detect_tdd_completion_with_green_commit(self, mock_run):
        """GREEN 커밋 감지 시 True 반환"""
        from moai_adk.core.validation import detect_tdd_completion

        # Given: GREEN 커밋이 있는 Git 로그
        mock_run.return_value = Mock(
            returncode=0,
            stdout="🟢 GREEN: Test coverage 85% achieved\n"
            "Some other commit\n"
            "Another commit",
        )

        # When: detect_tdd_completion() 호출
        result = detect_tdd_completion()

        # Then: True 반환
        assert result is True
        mock_run.assert_called_once()

    @patch("subprocess.run")
    def test_detect_tdd_completion_with_refactor_commit(self, mock_run):
        """REFACTOR 커밋 감지 시 True 반환"""
        from moai_adk.core.validation import detect_tdd_completion

        # Given: REFACTOR 커밋이 있는 Git 로그
        mock_run.return_value = Mock(
            returncode=0,
            stdout="♻️ REFACTOR: Optimize performance\n"
            "Some other commit",
        )

        # When: detect_tdd_completion() 호출
        result = detect_tdd_completion()

        # Then: True 반환
        assert result is True

    @patch("subprocess.run")
    def test_detect_tdd_completion_without_tdd_keywords(self, mock_run):
        """TDD 키워드 없으면 False 반환"""
        from moai_adk.core.validation import detect_tdd_completion

        # Given: TDD 키워드가 없는 Git 로그
        mock_run.return_value = Mock(
            returncode=0,
            stdout="docs: Update README\n" "chore: Update dependencies",
        )

        # When: detect_tdd_completion() 호출
        result = detect_tdd_completion()

        # Then: False 반환
        assert result is False

    @patch("subprocess.run")
    def test_detect_tdd_completion_git_command_fails(self, mock_run):
        """Git 명령 실패 시 False 반환"""
        from moai_adk.core.validation import detect_tdd_completion

        # Given: Git 명령 실패
        mock_run.return_value = Mock(returncode=1)

        # When: detect_tdd_completion() 호출
        result = detect_tdd_completion()

        # Then: False 반환
        assert result is False


# ==============================================================================
# Unit Tests for TRUST Validation Execution
# ==============================================================================


class TestTrustValidationExecution:
    """TRUST 검증 실행 로직 테스트"""

    @patch("subprocess.Popen")
    def test_trigger_trust_validation_creates_process(self, mock_popen):
        """비동기 검증 프로세스 생성 확인"""
        from moai_adk.core.validation import trigger_trust_validation

        # Given: subprocess.Popen 모킹
        mock_process = Mock(pid=12345)
        mock_popen.return_value = mock_process

        # When: trigger_trust_validation() 호출
        process = trigger_trust_validation()

        # Then: Popen 호출 확인 및 프로세스 반환
        assert process == mock_process
        assert process.pid == 12345
        mock_popen.assert_called_once()

    def test_collect_validation_result_success(self):
        """검증 성공 결과 수집"""
        from moai_adk.core.validation import collect_validation_result

        # Given: 성공한 검증 결과
        success_result = {
            "status": "passed",
            "test_coverage": 87.5,
            "code_constraints_passed": 25,
            "code_constraints_total": 25,
            "tag_integrity": True,
        }

        mock_process = Mock()
        mock_process.communicate.return_value = (
            json.dumps(success_result),
            "",
        )
        mock_process.returncode = 0

        # When: collect_validation_result() 호출
        result = collect_validation_result(mock_process)

        # Then: 결과 파싱 및 반환
        assert result["status"] == "passed"
        assert result["test_coverage"] == 87.5
        assert result["code_constraints_passed"] == 25

    def test_collect_validation_result_failure(self):
        """검증 실패 결과 수집"""
        from moai_adk.core.validation import collect_validation_result

        # Given: 검증 실패
        mock_process = Mock()
        mock_process.communicate.return_value = (
            "",
            "Test coverage 50% (require 85%)",
        )
        mock_process.returncode = 1

        # When: collect_validation_result() 호출
        result = collect_validation_result(mock_process)

        # Then: 실패 결과 반환
        assert result["status"] == "failed"
        assert "Test coverage" in result["error"]


# ==============================================================================
# Unit Tests for PostToolUse Handler Integration
# ==============================================================================


class TestPostToolUseHandler:
    """PostToolUse Hook 핸들러 테스트"""

    @patch("moai_adk.core.validation.detect_tdd_completion")
    @patch("moai_adk.core.validation.trigger_trust_validation")
    def test_handle_post_tool_use_triggers_validation(
        self, mock_trigger, mock_detect
    ):
        """TDD 완료 시 검증 트리거"""
        from moai_adk.core.validation import is_trust_validation_needed

        # Given: TDD 완료 상태
        mock_detect.return_value = True
        mock_trigger.return_value = Mock(pid=12345)

        payload = {
            "tool": "Bash",
            "input": {"command": "git commit -m '🟢 GREEN: Test'"},
        }

        # When: is_trust_validation_needed() 호출
        result = is_trust_validation_needed(payload)

        # Then: 검증 필요 확인
        assert result is True

    @patch("moai_adk.core.validation.detect_tdd_completion")
    def test_handle_post_tool_use_skips_non_tdd(self, mock_detect):
        """TDD 미완료 시 검증 스킵"""
        from moai_adk.core.validation import is_trust_validation_needed

        # Given: TDD 미완료 상태
        mock_detect.return_value = False

        payload = {
            "tool": "Read",
            "input": {"file_path": "/some/file.md"},
        }

        # When: is_trust_validation_needed() 호출
        result = is_trust_validation_needed(payload)

        # Then: 검증 불필요
        assert result is False

    def test_is_alfred_build_command_detection(self):
        """alfred:2-run 커맨드 감지"""
        from moai_adk.core.validation import is_alfred_build_command

        # Given: alfred:2-run 커맨드
        payload = {
            "tool": "SlashCommand",
            "input": {"command": "/alfred:2-run SPEC-001"},
        }

        # When: is_alfred_build_command() 호출
        result = is_alfred_build_command(payload)

        # Then: True 반환
        assert result is True


# ==============================================================================
# Unit Tests for Validation Result Formatting
# ==============================================================================


class TestValidationResultFormatting:
    """검증 결과 포맷팅 테스트"""

    def test_format_validation_success(self):
        """성공 결과 포맷팅"""
        from moai_adk.core.validation import format_validation_result

        result = {
            "status": "passed",
            "test_coverage": 87.5,
            "code_constraints_passed": 25,
            "code_constraints_total": 25,
        }

        formatted = format_validation_result(result)

        assert "✅" in formatted
        assert "TRUST 원칙 검증 통과" in formatted
        assert "87.5%" in formatted

    def test_format_validation_failure(self):
        """실패 결과 포맷팅"""
        from moai_adk.core.validation import format_validation_result

        result = {
            "status": "failed",
            "error": "Test coverage 50% (require 85%)",
            "test_coverage": 50,
            "recommendation": "Run pytest with --cov flag",
        }

        formatted = format_validation_result(result)

        assert "❌" in formatted
        assert "TRUST 원칙 검증 실패" in formatted
        assert "50%" in formatted


# ==============================================================================
# Performance Tests
# ==============================================================================


class TestPerformanceConstraints:
    """PostToolUse 100ms 제약 테스트"""

    @patch("subprocess.run")
    def test_tdd_detection_performance(self, mock_run):
        """TDD 감지 성능 (<10ms)"""
        import time

        from moai_adk.core.validation import detect_tdd_completion

        # Given: Git 호출 모킹
        mock_run.return_value = Mock(
            returncode=0,
            stdout="🟢 GREEN: Test"
        )

        # When: detect_tdd_completion() 실행
        start = time.time()
        result = detect_tdd_completion()
        elapsed = (time.time() - start) * 1000  # ms로 변환

        # Then: 100ms 이내 (실제 Git 호출은 subprocess로 1.0s 제약)
        assert elapsed < 100
        assert result is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
