#!/usr/bin/env python3
# @CODE:TEST-INTEGRATION-001 | SPEC: SPEC-TEST-INTEGRATION-001.md | TEST: tests/unit/test_moai_hooks.py
"""Self-contained Python hook script for MoAI-ADK Claude Code integration

@TAG:TEST-INTEGRATION-001
- SPEC: .moai/specs/SPEC-TEST-INTEGRATION-001/spec.md
- TEST: tests/unit/test_moai_hooks.py (49 tests, 94% coverage)
- VERSION: 0.1.0 (TDD 구현 완료)

TDD History:
- RED: 49개 테스트 작성 (21개 언어 감지, 10개 핸들러, 18개 JIT)
- GREEN: 모든 테스트 통과 구현 (20개 언어 지원, 9개 핸들러)
- REFACTOR: 문서화 강화, 아키텍처 명확화, 품질 개선

Architecture:
┌─────────────────────────────────────────────────────────────┐
│ Helpers (독립 함수 - 외부 의존성 없음)                       │
├─────────────────────────────────────────────────────────────┤
│ - detect_language(cwd) -> str                               │
│   20개 언어 자동 감지 (Python, TypeScript, Java, Go 등)     │
│                                                             │
│ - get_git_info(cwd) -> dict                                 │
│   Git 상태 조회 (branch, commit, changes)                   │
│                                                             │
│ - count_specs(cwd) -> dict                                  │
│   SPEC 진행도 조회 (completed, total, percentage)           │
│                                                             │
│ - get_jit_context(prompt, cwd) -> list[str]                 │
│   프롬프트 기반 JIT Context Retrieval                        │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Handlers (helper 함수 조합 - 이벤트 처리)                    │
├─────────────────────────────────────────────────────────────┤
│ - handle_session_start(payload) -> HookResult               │
│   세션 시작 시 프로젝트 상태 요약 표시                       │
│                                                             │
│ - handle_user_prompt_submit(payload) -> HookResult          │
│   사용자 프롬프트 기반 JIT 문서 로딩                         │
│                                                             │
│ - handle_pre_compact(payload) -> HookResult                 │
│   컨텍스트 초과 시 새 세션 제안                              │
│                                                             │
│ - handle_session_end, handle_pre_tool_use, ...             │
│   (기타 6개 핸들러 - 기본 구현)                             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Main Entry Point                                            │
├─────────────────────────────────────────────────────────────┤
│ - main() -> None                                            │
│   CLI 진입점, 이벤트 라우팅, JSON I/O 처리                   │
└─────────────────────────────────────────────────────────────┘

Usage:
    # Claude Code가 자동으로 호출 (사용자가 직접 실행하지 않음)
    python moai_hooks.py SessionStart < payload.json
    python moai_hooks.py UserPromptSubmit < payload.json

Dependencies:
    - Python 3.12+ (dataclasses, pathlib, subprocess)
    - No external packages required (self-contained)
"""

import json
import subprocess
import sys
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any, NotRequired, TypedDict


class HookPayload(TypedDict):
    """Claude Code Hook 이벤트 페이로드 타입 정의

    Claude Code가 Hook 스크립트에 전달하는 데이터 구조.
    이벤트에 따라 필드가 다를 수 있으므로 NotRequired 사용.
    """
    cwd: str
    userPrompt: NotRequired[str]  # UserPromptSubmit 이벤트만 포함


@dataclass
class HookResult:
    """Hook 실행 결과"""

    message: str | None = None
    blocked: bool = False
    contextFiles: list[str] = field(default_factory=list)
    suggestions: list[str] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        """딕셔너리로 변환"""
        return asdict(self)


# ============================================================================
# Helper Functions
# ============================================================================


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


def get_jit_context(prompt: str, cwd: str) -> list[str]:
    """프롬프트 기반 JIT Context Retrieval

    사용자 프롬프트를 분석하여 관련 문서를 자동으로 추천합니다.
    Alfred 커맨드, 키워드 기반 패턴 매칭으로 필요한 문서만 로드합니다.

    Args:
        prompt: 사용자 입력 프롬프트 (대소문자 무관)
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        추천 문서 경로 리스트 (상대 경로).
        매칭되는 패턴이 없거나 파일이 없으면 빈 리스트 []

    Patterns:
        - "/alfred:1-spec" → .moai/memory/spec-metadata.md
        - "/alfred:2-build" → .moai/memory/development-guide.md
        - "test" → tests/ (디렉토리가 존재하는 경우)

    Examples:
        >>> get_jit_context("/alfred:1-spec", "/project")
        ['.moai/memory/spec-metadata.md']
        >>> get_jit_context("implement test", "/project")
        ['tests/']
        >>> get_jit_context("unknown", "/project")
        []

    Notes:
        - Context Engineering: JIT Retrieval 원칙 준수
        - 필요한 문서만 로드하여 초기 컨텍스트 부담 최소화
        - 파일 존재 여부 확인 후 반환

    TDD History:
        - RED: 18개 시나리오 테스트 (커맨드 매칭, 키워드, 빈 결과)
        - GREEN: 패턴 매칭 딕셔너리 기반 구현
        - REFACTOR: 확장 가능한 패턴 구조, 파일 존재 검증 추가
    """
    context_files = []
    cwd_path = Path(cwd)

    # Pattern matching
    patterns = {
        "/alfred:1-spec": [".moai/memory/spec-metadata.md"],
        "/alfred:2-build": [".moai/memory/development-guide.md"],
        "test": ["tests/"],
    }

    for pattern, files in patterns.items():
        if pattern in prompt.lower():
            for file in files:
                file_path = cwd_path / file
                if file_path.exists():
                    context_files.append(file)

    return context_files


def detect_risky_operation(tool_name: str, tool_args: dict[str, Any], cwd: str) -> tuple[bool, str]:
    """위험한 작업 감지 (Event-Driven Checkpoint용)

    Claude Code tool 사용 전 위험한 작업을 자동으로 감지합니다.
    위험 감지 시 자동으로 checkpoint를 생성하여 롤백 가능하게 합니다.

    Args:
        tool_name: Claude Code tool 이름 (Bash, Edit, Write, MultiEdit)
        tool_args: Tool 인자 딕셔너리
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        (is_risky, operation_type) 튜플
        - is_risky: 위험한 작업 여부 (bool)
        - operation_type: 작업 유형 (str: delete, merge, script, critical-file, refactor)

    Risky Operations:
        - Bash tool: rm -rf, git merge, git reset --hard, git rebase
        - Edit/Write tool: CLAUDE.md, config.json, .moai/memory/*.md
        - MultiEdit tool: ≥10개 파일 동시 수정

    Examples:
        >>> detect_risky_operation("Bash", {"command": "rm -rf src/"}, ".")
        (True, 'delete')
        >>> detect_risky_operation("Edit", {"file_path": "CLAUDE.md"}, ".")
        (True, 'critical-file')
        >>> detect_risky_operation("Read", {"file_path": "test.py"}, ".")
        (False, '')

    Notes:
        - False Positive 최소화: 안전한 작업은 무시
        - 성능: 가벼운 문자열 매칭 (< 1ms)
        - 확장성: patterns 딕셔너리로 쉽게 추가 가능

    @TAG:CHECKPOINT-EVENT-001
    """
    # Bash tool: 위험한 명령어 감지
    if tool_name == "Bash":
        command = tool_args.get("command", "")

        # 대규모 삭제
        if any(pattern in command for pattern in ["rm -rf", "git rm"]):
            return (True, "delete")

        # Git 병합/리셋/리베이스
        if any(pattern in command for pattern in ["git merge", "git reset --hard", "git rebase"]):
            return (True, "merge")

        # 외부 스크립트 실행 (파괴적 가능성)
        if any(command.startswith(prefix) for prefix in ["python ", "node ", "bash ", "sh "]):
            return (True, "script")

    # Edit/Write tool: 중요 파일 감지
    if tool_name in ("Edit", "Write"):
        file_path = tool_args.get("file_path", "")

        critical_files = [
            "CLAUDE.md",
            "config.json",
            ".moai/memory/development-guide.md",
            ".moai/memory/spec-metadata.md",
            ".moai/config.json",
        ]

        if any(cf in file_path for cf in critical_files):
            return (True, "critical-file")

    # MultiEdit tool: 대규모 수정 감지
    if tool_name == "MultiEdit":
        edits = tool_args.get("edits", [])
        if len(edits) >= 10:
            return (True, "refactor")

    return (False, "")


def create_checkpoint(cwd: str, operation_type: str) -> str:
    """Checkpoint 생성 (Git local branch)

    위험한 작업 전 자동으로 checkpoint를 생성합니다.
    Git local branch로 생성하여 원격 저장소 오염을 방지합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로
        operation_type: 작업 유형 (delete, merge, script 등)

    Returns:
        checkpoint_branch: 생성된 브랜치명
        실패 시 "checkpoint-failed" 반환

    Branch Naming:
        before-{operation}-{YYYYMMDD-HHMMSS}
        예: before-delete-20251015-143000

    Examples:
        >>> create_checkpoint(".", "delete")
        'before-delete-20251015-143000'

    Notes:
        - Local branch만 생성 (원격 push 안 함)
        - Git 오류 시 fallback (무시하고 계속 진행)
        - Dirty working directory 체크 안 함 (커밋 안 된 변경사항 허용)
        - Checkpoint 로그 자동 기록 (.moai/checkpoints.log)

    @TAG:CHECKPOINT-EVENT-001
    """
    from datetime import datetime

    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    branch_name = f"before-{operation_type}-{timestamp}"

    try:
        # 현재 브랜치에서 새 local branch 생성 (체크아웃 안 함)
        result = subprocess.run(
            ["git", "branch", branch_name],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
            timeout=2,
        )

        # Checkpoint 로그 기록
        log_checkpoint(cwd, branch_name, operation_type)

        return branch_name

    except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
        # Git 오류 시 fallback (무시)
        return "checkpoint-failed"


def log_checkpoint(cwd: str, branch_name: str, operation_type: str) -> None:
    """Checkpoint 로그 기록 (.moai/checkpoints.log)

    Checkpoint 생성 이력을 JSON Lines 형식으로 기록합니다.
    SessionStart에서 이 로그를 읽어 checkpoint 목록을 표시합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로
        branch_name: 생성된 checkpoint 브랜치명
        operation_type: 작업 유형

    Log Format (JSON Lines):
        {"timestamp": "2025-10-15T14:30:00", "branch": "before-delete-...", "operation": "delete"}

    Examples:
        >>> log_checkpoint(".", "before-delete-20251015-143000", "delete")
        # .moai/checkpoints.log에 1줄 추가

    Notes:
        - 파일 없으면 자동 생성
        - append 모드로 기록 (기존 로그 보존)
        - 실패 시 무시 (critical하지 않음)

    @TAG:CHECKPOINT-EVENT-001
    """
    from datetime import datetime

    log_file = Path(cwd) / ".moai" / "checkpoints.log"

    try:
        log_file.parent.mkdir(parents=True, exist_ok=True)

        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "branch": branch_name,
            "operation": operation_type,
        }

        with log_file.open("a") as f:
            f.write(json.dumps(log_entry) + "\n")

    except (OSError, PermissionError):
        # 로그 실패는 무시 (critical하지 않음)
        pass


def list_checkpoints(cwd: str, max_count: int = 10) -> list[dict[str, str]]:
    """Checkpoint 목록 조회 (.moai/checkpoints.log 파싱)

    최근 생성된 checkpoint 목록을 반환합니다.
    SessionStart, /alfred:0-project restore 커맨드에서 사용합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로
        max_count: 반환할 최대 개수 (기본 10개)

    Returns:
        Checkpoint 목록 (최신순)
        [{"timestamp": "...", "branch": "...", "operation": "..."}, ...]

    Examples:
        >>> list_checkpoints(".")
        [
            {"timestamp": "2025-10-15T14:30:00", "branch": "before-delete-...", "operation": "delete"},
            {"timestamp": "2025-10-15T14:25:00", "branch": "before-merge-...", "operation": "merge"},
        ]

    Notes:
        - 로그 파일 없으면 빈 리스트 반환
        - JSON 파싱 실패한 줄은 무시
        - 최신 max_count개만 반환

    @TAG:CHECKPOINT-EVENT-001
    """
    log_file = Path(cwd) / ".moai" / "checkpoints.log"

    if not log_file.exists():
        return []

    checkpoints = []

    try:
        with log_file.open("r") as f:
            for line in f:
                try:
                    checkpoints.append(json.loads(line.strip()))
                except json.JSONDecodeError:
                    # 파싱 실패한 줄 무시
                    pass
    except (OSError, PermissionError):
        return []

    # 최근 max_count개만 반환 (최신순)
    return checkpoints[-max_count:]


# ============================================================================
# Hook Handlers
# ============================================================================


def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart 이벤트 핸들러 (Checkpoint 목록 포함)

    Claude Code 세션 시작 시 프로젝트 상태를 요약하여 표시합니다.
    언어, Git 상태, SPEC 진행도, Checkpoint 목록을 한눈에 확인할 수 있습니다.

    Args:
        payload: Claude Code 이벤트 페이로드 (cwd 키 필수)

    Returns:
        HookResult(message=프로젝트 상태 요약 메시지)

    Message Format:
        🚀 MoAI-ADK Session Started
           Language: {언어}
           Branch: {브랜치} ({커밋 해시})
           Changes: {변경 파일 수}
           SPEC Progress: {완료}/{전체} ({퍼센트}%)
           Checkpoints: {개수} available (최신 3개 표시)

    TDD History:
        - RED: 세션 시작 메시지 형식 테스트
        - GREEN: helper 함수 조합하여 상태 메시지 생성
        - REFACTOR: 메시지 포맷 개선, 가독성 향상, checkpoint 목록 추가

    @TAG:CHECKPOINT-EVENT-001
    """
    cwd = payload.get("cwd", ".")
    language = detect_language(cwd)
    git_info = get_git_info(cwd)
    specs = count_specs(cwd)
    checkpoints = list_checkpoints(cwd, max_count=10)

    branch = git_info.get("branch", "N/A")
    commit = git_info.get("commit", "N/A")[:7]
    changes = git_info.get("changes", 0)
    spec_progress = f"{specs['completed']}/{specs['total']}"

    message = (
        f"🚀 MoAI-ADK Session Started\n"
        f"   Language: {language}\n"
        f"   Branch: {branch} ({commit})\n"
        f"   Changes: {changes}\n"
        f"   SPEC Progress: {spec_progress} ({specs['percentage']}%)"
    )

    # Checkpoint 목록 추가 (최신 3개만 표시)
    if checkpoints:
        message += f"\n   Checkpoints: {len(checkpoints)} available"
        for cp in reversed(checkpoints[-3:]):  # 최신 3개
            branch_short = cp["branch"].replace("before-", "")
            message += f"\n      - {branch_short}"
        message += "\n   Restore: /alfred:0-project restore"

    return HookResult(message=message)


def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """UserPromptSubmit 이벤트 핸들러

    사용자 프롬프트를 분석하여 관련 문서를 자동으로 컨텍스트에 추가합니다.
    JIT (Just-in-Time) Retrieval 원칙에 따라 필요한 문서만 로드합니다.

    Args:
        payload: Claude Code 이벤트 페이로드
                 (userPrompt, cwd 키 포함)

    Returns:
        HookResult(
            message=로드된 파일 수 (또는 None),
            contextFiles=추천 문서 경로 리스트
        )

    TDD History:
        - RED: JIT 문서 로딩 시나리오 테스트
        - GREEN: get_jit_context() 호출하여 문서 추천
        - REFACTOR: 메시지 조건부 표시 (파일 있을 때만)
    """
    user_prompt = payload.get("userPrompt", "")
    cwd = payload.get("cwd", ".")
    context_files = get_jit_context(user_prompt, cwd)

    message = f"📎 Loaded {len(context_files)} context file(s)" if context_files else None

    return HookResult(message=message, contextFiles=context_files)


def handle_pre_compact(payload: HookPayload) -> HookResult:
    """PreCompact 이벤트 핸들러

    컨텍스트가 70% 이상 차면 새 세션 시작을 제안합니다.
    Context Engineering의 Compaction 원칙을 구현합니다.

    Args:
        payload: Claude Code 이벤트 페이로드

    Returns:
        HookResult(
            message=새 세션 시작 제안 메시지,
            suggestions=구체적인 액션 제안 리스트
        )

    Suggestions:
        - /clear 명령으로 새 세션 시작
        - /new 명령으로 새 대화 시작
        - 핵심 결정사항 요약 후 계속

    Notes:
        - Context Engineering: Compaction 원칙 준수
        - 토큰 사용량 > 70% 시 자동 호출
        - 성능 향상 및 컨텍스트 관리 개선

    TDD History:
        - RED: PreCompact 메시지 및 제안 테스트
        - GREEN: 고정 메시지 및 제안 리스트 반환
        - REFACTOR: 사용자 친화적 메시지 개선
    """
    suggestions = [
        "Use `/clear` to start a fresh session",
        "Use `/new` to begin a new conversation",
        "Consider summarizing key decisions before continuing",
    ]

    message = (
        "💡 Tip: Context is getting large. Consider starting a new session for better performance."
    )

    return HookResult(message=message, suggestions=suggestions)


def handle_session_end(payload: HookPayload) -> HookResult:
    """SessionEnd 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse 이벤트 핸들러 (Event-Driven Checkpoint 통합)

    위험한 작업 전 자동으로 checkpoint를 생성합니다.
    Claude Code tool 사용 전에 호출되며, 위험 감지 시 사용자에게 알립니다.

    Args:
        payload: Claude Code 이벤트 페이로드
                 (tool, arguments, cwd 키 포함)

    Returns:
        HookResult(
            message=checkpoint 생성 알림 (위험 감지 시),
            blocked=False (항상 작업 계속 진행)
        )

    Checkpoint Triggers:
        - Bash: rm -rf, git merge, git reset --hard
        - Edit/Write: CLAUDE.md, config.json
        - MultiEdit: ≥10 files

    Examples:
        Bash tool (rm -rf) 감지:
        → "🛡️ Checkpoint created: before-delete-20251015-143000"

    Notes:
        - 위험 감지 후에도 blocked=False 반환 (작업 계속)
        - Checkpoint 실패 시에도 작업 진행 (무시)
        - 투명한 백그라운드 동작

    @TAG:CHECKPOINT-EVENT-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # 위험한 작업 감지
    is_risky, operation_type = detect_risky_operation(tool_name, tool_args, cwd)

    # 위험 감지 시 checkpoint 생성
    if is_risky:
        checkpoint_branch = create_checkpoint(cwd, operation_type)

        if checkpoint_branch != "checkpoint-failed":
            message = (
                f"🛡️ Checkpoint created: {checkpoint_branch}\n"
                f"   Operation: {operation_type}\n"
                f"   Restore: /alfred:0-project restore"
            )

            return HookResult(message=message, blocked=False)

    return HookResult(blocked=False)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_notification(payload: HookPayload) -> HookResult:
    """Notification 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_subagent_stop(payload: HookPayload) -> HookResult:
    """SubagentStop 이벤트 핸들러 (기본 구현)"""
    return HookResult()


# ============================================================================
# Main Entry Point
# ============================================================================


def main() -> None:
    """메인 진입점 - Claude Code Hook 스크립트

    CLI 인수로 이벤트명을 받고, stdin으로 JSON 페이로드를 읽습니다.
    이벤트에 맞는 핸들러를 호출하고, 결과를 JSON으로 stdout에 출력합니다.

    Usage:
        python moai_hooks.py <event_name> < payload.json

    Supported Events:
        - SessionStart: 세션 시작 (프로젝트 상태 표시)
        - UserPromptSubmit: 프롬프트 제출 (JIT 문서 로딩)
        - PreCompact: 컨텍스트 초과 경고 (새 세션 제안)
        - SessionEnd, PreToolUse, PostToolUse, Notification, Stop, SubagentStop

    Exit Codes:
        - 0: 성공
        - 1: 에러 (인수 없음, JSON 파싱 실패, 예외 발생)

    Examples:
        $ echo '{"cwd": "."}' | python moai_hooks.py SessionStart
        {"message": "🚀 MoAI-ADK Session Started\\n...", ...}

    Notes:
        - Claude Code가 자동으로 호출 (사용자 직접 실행 불필요)
        - stdin/stdout으로 JSON I/O 처리
        - stderr로 에러 메시지 출력

    TDD History:
        - RED: 이벤트 라우팅, JSON I/O, 에러 처리 테스트
        - GREEN: 핸들러 맵 기반 라우팅 구현
        - REFACTOR: 에러 메시지 명확화, exit code 표준화
    """
    # Check for event argument
    if len(sys.argv) < 2:
        print("Usage: moai_hooks.py <event>", file=sys.stderr)
        sys.exit(1)

    event_name = sys.argv[1]

    try:
        # Read JSON from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data)

        cwd = data.get("cwd", ".")

        # Route to appropriate handler
        handlers = {
            "SessionStart": handle_session_start,
            "UserPromptSubmit": handle_user_prompt_submit,
            "PreCompact": handle_pre_compact,
            "SessionEnd": handle_session_end,
            "PreToolUse": handle_pre_tool_use,
            "PostToolUse": handle_post_tool_use,
            "Notification": handle_notification,
            "Stop": handle_stop,
            "SubagentStop": handle_subagent_stop,
        }

        handler = handlers.get(event_name)
        result = handler({"cwd": cwd, **data}) if handler else HookResult()

        # Output JSON to stdout
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
