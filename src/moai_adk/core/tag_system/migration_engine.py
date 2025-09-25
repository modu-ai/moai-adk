"""
Core migration engine for JSON ↔ SQLite conversion.

@FEATURE:MIGRATION-ENGINE-001 Core transformation logic
@DESIGN:SEPARATED-ENGINE-001 Extracted from oversized migration.py (644 LOC)
"""

import json
import time
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable
from datetime import datetime

from .database import TagDatabaseManager, DatabaseError
from .migration_models import MigrationResult, MigrationProgress, BackupManager


class MigrationEngine:
    """핵심 마이그레이션 엔진"""

    def __init__(self, database_path: Path, json_path: Path, backup_manager: BackupManager):
        self.database_path = database_path
        self.json_path = json_path
        self.backup_manager = backup_manager

    def migrate_json_to_sqlite_core(self, json_data: Dict[str, Any],
                                   db_manager: TagDatabaseManager,
                                   mode: str = 'full',
                                   batch_size: int = 1000,
                                   progress_callback: Optional[Callable] = None) -> MigrationResult:
        """JSON에서 SQLite로 핵심 마이그레이션"""
        result = MigrationResult(success=False)

        try:
            if mode == 'full':
                result = self._full_migration(db_manager, json_data, result, progress_callback, batch_size)
            elif mode == 'incremental':
                result = self._incremental_migration(db_manager, json_data, result, progress_callback)

            result.success = True

        except Exception as e:
            result.errors.append(f"마이그레이션 실행 중 오류: {e}")

        return result

    def migrate_sqlite_to_json(self) -> MigrationResult:
        """SQLite에서 JSON으로 마이그레이션"""
        result = MigrationResult(success=False)

        try:
            # SQLite 데이터베이스 매니저 직접 사용
            db_manager = TagDatabaseManager(self.database_path)
            db_manager.initialize()
            all_tags = db_manager.get_all_tags()

            # 기존 JSON 형식으로 변환
            json_data = self._convert_sqlite_to_original_json_format(all_tags, db_manager)

            # JSON 파일로 저장
            with open(self.json_path, 'w', encoding='utf-8') as f:
                json.dump(json_data, f, indent=2, ensure_ascii=False)

            result.migrated_tags_count = len(all_tags)
            result.success = True

        except Exception as e:
            result.errors.append(str(e))

        return result

    def _convert_sqlite_to_original_json_format(self, all_tags: List[Dict[str, Any]],
                                               db_manager: TagDatabaseManager) -> Dict[str, Any]:
        """SQLite 데이터를 원본 JSON 형식으로 변환"""
        # 기본 구조 생성
        json_data = {
            "version": "1.0.0",
            "updated": datetime.now().isoformat(),
            "statistics": {
                "total_tags": len(all_tags),
                "categories": {
                    "Primary": 0,
                    "Steering": 0,
                    "Implementation": 0,
                    "Quality": 0
                }
            },
            "index": {},
            "references": {}
        }

        # TAG 데이터 변환
        for tag in all_tags:
            tag_key = f"{tag['category']}:{tag['identifier']}"

            # 인덱스에 TAG 추가
            json_data["index"][tag_key] = [{
                "file": tag['file_path'],
                "line": tag.get('line_number', 1),
                "context": f"@{tag_key} {tag['description'] or ''}"
            }]

            # 카테고리 통계 업데이트
            category_group = self._get_category_group_for_stats(tag['category'])
            if category_group in json_data["statistics"]["categories"]:
                json_data["statistics"]["categories"][category_group] += 1

        # 참조 관계 변환 (간단한 구현)
        for tag in all_tags:
            references = db_manager.get_references_by_source(tag['id'])
            if references:
                source_key = f"{tag['category']}:{tag['identifier']}"
                json_data["references"][source_key] = []

                for ref in references:
                    target_tag = db_manager.get_tag_by_id(ref['target_tag_id'])
                    if target_tag:
                        target_key = f"{target_tag['category']}:{target_tag['identifier']}"
                        json_data["references"][source_key].append(target_key)

        return json_data

    def _get_category_group_for_stats(self, category: str) -> str:
        """통계용 카테고리 그룹 결정"""
        if category in ["REQ", "DESIGN", "TASK", "TEST"]:
            return "Primary"
        elif category in ["VISION", "STRUCT", "TECH", "ADR"]:
            return "Steering"
        elif category in ["FEATURE", "API", "UI", "DATA"]:
            return "Implementation"
        elif category in ["PERF", "SEC", "DOCS", "TAG"]:
            return "Quality"
        else:
            return "Primary"

    def _full_migration(self, db_manager: TagDatabaseManager, json_data: Dict[str, Any],
                       result: MigrationResult, progress_callback: Optional[Callable],
                       batch_size: int) -> MigrationResult:
        """전체 마이그레이션"""
        index_data = json_data.get("index", {})
        references_data = json_data.get("references", {})

        total_items = len(index_data) + len(references_data)
        completed = 0

        # 진행률 보고
        if progress_callback:
            progress = MigrationProgress("ANALYZING", 0, total_items, 0)
            progress_callback(progress)

        # TAG 마이그레이션
        if progress_callback:
            progress = MigrationProgress("MIGRATING_TAGS", completed, total_items, time.time())
            progress_callback(progress)

        tag_id_mapping = {}  # 기존 TAG ID → 새 DB ID 매핑

        for tag_key, tag_entries in index_data.items():
            try:
                if ":" not in tag_key:
                    continue

                category, identifier = tag_key.split(":", 1)

                # 첫 번째 엔트리로 기본 TAG 생성 (중복 방지)
                if tag_entries and tag_key not in tag_id_mapping:
                    first_entry = tag_entries[0]

                    # TAG 데이터 생성
                    tag_data = {
                        'category': category,
                        'identifier': identifier,
                        'description': first_entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                        'file_path': first_entry.get('file', ''),
                        'line_number': first_entry.get('line', 1)
                    }

                    tag_id = db_manager.insert_tag(**tag_data)
                    tag_id_mapping[tag_key] = tag_id
                    result.migrated_tags_count += 1

            except Exception as e:
                result.errors.append(f"TAG {tag_key} 처리 중 오류: {e}")

            completed += 1
            if progress_callback:
                progress = MigrationProgress("MIGRATING_TAGS", completed, total_items, time.time())
                progress_callback(progress)

        # 참조 관계 마이그레이션
        if progress_callback:
            progress = MigrationProgress("MIGRATING_REFERENCES", completed, total_items, time.time())
            progress_callback(progress)

        for source_tag, target_tags in references_data.items():
            if source_tag in tag_id_mapping:
                source_id = tag_id_mapping[source_tag]
                for target_tag in target_tags:
                    if target_tag in tag_id_mapping:
                        target_id = tag_id_mapping[target_tag]
                        try:
                            db_manager.create_reference(source_id, target_id, 'chain')
                            result.migrated_references_count += 1
                        except Exception as e:
                            result.errors.append(f"참조 {source_tag} → {target_tag} 생성 실패: {e}")

            completed += 1
            if progress_callback:
                progress = MigrationProgress("MIGRATING_REFERENCES", completed, total_items, time.time())
                progress_callback(progress)

        return result

    def _incremental_migration(self, db_manager: TagDatabaseManager, json_data: Dict[str, Any],
                              result: MigrationResult, progress_callback: Optional[Callable]) -> MigrationResult:
        """증분 마이그레이션"""
        existing_tags = {f"{tag['category']}:{tag['identifier']}" for tag in db_manager.get_all_tags()}
        index_data = json_data.get("index", {})

        new_tags = set(index_data.keys()) - existing_tags
        total_new = len(new_tags)

        if progress_callback:
            progress = MigrationProgress("INCREMENTAL_MIGRATION", 0, total_new, time.time())
            progress_callback(progress)

        completed = 0
        for tag_key in new_tags:
            try:
                if ":" not in tag_key:
                    continue

                category, identifier = tag_key.split(":", 1)
                tag_entries = index_data[tag_key]

                if tag_entries:
                    first_entry = tag_entries[0]
                    tag_data = {
                        'category': category,
                        'identifier': identifier,
                        'description': first_entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                        'file_path': first_entry.get('file', ''),
                        'line_number': first_entry.get('line', 1)
                    }

                    db_manager.insert_tag(**tag_data)
                    result.migrated_tags_count += 1

            except Exception as e:
                result.errors.append(f"증분 TAG {tag_key} 처리 중 오류: {e}")

            completed += 1
            if progress_callback:
                progress = MigrationProgress("INCREMENTAL_MIGRATION", completed, total_new, time.time())
                progress_callback(progress)

        result.preserved_tags_count = len(existing_tags)
        return result