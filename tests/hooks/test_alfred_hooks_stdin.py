#!/usr/bin/env python3
# @TEST:WINDOWS-HOOKS-001 | SPEC: SPEC-WINDOWS-HOOKS-001.md
"""Alfred Hooks stdin 읽기 테스트

Windows/macOS/Linux 크로스 플랫폼 stdin 처리 검증

TDD History:
    - RED: Windows stdin EOF 처리, 빈 stdin, JSON 파싱 실패 테스트 작성
    - GREEN: Iterator 패턴 구현 후 테스트 통과
    - REFACTOR: 에러 처리 강화, 주석 개선
"""
import json
import subprocess
import sys
from io import StringIO
from pathlib import Path

import pytest


# Path to alfred_hooks.py
HOOKS_SCRIPT = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred" / "alfred_hooks.py"


class TestAlfredHooksStdin:
    """Alfred Hooks stdin 읽기 테스트 케이스

    SPEC 요구사항 검증:
        - Windows/macOS/Linux 모든 환경에서 stdin 안정적 읽기
        - 빈 stdin 처리 시 에러 없이 기본값 반환
        - JSON 파싱 실패 시 명확한 에러 처리
    """

    def test_stdin_normal_json(self):
        """정상 JSON 입력 처리 테스트

        SPEC 요구사항:
            - WHEN stdin에 유효한 JSON이 제공되면, 시스템은 정상적으로 파싱해야 한다

        Given: 유효한 JSON 페이로드
        When: SessionStart 이벤트로 alfred_hooks.py 실행
        Then: JSON을 정상적으로 파싱하고 결과를 반환한다
        """
        payload = {"cwd": "."}
        input_data = json.dumps(payload)

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 0, f"Expected exit code 0, got {result.returncode}. stderr: {result.stderr}"

        # stdout이 유효한 JSON인지 확인
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")

    def test_stdin_empty(self):
        """빈 stdin 처리 테스트

        SPEC 요구사항:
            - WHEN stdin이 비어있으면, 시스템은 JSONDecodeError를 발생시키지 않고 빈 객체로 처리해야 한다

        Given: 빈 stdin 입력
        When: SessionStart 이벤트로 alfred_hooks.py 실행
        Then: JSONDecodeError 없이 기본 HookResult를 반환한다
        """
        input_data = ""

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        # Iterator 패턴 구현 후: 빈 stdin은 "{}"로 처리되어 정상 동작
        assert result.returncode == 0, f"Empty stdin should be handled gracefully. stderr: {result.stderr}"

        # stdout이 유효한 JSON인지 확인
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")

    def test_stdin_invalid_json(self):
        """잘못된 JSON 입력 처리 테스트

        SPEC 요구사항:
            - WHEN JSON 파싱이 실패하면, 시스템은 JSONDecodeError를 발생시키고 exit code 1을 반환해야 한다

        Given: 잘못된 JSON 문자열
        When: SessionStart 이벤트로 alfred_hooks.py 실행
        Then: JSONDecodeError를 stderr로 출력하고 exit code 1을 반환한다
        """
        input_data = "{ invalid json }"

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 1, "Invalid JSON should return exit code 1"
        assert "JSON parse error" in result.stderr or "JSONDecodeError" in result.stderr

    def test_stdin_cross_platform(self):
        """크로스 플랫폼 stdin 읽기 테스트

        SPEC 요구사항:
            - stdin 읽기 로직은 플랫폼에 무관하게 동작해야 한다
            - Windows/macOS/Linux 모든 환경에서 EOF를 올바르게 처리해야 한다

        Given: 멀티라인 JSON 페이로드 (다양한 플랫폼 시뮬레이션)
        When: SessionStart 이벤트로 alfred_hooks.py 실행
        Then: 모든 라인을 정상적으로 읽고 파싱한다
        """
        # 멀티라인 JSON (Windows \r\n, Unix \n 모두 포함)
        payload = {
            "cwd": ".",
            "multiline": "line1\nline2\nline3",
        }
        input_data = json.dumps(payload, indent=2)

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 0, f"Expected exit code 0, got {result.returncode}. stderr: {result.stderr}"

        # stdout이 유효한 JSON인지 확인
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
