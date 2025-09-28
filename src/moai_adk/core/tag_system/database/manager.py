"""
Tag Database Manager (Refactored)

SQLite 기반 TAG 데이터베이스의 통합 관리 기능을 제공합니다.

@DESIGN:TAG-DATABASE-MANAGER-002 - 모듈화된 관리자 통합
@TRUST:UNIFIED - 단일 책임: 모듈들을 조정하는 퍼사드 패턴
@REFACTOR:MODULARIZATION-001 - TRUST 원칙 준수를 위한 모듈 분해
"""

import logging
from pathlib import Path
from typing import Any, ContextManager

from .connection import DatabaseConnection, TagDatabase
from .crud_manager import TagCrudManager
from .schema_manager import SchemaManager
from .search_manager import TagSearchManager
from .transaction import PreparedInsertManager

# 로깅 설정
logger = logging.getLogger(__name__)


class TagDatabaseManager:
    """
    SQLite 기반 TAG 데이터베이스 통합 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 퍼사드 패턴 적용
    - Unified: 단일 책임 - 모듈들 간의 조정 역할만 담당
    - Secured: 각 모듈의 보안 기능 통합
    - Trackable: 구조화 로깅

    모듈 구성:
    - SchemaManager: 스키마 및 인덱스 관리
    - TagCrudManager: CRUD 작업 처리
    - TagSearchManager: 검색 기능 제공
    """

    def __init__(self, db_path: Path):
        """데이터베이스 관리자 초기화"""
        self.db_path = Path(db_path)

        # 연결 및 데이터베이스 초기화
        self._connection = DatabaseConnection(self.db_path)
        self._database = TagDatabase(self._connection)

        # 각 전문 관리자 인스턴스 생성
        self._schema_manager = SchemaManager(self._database)
        self._crud_manager = TagCrudManager(self._database)
        self._search_manager = TagSearchManager(self._database)

        # 로깅
        logger.debug(f"TagDatabaseManager 초기화 (모듈화): {self.db_path}")

    def initialize(self, create_indexes: bool = True) -> None:
        """데이터베이스 스키마 및 인덱스 초기화"""
        try:
            # 스키마 초기화
            self._schema_manager.initialize_schema(self.db_path)

            # 인덱스 생성
            if create_indexes:
                self._schema_manager.create_indexes()

            # 트랜잭션 커밋
            self._database.commit()
            logger.info(f"데이터베이스 초기화 완료 (모듈화): {self.db_path}")

        except Exception as e:
            logger.error(f"데이터베이스 초기화 실패: {e}")
            raise

    # === 스키마 관리 메서드 (SchemaManager 위임) ===
    def get_schema(self) -> dict[str, dict[str, str]]:
        """테이블 스키마 정보 반환"""
        return self._schema_manager.get_schema()

    def get_indexes(self) -> list[str]:
        """생성된 인덱스 목록 반환"""
        return self._schema_manager.get_indexes()

    # === CRUD 작업 메서드 (TagCrudManager 위임) ===
    def insert_tag(
        self,
        category: str,
        identifier: str,
        description: str = "",
        file_path: str = "",
        line_number: int | None = None,
    ) -> int:
        """TAG 삽입 (Performance 개선 + 입력 검증)"""
        return self._crud_manager.insert_tag(category, identifier, description, file_path, line_number)

    def get_tag_by_id(self, tag_id: int) -> dict[str, Any] | None:
        """ID로 TAG 검색"""
        return self._crud_manager.get_tag_by_id(tag_id)

    def update_tag(self, tag_id: int, **kwargs) -> int:
        """TAG 업데이트 (동적 필드 지원)"""
        return self._crud_manager.update_tag(tag_id, **kwargs)

    def delete_tag(self, tag_id: int) -> int:
        """TAG 삭제"""
        return self._crud_manager.delete_tag(tag_id)

    def get_all_tags(self) -> list[dict[str, Any]]:
        """모든 TAG 조회"""
        return self._crud_manager.get_all_tags()

    def bulk_insert_tags(self, tags_data: list[dict[str, Any]]) -> int:
        """대량 TAG 삽입 (최적화)"""
        return self._crud_manager.bulk_insert_tags(tags_data)

    def create_reference(
        self, source_tag_id: int, target_tag_id: int, reference_type: str = "chain"
    ) -> int:
        """TAG 참조 관계 생성"""
        return self._crud_manager.create_reference(source_tag_id, target_tag_id, reference_type)

    def get_references_by_source(self, source_tag_id: int) -> list[dict[str, Any]]:
        """소스 TAG로부터의 참조 관계 검색"""
        return self._crud_manager.get_references_by_source(source_tag_id)

    def get_tag_metadata(self, tag_id: int) -> dict[str, Any]:
        """TAG 메타데이터 조회 (참조 포함)"""
        return self._crud_manager.get_tag_metadata(tag_id)

    def transaction(self) -> ContextManager:
        """트랜잭션 컨텍스트 매니저 반환"""
        return self._crud_manager.transaction()

    def prepared_insert(self) -> PreparedInsertManager:
        """Prepared statement 삽입 관리자 반환"""
        return self._crud_manager.prepared_insert()

    # === 검색 작업 메서드 (TagSearchManager 위임) ===
    def search_tags_by_category(self, category: str) -> list[dict[str, Any]]:
        """카테고리별 TAG 검색"""
        return self._search_manager.search_tags_by_category(category)

    def search_tags_by_identifier(self, identifier: str) -> list[dict[str, Any]]:
        """식별자별 TAG 검색"""
        return self._search_manager.search_tags_by_identifier(identifier)

    def search_tags_by_file(self, file_path: str) -> list[dict[str, Any]]:
        """파일별 TAG 검색"""
        return self._search_manager.search_tags_by_file(file_path)

    def search_tags_by_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """패턴별 TAG 검색 (LIKE 사용)"""
        return self._search_manager.search_tags_by_pattern(pattern)

    def search_tags_by_file_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """파일 패턴별 TAG 검색"""
        return self._search_manager.search_tags_by_file_pattern(pattern)

    def search_tags_by_line_range(
        self, start_line: int, end_line: int, file_path: str | None = None
    ) -> list[dict[str, Any]]:
        """라인 범위별 TAG 검색 (파일 경로 옵션)"""
        return self._search_manager.search_tags_by_line_range(start_line, end_line, file_path)

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
        return self._search_manager.complex_search(
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
        return self._search_manager.search_tags_with_references(tag_id)

    def get_statistics(self) -> dict[str, Any]:
        """TAG 데이터베이스 통계 정보"""
        return self._search_manager.get_statistics()

    # === 유틸리티 메서드 ===
    def close(self) -> None:
        """데이터베이스 연결 종료"""
        self._connection.close()
        logger.debug("데이터베이스 연결 종료 (모듈화)")