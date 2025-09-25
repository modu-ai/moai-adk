# @TASK:TAG-COMPLETION-TOOL-011
"""
TAG 자동 완성 도구 - SPEC-011 구현

누락된 @TAG를 자동으로 파일에 추가하는 최소 구현입니다.
TRUST 5원칙에 따라 단순하고 안전한 접근법을 사용합니다.
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple


class TagMappingRules:
    """TAG 매핑 규칙 관리 - Single Responsibility Principle"""

    def __init__(self):
        self.mapping_rules = {
            'cli/__main__.py': '@FEATURE:CLI-ENTRY-011',
            'hooks/moai/policy_block.py': '@SEC:POLICY-BLOCK-011',
            'hooks/moai/pre_write_guard.py': '@SEC:PRE-WRITE-GUARD-011',
            'hooks/moai/language_detector.py': '@FEATURE:LANGUAGE-DETECT-011',
            'hooks/moai/steering_guard.py': '@SEC:STEERING-GUARD-011',
            'hooks/moai/run_tests_and_report.py': '@TASK:TEST-REPORT-011',
            'scripts/check_constitution.py': '@TASK:CONSTITUTION-CHECK-011',
            'scripts/doc_sync.py': '@TASK:DOC-SYNC-011',
            'scripts/validate_claude_standards.py': '@TASK:CLAUDE-STANDARDS-011',
            'scripts/check-traceability.py': '@TASK:TRACEABILITY-CHECK-011',
            'scripts/check_secrets.py': '@SEC:SECRETS-CHECK-011',
            'scripts/validate_tags.py': '@TASK:TAG-VALIDATE-011',
            'scripts/check_constitution.py': '@TASK:CONSTITUTION-CHECK-011',
            'scripts/repair_tags.py': '@TASK:TAG-REPAIR-011',
            'scripts/check_licenses.py': '@TASK:LICENSE-CHECK-011',
            'scripts/check_traceability.py': '@TASK:TRACEABILITY-CHECK-011',
            'scripts/check_coverage.py': '@TASK:COVERAGE-CHECK-011',
            'scripts/validate_stage.py': '@TASK:STAGE-VALIDATE-011'
        }

    def find_missing_tags(self) -> List[str]:
        """@TAG가 없는 파일들을 찾는다"""
        missing_tag_files = []

        for root, dirs, files in os.walk(self.src_dir):
            for file in files:
                if file.endswith('.py'):
                    file_path = os.path.join(root, file)
                    try:
                        with open(file_path, 'r', encoding='utf-8') as f:
                            content = f.read()
                            if not self.tag_pattern.search(content):
                                missing_tag_files.append(file_path)
                    except (UnicodeDecodeError, OSError):
                        # 파일 읽기 실패 시 누락된 것으로 간주
                        missing_tag_files.append(file_path)

        return missing_tag_files

    def suggest_tag_for_file(self, file_path: str) -> str:
        """파일 경로를 기반으로 적절한 @TAG 제안"""
        # 파일 경로에서 상대 경로 추출
        relative_path = os.path.relpath(file_path, self.src_dir)

        # 매핑 규칙에서 검색
        for pattern, tag in self.tag_mapping_rules.items():
            if pattern in relative_path:
                return tag

        # 기본 규칙: 파일 위치와 이름을 기반으로 TAG 생성
        if 'cli' in relative_path:
            return '@FEATURE:CLI-COMPONENT-011'
        elif 'core' in relative_path:
            return '@FEATURE:CORE-COMPONENT-011'
        elif 'install' in relative_path:
            return '@FEATURE:INSTALL-COMPONENT-011'
        elif 'utils' in relative_path:
            return '@FEATURE:UTILS-COMPONENT-011'
        elif 'templates' in relative_path:
            if 'hook' in relative_path:
                return '@TASK:TEMPLATE-HOOK-011'
            elif 'script' in relative_path:
                return '@TASK:TEMPLATE-SCRIPT-011'
            else:
                return '@TASK:TEMPLATE-RESOURCE-011'
        else:
            # 최후의 기본 TAG
            return '@TASK:CODE-COMPONENT-011'

    def add_tag_to_file(self, file_path: str, tag: str, dry_run: bool = True) -> bool:
        """파일에 @TAG를 추가한다"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # 이미 TAG가 있는지 확인
            if self.tag_pattern.search(content):
                print(f"File already has tag: {file_path}")
                return True

            # TAG를 추가할 위치 결정 (파일 시작 부분의 주석 블록)
            lines = content.split('\n')
            insert_position = 0

            # shebang이나 encoding 선언 이후에 삽입
            for i, line in enumerate(lines):
                if line.startswith('#!') or 'coding:' in line or 'encoding:' in line:
                    insert_position = i + 1
                elif line.strip().startswith('"""') or line.strip().startswith('"""'):
                    # docstring 시작 전에 삽입
                    break
                elif line.strip() and not line.startswith('#'):
                    # 첫 번째 코드 라인 전에 삽입
                    break

            # TAG 주석 생성
            tag_comment = f"# {tag}"

            # 내용 수정
            new_lines = lines[:insert_position] + [tag_comment] + lines[insert_position:]
            new_content = '\n'.join(new_lines)

            if not dry_run:
                # 실제 파일 수정
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)
                print(f"Added tag to: {file_path} -> {tag}")
            else:
                print(f"[DRY RUN] Would add tag to: {file_path} -> {tag}")

            return True

        except Exception as e:
            print(f"Error processing file {file_path}: {e}")
            return False

    def apply_tags(self, dry_run: bool = True) -> Dict[str, str]:
        """누락된 파일에 TAG 일괄 적용"""
        missing_files = self.find_missing_tags()
        results = {}

        print(f"Found {len(missing_files)} files without @TAG")

        for file_path in missing_files:
            suggested_tag = self.suggest_tag_for_file(file_path)
            success = self.add_tag_to_file(file_path, suggested_tag, dry_run)
            results[file_path] = suggested_tag if success else "ERROR"

        return results

    def validate_completion(self) -> Tuple[int, int, float]:
        """TAG 적용 완료 후 검증"""
        all_files = []
        for root, dirs, files in os.walk(self.src_dir):
            for file in files:
                if file.endswith('.py'):
                    all_files.append(os.path.join(root, file))

        missing_files = self.find_missing_tags()
        total_files = len(all_files)
        tagged_files = total_files - len(missing_files)
        coverage = tagged_files / total_files if total_files > 0 else 0

        return total_files, tagged_files, coverage


def main():
    """CLI 실행 함수"""
    import argparse

    parser = argparse.ArgumentParser(description="TAG Completion Tool - SPEC-011")
    parser.add_argument("--dry-run", action="store_true", default=True,
                       help="Show what would be done without making changes")
    parser.add_argument("--execute", action="store_true",
                       help="Actually apply the changes")

    args = parser.parse_args()

    tool = TagCompletionTool()

    # 현재 상태 출력
    total, tagged, coverage = tool.validate_completion()
    print(f"Current Status:")
    print(f"  Total Python files: {total}")
    print(f"  Files with @TAG: {tagged}")
    print(f"  Coverage: {coverage:.2%}")
    print()

    # TAG 적용
    dry_run = not args.execute
    results = tool.apply_tags(dry_run=dry_run)

    print(f"\nProcessed {len(results)} files:")
    for file_path, result in results.items():
        status = "✓" if result != "ERROR" else "✗"
        relative_path = os.path.relpath(file_path)
        print(f"  {status} {relative_path} -> {result}")

    # 완료 후 상태
    if not dry_run:
        total, tagged, coverage = tool.validate_completion()
        print(f"\nFinal Status:")
        print(f"  Coverage: {coverage:.2%}")

        if coverage >= 1.0:
            print("🎉 All files now have @TAG!")
        else:
            remaining = total - tagged
            print(f"⚠️  {remaining} files still missing @TAG")


if __name__ == "__main__":
    main()