# @TASK:TAG-SYSTEM-VALIDATOR-011
"""
TAG 시스템 종합 검증기 - SPEC-011 Refactor Phase

TRUST 5원칙을 모두 적용한 고급 TAG 시스템 검증 도구입니다.
"""

import os
import re
import json
import time
import logging
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Set
from dataclasses import dataclass, asdict
from collections import Counter, defaultdict
import subprocess
from datetime import datetime


@dataclass
class TagValidationResult:
    """TAG 검증 결과"""
    file_path: str
    has_tag: bool
    tag_count: int
    tags: List[str]
    valid_format: bool
    issues: List[str]


@dataclass
class SystemHealthReport:
    """시스템 건강도 리포트"""
    timestamp: str
    total_files: int
    tagged_files: int
    coverage_percent: float
    primary_chain_completion: float
    tag_distribution: Dict[str, int]
    duplicate_tags: List[str]
    inconsistent_tags: List[str]
    validation_time: float
    issues_summary: Dict[str, int]


class TagSystemValidator:
    """TAG 시스템 종합 검증기 - TRUST 5원칙 적용"""

    def __init__(self, project_root: str = None):
        self.project_root = project_root or "/Users/goos/MoAI/MoAI-ADK"
        self.src_dir = os.path.join(self.project_root, "src")

        # 패턴 정의
        self.tag_pattern = re.compile(r'@[A-Z]+:[A-Z-]+-\d+')
        self.standard_pattern = re.compile(r'@[A-Z]+:[A-Z-]+-\d{3}')

        # 16-Core TAG 시스템 정의
        self.core_categories = {
            'SPEC': ['REQ', 'DESIGN', 'TASK'],
            'PROJECT': ['VISION', 'STRUCT', 'TECH', 'ADR'],
            'IMPLEMENTATION': ['FEATURE', 'API', 'TEST', 'DATA'],
            'QUALITY': ['PERF', 'SEC', 'DEBT', 'TODO']
        }
        self.all_categories = set()
        for categories in self.core_categories.values():
            self.all_categories.update(categories)

        # Primary Chain 정의
        self.primary_chain = ['REQ', 'DESIGN', 'TASK', 'TEST']

        # 로거 초기화
        self.logger = self._setup_logger()

    def _setup_logger(self) -> logging.Logger:
        """구조화된 로깅 설정"""
        logger = logging.getLogger('tag_validator')
        logger.setLevel(logging.INFO)

        # 기존 핸들러 제거
        for handler in logger.handlers[:]:
            logger.removeHandler(handler)

        # 새 핸들러 추가
        handler = logging.StreamHandler()
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        handler.setFormatter(formatter)
        logger.addHandler(handler)

        return logger

    def find_all_python_files(self) -> List[str]:
        """모든 Python 파일 검색"""
        python_files = []

        for root, dirs, files in os.walk(self.src_dir):
            # 숨김 디렉토리나 __pycache__ 제외
            dirs[:] = [d for d in dirs if not d.startswith('.') and d != '__pycache__']

            for file in files:
                if file.endswith('.py') and not file.startswith('.'):
                    file_path = os.path.join(root, file)
                    python_files.append(file_path)

        return sorted(python_files)

    def validate_file(self, file_path: str) -> TagValidationResult:
        """개별 파일 TAG 검증"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except Exception as e:
            return TagValidationResult(
                file_path=file_path,
                has_tag=False,
                tag_count=0,
                tags=[],
                valid_format=False,
                issues=[f"File read error: {str(e)}"]
            )

        # TAG 추출
        tags = self.tag_pattern.findall(content)
        has_tag = len(tags) > 0
        issues = []

        # 형식 검증
        valid_format = True
        for tag in tags:
            if not self._validate_tag_format(tag):
                valid_format = False
                issues.append(f"Invalid format: {tag}")

        # 중복 TAG 검사
        tag_counts = Counter(tags)
        for tag, count in tag_counts.items():
            if count > 1:
                issues.append(f"Duplicate tag: {tag} ({count} times)")

        return TagValidationResult(
            file_path=file_path,
            has_tag=has_tag,
            tag_count=len(tags),
            tags=tags,
            valid_format=valid_format,
            issues=issues
        )

    def _validate_tag_format(self, tag: str) -> bool:
        """TAG 형식 검증"""
        # 기본 패턴 매치
        if not self.tag_pattern.match(tag):
            return False

        # 카테고리 유효성 검사
        try:
            category = tag.split(':')[0].replace('@', '')
            return category in self.all_categories
        except IndexError:
            return False

    def analyze_tag_distribution(self, all_results: List[TagValidationResult]) -> Dict[str, int]:
        """TAG 카테고리별 분포 분석"""
        distribution = defaultdict(int)

        for result in all_results:
            for tag in result.tags:
                try:
                    category = tag.split(':')[0].replace('@', '')
                    distribution[category] += 1
                except IndexError:
                    continue

        return dict(distribution)

    def calculate_primary_chain_completion(self, distribution: Dict[str, int]) -> float:
        """Primary Chain 완성도 계산"""
        present_categories = sum(1 for cat in self.primary_chain if distribution.get(cat, 0) > 0)
        return present_categories / len(self.primary_chain)

    def find_duplicate_tags(self, all_results: List[TagValidationResult]) -> List[str]:
        """전체 시스템에서 중복 TAG 찾기"""
        tag_files = defaultdict(list)
        duplicates = []

        for result in all_results:
            for tag in result.tags:
                tag_files[tag].append(result.file_path)

        for tag, files in tag_files.items():
            if len(files) > 1:
                duplicates.append(tag)

        return duplicates

    def find_inconsistent_tags(self, all_results: List[TagValidationResult]) -> List[str]:
        """일관성 없는 TAG 찾기"""
        inconsistent = set()

        for result in all_results:
            for tag in result.tags:
                if not self.standard_pattern.match(tag):
                    inconsistent.add(tag)

        return list(inconsistent)

    def generate_comprehensive_report(self) -> SystemHealthReport:
        """종합 시스템 건강도 리포트 생성"""
        start_time = time.time()
        self.logger.info("Starting comprehensive TAG system validation")

        # 모든 파일 검증
        python_files = self.find_all_python_files()
        all_results = []

        for i, file_path in enumerate(python_files, 1):
            if i % 20 == 0:  # 진행률 로깅
                self.logger.info(f"Validated {i}/{len(python_files)} files")

            result = self.validate_file(file_path)
            all_results.append(result)

        # 통계 계산
        tagged_files = sum(1 for r in all_results if r.has_tag)
        coverage_percent = (tagged_files / len(python_files)) * 100 if python_files else 0

        # 분석 수행
        distribution = self.analyze_tag_distribution(all_results)
        primary_chain_completion = self.calculate_primary_chain_completion(distribution)
        duplicate_tags = self.find_duplicate_tags(all_results)
        inconsistent_tags = self.find_inconsistent_tags(all_results)

        # 이슈 요약
        issues_summary = {
            'files_without_tags': len(python_files) - tagged_files,
            'duplicate_tags': len(duplicate_tags),
            'inconsistent_format': len(inconsistent_tags),
            'files_with_issues': sum(1 for r in all_results if r.issues)
        }

        validation_time = time.time() - start_time

        report = SystemHealthReport(
            timestamp=datetime.now().isoformat(),
            total_files=len(python_files),
            tagged_files=tagged_files,
            coverage_percent=coverage_percent,
            primary_chain_completion=primary_chain_completion * 100,
            tag_distribution=distribution,
            duplicate_tags=duplicate_tags,
            inconsistent_tags=inconsistent_tags,
            validation_time=validation_time,
            issues_summary=issues_summary
        )

        self.logger.info(f"Validation completed in {validation_time:.2f}s")
        return report

    def print_report(self, report: SystemHealthReport):
        """리포트를 읽기 쉽게 출력"""
        print("\n" + "="*60)
        print("🗿 MoAI-ADK TAG System Health Report")
        print("="*60)
        print(f"📅 Generated: {report.timestamp}")
        print(f"⏱️  Validation time: {report.validation_time:.2f}s")
        print()

        # 전체 상태
        print("📊 OVERALL STATUS")
        print(f"   Total Python files: {report.total_files}")
        print(f"   Files with @TAG: {report.tagged_files}")
        print(f"   Coverage: {report.coverage_percent:.1f}%")
        print(f"   Primary Chain completion: {report.primary_chain_completion:.1f}%")
        print()

        # TAG 분포
        print("🏷️  TAG DISTRIBUTION")
        if report.tag_distribution:
            for category, count in sorted(report.tag_distribution.items()):
                print(f"   {category}: {count}")
        else:
            print("   No tags found")
        print()

        # 이슈 요약
        print("⚠️  ISSUES SUMMARY")
        for issue_type, count in report.issues_summary.items():
            if count > 0:
                print(f"   {issue_type.replace('_', ' ').title()}: {count}")

        if sum(report.issues_summary.values()) == 0:
            print("   ✅ No issues found - system is healthy!")
        print()

        # 품질 점수 계산
        quality_score = self._calculate_quality_score(report)
        print(f"🎯 QUALITY SCORE: {quality_score:.1f}/100")

        if quality_score >= 90:
            print("   🟢 Excellent - Production Ready")
        elif quality_score >= 75:
            print("   🟡 Good - Minor improvements needed")
        elif quality_score >= 50:
            print("   🟠 Fair - Several issues to address")
        else:
            print("   🔴 Poor - Significant improvements required")

    def _calculate_quality_score(self, report: SystemHealthReport) -> float:
        """품질 점수 계산 (0-100)"""
        # 기본 점수는 커버리지
        score = report.coverage_percent * 0.4

        # Primary Chain 완성도
        score += report.primary_chain_completion * 0.3

        # 이슈가 적을수록 높은 점수
        total_issues = sum(report.issues_summary.values())
        if report.total_files > 0:
            issue_penalty = (total_issues / report.total_files) * 30
            score += max(0, 30 - issue_penalty)

        return min(100, max(0, score))

    def save_report(self, report: SystemHealthReport, output_path: str = None):
        """리포트를 JSON으로 저장"""
        if output_path is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            output_path = f"/tmp/claude/tag_health_report_{timestamp}.json"

        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(asdict(report), f, indent=2, ensure_ascii=False)

        self.logger.info(f"Report saved to: {output_path}")
        print(f"📄 Report saved to: {output_path}")


def main():
    """CLI 실행 함수"""
    import argparse

    parser = argparse.ArgumentParser(description="TAG System Validator - SPEC-011")
    parser.add_argument("--output", "-o", type=str,
                       help="Output JSON report path")
    parser.add_argument("--strict", action="store_true",
                       help="Strict mode - exit with error if issues found")
    parser.add_argument("--quiet", "-q", action="store_true",
                       help="Quiet mode - minimal output")

    args = parser.parse_args()

    validator = TagSystemValidator()

    if not args.quiet:
        print("🔍 Starting TAG system validation...")

    # 검증 실행
    report = validator.generate_comprehensive_report()

    # 리포트 출력
    if not args.quiet:
        validator.print_report(report)

    # JSON 저장
    if args.output:
        validator.save_report(report, args.output)
    else:
        validator.save_report(report)

    # 종료 코드 결정
    if args.strict:
        total_issues = sum(report.issues_summary.values())
        if total_issues > 0:
            exit(1)  # 이슈가 있으면 실패
        elif report.coverage_percent < 100:
            exit(1)  # 커버리지가 100% 미만이면 실패

    exit(0)


if __name__ == "__main__":
    main()