# @TEST:DOCS-003-CONTENT-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-CONTENT-001
# Related: @REQ:DOCS-003-INTRO-001, @REQ:DOCS-003-START-001

"""
SPEC-DOCS-003: 문서 내용 품질 검증 테스트

이 테스트는 문서 내용의 품질과 일관성을 검증합니다.
"""

import re
from pathlib import Path

import pytest


class TestDocumentContent:
    """문서 내용 테스트"""

    @pytest.fixture
    def docs_dir(self):
        """문서 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def test_introduction_has_core_problems(self, docs_dir):
        """TEST-INTRO-CONTENT-001: Introduction에 3가지 핵심 문제 포함 확인"""
        intro_file = docs_dir / "introduction.md"
        content = intro_file.read_text(encoding='utf-8')

        # 핵심 키워드 확인
        keywords = ["플랑켄슈타인", "추적성", "품질"]
        found_keywords = sum(1 for kw in keywords if kw in content)

        assert found_keywords >= 2, (
            f"Introduction에 핵심 문제 키워드가 충분히 포함되어야 합니다 (발견: {found_keywords}/3)"
        )

    def test_getting_started_has_commands(self, docs_dir):
        """TEST-START-CONTENT-001: Getting Started에 핵심 명령어 포함 확인"""
        installation = docs_dir / "getting-started" / "installation.md"
        quick_start = docs_dir / "getting-started" / "quick-start.md"

        # installation.md에 pip install 명령어 확인
        install_content = installation.read_text(encoding='utf-8')
        assert "pip install" in install_content or "poetry add" in install_content, \
            "installation.md에 설치 명령어가 포함되어야 합니다"

        # quick-start.md에 moai-adk init 명령어 확인
        quick_content = quick_start.read_text(encoding='utf-8')
        assert "moai-adk init" in quick_content or "init" in quick_content, \
            "quick-start.md에 초기화 명령어가 포함되어야 합니다"

    def test_agents_have_persona(self, docs_dir):
        """TEST-AGENT-CONTENT-001: 각 에이전트 문서에 페르소나 섹션 확인"""
        agents_dir = docs_dir / "agents"
        agent_files = [
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

        missing_persona = []
        for agent_file in agent_files:
            content = (agents_dir / agent_file).read_text(encoding='utf-8')
            # 페르소나 관련 키워드 확인 (유연하게)
            if not any(keyword in content.lower() for keyword in ["페르소나", "persona", "역할", "전문", "아이콘"]):
                missing_persona.append(agent_file)

        assert len(missing_persona) == 0, \
            f"페르소나 정보가 누락된 에이전트: {missing_persona}"

    def test_api_reference_has_docstrings(self, docs_dir):
        """TEST-API-CONTENT-001: API Reference에 mkdocstrings 지시자 확인"""
        api_dir = docs_dir / "api-reference"
        api_files = [
            "core-installer.md",
            "core-git.md",
            "core-tag.md",
            "core-template.md"
        ]

        missing_docstrings = []
        for api_file in api_files:
            content = (api_dir / api_file).read_text(encoding='utf-8')
            # mkdocstrings 지시자 확인
            if ":::" not in content:
                missing_docstrings.append(api_file)

        assert len(missing_docstrings) == 0, \
            f"mkdocstrings 지시자가 누락된 API 문서: {missing_docstrings}"

    def test_troubleshooting_has_sufficient_errors(self, docs_dir):
        """TEST-TROUBLESHOOT-CONTENT-001: Troubleshooting에 충분한 에러 문서화 확인"""
        common_errors = docs_dir / "troubleshooting" / "common-errors.md"
        content = common_errors.read_text(encoding='utf-8')

        # ### 로 시작하는 에러 섹션 개수 확인
        error_sections = len(re.findall(r'^###\s+\w+Error', content, re.MULTILINE))

        assert error_sections >= 5, \
            f"common-errors.md에 최소 5개 이상의 에러가 문서화되어야 합니다 (현재: {error_sections}개)"

    def test_troubleshooting_has_sufficient_faq(self, docs_dir):
        """TEST-FAQ-001: FAQ에 충분한 항목 확인"""
        faq = docs_dir / "troubleshooting" / "faq.md"
        content = faq.read_text(encoding='utf-8')

        # ### 로 시작하는 질문 섹션 개수 확인
        faq_sections = len(re.findall(r'^###\s+', content, re.MULTILINE))

        assert faq_sections >= 10, \
            f"faq.md에 최소 10개 이상의 FAQ가 포함되어야 합니다 (현재: {faq_sections}개)"

    def test_security_has_checklist(self, docs_dir):
        """TEST-SEC-CONTENT-001: Security 체크리스트 항목 확인"""
        checklist = docs_dir / "security" / "checklist.md"
        content = checklist.read_text(encoding='utf-8')

        # 체크박스 항목 개수 확인
        checkbox_items = len(re.findall(r'^\s*-\s+\[[ x]\]', content, re.MULTILINE))

        assert checkbox_items >= 5, \
            f"security/checklist.md에 최소 5개 이상의 체크리스트 항목이 필요합니다 (현재: {checkbox_items}개)"

    def test_contributing_has_setup_steps(self, docs_dir):
        """TEST-CONTRIB-CONTENT-001: Contributing에 개발 환경 설정 단계 확인"""
        dev_setup = docs_dir / "contributing" / "development-setup.md"
        content = dev_setup.read_text(encoding='utf-8')

        # 필수 도구 키워드 확인
        required_tools = ["poetry", "python", "git"]
        found_tools = sum(1 for tool in required_tools if tool.lower() in content.lower())

        assert found_tools >= 2, \
            f"development-setup.md에 필수 도구 설명이 충분히 포함되어야 합니다 (발견: {found_tools}/3)"
