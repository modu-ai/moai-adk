# @TEST:DOCS-003-LINKS-001 | Chain: @SPEC:DOCS-003 -> @TEST:DOCS-003-LINKS-001
# Related: @REQ:DOCS-003-NAV-001
"""
TEST-DOCS-003-LINKS-001: 문서 내부 링크 유효성 검증 테스트

이 테스트는 문서 간 내부 링크가 올바르게 연결되어 있는지 검증합니다.

테스트 대상:
- 모든 마크다운 링크 ([text](path.md)) 유효성
- 앵커 링크 (#section) 존재 여부
- 깨진 링크 (404) 감지
"""

import re
from pathlib import Path
from typing import List, Tuple

import pytest


# TEST-HAPPY-LINKS-001: 정상 동작 - 내부 링크 유효성 검증
class TestDocsLinks:
    """문서 링크 유효성 검증 테스트 스위트"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def extract_markdown_links(self, file_path: Path) -> List[Tuple[str, str]]:
        """마크다운 파일에서 모든 링크 추출

        Returns:
            List[Tuple[str, str]]: [(링크 텍스트, 링크 경로), ...]
        """
        content = file_path.read_text(encoding="utf-8")

        # [text](path) 형식의 링크 추출
        pattern = r'\[([^\]]+)\]\(([^)]+)\)'
        matches = re.findall(pattern, content)

        return [(text, path) for text, path in matches]

    def test_should_have_valid_internal_links(self, docs_root: Path):
        """TEST-LINKS-001: 모든 내부 링크는 유효해야 함"""
        broken_links = []

        for md_file in docs_root.rglob("*.md"):
            links = self.extract_markdown_links(md_file)

            for link_text, link_path in links:
                # 외부 링크 스킵 (http, https, mailto)
                if link_path.startswith(('http://', 'https://', 'mailto:')):
                    continue

                # 앵커만 있는 링크 스킵 (같은 페이지 내 이동)
                if link_path.startswith('#'):
                    continue

                # 상대 경로 해결
                if '#' in link_path:
                    # 앵커가 포함된 링크 (#section)
                    file_part, anchor = link_path.split('#', 1)
                    if file_part:
                        target_path = (md_file.parent / file_part).resolve()
                    else:
                        # 현재 파일의 앵커
                        continue
                else:
                    target_path = (md_file.parent / link_path).resolve()

                # 링크 대상 파일 존재 확인
                if not target_path.exists():
                    broken_links.append({
                        'source': str(md_file.relative_to(docs_root)),
                        'link_text': link_text,
                        'target': link_path,
                        'resolved': str(target_path)
                    })

        if broken_links:
            error_msg = "\n깨진 링크가 발견되었습니다:\n"
            for link in broken_links:
                error_msg += f"  - {link['source']}: [{link['link_text']}]({link['target']})\n"
                error_msg += f"    → {link['resolved']} (존재하지 않음)\n"
            pytest.fail(error_msg)

    def test_should_have_cross_references(self, docs_root: Path):
        """TEST-LINKS-002: 주요 문서는 서로 크로스 링크를 가져야 함"""
        # Introduction은 Getting Started로 링크되어야 함
        intro_file = docs_root / "introduction.md"
        if intro_file.exists():
            links = self.extract_markdown_links(intro_file)
            link_paths = [path for _, path in links]

            # Getting Started 관련 링크가 있는지 확인
            has_getting_started_link = any(
                'getting-started' in path.lower() for path in link_paths
            )
            assert has_getting_started_link, \
                "introduction.md는 getting-started로 연결되는 링크를 포함해야 합니다"


# TEST-ERROR-LINKS-001: 오류 상황 - 절대 경로 링크 금지
class TestDocsLinkConstraints:
    """문서 링크 제약사항 검증"""

    @pytest.fixture
    def docs_root(self) -> Path:
        """문서 루트 디렉토리 경로"""
        return Path(__file__).parent.parent / "docs"

    def extract_markdown_links(self, file_path: Path) -> List[Tuple[str, str]]:
        """마크다운 파일에서 모든 링크 추출"""
        content = file_path.read_text(encoding="utf-8")
        pattern = r'\[([^\]]+)\]\(([^)]+)\)'
        matches = re.findall(pattern, content)
        return [(text, path) for text, path in matches]

    def test_should_not_use_absolute_paths(self, docs_root: Path):
        """TEST-ERROR-001: 문서 내부 링크는 절대 경로를 사용하지 않아야 함"""
        absolute_path_links = []

        for md_file in docs_root.rglob("*.md"):
            links = self.extract_markdown_links(md_file)

            for link_text, link_path in links:
                # 외부 링크 스킵
                if link_path.startswith(('http://', 'https://', 'mailto:')):
                    continue

                # 절대 경로 검사 (/ 또는 file:// 로 시작)
                if link_path.startswith('/') or link_path.startswith('file://'):
                    absolute_path_links.append({
                        'source': str(md_file.relative_to(docs_root)),
                        'link_text': link_text,
                        'link_path': link_path
                    })

        if absolute_path_links:
            error_msg = "\n절대 경로를 사용하는 링크가 발견되었습니다:\n"
            for link in absolute_path_links:
                error_msg += f"  - {link['source']}: [{link['link_text']}]({link['link_path']})\n"
            pytest.fail(error_msg)
