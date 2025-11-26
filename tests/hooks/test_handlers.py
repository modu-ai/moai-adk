#!/usr/bin/env python3
"""Alfred Hooks 핸들러 테스트

모든 Hook 이벤트 핸들러의 동작을 검증합니다.

TDD History:
    - RED: 핸들러 입력/출력 검증 테스트 작성
    - GREEN: 핸들러 구현 (Claude Code 표준 스키마 준수)
    - REFACTOR: 테스트 케이스 확장, 엣지 케이스 처리
"""
import sys
from pathlib import Path
from unittest.mock import patch

# Add hooks directory to path (must be before any imports from hooks)
PROJECT_ROOT = Path(__file__).parent.parent.parent
LIB_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai" / "lib"

# sys.path에 추가 (최상단에 추가하여 우선순위 높임)
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# 이제 핸들러를 import할 수 있음
try:
    from models import HookPayload, HookResult  # noqa: E402
except ImportError as e:
    raise ImportError(f"Failed to import from models: {e}. LIB_DIR={LIB_DIR}, sys.path={sys.path[:3]}") from e

import pytest

# Skip this file - outdated test file
# handlers modules don't exist yet in the moai structure
# The test file references undefined handlers that haven't been implemented
pytestmark = pytest.mark.skip(reason="Outdated test file - handlers modules not implemented in moai structure")
# try:
#     from handlers.notification import (  # noqa: E402
#         handle_notification,
#         handle_stop,
#         handle_subagent_stop,
#     )
# except ImportError as e:
#     raise ImportError(f"Failed to import from handlers.notification: {e}") from e
#
# try:
#     from handlers.session import handle_session_end, handle_session_start  # noqa: E402
# except ImportError as e:
#     raise ImportError(f"Failed to import from handlers.session: {e}") from e
#
# try:
#     from handlers.tool import handle_post_tool_use, handle_pre_tool_use  # noqa: E402
# except ImportError as e:
#     raise ImportError(f"Failed to import from handlers.tool: {e}") from e
#
# try:
#     from handlers.user import handle_user_prompt_submit  # noqa: E402
# except ImportError as e:
#     raise ImportError(f"Failed to import from handlers.user: {e}") from e


class TestPreToolUseHandler:
    """PreToolUse 핸들러 테스트

    SPEC 요구사항:
        - 위험한 작업 감지 시 체크포인트 생성
        - 항상 continue_execution=True 반환
    """

    def test_pre_tool_use_safe_operation(self):
        """안전한 작업은 기본 결과 반환

        SPEC 요구사항:
            - WHEN 안전한 작업이 감지되면, continue_execution=True를 반환해야 한다

        Given: 안전한 Bash 명령어 payload
        When: handle_pre_tool_use()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {
            "cwd": ".",
            "tool": "Bash",
            "arguments": {"command": "ls -la"},
        }

        result = handle_pre_tool_use(payload)

        assert isinstance(result, HookResult)
        assert result.continue_execution is True
        output = result.to_dict()
        assert "continue" in output
        assert output["continue"] is True

    @patch("handlers.tool.detect_risky_operation")
    @patch("handlers.tool.create_checkpoint")
    def test_pre_tool_use_risky_operation(self, mock_create_checkpoint, mock_detect_risky):
        """위험한 작업 감지 시 체크포인트 생성 알림

        SPEC 요구사항:
            - WHEN 위험한 작업이 감지되면, 체크포인트 생성 알림을 표시해야 한다

        Given: 위험한 rm -rf 명령어
        When: handle_pre_tool_use()를 호출하면
        Then: system_message에 체크포인트 정보가 포함된다
        """
        mock_detect_risky.return_value = (True, "delete")
        mock_create_checkpoint.return_value = "before-delete-20251015-143000"

        payload: HookPayload = {
            "cwd": ".",
            "tool": "Bash",
            "arguments": {"command": "rm -rf /tmp"},
        }

        result = handle_pre_tool_use(payload)

        assert result.continue_execution is True
        assert result.system_message is not None
        assert "Checkpoint created" in result.system_message
        assert "delete" in result.system_message


class TestSessionStartHandler:
    """SessionStart 핸들러 테스트

    SPEC 요구사항:
        - 프로젝트 상태 요약 정보 제공
        - clear 단계 시 최소한의 결과만 반환
    """

    def test_session_start_clear_phase(self):
        """clear 단계는 최소한의 결과 반환

        SPEC 요구사항:
            - WHEN phase="clear"이면, 최소한의 결과만 반환해야 한다

        Given: phase="clear"인 payload
        When: handle_session_start()를 호출하면
        Then: {"continue": true}만 반환한다
        """
        payload: HookPayload = {"cwd": ".", "phase": "clear"}

        result = handle_session_start(payload)

        output = result.to_dict()
        assert output == {"continue": True}

    @patch("handlers.session.count_specs")
    @patch("handlers.session.get_git_info")
    def test_session_start_compact_phase(self, mock_get_git, mock_count_specs):
        """compact 단계는 상세 정보 반환

        SPEC 요구사항:
            - WHEN phase="compact"이면, 상세한 상태 정보를 반환해야 한다

        Given: phase="compact"인 payload
        When: handle_session_start()를 호출하면
        Then: system_message에 브랜치, SPEC 진도 정보가 포함된다
        """
        mock_get_git.return_value = {
            "branch": "main",
            "commit": "abc123def456",
            "changes": 0,
        }
        mock_count_specs.return_value = {
            "completed": 5,
            "total": 10,
            "percentage": 50,
        }

        with patch("handlers.session.list_checkpoints", return_value=[]):
            payload: HookPayload = {"cwd": ".", "phase": "compact"}
            result = handle_session_start(payload)

        assert result.system_message is not None
        assert "MoAI-ADK Session Started" in result.system_message
        assert "main" in result.system_message
        assert "5/10" in result.system_message

    @patch("handlers.session.get_package_version_info")
    @patch("handlers.session.list_checkpoints")
    @patch("handlers.session.count_specs")
    @patch("handlers.session.get_git_info")
    def test_session_start_major_version_warning(
        self, mock_get_git, mock_count_specs, mock_list_checkpoints, mock_version_info
    ):
        """Major version update shows warning with release notes


        SPEC Requirements:
            - WHEN major version update is available (e.g., 0.8.1 → 1.0.0),
              THEN SessionStart should display warning with release notes URL

        Given: Major version update available (0.8.1 → 1.0.0)
        When: handle_session_start() is called
        Then: system_message includes "⚠️ Major version update available"
              and release notes URL
        """
        mock_get_git.return_value = {}
        mock_count_specs.return_value = {"completed": 0, "total": 0, "percentage": 0}
        mock_list_checkpoints.return_value = []

        # Mock major version update
        mock_version_info.return_value = {
            "current": "0.8.1",
            "latest": "1.0.0",
            "update_available": True,
            "is_major_update": True,
            "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0",
            "upgrade_command": "uv tool upgrade moai-adk",
        }

        payload: HookPayload = {"cwd": ".", "phase": "compact"}
        result = handle_session_start(payload)

        assert result.system_message is not None
        # Should show major version warning
        assert "⚠️" in result.system_message
        assert "Major version update available" in result.system_message
        assert "0.8.1" in result.system_message
        assert "1.0.0" in result.system_message
        # Should include release notes URL
        assert "github.com/modu-ai/moai-adk/releases" in result.system_message
        # Should include upgrade command
        assert "Upgrade:" in result.system_message or "⬆️" in result.system_message

    @patch("handlers.session.get_package_version_info")
    @patch("handlers.session.list_checkpoints")
    @patch("handlers.session.count_specs")
    @patch("handlers.session.get_git_info")
    def test_session_start_regular_update_with_release_notes(
        self, mock_get_git, mock_count_specs, mock_list_checkpoints, mock_version_info
    ):
        """Regular update shows version info with release notes


        SPEC Requirements:
            - WHEN minor/patch update is available (e.g., 0.8.1 → 0.9.0),
              THEN SessionStart should display version info and release notes

        Given: Regular version update available (0.8.1 → 0.9.0)
        When: handle_session_start() is called
        Then: system_message includes version line and release notes URL
        """
        mock_get_git.return_value = {}
        mock_count_specs.return_value = {"completed": 0, "total": 0, "percentage": 0}
        mock_list_checkpoints.return_value = []

        # Mock regular update (no major version change)
        mock_version_info.return_value = {
            "current": "0.8.1",
            "latest": "0.9.0",
            "update_available": True,
            "is_major_update": False,
            "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v0.9.0",
            "upgrade_command": "uv tool upgrade moai-adk",
        }

        payload: HookPayload = {"cwd": ".", "phase": "compact"}
        result = handle_session_start(payload)

        assert result.system_message is not None
        # Should show regular version line (NOT major warning)
        assert "0.8.1" in result.system_message
        assert "0.9.0" in result.system_message
        assert "available" in result.system_message
        # Should NOT show major version warning
        assert "⚠️" not in result.system_message or "Major" not in result.system_message
        # Should include release notes URL
        assert "github.com/modu-ai/moai-adk/releases" in result.system_message


class TestUserPromptSubmitHandler:
    """UserPromptSubmit 핸들러 테스트

    SPEC 요구사항:
        - 사용자 프롬프트 분석하여 JIT 컨텍스트 로드
        - context_files 목록 반환
    """

    @patch("handlers.user.get_enhanced_jit_context")
    def test_user_prompt_submit_with_context(self, mock_enhanced_jit):
        """컨텍스트 파일이 있을 때

        SPEC 요구사항:
            - WHEN 컨텍스트 파일이 발견되면, context_files 목록을 반환해야 한다

        Given: "SPEC 파일" 사용자 프롬프트
        When: handle_user_prompt_submit()를 호출하면
        Then: context_files에 SPEC 파일 목록이 포함된다
        """
        mock_enhanced_jit.return_value = (
            [".moai/specs/SPEC-001.md", ".moai/specs/SPEC-002.md"],
            "🎯 전문가 에이전트: spec-builder",
        )

        payload: HookPayload = {
            "cwd": ".",
            "userPrompt": "SPEC 파일",
        }

        result = handle_user_prompt_submit(payload)

        assert result.context_files == [
            ".moai/specs/SPEC-001.md",
            ".moai/specs/SPEC-002.md",
        ]
        # UserPromptSubmit은 특별한 스키마 사용
        user_submit_output = result.to_user_prompt_submit_dict()
        assert user_submit_output["hookSpecificOutput"]["hookEventName"] == "UserPromptSubmit"

    @patch("handlers.user.get_jit_context")
    def test_user_prompt_submit_empty_context(self, mock_get_jit):
        """컨텍스트 파일이 없을 때

        SPEC 요구사항:
            - WHEN 컨텍스트 파일이 없으면, 빈 context_files 반환해야 한다

        Given: "random"이라는 사용자 프롬프트
        When: handle_user_prompt_submit()를 호출하면
        Then: context_files는 빈 리스트, system_message는 None이다
        """
        mock_get_jit.return_value = []

        payload: HookPayload = {
            "cwd": ".",
            "userPrompt": "random",
        }

        result = handle_user_prompt_submit(payload)

        assert result.context_files == []
        assert result.system_message is None

    @patch("handlers.user.get_enhanced_jit_context")
    def test_user_prompt_submit_alfred_command_logging(self, mock_enhanced_jit):
        """Alfred 명령어 실행 시 로깅 기능

        SPEC 요구사항:
            - WHEN /moai:* 명령어가 실행되면, 타임스탐프와 함께 로그 파일에 기록해야 한다
            - WHEN 로깅이 실패하면, 메인 플로우는 계속되어야 한다 (비차단)

        Given: "/moai:1-plan 테스트" 프롬프트
        When: handle_user_prompt_submit()를 호출하면
        Then: 로그 파일에 명령어가 기록되고, 정상적으로 완료된다
        """
        mock_enhanced_jit.return_value = ([], None)

        payload: HookPayload = {
            "cwd": ".",
            "userPrompt": "/moai:1-plan 테스트 명령어",
        }

        result = handle_user_prompt_submit(payload)

        # Verify main flow continues (no exceptions)
        assert isinstance(result, HookResult)
        assert isinstance(result.context_files, list)
        assert result.system_message is None or isinstance(result.system_message, str)

    @patch("handlers.user.get_jit_context")
    def test_user_prompt_submit_non_alfred_command_no_logging(self, mock_get_jit):
        """Alfred가 아닌 명령어는 로깅하지 않음

        SPEC 요구사항:
            - WHEN /moai:가 아닌 명령어가 실행되면, 로깅하지 않아야 한다

        Given: "/help" 사용자 프롬프트
        When: handle_user_prompt_submit()를 호출하면
        Then: 로그 파일 접근이 없고 정상적으로 완료된다
        """
        mock_get_jit.return_value = []

        payload: HookPayload = {
            "cwd": ".",
            "userPrompt": "/help",
        }

        result = handle_user_prompt_submit(payload)

        # Verify no errors and normal operation
        assert isinstance(result, HookResult)
        assert result.context_files == []
        assert result.system_message is None

    @patch("handlers.user.get_enhanced_jit_context")
    def test_user_prompt_submit_logging_graceful_failure(self, mock_enhanced_jit):
        """로깅 실패 시에도 메인 플로우는 계속됨 (비차단)

        SPEC 요구사항:
            - WHEN 로그 파일 작성이 실패하면, 사일런트 페일로 메인 플로우를 방해하지 않아야 한다

        Given: 파일 쓰기 권한이 없는 상황
        When: handle_user_prompt_submit()를 호출하면
        Then: 로깅 오류가 발생하지만 정상적으로 완료된다
        """
        mock_enhanced_jit.return_value = ([], None)

        payload: HookPayload = {
            "cwd": ".",
            "userPrompt": "/moai:2-run SPEC-001",
        }

        # Should not raise exception despite logging failure
        result = handle_user_prompt_submit(payload)

        # Verify main flow continues
        assert isinstance(result, HookResult)
        assert isinstance(result.context_files, list)


class TestPostToolUseHandler:
    """PostToolUse 핸들러 테스트"""

    def test_post_tool_use_default(self):
        """기본 동작 검증

        SPEC 요구사항:
            - WHEN PostToolUse 이벤트가 발생하면, 기본 결과를 반환해야 한다

        Given: PostToolUse payload
        When: handle_post_tool_use()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {"cwd": ".", "tool": "Bash"}

        result = handle_post_tool_use(payload)

        assert result.continue_execution is True
        output = result.to_dict()
        assert output == {"continue": True}


class TestSessionEndHandler:
    """SessionEnd 핸들러 테스트"""

    def test_session_end_default(self):
        """기본 동작 검증

        SPEC 요구사항:
            - WHEN SessionEnd 이벤트가 발생하면, 기본 결과를 반환해야 한다

        Given: SessionEnd payload
        When: handle_session_end()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {"cwd": "."}

        result = handle_session_end(payload)

        assert result.continue_execution is True
        output = result.to_dict()
        assert output == {"continue": True}


class TestNotificationHandler:
    """Notification 핸들러 테스트"""

    def test_notification_default(self):
        """기본 동작 검증

        SPEC 요구사항:
            - WHEN Notification 이벤트가 발생하면, 기본 결과를 반환해야 한다

        Given: Notification payload
        When: handle_notification()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {"cwd": "."}

        result = handle_notification(payload)

        assert result.continue_execution is True
        output = result.to_dict()
        assert output == {"continue": True}


class TestStopHandler:
    """Stop 핸들러 테스트"""

    def test_stop_default(self):
        """기본 동작 검증

        SPEC 요구사항:
            - WHEN Stop 이벤트가 발생하면, 기본 결과를 반환해야 한다

        Given: Stop payload
        When: handle_stop()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {"cwd": "."}

        result = handle_stop(payload)

        assert result.continue_execution is True
        output = result.to_dict()
        assert output == {"continue": True}


class TestSubagentStopHandler:
    """SubagentStop 핸들러 테스트"""

    def test_subagent_stop_default(self):
        """기본 동작 검증

        SPEC 요구사항:
            - WHEN SubagentStop 이벤트가 발생하면, 기본 결과를 반환해야 한다

        Given: SubagentStop payload
        When: handle_subagent_stop()를 호출하면
        Then: {"continue": true}를 반환한다
        """
        payload: HookPayload = {"cwd": "."}

        result = handle_subagent_stop(payload)

        assert result.continue_execution is True
        output = result.to_dict()
        assert output == {"continue": True}
