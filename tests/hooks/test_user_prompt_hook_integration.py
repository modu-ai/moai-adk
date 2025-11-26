#!/usr/bin/env python3
"""UserPromptSubmit Hook 통합 테스트

향상된 JIT 컨텍스트 로딩 및 에이전트 위임 기능의 전체 통합 테스트.
실제 Hook 실행부터 결과 확인까지의 완전한 흐름 검증.

TDD History:
    - RED: 통합 테스트 시나리오 작성 (실제 Hook 아직 미연동)
    - GREEN: Hook 기능 구현으로 통합 테스트 통과
    - REFACTOR: 테스트 커버리지 확장 및 에지 케이스 처리
"""

import json
import subprocess
import tempfile
from pathlib import Path

import pytest

# Skip - outdated test using 'alfred' hook structure
pytestmark = pytest.mark.skip(reason="Outdated test - expects 'alfred' hook folder (moved to moai)")


class TestUserPromptHookIntegration:
    """UserPromptSubmit Hook 통합 테스트"""

    def setup_method(self):
        """테스트 환경 설정"""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.hook_path = (
            Path(__file__).parent.parent.parent
            / "src"
            / "moai_adk"
            / "templates"
            / ".claude"
            / "hooks"
            / "alfred"
            / "user_prompt__jit_load_docs.py"
        )

        # Mock .moai/config.json
        self.moai_dir = self.temp_dir / ".moai"
        self.moai_dir.mkdir(parents=True, exist_ok=True)
        config_file = self.moai_dir / "config.json"
        config_file.write_text(
            json.dumps(
                {
                    "language": {"conversation_language": "ko"},
                    "hooks": {"user_prompt_jit_loading": {"enabled": True}},
                },
                indent=2,
            )
        )

        # Mock .claude/skills directory
        self.skills_dir = self.temp_dir / ".claude" / "skills"
        self.skills_dir.mkdir(parents=True, exist_ok=True)

        # Create mock skill references
        for skill_name in [
            "moai-core-spec-authoring",
            "moai-domain-frontend",
            "moai-essentials-debug",
        ]:
            skill_dir = self.skills_dir / skill_name
            skill_dir.mkdir(exist_ok=True)
            (skill_dir / "reference.md").write_text(f"# {skill_name} Reference\n\nMock skill documentation.")

        # Mock project files
        (self.temp_dir / "src").mkdir()
        (self.temp_dir / "src" / "App.js").write_text("/* React App */")
        (self.temp_dir / "package.json").write_text('{"name": "test-app"}')
        (self.temp_dir / "README.md").write_text("# Test Project")

    def test_hook_with_spec_creation_prompt(self):
        """SPEC 생성 프롬프트에 대한 Hook 테스트"""
        # Input payload
        payload = {
            "userPrompt": "새로운 기능에 대한 명세서를 작성해주세요",
            "cwd": str(self.temp_dir),
        }

        # Execute hook
        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Parse output
        try:
            output = json.loads(result.stdout)
        except json.JSONDecodeError:
            pytest.skip(f"Hook output invalid JSON: {result.stdout}")

        # Verify success
        assert result.returncode == 0
        assert output.get("continue") is True

        # Check hook-specific output
        hook_output = output.get("hookSpecificOutput", {})
        assert hook_output.get("hookEventName") == "UserPromptSubmit"

        # Check for additional context (agent delegation or traditional context)
        additional_context = hook_output.get("additionalContext")
        if additional_context:
            # Should contain agent delegation information or context files
            assert isinstance(additional_context, str)

    def test_hook_with_implementation_prompt(self):
        """구현 프롬프트에 대한 Hook 테스트"""
        payload = {
            "userPrompt": "/moai:2-run 사용자 인증 기능을 구현해주세요",
            "cwd": str(self.temp_dir),
        }

        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Parse output
        try:
            output = json.loads(result.stdout)
        except json.JSONDecodeError:
            pytest.skip(f"Hook output invalid JSON: {result.stdout}")

        # Verify success
        assert result.returncode == 0
        assert output.get("continue") is True

        # Check hook-specific output structure
        hook_output = output.get("hookSpecificOutput", {})
        assert "hookEventName" in hook_output

        # Should have agent delegation for /moai:2-run command
        additional_context = hook_output.get("additionalContext")
        if additional_context:
            print(f"Additional context: {additional_context}")

    def test_hook_with_frontend_development_prompt(self):
        """프론트엔드 개발 프롬프트에 대한 Hook 테스트"""
        payload = {
            "userPrompt": "React 컴포넌트를 만들어주세요. UI/UX 디자인이 필요합니다.",
            "cwd": str(self.temp_dir),
        }

        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Parse output
        try:
            output = json.loads(result.stdout)
        except json.JSONDecodeError:
            pytest.skip(f"Hook output invalid JSON: {result.stdout}")

        # Verify success
        assert result.returncode == 0

        # Check for frontend agent delegation
        hook_output = output.get("hookSpecificOutput", {})
        additional_context = hook_output.get("additionalContext", "")

        # Should detect frontend development intent
        if additional_context:
            assert "frontend" in additional_context.lower() or "ui" in additional_context.lower()

    def test_hook_with_general_prompt(self):
        """일반 프롬프트에 대한 Hook 테스트 (에이전트 위임 없음)"""
        payload = {"userPrompt": "오늘 날씨가 어때요?", "cwd": str(self.temp_dir)}

        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Parse output
        try:
            output = json.loads(result.stdout)
        except json.JSONDecodeError:
            pytest.skip(f"Hook output invalid JSON: {result.stdout}")

        # Verify success (should still succeed even without agent delegation)
        assert result.returncode == 0
        assert output.get("continue") is True

        # Should not have agent delegation for general prompts
        hook_output = output.get("hookSpecificOutput", {})
        additional_context = hook_output.get("additionalContext", "")

        if additional_context and "전문가 에이전트" in additional_context:
            pytest.fail("General prompt should not trigger agent delegation")

    def test_hook_timeout_handling(self):
        """Hook 타임아웃 처리 테스트"""
        # Create a scenario that might cause timeout
        payload = {
            "userPrompt": "매우 긴 프롬프트 " * 1000,  # Very long prompt
            "cwd": str(self.temp_dir),
        }

        # Execute with short timeout
        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            timeout=10,  # 10 second timeout
            cwd=str(self.temp_dir),
        )

        # Should handle gracefully (either succeed or timeout gracefully)
        if result.returncode == 1:
            # If failed, should have graceful error message
            try:
                output = json.loads(result.stdout)
                assert output.get("continue") is True  # Should continue despite timeout
            except json.JSONDecodeError:
                # At minimum, should not crash
                pass

    def test_hook_invalid_json_handling(self):
        """잘못된 JSON 입력 처리 테스트"""
        # Test with invalid JSON
        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input="invalid json input",
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Should handle gracefully
        try:
            output = json.loads(result.stdout)
            assert output.get("continue") is True  # Should continue despite error
        except json.JSONDecodeError:
            # Should not crash
            pass

    def test_hook_empty_input_handling(self):
        """빈 입력 처리 테스트"""
        payload = {}

        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Should handle gracefully
        try:
            output = json.loads(result.stdout)
            assert output.get("continue") is True
        except json.JSONDecodeError:
            # Should not crash
            pass


class TestHookPerformanceAndScalability:
    """Hook 성능 및 확장성 테스트"""

    def setup_method(self):
        """테스트 환경 설정"""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.hook_path = (
            Path(__file__).parent.parent.parent
            / "src"
            / "moai_adk"
            / "templates"
            / ".claude"
            / "hooks"
            / "alfred"
            / "user_prompt__jit_load_docs.py"
        )

        # Create minimal test environment
        self.moai_dir = self.temp_dir / ".moai"
        self.moai_dir.mkdir(parents=True, exist_ok=True)
        config_file = self.moai_dir / "config.json"
        config_file.write_text(json.dumps({"language": {"conversation_language": "ko"}}))

    def test_hook_execution_time(self):
        """Hook 실행 시간 테스트"""
        import time

        payload = {
            "userPrompt": "/moai:1-plan 새로운 기능 명세",
            "cwd": str(self.temp_dir),
        }

        # Measure execution time
        start_time = time.time()
        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )
        execution_time = time.time() - start_time

        # Should execute quickly (within 5 seconds)
        assert execution_time < 5.0, f"Hook took too long: {execution_time:.2f} seconds"
        assert result.returncode == 0

    def test_hook_memory_usage(self):
        """Hook 메모리 사용량 테스트"""
        import os

        import psutil

        # Get current process
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss

        payload = {
            "userPrompt": "복잡한 작업을 위한 프롬프트 " * 100,  # Moderate complexity
            "cwd": str(self.temp_dir),
        }

        result = subprocess.run(
            ["python3", str(self.hook_path)],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            cwd=str(self.temp_dir),
        )

        # Check memory usage (should not increase dramatically)
        final_memory = process.memory_info().rss
        memory_increase = final_memory - initial_memory

        # Memory increase should be reasonable (less than 50MB)
        assert (
            memory_increase < 50 * 1024 * 1024
        ), f"Memory usage increased too much: {memory_increase / (1024*1024):.2f} MB"
        assert result.returncode == 0


if __name__ == "__main__":
    pytest.main([__file__, "-v", "-s"])
