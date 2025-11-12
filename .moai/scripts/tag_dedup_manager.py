#!/usr/bin/env python3
"""
MoAI-ADK TAG De-duplication Manager (Unified)

Merged from:
  - tag_dedup_detector.py: TAG 중복 탐지
  - tag_auto_corrector.py: TAG 자동 수정

통합된 단일 명령어로 TAG 시스템의 중복을 감지하고 수정합니다.
"""

import json
import re
import os
import sys
import subprocess
from pathlib import Path
from typing import Dict, List, Tuple, Set, Optional
from collections import defaultdict, Counter
from dataclasses import dataclass, field
from datetime import datetime


# ============================================================================
# Data Structures (공유 데이터 구조)
# ============================================================================

@dataclass
class TagInfo:
    """TAG 정보 구조체"""
    tag: str
    file_path: str
    line_number: int
    line_content: str
    is_topline: bool
    tag_type: str  # SPEC, TEST, CODE, DOC
    domain: str
    id_number: str
    authority_score: int


@dataclass
class DuplicateGroup:
    """중복 그룹 정보"""
    tag_pattern: str
    occurrences: List[TagInfo] = field(default_factory=list)
    primary_candidate: Optional[TagInfo] = None
    duplicates: List[TagInfo] = field(default_factory=list)


@dataclass
class CorrectionAction:
    """수정 액션 정보"""
    action_type: str  # renumber, remove, update_reference
    file_path: str
    line_number: int
    old_tag: str
    new_tag: Optional[str]
    confidence: float
    impact: str  # low, medium, high, critical
    description: str


@dataclass
class CorrectionResult:
    """수정 결과 정보"""
    action: CorrectionAction
    success: bool
    error_message: Optional[str] = None
    backup_path: Optional[str] = None


# ============================================================================
# TAG 탐지기 (Detection Engine)
# ============================================================================

class TagDetectionEngine:
    """TAG 중복 탐지 엔진"""

    def __init__(self, config_path: str = None):
        self.project_root = Path.cwd()
        self.config = self._load_config(config_path)
        self.tag_pattern = re.compile(r'@([A-Z]+):([A-Z_]+)-([0-9]{3,})')
        self.all_tags: List[TagInfo] = []
        self.duplicate_groups: List[DuplicateGroup] = []

    def _load_config(self, config_path: str = None) -> Dict:
        """설정 파일 로드"""
        if config_path is None:
            config_path = self.project_root / ".moai" / "tag-policy.json"

        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except FileNotFoundError:
            return self._default_config()

    def _default_config(self) -> Dict:
        """기본 설정"""
        return {
            "scan_paths": ["src/", ".moai/", ".claude/"],
            "exclude_paths": ["node_modules/", ".git/", "__pycache__/"],
            "tag_types": ["SPEC", "TEST", "CODE", "DOC"],
            "min_duplicates": 2,
        }

    def _calculate_authority_score(self, file_path: str, line_number: int) -> int:
        """권한 점수 계산"""
        score = 50

        if ".moai/specs/" in str(file_path):
            score += 30
        elif ".claude/" in str(file_path):
            score += 20
        elif "src/" in str(file_path):
            score += 10

        if line_number < 10:
            score += 10

        return min(score, 100)

    def _is_topline_tag(self, line_content: str, line_number: int) -> bool:
        """헤더라인 TAG 여부"""
        return bool(re.match(r'^#+ ', line_content))

    def _extract_tags_from_file(self, file_path: str) -> List[TagInfo]:
        """파일에서 TAG 추출"""
        tags = []
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                for line_num, line_content in enumerate(f, 1):
                    for match in self.tag_pattern.finditer(line_content):
                        tag_type = match.group(1)
                        domain = match.group(2)
                        id_number = match.group(3)
                        full_tag = match.group(0)

                        tag_info = TagInfo(
                            tag=full_tag,
                            file_path=str(file_path),
                            line_number=line_num,
                            line_content=line_content.strip(),
                            is_topline=self._is_topline_tag(line_content, line_num),
                            tag_type=tag_type,
                            domain=domain,
                            id_number=id_number,
                            authority_score=self._calculate_authority_score(
                                str(file_path), line_num
                            ),
                        )
                        tags.append(tag_info)
        except (UnicodeDecodeError, IOError):
            pass

        return tags

    def scan_all_files(self) -> None:
        """모든 파일 스캔"""
        exclude_dirs = {
            d.lower() for d in self.config.get("exclude_paths", [])
        }

        for root, dirs, files in os.walk(self.project_root):
            dirs[:] = [
                d for d in dirs if d.lower() not in exclude_dirs
            ]

            for file in files:
                if file.endswith((".py", ".md", ".sh", ".js", ".ts")):
                    file_path = Path(root) / file
                    tags = self._extract_tags_from_file(file_path)
                    self.all_tags.extend(tags)

    def find_duplicates(self) -> None:
        """중복 탐지"""
        tag_occurrences: Dict[str, List[TagInfo]] = defaultdict(list)

        for tag in self.all_tags:
            pattern = f"{tag.tag_type}:{tag.domain}-*"
            tag_occurrences[pattern].append(tag)

        for pattern, occurrences in tag_occurrences.items():
            if len(occurrences) >= self.config.get("min_duplicates", 2):
                dup_group = DuplicateGroup(
                    tag_pattern=pattern,
                    occurrences=occurrences,
                )

                sorted_tags = sorted(
                    occurrences,
                    key=lambda t: (-t.authority_score, t.line_number),
                )
                dup_group.primary_candidate = sorted_tags[0]
                dup_group.duplicates = sorted_tags[1:]

                self.duplicate_groups.append(dup_group)

    def analyze_duplicates(self) -> Dict:
        """중복 분석"""
        analysis = {
            "total_tags": len(self.all_tags),
            "duplicate_groups": len(self.duplicate_groups),
            "total_duplicates": sum(
                len(group.duplicates) for group in self.duplicate_groups
            ),
            "details": [],
        }

        for group in self.duplicate_groups:
            group_detail = {
                "pattern": group.tag_pattern,
                "count": len(group.occurrences),
                "primary": {
                    "tag": group.primary_candidate.tag,
                    "file": group.primary_candidate.file_path,
                    "line": group.primary_candidate.line_number,
                },
                "duplicates": [
                    {
                        "tag": dup.tag,
                        "file": dup.file_path,
                        "line": dup.line_number,
                    }
                    for dup in group.duplicates
                ],
            }
            analysis["details"].append(group_detail)

        return analysis


# ============================================================================
# TAG 수정기 (Correction Engine)
# ============================================================================

class TagCorrectionEngine:
    """TAG 자동 수정 엔진"""

    def __init__(self, config_path: str = None, dry_run: bool = True):
        self.project_root = Path.cwd()
        self.config = self._load_config(config_path)
        self.dry_run = dry_run
        self.detector = TagDetectionEngine(config_path)
        self.corrections: List[CorrectionAction] = []
        self.results: List[CorrectionResult] = []

    def _load_config(self, config_path: str = None) -> Dict:
        """설정 파일 로드"""
        if config_path is None:
            config_path = self.project_root / ".moai" / "tag-policy.json"

        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except FileNotFoundError:
            return {}

    def _create_backup(self) -> str:
        """백업 생성"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = self.project_root / ".moai" / "backups" / "tag_corrections"
        backup_dir.mkdir(parents=True, exist_ok=True)

        backup_path = backup_dir / f"backup_{timestamp}.tar.gz"
        os.system(f"tar -czf {backup_path} .moai/specs/ src/ .claude/")

        return str(backup_path)

    def _validate_correction(
        self, action: CorrectionAction
    ) -> Tuple[bool, str]:
        """수정 검증"""
        if not action.new_tag:
            return True, "제거 작업"

        tag_pattern = re.compile(r'^@[A-Z]+:[A-Z_]+-\d{3,}$')
        if not tag_pattern.match(action.new_tag):
            return False, f"유효하지 않은 TAG 형식: {action.new_tag}"

        return True, "검증 완료"

    def _apply_line_correction(self, action: CorrectionAction) -> bool:
        """라인 수정 적용"""
        try:
            file_path = Path(action.file_path)

            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            if action.line_number > len(lines):
                return False

            target_line = lines[action.line_number - 1]

            if action.new_tag:
                new_line = target_line.replace(
                    action.old_tag, action.new_tag
                )
            else:
                new_line = target_line.replace(action.old_tag, "")

            if not self.dry_run:
                lines[action.line_number - 1] = new_line
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.writelines(lines)

            return True
        except Exception as e:
            return False

    def generate_corrections(
        self, duplicate_groups: List[DuplicateGroup]
    ) -> None:
        """수정 계획 생성"""
        for group in duplicate_groups:
            for dup in group.duplicates:
                new_tag = self._generate_new_tag(
                    dup.tag, dup.domain, dup.tag_type
                )

                action = CorrectionAction(
                    action_type="renumber",
                    file_path=dup.file_path,
                    line_number=dup.line_number,
                    old_tag=dup.tag,
                    new_tag=new_tag,
                    confidence=0.95,
                    impact="high",
                    description=f"중복 TAG 제거: {dup.tag} → {new_tag}",
                )

                self.corrections.append(action)

    def _generate_new_tag(
        self, old_tag: str, domain: str, tag_type: str
    ) -> str:
        """새 TAG 생성"""
        next_num = 1

        for tag in self.detector.all_tags:
            if tag.domain == domain and tag.tag_type == tag_type:
                try:
                    num = int(tag.id_number)
                    next_num = max(next_num, num + 1)
                except ValueError:
                    pass

        return f"@{tag_type}:{domain}-{next_num:03d}"

    def apply_corrections(self) -> List[CorrectionResult]:
        """수정 적용"""
        backup_path = None

        if not self.dry_run and self.corrections:
            backup_path = self._create_backup()

        for action in self.corrections:
            is_valid, validation_msg = self._validate_correction(action)

            if not is_valid:
                result = CorrectionResult(
                    action=action,
                    success=False,
                    error_message=validation_msg,
                    backup_path=backup_path,
                )
                self.results.append(result)
                continue

            success = self._apply_line_correction(action)

            result = CorrectionResult(
                action=action,
                success=success,
                error_message=None if success else "적용 실패",
                backup_path=backup_path,
            )
            self.results.append(result)

        return self.results

    def save_correction_report(
        self, output_path: str = None
    ) -> None:
        """수정 리포트 저장"""
        if output_path is None:
            output_path = (
                self.project_root
                / ".moai"
                / "reports"
                / f"tag_corrections_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
            )

        Path(output_path).parent.mkdir(parents=True, exist_ok=True)

        report = {
            "timestamp": datetime.now().isoformat(),
            "total_corrections": len(self.corrections),
            "successful": sum(1 for r in self.results if r.success),
            "failed": sum(1 for r in self.results if not r.success),
            "dry_run": self.dry_run,
            "results": [
                {
                    "action": action.action_type,
                    "file": action.file_path,
                    "line": action.line_number,
                    "old_tag": action.old_tag,
                    "new_tag": action.new_tag,
                    "success": result.success,
                    "error": result.error_message,
                }
                for action, result in zip(self.corrections, self.results)
            ],
        }

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        print(f"📄 리포트 저장: {output_path}")


# ============================================================================
# 통합 관리자 (Unified Manager)
# ============================================================================

class TagDedupManager:
    """TAG 중복 관리자 (통합)"""

    def __init__(self, config_path: str = None):
        self.config_path = config_path
        self.detector = TagDetectionEngine(config_path)
        self.corrector = None

    def scan_only(self) -> Dict:
        """스캔 전용 모드"""
        print("🔍 TAG 중복 스캔 중...")
        self.detector.scan_all_files()
        self.detector.find_duplicates()

        analysis = self.detector.analyze_duplicates()

        if analysis["duplicate_groups"] > 0:
            print(f"\n⚠️  중복 그룹 발견: {analysis['duplicate_groups']}")
            print(f"   총 중복 TAG 수: {analysis['total_duplicates']}")
        else:
            print("\n✅ 중복 TAG가 없습니다!")

        return analysis

    def dry_run(self) -> Dict:
        """검토 모드"""
        print("📋 TAG 중복 검토 중 (변경 없음)...")
        self.detector.scan_all_files()
        self.detector.find_duplicates()

        self.corrector = TagCorrectionEngine(self.config_path, dry_run=True)
        self.corrector.detector = self.detector
        self.corrector.generate_corrections(self.detector.duplicate_groups)

        analysis = self.detector.analyze_duplicates()

        print(f"\n📊 수정 계획:")
        print(f"   적용될 수정: {len(self.corrector.corrections)}")

        for action in self.corrector.corrections[:5]:
            print(
                f"   - {action.old_tag} → {action.new_tag} ({action.file_path}:{action.line_number})"
            )

        if len(self.corrector.corrections) > 5:
            print(f"   ... 외 {len(self.corrector.corrections) - 5}개")

        return analysis

    def apply(self) -> List[CorrectionResult]:
        """적용 모드"""
        print("✏️  TAG 중복 수정 중...")
        self.detector.scan_all_files()
        self.detector.find_duplicates()

        self.corrector = TagCorrectionEngine(self.config_path, dry_run=False)
        self.corrector.detector = self.detector
        self.corrector.generate_corrections(self.detector.duplicate_groups)

        results = self.corrector.apply_corrections()

        successful = sum(1 for r in results if r.success)
        failed = len(results) - successful

        print(f"\n✅ 수정 완료:")
        print(f"   성공: {successful}")
        print(f"   실패: {failed}")

        if results and results[0].backup_path:
            print(f"   백업: {results[0].backup_path}")

        self.corrector.save_correction_report()

        return results

    def full_workflow(self) -> Dict:
        """전체 워크플로우"""
        print("🚀 전체 TAG 중복 정리 시작...\n")

        analysis = self.scan_only()
        print()

        if analysis["duplicate_groups"] > 0:
            self.dry_run()
            print()
            self.apply()
        else:
            print("\n✅ 정리할 중복이 없습니다!")

        return analysis


# ============================================================================
# CLI
# ============================================================================

def main():
    """메인 함수"""
    import argparse

    parser = argparse.ArgumentParser(
        description="MoAI-ADK TAG 중복 관리 도구 (통합)",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
사용 예시:
  # 스캔 전용 (변경 없음)
  python3 tag_dedup_manager.py --scan-only

  # 검토 모드 (수정 계획만 표시)
  python3 tag_dedup_manager.py --dry-run

  # 수정 적용
  python3 tag_dedup_manager.py --apply

  # 전체 워크플로우 (스캔 → 검토 → 적용)
  python3 tag_dedup_manager.py --full

  # 설정 파일 지정
  python3 tag_dedup_manager.py --config .moai/my-config.json --apply
        """,
    )

    parser.add_argument(
        "--config",
        type=str,
        help="설정 파일 경로",
        default=None,
    )

    # 모드 선택
    mode_group = parser.add_mutually_exclusive_group(required=True)
    mode_group.add_argument(
        "--scan-only",
        action="store_true",
        help="중복 스캔만 수행 (변경 없음)",
    )
    mode_group.add_argument(
        "--dry-run",
        action="store_true",
        help="수정 계획을 보고 변경하지 않음",
    )
    mode_group.add_argument(
        "--apply",
        action="store_true",
        help="중복 TAG 수정 적용",
    )
    mode_group.add_argument(
        "--full",
        action="store_true",
        help="전체 워크플로우 실행 (스캔 → 검토 → 적용)",
    )

    args = parser.parse_args()

    manager = TagDedupManager(config_path=args.config)

    if args.scan_only:
        manager.scan_only()
    elif args.dry_run:
        manager.dry_run()
    elif args.apply:
        manager.apply()
    elif args.full:
        manager.full_workflow()


if __name__ == "__main__":
    main()
