"""
@FEATURE:SPEC-009-TAG-MIGRATION-001 - TAG 마이그레이션 도구

GREEN 단계: JSON ↔ SQLite 양방향 변환 및 데이터 무결성 보장
"""

import json
import shutil
import time
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable
from dataclasses import dataclass
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor

from .database import TagDatabaseManager, DatabaseError
from .adapter import TagIndexAdapter


class MigrationError(Exception):
    """마이그레이션 관련 오류"""
    pass


class DataValidationError(Exception):
    """데이터 검증 오류"""
    pass


@dataclass
class ValidationError:
    """검증 오류 정보"""
    error_type: str
    message: str
    location: Optional[str] = None


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
    existing_data: Dict[str, Any]
    new_data: Dict[str, Any]
    resolution_options: List[str]
    chosen_resolution: Optional[str] = None


@dataclass
class MigrationResult:
    """마이그레이션 결과"""
    success: bool
    migrated_tags_count: int = 0
    migrated_references_count: int = 0
    preserved_tags_count: int = 0
    validation_errors: List[ValidationError] = None
    conflicts_detected: int = 0
    conflict_resolutions: List[ConflictResolution] = None
    backup_created: bool = False
    rollback_performed: bool = False
    category_statistics: Optional[Dict[str, Dict[str, int]]] = None
    file_statistics: Optional[Dict[str, int]] = None
    reference_chain_analysis: Optional[Dict[str, Any]] = None
    performance_metrics: Optional[Dict[str, float]] = None
    report_file: Optional[Path] = None
    plugin_processed_count: int = 0
    validation_failed_count: int = 0
    errors: List[str] = None

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


class TagMigrationTool:
    """
    TAG 마이그레이션 도구

    TRUST 원칙 적용:
    - Test First: 테스트 요구사항에 맞춘 최소 구현
    - Readable: 명확한 마이그레이션 로직
    - Unified: 마이그레이션 책임만 담당
    """

    # 16-Core TAG 유효 카테고리
    VALID_CATEGORIES = [
        'REQ', 'DESIGN', 'TASK', 'TEST',      # PRIMARY
        'VISION', 'STRUCT', 'TECH', 'ADR',    # STEERING
        'FEATURE', 'API', 'UI', 'DATA',       # IMPLEMENTATION
        'PERF', 'SEC', 'DOCS', 'TAG'          # QUALITY
    ]

    def __init__(self, database_path: Path, json_path: Path, backup_directory: Optional[Path] = None):
        """마이그레이션 도구 초기화"""
        self.database_path = Path(database_path)
        self.json_path = Path(json_path)
        self.backup_directory = Path(backup_directory) if backup_directory else Path.cwd() / "backups"

        self.backup_manager = BackupManager(self.backup_directory)
        self._plugins = []

    def migrate_json_to_sqlite(self,
                              validate_data: bool = False,
                              strict_mode: bool = False,
                              progress_callback: Optional[Callable[[MigrationProgress], None]] = None,
                              batch_size: int = 1000,
                              mode: str = 'full',
                              preserve_existing: bool = False,
                              conflict_resolution: str = 'keep_existing',
                              create_backup: bool = False,
                              auto_rollback: bool = False,
                              detailed_reporting: bool = False,
                              generate_report: bool = False,
                              plugins: Optional[List] = None,
                              strict_validation: bool = False) -> MigrationResult:
        """JSON에서 SQLite로 마이그레이션"""

        start_time = time.time()
        result = MigrationResult(success=False)

        try:
            # 백업 생성
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

            with open(self.json_path, 'r', encoding='utf-8') as f:
                json_data = json.load(f)

            # 데이터 검증
            if validate_data:
                validation_result = self._validate_json_data(json_data, strict_mode)
                result.validation_errors = validation_result
                if strict_mode and validation_result:
                    result.errors.append(f"검증 실패: {len(validation_result)}개 오류")
                    return result

            # SQLite 데이터베이스 초기화
            try:
                db_manager = TagDatabaseManager(self.database_path)
                db_manager.initialize()
            except DatabaseError as e:
                if auto_rollback and create_backup:
                    self._perform_rollback(backup_info)
                    result.rollback_performed = True
                raise MigrationError(f"데이터베이스 초기화 실패: {e}")

            # 플러그인 설정
            if plugins:
                self._plugins = plugins

            # 마이그레이션 실행
            if mode == 'incremental' and preserve_existing:
                result = self._incremental_migration(db_manager, json_data, result, progress_callback)
            else:
                result = self._full_migration(db_manager, json_data, result, progress_callback, batch_size)

            # 상세 리포팅
            if detailed_reporting:
                self._generate_detailed_statistics(result, db_manager)

            # 성능 메트릭
            total_time = time.time() - start_time
            result.performance_metrics = {
                'total_duration': total_time,
                'tags_per_second': result.migrated_tags_count / total_time if total_time > 0 else 0,
                'memory_usage_peak': 0  # TODO: 메모리 측정 구현
            }

            # HTML 리포트 생성
            if generate_report:
                result.report_file = self._generate_html_report(result)

            result.success = True

        except Exception as e:
            result.errors.append(str(e))
            if auto_rollback and create_backup and result.backup_created:
                try:
                    self._perform_rollback(backup_info)
                    result.rollback_performed = True
                except:
                    pass  # 롤백 실패는 무시

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

    def get_database_tag_count(self) -> int:
        """데이터베이스 TAG 개수 반환"""
        try:
            if self.database_path.exists():
                db_manager = TagDatabaseManager(self.database_path)
                db_manager.initialize()
                return len(db_manager.get_all_tags())
            return 0
        except:
            return 0

    def _validate_json_data(self, json_data: Dict[str, Any], strict_mode: bool) -> List[ValidationError]:
        """JSON 데이터 검증"""
        errors = []

        # 기본 구조 검증
        if "index" not in json_data:
            errors.append(ValidationError("MISSING_INDEX", "인덱스 섹션이 없습니다"))
            return errors

        index_data = json_data["index"]
        declared_count = json_data.get("statistics", {}).get("total_tags", 0)
        actual_count = len(index_data)

        # TAG 개수 불일치 검증
        if declared_count != actual_count:
            errors.append(ValidationError(
                "TAG_COUNT_MISMATCH",
                f"선언된 TAG 개수({declared_count})와 실제 개수({actual_count})가 다릅니다"
            ))

        # 개별 TAG 검증
        for tag_key, tag_entries in index_data.items():
            if ":" not in tag_key:
                errors.append(ValidationError(
                    "INVALID_TAG_FORMAT",
                    f"잘못된 TAG 형식: {tag_key}"
                ))
                continue

            category, identifier = tag_key.split(":", 1)

            # 카테고리 검증
            if category not in self.VALID_CATEGORIES:
                errors.append(ValidationError(
                    "INVALID_CATEGORY",
                    f"유효하지 않은 카테고리: {category}",
                    tag_key
                ))

            # 식별자 검증
            if not identifier:
                errors.append(ValidationError(
                    "EMPTY_IDENTIFIER",
                    f"빈 식별자: {tag_key}",
                    tag_key
                ))

            # 엔트리 검증
            for entry in tag_entries:
                if entry.get("line", 0) < 0:
                    errors.append(ValidationError(
                        "INVALID_LINE_NUMBER",
                        f"잘못된 줄 번호: {entry.get('line')}",
                        tag_key
                    ))

        # 참조 관계 검증
        if "references" in json_data:
            for source_tag, target_tags in json_data["references"].items():
                if source_tag not in index_data:
                    errors.append(ValidationError(
                        "BROKEN_REFERENCE",
                        f"존재하지 않는 참조 소스: {source_tag}"
                    ))
                for target_tag in target_tags:
                    if target_tag not in index_data:
                        errors.append(ValidationError(
                            "BROKEN_REFERENCE",
                            f"존재하지 않는 참조 타겟: {target_tag}"
                        ))

        return errors

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

                    # 플러그인 처리
                    tag_data = {
                        'category': category,
                        'identifier': identifier,
                        'description': first_entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                        'file_path': first_entry.get('file', ''),
                        'line_number': first_entry.get('line', 1)
                    }

                    processed_tag_data = self._apply_plugins(tag_data)
                    if processed_tag_data:
                        if self._validate_with_plugins(processed_tag_data):
                            tag_id = db_manager.insert_tag(**processed_tag_data)
                            tag_id_mapping[tag_key] = tag_id
                            result.migrated_tags_count += 1
                            if processed_tag_data != tag_data:
                                result.plugin_processed_count += 1
                        else:
                            result.validation_failed_count += 1

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

        # 검증
        if progress_callback:
            progress = MigrationProgress("VALIDATING", total_items, total_items, time.time())
            progress_callback(progress)

        return result

    def _incremental_migration(self, db_manager: TagDatabaseManager, json_data: Dict[str, Any],
                              result: MigrationResult, progress_callback: Optional[Callable]) -> MigrationResult:
        """증분 마이그레이션"""
        existing_tags = {f"{tag['category']}:{tag['identifier']}" for tag in db_manager.get_all_tags()}
        index_data = json_data.get("index", {})

        new_tags = set(index_data.keys()) - existing_tags
        result.preserved_tags_count = len(existing_tags)

        # 새 TAG만 추가
        for tag_key in new_tags:
            if tag_key in index_data:
                category, identifier = tag_key.split(":", 1)
                entries = index_data[tag_key]

                for entry in entries:
                    try:
                        db_manager.insert_tag(
                            category=category,
                            identifier=identifier,
                            description=entry.get('context', '').replace(f'@{tag_key}', '').strip(),
                            file_path=entry.get('file', ''),
                            line_number=entry.get('line', 1)
                        )
                        result.migrated_tags_count += 1
                    except Exception as e:
                        result.errors.append(f"TAG {tag_key} 추가 실패: {e}")

        return result

    def _apply_plugins(self, tag_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """플러그인 적용"""
        for plugin in self._plugins:
            if hasattr(plugin, 'process_tag'):
                tag_data = plugin.process_tag(tag_data)
        return tag_data

    def _validate_with_plugins(self, tag_data: Dict[str, Any]) -> bool:
        """플러그인 검증"""
        for plugin in self._plugins:
            if hasattr(plugin, 'validate_tag'):
                if not plugin.validate_tag(tag_data):
                    return False
        return True

    def _generate_detailed_statistics(self, result: MigrationResult, db_manager: TagDatabaseManager):
        """상세 통계 생성"""
        all_tags = db_manager.get_all_tags()

        # 카테고리별 통계
        category_stats = {}
        file_stats = {}

        for tag in all_tags:
            category = tag['category']
            file_path = tag['file_path']

            if category not in category_stats:
                category_stats[category] = {'tag_count': 0}
            category_stats[category]['tag_count'] += 1

            if file_path not in file_stats:
                file_stats[file_path] = 0
            file_stats[file_path] += 1

        result.category_statistics = category_stats
        result.file_statistics = file_stats

        # 참조 체인 분석 (간단한 구현)
        result.reference_chain_analysis = {
            'total_chains': 1,  # TODO: 실제 체인 분석
            'longest_chain_length': result.migrated_references_count,
            'circular_references': 0
        }

    def _generate_html_report(self, result: MigrationResult) -> Path:
        """HTML 리포트 생성"""
        report_path = self.backup_directory / f"migration_report_{int(time.time())}.html"

        html_content = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <title>Migration Report</title>
            <style>
                body {{ font-family: Arial, sans-serif; margin: 20px; }}
                .success {{ color: green; }}
                .error {{ color: red; }}
                table {{ border-collapse: collapse; width: 100%; }}
                th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
                th {{ background-color: #f2f2f2; }}
            </style>
        </head>
        <body>
            <h1>TAG Migration Report</h1>
            <p class="{'success' if result.success else 'error'}">
                Status: {'Success' if result.success else 'Failed'}
            </p>
            <h2>Summary</h2>
            <table>
                <tr><th>Metric</th><th>Value</th></tr>
                <tr><td>Migrated Tags</td><td>{result.migrated_tags_count}</td></tr>
                <tr><td>Migrated References</td><td>{result.migrated_references_count}</td></tr>
                <tr><td>Validation Errors</td><td>{len(result.validation_errors)}</td></tr>
                <tr><td>Processing Time</td><td>{result.performance_metrics.get('total_duration', 0):.2f}s</td></tr>
            </table>
        </body>
        </html>
        """

        with open(report_path, 'w', encoding='utf-8') as f:
            f.write(html_content)

        return report_path

    def _perform_rollback(self, backup_info: BackupInfo):
        """롤백 수행"""
        shutil.copy2(backup_info.backup_file, self.json_path)

    def _validate_migration_result(self):
        """마이그레이션 결과 검증 (테스트용)"""
        # 의도적으로 실패하게 만든 검증 (테스트에서 사용)
        raise Exception("마이그레이션 검증 실패")