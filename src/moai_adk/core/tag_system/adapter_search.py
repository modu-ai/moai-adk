"""
Advanced search and traceability features for tag adapter.

@FEATURE:ADAPTER-SEARCH-001 Complex search and traceability operations
@DESIGN:SEPARATED-SEARCH-001 Extracted from oversized adapter.py (631 LOC)
"""

from typing import Any

from .database import TagDatabaseManager


class AdapterSearch:
    """검색 및 추적 기능"""

    def __init__(self, db_manager: TagDatabaseManager):
        """검색 엔진 초기화"""
        self.db_manager = db_manager

    def _get_category_group(self, category: str) -> str:
        """카테고리 그룹 결정"""
        if category in ["REQ", "DESIGN", "TASK", "TEST"]:
            return "Primary"
        elif category in ["VISION", "STRUCT", "TECH", "ADR"]:
            return "Steering"
        elif category in ["FEATURE", "API", "UI", "DATA"]:
            return "Implementation"
        elif category in ["PERF", "SEC", "DOCS", "TAG"]:
            return "Quality"
        else:
            return "Other"

    def search_by_category(self, category: str, **filters) -> list[dict[str, Any]]:
        """카테고리별 검색 (고급 필터링 포함)"""
        try:
            # 기본 카테고리 검색
            tags = self.db_manager.search_tags_by_category(category)

            # 추가 필터 적용
            filtered_tags = []
            for tag in tags:
                include_tag = True

                # 파일 경로 필터
                if "file_path" in filters:
                    file_pattern = filters["file_path"]
                    if file_pattern not in tag.get("file_path", ""):
                        include_tag = False

                # 식별자 패턴 필터
                if "identifier_pattern" in filters:
                    pattern = filters["identifier_pattern"].lower()
                    identifier = tag.get("identifier", "").lower()
                    if pattern not in identifier:
                        include_tag = False

                # 설명 키워드 필터
                if "description_keywords" in filters:
                    keywords = filters["description_keywords"]
                    description = tag.get("description", "").lower()
                    if not any(keyword.lower() in description for keyword in keywords):
                        include_tag = False

                # 라인 범위 필터
                if "line_range" in filters:
                    start_line, end_line = filters["line_range"]
                    tag_line = tag.get("line_number", 0)
                    if not (start_line <= tag_line <= end_line):
                        include_tag = False

                # 그룹 필터
                if "category_group" in filters:
                    target_group = filters["category_group"]
                    tag_group = self._get_category_group(tag.get("category", ""))
                    if tag_group != target_group:
                        include_tag = False

                if include_tag:
                    # 추가 메타데이터 포함
                    enriched_tag = tag.copy()
                    enriched_tag["category_group"] = self._get_category_group(
                        tag.get("category", "")
                    )
                    enriched_tag["tag_key"] = (
                        f"{tag.get('category', '')}:{tag.get('identifier', '')}"
                    )
                    filtered_tags.append(enriched_tag)

            # 정렬 옵션
            sort_by = filters.get("sort_by", "identifier")
            reverse = filters.get("reverse", False)

            if sort_by == "identifier":
                filtered_tags.sort(
                    key=lambda x: x.get("identifier", ""), reverse=reverse
                )
            elif sort_by == "file_path":
                filtered_tags.sort(
                    key=lambda x: x.get("file_path", ""), reverse=reverse
                )
            elif sort_by == "line_number":
                filtered_tags.sort(
                    key=lambda x: x.get("line_number", 0), reverse=reverse
                )

            # 결과 제한
            limit = filters.get("limit")
            if limit:
                filtered_tags = filtered_tags[:limit]

            return filtered_tags

        except Exception as e:
            raise Exception(f"카테고리 검색 실패 ({category}): {e}")

    def get_traceability_chain(
        self,
        tag_identifier: str,
        direction: str = "both",
        max_depth: int = 5,
        include_details: bool = True,
        category_filter: list[str] | None = None,
    ) -> dict[str, Any]:
        """추적성 체인 분석 (고도화된 버전)"""
        try:
            # TAG 식별자에서 카테고리와 ID 분리
            if ":" not in tag_identifier:
                raise ValueError(f"올바르지 않은 TAG 형식: {tag_identifier}")

            category, identifier = tag_identifier.split(":", 1)

            # 시작 TAG 찾기
            start_tag = None
            all_tags = self.db_manager.get_all_tags()
            for tag in all_tags:
                if tag["category"] == category and tag["identifier"] == identifier:
                    start_tag = tag
                    break

            if not start_tag:
                return {
                    "start_tag": tag_identifier,
                    "found": False,
                    "error": "TAG를 찾을 수 없습니다",
                }

            # 체인 분석 결과
            result = {
                "start_tag": tag_identifier,
                "found": True,
                "forward_chain": [],  # 이 TAG가 참조하는 TAG들
                "backward_chain": [],  # 이 TAG를 참조하는 TAG들
                "statistics": {
                    "total_forward": 0,
                    "total_backward": 0,
                    "max_depth_reached": 0,
                    "categories_involved": set(),
                },
            }

            # Forward chain (이 TAG가 참조하는 것들)
            if direction in ["forward", "both"]:
                forward_visited = {start_tag["id"]}
                forward_chain = self._build_reference_chain(
                    start_tag["id"],
                    "forward",
                    max_depth,
                    forward_visited,
                    include_details,
                    category_filter,
                )
                result["forward_chain"] = forward_chain
                result["statistics"]["total_forward"] = self._count_chain_nodes(
                    forward_chain
                )

            # Backward chain (이 TAG를 참조하는 것들)
            if direction in ["backward", "both"]:
                backward_visited = {start_tag["id"]}
                backward_chain = self._build_reference_chain(
                    start_tag["id"],
                    "backward",
                    max_depth,
                    backward_visited,
                    include_details,
                    category_filter,
                )
                result["backward_chain"] = backward_chain
                result["statistics"]["total_backward"] = self._count_chain_nodes(
                    backward_chain
                )

            # 통계 계산
            all_categories = set()
            for chain in [result["forward_chain"], result["backward_chain"]]:
                self._collect_categories_from_chain(chain, all_categories)

            result["statistics"]["categories_involved"] = list(all_categories)
            result["statistics"]["max_depth_reached"] = min(
                max_depth,
                max(
                    self._get_chain_depth(result["forward_chain"]),
                    self._get_chain_depth(result["backward_chain"]),
                ),
            )

            return result

        except Exception as e:
            return {"start_tag": tag_identifier, "found": False, "error": str(e)}

    def _build_reference_chain(
        self,
        tag_id: int,
        direction: str,
        max_depth: int,
        visited: set,
        include_details: bool,
        category_filter: list[str] | None,
        current_depth: int = 0,
    ) -> list[dict[str, Any]]:
        """참조 체인 구축"""
        if current_depth >= max_depth:
            return []

        chain = []
        if direction == "forward":
            references = self.db_manager.get_references_by_source(tag_id)
            ref_key = "target_tag_id"
        else:
            references = self.db_manager.get_references_by_target(tag_id)
            ref_key = "source_tag_id"

        for ref in references:
            target_tag_id = ref[ref_key]
            if target_tag_id in visited:
                continue  # 순환 참조 방지

            target_tag = self.db_manager.get_tag_by_id(target_tag_id)
            if not target_tag:
                continue

            # 카테고리 필터 적용
            if category_filter and target_tag["category"] not in category_filter:
                continue

            visited.add(target_tag_id)

            # 체인 노드 생성
            node = {
                "tag_id": target_tag_id,
                "category": target_tag["category"],
                "identifier": target_tag["identifier"],
                "tag_key": f"{target_tag['category']}:{target_tag['identifier']}",
                "depth": current_depth + 1,
                "reference_type": ref.get("reference_type", "chain"),
            }

            if include_details:
                node.update(
                    {
                        "description": target_tag.get("description", ""),
                        "file_path": target_tag.get("file_path", ""),
                        "line_number": target_tag.get("line_number", 0),
                        "category_group": self._get_category_group(
                            target_tag["category"]
                        ),
                    }
                )

            # 재귀적으로 다음 레벨 구축
            sub_chain = self._build_reference_chain(
                target_tag_id,
                direction,
                max_depth,
                visited.copy(),
                include_details,
                category_filter,
                current_depth + 1,
            )
            if sub_chain:
                node["children"] = sub_chain

            chain.append(node)

        return chain

    def _count_chain_nodes(self, chain: list[dict[str, Any]]) -> int:
        """체인의 총 노드 수 계산"""
        count = len(chain)
        for node in chain:
            if "children" in node:
                count += self._count_chain_nodes(node["children"])
        return count

    def _collect_categories_from_chain(
        self, chain: list[dict[str, Any]], categories: set
    ):
        """체인에서 모든 카테고리 수집"""
        for node in chain:
            categories.add(node.get("category", ""))
            if "children" in node:
                self._collect_categories_from_chain(node["children"], categories)

    def _get_chain_depth(self, chain: list[dict[str, Any]]) -> int:
        """체인의 최대 깊이 계산"""
        if not chain:
            return 0

        max_depth = 0
        for node in chain:
            current_depth = node.get("depth", 0)
            if "children" in node:
                child_depth = self._get_chain_depth(node["children"])
                current_depth = max(current_depth, child_depth)
            max_depth = max(max_depth, current_depth)

        return max_depth
