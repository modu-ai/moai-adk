# @TEST:DOCS-003-STRUCTURE-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-STRUCTURE-001
# Related: @REQ:DOCS-003-STRUCTURE-001, @REQ:DOCS-003-NAV-001

"""
SPEC-DOCS-003: 문서 구조 및 네비게이션 검증 테스트

이 테스트는 11단계 사용자 여정 기반 문서 구조를 검증합니다.
"""

from pathlib import Path

import pytest
import yaml


class TestDocumentStructure:
    """문서 구조 테스트"""

    @pytest.fixture
    def docs_dir(self):
        """문서 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    @pytest.fixture
    def mkdocs_config(self):
        """MkDocs 설정 파일 로드"""
        config_path = Path(__file__).parent.parent / "mkdocs.yml"
        with open(config_path, 'r', encoding='utf-8') as f:
            # Skip Python object tags (pymdownx.superfences.fence_code_format)
            # to avoid ConstructorError in CI/CD environments
            return yaml.safe_load(f.read().replace('!!python/name:', '#!!python/name:'))

    def test_introduction_exists(self, docs_dir):
        """TEST-INTRO-001: Introduction 파일 존재 확인"""
        intro_file = docs_dir / "introduction.md"
        assert intro_file.exists(), "introduction.md 파일이 존재해야 합니다"

    def test_getting_started_structure(self, docs_dir):
        """TEST-START-001: Getting Started 구조 확인 (3개 파일)"""
        getting_started = docs_dir / "getting-started"
        required_files = [
            "installation.md",
            "quick-start.md",
            "first-project.md"
        ]

        for file_name in required_files:
            file_path = getting_started / file_name
            assert file_path.exists(), f"getting-started/{file_name}이 존재해야 합니다"

    def test_configuration_structure(self, docs_dir):
        """TEST-CONFIG-001: Configuration 구조 확인 (3개 파일)"""
        configuration = docs_dir / "configuration"
        required_files = [
            "config-json.md",
            "personal-vs-team.md",
            "advanced-settings.md"
        ]

        for file_name in required_files:
            file_path = configuration / file_name
            assert file_path.exists(), f"configuration/{file_name}이 존재해야 합니다"

    def test_workflow_exists(self, docs_dir):
        """TEST-WORKFLOW-001: Workflow 파일 존재 확인 (기존 유지)"""
        workflow_file = docs_dir / "workflow.md"
        assert workflow_file.exists(), "workflow.md 파일이 존재해야 합니다"

    def test_commands_structure(self, docs_dir):
        """TEST-CMD-001: Commands 구조 확인 (3개 파일)"""
        commands = docs_dir / "commands"
        required_files = [
            "cli-reference.md",
            "alfred-commands.md",
            "agent-commands.md"
        ]

        for file_name in required_files:
            file_path = commands / file_name
            assert file_path.exists(), f"commands/{file_name}이 존재해야 합니다"

    def test_agents_structure(self, docs_dir):
        """TEST-AGENT-001: Agents 구조 확인 (9개 파일)"""
        agents = docs_dir / "agents"
        required_files = [
            "spec-builder.md",
            "code-builder.md",
            "doc-syncer.md",
            "tag-agent.md",
            "git-manager.md",
            "debug-helper.md",
            "trust-checker.md",
            "cc-manager.md",
            "project-manager.md"
        ]

        for file_name in required_files:
            file_path = agents / file_name
            assert file_path.exists(), f"agents/{file_name}이 존재해야 합니다"

    def test_hooks_structure(self, docs_dir):
        """TEST-HOOK-001: Hooks 구조 확인 (5개 파일)"""
        hooks = docs_dir / "hooks"
        required_files = [
            "overview.md",
            "session-start-hook.md",
            "pre-tool-use-hook.md",
            "post-tool-use-hook.md",
            "custom-hooks.md"
        ]

        for file_name in required_files:
            file_path = hooks / file_name
            assert file_path.exists(), f"hooks/{file_name}이 존재해야 합니다"

    def test_api_reference_structure(self, docs_dir):
        """TEST-API-001: API Reference 구조 확인 (5개 파일)"""
        api_reference = docs_dir / "api-reference"
        required_files = [
            "core-installer.md",
            "core-git.md",
            "core-tag.md",
            "core-template.md",
            "agents.md"
        ]

        for file_name in required_files:
            file_path = api_reference / file_name
            assert file_path.exists(), f"api-reference/{file_name}이 존재해야 합니다"

    def test_contributing_structure(self, docs_dir):
        """TEST-CONTRIB-001: Contributing 구조 확인 (5개 파일)"""
        contributing = docs_dir / "contributing"
        required_files = [
            "overview.md",
            "development-setup.md",
            "code-style.md",
            "testing.md",
            "pull-request-process.md"
        ]

        for file_name in required_files:
            file_path = contributing / file_name
            assert file_path.exists(), f"contributing/{file_name}이 존재해야 합니다"

    def test_security_structure(self, docs_dir):
        """TEST-SEC-001: Security 구조 확인 (4개 파일)"""
        security = docs_dir / "security"
        required_files = [
            "overview.md",
            "template-security.md",
            "best-practices.md",
            "checklist.md"
        ]

        for file_name in required_files:
            file_path = security / file_name
            assert file_path.exists(), f"security/{file_name}이 존재해야 합니다"

    def test_troubleshooting_structure(self, docs_dir):
        """TEST-TROUBLESHOOT-001: Troubleshooting 구조 확인 (3개 파일)"""
        troubleshooting = docs_dir / "troubleshooting"
        required_files = [
            "common-errors.md",
            "debugging-guide.md",
            "faq.md"
        ]

        for file_name in required_files:
            file_path = troubleshooting / file_name
            assert file_path.exists(), f"troubleshooting/{file_name}이 존재해야 합니다"

    def test_navigation_completeness(self, mkdocs_config):
        """TEST-NAV-001: mkdocs.yml 네비게이션 완전성 확인"""
        nav = mkdocs_config.get("nav", [])

        # 11단계 구조의 모든 섹션 확인
        nav_sections = []
        for item in nav:
            if isinstance(item, dict):
                nav_sections.extend(item.keys())

        # 필수 섹션 확인
        required_sections = [
            "Introduction",
            "Getting Started",
            "Configuration",
            "Workflow",
            "Commands",
            "Agents",
            "Hooks",
            "API Reference",
            "Contributing",
            "Security",
            "Troubleshooting"
        ]

        for section in required_sections:
            assert section in str(nav), f"네비게이션에 {section} 섹션이 포함되어야 합니다"
