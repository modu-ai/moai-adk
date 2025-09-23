#!/usr/bin/env python3
"""
MoAI-ADK Steering filename migration tool

Migrates legacy Steering filenames to the standard ones:
  - vision.md       -> product.md
  - architecture.md -> structure.md
  - techstack.md    -> tech.md

Safety:
  - Dry-run by default (prints plan without modifying files)
  - Creates a timestamped backup directory before changes
  - Skips if targets already exist (unless --force)

Usage:
  python scripts/migrate_steering_filenames.py [--apply] [--force]

Options:
  --apply   Apply changes (default: dry-run)
  --force   Overwrite existing target files (after backing them up)
"""
from __future__ import annotations

import argparse
import shutil
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple


LEGACY_TO_STANDARD: Dict[str, str] = {
    "vision.md": "product.md",
    "architecture.md": "structure.md",
    "techstack.md": "tech.md",
}


def find_migrations(steering_dir: Path) -> List[Tuple[Path, Path]]:
    plan: List[Tuple[Path, Path]] = []
    for legacy, standard in LEGACY_TO_STANDARD.items():
        src = steering_dir / legacy
        dst = steering_dir / standard
        if src.exists():
            plan.append((src, dst))
    return plan


def ensure_backup_dir(project_root: Path) -> Path:
    ts = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup = project_root / f".moai_backup_{ts}"
    backup.mkdir(parents=True, exist_ok=True)
    return backup


def backup_file(src: Path, backup_root: Path) -> Path:
    rel = src.relative_to(Path.cwd()) if src.is_absolute() else src
    dest = backup_root / rel
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest)
    return dest


def migrate(steering_dir: Path, apply: bool, force: bool) -> int:
    project_root = Path.cwd()
    plan = find_migrations(steering_dir)
    if not plan:
        print("✅ No legacy Steering filenames found. Nothing to migrate.")
        return 0

    print("🔎 Migration plan (legacy -> standard):")
    for src, dst in plan:
        status = "overwrite" if dst.exists() else "create"
        print(f" - {src} -> {dst} ({status})")

    if not apply:
        print("\nℹ️ Dry-run only. Re-run with --apply to perform changes.")
        return 0

    backup_root = ensure_backup_dir(project_root)
    print(f"\n🗄️  Backup directory: {backup_root}")

    moved = 0
    for src, dst in plan:
        # backup source
        backup_file(src, backup_root)
        # backup existing target if any
        if dst.exists():
            if not force:
                print(f"⏭️  Skip: target exists (use --force to overwrite): {dst}")
                continue
            backup_file(dst, backup_root)
        # perform move
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(src), str(dst))
        print(f"✅ Moved {src.name} -> {dst.name}")
        moved += 1

    print(f"\n🎉 Migration completed. Files moved: {moved}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Migrate legacy Steering filenames to standard ones")
    parser.add_argument("--apply", action="store_true", help="Apply changes (default: dry-run)")
    parser.add_argument("--force", action="store_true", help="Overwrite existing targets after backup")
    parser.add_argument("--path", type=str, default=".moai/project", help="Project doc directory path")
    args = parser.parse_args()

    steering_dir = Path(args.path)
    if not steering_dir.exists():
        print(f"❌ Steering directory not found: {steering_dir}")
        return 1

    return migrate(steering_dir, apply=args.apply, force=args.force)


if __name__ == "__main__":
    raise SystemExit(main())
