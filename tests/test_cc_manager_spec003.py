"""
SPEC-003: cc-manager 중심 Claude Code 최적화 - RED 단계 테스트

TDD RED 단계: 실패하는 테스트 작성
모든 테스트는 현재 상태에서 실패해야 하며, 이는 의도된 동작입니다.
"""

import os
import re
import json
import yaml
from pathlib import Path
import pytest
from typing import Dict, List, Any


@pytest.fixture
def project_root():
    """프로젝트 루트 디렉토리 경로를 반환합니다."""
    return Path(__file__).parent.parent


@pytest.fixture
def claude_dir(project_root):
    """Claude Code 디렉토리 경로를 반환합니다."""
    return project_root / ".claude"


@pytest.fixture
def cc_manager_path(claude_dir):
    """cc-manager.md 파일 경로를 반환합니다."""
    return claude_dir / "agents" / "moai" / "cc-manager.md"


@pytest.fixture
def validate_script_path(project_root):
    """검증 스크립트 경로를 반환합니다."""
    return project_root / ".moai" / "scripts" / "validate_claude_standards.py"


class TestCCManagerTemplateGuidelines:
    """
    AC1: cc-manager 템플릿 지침 내장 테스트
    cc-manager.md에 완전한 템플릿 지침이 포함되어야 함
    """

    def test_cc_manager_contains_command_template_guidelines(self, cc_manager_path):
        """
        Given: cc-manager.md 파일이 존재할 때
        When: 파일 내용을 검사하면
        Then: 커맨드 표준 템플릿 지침이 완전히 포함되어 있어야 함
        """
        assert cc_manager_path.exists(), "cc-manager.md 파일이 존재하지 않습니다"

        content = cc_manager_path.read_text(encoding="utf-8")

        # 현재 실패할 검증들 (SPEC-003 완료 후 통과)
        required_command_sections = [
            "커맨드 표준 템플릿 지침",  # 완전한 지침 제목
            "Claude Code 공식 문서 통합",  # 외부 참조 제거
            "파일 생성 시 자동 검증",  # 자동화 기능
            "표준 위반 시 수정 제안",  # 오류 방지 기능
        ]

        for section in required_command_sections:
            assert section in content, (
                f"필수 섹션 '{section}'이 cc-manager.md에 없습니다 (의도된 실패)"
            )

    def test_cc_manager_contains_agent_template_guidelines(self, cc_manager_path):
        """
        Given: cc-manager.md 파일이 존재할 때
        When: 파일 내용을 검사하면
        Then: 에이전트 표준 템플릿 지침이 완전히 포함되어 있어야 함
        """
        assert cc_manager_path.exists()

        content = cc_manager_path.read_text(encoding="utf-8")

        # 현재 실패할 검증들 (SPEC-003 완료 후 통과)
        required_agent_sections = [
            "에이전트 표준 템플릿 지침",
            "프로액티브 트리거 조건 완전 가이드",
            "도구 권한 최소화 자동 검증",
            "중구난방 지침 방지 시스템",  # SPEC에서 강조한 부분
        ]

        for section in required_agent_sections:
            assert section in content, (
                f"필수 섹션 '{section}'이 cc-manager.md에 없습니다 (의도된 실패)"
            )

    def test_cc_manager_central_control_tower_role(self, cc_manager_path):
        """
        Given: cc-manager.md 파일이 존재할 때
        When: 파일 내용을 검사하면
        Then: 중앙 관제탑 역할이 명확하게 정의되어 있어야 함
        """
        assert cc_manager_path.exists()

        content = cc_manager_path.read_text(encoding="utf-8")

        # 현재 실패할 검증들
        required_control_tower_features = [
            "중앙 관제탑으로서의 완전한 표준 제공",
            "외부 문서 참조 없는 독립적 지침",
            "모든 Claude Code 파일 생성/수정 관리",
            "실시간 표준 검증 및 수정 제안",
        ]

        for feature in required_control_tower_features:
            assert feature in content, (
                f"중앙 관제탑 기능 '{feature}'이 정의되지 않았습니다 (의도된 실패)"
            )


class TestClaudeCodeStandardCompliance:
    """
    AC2: Claude Code 표준 준수 테스트
    기존 커맨드/에이전트 파일들이 공식 구조를 준수해야 함
    """

    def test_command_files_have_standard_yaml_frontmatter(self, claude_dir):
        """
        Given: .claude/commands/moai/*.md 파일들이 존재할 때
        When: YAML frontmatter를 검사하면
        Then: 표준 구조(name, description, argument-hint, allowed-tools, model)를 가져야 함
        """
        commands_dir = claude_dir / "commands" / "moai"
        command_files = list(commands_dir.glob("*.md"))

        assert len(command_files) >= 5, (
            f"커맨드 파일이 5개 미만입니다: {len(command_files)}"
        )

        required_fields = [
            "name",
            "description",
            "argument-hint",
            "allowed-tools",
            "model",
        ]

        for cmd_file in command_files:
            content = cmd_file.read_text(encoding="utf-8")

            # YAML frontmatter 추출
            yaml_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
            assert yaml_match, f"{cmd_file.name}에 YAML frontmatter가 없습니다"

            try:
                yaml_data = yaml.safe_load(yaml_match.group(1))
            except yaml.YAMLError as e:
                pytest.fail(f"{cmd_file.name}의 YAML 파싱 실패: {e}")

            # 현재 실패할 검증들 (일부 필드가 누락되어 있음)
            for field in required_fields:
                assert field in yaml_data, (
                    f"{cmd_file.name}에 필수 필드 '{field}'가 없습니다 (의도된 실패)"
                )

            # argument-hint가 배열인지 확인 (현재 실패할 가능성)
            if "argument-hint" in yaml_data:
                assert isinstance(yaml_data["argument-hint"], (list, str)), (
                    f"{cmd_file.name}의 argument-hint가 올바른 형식이 아닙니다 (의도된 실패)"
                )

    def test_agent_files_have_proactive_pattern(self, claude_dir):
        """
        Given: .claude/agents/moai/*.md 파일들이 존재할 때
        When: YAML frontmatter를 검사하면
        Then: description에 "Use PROACTIVELY for" 패턴이 포함되어야 함
        """
        agents_dir = claude_dir / "agents" / "moai"
        agent_files = list(agents_dir.glob("*.md"))

        assert len(agent_files) >= 7, (
            f"에이전트 파일이 7개 미만입니다: {len(agent_files)}"
        )

        for agent_file in agent_files:
            content = agent_file.read_text(encoding="utf-8")

            yaml_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
            assert yaml_match, f"{agent_file.name}에 YAML frontmatter가 없습니다"

            try:
                yaml_data = yaml.safe_load(yaml_match.group(1))
            except yaml.YAMLError as e:
                pytest.fail(f"{agent_file.name}의 YAML 파싱 실패: {e}")

            # 현재 일부 파일에서 실패할 검증
            assert "description" in yaml_data, (
                f"{agent_file.name}에 description 필드가 없습니다"
            )

            description = yaml_data.get("description", "")
            # 현재 실패할 가능성이 높은 검증 (모든 파일이 패턴을 준수하지 않음)
            assert "Use PROACTIVELY for" in description, (
                f"{agent_file.name}의 description에 'Use PROACTIVELY for' 패턴이 없습니다 (의도된 실패)"
            )

    def test_agent_files_have_minimal_tool_permissions(self, claude_dir):
        """
        Given: .claude/agents/moai/*.md 파일들이 존재할 때
        When: tools 필드를 검사하면
        Then: 최소 권한 원칙에 따른 도구만 허용되어야 함
        """
        agents_dir = claude_dir / "agents" / "moai"
        agent_files = list(agents_dir.glob("*.md"))

        # 허용된 표준 도구 목록 (SPEC-003 기준)
        allowed_tools = {
            "Read",
            "Write",
            "Edit",
            "MultiEdit",
            "Bash",
            "Glob",
            "Grep",
            "TodoWrite",
            "WebFetch",
            "WebSearch",
            "Task",
        }

        for agent_file in agent_files:
            content = agent_file.read_text(encoding="utf-8")

            yaml_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
            if yaml_match:
                yaml_data = yaml.safe_load(yaml_match.group(1))

                if "tools" in yaml_data:
                    tools = yaml_data["tools"]
                    if isinstance(tools, str):
                        tools_list = [tool.strip() for tool in tools.split(",")]
                    elif isinstance(tools, list):
                        tools_list = tools
                    else:
                        pytest.fail(
                            f"{agent_file.name}의 tools 필드 형식이 잘못되었습니다"
                        )

                    # 현재 실패할 가능성이 있는 검증 (일부 에이전트가 과도한 권한을 가질 수 있음)
                    for tool in tools_list:
                        assert tool in allowed_tools, (
                            f"{agent_file.name}에 허용되지 않은 도구 '{tool}'이 있습니다 (의도된 실패)"
                        )


class TestValidationToolExistence:
    """
    AC3: 검증 도구 동작 테스트
    validate_claude_standards.py 스크립트가 존재하고 동작해야 함
    """

    def test_validation_script_exists(self, validate_script_path):
        """
        Given: validate_claude_standards.py 스크립트가 있어야 할 때
        When: 파일 존재 여부를 확인하면
        Then: 파일이 존재해야 함
        """
        # 현재 실패할 검증 (스크립트가 아직 없음)
        assert validate_script_path.exists(), (
            f"검증 스크립트가 존재하지 않습니다: {validate_script_path} (의도된 실패)"
        )

    def test_validation_script_has_required_functions(self, validate_script_path):
        """
        Given: validate_claude_standards.py 스크립트가 존재할 때
        When: 스크립트 내용을 검사하면
        Then: 필수 검증 함수들이 구현되어 있어야 함
        """
        if not validate_script_path.exists():
            pytest.skip("검증 스크립트가 없으므로 스킵")

        content = validate_script_path.read_text(encoding="utf-8")

        # 현재 실패할 검증들 (함수가 구현되지 않음)
        required_functions = [
            "validate_yaml_frontmatter",
            "check_required_fields",
            "validate_proactive_pattern",
            "generate_violation_report",
            "suggest_fixes",
        ]

        for func_name in required_functions:
            assert f"def {func_name}" in content, (
                f"필수 함수 '{func_name}'이 구현되지 않았습니다 (의도된 실패)"
            )


class TestCoreDocumentOptimization:
    """
    AC4: 핵심 문서 최적화 테스트
    CLAUDE.md, settings.json 등이 cc-manager 중심으로 최적화되어야 함
    """

    def test_claude_md_emphasizes_cc_manager_role(self, project_root):
        """
        Given: CLAUDE.md 파일이 업데이트될 때
        When: 파일 내용을 확인하면
        Then: cc-manager의 중앙 관제탑 역할이 강조되어야 함
        """
        claude_md_path = project_root / "CLAUDE.md"
        assert claude_md_path.exists(), "CLAUDE.md 파일이 존재하지 않습니다"

        content = claude_md_path.read_text(encoding="utf-8")

        # 현재 실패할 검증들 (cc-manager 역할이 충분히 강조되지 않음)
        required_cc_manager_mentions = [
            "cc-manager 중앙 관제탑",
            "Claude Code 표준화의 핵심",
            "모든 파일 생성/수정의 중심",
            "표준 검증 자동화",
        ]

        for mention in required_cc_manager_mentions:
            assert mention in content, (
                f"CLAUDE.md에 cc-manager 역할 설명 '{mention}'이 없습니다 (의도된 실패)"
            )

    def test_settings_json_has_optimized_permissions(self, claude_dir):
        """
        Given: .claude/settings.json 파일이 업데이트될 때
        When: permissions 섹션을 확인하면
        Then: WebSearch, BashOutput 등 추가 도구들이 허용되어야 함
        """
        settings_path = claude_dir / "settings.json"

        if not settings_path.exists():
            pytest.skip("settings.json이 없으므로 스킵")

        try:
            with open(settings_path, "r", encoding="utf-8") as f:
                settings = json.load(f)
        except json.JSONDecodeError as e:
            pytest.fail(f"settings.json 파싱 실패: {e}")

        # 현재 실패할 검증들 (새로운 권한이 추가되지 않음)
        required_new_permissions = [
            "WebSearch",
            "BashOutput",
            "KillShell",
            "Bash(gemini:*)",
            "Bash(codex:*)",
        ]

        permissions = settings.get("permissions", {})
        allowed_tools = permissions.get("allow", [])

        for perm in required_new_permissions:
            assert perm in allowed_tools, (
                f"settings.json에 필수 권한 '{perm}'이 없습니다 (의도된 실패)"
            )


class TestStandardComplianceIntegration:
    """
    통합 테스트: 전체 표준 준수 시스템 동작 검증
    """

    def test_all_files_pass_standard_validation(self, claude_dir, validate_script_path):
        """
        Given: 검증 스크립트와 Claude Code 파일들이 존재할 때
        When: 전체 표준 검증을 실행하면
        Then: 모든 파일이 표준을 준수해야 함
        """
        if not validate_script_path.exists():
            pytest.skip("검증 스크립트가 없으므로 스킵")

        # 현재 실패할 검증 (표준을 완전히 준수하지 않는 파일들 존재)
        commands_dir = claude_dir / "commands" / "moai"
        agents_dir = claude_dir / "agents" / "moai"

        violation_count = 0

        # 커맨드 파일들 검증
        for cmd_file in commands_dir.glob("*.md"):
            # 임시 검증 로직 (실제로는 validate_script를 사용)
            content = cmd_file.read_text(encoding="utf-8")
            if not re.match(
                r"^---\n.*\nname:.*\ndescription:.*\nargument-hint:.*\nallowed-tools:.*\nmodel:.*\n---",
                content,
                re.DOTALL,
            ):
                violation_count += 1

        # 에이전트 파일들 검증
        for agent_file in agents_dir.glob("*.md"):
            content = agent_file.read_text(encoding="utf-8")
            if "Use PROACTIVELY for" not in content:
                violation_count += 1

        # 현재 실패할 검증 (위반사항이 존재함)
        assert violation_count == 0, (
            f"{violation_count}개 파일이 표준을 위반했습니다 (의도된 실패)"
        )

    def test_cc_manager_can_generate_standard_files(self, cc_manager_path):
        """
        Given: cc-manager가 강화된 템플릿 지침을 가질 때
        When: 새 파일 생성을 요청하면
        Then: 표준에 맞는 파일이 생성되어야 함
        """
        assert cc_manager_path.exists()

        content = cc_manager_path.read_text(encoding="utf-8")

        # 현재 실패할 검증 (파일 생성 기능이 완전하지 않음)
        required_generation_capabilities = [
            "자동 파일 생성 시 표준 템플릿 적용",
            "실시간 표준 검증 및 오류 방지",
            "기존 파일 수정 시 표준 준수 확인",
            "표준 위반 시 즉시 수정 제안",
        ]

        for capability in required_generation_capabilities:
            assert capability in content, (
                f"cc-manager에 파일 생성 기능 '{capability}'이 없습니다 (의도된 실패)"
            )


if __name__ == "__main__":
    # 이 테스트들은 모두 현재 상태에서 실패해야 합니다 (TDD RED 단계)
    print("SPEC-003 RED 단계 테스트 실행")
    print("모든 테스트는 의도적으로 실패해야 합니다.")
    pytest.main([__file__, "-v"])
