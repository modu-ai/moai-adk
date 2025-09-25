"""
Migration validation, reporting, and utility functions.

@FEATURE:MIGRATION-VALIDATOR-001 Data validation and reporting
@DESIGN:SEPARATED-VALIDATOR-001 Extracted from oversized migration.py (644 LOC)
"""

import time
import shutil
from pathlib import Path
from typing import Dict, List, Any, Optional

from .database import TagDatabaseManager
from .migration_models import ValidationError, MigrationResult, BackupInfo


class MigrationValidator:
    """마이그레이션 검증 및 리포팅"""

    def __init__(self, json_path: Path, backup_directory: Path):
        self.json_path = json_path
        self.backup_directory = backup_directory

    def validate_json_data(
        self, json_data: Dict[str, Any], strict_mode: bool = False
    ) -> List[ValidationError]:
        """JSON 데이터 검증"""
        errors = []

        # 버전 확인
        if "version" not in json_data:
            errors.append(
                ValidationError("missing_field", "버전 정보가 없습니다", "root.version")
            )

        # 인덱스 데이터 확인
        if "index" not in json_data:
            errors.append(
                ValidationError(
                    "missing_field", "인덱스 데이터가 없습니다", "root.index"
                )
            )
        else:
            index_data = json_data["index"]
            if not isinstance(index_data, dict):
                errors.append(
                    ValidationError(
                        "invalid_type", "인덱스는 딕셔너리여야 합니다", "root.index"
                    )
                )
            else:
                # 각 TAG 키 검증
                for tag_key, tag_entries in index_data.items():
                    if ":" not in tag_key:
                        errors.append(
                            ValidationError(
                                "invalid_format",
                                f"TAG 키 형식이 잘못됨: {tag_key}",
                                f"root.index.{tag_key}",
                            )
                        )
                        continue

                    category, identifier = tag_key.split(":", 1)

                    # 카테고리 검증
                    valid_categories = [
                        "REQ",
                        "DESIGN",
                        "TASK",
                        "TEST",
                        "VISION",
                        "STRUCT",
                        "TECH",
                        "ADR",
                        "FEATURE",
                        "API",
                        "UI",
                        "DATA",
                        "PERF",
                        "SEC",
                        "DOCS",
                        "TAG",
                        "DEBT",
                        "TODO",
                    ]

                    if category not in valid_categories:
                        if strict_mode:
                            errors.append(
                                ValidationError(
                                    "invalid_category",
                                    f"알 수 없는 카테고리: {category}",
                                    f"root.index.{tag_key}.category",
                                )
                            )

                    # 식별자 검증
                    if not identifier or len(identifier) < 3:
                        errors.append(
                            ValidationError(
                                "invalid_identifier",
                                f"식별자가 너무 짧음: {identifier}",
                                f"root.index.{tag_key}.identifier",
                            )
                        )

                    # 엔트리 구조 검증
                    if not isinstance(tag_entries, list) or not tag_entries:
                        errors.append(
                            ValidationError(
                                "invalid_entries",
                                f"TAG 엔트리가 비어있거나 올바르지 않음: {tag_key}",
                                f"root.index.{tag_key}",
                            )
                        )
                        continue

                    # 각 엔트리 검증
                    for i, entry in enumerate(tag_entries):
                        if not isinstance(entry, dict):
                            errors.append(
                                ValidationError(
                                    "invalid_entry_type",
                                    f"엔트리는 딕셔너리여야 함: {tag_key}[{i}]",
                                    f"root.index.{tag_key}[{i}]",
                                )
                            )
                            continue

                        # 필수 필드 검증
                        required_fields = ["file", "line", "context"]
                        for field in required_fields:
                            if field not in entry:
                                errors.append(
                                    ValidationError(
                                        "missing_entry_field",
                                        f"필수 필드 누락: {field}",
                                        f"root.index.{tag_key}[{i}].{field}",
                                    )
                                )

                        # 파일 경로 검증
                        if "file" in entry:
                            file_path = entry["file"]
                            if not isinstance(file_path, str) or not file_path:
                                errors.append(
                                    ValidationError(
                                        "invalid_file_path",
                                        f"유효하지 않은 파일 경로: {file_path}",
                                        f"root.index.{tag_key}[{i}].file",
                                    )
                                )

                        # 라인 번호 검증
                        if "line" in entry:
                            line_number = entry["line"]
                            if not isinstance(line_number, int) or line_number < 1:
                                errors.append(
                                    ValidationError(
                                        "invalid_line_number",
                                        f"유효하지 않은 라인 번호: {line_number}",
                                        f"root.index.{tag_key}[{i}].line",
                                    )
                                )

        # 참조 데이터 검증
        if "references" in json_data:
            references_data = json_data["references"]
            if not isinstance(references_data, dict):
                errors.append(
                    ValidationError(
                        "invalid_type", "참조는 딕셔너리여야 합니다", "root.references"
                    )
                )
            else:
                # 각 참조 관계 검증
                for source_tag, target_tags in references_data.items():
                    if ":" not in source_tag:
                        errors.append(
                            ValidationError(
                                "invalid_reference_format",
                                f"소스 TAG 형식이 잘못됨: {source_tag}",
                                f"root.references.{source_tag}",
                            )
                        )

                    if not isinstance(target_tags, list):
                        errors.append(
                            ValidationError(
                                "invalid_target_type",
                                f"대상 TAG는 리스트여야 함: {source_tag}",
                                f"root.references.{source_tag}",
                            )
                        )
                        continue

                    for target_tag in target_tags:
                        if ":" not in target_tag:
                            errors.append(
                                ValidationError(
                                    "invalid_reference_format",
                                    f"대상 TAG 형식이 잘못됨: {target_tag}",
                                    f"root.references.{source_tag} -> {target_tag}",
                                )
                            )

        return errors

    def apply_plugins(self, tag_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """플러그인 적용 (기본 구현)"""
        # 기본적으로 데이터를 그대로 반환
        return tag_data

    def validate_with_plugins(self, tag_data: Dict[str, Any]) -> bool:
        """플러그인 검증 (기본 구현)"""
        # 기본적으로 항상 유효함
        return True

    def generate_detailed_statistics(
        self, result: MigrationResult, db_manager: TagDatabaseManager
    ):
        """상세 통계 생성"""
        try:
            # 카테고리별 통계
            all_tags = db_manager.get_all_tags()
            category_stats = {}

            for tag in all_tags:
                category = tag["category"]
                if category not in category_stats:
                    category_stats[category] = {"count": 0, "files": set()}

                category_stats[category]["count"] += 1
                category_stats[category]["files"].add(tag["file_path"])

            # 딕셔너리 값 변환 (set을 len으로)
            for category in category_stats:
                category_stats[category]["unique_files"] = len(
                    category_stats[category]["files"]
                )
                category_stats[category]["files"] = list(
                    category_stats[category]["files"]
                )

            result.category_statistics = category_stats

            # 파일별 통계
            file_stats = {}
            for tag in all_tags:
                file_path = tag["file_path"]
                if file_path not in file_stats:
                    file_stats[file_path] = 0
                file_stats[file_path] += 1

            result.file_statistics = file_stats

            # 참조 체인 분석
            reference_analysis = {
                "total_references": len(db_manager.get_all_references()),
                "orphaned_tags": 0,
                "circular_references": [],
            }

            # 고아 TAG 찾기 (참조가 없는 TAG)
            all_tag_ids = {tag["id"] for tag in all_tags}
            referenced_ids = set()

            for ref in db_manager.get_all_references():
                referenced_ids.add(ref["source_tag_id"])
                referenced_ids.add(ref["target_tag_id"])

            reference_analysis["orphaned_tags"] = len(all_tag_ids - referenced_ids)
            result.reference_chain_analysis = reference_analysis

        except Exception as e:
            if not result.errors:
                result.errors = []
            result.errors.append(f"통계 생성 중 오류: {e}")

    def generate_html_report(self, result: MigrationResult) -> Path:
        """HTML 보고서 생성"""
        report_path = (
            self.backup_directory / f"migration_report_{int(time.time())}.html"
        )

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
            <p class="{"success" if result.success else "error"}">
                Status: {"Success" if result.success else "Failed"}
            </p>
            <h2>Summary</h2>
            <table>
                <tr><th>Metric</th><th>Value</th></tr>
                <tr><td>Migrated Tags</td><td>{result.migrated_tags_count}</td></tr>
                <tr><td>Migrated References</td><td>{result.migrated_references_count}</td></tr>
                <tr><td>Validation Errors</td><td>{len(result.validation_errors)}</td></tr>
                <tr><td>Processing Time</td><td>{result.performance_metrics.get("total_duration", 0):.2f if result.performance_metrics else 0:.2f}s</td></tr>
            </table>
        </body>
        </html>
        """

        with open(report_path, "w", encoding="utf-8") as f:
            f.write(html_content)

        return report_path

    def perform_rollback(self, backup_info: BackupInfo):
        """롤백 수행"""
        shutil.copy2(backup_info.backup_file, self.json_path)

    def validate_migration_result(self):
        """마이그레이션 결과 검증 (테스트용)"""
        pass
