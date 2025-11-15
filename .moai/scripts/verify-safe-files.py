#!/usr/bin/env python3
"""
verify-safe-files.py

Pre-commit verification script that ensures protected local-only files
are not accidentally deleted or modified in dangerous ways.

Protected files (must never be deleted):
  - .moai/yoda/              (Lecture material generation)
  - .moai/release/           (/moai:release command tool)
  - .moai/docs/              (Local documentation)
  - .claude/commands/moai/   (Local command definitions)

Usage:
  uv run .moai/scripts/verify-safe-files.py

Exit codes:
  0: All safe files verified
  1: Protected file deletion detected (commit blocked)
  2: Configuration error
"""

import subprocess
import sys
from pathlib import Path
from typing import List, Tuple

# ANSI color codes
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
BLUE = "\033[0;34m"
NC = "\033[0m"  # No Color

# Protected paths that must never be deleted
PROTECTED_PATHS = [
    ".moai/yoda",
    ".moai/release",
    ".moai/docs",
    ".claude/commands/moai",
]

# Protected files that must never be deleted
PROTECTED_FILES = [
    ".claude/commands/moai/release.md",
]


def log_info(msg: str) -> None:
    """Log informational message."""
    print(f"{BLUE}ℹ️  {msg}{NC}")


def log_success(msg: str) -> None:
    """Log success message."""
    print(f"{GREEN}✅ {msg}{NC}")


def log_warning(msg: str) -> None:
    """Log warning message."""
    print(f"{YELLOW}⚠️  {msg}{NC}")


def log_error(msg: str) -> None:
    """Log error message."""
    print(f"{RED}❌ {msg}{NC}")


def get_git_root() -> Path:
    """Get git repository root directory."""
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True,
            text=True,
            check=True,
        )
        return Path(result.stdout.strip())
    except subprocess.CalledProcessError:
        log_error("Not in a git repository")
        sys.exit(2)


def get_staged_files() -> Tuple[List[str], List[str], List[str]]:
    """
    Get staged files from git index.

    Returns:
        Tuple of (added_files, modified_files, deleted_files)
    """
    try:
        result = subprocess.run(
            ["git", "diff", "--cached", "--name-status"],
            capture_output=True,
            text=True,
            check=True,
        )

        added = []
        modified = []
        deleted = []

        for line in result.stdout.strip().split("\n"):
            if not line:
                continue
            status, filepath = line.split(maxsplit=1)
            if status == "A":
                added.append(filepath)
            elif status == "M":
                modified.append(filepath)
            elif status == "D":
                deleted.append(filepath)

        return added, modified, deleted
    except subprocess.CalledProcessError as e:
        log_error(f"Failed to get git status: {e}")
        sys.exit(2)


def is_protected(path: str) -> bool:
    """Check if path is protected."""
    for protected in PROTECTED_PATHS:
        if path.startswith(protected):
            return True

    for protected_file in PROTECTED_FILES:
        if path == protected_file:
            return True

    return False


def verify_protected_files(git_root: Path) -> int:
    """
    Verify that protected files are not deleted in staged changes.

    Returns:
        0 if all protected files are safe, 1 if violations detected
    """
    print(f"\n{BLUE}{'='*70}")
    print(f"🔒 Safe File Verification")
    print(f"{'='*70}{NC}\n")

    added, modified, deleted = get_staged_files()

    violations = []

    # Check for deleted protected files
    log_info("Checking for protected file deletions...")
    for deleted_file in deleted:
        if is_protected(deleted_file):
            violations.append(
                (
                    "DELETE",
                    deleted_file,
                    "Protected file deletion detected - BLOCKING COMMIT",
                )
            )

    # Check for deleted protected directories
    for deleted_dir in deleted:
        for protected_path in PROTECTED_PATHS:
            if deleted_dir.startswith(protected_path):
                violations.append(
                    (
                        "DELETE",
                        deleted_dir,
                        f"File in protected directory '{protected_path}' - BLOCKING COMMIT",
                    )
                )

    if violations:
        print()
        log_error(f"DETECTED {len(violations)} VIOLATION(S):")
        print()
        for status, path, reason in violations:
            log_error(f"  {status}: {path}")
            print(f"      → {reason}")
        print()
        print(
            f"{RED}═══════════════════════════════════════════════════════════════════{NC}"
        )
        print(f"{RED}❌ COMMIT BLOCKED - Protected Files at Risk{NC}")
        print(f"{RED}═══════════════════════════════════════════════════════════════════{NC}")
        print()
        print(f"These files are critical to MoAI-ADK and must never be deleted:")
        print()
        for protected in PROTECTED_PATHS:
            full_path = git_root / protected
            if full_path.exists():
                print(f"  {YELLOW}📁 {protected}{NC}")
        for protected_file in PROTECTED_FILES:
            full_path = git_root / protected_file
            if full_path.exists():
                print(f"  {YELLOW}📄 {protected_file}{NC}")
        print()
        print(f"If you need to modify these files, please:")
        print(f"  1. Create an issue explaining the reason")
        print(f"  2. Update .gitignore if removing protection is necessary")
        print(f"  3. Update DEPLOYMENT.md documentation")
        print()
        return 1

    # Verify protected files still exist in working directory
    log_info("Verifying protected files exist...")
    missing_files = []
    for protected_path in PROTECTED_PATHS:
        full_path = git_root / protected_path
        if not full_path.exists():
            missing_files.append(protected_path)

    for protected_file in PROTECTED_FILES:
        full_path = git_root / protected_file
        if not full_path.exists():
            missing_files.append(protected_file)

    if missing_files:
        log_warning(f"Missing {len(missing_files)} protected file(s):")
        for missing in missing_files:
            print(f"  {YELLOW}  {missing}{NC}")

    # Success
    print()
    log_success("All protected files verified - Safe to commit")
    print()
    print(f"{BLUE}═══════════════════════════════════════════════════════════════════{NC}")
    print(f"📊 Staged Changes Summary")
    print(f"{BLUE}═══════════════════════════════════════════════════════════════════{NC}")
    print(f"  Added:    {len(added)} file(s)")
    print(f"  Modified: {len(modified)} file(s)")
    print(f"  Deleted:  {len(deleted)} file(s)")
    print()

    return 0


def main() -> int:
    """Main entry point."""
    try:
        git_root = get_git_root()
        return verify_protected_files(git_root)
    except KeyboardInterrupt:
        print(f"\n{YELLOW}Interrupted by user{NC}")
        return 2
    except Exception as e:
        log_error(f"Unexpected error: {e}")
        return 2


if __name__ == "__main__":
    sys.exit(main())
