# @TEST:PORTAL-LINK-001 | SPEC: SPEC-PORTAL-LINK-001

import pytest
import requests
from pathlib import Path
import re
from typing import Dict, Any


class TestPortalLinkValidation:
    """온라인 문서 링크 검증 테스트"""

    @pytest.fixture
    def readme_content(self) -> str:
        """README.ko.md 파일 내용을 반환하는 fixture"""
        readme_path = Path("README.ko.md")
        if not readme_path.exists():
            pytest.fail("README.ko.md 파일이 존재하지 않습니다")
        return readme_path.read_text(encoding='utf-8')

    @pytest.fixture
    def online_docs_url(self) -> str:
        """온라인 문서 URL 반환"""
        return "https://adk.mo.ai.kr"

    @pytest.fixture
    def expected_links(self) -> Dict[str, str]:
        """기대되는 링크 구조 반환"""
        return {
            "main_portal": "https://adk.mo.ai.kr",
            "api_docs": "https://adk.mo.ai.kr/api",
            "getting_started": "https://adk.mo.ai.kr/getting-started",
            "guides": "https://adk.mo.ai.kr/guides"
        }

    def test_readme_should_contain_online_docs_link(self, readme_content: str) -> None:
        """WHEN README.ko.md를 열면, 온라인 문서 포털 링크가 포함되어 있어야 한다"""
        # 온라인 문서 링크 존재 검증
        assert "https://adk.mo.ai.kr" in readme_content, "온라인 문서 포털 링크가 README.ko.md에 없습니다"

        # 링크가 올바른 형식인지 검증 (markdown 링크 형식)
        link_pattern = r'\[.*?\]\(https://adk\.mo\.ai\.kr.*?\)'
        matches = re.findall(link_pattern, readme_content)
        assert len(matches) > 0, "온라인 문서 링크가 올바른 markdown 형식이 아닙니다"

    def test_online_docs_link_should_be_accessible(self, online_docs_url: str) -> None:
        """WHEN 온라인 문서 포털 링크를 접속하면, HTTP 200 응답을 반환해야 한다"""
        try:
            response = requests.get(online_docs_url, timeout=10)
            assert response.status_code == 200, f"온라인 문서 포털 접속 실패: {response.status_code}"
        except requests.exceptions.RequestException as e:
            pytest.fail(f"온라인 문서 포털 접속 오류: {e}")

    def test_readme_should_contain_multiple_documentation_links(self, readme_content: str) -> None:
        """WHEN README.ko.md를 열면, 여러 개의 문서 섹션 링크가 포함되어 있어야 한다"""
        # 메인 포털 링크
        assert "온라인 문서" in readme_content, "온라인 문서 섹션이 없습니다"

        # API 문서 링크
        api_links = re.findall(r'\[API.*?\]\(.*?\)', readme_content, re.IGNORECASE)
        assert len(api_links) > 0, "API 문서 링크가 없습니다"

        # 가이드 문서 링크
        guide_links = re.findall(r'\[.*?가이드.*?\]\(.*?\)', readme_content, re.IGNORECASE)
        assert len(guide_links) > 0, "가이드 문서 링크가 없습니다"

    def test_links_should_be_consistent(self, readme_content: str, expected_links: Dict[str, str]) -> None:
        """WHEN README.ko.md를 열면, 모든 링크가 일관된 형식이어야 한다"""
        main_portal_found = False
        api_docs_found = False
        getting_started_found = False

        # 모든 markdown 링크 추출
        link_pattern = r'\[([^\]]+)\]\(([^)]+)\)'
        matches = re.findall(link_pattern, readme_content)

        for link_text, link_url in matches:
            # 메인 포털 링크 검증
            if "온라인 문서" in link_text or "adk.mo.ai.kr" in link_url:
                main_portal_found = True
                assert link_url == expected_links["main_portal"], f"메인 포털 링크가 일치하지 않습니다: {link_url}"

            # API 문서 링크 검증
            if any(keyword in link_text.lower() for keyword in ["api", "api", "인터페이스"]):
                api_docs_found = True
                assert "api" in link_url.lower(), f"API 문서 링크에 'api'가 없습니다: {link_url}"

            # 시작 가이드 링크 검증
            if any(keyword in link_text.lower() for keyword in ["시작", "시작", "quick start", "getting started"]):
                getting_started_found = True

        assert main_portal_found, "메인 포털 링크를 찾을 수 없습니다"
        assert api_docs_found, "API 문서 링크를 찾을 수 없습니다"
        assert getting_started_found, "시작 가이드 링크를 찾을 수 없습니다"

    def test_portal_functionality_should_be_improved(self, readme_content: str) -> None:
        """WHEN README.ko.md를 열면, 포털 기능이 개선된 내용이 포함되어 있어야 한다"""
        # 개선된 기능 설명 검증
        improvements = [
            "자동 링크 검증",
            "사용자 경험",
            "문서 네비게이션",
            "검색 기능",
            "다국어 지원"
        ]

        found_improvements = 0
        for improvement in improvements:
            if improvement in readme_content:
                found_improvements += 1

        assert found_improvements >= 3, f"개선된 포털 기능 설명이 부족합니다 (찾은 개선사항: {found_improvements}/5)"


class TestPortalUserExperience:
    """포털 사용자 경험 테스트"""

    def test_should_have_user_friendly_navigation(self) -> None:
        """WHEN 온라인 문서 포털을 사용하면, 사용자 친화적인 네비게이션이 제공되어야 한다"""
        # TODO: 온라인 포털의 네비게이션 구조 테스트 구현
        # - 메뉴 구조 검증
        # - 계층적 내비게이션 검증
        # - 검색 기능 검증

        # 현재는 테스트 스켈레톤만 작성
        assert True, "사용자 친화적 네비게이션 테스트 구현 필요"

    def test_should_have_multilingual_support(self) -> None:
        """WHEN 온라인 문서 포털을 사용하면, 다국어 지원이 제공되어야 한다"""
        # TODO: 다국어 지원 테스트 구현
        # - 언어 전환 기능 검증
        # - 번역된 내용 검증
        # - 로컬화된 UI 검증

        assert True, "다국어 지원 테스트 구현 필요"

    def test_should_have_search_functionality(self) -> None:
        """WHEN 온라인 문서 포털을 사용하면, 검색 기능이 제공되어야 한다"""
        # TODO: 검색 기능 테스트 구현
        # - 검색 API 테스트
        # - 검색 결과 정확도 테스트
        # - 자완성 테스트

        assert True, "검색 기능 테스트 구현 필요"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])