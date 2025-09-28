"""
TAG CRUD Operations Manager

TAG 데이터베이스의 생성, 읽기, 수정, 삭제 작업을 담당합니다.

@DESIGN:TAG-CRUD-MANAGER-001 - CRUD 작업 분리
@TRUST:UNIFIED - 단일 책임: CRUD 작업만 담당
"""

import logging
import time
from typing import Any, ContextManager

from .models import DatabaseError
from .transaction import PreparedInsertManager, TransactionManager

# 로깅 설정
logger = logging.getLogger(__name__)


class TagCrudManager:
    """
    TAG CRUD 작업 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 CRUD 작업 로직
    - Unified: 단일 책임 - CRUD 작업만 담당
    - Secured: 입력 검증 및 오류 처리
    - Trackable: 구조화 로깅과 성능 측정
    """

    # 유효한 카테고리 목록
    VALID_CATEGORIES = [
        "REQ", "DESIGN", "TASK", "TEST", "VISION", "STRUCT", "TECH", "ADR",
        "FEATURE", "API", "UI", "DATA", "PERF", "SEC", "DOCS", "TAG", "CUSTOM",
    ]

    def __init__(self, database):
        """CRUD 관리자 초기화"""
        self._database = database
        logger.debug("TagCrudManager 초기화")

    def insert_tag(
        self,
        category: str,
        identifier: str,
        description: str = "",
        file_path: str = "",
        line_number: int | None = None,
    ) -> int:
        """TAG 삽입 (Performance 개선 + 입력 검증)"""
        start_time = time.time()

        # 입력 검증 (TRUST 원칙: Secured)
        self._validate_tag_input(category, identifier)

        try:
            sql = """
            INSERT INTO tags (category, identifier, description, file_path, line_number)
            VALUES (?, ?, ?, ?, ?)
            """
            cursor = self._database.execute(
                sql, (category, identifier, description, file_path, line_number)
            )

            # 트랜잭션 내부가 아닌 경우에만 커밋
            if not self._is_in_transaction():
                self._database.commit()

            tag_id = cursor.lastrowid
            duration_ms = (time.time() - start_time) * 1000

            # 구조화 로깅
            logger.info({
                "operation": "insert_tag",
                "category": category,
                "identifier": identifier,
                "duration_ms": round(duration_ms, 2),
                "success": True,
            })

            return tag_id

        except Exception as e:
            duration_ms = (time.time() - start_time) * 1000
            logger.error({
                "operation": "insert_tag",
                "error": str(e),
                "duration_ms": round(duration_ms, 2),
                "success": False,
            })
            raise DatabaseError(f"TAG 삽입 실패: {e}")

    def get_tag_by_id(self, tag_id: int) -> dict[str, Any] | None:
        """ID로 TAG 검색"""
        try:
            cursor = self._database.execute("SELECT * FROM tags WHERE id = ?", (tag_id,))
            row = cursor.fetchone()
            return dict(row) if row else None
        except Exception as e:
            logger.error(f"TAG 검색 실패 (ID: {tag_id}): {e}")
            return None

    def update_tag(self, tag_id: int, **kwargs) -> int:
        """TAG 업데이트 (동적 필드 지원)"""
        if not kwargs:
            return 0

        set_clauses = []
        params = []

        # 허용된 필드만 업데이트
        allowed_fields = ["category", "identifier", "description", "file_path", "line_number"]

        for field, value in kwargs.items():
            if field in allowed_fields:
                set_clauses.append(f"{field} = ?")
                params.append(value)

        if not set_clauses:
            return 0

        params.append(tag_id)
        sql = f"UPDATE tags SET {', '.join(set_clauses)} WHERE id = ?"

        try:
            cursor = self._database.execute(sql, params)
            self._database.commit()

            logger.info({
                "operation": "update_tag",
                "tag_id": tag_id,
                "fields": list(kwargs.keys()),
                "rows_affected": cursor.rowcount,
            })

            return cursor.rowcount

        except Exception as e:
            logger.error(f"TAG 업데이트 실패: {e}")
            raise DatabaseError(f"TAG 업데이트 실패: {e}")

    def delete_tag(self, tag_id: int) -> int:
        """TAG 삭제"""
        try:
            cursor = self._database.execute("DELETE FROM tags WHERE id = ?", (tag_id,))
            self._database.commit()

            logger.info({
                "operation": "delete_tag",
                "tag_id": tag_id,
                "rows_affected": cursor.rowcount,
            })

            return cursor.rowcount

        except Exception as e:
            logger.error(f"TAG 삭제 실패: {e}")
            raise DatabaseError(f"TAG 삭제 실패: {e}")

    def get_all_tags(self) -> list[dict[str, Any]]:
        """모든 TAG 조회"""
        try:
            cursor = self._database.execute("SELECT * FROM tags ORDER BY created_at")
            return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            logger.error(f"전체 TAG 조회 실패: {e}")
            return []

    def bulk_insert_tags(self, tags_data: list[dict[str, Any]]) -> int:
        """대량 TAG 삽입 (최적화)"""
        if not tags_data:
            return 0

        sql = """
        INSERT INTO tags (category, identifier, description, file_path, line_number)
        VALUES (?, ?, ?, ?, ?)
        """

        params_list = []
        for tag in tags_data:
            # 각 태그에 대해 입력 검증
            category = tag.get("category", "")
            identifier = tag.get("identifier", "")

            if category and identifier:
                self._validate_tag_input(category, identifier, raise_error=False)

            params_list.append((
                category,
                identifier,
                tag.get("description", ""),
                tag.get("file_path", ""),
                tag.get("line_number"),
            ))

        try:
            with self.transaction():
                self._database.executemany(sql, params_list)

            logger.info({
                "operation": "bulk_insert_tags",
                "count": len(params_list),
                "success": True,
            })

            return len(params_list)

        except Exception as e:
            logger.error(f"대량 삽입 실패: {e}")
            raise DatabaseError(f"대량 삽입 실패: {e}")

    def create_reference(
        self, source_tag_id: int, target_tag_id: int, reference_type: str = "chain"
    ) -> int:
        """TAG 참조 관계 생성"""
        try:
            sql = """
            INSERT INTO tag_references (source_tag_id, target_tag_id, reference_type)
            VALUES (?, ?, ?)
            """
            cursor = self._database.execute(sql, (source_tag_id, target_tag_id, reference_type))
            self._database.commit()

            logger.info({
                "operation": "create_reference",
                "source_tag_id": source_tag_id,
                "target_tag_id": target_tag_id,
                "reference_type": reference_type,
            })

            return cursor.lastrowid

        except Exception as e:
            logger.error(f"참조 생성 실패: {e}")
            raise DatabaseError(f"참조 생성 실패: {e}")

    def get_references_by_source(self, source_tag_id: int) -> list[dict[str, Any]]:
        """소스 TAG로부터의 참조 관계 검색"""
        try:
            cursor = self._database.execute(
                "SELECT * FROM tag_references WHERE source_tag_id = ?", (source_tag_id,)
            )
            return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            logger.error(f"참조 관계 검색 실패: {e}")
            return []

    def get_tag_metadata(self, tag_id: int) -> dict[str, Any]:
        """TAG 메타데이터 조회 (참조 포함)"""
        tag = self.get_tag_by_id(tag_id)
        if not tag:
            return {}

        references = self.get_references_by_source(tag_id)
        return {
            "tag": tag,
            "reference_count": len(references),
            "references": references,
        }

    def transaction(self) -> ContextManager:
        """트랜잭션 컨텍스트 매니저 반환"""
        return TransactionManager(self._database)

    def prepared_insert(self) -> PreparedInsertManager:
        """Prepared statement 삽입 관리자 반환"""
        return PreparedInsertManager(self._database)

    def _validate_tag_input(self, category: str, identifier: str, raise_error: bool = True) -> bool:
        """TAG 입력 검증"""
        # 카테고리 검증
        if category not in self.VALID_CATEGORIES:
            if raise_error:
                raise ValueError(f"Invalid category: {category}")
            return False

        # 식별자 검증
        if not identifier.strip():
            if raise_error:
                raise ValueError("Identifier cannot be empty")
            return False

        return True

    def _is_in_transaction(self) -> bool:
        """현재 트랜잭션 내부인지 확인"""
        return (hasattr(self._database, '_transaction_manager') and
                self._database._transaction_manager and
                self._database._transaction_manager._in_transaction)