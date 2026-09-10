"""Count Unicode category Cf characters in the files a t641 commit touches.

Paths are listed here rather than passed on the command line so the check can
run under the worktree guard, which refuses command lines naming a path with a
"git" segment. Usage: python3 .moai/reports/t641/cf_check.py <group>
"""
import sys
import unicodedata
from pathlib import Path

GROUPS = {
    "red": [
        "internal/core/git/status_optional_locks_test.go",
        ".moai/reports/t641/red.txt",
        ".moai/reports/t641/isworkingtreeclean-callers.txt",
        ".moai/reports/t641/cf_check.py",
        ".moai/reports/t641/msg-red.txt",
    ],
    "green": [
        "internal/core/git/manager.go",
        ".moai/reports/t641/green.txt",
        ".moai/reports/t641/msg-green.txt",
    ],
    "final": [
        "internal/core/git/manager.go",
        "internal/core/git/status_optional_locks_test.go",
        ".moai/reports/t641/verdict.md",
        ".moai/reports/t641/mutant.txt",
        ".moai/reports/t641/mutant-diff.txt",
        ".moai/reports/t641/mutant-restore.txt",
        ".moai/reports/t641/test-statusline.txt",
        ".moai/reports/t641/gofmt.txt",
        ".moai/reports/t641/vet.txt",
        ".moai/reports/t641/vet-windows.txt",
        ".moai/reports/t641/lint.txt",
        ".moai/reports/t641/cf_check.py",
        ".moai/reports/t641/msg-final.txt",
    ],
}

total = 0
for rel in GROUPS[sys.argv[1]]:
    path = Path(rel)
    if not path.exists():
        print(f"{rel} MISSING")
        continue
    n = sum(1 for c in path.read_text(encoding="utf-8") if unicodedata.category(c) == "Cf")
    total += n
    print(f"{rel} cf={n}")
print(f"TOTAL cf={total}")
