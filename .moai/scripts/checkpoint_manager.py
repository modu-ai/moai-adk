#!/usr/bin/env python3
"""
MoAI 체크포인트 관리자 v0.3.0
개인 모드 전용 안전한 실험 환경 제공 – Annotated Tag 기반 스냅샷 방식

@REQ:GIT-CHECKPOINT-001
@FEATURE:CHECKPOINT-SYSTEM-001
@API:CHECKPOINT-INTERFACE-001
@DESIGN:CHECKPOINT-WORKFLOW-003
@TECH:PERSONAL-MODE-001
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tarfile
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, List, Optional, cast


class CheckpointManager:
    """체크포인트 관리자.

    Annotated Tag(`moai_cp/*`)를 활용해 작업 히스토리를 보존하면서
    메타데이터(.moai/checkpoints/metadata.json)를 유지한다.
    git stash는 임시 스냅샷 생성 시점에서만 사용하며 즉시 Tag로 승격한다.
    """

    TAG_PREFIX = "moai_cp/"

    def __init__(self) -> None:
        self.project_root = Path(__file__).resolve().parents[2]
        self.config_path = self.project_root / ".moai" / "config.json"
        self.checkpoints_dir = self.project_root / ".moai" / "checkpoints"
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)
        self.metadata_path = self.checkpoints_dir / "metadata.json"
        tmp_dir = self.checkpoints_dir / "tmp"
        tmp_dir.mkdir(parents=True, exist_ok=True)
        self.git_env = os.environ.copy()
        self.git_env.setdefault("TMPDIR", str(tmp_dir))

    # ------------------------------------------------------------------
    # Git helpers
    # ------------------------------------------------------------------
    def _run_git(self, args: List[str], *, check: bool = True) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            args,
            cwd=self.project_root,
            capture_output=True,
            text=True,
            check=check,
            env=self.git_env,
        )

    def _git_output(self, args: List[str]) -> str:
        return self._run_git(args).stdout.strip()

    # ------------------------------------------------------------------
    # Configuration / metadata helpers
    # ------------------------------------------------------------------
    def load_config(self) -> Dict[str, Any]:
        try:
            with open(self.config_path, "r", encoding="utf-8") as fh:
                return cast(Dict[str, Any], json.load(fh))
        except FileNotFoundError:
            return {"project": {"mode": "personal"}, "git_strategy": {"personal": {}}}

    def _load_metadata(self) -> Dict[str, Any]:
        if not self.metadata_path.exists():
            return {"checkpoints": []}
        with open(self.metadata_path, "r", encoding="utf-8") as fh:
            try:
                return cast(Dict[str, Any], json.load(fh))
            except json.JSONDecodeError:
                return {"checkpoints": []}

    def _save_metadata(self, data: Dict[str, Any]) -> None:
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)
        with open(self.metadata_path, "w", encoding="utf-8") as fh:
            json.dump(data, fh, ensure_ascii=False, indent=2)

    # ------------------------------------------------------------------
    # Stash utilities (임시 스냅샷 생성 용도)
    # ------------------------------------------------------------------
    def _list_stash_entries(self) -> List[Dict[str, str]]:
        result = self._run_git([
            "git",
            "stash",
            "list",
            "--format=%H %gd %gs",
        ], check=False)
        entries: List[Dict[str, str]] = []
        for line in result.stdout.strip().splitlines():
            try:
                commit, ref, message = line.split(" ", 2)
            except ValueError:
                continue
            entries.append({"commit": commit, "ref": ref, "message": message})
        return entries

    def _find_stash_by_commit(self, commit_hash: str) -> Optional[Dict[str, str]]:
        for entry in self._list_stash_entries():
            if entry["commit"] == commit_hash:
                return entry
        return None

    def _find_stash_by_marker(self, marker: str) -> Optional[Dict[str, str]]:
        for entry in self._list_stash_entries():
            if marker in entry["message"]:
                return entry
        return None

    def _drop_stash_entry(self, commit_hash: str) -> None:
        entry = self._find_stash_by_commit(commit_hash)
        if entry:
            self._run_git(["git", "stash", "drop", entry["ref"]], check=False)

    def _files_changed_for_stash(self, ref: str) -> int:
        result = self._run_git(["git", "stash", "show", ref, "--stat"], check=False)
        for line in result.stdout.splitlines()[::-1]:
            if "files changed" in line:
                parts = line.strip().split()
                try:
                    return int(parts[0])
                except (IndexError, ValueError):
                    return 0
        return 0

    # ------------------------------------------------------------------
    # Filesystem fallback utilities
    # ------------------------------------------------------------------
    def _should_skip_path(self, path: Path) -> bool:
        try:
            relative = path.relative_to(self.project_root)
        except ValueError:
            return True
        parts = relative.parts
        if not parts:
            return True
        if parts[0] == ".git":
            return True
        if len(parts) >= 2 and parts[0] == ".moai" and parts[1] == "checkpoints":
            if len(parts) >= 3 and parts[2] in {"snapshots", "tmp"}:
                return True
        return False

    def _create_filesystem_snapshot(self, checkpoint_id: str) -> Optional[str]:
        snapshots_dir = self.checkpoints_dir / "snapshots"
        snapshots_dir.mkdir(parents=True, exist_ok=True)
        archive_path = snapshots_dir / f"{checkpoint_id}.tar.gz"
        try:
            with tarfile.open(archive_path, "w:gz") as tar:
                for path in self.project_root.rglob("*"):
                    if self._should_skip_path(path):
                        continue
                    tar.add(path, arcname=str(path.relative_to(self.project_root)))
            return str(archive_path.relative_to(self.project_root))
        except Exception:
            archive_path.unlink(missing_ok=True)
            return None

    # ------------------------------------------------------------------
    # Core behaviour
    # ------------------------------------------------------------------
    def _current_time(self) -> datetime:
        return datetime.now()

    def check_personal_mode(self, *, quiet: bool = False) -> bool:
        mode = self.load_config().get("project", {}).get("mode", "personal")
        if mode != "personal":
            if not quiet:
                print("⚠️ 체크포인트는 개인 모드에서만 지원됩니다.")
                print(f"현재 모드: {mode}")
                print("개인 모드로 전환: .moai/config.json에서 mode를 'personal'로 변경")
            return False
        if not quiet:
            print("✅ 개인 모드 확인 완료")
        return True

    def _has_uncommitted_changes(self) -> bool:
        status = self._git_output(["git", "status", "--porcelain"])
        return bool(status)

    def _tag_exists(self, tag_name: str) -> bool:
        result = self._run_git(["git", "tag", "-l", tag_name], check=False)
        return tag_name in result.stdout.splitlines()

    def _create_tag(self, tag_name: str, commit: str, message: str) -> bool:
        result = self._run_git([
            "git",
            "tag",
            "-a",
            tag_name,
            commit,
            "-m",
            message,
        ], check=False)
        return result.returncode == 0

    def _delete_tag(self, tag_name: str) -> None:
        if tag_name:
            self._run_git(["git", "tag", "-d", tag_name], check=False)

    def generate_checkpoint_id(self) -> str:
        base = self._current_time().strftime("%Y%m%d_%H%M%S")
        candidate = f"{self.TAG_PREFIX}{base}"
        suffix = 1
        while self._tag_exists(candidate):
            candidate = f"{self.TAG_PREFIX}{base}_{suffix:02d}"
            suffix += 1
        return candidate

    def _delete_checkpoint_resource(self, entry: Dict[str, Any]) -> None:
        kind = entry.get("kind", "tag")
        if kind == "tag":
            tag_name = entry.get("tag") or entry.get("id")
            if tag_name:
                self._delete_tag(str(tag_name))
        elif kind == "filesystem":
            snapshot_rel = entry.get("snapshot")
            if snapshot_rel:
                (self.project_root / str(snapshot_rel)).unlink(missing_ok=True)
        elif kind == "stash":
            commit = entry.get("stash_commit")
            if commit:
                self._drop_stash_entry(str(commit))
        elif kind == "legacy":
            branch = entry.get("branch")
            if branch:
                self._run_git(["git", "branch", "-D", branch], check=False)

    def _record_checkpoint(self, entry: Dict[str, Any]) -> None:
        metadata = self._load_metadata()
        metadata.setdefault("checkpoints", []).append(entry)

        config = self.load_config()
        max_entries = config.get("git_strategy", {}).get("personal", {}).get("max_checkpoints", 50)
        metadata["checkpoints"].sort(key=lambda cp: cp.get("timestamp", ""))
        extras = max(0, len(metadata["checkpoints"]) - max_entries)
        for _ in range(extras):
            removed = metadata["checkpoints"].pop(0)
            self._delete_checkpoint_resource(removed)
        self._save_metadata(metadata)

    def list_checkpoints(self) -> None:
        metadata = self._load_metadata()
        checkpoints = sorted(metadata.get("checkpoints", []), key=lambda cp: cp.get("timestamp", ""), reverse=True)
        if not checkpoints:
            print("📋 체크포인트가 없습니다.")
            return
        print("📋 사용 가능한 체크포인트 (최신 10개):")
        for entry in checkpoints[:10]:
            timestamp = entry.get("timestamp", "-")
            message = entry.get("message", "")
            origin = entry.get("source", "manual")
            files = entry.get("files_changed", 0)
            kind = entry.get("kind", "tag")
            print(f"  📍 {entry.get('id', 'unknown')} | {timestamp} | {origin}/{kind} | 파일 {files}개 | {message}")

    def show_status(self) -> None:
        metadata = self._load_metadata()
        checkpoints = metadata.get("checkpoints", [])
        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        print("=== 체크포인트 시스템 상태 ===")
        print(f"🎯 모드: {mode}")
        print(f"💾 총 체크포인트: {len(checkpoints)}개")
        if checkpoints:
            latest = max(checkpoints, key=lambda cp: cp.get("timestamp", ""))
            print(f"🆔 최신 ID: {latest.get('id')}")
            print(f"📅 생성 시간: {latest.get('timestamp')}")
            print(f"📝 메시지: {latest.get('message', '')}")

    def cleanup_old_checkpoints(self) -> None:
        metadata = self._load_metadata()
        checkpoints = metadata.get("checkpoints", [])
        if not checkpoints:
            print("정리할 체크포인트가 없습니다.")
            return

        current = self._current_time()
        remaining: List[Dict[str, Any]] = []
        removed = 0
        for entry in checkpoints:
            try:
                timestamp = datetime.fromisoformat(entry.get("timestamp", ""))
            except ValueError:
                timestamp = None
            if timestamp and (current - timestamp) > timedelta(days=7):
                self._delete_checkpoint_resource(entry)
                removed += 1
            else:
                remaining.append(entry)
        metadata["checkpoints"] = remaining
        self._save_metadata(metadata)
        if removed:
            print(f"✅ {removed}개 체크포인트 정리 완료")
        else:
            print("정리할 오래된 체크포인트가 없습니다.")

    # ------------------------------------------------------------------
    # Internal storage helpers
    # ------------------------------------------------------------------
    def _capture_stash_entry(
        self,
        checkpoint_id: str,
        clean_message: str,
        *,
        quiet: bool,
    ) -> tuple[str, Optional[Dict[str, str]]]:
        label = f"{checkpoint_id} :: {clean_message}"
        result = self._run_git([
            "git",
            "stash",
            "push",
            "--include-untracked",
            "-m",
            label,
        ], check=False)
        if "No local changes" in result.stdout:
            if not quiet:
                print("ℹ️ 변경사항이 없어 체크포인트를 건너뜁니다.")
            return "skip", None
        if result.returncode != 0:
            if not quiet:
                error = result.stderr.strip() or "git stash push 실패"
                print(f"⚠️ Git 스냅샷 생성 실패({error}) – 파일 시스템 스냅샷으로 전환합니다.")
            return "error", None
        entry = self._find_stash_by_marker(checkpoint_id)
        if not entry:
            if not quiet:
                print("❌ 생성된 스냅샷을 찾을 수 없습니다. 최신 stash를 복구합니다.")
            self._run_git(["git", "stash", "pop"], check=False)
            return "missing", None
        return "success", entry

    def _store_filesystem_checkpoint(
        self,
        checkpoint_id: str,
        clean_message: str,
        *,
        source: str,
        branch: str,
        base_commit: str,
        quiet: bool,
    ) -> bool:
        snapshot_rel = self._create_filesystem_snapshot(checkpoint_id)
        if not snapshot_rel:
            if not quiet:
                print("❌ 체크포인트 생성 실패: 파일 시스템 스냅샷 생성 불가")
            return False
        metadata_entry = {
            "id": checkpoint_id,
            "timestamp": self._current_time().isoformat(),
            "message": clean_message,
            "source": source,
            "kind": "filesystem",
            "snapshot": snapshot_rel,
            "files_changed": 0,
            "mode": "personal",
            "branch": branch,
            "base_commit": base_commit,
        }
        self._record_checkpoint(metadata_entry)
        if not quiet:
            print("=== 체크포인트 생성 결과 ===")
            print(f"🆔 ID: {checkpoint_id}")
            print(f"💾 스냅샷: {snapshot_rel}")
        return True

    def _store_tag_checkpoint(
        self,
        checkpoint_id: str,
        clean_message: str,
        stash_entry: Dict[str, str],
        *,
        source: str,
        branch: str,
        base_commit: str,
        quiet: bool,
    ) -> bool:
        commit_hash = stash_entry["commit"]
        files_changed = self._files_changed_for_stash(stash_entry["ref"])
        if not self._create_tag(checkpoint_id, commit_hash, clean_message):
            if not quiet:
                print("⚠️ Tag 생성에 실패하여 파일 시스템 스냅샷으로 대체합니다.")
            self._run_git(["git", "stash", "apply", stash_entry["ref"]], check=False)
            self._drop_stash_entry(commit_hash)
            return self._store_filesystem_checkpoint(
                checkpoint_id,
                clean_message,
                source=source,
                branch=branch,
                base_commit=base_commit,
                quiet=quiet,
            )
        apply_result = self._run_git(["git", "stash", "apply", stash_entry["ref"]], check=False)
        if apply_result.returncode != 0:
            if not quiet:
                print(f"❌ 작업 복구 실패: {apply_result.stderr.strip()}")
            self._delete_tag(checkpoint_id)
            self._drop_stash_entry(commit_hash)
            return False
        self._run_git(["git", "stash", "drop", stash_entry["ref"]], check=False)
        metadata_entry = {
            "id": checkpoint_id,
            "timestamp": self._current_time().isoformat(),
            "message": clean_message,
            "source": source,
            "kind": "tag",
            "tag": checkpoint_id,
            "stash_commit": commit_hash,
            "files_changed": files_changed,
            "mode": "personal",
            "branch": branch,
            "base_commit": base_commit,
        }
        self._record_checkpoint(metadata_entry)
        if not quiet:
            print("=== 체크포인트 생성 결과 ===")
            print(f"🆔 ID(Tag): {checkpoint_id}")
            print(f"🗂️ 변경 파일: {files_changed}개")
        return True

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------
    def create_checkpoint(self, message: str = "", *, source: str = "manual", quiet: bool = False) -> bool:
        if not self.check_personal_mode(quiet=quiet):
            return False
        if not self._has_uncommitted_changes():
            if not quiet:
                print("ℹ️ 저장할 변경사항이 없습니다.")
            return False

        checkpoint_id = self.generate_checkpoint_id()
        clean_message = message.strip() or f"Snapshot {self._current_time():%Y-%m-%d %H:%M:%S}"
        current_branch = self._git_output(["git", "branch", "--show-current"]) or "unknown"
        base_commit = self._git_output(["git", "rev-parse", "HEAD"]) or ""

        if not quiet:
            print(f"💾 체크포인트 생성: {clean_message}")

        status, stash_entry = self._capture_stash_entry(
            checkpoint_id,
            clean_message,
            quiet=quiet,
        )

        if status == "skip":
            return False
        if status == "error":
            return self._store_filesystem_checkpoint(
                checkpoint_id,
                clean_message,
                source=source,
                branch=current_branch,
                base_commit=base_commit,
                quiet=quiet,
            )
        if status != "success" or not stash_entry:
            return False

        return self._store_tag_checkpoint(
            checkpoint_id,
            clean_message,
            stash_entry,
            source=source,
            branch=current_branch,
            base_commit=base_commit,
            quiet=quiet,
        )

    def run(self, args: List[str]) -> None:
        if not args:
            self.create_checkpoint()
            return
        action = args[0]
        if action == "--list":
            self.list_checkpoints()
        elif action == "--status":
            self.show_status()
        elif action == "--cleanup":
            self.cleanup_old_checkpoints()
        else:
            message = " ".join(args)
            self.create_checkpoint(message)


def main() -> None:
    manager = CheckpointManager()
    manager.run(sys.argv[1:])


if __name__ == "__main__":
    main()
