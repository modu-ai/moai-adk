#!/usr/bin/env python3
"""Project metadata utilities

프로젝트 정보 조회 (언어, Git, SPEC 진행도 등)
"""

import json
import subprocess
from pathlib import Path
from typing import Any


def detect_language(cwd: str) -> str:
    """프로젝트 언어 감지 (20개 언어 지원)

    파일 시스템을 탐색하여 프로젝트의 주 개발 언어를 감지합니다.
    pyproject.toml, tsconfig.json 등의 설정 파일을 우선 검사하며,
    TypeScript 우선 원칙을 적용합니다 (tsconfig.json 존재 시).

    Args:
        cwd: 프로젝트 루트 디렉토리 경로 (절대/상대 경로 모두 가능)

    Returns:
        감지된 언어명 (소문자). 감지 실패 시 "Unknown Language" 반환.
        지원 언어: python, typescript, javascript, java, go, rust,
                  dart, swift, kotlin, php, ruby, elixir, scala,
                  clojure, cpp, c, csharp, haskell, shell, lua

    Examples:
        >>> detect_language("/path/to/python/project")
        'python'
        >>> detect_language("/path/to/typescript/project")
        'typescript'
        >>> detect_language("/path/to/unknown/project")
        'Unknown Language'

    TDD History:
        - RED: 21개 언어 감지 테스트 작성 (20개 언어 + 1개 unknown)
        - GREEN: 20개 언어 + unknown 구현, 모든 테스트 통과
        - REFACTOR: 파일 검사 순서 최적화, TypeScript 우선 원칙 적용
    """
    cwd_path = Path(cwd)

    # Language detection mapping
    language_files = {
        "pyproject.toml": "python",
        "tsconfig.json": "typescript",
        "package.json": "javascript",
        "pom.xml": "java",
        "go.mod": "go",
        "Cargo.toml": "rust",
        "pubspec.yaml": "dart",
        "Package.swift": "swift",
        "build.gradle.kts": "kotlin",
        "composer.json": "php",
        "Gemfile": "ruby",
        "mix.exs": "elixir",
        "build.sbt": "scala",
        "project.clj": "clojure",
        "CMakeLists.txt": "cpp",
        "Makefile": "c",
    }

    # Check standard language files
    for file_name, language in language_files.items():
        if (cwd_path / file_name).exists():
            # Special handling for package.json - prefer typescript if tsconfig exists
            if file_name == "package.json" and (cwd_path / "tsconfig.json").exists():
                return "typescript"
            return language

    # Check for C# project files (*.csproj)
    if any(cwd_path.glob("*.csproj")):
        return "csharp"

    # Check for Haskell project files (*.cabal)
    if any(cwd_path.glob("*.cabal")):
        return "haskell"

    # Check for Shell scripts (*.sh)
    if any(cwd_path.glob("*.sh")):
        return "shell"

    # Check for Lua files (*.lua)
    if any(cwd_path.glob("*.lua")):
        return "lua"

    return "Unknown Language"


def _run_git_command(args: list[str], cwd: str, timeout: int = 2) -> str:
    """Git 명령어 실행 헬퍼 함수

    Git 명령어를 안전하게 실행하고 출력을 반환합니다.
    코드 중복을 제거하고 일관된 에러 처리를 제공합니다.

    Args:
        args: Git 명령어 인자 리스트 (git은 자동 추가)
        cwd: 실행 디렉토리 경로
        timeout: 타임아웃 (초, 기본 2초)

    Returns:
        Git 명령어 출력 (stdout, 앞뒤 공백 제거)

    Raises:
        subprocess.TimeoutExpired: 타임아웃 초과
        subprocess.CalledProcessError: Git 명령어 실패

    Examples:
        >>> _run_git_command(["branch", "--show-current"], ".")
        'main'
    """
    result = subprocess.run(
        ["git"] + args,
        cwd=cwd,
        capture_output=True,
        text=True,
        timeout=timeout,
        check=True,
    )
    return result.stdout.strip()


def get_git_info(cwd: str) -> dict[str, Any]:
    """Git 리포지토리 정보 수집

    Git 리포지토리의 현재 상태를 조회합니다.
    브랜치명, 커밋 해시, 변경사항 개수를 반환하며,
    Git 리포지토리가 아닌 경우 빈 딕셔너리를 반환합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        Git 정보 딕셔너리. 다음 키를 포함:
        - branch: 현재 브랜치명 (str)
        - commit: 현재 커밋 해시 (str, full hash)
        - changes: 변경된 파일 개수 (int, staged + unstaged)

        Git 리포지토리가 아니거나 조회 실패 시 빈 딕셔너리 {}

    Examples:
        >>> get_git_info("/path/to/git/repo")
        {'branch': 'main', 'commit': 'abc123...', 'changes': 3}
        >>> get_git_info("/path/to/non-git")
        {}

    Notes:
        - 타임아웃: 각 Git 명령어 2초
        - 보안: subprocess.run(shell=False)로 안전한 실행
        - 에러 처리: 모든 예외 시 빈 딕셔너리 반환

    TDD History:
        - RED: 3개 시나리오 테스트 (Git 리포, 비 Git, 에러)
        - GREEN: subprocess 기반 Git 명령어 실행 구현
        - REFACTOR: 타임아웃 추가 (2초), 예외 처리 강화, 헬퍼 함수로 중복 제거
    """
    try:
        # Check if it's a git repository
        _run_git_command(["rev-parse", "--git-dir"], cwd)

        # Get branch name, commit hash, and changes
        branch = _run_git_command(["branch", "--show-current"], cwd)
        commit = _run_git_command(["rev-parse", "HEAD"], cwd)
        status_output = _run_git_command(["status", "--short"], cwd)
        changes = len([line for line in status_output.splitlines() if line])

        return {
            "branch": branch,
            "commit": commit,
            "changes": changes,
        }

    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        return {}


def count_specs(cwd: str) -> dict[str, int]:
    """SPEC 파일 카운트 및 진행도 계산

    .moai/specs/ 디렉토리를 탐색하여 SPEC 파일 개수와
    완료 상태(status: completed)인 SPEC 개수를 집계합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        SPEC 진행도 딕셔너리. 다음 키를 포함:
        - completed: 완료된 SPEC 개수 (int)
        - total: 전체 SPEC 개수 (int)
        - percentage: 완료율 (int, 0~100)

        .moai/specs/ 디렉토리가 없으면 모두 0

    Examples:
        >>> count_specs("/path/to/project")
        {'completed': 2, 'total': 5, 'percentage': 40}
        >>> count_specs("/path/to/no-specs")
        {'completed': 0, 'total': 0, 'percentage': 0}

    Notes:
        - SPEC 파일 위치: .moai/specs/SPEC-{ID}/spec.md
        - 완료 조건: YAML front matter에 "status: completed" 포함
        - 파싱 실패 시 해당 SPEC은 미완료로 간주

    TDD History:
        - RED: 5개 시나리오 테스트 (0/0, 2/5, 5/5, 디렉토리 없음, 파싱 에러)
        - GREEN: Path.iterdir()로 SPEC 탐색, YAML 파싱 구현
        - REFACTOR: 예외 처리 강화, 퍼센트 계산 안전성 개선
    """
    specs_dir = Path(cwd) / ".moai" / "specs"

    if not specs_dir.exists():
        return {"completed": 0, "total": 0, "percentage": 0}

    completed = 0
    total = 0

    for spec_dir in specs_dir.iterdir():
        if not spec_dir.is_dir() or not spec_dir.name.startswith("SPEC-"):
            continue

        spec_file = spec_dir / "spec.md"
        if not spec_file.exists():
            continue

        total += 1

        # Parse YAML front matter
        try:
            content = spec_file.read_text()
            if content.startswith("---"):
                yaml_end = content.find("---", 3)
                if yaml_end > 0:
                    yaml_content = content[3:yaml_end]
                    if "status: completed" in yaml_content:
                        completed += 1
        except (OSError, UnicodeDecodeError):
            # 파일 읽기 실패 또는 인코딩 오류 - 미완료로 간주
            pass

    percentage = int(completed / total * 100) if total > 0 else 0

    return {
        "completed": completed,
        "total": total,
        "percentage": percentage,
    }


def get_project_language(cwd: str) -> str:
    """Determine the primary project language (prefers config.json).

    Args:
        cwd: Project root directory.

    Returns:
        Language string in lower-case.

    Notes:
        - Reads ``.moai/config.json`` first for a quick answer.
        - Falls back to ``detect_language`` if configuration is missing.
    """
    config_path = Path(cwd) / ".moai" / "config.json"
    if config_path.exists():
        try:
            config = json.loads(config_path.read_text())
            lang = config.get("language", "")
            if lang:
                return lang
        except (OSError, json.JSONDecodeError):
            # Fall back to detection on parse errors
            pass

    # Fall back to the original language detection routine
    return detect_language(cwd)


__all__ = [
    "detect_language",
    "get_git_info",
    "count_specs",
    "get_project_language",
]
