"""
Basic TAG Search Operations

TAG 데이터베이스의 기본 검색 기능을 담당합니다.

@DESIGN:TAG-BASIC-SEARCH-001 - 기본 검색 기능 분리
@TRUST:UNIFIED - 단일 책임: 기본 검색만 담당
"""

import logging
from typing import Any

# 로깅 설정
logger = logging.getLogger(__name__)


class BasicTagSearch:
    """
    TAG 기본 검색 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 검색 로직
    - Unified: 단일 책임 - 기본 검색만 담당
    - Secured: 안전한 SQL 쿼리 처리
    - Trackable: 구조화 로깅
    """

    def __init__(self, database):
        """기본 검색 관리자 초기화"""
        self._database = database
        logger.debug("BasicTagSearch 초기화")

    def search_by_category(self, category: str) -> list[dict[str, Any]]:
        """카테고리별 TAG 검색"""
        try:
            cursor = self._database.execute(
                "SELECT * FROM tags WHERE category = ? ORDER BY created_at", (category,)
            )
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_category",
                "category": category,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"카테고리 검색 실패: {e}")
            return []

    def search_by_identifier(self, identifier: str) -> list[dict[str, Any]]:
        """식별자별 TAG 검색"""
        try:
            cursor = self._database.execute(
                "SELECT * FROM tags WHERE identifier = ? ORDER BY created_at", (identifier,)
            )
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_identifier",
                "identifier": identifier,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"식별자 검색 실패: {e}")
            return []

    def search_by_file(self, file_path: str) -> list[dict[str, Any]]:
        """파일별 TAG 검색"""
        try:
            cursor = self._database.execute(
                "SELECT * FROM tags WHERE file_path = ? ORDER BY line_number", (file_path,)
            )
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_file",
                "file_path": file_path,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"파일 검색 실패: {e}")
            return []

    def search_by_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """패턴별 TAG 검색 (LIKE 사용)"""
        try:
            cursor = self._database.execute(
                """SELECT * FROM tags
                   WHERE identifier LIKE ? OR description LIKE ?
                   ORDER BY created_at""",
                (f"%{pattern}%", f"%{pattern}%"),
            )
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_pattern",
                "pattern": pattern,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"패턴 검색 실패: {e}")
            return []

    def search_by_file_pattern(self, pattern: str) -> list[dict[str, Any]]:
        """파일 패턴별 TAG 검색"""
        try:
            cursor = self._database.execute(
                """SELECT * FROM tags
                   WHERE file_path LIKE ?
                   ORDER BY file_path, line_number""",
                (f"%{pattern}%",),
            )
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_file_pattern",
                "pattern": pattern,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"파일 패턴 검색 실패: {e}")
            return []

    def search_by_line_range(
        self, start_line: int, end_line: int, file_path: str | None = None
    ) -> list[dict[str, Any]]:
        """라인 범위별 TAG 검색 (파일 경로 옵션)"""
        try:
            if file_path:
                cursor = self._database.execute(
                    """SELECT * FROM tags
                       WHERE file_path = ? AND line_number BETWEEN ? AND ?
                       ORDER BY line_number""",
                    (file_path, start_line, end_line),
                )
            else:
                cursor = self._database.execute(
                    """SELECT * FROM tags
                       WHERE line_number BETWEEN ? AND ?
                       ORDER BY line_number, file_path""",
                    (start_line, end_line),
                )

            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "search_by_line_range",
                "start_line": start_line,
                "end_line": end_line,
                "file_path": file_path,
                "count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"라인 범위 검색 실패: {e}")
            return []