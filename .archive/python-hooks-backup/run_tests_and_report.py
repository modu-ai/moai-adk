#!/usr/bin/env python3
# @CODE:TEST-REPORT-011
"""Stop hook that runs lightweight tests and reports the outcome."""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

Command = tuple[str, list[str]]
TIMEOUT_SECONDS = 300


def _load_input() -> dict:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError:
        return {}


def _project_root() -> Path:
    env = os.environ.get("CLAUDE_PROJECT_DIR")
    return Path(env).resolve() if env else Path.cwd().resolve()


def _command_exists(executable: str) -> bool:
    return shutil.which(executable) is not None


def _detect_pytest(root: Path) -> Command | None:
    tests_dir = root / "tests"
    if tests_dir.exists() and _command_exists("pytest"):
        return ("pytest", [sys.executable or "python3", "-m", "pytest", "-q"])
    return None


def _detect_npm(root: Path) -> Command | None:
    if (root / "package.json").exists() and _command_exists("npm"):
        return ("npm test", ["npm", "test", "--", "--watch=false"])
    return None


def _detect_go(root: Path) -> Command | None:
    if (root / "go.mod").exists() and _command_exists("go"):
        return ("go test", ["go", "test", "./..."])
    return None


def _detect_cargo(root: Path) -> Command | None:
    if (root / "Cargo.toml").exists() and _command_exists("cargo"):
        return ("cargo test", ["cargo", "test", "--quiet"])
    return None


def _collect_commands(root: Path) -> list[Command]:
    commands: list[Command] = []
    for detector in (_detect_pytest, _detect_npm, _detect_go, _detect_cargo):
        match = detector(root)
        if match:
            commands.append(match)
    return commands


def _run_command(command: Command, root: Path) -> tuple[int, str, str]:
    name, argv = command
    try:
        proc = subprocess.run(
            argv,
            cwd=root,
            capture_output=True,
            text=True,
            timeout=TIMEOUT_SECONDS,
        )
        return proc.returncode, proc.stdout.strip(), proc.stderr.strip()
    except subprocess.TimeoutExpired:
        return 124, "", f"{name} timed out after {TIMEOUT_SECONDS}s"
    except Exception as exc:  # pragma: no cover - defensive
        return 1, "", f"{name} failed: {exc}"


def main() -> None:
    print(
        "Stop Hook: 비활성화됨 - 테스트는 /moai:2-build 단계에서만 실행됩니다.",
        flush=True,
    )
    sys.exit(0)


if __name__ == "__main__":
    main()
