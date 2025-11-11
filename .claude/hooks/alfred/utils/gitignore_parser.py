#!/usr/bin/env python3
# @CODE:UTIL-GITIGNORE-001 | SPEC: UTIL-GITIGNORE-001 | TEST: tests/hooks/test_gitignore_parser.py
""".gitignore 파서 유틸리티

.gitignore 파일을 파싱하여 TAG 검증에서 제외할 패턴을 추출합니다.

기능:
- .gitignore 파일 파싱
- 코멘트 및 빈 줄 제외
- glob 패턴 정규화
- 디렉토리 패턴 처리

사용법:
    from .utils.gitignore_parser import load_gitignore_patterns
    patterns = load_gitignore_patterns()
"""

from pathlib import Path
from typing import List, Set


def load_gitignore_patterns(gitignore_path: str = ".gitignore") -> List[str]:
    """.gitignore에서 패턴 로드

    Args:
        gitignore_path: .gitignore 파일 경로 (기본값: ".gitignore")

    Returns:
        제외 패턴 목록
    """
    patterns = []

    try:
        gitignore_file = Path(gitignore_path)
        if not gitignore_file.exists():
            return patterns

        with open(gitignore_file, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()

                # 빈 줄이나 코멘트 제외
                if not line or line.startswith('#'):
                    continue

                # 부정 패턴 (!) 제외 - TAG 검증에서는 사용하지 않음
                if line.startswith('!'):
                    continue

                # 패턴 정규화
                pattern = normalize_pattern(line)
                if pattern:
                    patterns.append(pattern)

    except Exception:
        # 파일 읽기 실패 시 빈 리스트 반환
        pass

    return patterns


def normalize_pattern(pattern: str) -> str:
    """패턴 정규화

    Args:
        pattern: 원본 패턴

    Returns:
        정규화된 패턴
    """
    # 앞뒤 공백 제거
    pattern = pattern.strip()

    # 절대 경로 패턴 (/ 로 시작)
    if pattern.startswith('/'):
        # / 제거하고 반환
        return pattern[1:]

    # 일반 패턴은 그대로 반환 (디렉토리 슬래시 포함)
    return pattern


def is_path_ignored(
    file_path: str,
    ignore_patterns: List[str],
    protected_paths: List[str] = None
) -> bool:
    """파일 경로가 ignore 패턴에 매칭되는지 확인

    Args:
        file_path: 확인할 파일 경로
        ignore_patterns: ignore 패턴 목록
        protected_paths: 보호되어야 하는 경로 (무시하지 않음)

    Returns:
        매칭되면 True
    """
    import fnmatch

    # 기본 보호 경로 설정
    if protected_paths is None:
        protected_paths = [".moai/specs/"]

    # 보호된 경로인 경우 무시하지 않음
    for protected in protected_paths:
        if file_path.startswith(protected):
            return False

    # 경로 정규화
    path_parts = Path(file_path).parts

    for pattern in ignore_patterns:
        # 와일드카드 디렉토리 패턴 (hooks_backup_*/, *_backup_*/)
        if '*' in pattern and pattern.endswith('/'):
            pattern_without_slash = pattern[:-1]
            # 각 경로 부분과 매칭
            for part in path_parts:
                if fnmatch.fnmatch(part, pattern_without_slash):
                    return True

        # 와일드카드 패턴 매칭 (*.ext, *backup*)
        elif '*' in pattern:
            # 전체 경로 매칭
            if fnmatch.fnmatch(file_path, pattern):
                return True
            # 각 경로 부분 매칭
            for part in path_parts:
                if fnmatch.fnmatch(part, pattern):
                    return True

        # 디렉토리 패턴 매칭
        elif pattern.endswith('/'):
            dir_name = pattern[:-1]
            if dir_name in path_parts:
                return True

        # 단순 문자열 매칭 (파일명 또는 디렉토리명)
        else:
            if pattern in file_path:
                return True

    return False


def get_combined_exclude_patterns(
    base_patterns: List[str],
    gitignore_path: str = ".gitignore"
) -> List[str]:
    """기본 패턴과 .gitignore 패턴 결합

    Args:
        base_patterns: 기본 제외 패턴
        gitignore_path: .gitignore 파일 경로

    Returns:
        결합된 제외 패턴 목록 (중복 제거)
    """
    # 기본 패턴 시작
    patterns_set: Set[str] = set(base_patterns)

    # .gitignore 패턴 추가
    gitignore_patterns = load_gitignore_patterns(gitignore_path)
    patterns_set.update(gitignore_patterns)

    # 정렬하여 반환
    return sorted(list(patterns_set))


if __name__ == "__main__":
    # 테스트 코드
    patterns = load_gitignore_patterns()
    print(f"Loaded {len(patterns)} patterns from .gitignore")
    for pattern in patterns[:10]:
        print(f"  - {pattern}")
