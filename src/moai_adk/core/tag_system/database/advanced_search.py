"""
Advanced TAG Search Operations

TAG 데이터베이스의 고급 검색 및 통계 기능을 담당합니다.

@DESIGN:TAG-ADVANCED-SEARCH-001 - 고급 검색 기능 분리
@TRUST:UNIFIED - 단일 책임: 고급 검색과 통계만 담당
"""

import logging
from typing import Any

# 로깅 설정
logger = logging.getLogger(__name__)


class AdvancedTagSearch:
    """
    TAG 고급 검색 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 고급 검색 로직
    - Unified: 단일 책임 - 고급 검색과 통계만 담당
    - Secured: 안전한 SQL 쿼리 처리
    - Trackable: 구조화 로깅
    """

    def __init__(self, database):
        """고급 검색 관리자 초기화"""
        self._database = database
        logger.debug("AdvancedTagSearch 초기화")

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
        try:
            conditions = []
            params = []

            # 검색 조건 구성
            self._add_condition(conditions, params, "category = ?", category)
            self._add_condition(conditions, params, "identifier LIKE ?",
                              f"%{identifier_pattern}%" if identifier_pattern else None)
            self._add_condition(conditions, params, "file_path LIKE ?",
                              f"%{file_pattern}%" if file_pattern else None)
            self._add_condition(conditions, params, "description LIKE ?",
                              f"%{description_pattern}%" if description_pattern else None)
            self._add_condition(conditions, params, "line_number >= ?", line_min)
            self._add_condition(conditions, params, "line_number <= ?", line_max)
            self._add_condition(conditions, params, "created_at >= ?", created_after)
            self._add_condition(conditions, params, "created_at <= ?", created_before)

            where_clause = " AND ".join(conditions) if conditions else "1=1"
            sql = f"SELECT * FROM tags WHERE {where_clause} ORDER BY created_at"

            cursor = self._database.execute(sql, params)
            results = [dict(row) for row in cursor.fetchall()]

            logger.debug({
                "operation": "complex_search",
                "conditions_count": len(conditions),
                "results_count": len(results),
            })

            return results

        except Exception as e:
            logger.error(f"복잡한 검색 실패: {e}")
            return []

    def search_with_references(self, tag_id: int) -> dict[str, Any]:
        """TAG와 관련 참조들을 함께 검색"""
        try:
            # 기본 TAG 정보
            tag_cursor = self._database.execute("SELECT * FROM tags WHERE id = ?", (tag_id,))
            tag_row = tag_cursor.fetchone()

            if not tag_row:
                return {}

            tag = dict(tag_row)

            # 이 TAG에서 참조하는 다른 TAG들
            outgoing_cursor = self._database.execute(
                """SELECT tr.*, t.category, t.identifier
                   FROM tag_references tr
                   JOIN tags t ON tr.target_tag_id = t.id
                   WHERE tr.source_tag_id = ?""",
                (tag_id,)
            )
            outgoing_refs = [dict(row) for row in outgoing_cursor.fetchall()]

            # 이 TAG를 참조하는 다른 TAG들
            incoming_cursor = self._database.execute(
                """SELECT tr.*, t.category, t.identifier
                   FROM tag_references tr
                   JOIN tags t ON tr.source_tag_id = t.id
                   WHERE tr.target_tag_id = ?""",
                (tag_id,)
            )
            incoming_refs = [dict(row) for row in incoming_cursor.fetchall()]

            result = {
                "tag": tag,
                "outgoing_references": outgoing_refs,
                "incoming_references": incoming_refs,
                "total_references": len(outgoing_refs) + len(incoming_refs),
            }

            logger.debug({
                "operation": "search_with_references",
                "tag_id": tag_id,
                "outgoing_count": len(outgoing_refs),
                "incoming_count": len(incoming_refs),
            })

            return result

        except Exception as e:
            logger.error(f"참조 포함 검색 실패: {e}")
            return {}

    def get_statistics(self) -> dict[str, Any]:
        """TAG 데이터베이스 통계 정보"""
        try:
            # 전체 TAG 수
            total_cursor = self._database.execute("SELECT COUNT(*) FROM tags")
            total_tags = total_cursor.fetchone()[0]

            # 카테고리별 TAG 수
            category_cursor = self._database.execute(
                "SELECT category, COUNT(*) FROM tags GROUP BY category ORDER BY COUNT(*) DESC"
            )
            categories = [{"category": row[0], "count": row[1]} for row in category_cursor.fetchall()]

            # 전체 참조 수
            ref_cursor = self._database.execute("SELECT COUNT(*) FROM tag_references")
            total_references = ref_cursor.fetchone()[0]

            # 파일별 TAG 수 (상위 10개)
            file_cursor = self._database.execute(
                """SELECT file_path, COUNT(*) as count
                   FROM tags
                   WHERE file_path != ''
                   GROUP BY file_path
                   ORDER BY count DESC
                   LIMIT 10"""
            )
            top_files = [{"file_path": row[0], "count": row[1]} for row in file_cursor.fetchall()]

            stats = {
                "total_tags": total_tags,
                "total_references": total_references,
                "categories": categories,
                "top_files": top_files,
            }

            logger.debug({
                "operation": "get_statistics",
                "total_tags": total_tags,
                "categories_count": len(categories),
            })

            return stats

        except Exception as e:
            logger.error(f"통계 조회 실패: {e}")
            return {}

    def get_tag_chains(self, start_tag_id: int, max_depth: int = 5) -> list[dict[str, Any]]:
        """TAG 체인 추적 (재귀적 참조 탐색)"""
        try:
            chains = []
            visited = set()

            def _build_chain(tag_id: int, chain: list, depth: int):
                if depth > max_depth or tag_id in visited:
                    return

                visited.add(tag_id)

                # 현재 TAG 정보
                tag_cursor = self._database.execute("SELECT * FROM tags WHERE id = ?", (tag_id,))
                tag_row = tag_cursor.fetchone()

                if not tag_row:
                    return

                current_chain = chain + [dict(tag_row)]

                # 참조되는 TAG들 탐색
                ref_cursor = self._database.execute(
                    "SELECT target_tag_id FROM tag_references WHERE source_tag_id = ?", (tag_id,)
                )

                references = ref_cursor.fetchall()

                if not references:
                    # 체인 종료점
                    chains.append(current_chain)
                else:
                    # 재귀적으로 체인 탐색
                    for (target_id,) in references:
                        _build_chain(target_id, current_chain, depth + 1)

                visited.remove(tag_id)

            _build_chain(start_tag_id, [], 0)

            logger.debug({
                "operation": "get_tag_chains",
                "start_tag_id": start_tag_id,
                "chains_found": len(chains),
                "max_depth": max_depth,
            })

            return chains

        except Exception as e:
            logger.error(f"TAG 체인 추적 실패: {e}")
            return []

    def _add_condition(self, conditions: list, params: list, condition: str, value: Any) -> None:
        """검색 조건 추가 (None 값 무시)"""
        if value is not None:
            conditions.append(condition)
            params.append(value)