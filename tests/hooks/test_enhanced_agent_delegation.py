#!/usr/bin/env python3
"""향상된 에이전트 위임 Hook 테스트

사용자 프롬프트 분석을 통한 전문가 에이전트 위임 및 Skills JIT 컨텍스트 로딩 기능 테스트.

TDD History:
    - RED: 에이전트 위임 기능 테스트 작성 (아직 미구현)
    - GREEN: agent_context.py 모듈 구현으로 기능 완성
    - REFACTOR: 확장성 있는 패턴 매칭 및 설정 파일화
"""

import sys
import json
from pathlib import Path
from unittest.mock import Mock, patch

# Hook 디렉토리를 sys.path에 추가
HOOKS_DIR = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred"
SHARED_DIR = HOOKS_DIR / "shared"
UTILS_DIR = HOOKS_DIR / "utils"
SRC_DIR = Path(__file__).parent.parent.parent / "src"

# sys.path에 추가 (중복 방지)
for path in [str(SHARED_DIR), str(HOOKS_DIR), str(UTILS_DIR), str(SRC_DIR)]:
    if path not in sys.path:
        sys.path.insert(0, path)

import pytest


class TestAgentContextModule:
    """에이전트 컨텍스트 모듈 테스트"""

    def test_load_agent_skills_mapping(self):
        """에이전트-Skills 매핑 로드 테스트"""
        from shared.core.agent_context import load_agent_skills_mapping

        # 매핑 파일 로드
        mapping = load_agent_skills_mapping()

        # 필수 키 확인
        assert "agent_skills_mapping" in mapping
        assert "prompt_patterns" in mapping

        # 에이전트-Skills 매핑 확인
        agent_mapping = mapping["agent_skills_mapping"]
        assert "spec-builder" in agent_mapping
        assert "tdd-implementer" in agent_mapping
        assert "test-engineer" in agent_mapping

        # 패턴 설정 확인
        prompt_patterns = mapping["prompt_patterns"]
        assert "spec_creation" in prompt_patterns
        assert "implementation" in prompt_patterns
        assert "testing" in prompt_patterns

    def test_analyze_prompt_intent_spec_creation(self):
        """SPEC 생성 의도 분석 테스트"""
        from shared.core.agent_context import analyze_prompt_intent, load_agent_skills_mapping

        mapping = load_agent_skills_mapping()
        prompt = "새로운 기능에 대한 명세서를 작성해주세요"

        result = analyze_prompt_intent(prompt, mapping)

        # 결과 확인
        assert result is not None
        assert result["intent"] == "spec_creation"
        assert result["primary_agent"] == "spec-builder"
        assert "tdd-implementer" not in result["primary_agent"]
        assert result["confidence"] > 0.5
        assert len(result["matched_keywords"]) > 0
        assert "spec-builder" in result["recommended_skills"]

    def test_analyze_prompt_intent_implementation(self):
        """구현 의도 분석 테스트"""
        from shared.core.agent_context import analyze_prompt_intent, load_agent_skills_mapping

        mapping = load_agent_skills_mapping()
        prompt = "/alfred:2-run SPEC-001을 구현해주세요"

        result = analyze_prompt_intent(prompt, mapping)

        # 결과 확인
        assert result is not None
        assert result["intent"] == "implementation"
        assert result["primary_agent"] == "tdd-implementer"
        assert "test-engineer" in result["secondary_agents"]
        assert "quality-gate" in result["secondary_agents"]
        assert result["confidence"] > 0.5

    def test_analyze_prompt_intent_no_match(self):
        """일치하는 패턴 없을 때 테스트"""
        from shared.core.agent_context import analyze_prompt_intent, load_agent_skills_mapping

        mapping = load_agent_skills_mapping()
        prompt = "오늘 날씨 어때요?"

        result = analyze_prompt_intent(prompt, mapping)

        # 일치하는 패턴 없음
        assert result is None

    def test_get_agent_delegation_context_with_intent(self):
        """에이전트 위임 컨텍스트 생성 테스트 (의도 있음)"""
        from shared.core.agent_context import get_agent_delegation_context

        prompt = "테스트 코드를 작성해주세요"
        cwd = "/tmp/test_project"

        # Mock Path.exists() for context files
        with patch('pathlib.Path.exists') as mock_exists:
            mock_exists.return_value = True

            result = get_agent_delegation_context(prompt, cwd)

            # 결과 확인
            assert result["intent_detected"] is True
            assert result["primary_agent"] == "test-engineer"
            assert result["confidence"] > 0.5
            assert "test" in result["matched_keywords"]
            assert len(result["recommended_skills"]) > 0
            assert len(result["context_files"]) >= 0  # 파일이 있을 수도 있고 없을 수도 있음

    def test_get_agent_delegation_context_without_intent(self):
        """에이전트 위임 컨텍스트 생성 테스트 (의도 없음)"""
        from shared.core.agent_context import get_agent_delegation_context

        prompt = "일반적인 대화"
        cwd = "/tmp/test_project"

        result = get_agent_delegation_context(prompt, cwd)

        # 결과 확인
        assert result["intent_detected"] is False
        assert "primary_agent" not in result
        assert "traditional_context" in result

    def test_format_agent_delegation_message_with_high_confidence(self):
        """높은 신뢰도 에이전트 위임 메시지 포맷팅 테스트"""
        from shared.core.agent_context import format_agent_delegation_message

        context = {
            "intent_detected": True,
            "primary_agent": "frontend-expert",
            "confidence": 0.8,
            "intent": "frontend",
            "matched_keywords": ["react", "ui"],
            "recommended_skills": ["moai-domain-frontend", "component-designer"],
            "secondary_agents": ["ui-ux-expert"],
            "context_files": ["src/App.js", "package.json"],
            "skill_references": [".claude/skills/moai-domain-frontend/reference.md"]
        }

        message = format_agent_delegation_message(context)

        # 메시지 확인
        assert message is not None
        assert "frontend-expert" in message
        assert "react" in message or "ui" in message
        assert "moai-domain-frontend" in message
        assert "ui-ux-expert" in message
        assert "개 파일" in message

    def test_format_agent_delegation_message_with_low_confidence(self):
        """낮은 신뢰도 에이전트 위임 메시지 포맷팅 테스트"""
        from shared.core.agent_context import format_agent_delegation_message

        context = {
            "intent_detected": True,
            "primary_agent": "backend-expert",
            "confidence": 0.3,  # 낮은 신뢰도
            "intent": "backend",
            "matched_keywords": []
        }

        message = format_agent_delegation_message(context)

        # 신뢰도가 낮으면 메시지 없음
        assert message is None

    def test_get_enhanced_jit_context_integration(self):
        """향상된 JIT 컨텍스트 통합 테스트"""
        from shared.core.agent_context import get_enhanced_jit_context

        prompt = "새로운 API를 설계하고 구현해주세요"
        cwd = "/tmp/test_project"

        # Mock file existence
        with patch('pathlib.Path.exists') as mock_exists:
            # Mock different files differently
            def exists_side_effect(self):
                if "skill" in str(self).lower():
                    return True
                elif "src/api" in str(self):
                    return True
                return False

            mock_exists.side_effect = exists_side_effect

            context_files, system_message = get_enhanced_jit_context(prompt, cwd)

            # 결과 확인
            assert isinstance(context_files, list)
            assert len(context_files) >= 0
            assert system_message is not None
            assert "전문가 에이전트" in system_message or "backend" in system_message.lower()

            # 컨텍스트 파일 중복 확인
            assert len(context_files) == len(set(context_files))  # 중복 없음


class TestEnhancedUserHandler:
    """향상된 사용자 핸들러 테스트"""

    def test_handle_user_prompt_submit_with_agent_delegation(self):
        """에이전트 위임과 함께 사용자 프롬프트 제출 핸들러 테스트"""
        from shared.handlers.user import handle_user_prompt_submit
        from shared.core import HookPayload

        # Mock the enhanced context
        with patch('shared.handlers.user.get_enhanced_jit_context') as mock_context:
            mock_context.return_value (
                [".claude/skills/moai-domain-backend/reference.md", "src/api/"],
                "🎯 전문가 에이전트 추천: backend-expert"
            )

            payload = HookPayload(
                userPrompt="API 엔드포인트를 구현해주세요",
                cwd="/tmp/test_project"
            )

            result = handle_user_prompt_submit(payload)

            # 결과 확인
            assert result is not None
            assert result.system_message is not None
            assert "에이전트" in result.system_message
            assert len(result.context_files) > 0
            assert any("skills" in str(f) for f in result.context_files)

    def test_handle_user_prompt_submit_backward_compatibility(self):
        """이전 버전과의 호환성 테스트"""
        from shared.handlers.user import handle_user_prompt_submit
        from shared.core import HookPayload

        # Mock traditional context (no agent delegation)
        with patch('shared.handlers.user.get_enhanced_jit_context') as mock_context:
            mock_context.return_value ([], None)

            payload = HookPayload(
                userPrompt="간단한 질문입니다",
                cwd="/tmp/test_project"
            )

            result = handle_user_prompt_submit(payload)

            # 결과 확인
            assert result is not None
            assert result.context_files == []

    def test_handle_user_prompt_submit_alfred_command_logging(self):
        """Alfred 명령어 로깅 테스트"""
        from shared.handlers.user import handle_user_prompt_submit
        from shared.core import HookPayload

        with patch('shared.handlers.user.get_enhanced_jit_context') as mock_context, \
             patch('builtins.open', create=True) as mock_open, \
             patch('pathlib.Path.mkdir') as mock_mkdir, \
             patch('datetime.datetime.now') as mock_datetime:

            # Mock enhanced context with agent delegation
            mock_context.return_value (
                [".claude/skills/moai-alfred-spec-authoring/reference.md"],
                "🎯 전문가 에이전트 추천: spec-builder"
            )

            # Mock datetime
            mock_now = Mock()
            mock_now.isoformat.return_value = "2024-01-01T12:00:00"
            mock_datetime.now.return_value = mock_now

            payload = HookPayload(
                userPrompt="/alfred:1-plan 새로운 기능 명세",
                cwd="/tmp/test_project"
            )

            result = handle_user_prompt_submit(payload)

            # 로깅 확인 (mkdir 호출)
            mock_mkdir.assert_called_once()

            # 결과 확인
            assert result is not None
            assert result.system_message is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])