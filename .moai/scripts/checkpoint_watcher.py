#!/usr/bin/env python3
"""Convenience wrapper for the auto checkpoint file watcher."""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parents[2]
FILE_WATCHER = SCRIPT_DIR / ".claude" / "hooks" / "moai" / "file_watcher.py"


def run(args: list[str], cwd: Path) -> int:
    try:
        result = subprocess.run(
            [sys.executable or "python3", str(FILE_WATCHER)] + args,
            cwd=cwd,
            check=False,
            text=True,
        )
        return result.returncode
    except FileNotFoundError:
        print("❌ file_watcher.py 를 찾을 수 없습니다. (.claude/hooks/moai/file_watcher.py)")
        return 1


def main() -> None:
    parser = argparse.ArgumentParser(description="MoAI 자동 체크포인트 워처 컨트롤러")
    parser.add_argument("action", choices=["start", "stop", "status", "once"], help="실행/중지/상태/단발성")
    parser.add_argument("path", nargs="?", help="대상 프로젝트 경로 (기본값: 현재)" )
    args = parser.parse_args()

    project_root = Path(args.path).resolve() if args.path else SCRIPT_DIR
    if not FILE_WATCHER.exists():
        print(f"❌ file_watcher.py 를 찾을 수 없습니다: {FILE_WATCHER}")
        sys.exit(1)

    if args.action == "start":
        exit_code = run([str(project_root), "--start"], project_root)
    elif args.action == "stop":
        exit_code = run([str(project_root), "--stop"], project_root)
    elif args.action == "status":
        exit_code = run([str(project_root), "--status"], project_root)
    else:  # once
        exit_code = run([str(project_root), "--once"], project_root)

    sys.exit(exit_code)


if __name__ == "__main__":
    main()
