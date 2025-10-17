#!/usr/bin/env python3
"""Event-Driven Checkpoint system

위험한 작업 감지 및 자동 Checkpoint 생성
@TAG:CHECKPOINT-EVENT-001
"""

import json
import re
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any

# MoAI-ADK 지원 언어별 스크립트 실행 패턴
# Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin + Shell
SCRIPT_EXECUTION_PATTERN = re.compile(
    r"\b("
    # Python ecosystem
    r"python3?|pytest|pip|uv|"
    # JavaScript/TypeScript ecosystem
    r"node|npm|npx|yarn|bun|tsx|ts-node|vitest|jest|"
    # Java ecosystem
    r"java|javac|mvn|gradle|"
    # Go
    r"go|"
    # Rust
    r"cargo|"
    # Dart/Flutter
    r"dart|flutter|"
    # Swift
    r"swift|xcodebuild|"
    # Kotlin
    r"kotlinc?|"
    # Shell scripts and build tools
    r"bash|sh|zsh|fish|make"
    r")\b"
)


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
        - Bash tool: rm -rf, git merge, git reset --hard, git rebase, script execution
        - Edit/Write tool: CLAUDE.md, config.json, .moai/memory/*.md
        - MultiEdit tool: ≥10개 파일 동시 수정
        - Script execution: Python, Node, Java, Go, Rust, Dart, Swift, Kotlin, Shell scripts

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


__all__ = [
    "detect_risky_operation",
    "create_checkpoint",
    "log_checkpoint",
    "list_checkpoints",
]
