#!/usr/bin/env python3
# @CODE:LANGUAGE-DIRS-001 | @SPEC:TAG-LANGUAGE-DETECTION-001 | @DOC:LANGUAGE-DIRS-CONFIG-001
"""언어별 코드 디렉토리 감지 및 설정

프로젝트 언어에 따라 기대되는 코드 디렉토리를 자동으로 감지하고,
사용자 정의 설정과 병합하여 최종 디렉토리 패턴을 반환합니다.

지원하는 언어 (10개):
  - Python, JavaScript, TypeScript
  - Go, Rust
  - Kotlin, Ruby, PHP
  - Java, C#
"""

from typing import Dict, List, Optional, Set
from pathlib import Path


# Language-specific code directory patterns
# LANGUAGE_DIRECTORY_MAP: 각 언어별 일반적인 코드 디렉토리 패턴
LANGUAGE_DIRECTORY_MAP: Dict[str, List[str]] = {
    "python": [
        "src/",
        "lib/",
        "{package_name}/",  # Package name replaces with actual name
    ],
    "javascript": [
        "src/",
        "lib/",
        "app/",
        "pages/",
        "components/",
    ],
    "typescript": [
        "src/",
        "lib/",
        "app/",
        "pages/",
        "components/",
    ],
    "go": [
        "cmd/",
        "pkg/",
        "internal/",
    ],
    "rust": [
        "src/",
        "crates/",
    ],
    "kotlin": [
        "src/main/kotlin/",
        "src/test/kotlin/",
    ],
    "ruby": [
        "lib/",
        "app/",
    ],
    "php": [
        "src/",
        "app/",
    ],
    "java": [
        "src/main/java/",
        "src/test/java/",
    ],
    "csharp": [
        "src/",
        "App/",
    ],
}

# 모든 언어 공통 제외 패턴
COMMON_EXCLUDE_PATTERNS: List[str] = [
    "tests/",
    "test/",
    "__tests__/",
    "spec/",
    "specs/",
    ".moai/",
    ".claude/",
    "node_modules/",
    "dist/",
    "build/",
    ".next/",
    ".nuxt/",
    "examples/",
    "docs/",
    "documentation/",
    "templates/",
    ".git/",
    ".github/",
    "venv/",
    ".venv/",
    "vendor/",
    "target/",
    "bin/",
    "__pycache__/",
    "*.egg-info/",
]


def detect_directories(
    config: Optional[Dict] = None,
    language: Optional[str] = None,
) -> List[str]:
    """
    프로젝트 설정과 언어에 따라 코드 디렉토리 패턴 반환.

    동작 순서:
    1. config에 커스텀 패턴이 있으면 → 커스텀 패턴 사용
    2. 커스텀 없으면 → 언어별 기본 패턴 사용
    3. 하이브리드 모드 → 언어 기본 + 커스텀 병합

    Args:
        config: 프로젝트 설정 (예: .moai/config.json)
        language: 프로젝트 언어 (config에 없는 경우 직접 전달 가능)

    Returns:
        감지된 코드 디렉토리 패턴 리스트
    """
    config = config or {}

    # 언어 결정
    if language is None:
        language = config.get("project", {}).get("language", "python").lower()

    # 커스텀 설정 추출
    tags_policy = config.get("tags", {}).get("policy", {})
    code_dirs_config = tags_policy.get("code_directories", {})

    detection_mode = code_dirs_config.get("detection_mode", "auto")
    custom_patterns = code_dirs_config.get("patterns", [])

    # Mode별 처리
    if detection_mode == "manual" and custom_patterns:
        # 수동 모드: 커스텀 패턴만 사용
        return custom_patterns

    elif detection_mode == "auto":
        # 자동 모드: 언어별 기본 패턴 사용
        return LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])

    elif detection_mode == "hybrid":
        # 하이브리드 모드: 언어 기본 + 커스텀 병합
        base_patterns = LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])
        combined = list(set(base_patterns + custom_patterns))
        return sorted(combined)

    # 기본값: 자동 감지
    return LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])


def get_exclude_patterns(
    config: Optional[Dict] = None,
) -> List[str]:
    """
    제외할 디렉토리 패턴 반환.

    순서:
    1. config에 커스텀 제외 패턴이 있으면 → 커스텀 사용
    2. 없으면 → 공통 제외 패턴 사용
    3. 병합 모드 → 공통 + 커스텀 합치기

    Args:
        config: 프로젝트 설정

    Returns:
        제외할 디렉토리 패턴 리스트
    """
    config = config or {}

    tags_policy = config.get("tags", {}).get("policy", {})
    code_dirs_config = tags_policy.get("code_directories", {})

    custom_exclude = code_dirs_config.get("exclude_patterns", [])
    merge_with_common = code_dirs_config.get("merge_exclude_patterns", True)

    if custom_exclude and not merge_with_common:
        # 커스텀만 사용
        return custom_exclude

    if custom_exclude and merge_with_common:
        # 공통 + 커스텀 병합
        return list(set(COMMON_EXCLUDE_PATTERNS + custom_exclude))

    # 기본값: 공통 제외 패턴만
    return COMMON_EXCLUDE_PATTERNS


def is_code_directory(
    path: Path,
    config: Optional[Dict] = None,
    language: Optional[str] = None,
) -> bool:
    """
    주어진 경로가 코드 디렉토리인지 확인.

    Args:
        path: 확인할 경로
        config: 프로젝트 설정
        language: 프로젝트 언어

    Returns:
        True if path is a code directory, False otherwise
    """
    code_dirs = detect_directories(config, language)
    exclude_patterns = get_exclude_patterns(config)

    path_str = str(path)

    # 제외 패턴 확인
    for exclude in exclude_patterns:
        if exclude.endswith("/"):
            if path_str.startswith(exclude) or f"/{exclude}" in path_str:
                return False
        else:
            if exclude in path_str:
                return False

    # 코드 디렉토리 패턴 확인
    for code_dir in code_dirs:
        if code_dir.endswith("/"):
            if path_str.startswith(code_dir) or f"/{code_dir}" in path_str:
                return True
        else:
            if code_dir in path_str:
                return True

    return False


def get_language_by_file_extension(file_path: Path) -> Optional[str]:
    """
    파일 확장자에서 언어 추론.

    Args:
        file_path: 파일 경로

    Returns:
        추론된 언어, 또는 None
    """
    extension_map = {
        ".py": "python",
        ".js": "javascript",
        ".jsx": "javascript",
        ".ts": "typescript",
        ".tsx": "typescript",
        ".go": "go",
        ".rs": "rust",
        ".kt": "kotlin",
        ".kts": "kotlin",
        ".rb": "ruby",
        ".php": "php",
        ".java": "java",
        ".cs": "csharp",
    }

    suffix = file_path.suffix.lower()
    return extension_map.get(suffix)


def get_all_supported_languages() -> List[str]:
    """
    지원하는 모든 언어 반환.

    Returns:
        지원하는 언어 리스트 (정렬됨)
    """
    return sorted(LANGUAGE_DIRECTORY_MAP.keys())


def validate_language(language: str) -> bool:
    """
    언어가 지원되는지 확인.

    Args:
        language: 확인할 언어

    Returns:
        True if supported, False otherwise
    """
    return language.lower() in LANGUAGE_DIRECTORY_MAP


def merge_directory_patterns(
    base_patterns: List[str],
    custom_patterns: List[str],
) -> List[str]:
    """
    기본 패턴과 커스텀 패턴을 병합.

    중복 제거 후 정렬하여 반환.

    Args:
        base_patterns: 기본 디렉토리 패턴
        custom_patterns: 커스텀 디렉토리 패턴

    Returns:
        병합된 패턴 (중복 제거, 정렬됨)
    """
    merged = list(set(base_patterns + custom_patterns))
    return sorted(merged)


def expand_package_placeholder(
    patterns: List[str],
    package_name: Optional[str] = None,
) -> List[str]:
    """
    패턴의 {package_name} 플레이스홀더를 실제 패키지명으로 치환.

    Args:
        patterns: 원본 패턴
        package_name: 패키지명 (없으면 플레이스홀더 유지)

    Returns:
        치환된 패턴
    """
    if not package_name:
        return [p for p in patterns if "{package_name}" not in p]

    expanded = []
    for pattern in patterns:
        if "{package_name}" in pattern:
            expanded.append(pattern.replace("{package_name}", package_name))
        else:
            expanded.append(pattern)

    return expanded
