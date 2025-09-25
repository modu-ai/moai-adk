#!/usr/bin/env python3
# @TASK:TAG-VALIDATE-011
"""
MoAI-ADK Tag System Validator v0.1.12
16-Core @TAG 무결성 검사 및 추적성 매트릭스 검증

이 스크립트는 프로젝트 전체의 @TAG 시스템을:
- 16-Core 태그 체계 준수 검증
- 고아 태그 및 연결 끊김 감지
- 태그 인덱스 일관성 확인 (SQLite 백엔드)
- 추적성 매트릭스 업데이트
- 태그 품질 점수 계산

⚠️  NOTE: 이 스크립트는 SQLite 전용입니다. JSON 호환성은 완전히 제거되었습니다.
"""

import sqlite3
import re
import sys
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Set, Any


@dataclass
class TagReference:
    """태그 참조 정보"""

    tag_type: str
    tag_id: str
    file_path: str
    line_number: int
    context: str


@dataclass
class TagHealthReport:
    """태그 건강도 리포트"""

    total_tags: int = 0
    valid_tags: int = 0
    invalid_tags: int = 0
    orphan_tags: int = 0
    broken_links: int = 0
    quality_score: float = 0.0
    issues: list[str] = field(default_factory=list)
    recommendations: list[str] = field(default_factory=list)


class TagValidator:
    """16-Core TAG 시스템 검증기"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"
        self.indexes_dir = self.moai_dir / "indexes"

        # 16-Core 태그 체계
        self.tag_categories = {
            "Primary": ["REQ", "SPEC", "DESIGN", "TASK", "TEST"],
            "Steering": ["VISION", "STRUCT", "TECH", "ADR"],
            "Implementation": ["FEATURE", "API", "DATA"],
            "Quality": ["PERF", "SEC", "DEBT", "TODO"],
            "Legacy": ["US", "FR", "NFR", "BUG", "REVIEW"],
        }

        self.valid_tag_types = []
        for category_tags in self.tag_categories.values():
            self.valid_tag_types.extend(category_tags)

        # 추적성 체인 정의
        self.traceability_chains = {
            "Primary": ["REQ", "DESIGN", "TASK", "TEST"],
            "Development": ["SPEC", "ADR", "TASK", "API", "TEST"],
            "Quality": ["PERF", "SEC", "DEBT", "REVIEW"],
        }

        # 스캔 결과
        self.all_tags: list[TagReference] = []
        self.tag_index: dict[str, list[TagReference]] = {}
        self.violations: list[str] = []

    def scan_project_files(self) -> list[TagReference]:
        """프로젝트 파일에서 모든 태그 스캔"""

        tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-_]+)(?:\s+"([^"]*)")?'
        found_tags = []

        # 스캔할 파일 확장자 (JSON 제외)
        scan_extensions = [".md", ".py", ".js", ".ts", ".tsx", ".jsx", ".yml", ".yaml"]

        # 제외할 디렉토리
        exclude_dirs = {
            "node_modules",
            "__pycache__",
            ".git",
            "dist",
            "build",
            "venv",
            ".env",
        }

        for file_path in self.project_root.rglob("*"):
            if file_path.is_file() and file_path.suffix in scan_extensions:
                # 제외 디렉토리 확인
                if any(excluded in file_path.parts for excluded in exclude_dirs):
                    continue

                try:
                    content = file_path.read_text(encoding="utf-8", errors="ignore")
                    lines = content.split("\n")

                    for line_num, line in enumerate(lines, 1):
                        matches = re.finditer(tag_pattern, line)

                        for match in matches:
                            tag_type = match.group(1)
                            tag_id = match.group(2)
                            description = match.group(3) or ""

                            # 상대 경로로 변환
                            rel_path = file_path.relative_to(self.project_root)

                            tag_ref = TagReference(
                                tag_type=tag_type,
                                tag_id=tag_id,
                                file_path=str(rel_path),
                                line_number=line_num,
                                context=line.strip(),
                            )

                            found_tags.append(tag_ref)

                except Exception as error:
                    print(f"Warning: Could not scan {file_path}: {error}")

        return found_tags

    def build_tag_index(
        self, tags: list[TagReference]
    ) -> dict[str, list[TagReference]]:
        """태그 인덱스 구축"""
        index = {}

        for tag in tags:
            tag_key = f"{tag.tag_type}:{tag.tag_id}"
            if tag_key not in index:
                index[tag_key] = []
            index[tag_key].append(tag)

        return index

    def validate_tag_format(self, tag: TagReference) -> tuple[bool, str | None]:
        """태그 형식 검증"""

        # 유효한 태그 타입 확인
        if tag.tag_type not in self.valid_tag_types:
            return (
                False,
                f"Invalid tag type '{tag.tag_type}' (valid: {', '.join(self.valid_tag_types)})",
            )

        # 태그 ID 형식 확인
        if not re.match(r"^[A-Z0-9-_]+$", tag.tag_id):
            return (
                False,
                f"Invalid tag ID format '{tag.tag_id}' (use uppercase, numbers, hyphens, underscores only)",
            )

        # 길이 제한
        if len(tag.tag_id) < 2:
            return False, f"Tag ID '{tag.tag_id}' too short (minimum 2 characters)"

        if len(tag.tag_id) > 50:
            return False, f"Tag ID '{tag.tag_id}' too long (maximum 50 characters)"

        return True, None

    def find_orphan_tags(self, tag_index: dict[str, list[TagReference]]) -> list[str]:
        """고아 태그 (참조되지 않는 태그) 찾기"""

        orphan_tags = []

        # 모든 태그에서 다른 태그로의 참조 추출
        referenced_tags = set()

        for tag_key, tag_refs in tag_index.items():
            for tag_ref in tag_refs:
                # 컨텍스트에서 다른 태그 참조 찾기
                context_tags = re.findall(
                    r"@([A-Z]+)[-:]([A-Z0-9-_]+)", tag_ref.context
                )
                for ref_type, ref_id in context_tags:
                    if f"{ref_type}:{ref_id}" != tag_key:  # 자기 자신 제외
                        referenced_tags.add(f"{ref_type}:{ref_id}")

        # 참조되지 않는 태그 찾기 (단, 루트 태그는 제외)
        root_tag_types = ["REQ", "SPEC", "VISION"]  # 루트가 될 수 있는 태그

        for tag_key in tag_index:
            tag_type = tag_key.split(":")[0]

            if tag_key not in referenced_tags and tag_type not in root_tag_types:
                orphan_tags.append(tag_key)

        return orphan_tags

    def find_broken_links(
        self, tag_index: dict[str, list[TagReference]]
    ) -> list[tuple[str, str]]:
        """깨진 링크 (존재하지 않는 태그 참조) 찾기"""

        broken_links = []

        for tag_key, tag_refs in tag_index.items():
            for tag_ref in tag_refs:
                # 컨텍스트에서 다른 태그 참조 찾기
                context_tags = re.findall(
                    r"@([A-Z]+)[-:]([A-Z0-9-_]+)", tag_ref.context
                )

                for ref_type, ref_id in context_tags:
                    referenced_tag = f"{ref_type}:{ref_id}"

                    # 자기 자신이 아니고 인덱스에 없는 경우
                    if referenced_tag != tag_key and referenced_tag not in tag_index:
                        broken_links.append((tag_key, referenced_tag))

        return broken_links

    def validate_traceability_chains(
        self, tag_index: dict[str, list[TagReference]]
    ) -> list[str]:
        """추적성 체인 검증"""

        chain_violations = []

        for chain_name, chain_types in self.traceability_chains.items():
            # 체인의 각 단계별 태그 수집
            chain_tags = {}
            for tag_type in chain_types:
                chain_tags[tag_type] = [
                    tag_key
                    for tag_key in tag_index
                    if tag_key.startswith(f"{tag_type}:")
                ]

            # 체인 연결성 검사
            for i in range(len(chain_types) - 1):
                current_type = chain_types[i]
                next_type = chain_types[i + 1]

                current_tags = chain_tags[current_type]
                next_tags = chain_tags[next_type]

                # 현재 단계 태그가 있는데 다음 단계 태그가 없는 경우
                if current_tags and not next_tags:
                    chain_violations.append(
                        f"{chain_name} chain broken: {current_type} tags exist but no {next_type} tags found"
                    )

        return chain_violations

    def calculate_tag_quality_score(
        self, total_tags: int, valid_tags: int, orphan_tags: int, broken_links: int
    ) -> float:
        """태그 품질 점수 계산"""

        if total_tags == 0:
            return 1.0

        # 기본 점수 (유효성)
        validity_score = valid_tags / total_tags if total_tags > 0 else 0

        # 연결성 점수 (고아 태그 감점)
        orphan_penalty = (orphan_tags / total_tags) * 0.3 if total_tags > 0 else 0
        connectivity_score = max(0, 1.0 - orphan_penalty)

        # 무결성 점수 (깨진 링크 감점)
        broken_penalty = (broken_links / total_tags) * 0.4 if total_tags > 0 else 0
        integrity_score = max(0, 1.0 - broken_penalty)

        # 가중 평균
        quality_score = (
            validity_score * 0.4 + connectivity_score * 0.3 + integrity_score * 0.3
        )

        return round(quality_score, 3)

    def update_tag_indexes(self, tag_index: dict[str, list[TagReference]]) -> None:
        """태그 인덱스 파일 업데이트"""

        # tags.db (SQLite) 업데이트 - TODO: 실제 SQLite 로직으로 전환 필요
        tags_file = self.indexes_dir / "tags.json"  # 임시: JSON 호환성 유지

        tags_data = {
            "version": "0.1.9",
            "updated": datetime.now().isoformat(),
            "statistics": {"total_tags": len(tag_index), "categories": {}},
            "index": {},
        }

        # 카테고리별 통계
        for category, tag_types in self.tag_categories.items():
            count = sum(
                1
                for tag_key in tag_index
                if any(tag_key.startswith(f"{t}:") for t in tag_types)
            )
            tags_data["statistics"]["categories"][category] = count

        # 인덱스 데이터
        for tag_key, tag_refs in tag_index.items():
            tags_data["index"][tag_key] = [
                {
                    "file": ref.file_path,
                    "line": ref.line_number,
                    "context": ref.context[:100],  # 처음 100자만
                }
                for ref in tag_refs
            ]

        # 파일 저장
        self.indexes_dir.mkdir(parents=True, exist_ok=True)
        tags_file.write_text(json.dumps(tags_data, indent=2, ensure_ascii=False))

        print(f"📄 Updated tag index: {tags_file}")

    def generate_health_report(
        self,
        total_tags: int,
        valid_tags: int,
        invalid_tags: int,
        orphan_tags: list[str],
        broken_links: list[tuple[str, str]],
        chain_violations: list[str],
    ) -> TagHealthReport:
        """태그 시스템 건강도 리포트 생성"""

        quality_score = self.calculate_tag_quality_score(
            total_tags, valid_tags, len(orphan_tags), len(broken_links)
        )

        issues = []
        recommendations = []

        # 이슈 수집
        if invalid_tags > 0:
            issues.append(f"{invalid_tags} invalid tag format(s)")
            recommendations.append(
                "Fix invalid tag formats using 16-Core naming conventions"
            )

        if orphan_tags:
            issues.append(f"{len(orphan_tags)} orphan tag(s) found")
            recommendations.append(
                "Link orphan tags to parent tags or remove if unused"
            )

        if broken_links:
            issues.append(f"{len(broken_links)} broken link(s)")
            recommendations.append("Fix broken tag references or create missing tags")

        if chain_violations:
            issues.extend(chain_violations)
            recommendations.append("Complete traceability chains for all requirements")

        if quality_score < 0.7:
            recommendations.append(
                "Improve overall tag quality by addressing issues above"
            )

        return TagHealthReport(
            total_tags=total_tags,
            valid_tags=valid_tags,
            invalid_tags=invalid_tags,
            orphan_tags=len(orphan_tags),
            broken_links=len(broken_links),
            quality_score=quality_score,
            issues=issues,
            recommendations=recommendations,
        )

    def run_validation(self) -> TagHealthReport:
        """전체 태그 검증 실행"""

        print("🏷️  Starting 16-Core TAG system validation...")

        # 1. 프로젝트 스캔
        print("  Scanning project files for tags...")
        self.all_tags = self.scan_project_files()
        print(f"  Found {len(self.all_tags)} tag references")

        # 2. 태그 인덱스 구축
        self.tag_index = self.build_tag_index(self.all_tags)
        print(f"  Unique tags: {len(self.tag_index)}")

        # 3. 형식 검증
        print("  Validating tag formats...")
        valid_tags = 0
        invalid_tags = 0

        for tag in self.all_tags:
            is_valid, error_msg = self.validate_tag_format(tag)
            if is_valid:
                valid_tags += 1
            else:
                invalid_tags += 1
                self.violations.append(
                    f"{tag.file_path}:{tag.line_number} - {error_msg}"
                )

        # 4. 고아 태그 찾기
        print("  Finding orphan tags...")
        orphan_tags = self.find_orphan_tags(self.tag_index)

        # 5. 깨진 링크 찾기
        print("  Finding broken links...")
        broken_links = self.find_broken_links(self.tag_index)

        # 6. 추적성 체인 검증
        print("  Validating traceability chains...")
        chain_violations = self.validate_traceability_chains(self.tag_index)

        # 7. 인덱스 업데이트
        print("  Updating tag indexes...")
        self.update_tag_indexes(self.tag_index)

        # 8. 건강도 리포트 생성
        report = self.generate_health_report(
            len(self.all_tags),
            valid_tags,
            invalid_tags,
            orphan_tags,
            broken_links,
            chain_violations,
        )

        print(f"  Tag quality score: {report.quality_score:.1%}")

        return report


def main():
    """메인 실행 함수"""

    project_root = Path.cwd()

    # MoAI 프로젝트 확인
    if not (project_root / ".moai").exists():
        print("❌ This is not a MoAI-ADK project")
        sys.exit(1)

    try:
        # 태그 검증 실행
        validator = TagValidator(project_root)
        report = validator.run_validation()

        # 결과 출력
        print("\n" + "=" * 60)
        print("🏷️  16-CORE TAG VALIDATION REPORT")
        print("=" * 60)

        print(f"Total Tags: {report.total_tags}")
        print(f"Valid Tags: {report.valid_tags}")
        print(f"Invalid Tags: {report.invalid_tags}")
        print(f"Orphan Tags: {report.orphan_tags}")
        print(f"Broken Links: {report.broken_links}")
        print(f"Quality Score: {report.quality_score:.1%}")

        # 상태 판정
        if report.quality_score >= 0.8:
            status = "✅ EXCELLENT"
        elif report.quality_score >= 0.6:
            status = "🟡 GOOD"
        elif report.quality_score >= 0.4:
            status = "🟠 NEEDS IMPROVEMENT"
        else:
            status = "❌ POOR"

        print(f"Overall Status: {status}")

        # 이슈 출력
        if report.issues:
            print(f"\n⚠️  Issues Found ({len(report.issues)}):")
            for issue in report.issues[:10]:  # 상위 10개만
                print(f"  • {issue}")
            if len(report.issues) > 10:
                print(f"  ... and {len(report.issues) - 10} more")

        # 권장사항
        if report.recommendations:
            print("\n💡 Recommendations:")
            for rec in report.recommendations:
                print(f"  • {rec}")

        # 위반사항 상세 (처음 5개만)
        if validator.violations:
            print("\n🔍 Validation Violations (first 5):")
            for violation in validator.violations[:5]:
                print(f"  • {violation}")
            if len(validator.violations) > 5:
                print(f"  ... and {len(validator.violations) - 5} more violations")

        # 리포트 저장
        report_data = {
            "tag_health": report.__dict__,
            "violations": validator.violations,
            "scan_info": {
                "timestamp": datetime.now().isoformat(),
                "total_files_scanned": len(
                    set(tag.file_path for tag in validator.all_tags)
                ),
                "version": "0.1.9",
            },
        }

        # SQLite 보고서 저장
        report_file = project_root / ".moai" / "reports" / "tag_validation.db"
        report_file.parent.mkdir(parents=True, exist_ok=True)

        try:
            conn = sqlite3.connect(report_file)
            cursor = conn.cursor()

            # 보고서 테이블 생성
            cursor.execute("DROP TABLE IF EXISTS validation_report")
            cursor.execute("""
                CREATE TABLE validation_report (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    metric TEXT NOT NULL,
                    value TEXT NOT NULL,
                    details TEXT,
                    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            # 보고서 데이터 삽입
            cursor.execute(
                "INSERT INTO validation_report (metric, value, details) VALUES ('timestamp', ?, '')",
                (report_data["timestamp"],),
            )
            cursor.execute(
                "INSERT INTO validation_report (metric, value, details) VALUES ('total_tags', ?, '')",
                (str(report_data["summary"]["total_tags"]),),
            )
            cursor.execute(
                "INSERT INTO validation_report (metric, value, details) VALUES ('valid_tags', ?, '')",
                (str(report_data["summary"]["valid_tags"]),),
            )
            cursor.execute(
                "INSERT INTO validation_report (metric, value, details) VALUES ('quality_score', ?, '')",
                (str(report_data["summary"]["quality_score"]),),
            )

            # 단순화된 보고서로 대체
            for issue_type, issues in report_data["issues"].items():
                cursor.execute(
                    "INSERT INTO validation_report (metric, value, details) VALUES (?, ?, ?)",
                    (issue_type, str(len(issues)), str(issues)[:500]),
                )

            conn.commit()
            conn.close()

            print(f"\n📄 Detailed report saved to SQLite: {report_file}")
        except Exception as e:
            print(f"\n⚠️  Failed to save report: {e}")

        # Exit code (PASS if quality >= 60%)
        sys.exit(0 if report.quality_score >= 0.6 else 1)

    except Exception as error:
        print(f"❌ Tag validation failed: {error}")
        sys.exit(1)


if __name__ == "__main__":
    main()
