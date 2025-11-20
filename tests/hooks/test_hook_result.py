#!/usr/bin/env python3
"""HookResult JSON 스키마 검증 테스트

Claude Code Hook 표준 스키마를 준수하는지 검증합니다.

TDD History:
    - RED: Claude Code 표준 스키마 검증 테스트 작성
    - GREEN: HookResult.to_dict() 메서드 구현 (표준 스키마 준수)
    - REFACTOR: 실제 HookResult API에 맞게 테스트 수정
"""
import json
import sys
from pathlib import Path

# Add hooks directory to path
PROJECT_ROOT = Path(__file__).parent.parent.parent
LIB_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai" / "lib"
sys.path.insert(0, str(LIB_DIR))

from models import HookResult  # noqa: E402


class TestHookResultSchema:
    """HookResult JSON 스키마 검증 테스트 케이스

    SPEC 요구사항 검증:
        - to_dict()는 Claude Code 표준 스키마를 반환해야 한다
        - 실제 필드: continue_execution, block_execution, system_message, context_files, hook_specific_output
        - 불필요한 필드는 제외되어야 한다
    """

    def test_hook_result_default_continue_true(self):
        """기본 HookResult는 continue_execution=True를 반환해야 한다

        SPEC 요구사항:
            - WHEN HookResult()가 생성되면, to_dict()는 {"continue_execution": True, "block_execution": False}를 반환해야 한다

        Given: 기본 HookResult 객체
        When: to_dict()를 호출하면
        Then: {"continue_execution": True, "block_execution": False}를 반환한다
        """
        result = HookResult()
        output = result.to_dict()

        assert output == {"continue_execution": True, "block_execution": False}
        assert isinstance(output, dict)

    def test_hook_result_with_system_message(self):
        """system_message가 있을 때 system_message 필드에 포함되어야 한다

        SPEC 요구사항:
            - WHEN system_message가 설정되면, system_message가 필드로 포함되어야 한다

        Given: system_message="Test message"인 HookResult
        When: to_dict()를 호출하면
        Then: system_message="Test message"를 반환한다
        """
        result = HookResult(system_message="Test message")
        output = result.to_dict()

        assert "continue_execution" in output
        assert output["continue_execution"] is True
        assert "system_message" in output
        assert output["system_message"] == "Test message"
        assert output["block_execution"] is False

    def test_hook_result_with_context_files(self):
        """context_files가 설정되면 to_dict()에 포함되어야 한다

        SPEC 요구사항:
            - WHEN context_files가 설정되면, context_files 필드로 포함되어야 한다

        Given: context_files=["file1.txt", "file2.txt"]인 HookResult
        When: to_dict()를 호출하면
        Then: context_files가 JSON에 포함된다
        """
        result = HookResult(context_files=["file1.txt", "file2.txt"])
        output = result.to_dict()

        assert "context_files" in output
        assert output["context_files"] == ["file1.txt", "file2.txt"]
        assert result.context_files == ["file1.txt", "file2.txt"]

    def test_hook_result_block_execution_true(self):
        """block_execution=True일 때 반환되어야 한다

        SPEC 요구사항:
            - WHEN block_execution=True이면, block_execution: True를 반환해야 한다

        Given: block_execution=True인 HookResult
        When: to_dict()를 호출하면
        Then: {"block_execution": True}를 포함한다
        """
        result = HookResult(block_execution=True)
        output = result.to_dict()

        assert output["block_execution"] is True
        assert output["continue_execution"] is True  # 기본값

    def test_hook_result_continue_false_block_true(self):
        """continue_execution=False, block_execution=True일 때 모두 반환되어야 한다

        SPEC 요구사항:
            - WHEN continue_execution=False이고 block_execution=True이면, 둘 다 반환해야 한다

        Given: continue_execution=False, block_execution=True인 HookResult
        When: to_dict()를 호출하면
        Then: 두 필드 모두 포함된다
        """
        result = HookResult(continue_execution=False, block_execution=True)
        output = result.to_dict()

        assert output["continue_execution"] is False
        assert output["block_execution"] is True

    def test_hook_result_with_hook_specific_output(self):
        """hook_specific_output이 설정되면 반환되어야 한다

        SPEC 요구사항:
            - WHEN hook_specific_output이 설정되면, 반환되어야 한다

        Given: hook_specific_output={"key": "value"}인 HookResult
        When: to_dict()를 호출하면
        Then: {"hook_specific_output": {"key": "value"}}를 포함한다
        """
        result = HookResult(hook_specific_output={"key": "value"})
        output = result.to_dict()

        assert "hook_specific_output" in output
        assert output["hook_specific_output"] == {"key": "value"}

    def test_hook_result_full_spec(self):
        """전체 필드가 설정된 HookResult

        SPEC 요구사항:
            - WHEN 모든 필드가 설정되면, Claude Code 표준 스키마로 반환되어야 한다

        Given: 모든 필드가 설정된 HookResult
        When: to_dict()를 호출하면
        Then: 모든 필드가 포함된다
        """
        result = HookResult(
            continue_execution=True,
            block_execution=False,
            system_message="Status message",
            context_files=["file1.txt"],
            hook_specific_output={"custom": "data"},
        )
        output = result.to_dict()

        assert output["continue_execution"] is True
        assert output["block_execution"] is False
        assert output["system_message"] == "Status message"
        assert output["context_files"] == ["file1.txt"]
        assert output["hook_specific_output"] == {"custom": "data"}

    def test_hook_result_empty_values_omitted(self):
        """빈 값은 to_dict()에서 제외되어야 한다

        SPEC 요구사항:
            - WHEN 값이 None, 빈 리스트, 빈 딕셔너리이면, 제외되어야 한다
            - 단, False와 0은 유지되어야 한다

        Given: 빈 값이 있는 HookResult
        When: to_dict()를 호출하면
        Then: 빈 값은 제외되고 False/0은 유지된다
        """
        result = HookResult(
            system_message=None,
            context_files=[],
            hook_specific_output={},
        )
        output = result.to_dict()

        # 빈 리스트와 빈 딕셔너리는 제외
        assert "system_message" not in output
        assert "context_files" not in output
        assert "hook_specific_output" not in output

        # 기본 boolean 필드는 유지 (False도 유지됨)
        assert "continue_execution" in output
        assert "block_execution" in output

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
            hook_specific_output={"action": "test"},
        )
        output = result.to_dict()

        # Should not raise JSONEncodeError
        json_str = json.dumps(output)
        assert isinstance(json_str, str)

        # Should be parseable
        parsed = json.loads(json_str)
        assert parsed["continue_execution"] is True

    def test_hook_result_default_empty_lists(self):
        """빈 리스트는 __post_init__에서 설정되어야 한다

        SPEC 요구사항:
            - WHEN context_files이 None이면, 빈 리스트로 초기화되어야 한다

        Given: context_files를 설정하지 않은 HookResult
        When: 객체를 생성하면
        Then: context_files는 빈 리스트로 초기화된다
        """
        result = HookResult()

        # 객체 속성으로는 빈 리스트로 초기화됨
        assert result.context_files == []
        assert isinstance(result.context_files, list)

    def test_hook_result_default_empty_dict(self):
        """빈 딕셔너리는 __post_init__에서 설정되어야 한다

        SPEC 요구사항:
            - WHEN hook_specific_output이 None이면, 빈 딕셔너리로 초기화되어야 한다

        Given: hook_specific_output을 설정하지 않은 HookResult
        When: 객체를 생성하면
        Then: hook_specific_output은 빈 딕셔너리로 초기화된다
        """
        result = HookResult()

        # 객체 속성으로는 빈 딕셔너리로 초기화됨
        assert result.hook_specific_output == {}
        assert isinstance(result.hook_specific_output, dict)

    def test_hook_result_false_values_preserved(self):
        """False와 0 값은 to_dict()에서 유지되어야 한다

        SPEC 요구사항:
            - WHEN 값이 False 또는 0이면, to_dict()에 포함되어야 한다

        Given: continue_execution=False, block_execution=False인 HookResult
        When: to_dict()를 호출하면
        Then: False 값이 유지된다
        """
        result = HookResult(continue_execution=False, block_execution=False)
        output = result.to_dict()

        # False 값도 포함되어야 함 (빈 값이 아니므로)
        assert "continue_execution" in output
        assert output["continue_execution"] is False
        assert "block_execution" in output
        assert output["block_execution"] is False

    def test_hook_result_complex_hook_specific_output(self):
        """복잡한 hook_specific_output이 올바르게 직렬화되어야 한다

        SPEC 요구사항:
            - WHEN hook_specific_output에 중첩된 데이터가 있으면, 올바르게 직렬화되어야 한다

        Given: 복잡한 구조의 hook_specific_output
        When: to_dict()를 호출하면
        Then: 중첩된 데이터가 올바르게 포함된다
        """
        result = HookResult(
            hook_specific_output={
                "nested": {"key": "value"},
                "list": [1, 2, 3],
                "mixed": {"items": ["a", "b"], "count": 2},
            }
        )
        output = result.to_dict()

        assert "hook_specific_output" in output
        assert output["hook_specific_output"]["nested"]["key"] == "value"
        assert output["hook_specific_output"]["list"] == [1, 2, 3]
        assert output["hook_specific_output"]["mixed"]["count"] == 2

        # JSON 직렬화 가능 확인
        json_str = json.dumps(output)
        parsed = json.loads(json_str)
        assert parsed["hook_specific_output"]["nested"]["key"] == "value"
