"""
TAG Search Operations Manager (Refactored)

TAG 데이터베이스의 통합 검색 기능을 제공합니다.

@DESIGN:TAG-SEARCH-MANAGER-002 - 모듈화된 검색 관리자
@TRUST:UNIFIED - 단일 책임: 검색 모듈들을 조정하는 퍼사드 패턴
@REFACTOR:SEARCH-MODULARIZATION-001 - 326줄 → 3개 모듈로 분해
"""

import logging
from typing import Any

from .basic_search import BasicTagSearch
from .advanced_search import AdvancedTagSearch

# 로깅 설정
logger = logging.getLogger(__name__)


class TagSearchManager:
    """
    TAG 검색 통합 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 퍼사드 패턴 적용
    - Unified: 단일 책임 - 검색 모듈들 간의 조정 역할만 담당
    - Secured: 각 모듈의 보안 기능 통합
    - Trackable: 구조화 로깅

    모듈 구성:
    - BasicTagSearch: 기본 검색 기능
    - AdvancedTagSearch: 고급 검색 및 통계
    """

    def __init__(self, database):
        """검색 관리자 초기화"""
        self._database = database

        # 각 전문 검색 관리자 인스턴스 생성
        self._basic_search = BasicTagSearch(database)
        self._advanced_search = AdvancedTagSearch(database)

        logger.debug("TagSearchManager 초기화 (모듈화)")

    # === 기본 검색 메서드 (BasicTagSearch 위임) ===
    def search_tags_by_category(self, category: str) -> list[dict[str, Any]]:
        """카테고리별 TAG 검색"""
        return self._basic_search.search_by_category(category)

    def search_tags_by_identifier(self, identifier: str) -> list[dict[str, Any]]:
        """식별자별 TAG 검색"""
        return self._basic_search.search_by_identifier(identifier)

    def search_tags_by_file(self, file_path: str) -> list[dict[str, Any]]:
        """파일별 TAG 검색"""
        return self._basic_search.search_by_file(file_path)

    def search_tags_by_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """패턴별 TAG 검색 (LIKE 사용)"""
        return self._basic_search.search_by_pattern(pattern)

    def search_tags_by_file_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """파일 패턴별 TAG 검색"""
        return self._basic_search.search_by_file_pattern(pattern)

    def search_tags_by_line_range(
        self, start_line: int, end_line: int, file_path: str | None = None
    ) -> list[dict[str, Any]]:
        """라인 범위별 TAG 검색 (파일 경로 옵션)"""
        return self._basic_search.search_by_line_range(start_line, end_line, file_path)

    # === 고급 검색 메서드 (AdvancedTagSearch 위임) ===
    def complex_search(
        self,
        category: str | None = None,
        identifier_pattern: str | None = None,
        file_pattern: str | None = None,
        description_pattern: str | None = None,
        line_min: int | None = None,
        line_max: int | None = None,
        created_after: str | None = None,
        created_before: str | None = None,
    ) -> list[dict[str, Any]]:
        """복잡한 조건으로 TAG 검색"""
        return self._advanced_search.complex_search(
            category=category,
            identifier_pattern=identifier_pattern,
            file_pattern=file_pattern,
            description_pattern=description_pattern,
            line_min=line_min,
            line_max=line_max,
            created_after=created_after,
            created_before=created_before,
        )

    def search_tags_with_references(self, tag_id: int) -> dict[str, Any]:
        """TAG와 관련 참조들을 함께 검색"""
        return self._advanced_search.search_with_references(tag_id)

    def get_statistics(self) -> dict[str, Any]:
        """TAG 데이터베이스 통계 정보"""
        return self._advanced_search.get_statistics()

    def get_tag_chains(self, start_tag_id: int, max_depth: int = 5) -> list[dict[str, Any]]:
        """TAG 체인 추적 (재귀적 참조 탐색)"""
        return self._advanced_search.get_tag_chains(start_tag_id, max_depth)