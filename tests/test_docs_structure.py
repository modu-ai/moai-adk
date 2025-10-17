# @TEST:DOCS-003-STRUCTURE-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-STRUCTURE-001
# Related: @REQ:DOCS-003-STRUCTURE-001
"""
TEST-DOCS-003-STRUCTURE-001: 11단계 문서 구조 검증 테스트

이 테스트는 SPEC-DOCS-003에서 정의한 11단계 사용자 여정 기반 문서 구조가
올바르게 생성되었는지 검증합니다.

테스트 대상:
1. Introduction (docs/introduction.md)
2. Getting Started (docs/getting-started/)
3. Configuration (docs/configuration/)
4. Workflow (docs/workflow.md)
5. Commands (docs/commands/)
6. Agents (docs/agents/)
7. Hooks (docs/hooks/)
8. API Reference (docs/api-reference/)
9. Contributing (docs/contributing/)
10. Security (docs/security/)
11. Troubleshooting (docs/troubleshooting/)
"""

import pytest
from pathlib import Path


# TEST-HAPPY-STRUCTURE-001: 정상 동작 - 11단계 디렉토리 존재 확인
class TestDocsStructure:
    """문서 구조 검증 테스트 스위트"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def test_should_have_introduction_file(self, docs_root: Path):
        """TEST-STRUCTURE-001: Introduction 파일이 존재해야 함"""
        intro_file = docs_root / "introduction.md"
        assert intro_file.exists(), f"introduction.md가 {docs_root}에 존재하지 않습니다"

    def test_should_have_getting_started_directory(self, docs_root: Path):
        """TEST-STRUCTURE-002: Getting Started 디렉토리가 존재해야 함"""
        getting_started = docs_root / "getting-started"
        assert getting_started.exists(), "getting-started 디렉토리가 존재하지 않습니다"
        assert getting_started.is_dir(), "getting-started는 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = ["installation.md", "quick-start.md", "first-project.md"]
        for file in required_files:
            file_path = getting_started / file
            assert file_path.exists(), f"{file}이 getting-started/에 존재하지 않습니다"

    def test_should_have_configuration_directory(self, docs_root: Path):
        """TEST-STRUCTURE-003: Configuration 디렉토리가 존재해야 함"""
        config_dir = docs_root / "configuration"
        assert config_dir.exists(), "configuration 디렉토리가 존재하지 않습니다"
        assert config_dir.is_dir(), "configuration은 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = ["config-json.md", "personal-vs-team.md", "advanced-settings.md"]
        for file in required_files:
            file_path = config_dir / file
            assert file_path.exists(), f"{file}이 configuration/에 존재하지 않습니다"

    def test_should_have_workflow_file(self, docs_root: Path):
        """TEST-STRUCTURE-004: Workflow 파일이 존재해야 함"""
        workflow_file = docs_root / "workflow.md"
        assert workflow_file.exists(), f"workflow.md가 {docs_root}에 존재하지 않습니다"

    def test_should_have_commands_directory(self, docs_root: Path):
        """TEST-STRUCTURE-005: Commands 디렉토리가 존재해야 함"""
        commands_dir = docs_root / "commands"
        assert commands_dir.exists(), "commands 디렉토리가 존재하지 않습니다"
        assert commands_dir.is_dir(), "commands는 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = ["cli-reference.md", "alfred-commands.md", "agent-commands.md"]
        for file in required_files:
            file_path = commands_dir / file
            assert file_path.exists(), f"{file}이 commands/에 존재하지 않습니다"

    def test_should_have_agents_directory(self, docs_root: Path):
        """TEST-STRUCTURE-006: Agents 디렉토리가 존재해야 함"""
        agents_dir = docs_root / "agents"
        assert agents_dir.exists(), "agents 디렉토리가 존재하지 않습니다"
        assert agents_dir.is_dir(), "agents는 디렉토리여야 합니다"

        # 9개 에이전트 파일 검증
        required_agents = [
            "spec-builder.md",
            "code-builder.md",
            "doc-syncer.md",
            "tag-agent.md",
            "git-manager.md",
            "debug-helper.md",
            "trust-checker.md",
            "cc-manager.md",
            "project-manager.md",
        ]
        for agent_file in required_agents:
            file_path = agents_dir / agent_file
            assert file_path.exists(), f"{agent_file}이 agents/에 존재하지 않습니다"

    def test_should_have_hooks_directory(self, docs_root: Path):
        """TEST-STRUCTURE-007: Hooks 디렉토리가 존재해야 함"""
        hooks_dir = docs_root / "hooks"
        assert hooks_dir.exists(), "hooks 디렉토리가 존재하지 않습니다"
        assert hooks_dir.is_dir(), "hooks는 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = [
            "overview.md",
            "session-start-hook.md",
            "pre-tool-use-hook.md",
            "post-tool-use-hook.md",
            "custom-hooks.md",
        ]
        for file in required_files:
            file_path = hooks_dir / file
            assert file_path.exists(), f"{file}이 hooks/에 존재하지 않습니다"

    def test_should_have_api_reference_directory(self, docs_root: Path):
        """TEST-STRUCTURE-008: API Reference 디렉토리가 존재해야 함"""
        api_dir = docs_root / "api-reference"
        assert api_dir.exists(), "api-reference 디렉토리가 존재하지 않습니다"
        assert api_dir.is_dir(), "api-reference는 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = [
            "core-installer.md",
            "core-git.md",
            "core-tag.md",
            "core-template.md",
            "agents.md",
        ]
        for file in required_files:
            file_path = api_dir / file
            assert file_path.exists(), f"{file}이 api-reference/에 존재하지 않습니다"

    def test_should_have_contributing_directory(self, docs_root: Path):
        """TEST-STRUCTURE-009: Contributing 디렉토리가 존재해야 함"""
        contrib_dir = docs_root / "contributing"
        assert contrib_dir.exists(), "contributing 디렉토리가 존재하지 않습니다"
        assert contrib_dir.is_dir(), "contributing은 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = [
            "overview.md",
            "development-setup.md",
            "code-style.md",
            "testing.md",
            "pull-request-process.md",
        ]
        for file in required_files:
            file_path = contrib_dir / file
            assert file_path.exists(), f"{file}이 contributing/에 존재하지 않습니다"

    def test_should_have_security_directory(self, docs_root: Path):
        """TEST-STRUCTURE-010: Security 디렉토리가 존재해야 함"""
        security_dir = docs_root / "security"
        assert security_dir.exists(), "security 디렉토리가 존재하지 않습니다"
        assert security_dir.is_dir(), "security는 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = [
            "overview.md",
            "template-security.md",
            "best-practices.md",
            "checklist.md",
        ]
        for file in required_files:
            file_path = security_dir / file
            assert file_path.exists(), f"{file}이 security/에 존재하지 않습니다"

    def test_should_have_troubleshooting_directory(self, docs_root: Path):
        """TEST-STRUCTURE-011: Troubleshooting 디렉토리가 존재해야 함"""
        troubleshooting_dir = docs_root / "troubleshooting"
        assert troubleshooting_dir.exists(), "troubleshooting 디렉토리가 존재하지 않습니다"
        assert troubleshooting_dir.is_dir(), "troubleshooting은 디렉토리여야 합니다"

        # 필수 파일 검증
        required_files = ["common-errors.md", "debugging-guide.md", "faq.md"]
        for file in required_files:
            file_path = troubleshooting_dir / file
            assert file_path.exists(), f"{file}이 troubleshooting/에 존재하지 않습니다"


# TEST-EDGE-STRUCTURE-001: 경계 조건 - 파일 크기 검증
class TestDocsFileSizes:
    """문서 파일 크기 검증 (품질 제약)"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def test_should_have_reasonable_file_sizes(self, docs_root: Path):
        """TEST-EDGE-001: 문서 파일은 적절한 크기를 가져야 함 (< 50KB)"""
        # 너무 큰 문서는 여러 파일로 분리되어야 함
        max_file_size = 50 * 1024  # 50KB

        for md_file in docs_root.rglob("*.md"):
            file_size = md_file.stat().st_size
            assert (
                file_size < max_file_size
            ), f"{md_file.name}의 크기가 너무 큽니다 ({file_size / 1024:.1f}KB > 50KB)"


# TEST-ERROR-STRUCTURE-001: 오류 상황 - 금지된 파일 확인
class TestDocsConstraints:
    """문서 제약사항 검증"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def test_should_not_have_checkpoint_policies(self, docs_root: Path):
        """TEST-ERROR-001: checkpoint-policies.md 같은 과도한 세부 문서는 존재하지 않아야 함"""
        # CONS-DOCS-003-002: 과도하게 세부적인 문서 제외
        forbidden_files = ["checkpoint-policies.md", "advanced-optimization.md"]

        for forbidden_file in forbidden_files:
            for found_file in docs_root.rglob(forbidden_file):
                pytest.fail(f"금지된 파일이 발견되었습니다: {found_file}")
