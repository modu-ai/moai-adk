"""
@TEST:SPEC-009-TAG-MIGRATION-001 - TAG 마이그레이션 도구 실패 테스트

RED 단계: JSON ↔ SQLite 양방향 변환 및 데이터 무결성 실패 테스트
"""

import pytest
import json
import tempfile
import time
from pathlib import Path
from typing import Dict, List, Any, Optional
from unittest.mock import Mock, patch
from datetime import datetime

# 아직 구현되지 않은 모듈들 - 실패할 예정
from moai_adk.core.tag_system.migration import (
    TagMigrationTool,
    MigrationResult,
    MigrationError,
    DataValidationError,
    MigrationProgress,
    BackupManager
)
from moai_adk.core.tag_system.database import TagDatabaseManager


class TestTagMigrationTool:
    """TAG 마이그레이션 도구 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_db = Path(tempfile.mktemp(suffix='.db'))
        self.temp_json = Path(tempfile.mktemp(suffix='.json'))
        self.backup_dir = Path(tempfile.mkdtemp())

        self.migration_tool = TagMigrationTool(
            database_path=self.temp_db,
            json_path=self.temp_json,
            backup_directory=self.backup_dir
        )

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil

        if self.temp_db.exists():
            self.temp_db.unlink()
        if self.temp_json.exists():
            self.temp_json.unlink()
        if self.backup_dir.exists():
            shutil.rmtree(self.backup_dir, ignore_errors=True)

    def test_should_migrate_json_to_sqlite_with_complete_data_preservation(self):
        """
        Given: 현재 프로젝트의 실제 JSON 인덱스 데이터
        When: SQLite로 마이그레이션을 수행할 때
        Then: 모든 데이터가 손실 없이 변환되어야 함
        """
        # GIVEN: 실제 프로젝트 JSON 구조 시뮬레이션
        source_json_data = {
            "version": "0.1.9",
            "updated": "2025-09-24T19:14:30.319589",
            "statistics": {
                "total_tags": 441,
                "categories": {
                    "Primary": 276,
                    "Steering": 37,
                    "Implementation": 72,
                    "Quality": 50,
                    "Legacy": 0
                }
            },
            "index": {
                "REQ:USER-AUTH-001": [
                    {
                        "file": "requirements.md",
                        "line": 10,
                        "context": "@REQ:USER-AUTH-001 사용자 인증 기능"
                    },
                    {
                        "file": "design.md",
                        "line": 5,
                        "context": "참조: @REQ:USER-AUTH-001"
                    }
                ],
                "DESIGN:JWT-SYSTEM-001": [
                    {
                        "file": "design.md",
                        "line": 20,
                        "context": "@DESIGN:JWT-SYSTEM-001 → @REQ:USER-AUTH-001"
                    }
                ],
                "TASK:API-IMPL-001": [
                    {
                        "file": "tasks.md",
                        "line": 15,
                        "context": "@TASK:API-IMPL-001 API 구현 작업"
                    }
                ]
            },
            "references": {
                "REQ:USER-AUTH-001": ["DESIGN:JWT-SYSTEM-001", "TASK:API-IMPL-001"],
                "DESIGN:JWT-SYSTEM-001": ["TASK:API-IMPL-001"]
            }
        }

        with open(self.temp_json, 'w', encoding='utf-8') as f:
            json.dump(source_json_data, f, indent=2, ensure_ascii=False)

        # WHEN: JSON → SQLite 마이그레이션
        migration_result = self.migration_tool.migrate_json_to_sqlite()

        # THEN: 완전한 데이터 보존 확인
        assert migration_result.success is True
        assert migration_result.migrated_tags_count == 3  # REQ, DESIGN, TASK
        assert migration_result.migrated_references_count == 3  # 2 + 1 참조
        assert migration_result.errors == []

        # SQLite 데이터 검증
        db_manager = TagDatabaseManager(self.temp_db)
        db_manager.initialize()

        # 모든 TAG가 올바르게 저장되었는지 확인
        all_tags = db_manager.get_all_tags()
        assert len(all_tags) == 3

        # 특정 TAG 세부사항 확인
        auth_tag = db_manager.search_tags_by_identifier('USER-AUTH-001')
        assert len(auth_tag) == 1
        assert auth_tag[0]['category'] == 'REQ'
        assert auth_tag[0]['description'] == '사용자 인증 기능'

        # 참조 관계 확인
        auth_tag_id = auth_tag[0]['id']
        references = db_manager.get_references_by_source(auth_tag_id)
        assert len(references) == 2  # DESIGN, TASK로의 참조

    def test_should_migrate_sqlite_to_json_maintaining_original_format(self):
        """
        Given: SQLite 데이터베이스에 저장된 TAG 데이터
        When: JSON 형식으로 역마이그레이션을 수행할 때
        Then: 기존 JSON 형식과 완전히 호환되는 구조로 내보내야 함
        """
        # GIVEN: SQLite에 테스트 데이터 준비
        db_manager = TagDatabaseManager(self.temp_db)
        db_manager.initialize()

        # 테스트 TAG들 삽입
        tags_data = [
            ('REQ', 'USER-LOGIN-001', '사용자 로그인', '/path/req.md', 10),
            ('DESIGN', 'AUTH-ARCH-001', '인증 아키텍처', '/path/design.md', 20),
            ('TASK', 'LOGIN-IMPL-001', '로그인 구현', '/path/task.md', 30)
        ]

        tag_ids = []
        for category, identifier, description, file_path, line_number in tags_data:
            tag_id = db_manager.insert_tag(
                category=category,
                identifier=identifier,
                description=description,
                file_path=file_path,
                line_number=line_number
            )
            tag_ids.append(tag_id)

        # 참조 관계 생성
        db_manager.create_reference(tag_ids[0], tag_ids[1], 'chain')  # REQ → DESIGN
        db_manager.create_reference(tag_ids[1], tag_ids[2], 'chain')  # DESIGN → TASK

        # WHEN: SQLite → JSON 마이그레이션
        export_result = self.migration_tool.migrate_sqlite_to_json()

        # THEN: JSON 형식 호환성 확인
        assert export_result.success is True
        assert self.temp_json.exists()

        with open(self.temp_json, 'r', encoding='utf-8') as f:
            exported_data = json.load(f)

        # 기존 JSON 구조와 완전 호환
        expected_keys = ["version", "updated", "statistics", "index", "references"]
        for key in expected_keys:
            assert key in exported_data

        # TAG 개수 검증
        assert exported_data["statistics"]["total_tags"] == 3

        # 인덱스 구조 검증
        index_data = exported_data["index"]
        assert "REQ:USER-LOGIN-001" in index_data
        assert "DESIGN:AUTH-ARCH-001" in index_data
        assert "TASK:LOGIN-IMPL-001" in index_data

        # 각 TAG의 구조 검증 (기존 JSON 형식과 동일)
        req_entries = index_data["REQ:USER-LOGIN-001"]
        assert len(req_entries) == 1
        assert req_entries[0]["file"] == "/path/req.md"
        assert req_entries[0]["line"] == 10
        assert "context" in req_entries[0]

        # 참조 관계 검증
        references = exported_data["references"]
        assert "REQ:USER-LOGIN-001" in references
        assert "DESIGN:AUTH-ARCH-001" in references["REQ:USER-LOGIN-001"]

    def test_should_validate_data_integrity_during_migration(self):
        """
        Given: 마이그레이션할 데이터
        When: 데이터 무결성 검증을 수행할 때
        Then: 모든 데이터 일관성을 확인하고 문제 발생 시 롤백해야 함
        """
        # GIVEN: 문제가 있는 JSON 데이터 (일부러 손상된 데이터)
        corrupted_json_data = {
            "version": "0.1.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {
                "total_tags": 5,  # 실제와 다른 개수
                "categories": {}
            },
            "index": {
                "REQ:INVALID-TAG": [
                    {
                        "file": "nonexistent.md",
                        "line": -1,  # 잘못된 줄 번호
                        "context": "@REQ:INVALID-TAG"
                    }
                ],
                "INVALID:CATEGORY-001": [  # 존재하지 않는 카테고리
                    {
                        "file": "test.md",
                        "line": 1,
                        "context": "@INVALID:CATEGORY-001"
                    }
                ],
                "REQ:DUPLICATE-TAG": [
                    {
                        "file": "file1.md",
                        "line": 10,
                        "context": "@REQ:DUPLICATE-TAG"
                    },
                    {
                        "file": "file2.md",
                        "line": 10,
                        "context": "@REQ:DUPLICATE-TAG"
                    }
                ]
            },
            "references": {
                "REQ:INVALID-TAG": ["NONEXISTENT:TAG-001"]  # 존재하지 않는 참조
            }
        }

        with open(self.temp_json, 'w', encoding='utf-8') as f:
            json.dump(corrupted_json_data, f, indent=2, ensure_ascii=False)

        # WHEN: 데이터 검증과 함께 마이그레이션 시도
        migration_result = self.migration_tool.migrate_json_to_sqlite(
            validate_data=True,
            strict_mode=True
        )

        # THEN: 검증 실패로 인한 마이그레이션 중단
        assert migration_result.success is False
        assert len(migration_result.validation_errors) > 0

        # 구체적인 검증 오류 확인
        error_types = [error.error_type for error in migration_result.validation_errors]
        assert 'INVALID_CATEGORY' in error_types
        assert 'INVALID_LINE_NUMBER' in error_types
        assert 'BROKEN_REFERENCE' in error_types
        assert 'TAG_COUNT_MISMATCH' in error_types

        # SQLite 파일이 생성되지 않거나 롤백되었는지 확인
        assert not self.temp_db.exists() or self.migration_tool.get_database_tag_count() == 0

    def test_should_handle_large_dataset_migration_with_progress_reporting(self):
        """
        Given: 대용량 TAG 데이터셋 (1000+ TAGs)
        When: 마이그레이션을 수행할 때
        Then: 진행률 보고와 함께 안정적으로 처리되어야 함
        """
        # GIVEN: 대용량 JSON 데이터 생성
        large_json_data = {
            "version": "1.0.0",
            "updated": datetime.now().isoformat(),
            "statistics": {
                "total_tags": 2000,
                "categories": {
                    "Primary": 1000,
                    "Steering": 300,
                    "Implementation": 400,
                    "Quality": 300
                }
            },
            "index": {},
            "references": {}
        }

        # 2000개 TAG 생성
        categories = ['REQ', 'DESIGN', 'TASK', 'TEST', 'VISION', 'STRUCT', 'TECH', 'FEATURE', 'API', 'PERF']
        for i in range(2000):
            category = categories[i % len(categories)]
            tag_key = f"{category}:LARGE-DATASET-{i:05d}"

            large_json_data["index"][tag_key] = [
                {
                    "file": f"/large/dataset/file_{i // 100}.md",
                    "line": (i % 100) + 1,
                    "context": f"@{tag_key} 대용량 데이터셋 테스트 TAG {i}"
                }
            ]

            # 일부 참조 관계 생성 (10%만)
            if i % 10 == 0 and i > 0:
                prev_tag = f"{categories[(i-1) % len(categories)]}:LARGE-DATASET-{i-1:05d}"
                if prev_tag not in large_json_data["references"]:
                    large_json_data["references"][prev_tag] = []
                large_json_data["references"][prev_tag].append(tag_key)

        with open(self.temp_json, 'w', encoding='utf-8') as f:
            json.dump(large_json_data, f, indent=2)

        # WHEN: 진행률 콜백과 함께 마이그레이션
        progress_reports = []

        def progress_callback(progress: MigrationProgress):
            progress_reports.append({
                'stage': progress.current_stage,
                'completed': progress.completed_items,
                'total': progress.total_items,
                'percentage': progress.percentage,
                'elapsed_time': progress.elapsed_time
            })

        start_time = time.time()
        migration_result = self.migration_tool.migrate_json_to_sqlite(
            progress_callback=progress_callback,
            batch_size=100  # 배치 처리로 메모리 효율성 확보
        )
        total_time = time.time() - start_time

        # THEN: 성공적인 대용량 처리 확인
        assert migration_result.success is True
        assert migration_result.migrated_tags_count == 2000
        assert migration_result.migrated_references_count == 200  # 10%

        # 성능 요구사항 확인 (2000개 TAG를 30초 내 처리)
        assert total_time < 30.0, f"마이그레이션 시간 {total_time:.2f}초가 목표 30초를 초과"

        # 진행률 보고 확인
        assert len(progress_reports) > 0

        # 진행률이 점진적으로 증가했는지 확인
        percentages = [report['percentage'] for report in progress_reports]
        assert percentages[0] < percentages[-1]  # 증가하는 패턴
        assert percentages[-1] == 100  # 완료 시 100%

        # 마이그레이션 단계별 보고 확인
        stages = set(report['stage'] for report in progress_reports)
        expected_stages = {'ANALYZING', 'MIGRATING_TAGS', 'MIGRATING_REFERENCES', 'VALIDATING'}
        assert stages.intersection(expected_stages) == expected_stages

    def test_should_create_and_restore_backups_during_migration(self):
        """
        Given: 기존 데이터와 마이그레이션 요청
        When: 마이그레이션 중 백업을 생성하고 문제 발생 시
        Then: 자동으로 백업에서 복구되어야 함
        """
        # GIVEN: 기존 데이터 준비
        original_json_data = {
            "version": "1.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {"total_tags": 1, "categories": {}},
            "index": {
                "REQ:ORIGINAL-001": [
                    {
                        "file": "original.md",
                        "line": 1,
                        "context": "@REQ:ORIGINAL-001 원본 데이터"
                    }
                ]
            },
            "references": {}
        }

        with open(self.temp_json, 'w', encoding='utf-8') as f:
            json.dump(original_json_data, f, indent=2)

        # WHEN: 백업 생성과 함께 마이그레이션
        backup_manager = BackupManager(self.backup_dir)
        backup_info = backup_manager.create_backup(
            json_path=self.temp_json,
            description="마이그레이션 전 백업"
        )

        # 마이그레이션 중 의도적 실패 시뮬레이션
        with patch.object(self.migration_tool, '_validate_migration_result') as mock_validate:
            mock_validate.side_effect = Exception("마이그레이션 검증 실패")

            migration_result = self.migration_tool.migrate_json_to_sqlite(
                create_backup=True,
                auto_rollback=True
            )

        # THEN: 실패 후 자동 복구 확인
        assert migration_result.success is False
        assert migration_result.backup_created is True
        assert migration_result.rollback_performed is True

        # 백업에서 복구되었는지 확인
        with open(self.temp_json, 'r', encoding='utf-8') as f:
            restored_data = json.load(f)

        assert restored_data == original_json_data

        # 백업 파일 존재 확인
        assert backup_info.backup_file.exists()
        assert backup_info.metadata_file.exists()

        # 백업 메타데이터 확인
        with open(backup_info.metadata_file, 'r') as f:
            backup_metadata = json.load(f)

        assert backup_metadata['description'] == "마이그레이션 전 백업"
        assert 'created_at' in backup_metadata
        assert 'original_file_size' in backup_metadata

    def test_should_support_incremental_migration_for_updates(self):
        """
        Given: 이미 마이그레이션된 SQLite 데이터베이스
        When: 새로운 TAG가 추가된 JSON을 증분 마이그레이션할 때
        Then: 기존 데이터는 유지하고 새 데이터만 추가되어야 함
        """
        # GIVEN: 초기 데이터로 첫 마이그레이션
        initial_json = {
            "version": "1.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {"total_tags": 2, "categories": {}},
            "index": {
                "REQ:INITIAL-001": [{"file": "req.md", "line": 1, "context": "@REQ:INITIAL-001"}],
                "DESIGN:INITIAL-001": [{"file": "design.md", "line": 1, "context": "@DESIGN:INITIAL-001"}]
            },
            "references": {"REQ:INITIAL-001": ["DESIGN:INITIAL-001"]}
        }

        with open(self.temp_json, 'w') as f:
            json.dump(initial_json, f)

        # 초기 마이그레이션
        self.migration_tool.migrate_json_to_sqlite()

        # 업데이트된 데이터 (새 TAG 추가)
        updated_json = {
            "version": "1.1",
            "updated": "2024-01-02T00:00:00Z",
            "statistics": {"total_tags": 4, "categories": {}},
            "index": {
                "REQ:INITIAL-001": [{"file": "req.md", "line": 1, "context": "@REQ:INITIAL-001"}],
                "DESIGN:INITIAL-001": [{"file": "design.md", "line": 1, "context": "@DESIGN:INITIAL-001"}],
                "TASK:NEW-001": [{"file": "task.md", "line": 1, "context": "@TASK:NEW-001 새 작업"}],
                "TEST:NEW-001": [{"file": "test.md", "line": 1, "context": "@TEST:NEW-001 새 테스트"}]
            },
            "references": {
                "REQ:INITIAL-001": ["DESIGN:INITIAL-001"],
                "DESIGN:INITIAL-001": ["TASK:NEW-001"],
                "TASK:NEW-001": ["TEST:NEW-001"]
            }
        }

        with open(self.temp_json, 'w') as f:
            json.dump(updated_json, f)

        # WHEN: 증분 마이그레이션 수행
        incremental_result = self.migration_tool.migrate_json_to_sqlite(
            mode='incremental',
            preserve_existing=True
        )

        # THEN: 증분 업데이트 성공 확인
        assert incremental_result.success is True
        assert incremental_result.migrated_tags_count == 2  # 새로 추가된 TAG만
        assert incremental_result.preserved_tags_count == 2  # 기존 TAG 유지

        # SQLite 데이터 검증
        db_manager = TagDatabaseManager(self.temp_db)
        all_tags = db_manager.get_all_tags()
        assert len(all_tags) == 4

        # 기존 TAG가 유지되었는지 확인
        initial_tags = db_manager.search_tags_by_pattern('INITIAL')
        assert len(initial_tags) == 2

        # 새 TAG가 추가되었는지 확인
        new_tags = db_manager.search_tags_by_pattern('NEW')
        assert len(new_tags) == 2

    def test_should_detect_and_resolve_migration_conflicts(self):
        """
        Given: 충돌이 발생할 수 있는 마이그레이션 시나리오
        When: 동일한 TAG에 대해 다른 데이터가 존재할 때
        Then: 충돌을 감지하고 해결 전략을 제시해야 함
        """
        # GIVEN: SQLite에 기존 데이터
        db_manager = TagDatabaseManager(self.temp_db)
        db_manager.initialize()

        existing_tag_id = db_manager.insert_tag(
            category='REQ',
            identifier='CONFLICT-001',
            description='기존 설명',
            file_path='/old/path.md',
            line_number=10
        )

        # JSON에 충돌되는 데이터
        conflicting_json = {
            "version": "1.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {"total_tags": 1, "categories": {}},
            "index": {
                "REQ:CONFLICT-001": [
                    {
                        "file": "/new/path.md",  # 다른 파일 경로
                        "line": 20,  # 다른 줄 번호
                        "context": "@REQ:CONFLICT-001 새로운 설명"  # 다른 설명
                    }
                ]
            },
            "references": {}
        }

        with open(self.temp_json, 'w') as f:
            json.dump(conflicting_json, f)

        # WHEN: 충돌 해결 전략과 함께 마이그레이션
        migration_result = self.migration_tool.migrate_json_to_sqlite(
            mode='incremental',
            conflict_resolution='interactive'  # 대화형 충돌 해결
        )

        # THEN: 충돌 감지 및 해결
        assert migration_result.conflicts_detected > 0
        assert len(migration_result.conflict_resolutions) > 0

        # 충돌 해결 상세 정보 확인
        conflict_info = migration_result.conflict_resolutions[0]
        assert conflict_info.tag_identifier == 'CONFLICT-001'
        assert conflict_info.conflict_type == 'DATA_MISMATCH'
        assert conflict_info.existing_data is not None
        assert conflict_info.new_data is not None

        # 해결 전략 옵션 확인
        assert 'KEEP_EXISTING' in conflict_info.resolution_options
        assert 'USE_NEW' in conflict_info.resolution_options
        assert 'MERGE' in conflict_info.resolution_options

    def test_should_provide_detailed_migration_statistics_and_reporting(self):
        """
        Given: 복잡한 TAG 데이터셋 마이그레이션
        When: 마이그레이션 완료 후
        Then: 상세한 통계와 리포트를 제공해야 함
        """
        # GIVEN: 다양한 종류의 데이터 준비
        complex_json = {
            "version": "2.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {"total_tags": 10, "categories": {}},
            "index": {},
            "references": {}
        }

        # 다양한 카테고리와 패턴의 TAG 생성
        categories = ['REQ', 'DESIGN', 'TASK', 'TEST', 'VISION', 'STRUCT', 'TECH', 'FEATURE', 'API', 'PERF']
        for i, category in enumerate(categories):
            tag_key = f"{category}:STATS-TEST-{i:03d}"
            complex_json["index"][tag_key] = [
                {
                    "file": f"/stats/{category.lower()}/file_{i}.md",
                    "line": (i * 10) + 5,
                    "context": f"@{tag_key} 통계 테스트용 TAG {i}"
                }
            ]

            # 체인 참조 생성 (순환 참조)
            if i > 0:
                prev_tag = f"{categories[i-1]}:STATS-TEST-{i-1:03d}"
                complex_json["references"][prev_tag] = [tag_key]

        with open(self.temp_json, 'w') as f:
            json.dump(complex_json, f)

        # WHEN: 상세 리포팅과 함께 마이그레이션
        migration_result = self.migration_tool.migrate_json_to_sqlite(
            detailed_reporting=True,
            generate_report=True
        )

        # THEN: 상세 통계 및 리포트 확인
        assert migration_result.success is True

        # 기본 통계
        assert migration_result.migrated_tags_count == 10
        assert migration_result.migrated_references_count == 9

        # 카테고리별 통계
        category_stats = migration_result.category_statistics
        assert len(category_stats) == 10
        for category in categories:
            assert category in category_stats
            assert category_stats[category]['tag_count'] == 1

        # 파일별 통계
        file_stats = migration_result.file_statistics
        assert len(file_stats) == 10  # 각 카테고리마다 하나의 파일

        # 참조 체인 분석
        chain_analysis = migration_result.reference_chain_analysis
        assert chain_analysis['total_chains'] == 1
        assert chain_analysis['longest_chain_length'] == 9
        assert chain_analysis['circular_references'] == 0

        # 성능 메트릭
        performance_metrics = migration_result.performance_metrics
        assert 'total_duration' in performance_metrics
        assert 'tags_per_second' in performance_metrics
        assert 'memory_usage_peak' in performance_metrics

        # HTML 리포트 생성 확인
        if migration_result.report_file:
            assert migration_result.report_file.exists()
            assert migration_result.report_file.suffix == '.html'

    def test_should_support_custom_migration_plugins(self):
        """
        Given: 사용자 정의 마이그레이션 로직이 필요한 상황
        When: 플러그인 시스템을 통해 커스텀 처리를 등록할 때
        Then: 플러그인이 마이그레이션 과정에 올바르게 통합되어야 함
        """
        # GIVEN: 커스텀 마이그레이션 플러그인
        class CustomTagProcessor:
            def process_tag(self, tag_data):
                # 특정 패턴의 TAG에 대해 추가 처리
                if 'CUSTOM' in tag_data['identifier']:
                    tag_data['description'] = f"[CUSTOM] {tag_data['description']}"
                    tag_data['metadata'] = {'processed': True, 'processor': 'CustomTagProcessor'}
                return tag_data

            def validate_tag(self, tag_data):
                # 커스텀 검증 로직
                if tag_data['category'] == 'CUSTOM':
                    return len(tag_data['identifier']) > 10
                return True

        # 테스트 데이터
        plugin_test_json = {
            "version": "1.0",
            "updated": "2024-01-01T00:00:00Z",
            "statistics": {"total_tags": 3, "categories": {}},
            "index": {
                "CUSTOM:CUSTOM-TAG-001": [{"file": "custom.md", "line": 1, "context": "@CUSTOM:CUSTOM-TAG-001 커스텀 TAG"}],
                "REQ:NORMAL-TAG-001": [{"file": "normal.md", "line": 1, "context": "@REQ:NORMAL-TAG-001 일반 TAG"}],
                "CUSTOM:SHORT": [{"file": "invalid.md", "line": 1, "context": "@CUSTOM:SHORT 짧은 식별자"}]  # 검증 실패 예정
            },
            "references": {}
        }

        with open(self.temp_json, 'w') as f:
            json.dump(plugin_test_json, f)

        # WHEN: 플러그인과 함께 마이그레이션
        custom_processor = CustomTagProcessor()
        migration_result = self.migration_tool.migrate_json_to_sqlite(
            plugins=[custom_processor],
            strict_validation=True
        )

        # THEN: 플러그인 처리 결과 확인
        assert migration_result.success is True
        assert migration_result.migrated_tags_count == 2  # 하나는 검증 실패로 제외
        assert migration_result.plugin_processed_count == 1  # CUSTOM-TAG-001만 처리
        assert migration_result.validation_failed_count == 1  # SHORT는 검증 실패

        # SQLite에서 플러그인 처리 결과 확인
        db_manager = TagDatabaseManager(self.temp_db)
        custom_tags = db_manager.search_tags_by_category('CUSTOM')

        processed_tag = None
        for tag in custom_tags:
            if 'CUSTOM-TAG-001' in tag['identifier']:
                processed_tag = tag
                break

        assert processed_tag is not None
        assert processed_tag['description'] == "[CUSTOM] 커스텀 TAG"
        # 메타데이터는 별도 테이블에 저장될 수 있음
        metadata = db_manager.get_tag_metadata(processed_tag['id'])
        assert metadata['processed'] is True