#!/usr/bin/env python3
"""
Auto Checkpoint System for MoAI-ADK Personal Mode v0.2.0

Git 히스토리를 오염시키지 않는 스냅샷 기반 체크포인트를 일정 주기로 생성한다.

@REQ:AUTO-CHECKPOINT-001
@FEATURE:AUTO-BACKUP-001
@API:CHECKPOINT-AUTOMATION-001
@DESIGN:PERSONAL-MODE-ONLY-001
@TECH:TIME-BASED-CHECKPOINT-002
"""

from __future__ import annotations

import json
import os
import sys
import time
from pathlib import Path
from typing import Dict

# 프로젝트 스크립트 접근 경로 추가
ROOT_DIR = Path(__file__).resolve().parents[2]
SCRIPTS_DIR = ROOT_DIR / ".moai" / "scripts"
sys.path.insert(0, str(SCRIPTS_DIR))

from checkpoint_manager import CheckpointManager  # noqa: E402


class AutoCheckpointManager:
    """개인 모드에서 자동 체크포인트를 생성하는 관리자."""

    def __init__(self, project_root: Path) -> None:
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"
        self.config_file = self.moai_dir / "config.json"
        self.checkpoints_dir = self.moai_dir / "checkpoints"
        self.last_checkpoint_file = self.checkpoints_dir / ".last_checkpoint"
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)
        self.manager = CheckpointManager()
        tmp_dir = self.checkpoints_dir / "tmp"
        tmp_dir.mkdir(parents=True, exist_ok=True)
        self.git_env = os.environ.copy()
        self.git_env.setdefault("TMPDIR", str(tmp_dir))

    # ------------------------------------------------------------------
    # Configuration helpers
    # ------------------------------------------------------------------
    def load_config(self) -> Dict[str, any]:
        try:
            with open(self.config_file, "r", encoding="utf-8") as fh:
                return json.load(fh)
        except FileNotFoundError:
            return {}

    def is_personal_mode(self) -> bool:
        return self.load_config().get("project", {}).get("mode") == "personal"

    def is_auto_checkpoint_enabled(self) -> bool:
        return (
            self.load_config()
            .get("git_strategy", {})
            .get("personal", {})
            .get("auto_checkpoint", False)
        )

    def get_checkpoint_interval(self) -> int:
        return (
            self.load_config()
            .get("git_strategy", {})
            .get("personal", {})
            .get("checkpoint_interval", 300)
        )

    # ------------------------------------------------------------------
    # Git state helpers
    # ------------------------------------------------------------------
    def _run_git(self, args: list[str]) -> str:
        from subprocess import run

        result = run(
            args,
            cwd=self.project_root,
            capture_output=True,
            text=True,
            env=self.git_env,
        )
        if result.returncode != 0:
            return ""
        return result.stdout.strip()

    def has_uncommitted_changes(self) -> bool:
        return bool(self._run_git(["git", "status", "--porcelain"]))

    def is_git_repository(self) -> bool:
        return bool(self._run_git(["git", "rev-parse", "--is-inside-work-tree"]))

    def time_since_last_checkpoint(self) -> float:
        if not self.last_checkpoint_file.exists():
            return float("inf")
        try:
            return time.time() - float(self.last_checkpoint_file.read_text())
        except ValueError:
            return float("inf")

    # ------------------------------------------------------------------
    def should_create_checkpoint(self) -> bool:
        if not self.is_personal_mode() or not self.is_auto_checkpoint_enabled():
            return False
        if not self.is_git_repository():
            return False
        if not self.has_uncommitted_changes():
            return False
        return self.time_since_last_checkpoint() >= self.get_checkpoint_interval()

    def update_last_checkpoint_time(self) -> None:
        self.last_checkpoint_file.write_text(str(time.time()))

    def create_checkpoint(self) -> bool:
        success = self.manager.create_checkpoint("Auto checkpoint", source="auto", quiet=True)
        if success:
            self.update_last_checkpoint_time()
            print("💾 자동 체크포인트 생성 완료")
        return success

    def cleanup_old_checkpoints(self) -> None:
        self.manager.cleanup_old_checkpoints()

    def run_once(self) -> bool:
        if self.should_create_checkpoint():
            return self.create_checkpoint()
        return False

    def run_daemon(self, interval: int = 60) -> None:
        print(f"🔄 Auto-checkpoint daemon started (interval={interval}s)")
        try:
            while True:
                if self.should_create_checkpoint():
                    self.create_checkpoint()
                if int(time.time()) % 3600 == 0:
                    self.cleanup_old_checkpoints()
                time.sleep(interval)
        except KeyboardInterrupt:
            print("\n⏹️ Auto-checkpoint daemon stopped")


def main() -> None:
    if len(sys.argv) < 2:
        print("Usage: auto_checkpoint.py <project_root> [--daemon] [--once]")
        sys.exit(1)

    project_root = Path(sys.argv[1]).resolve()
    if not project_root.exists():
        print(f"❌ Project root does not exist: {project_root}")
        sys.exit(1)

    manager = AutoCheckpointManager(project_root)

    if "--daemon" in sys.argv:
        manager.run_daemon()
    elif "--once" in sys.argv:
        manager.run_once()
    else:
        manager.run_once()


if __name__ == "__main__":
    main()
