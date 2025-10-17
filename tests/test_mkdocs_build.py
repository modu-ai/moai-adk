# @TEST:DOCS-003-BUILD-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-BUILD-001
# Related: @REQ:DOCS-003-MKDOCS-001
"""
TEST-DOCS-003-BUILD-001: MkDocs 빌드 검증 테스트

이 테스트는 MkDocs가 오류 없이 빌드되는지 검증합니다.

테스트 대상:
- mkdocs build --strict 성공 여부
- 빌드된 사이트 구조 검증
- 필수 페이지 생성 확인
"""

import pytest
import subprocess
from pathlib import Path


# TEST-HAPPY-BUILD-001: 정상 동작 - MkDocs 빌드 성공
class TestMkDocsBuild:
    """MkDocs 빌드 검증 테스트 스위트"""

    @pytest.fixture
    def project_root(self) -> Path:
        """프로젝트 루트 디렉토리 경로"""
        return Path(__file__).parent.parent

    @pytest.fixture
    def site_dir(self, project_root: Path) -> Path:
        """빌드된 사이트 디렉토리 경로"""
        return project_root / "site"

    def test_should_build_successfully_with_strict_mode(self, project_root: Path):
        """TEST-BUILD-001: mkdocs build --strict가 성공해야 함"""
        # mkdocs build --strict 실행
        result = subprocess.run(
            ["mkdocs", "build", "--strict"],
            cwd=project_root,
            capture_output=True,
            text=True
        )

        # 빌드 성공 확인
        if result.returncode != 0:
            error_msg = f"\nMkDocs 빌드 실패:\n"
            error_msg += f"Return Code: {result.returncode}\n"
            error_msg += f"STDOUT:\n{result.stdout}\n"
            error_msg += f"STDERR:\n{result.stderr}\n"
            pytest.fail(error_msg)

    def test_should_generate_index_page(self, site_dir: Path):
        """TEST-BUILD-002: index.html이 생성되어야 함"""
        index_file = site_dir / "index.html"
        assert index_file.exists(), "index.html이 생성되지 않았습니다"

    def test_should_generate_all_section_pages(self, site_dir: Path):
        """TEST-BUILD-003: 모든 주요 섹션 페이지가 생성되어야 함"""
        required_pages = [
            "introduction/index.html",
            "getting-started/installation/index.html",
            "configuration/config-json/index.html",
            "workflow/index.html",
            "commands/cli-reference/index.html",
            "agents/spec-builder/index.html",
            "hooks/overview/index.html",
            "api-reference/core-installer/index.html",
            "contributing/overview/index.html",
            "security/overview/index.html",
            "troubleshooting/common-errors/index.html",
        ]

        missing_pages = []
        for page in required_pages:
            page_path = site_dir / page
            if not page_path.exists():
                missing_pages.append(page)

        if missing_pages:
            error_msg = "\n다음 페이지가 생성되지 않았습니다:\n"
            for page in missing_pages:
                error_msg += f"  - {page}\n"
            pytest.fail(error_msg)


# TEST-ERROR-BUILD-001: 오류 상황 - 빌드 경고 감지
class TestMkDocsBuildWarnings:
    """MkDocs 빌드 경고 검증"""

    @pytest.fixture
    def project_root(self) -> Path:
        """프로젝트 루트 디렉토리 경로"""
        return Path(__file__).parent.parent

    def test_should_not_have_build_warnings(self, project_root: Path):
        """TEST-ERROR-001: 빌드 경고가 없어야 함 (--strict 모드)"""
        # mkdocs build --strict는 경고를 오류로 처리
        result = subprocess.run(
            ["mkdocs", "build", "--strict"],
            cwd=project_root,
            capture_output=True,
            text=True
        )

        # 경고 문구 검사
        warning_keywords = ["WARNING:", "warning:", "WARN:"]
        warnings_found = []

        for keyword in warning_keywords:
            if keyword in result.stdout or keyword in result.stderr:
                warnings_found.append({
                    'keyword': keyword,
                    'stdout': result.stdout,
                    'stderr': result.stderr
                })

        if warnings_found:
            error_msg = "\n빌드 중 경고가 발견되었습니다:\n"
            for warning in warnings_found:
                error_msg += f"  - 키워드: {warning['keyword']}\n"
                if warning['stdout']:
                    error_msg += f"    STDOUT: {warning['stdout']}\n"
                if warning['stderr']:
                    error_msg += f"    STDERR: {warning['stderr']}\n"
            pytest.fail(error_msg)
