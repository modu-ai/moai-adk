#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# ///
"""Test Hook Output Validation

자동 테스트: Claude Code Hook JSON 스키마 검증

- SessionStart Hook JSON 출력 검증
- UserPromptSubmit Hook 특수 스키마 검증
- 모든 Hook 이벤트 스키마 일관성 검증

실행:
    uv run test_hook_output.py
"""

import json
import sys
from pathlib import Path

# Add hooks directory to sys.path
HOOKS_DIR = Path(__file__).parent
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))

from core import HookResult  # noqa: E402


def test_basic_output():
    """Test 1: Basic output with only continue flag"""
    result = HookResult(continue_execution=True)
    output = result.to_dict()

    assert output == {"continue": True}, f"Expected {{'continue': True}}, got {output}"
    print("✅ Test 1: Basic output - PASSED")


def test_system_message_top_level():
    """Test 2: systemMessage at TOP-LEVEL (not in hookSpecificOutput)"""
    result = HookResult(system_message="Test message")
    output = result.to_dict()

    assert "systemMessage" in output, "systemMessage not found in output"
    assert output["systemMessage"] == "Test message"
    assert "hookSpecificOutput" not in output, "hookSpecificOutput should not be in to_dict() output"
    print("✅ Test 2: systemMessage (top-level) - PASSED")


def test_decision_with_reason():
    """Test 3: decision + reason (block pattern)"""
    result = HookResult(decision="block", reason="Dangerous operation")
    output = result.to_dict()

    assert output.get("decision") == "block"
    assert output.get("reason") == "Dangerous operation"
    assert "continue" not in output, "continue should not appear when decision is set"
    print("✅ Test 3: decision + reason - PASSED")


def test_user_prompt_submit_schema():
    """Test 4: UserPromptSubmit special schema"""
    result = HookResult(context_files=["tests/", "docs/"])
    output = result.to_user_prompt_submit_dict()

    assert "hookSpecificOutput" in output
    assert output["hookSpecificOutput"]["hookEventName"] == "UserPromptSubmit"
    assert "additionalContext" in output["hookSpecificOutput"]
    assert "📎 Context: tests/" in output["hookSpecificOutput"]["additionalContext"]
    print("✅ Test 4: UserPromptSubmit schema - PASSED")


def test_permission_decision():
    """Test 5: permissionDecision field"""
    result = HookResult(permission_decision="deny")
    output = result.to_dict()

    assert output.get("permissionDecision") == "deny"
    assert "continue" in output  # continue should still be present
    print("✅ Test 5: permissionDecision - PASSED")


def test_session_start_typical_output():
    """Test 6: Typical SessionStart output"""
    result = HookResult(
        continue_execution=True,
        system_message="🚀 MoAI-ADK Session Started\n   Language: python\n   Branch: develop"
    )
    output = result.to_dict()

    # Validate schema
    assert "continue" in output or "decision" in output, "Missing continue or decision"
    assert output.get("systemMessage", "").startswith("🚀 MoAI-ADK")

    # Ensure internal fields are NOT in output
    assert "context_files" not in output, "Internal field context_files leaked to output"
    assert "suggestions" not in output, "Internal field suggestions leaked to output"
    assert "exit_code" not in output, "Internal field exit_code leaked to output"

    print("✅ Test 6: SessionStart typical output - PASSED")


def test_json_serializable():
    """Test 7: Output is JSON serializable"""
    result = HookResult(
        system_message="Test",
        decision="approve",
        reason="Valid operation"
    )
    output = result.to_dict()

    try:
        json_str = json.dumps(output)
        parsed = json.loads(json_str)
        assert parsed == output
        print("✅ Test 7: JSON serializable - PASSED")
    except Exception as e:
        print(f"❌ Test 7: JSON serialization FAILED: {e}")
        sys.exit(1)


def test_user_prompt_submit_with_system_message():
    """Test 8: UserPromptSubmit with both context and system message"""
    result = HookResult(
        context_files=["src/"],
        system_message="Loading context..."
    )
    output = result.to_user_prompt_submit_dict()

    assert "hookSpecificOutput" in output
    assert "Loading context..." in output["hookSpecificOutput"]["additionalContext"]
    assert "📎 Context: src/" in output["hookSpecificOutput"]["additionalContext"]
    print("✅ Test 8: UserPromptSubmit with system_message - PASSED")


def main():
    """Run all tests"""
    print("\n" + "="*60)
    print("🧪 Claude Code Hook Output Validation Tests")
    print("="*60 + "\n")

    tests = [
        test_basic_output,
        test_system_message_top_level,
        test_decision_with_reason,
        test_user_prompt_submit_schema,
        test_permission_decision,
        test_session_start_typical_output,
        test_json_serializable,
        test_user_prompt_submit_with_system_message,
    ]

    failed = 0
    for test in tests:
        try:
            test()
        except AssertionError as e:
            print(f"❌ {test.__name__}: FAILED - {e}")
            failed += 1
        except Exception as e:
            print(f"❌ {test.__name__}: ERROR - {e}")
            failed += 1

    print("\n" + "="*60)
    if failed == 0:
        print(f"✅ ALL {len(tests)} TESTS PASSED")
        print("="*60 + "\n")
        sys.exit(0)
    else:
        print(f"❌ {failed}/{len(tests)} TESTS FAILED")
        print("="*60 + "\n")
        sys.exit(1)


if __name__ == "__main__":
    main()
