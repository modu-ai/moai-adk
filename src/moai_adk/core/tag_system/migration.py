"""
@FEATURE:SPEC-009-TAG-MIGRATION-001 - TAG 마이그레이션 도구

GREEN 단계: JSON ↔ SQLite 양방향 변환 및 데이터 무결성 보장
@DESIGN:REFACTORED-MIGRATION-001 Rebuilt from 644 LOC to modular architecture (< 200 LOC)
"""

import json
import time
from collections.abc import Callable
from pathlib import Path

from .database import TagDatabaseManager
from .migration_engine import MigrationEngine
from .migration_models import (
    BackupManager,
    MigrationProgress,
    MigrationResult,
)
from .migration_validator import MigrationValidator


class TagMigrationTool:
    """
    TAG 마이그레이션 오케스트레이터

    TRUST 원칙 적용:
    - Test First: 테스트 요구사항에 맞춘 최소 구현
    - Readable: 명확한 마이그레이션 로직
    - Unified: 마이그레이션 오케스트레이션 책임만 담당
    @DESIGN:REFACTORED-ORCHESTRATOR-001 Now coordinates specialized modules
    """

    def __init__(self, database_path: Path, json_path: Path, backup_directory: Path | None = None):
        self.database_path = database_path
        self.json_path = json_path

        # 백업 디렉토리 설정
        if backup_directory:
            self.backup_directory = Path(backup_directory)
        else:
            self.backup_directory = json_path.parent / ".moai" / "backups"

        # 전용 모듈들 초기화
        self.backup_manager = BackupManager(self.backup_directory)
        self.engine = MigrationEngine(database_path, json_path, self.backup_manager)
        self.validator = MigrationValidator(json_path, self.backup_directory)

    def migrate_json_to_sqlite(self,
                              validate_data: bool = False,
                              strict_mode: bool = False,
                              progress_callback: Callable[[MigrationProgress], None] | None = None,
                              batch_size: int = 1000,
                              mode: str = 'full',
                              preserve_existing: bool = False,
                              conflict_resolution: str = 'keep_existing',
                              create_backup: bool = False,
                              auto_rollback: bool = False,
                              detailed_reporting: bool = False,
                              generate_report: bool = False,
                              plugins: list | None = None,
                              strict_validation: bool = False) -> MigrationResult:
        """JSON에서 SQLite로 마이그레이션 (오케스트레이터)"""

        start_time = time.time()
        result = MigrationResult(success=False)

        try:
            # 백업 생성
            backup_info = None
            if create_backup and self.json_path.exists():
                backup_info = self.backup_manager.create_backup(
                    self.json_path,
                    "마이그레이션 전 자동 백업"
                )
                result.backup_created = True

            # JSON 데이터 로드
            if not self.json_path.exists():
                result.errors.append("JSON 파일이 존재하지 않습니다")
                return result

            with open(self.json_path, encoding='utf-8') as f:
                json_data = json.load(f)

            # 데이터 검증
            if validate_data:
                validation_result = self.validator.validate_json_data(json_data, strict_mode)
                result.validation_errors = validation_result
                if strict_mode and validation_result:
                    result.errors.append(f"검증 실패: {len(validation_result)}개 오류")
                    return result

            # SQLite 데이터베이스 초기화
            db_manager = TagDatabaseManager(self.database_path)
            db_manager.initialize()

            # 마이그레이션 엔진 실행
            result = self.engine.migrate_json_to_sqlite_core(
                json_data, db_manager, mode, batch_size, progress_callback
            )

            # 상세 보고서 생성
            if detailed_reporting:
                self.validator.generate_detailed_statistics(result, db_manager)

            if generate_report:
                result.report_file = self.validator.generate_html_report(result)

            # 성능 메트릭 추가
            result.performance_metrics = {
                'total_duration': time.time() - start_time
            }

        except Exception as e:
            result.errors.append(str(e))
            if auto_rollback and create_backup and backup_info:
                try:
                    self.validator.perform_rollback(backup_info)
                    result.rollback_performed = True
                except:
                    pass  # 롤백 실패는 무시

        return result

    def migrate_sqlite_to_json(self) -> MigrationResult:
        """SQLite에서 JSON으로 마이그레이션 (델리게이트)"""
        return self.engine.migrate_sqlite_to_json()

    def get_database_tag_count(self) -> int:
        """데이터베이스 TAG 개수 반환"""
        try:
            if self.database_path.exists():
                db_manager = TagDatabaseManager(self.database_path)
                db_manager.initialize()
                all_tags = db_manager.get_all_tags()
                return len(all_tags)
            return 0
        except Exception:
            return 0

    def validate_migration_result(self):
        """마이그레이션 결과 검증 (델리게이트)"""
        return self.validator.validate_migration_result()
