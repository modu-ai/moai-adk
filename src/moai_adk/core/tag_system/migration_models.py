"""
Migration data models and backup management.

@FEATURE:MIGRATION-MODELS-001 Data structures and backup utilities
@DESIGN:SEPARATED-MODELS-001 Extracted from oversized migration.py (644 LOC)
"""

import json
import shutil
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


class MigrationError(Exception):
    """마이그레이션 관련 오류"""


class DataValidationError(Exception):
    """데이터 검증 오류"""


@dataclass
class ValidationError:
    """검증 오류 정보"""
    error_type: str
    message: str
    location: str | None = None


@dataclass
class MigrationProgress:
    """마이그레이션 진행률 정보"""
    current_stage: str
    completed_items: int
    total_items: int
    elapsed_time: float

    @property
    def percentage(self) -> float:
        """진행률 퍼센트"""
        if self.total_items == 0:
            return 100.0
        return (self.completed_items / self.total_items) * 100.0


@dataclass
class ConflictResolution:
    """충돌 해결 정보"""
    tag_identifier: str
    conflict_type: str
    existing_data: dict[str, Any]
    new_data: dict[str, Any]
    resolution_options: list[str]
    chosen_resolution: str | None = None


@dataclass
class MigrationResult:
    """마이그레이션 결과"""
    success: bool
    migrated_tags_count: int = 0
    migrated_references_count: int = 0
    preserved_tags_count: int = 0
    validation_errors: list[ValidationError] = None
    conflicts_detected: int = 0
    conflict_resolutions: list[ConflictResolution] = None
    backup_created: bool = False
    rollback_performed: bool = False
    category_statistics: dict[str, dict[str, int]] | None = None
    file_statistics: dict[str, int] | None = None
    reference_chain_analysis: dict[str, Any] | None = None
    performance_metrics: dict[str, float] | None = None
    report_file: Path | None = None
    plugin_processed_count: int = 0
    validation_failed_count: int = 0
    errors: list[str] = None

    def __post_init__(self):
        if self.validation_errors is None:
            self.validation_errors = []
        if self.conflict_resolutions is None:
            self.conflict_resolutions = []
        if self.errors is None:
            self.errors = []


@dataclass
class BackupInfo:
    """백업 정보"""
    backup_file: Path
    metadata_file: Path
    created_at: datetime


class BackupManager:
    """백업 관리자"""

    def __init__(self, backup_directory: Path):
        self.backup_directory = Path(backup_directory)
        self.backup_directory.mkdir(parents=True, exist_ok=True)

    def create_backup(self, json_path: Path, description: str = "") -> BackupInfo:
        """백업 생성"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = f"tags_backup_{timestamp}"

        backup_file = self.backup_directory / f"{backup_name}.json"
        metadata_file = self.backup_directory / f"{backup_name}_meta.json"

        # 백업 파일 복사
        shutil.copy2(json_path, backup_file)

        # 메타데이터 생성
        metadata = {
            "original_file": str(json_path),
            "backup_file": str(backup_file),
            "created_at": datetime.now().isoformat(),
            "description": description,
            "original_file_size": json_path.stat().st_size
        }

        with open(metadata_file, 'w', encoding='utf-8') as f:
            json.dump(metadata, f, indent=2)

        return BackupInfo(
            backup_file=backup_file,
            metadata_file=metadata_file,
            created_at=datetime.now()
        )
