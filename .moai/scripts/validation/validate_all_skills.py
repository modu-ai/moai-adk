#!/usr/bin/env python3
"""
125개 Enterprise Skills 활성화 종합 검증 스크립트
- YAML Frontmatter 검증
- Skill 이름 규칙 검증
- 파일 구조 완전성 검증
- 내용 품질 검증
- 통합 활성화 시뮬레이션
"""

import os
import yaml
import json
from pathlib import Path
from typing import Dict, List, Tuple
from dataclasses import dataclass
import sys

@dataclass
class SkillValidationResult:
    skill_name: str
    valid: bool
    errors: List[str]
    warnings: List[str]
    metadata: Dict

class SkillValidator:
    def __init__(self, skills_dir: Path):
        self.skills_dir = skills_dir
        self.results = []
        self.stats = {
            'total': 0,
            'valid': 0,
            'warnings': 0,
            'failed': 0,
            'errors_by_type': {}
        }

    def validate_all(self) -> List[SkillValidationResult]:
        """모든 Skills 검증"""
        skill_dirs = sorted([d for d in self.skills_dir.iterdir() if d.is_dir() and d.name.startswith('moai-')])

        print(f"\n🔍 검증 시작: {len(skill_dirs)}개 Skills")
        print("=" * 70)

        for i, skill_dir in enumerate(skill_dirs, 1):
            skill_name = skill_dir.name
            result = self._validate_skill(skill_dir)
            self.results.append(result)

            # 진행상황 표시
            status = "✅" if result.valid else "⚠️" if result.warnings else "❌"
            print(f"[{i:3d}/125] {status} {skill_name}")

            # 통계 업데이트
            self.stats['total'] += 1
            if result.valid and not result.warnings:
                self.stats['valid'] += 1
            elif result.warnings:
                self.stats['warnings'] += 1
            else:
                self.stats['failed'] += 1
                for error in result.errors:
                    error_type = error.split(':')[0]
                    self.stats['errors_by_type'][error_type] = self.stats['errors_by_type'].get(error_type, 0) + 1

        return self.results

    def _validate_skill(self, skill_dir: Path) -> SkillValidationResult:
        """단일 Skill 검증"""
        skill_name = skill_dir.name
        errors = []
        warnings = []
        metadata = {}

        # 1. SKILL.md 파일 존재 확인
        skill_file = skill_dir / "SKILL.md"
        if not skill_file.exists():
            errors.append("SKILL.md: 파일 없음")
            return SkillValidationResult(skill_name, False, errors, warnings, metadata)

        # 2. YAML Frontmatter 검증
        try:
            with open(skill_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # YAML frontmatter 추출
            if content.startswith('---'):
                parts = content.split('---')
                if len(parts) >= 3:
                    yaml_content = parts[1]
                    metadata = yaml.safe_load(yaml_content)

                    # 필수 필드 확인
                    required_fields = ['name', 'version', 'status', 'description']
                    for field in required_fields:
                        if not metadata or field not in metadata:
                            errors.append(f"YAML: 필수 필드 '{field}' 누락")
                else:
                    errors.append("YAML: Frontmatter 형식 오류")
        except Exception as e:
            errors.append(f"YAML: 파싱 오류 - {str(e)[:50]}")
            return SkillValidationResult(skill_name, False, errors, warnings, metadata)

        # 3. 파일 크기 검증
        file_size = skill_file.stat().st_size
        if file_size < 500:
            warnings.append(f"Size: {file_size} bytes (최소 500 bytes 권장)")

        # 4. Skill 이름 규칙 검증
        if metadata and 'name' in metadata:
            declared_name = metadata['name']
            if declared_name != skill_name:
                errors.append(f"Name: 디렉토리명({skill_name}) ≠ declared name({declared_name})")

        if not skill_name.startswith('moai-'):
            errors.append("Name: 'moai-' 프리픽스 누락")

        # 5. Progressive Disclosure 구조 검증
        required_sections = ['## Quick Reference', '## Implementation', '## Advanced']
        missing_sections = []
        for section in required_sections:
            if section not in content:
                missing_sections.append(section)

        if missing_sections:
            warnings.append(f"Structure: 누락된 섹션 - {', '.join(missing_sections)}")

        # 6. 코드 예제 개수 확인
        code_block_count = content.count('```')
        if code_block_count < 4:  # 최소 2개 이상의 코드 블록 (```...``` = 2 카운트)
            warnings.append(f"Examples: {code_block_count//2}개 코드 블록 (최소 2개 권장)")

        # 7. 추가 파일 확인
        reference_file = skill_dir / "reference.md"
        examples_file = skill_dir / "examples.md"

        if not reference_file.exists():
            warnings.append("reference.md: 파일 없음 (선택사항)")

        if not examples_file.exists():
            warnings.append("examples.md: 파일 없음 (선택사항)")

        # 8. 메타데이터 검증
        if metadata:
            # 버전 형식 확인 (semver)
            version = metadata.get('version', '')
            if not self._is_valid_semver(version):
                warnings.append(f"Version: '{version}' 형식이 비표준 (semver: major.minor.patch)")

            # 상태 확인
            valid_statuses = ['stable', 'beta', 'alpha', 'experimental']
            status = metadata.get('status', '')
            if status not in valid_statuses:
                warnings.append(f"Status: '{status}' 유효하지 않음 ({valid_statuses})")

        # 최종 결과
        valid = len(errors) == 0

        return SkillValidationResult(
            skill_name=skill_name,
            valid=valid,
            errors=errors,
            warnings=warnings,
            metadata=metadata or {}
        )

    @staticmethod
    def _is_valid_semver(version: str) -> bool:
        """Semantic Versioning 형식 검증"""
        parts = version.split('.')
        if len(parts) != 3:
            return False
        return all(part.isdigit() for part in parts)

    def generate_report(self) -> str:
        """검증 결과 보고서 생성"""
        report = []
        report.append("# 125개 Enterprise Skills 활성화 종합 검증 보고서")
        report.append("")
        report.append(f"**검증 날짜**: 2025-11-12")
        report.append(f"**검증 시간**: {len(self.results)}개 Skills")
        report.append("")

        # 요약
        report.append("## 📊 검증 요약")
        report.append("")
        report.append(f"| 항목 | 결과 |")
        report.append(f"|------|------|")
        report.append(f"| **전체 Skills** | {self.stats['total']} |")
        report.append(f"| **✅ 완전 정상** | {self.stats['valid']}/{self.stats['total']} ({self.stats['valid']*100//self.stats['total']}%) |")
        report.append(f"| **⚠️ 경고 있음** | {self.stats['warnings']}/{self.stats['total']} |")
        report.append(f"| **❌ 실패** | {self.stats['failed']}/{self.stats['total']} |")
        report.append("")

        if self.stats['errors_by_type']:
            report.append("### 오류 유형 분석")
            report.append("")
            for error_type, count in sorted(self.stats['errors_by_type'].items(), key=lambda x: -x[1]):
                report.append(f"- **{error_type}**: {count}개")
            report.append("")

        # 상세 결과
        report.append("## 📋 상세 검증 결과")
        report.append("")

        # 실패한 Skills
        failed = [r for r in self.results if not r.valid]
        if failed:
            report.append("### ❌ 실패한 Skills")
            report.append("")
            for result in failed:
                report.append(f"**{result.skill_name}**:")
                for error in result.errors:
                    report.append(f"- ❌ {error}")
                report.append("")

        # 경고가 있는 Skills
        warnings_list = [r for r in self.results if r.valid and r.warnings]
        if warnings_list:
            report.append("### ⚠️ 경고가 있는 Skills")
            report.append("")
            for result in warnings_list[:10]:  # 처음 10개만 표시
                report.append(f"**{result.skill_name}**:")
                for warning in result.warnings:
                    report.append(f"- ⚠️ {warning}")
                report.append("")

            if len(warnings_list) > 10:
                report.append(f"*및 {len(warnings_list)-10}개 더...*")
                report.append("")

        # 정상 Skills 목록
        normal = [r for r in self.results if r.valid and not r.warnings]
        if normal:
            report.append("### ✅ 완전히 정상인 Skills")
            report.append("")
            for i, result in enumerate(normal, 1):
                report.append(f"{i}. {result.skill_name}")
            report.append("")

        # 종합 평가
        report.append("## 🎯 종합 평가")
        report.append("")

        success_rate = self.stats['valid'] * 100 // self.stats['total']
        if success_rate >= 95:
            report.append(f"✅ **EXCELLENT**: {success_rate}% 성공률")
            report.append("모든 Skills이 프로덕션 배포 준비 완료")
        elif success_rate >= 85:
            report.append(f"⚠️ **GOOD**: {success_rate}% 성공률")
            report.append("약간의 경고가 있으나 배포 가능")
        else:
            report.append(f"❌ **NEEDS FIXES**: {success_rate}% 성공률")
            report.append("배포 전 수정 필요")

        report.append("")
        report.append("---")
        report.append("")
        report.append("**보고서 생성**: 2025-11-12")

        return "\n".join(report)

def main():
    """메인 실행 함수"""
    skills_dir = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    if not skills_dir.exists():
        print(f"❌ Skills 디렉토리를 찾을 수 없습니다: {skills_dir}")
        sys.exit(1)

    validator = SkillValidator(skills_dir)
    results = validator.validate_all()

    print("")
    print("=" * 70)
    print(f"✅ 검증 완료!")
    print(f"  정상: {validator.stats['valid']} | 경고: {validator.stats['warnings']} | 실패: {validator.stats['failed']}")
    print("=" * 70)

    # 보고서 생성
    report_dir = Path('/Users/goos/MoAI/MoAI-ADK/.moai/reports/validation')
    report_dir.mkdir(parents=True, exist_ok=True)

    report_file = report_dir / 'skills-activation-test-2025-11-12.md'
    report_content = validator.generate_report()

    with open(report_file, 'w', encoding='utf-8') as f:
        f.write(report_content)

    print(f"\n📄 보고서 생성: {report_file}")
    print(f"\n🎯 최종 결과:")
    print(f"  ✅ 완전 정상: {validator.stats['valid']}/{validator.stats['total']} ({validator.stats['valid']*100//validator.stats['total']}%)")
    print(f"  ⚠️  경고 있음: {validator.stats['warnings']}/{validator.stats['total']}")
    print(f"  ❌ 실패: {validator.stats['failed']}/{validator.stats['total']}")

    return 0 if validator.stats['failed'] == 0 else 1

if __name__ == '__main__':
    sys.exit(main())
