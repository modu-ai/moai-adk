#!/usr/bin/env python3
# @TEST:HOOKS-SCHEMA-001 | SPEC: SPEC-HOOKS-SCHEMA-001.md
"""HookResult JSON 스키마 검증 테스트

Claude Code Hook 표준 스키마를 준수하는지 검증합니다.

TDD History:
    - RED: Claude Code 표준 스키마 검증 테스트 작성
    - GREEN: HookResult.to_dict() 메서드 구현 (표준 스키마 준수)
    - REFACTOR: 테스트 케이스 확장, 엣지 케이스 처리
"""
import json
import sys
from pathlib import Path

import pytest

# Add hooks directory to path
HOOKS_DIR = Path(__file__).parent.parent.parent / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
sys.path.insert(0, str(HOOKS_DIR))

from core import HookResult


class TestHookResultSchema:
    """HookResult JSON 스키마 검증 테스트 케이스

    SPEC 요구사항 검증:
        - to_dict()는 Claude Code 표준 스키마를 반환해야 한다
        - 최상위 필드: continue, decision, reason, suppressOutput, permissionDecision
        - 커스텀 필드: hookSpecificOutput 내부에 포함
        - 불필요한 필드는 제외되어야 한다
    """

    def test_hook_result_default_continue_true(self):
        """기본 HookResult는 continue=True를 반환해야 한다

        SPEC 요구사항:
            - WHEN HookResult()가 생성되면, to_dict()는 {"continue": true}를 반환해야 한다

        Given: 기본 HookResult 객체
        When: to_dict()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        result = HookResult()
        output = result.to_dict()

        assert output == {"continue": True}
        assert isinstance(output, dict)

    def test_hook_result_with_system_message(self):
        """system_message가 있을 때 hookSpecificOutput에 포함되어야 한다

        SPEC 요구사항:
            - WHEN system_message가 설정되면, hookSpecificOutput.systemMessage에 포함되어야 한다

        Given: system_message="Test message"인 HookResult
        When: to_dict()를 호출하면
        Then: hookSpecificOutput.systemMessage="Test message"를 포함한다
        """
        result = HookResult(system_message="Test message")
        output = result.to_dict()

        assert "continue" in output
        assert output["continue"] is True
        assert "hookSpecificOutput" in output
        assert output["hookSpecificOutput"]["systemMessage"] == "Test message"

    def test_hook_result_with_context_files(self):
        """context_files가 hookSpecificOutput에 포함되어야 한다

        SPEC 요구사항:
            - WHEN context_files가 설정되면, hookSpecificOutput.contextFiles에 포함되어야 한다

        Given: context_files=["file1.txt", "file2.txt"]인 HookResult
        When: to_dict()를 호출하면
        Then: hookSpecificOutput.contextFiles에 목록이 포함된다
        """
        result = HookResult(context_files=["file1.txt", "file2.txt"])
        output = result.to_dict()

        assert "hookSpecificOutput" in output
        assert output["hookSpecificOutput"]["contextFiles"] == ["file1.txt", "file2.txt"]

    def test_hook_result_decision_block(self):
        """decision="block"일 때 reason과 함께 반환되어야 한다

        SPEC 요구사항:
            - WHEN decision="block"이고 reason이 설정되면, decision과 reason을 반환해야 한다

        Given: decision="block", reason="Dangerous"인 HookResult
        When: to_dict()를 호출하면
        Then: {"decision": "block", "reason": "Dangerous"}를 반환한다
        """
        result = HookResult(decision="block", reason="Dangerous operation")
        output = result.to_dict()

        assert output["decision"] == "block"
        assert output["reason"] == "Dangerous operation"
        assert "continue" not in output

    def test_hook_result_suppress_output(self):
        """suppress_output=True일 때 suppressOutput이 포함되어야 한다

        SPEC 요구사항:
            - WHEN suppress_output=True이면, suppressOutput을 반환해야 한다

        Given: suppress_output=True인 HookResult
        When: to_dict()를 호출하면
        Then: {"suppressOutput": true}를 포함한다
        """
        result = HookResult(suppress_output=True)
        output = result.to_dict()

        assert output["suppressOutput"] is True

    def test_hook_result_permission_decision(self):
        """permission_decision이 설정되면 반환되어야 한다

        SPEC 요구사항:
            - WHEN permission_decision이 설정되면, 반환되어야 한다

        Given: permission_decision="allow"인 HookResult
        When: to_dict()를 호출하면
        Then: {"permissionDecision": "allow"}를 포함한다
        """
        result = HookResult(permission_decision="allow")
        output = result.to_dict()

        assert output["permissionDecision"] == "allow"

    def test_hook_result_full_spec(self):
        """전체 필드가 설정된 HookResult

        SPEC 요구사항:
            - WHEN 모든 필드가 설정되면, 모든 필드가 올바르게 반환되어야 한다

        Given: 모든 필드가 설정된 HookResult
        When: to_dict()를 호출하면
        Then: 모든 필드가 올바르게 포함된다
        """
        result = HookResult(
            continue_execution=True,
            suppress_output=False,
            decision=None,
            reason=None,
            permission_decision="ask",
            system_message="Status message",
            context_files=["file1.txt"],
            suggestions=["Do this first"],
            exit_code=0,
        )
        output = result.to_dict()

        assert output["continue"] is True
        assert output["permissionDecision"] == "ask"
        assert output["hookSpecificOutput"]["systemMessage"] == "Status message"
        assert output["hookSpecificOutput"]["contextFiles"] == ["file1.txt"]
        assert output["hookSpecificOutput"]["suggestions"] == ["Do this first"]
        assert "exitCode" not in output["hookSpecificOutput"]  # 0 should be omitted

    def test_hook_result_no_old_fields(self):
        """이전 필드명(message, blocked, contextFiles)이 없어야 한다

        SPEC 요구사항:
            - WHEN to_dict()를 호출하면, 이전 필드명이 없어야 한다

        Given: HookResult 객체
        When: to_dict()를 호출하면
        Then: message, blocked, contextFiles 필드가 없다
        """
        result = HookResult(system_message="Test", context_files=["file.txt"])
        output = result.to_dict()

        # 최상위 레벨에 없어야 함
        assert "message" not in output
        assert "blocked" not in output
        assert "contextFiles" not in output

        # camelCase 버전도 없어야 함
        assert "systemMessage" not in output

    def test_hook_result_json_serializable(self):
        """to_dict() 결과가 JSON 직렬화 가능해야 한다

        SPEC 요구사항:
            - WHEN to_dict() 결과를 JSON으로 직렬화하면, 성공해야 한다

        Given: HookResult 객체
        When: json.dumps()로 직렬화하면
        Then: 유효한 JSON 문자열을 반환한다
        """
        result = HookResult(
            system_message="Test",
            context_files=["file.txt"],
            suggestions=["Action"],
        )
        output = result.to_dict()

        # Should not raise JSONEncodeError
        json_str = json.dumps(output)
        assert isinstance(json_str, str)

        # Should be parseable
        parsed = json.loads(json_str)
        assert parsed["continue"] is True

    def test_hook_result_user_prompt_submit_dict(self):
        """UserPromptSubmit 전용 스키마를 반환해야 한다

        SPEC 요구사항:
            - WHEN to_user_prompt_submit_dict()를 호출하면, UserPromptSubmit 스키마를 반환해야 한다

        Given: context_files가 있는 HookResult
        When: to_user_prompt_submit_dict()를 호출하면
        Then: hookSpecificOutput.hookEventName="UserPromptSubmit"를 포함한다
        """
        result = HookResult(context_files=["file1.txt"], system_message="Loaded files")
        output = result.to_user_prompt_submit_dict()

        assert "continue" in output
        assert output["continue"] is True
        assert "hookSpecificOutput" in output
        assert output["hookSpecificOutput"]["hookEventName"] == "UserPromptSubmit"
        assert "📎 Context: file1.txt" in output["hookSpecificOutput"]["additionalContext"]
        assert "Loaded files" in output["hookSpecificOutput"]["additionalContext"]

    def test_hook_result_empty_lists_omitted(self):
        """빈 리스트는 hookSpecificOutput에서 제외되어야 한다

        SPEC 요구사항:
            - WHEN context_files, suggestions이 빈 리스트이면, 제외되어야 한다

        Given: 빈 리스트인 HookResult
        When: to_dict()를 호출하면
        Then: hookSpecificOutput에 포함되지 않는다
        """
        result = HookResult(context_files=[], suggestions=[])
        output = result.to_dict()

        # hookSpecificOutput이 없거나 contextFiles, suggestions이 없어야 함
        if "hookSpecificOutput" in output:
            assert "contextFiles" not in output["hookSpecificOutput"]
            assert "suggestions" not in output["hookSpecificOutput"]

    def test_hook_result_exit_code_nonzero(self):
        """exit_code가 0이 아니면 hookSpecificOutput에 포함되어야 한다

        SPEC 요구사항:
            - WHEN exit_code가 0이 아니면, hookSpecificOutput.exitCode에 포함되어야 한다

        Given: exit_code=1인 HookResult
        When: to_dict()를 호출하면
        Then: hookSpecificOutput.exitCode=1를 포함한다
        """
        result = HookResult(exit_code=1)
        output = result.to_dict()

        assert "hookSpecificOutput" in output
        assert output["hookSpecificOutput"]["exitCode"] == 1

    def test_hook_result_exit_code_zero_omitted(self):
        """exit_code가 0이면 hookSpecificOutput에서 제외되어야 한다

        SPEC 요구사항:
            - WHEN exit_code가 0이면, hookSpecificOutput에서 제외되어야 한다

        Given: exit_code=0인 HookResult
        When: to_dict()를 호출하면
        Then: hookSpecificOutput에 exitCode가 없다
        """
        result = HookResult(exit_code=0)
        output = result.to_dict()

        if "hookSpecificOutput" in output:
            assert "exitCode" not in output["hookSpecificOutput"]
