"""
Core adapter functionality for SQLite backend and JSON API compatibility.

@FEATURE:ADAPTER-CORE-001 Basic CRUD operations and index management
@DESIGN:SEPARATED-CORE-001 Extracted from oversized adapter.py (631 LOC)
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Any

from .database import DatabaseError, TagDatabaseManager


class ApiCompatibilityError(Exception):
    """API 호환성 관련 오류"""


class AdapterCore:
    """핵심 어댑터 기능"""

    def __init__(self, database_path: Path, json_fallback_path: Path | None = None):
        """어댑터 코어 초기화"""
        self.database_path = Path(database_path)
        self.json_fallback_path = Path(json_fallback_path) if json_fallback_path else None

        # SQLite 백엔드 초기화
        try:
            self.db_manager = TagDatabaseManager(self.database_path)
            self.db_manager.initialize()
            self.backend_available = True
        except Exception:
            self.backend_available = False

    def initialize(self) -> None:
        """어댑터 초기화"""
        if self.backend_available:
            try:
                self.db_manager.initialize()
            except DatabaseError as e:
                raise ApiCompatibilityError(f"SQLite 백엔드 초기화 실패: {e}")
        else:
            raise ApiCompatibilityError("SQLite 백엔드를 사용할 수 없습니다")

    def initialize_index(self) -> None:
        """인덱스 초기화 (호환성용)"""
        self.initialize()

    def load_index(self) -> dict[str, Any]:
        """인덱스 로드 (기존 JSON API 호환)"""
        if self.backend_available:
            return self._load_from_sqlite()
        elif self.json_fallback_path and self.json_fallback_path.exists():
            return self._load_from_json_fallback()
        else:
            return self._create_empty_index()

    def _load_from_sqlite(self) -> dict[str, Any]:
        """SQLite에서 인덱스 로드"""
        try:
            all_tags = self.db_manager.get_all_tags()
            all_references = self.db_manager.get_all_references()

            # JSON 형식으로 변환
            index_data = {
                "version": "2.0.0",
                "backend": "sqlite",
                "updated": datetime.now().isoformat(),
                "statistics": {
                    "total_tags": len(all_tags),
                    "categories": {}
                },
                "index": {},
                "references": {}
            }

            # 카테고리별 통계 계산
            category_counts = {}
            for tag in all_tags:
                category = tag['category']
                category_counts[category] = category_counts.get(category, 0) + 1

            index_data["statistics"]["categories"] = category_counts

            # 인덱스 구성
            for tag in all_tags:
                tag_key = f"{tag['category']}:{tag['identifier']}"
                index_data["index"][tag_key] = [{
                    "file": tag['file_path'],
                    "line": tag.get('line_number', 1),
                    "context": f"@{tag_key} {tag.get('description', '')}"
                }]

            # 참조 관계 구성
            reference_map = {}
            for ref in all_references:
                source_tag = self.db_manager.get_tag_by_id(ref['source_tag_id'])
                target_tag = self.db_manager.get_tag_by_id(ref['target_tag_id'])

                if source_tag and target_tag:
                    source_key = f"{source_tag['category']}:{source_tag['identifier']}"
                    target_key = f"{target_tag['category']}:{target_tag['identifier']}"

                    if source_key not in reference_map:
                        reference_map[source_key] = []
                    reference_map[source_key].append(target_key)

            index_data["references"] = reference_map
            return index_data

        except Exception as e:
            raise ApiCompatibilityError(f"SQLite 인덱스 로드 실패: {e}")

    def _load_from_json_fallback(self) -> dict[str, Any]:
        """JSON fallback에서 로드"""
        try:
            with open(self.json_fallback_path, encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            raise ApiCompatibilityError(f"JSON fallback 로드 실패: {e}")

    def save_index(self, index_data: dict[str, Any]) -> None:
        """인덱스 저장 (기존 JSON API 호환)"""
        if self.backend_available:
            self._save_to_sqlite(index_data)
        else:
            raise ApiCompatibilityError("SQLite 백엔드를 사용할 수 없습니다")

    def _save_to_sqlite(self, index_data: dict[str, Any]) -> None:
        """SQLite에 인덱스 저장"""
        try:
            # 기존 데이터 클리어 (전체 교체)
            self.db_manager.clear_all_data()

            # TAG 데이터 저장
            index_entries = index_data.get("index", {})
            tag_id_mapping = {}

            for tag_key, entries in index_entries.items():
                if ":" not in tag_key or not entries:
                    continue

                category, identifier = tag_key.split(":", 1)
                first_entry = entries[0]

                tag_id = self.db_manager.insert_tag(
                    category=category,
                    identifier=identifier,
                    description=first_entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                    file_path=first_entry.get('file', ''),
                    line_number=first_entry.get('line', 1)
                )
                tag_id_mapping[tag_key] = tag_id

            # 참조 관계 저장
            references = index_data.get("references", {})
            for source_key, target_keys in references.items():
                if source_key in tag_id_mapping:
                    source_id = tag_id_mapping[source_key]
                    for target_key in target_keys:
                        if target_key in tag_id_mapping:
                            target_id = tag_id_mapping[target_key]
                            self.db_manager.create_reference(source_id, target_id, 'chain')

        except Exception as e:
            raise ApiCompatibilityError(f"SQLite 인덱스 저장 실패: {e}")

    def validate_index_schema(self, index_data: dict[str, Any]) -> bool:
        """인덱스 스키마 검증"""
        required_fields = ['version', 'index']
        return all(field in index_data for field in required_fields)

    def _create_empty_index(self) -> dict[str, Any]:
        """빈 인덱스 생성"""
        return {
            "version": "2.0.0",
            "backend": "sqlite",
            "updated": datetime.now().isoformat(),
            "statistics": {
                "total_tags": 0,
                "categories": {}
            },
            "index": {},
            "references": {}
        }

    def get_configuration_info(self) -> dict[str, Any]:
        """설정 정보 반환"""
        return {
            "backend_type": "sqlite",
            "database_path": str(self.database_path),
            "backend_available": self.backend_available,
            "json_fallback_path": str(self.json_fallback_path) if self.json_fallback_path else None
        }

    def close(self):
        """리소스 정리"""
        if hasattr(self, 'db_manager') and self.db_manager:
            try:
                self.db_manager.close()
            except:
                pass
