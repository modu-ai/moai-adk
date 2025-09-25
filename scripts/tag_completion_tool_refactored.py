# @TASK:TAG-COMPLETION-TOOL-011
"""
TAG 자동 완성 도구 - REFACTOR 버전

TRUST 5원칙을 적용하여 단일 책임 원칙과 가독성을 향상한 리팩토링 버전입니다.
"""

import os
import re
import json
import logging
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, asdict
from enum import Enum


# 구조화된 로깅 설정
def setup_structured_logging():
    """Secured: 구조화된 로깅 설정"""
    class StructuredFormatter(logging.Formatter):
        def format(self, record):
            log_entry = {
                'timestamp': self.formatTime(record),
                'level': record.levelname,
                'message': record.getMessage(),
                'module': record.module,
                'function': record.funcName,
                'line': record.lineno
            }
            # 민감정보 마스킹
            if 'path' in log_entry['message'].lower():
                log_entry['message'] = re.sub(r'/Users/[^/]+', '/Users/***redacted***', log_entry['message'])
            return json.dumps(log_entry)

    logger = logging.getLogger('tag_completion')
    logger.setLevel(logging.INFO)
    handler = logging.StreamHandler()
    handler.setFormatter(StructuredFormatter())
    logger.addHandler(handler)
    return logger


class TagCategory(Enum):
    """16-Core TAG 카테고리 정의"""
    # SPEC Category
    REQ = "REQ"
    DESIGN = "DESIGN"
    TASK = "TASK"

    # PROJECT Category
    VISION = "VISION"
    STRUCT = "STRUCT"
    TECH = "TECH"
    ADR = "ADR"

    # IMPLEMENTATION Category
    FEATURE = "FEATURE"
    API = "API"
    TEST = "TEST"
    DATA = "DATA"

    # QUALITY Category
    PERF = "PERF"
    SEC = "SEC"
    DEBT = "DEBT"
    TODO = "TODO"


@dataclass
class TagSuggestion:
    """TAG 제안 정보"""
    file_path: str
    suggested_tag: str
    confidence: float
    reason: str


class TagMappingRules:
    """TAG 매핑 규칙 관리 - Single Responsibility"""

    def __init__(self):
        self._init_pattern_rules()
        self._init_directory_rules()

    def _init_pattern_rules(self):
        """파일 패턴별 특정 TAG 규칙"""
        self.specific_patterns = {
            'cli/__main__.py': TagCategory.FEATURE,
            'hooks/moai/policy_block.py': TagCategory.SEC,
            'hooks/moai/pre_write_guard.py': TagCategory.SEC,
            'hooks/moai/language_detector.py': TagCategory.FEATURE,
            'hooks/moai/steering_guard.py': TagCategory.SEC,
            'hooks/moai/run_tests_and_report.py': TagCategory.TASK,
            'scripts/check_constitution.py': TagCategory.TASK,
            'scripts/doc_sync.py': TagCategory.TASK,
            'scripts/validate_claude_standards.py': TagCategory.TASK,
            'scripts/check_secrets.py': TagCategory.SEC,
            'scripts/validate_tags.py': TagCategory.TASK,
            'scripts/repair_tags.py': TagCategory.TASK,
            'scripts/check_licenses.py': TagCategory.TASK,
            'scripts/check_traceability.py': TagCategory.TASK,
            'scripts/check_coverage.py': TagCategory.TASK,
            'scripts/validate_stage.py': TagCategory.TASK,
        }

    def _init_directory_rules(self):
        """디렉토리별 기본 TAG 규칙"""
        self.directory_patterns = {
            'cli': TagCategory.FEATURE,
            'core': TagCategory.FEATURE,
            'install': TagCategory.FEATURE,
            'utils': TagCategory.FEATURE,
            'commands': TagCategory.FEATURE,
        }

    def suggest_category(self, file_path: str) -> TagCategory:
        """파일 경로 기반 TAG 카테고리 제안"""
        # 특정 패턴 먼저 확인
        for pattern, category in self.specific_patterns.items():
            if pattern in file_path:
                return category

        # 디렉토리 패턴 확인
        for directory, category in self.directory_patterns.items():
            if directory in file_path:
                return category

        # 기본값
        return TagCategory.TASK

    def generate_tag_id(self, file_path: str, category: TagCategory) -> str:
        """TAG ID 생성"""
        # 파일 경로에서 도메인 추출
        path_parts = Path(file_path).parts

        # 의미있는 도메인 이름 생성
        if 'cli' in path_parts:
            domain = "CLI-COMPONENT"
        elif 'core' in path_parts:
            domain = "CORE-COMPONENT"
        elif 'templates' in path_parts:
            if 'hook' in file_path:
                domain = "TEMPLATE-HOOK"
            elif 'script' in file_path:
                domain = "TEMPLATE-SCRIPT"
            else:
                domain = "TEMPLATE-RESOURCE"
        elif 'install' in path_parts:
            domain = "INSTALL-COMPONENT"
        elif 'utils' in path_parts:
            domain = "UTILS-COMPONENT"
        else:
            # 파일명에서 도메인 추출 시도
            file_stem = Path(file_path).stem.upper().replace('_', '-')
            domain = file_stem[:20]  # 길이 제한

        return f"@{category.value}:{domain}-011"


class TagScanner:
    """TAG 스캔 전용 클래스 - Single Responsibility"""

    def __init__(self, src_dir: str):
        self.src_dir = src_dir
        self.tag_pattern = re.compile(r'@[A-Z]+:[A-Z-]+-\d+')

    def find_all_python_files(self) -> List[str]:
        """모든 Python 파일 찾기"""
        python_files = []
        for root, dirs, files in os.walk(self.src_dir):
            for file in files:
                if file.endswith('.py'):
                    python_files.append(os.path.join(root, file))
        return python_files

    def has_tag(self, file_path: str) -> bool:
        """파일에 TAG가 있는지 확인"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                return bool(self.tag_pattern.search(content))
        except (UnicodeDecodeError, OSError):
            return False

    def find_missing_tag_files(self) -> List[str]:
        """TAG가 없는 파일들 찾기"""
        missing_files = []
        for file_path in self.find_all_python_files():
            if not self.has_tag(file_path):
                missing_files.append(file_path)
        return missing_files


class TagApplicator:
    """TAG 적용 전용 클래스 - Single Responsibility"""

    def add_tag_to_file(self, file_path: str, tag: str, dry_run: bool = True) -> bool:
        """파일에 TAG 추가"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            # 이미 TAG가 있으면 스킵
            if re.search(r'@[A-Z]+:[A-Z-]+-\d+', content):
                return True

            # TAG 삽입 위치 결정
            lines = content.split('\n')
            insert_pos = self._find_insert_position(lines)

            # TAG 추가
            tag_comment = f"# {tag}"
            new_lines = lines[:insert_pos] + [tag_comment] + lines[insert_pos:]
            new_content = '\n'.join(new_lines)

            if not dry_run:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)

            return True

        except Exception as e:
            print(f"Error processing {file_path}: {e}")
            return False

    def _find_insert_position(self, lines: List[str]) -> int:
        """TAG 삽입 위치 찾기"""
        for i, line in enumerate(lines):
            if line.startswith('#!') or 'coding:' in line or 'encoding:' in line:
                continue
            elif line.strip().startswith('"""') or line.strip().startswith("'''"):
                return i  # docstring 전에 삽입
            elif line.strip() and not line.startswith('#'):
                return i  # 첫 번째 코드 라인 전에 삽입
        return 0


class TagCompletionOrchestrator:
    """TAG 완성 프로세스 오케스트레이터 - Facade Pattern"""

    def __init__(self, project_root: str = None):
        self.project_root = project_root or "/Users/goos/MoAI/MoAI-ADK"
        self.src_dir = os.path.join(self.project_root, "src")

        # 컴포넌트 초기화
        self.mapping_rules = TagMappingRules()
        self.scanner = TagScanner(self.src_dir)
        self.applicator = TagApplicator()

    def analyze_current_state(self) -> Dict[str, int]:
        """현재 TAG 상태 분석"""
        all_files = self.scanner.find_all_python_files()
        missing_files = self.scanner.find_missing_tag_files()

        return {
            'total_files': len(all_files),
            'tagged_files': len(all_files) - len(missing_files),
            'missing_files': len(missing_files),
            'coverage_percent': int((len(all_files) - len(missing_files)) / len(all_files) * 100)
        }

    def generate_suggestions(self) -> List[TagSuggestion]:
        """TAG 제안 생성"""
        missing_files = self.scanner.find_missing_tag_files()
        suggestions = []

        for file_path in missing_files:
            category = self.mapping_rules.suggest_category(file_path)
            suggested_tag = self.mapping_rules.generate_tag_id(file_path, category)

            suggestion = TagSuggestion(
                file_path=file_path,
                suggested_tag=suggested_tag,
                confidence=0.9,  # 고정된 confidence
                reason=f"Based on file path pattern and {category.value} category"
            )
            suggestions.append(suggestion)

        return suggestions

    def apply_suggestions(self, suggestions: List[TagSuggestion], dry_run: bool = True) -> Dict[str, bool]:
        """제안된 TAG들을 적용"""
        results = {}

        for suggestion in suggestions:
            success = self.applicator.add_tag_to_file(
                suggestion.file_path,
                suggestion.suggested_tag,
                dry_run
            )
            results[suggestion.file_path] = success

            if not dry_run and success:
                print(f"✓ Added: {suggestion.suggested_tag} -> {os.path.relpath(suggestion.file_path)}")
            elif dry_run and success:
                print(f"[DRY] Would add: {suggestion.suggested_tag} -> {os.path.relpath(suggestion.file_path)}")

        return results

    def run_completion_process(self, dry_run: bool = True) -> Dict:
        """완전한 TAG 완성 프로세스 실행"""
        print("🗿 MoAI-ADK TAG Completion Tool - REFACTORED")
        print("=" * 50)

        # 현재 상태 분석
        state = self.analyze_current_state()
        print(f"📊 Current State:")
        print(f"   Total files: {state['total_files']}")
        print(f"   Tagged files: {state['tagged_files']}")
        print(f"   Missing tags: {state['missing_files']}")
        print(f"   Coverage: {state['coverage_percent']}%")
        print()

        if state['missing_files'] == 0:
            print("🎉 All files already have @TAG!")
            return state

        # 제안 생성
        suggestions = self.generate_suggestions()
        print(f"🔍 Generated {len(suggestions)} tag suggestions")

        # 제안 적용
        results = self.apply_suggestions(suggestions, dry_run)
        success_count = sum(1 for success in results.values() if success)

        print()
        print(f"📈 Process Summary:")
        print(f"   Processed: {len(results)} files")
        print(f"   Successful: {success_count}")
        print(f"   Failed: {len(results) - success_count}")

        if not dry_run:
            # 최종 상태 확인
            final_state = self.analyze_current_state()
            print(f"   Final coverage: {final_state['coverage_percent']}%")

            if final_state['coverage_percent'] == 100:
                print("🎉 TAG completion achieved 100% coverage!")

        return {
            'initial_state': state,
            'suggestions': len(suggestions),
            'results': results,
            'success_rate': success_count / len(results) if results else 0
        }


def main():
    """CLI 실행 함수"""
    import argparse

    parser = argparse.ArgumentParser(description="TAG Completion Tool - REFACTORED")
    parser.add_argument("--dry-run", action="store_true", default=True,
                       help="Show what would be done without making changes")
    parser.add_argument("--execute", action="store_true",
                       help="Actually apply the changes")

    args = parser.parse_args()

    orchestrator = TagCompletionOrchestrator()
    dry_run = not args.execute

    result = orchestrator.run_completion_process(dry_run=dry_run)

    # 종료 코드 결정
    if result.get('success_rate', 0) == 1.0:
        exit_code = 0  # 성공
    else:
        exit_code = 1  # 부분 실패

    exit(exit_code)


if __name__ == "__main__":
    main()