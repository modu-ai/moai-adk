# @TEST:DOCS-003-API-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-API-001
# Related: @REQ:DOCS-003-API-001, @REQ:DOCS-003-API-USAGE-001
"""
TEST-DOCS-003-API-001: API 문서 자동 생성 검증 테스트

이 테스트는 mkdocstrings를 통한 API 문서 자동 생성이 정상 작동하는지 검증합니다.

테스트 대상:
- API 참조 문서 파일 존재 여부
- mkdocstrings 마크다운 문법 유효성
- Python docstring 파싱 가능 여부
"""

import pytest
import re
from pathlib import Path
from typing import List


# TEST-HAPPY-API-001: 정상 동작 - API 문서 자동 생성 확인
class TestApiDocs:
    """API 문서 자동 생성 검증 테스트 스위트"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    @pytest.fixture
    def api_docs_dir(self, docs_root: Path) -> Path:
        """API 문서 디렉토리 경로"""
        return docs_root / "api-reference"

    def extract_mkdocstrings_references(self, file_path: Path) -> List[str]:
        """mkdocstrings 참조 추출

        ::: moai_adk.core.installer 형식의 참조를 추출합니다.

        Returns:
            List[str]: mkdocstrings 참조 목록
        """
        content = file_path.read_text(encoding="utf-8")

        # ::: module.path 형식의 참조 추출
        pattern = r':::\s+([a-zA-Z0-9_.]+)'
        matches = re.findall(pattern, content)

        return matches

    def test_should_have_api_reference_files(self, api_docs_dir: Path):
        """TEST-API-001: API 참조 파일이 존재해야 함"""
        required_files = [
            "core-installer.md",
            "core-git.md",
            "core-tag.md",
            "core-template.md",
            "agents.md",
        ]

        for file in required_files:
            file_path = api_docs_dir / file
            assert file_path.exists(), f"{file}이 api-reference/에 존재하지 않습니다"

    def test_should_use_mkdocstrings_syntax(self, api_docs_dir: Path):
        """TEST-API-002: API 문서는 mkdocstrings 문법을 사용해야 함"""
        files_without_mkdocstrings = []

        # 예외 파일: 실제 Python 모듈이 아닌 문서
        exceptions = {
            "agents.md",  # Claude Code 에이전트 설정 (Python 모듈 아님)
            "core-tag.md",  # 주석 기반 TAG 시스템 (Python 모듈 아님)
        }

        for md_file in api_docs_dir.glob("*.md"):
            # 예외 파일 스킵
            if md_file.name in exceptions:
                continue

            references = self.extract_mkdocstrings_references(md_file)

            # mkdocstrings 참조가 없는 파일 기록
            if not references:
                files_without_mkdocstrings.append(md_file.name)

        if files_without_mkdocstrings:
            error_msg = "\n다음 API 문서 파일에 mkdocstrings 참조가 없습니다:\n"
            for file in files_without_mkdocstrings:
                error_msg += f"  - {file}\n"
            error_msg += "\n예상 형식: ::: moai_adk.module.path\n"
            pytest.fail(error_msg)

    def test_should_reference_valid_modules(self, api_docs_dir: Path):
        """TEST-API-003: mkdocstrings는 유효한 Python 모듈을 참조해야 함"""
        invalid_references = []

        for md_file in api_docs_dir.glob("*.md"):
            references = self.extract_mkdocstrings_references(md_file)

            for ref in references:
                # 모든 참조는 moai_adk로 시작해야 함
                if not ref.startswith('moai_adk'):
                    invalid_references.append({
                        'file': md_file.name,
                        'reference': ref,
                        'reason': 'moai_adk로 시작하지 않음'
                    })

        if invalid_references:
            error_msg = "\n유효하지 않은 모듈 참조가 발견되었습니다:\n"
            for inv_ref in invalid_references:
                error_msg += f"  - {inv_ref['file']}: {inv_ref['reference']}\n"
                error_msg += f"    이유: {inv_ref['reason']}\n"
            pytest.fail(error_msg)


# TEST-EDGE-API-001: 경계 조건 - API 문서 품질 검증
class TestApiDocsQuality:
    """API 문서 품질 검증"""

    @pytest.fixture
    def api_docs_dir(self) -> Path:
        """API 문서 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs" / "api-reference"

    def test_should_have_descriptive_content(self, api_docs_dir: Path):
        """TEST-EDGE-001: API 문서는 설명 콘텐츠를 포함해야 함"""
        # mkdocstrings 참조만 있는 것이 아니라 설명도 포함되어야 함
        files_without_description = []

        for md_file in api_docs_dir.glob("*.md"):
            content = md_file.read_text(encoding="utf-8")

            # 헤더와 mkdocstrings 참조를 제외한 실제 콘텐츠 확인
            lines = content.split('\n')
            content_lines = [
                line for line in lines
                if line.strip() and not line.startswith('#') and not line.strip().startswith(':::')
            ]

            # 최소 3줄 이상의 설명이 있어야 함
            if len(content_lines) < 3:
                files_without_description.append(md_file.name)

        if files_without_description:
            error_msg = "\n다음 API 문서에 충분한 설명이 없습니다 (최소 3줄 이상 필요):\n"
            for file in files_without_description:
                error_msg += f"  - {file}\n"
            pytest.fail(error_msg)
