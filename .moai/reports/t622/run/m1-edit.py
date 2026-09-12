"""M1 line-level edit for card t622 (non manager-git files).

Each old string must occur exactly once in both the local and the template
copy; run with --dry first. Local and template copies are edited in place at
their own positions (no whole-file copy).
"""
import sys

T = "internal/template/templates/"

DL_OLD = """Auto-merge trigger conditions:
- `is_worktree_context == true` AND `--no-merge` flag NOT set
- OR `--merge` flag explicitly set (deprecated, logged as warning)

When auto-merge is triggered:
1. Verify all CI/CD checks pass (gh pr checks)
2. Verify zero merge conflicts (gh pr view --json mergeable)
3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

##### Flag Behavior

- `--no-merge`: Skip auto-merge even in worktree context. PR is created but not merged.
- `--merge`: Deprecated. Logs warning: "The --merge flag is deprecated. Auto-merge is now the default for worktree contexts."
"""

DL_NEW = """Merging is opt-in. The single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge; worktree context alone never triggers a merge.

Auto-merge trigger conditions:
- `--auto-merge` flag set
- OR `--merge` flag set (deprecated alias of `--auto-merge`, logged as warning)

Mode conditions (same as `manager-git.md` § PR Auto-Merge):
- In team mode, `--auto-merge` merges only after all approvals are obtained.
- In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve).

When auto-merge is triggered:
1. Verify all CI/CD checks pass (gh pr checks)
2. Verify zero merge conflicts (gh pr view --json mergeable)
3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

##### Flag Behavior

- `--auto-merge`: Opt in to merging the PR after sync, under the mode conditions above.
- `--merge`: deprecated alias of `--auto-merge` (logs a warning).
- `--no-merge`: Deprecated no-op kept for compatibility (logs a warning); not merging is already the default.
"""

NEXT_OLD = "- Auto-Merge PR (/moai sync --merge)\n"
NEXT_NEW = "- Auto-Merge PR (/moai sync --auto-merge)\n"

DE_OLD = "This affects auto-merge behavior: worktree contexts default to auto-merge.\n"
DE_NEW = ("This does not decide auto-merge: the PR merges only on the `--auto-merge` opt-in "
          "defined in `manager-git.md` § PR Auto-Merge, never on worktree context alone.\n")

SKILL_OLD = "Modes: auto, force, status, project. Flags: --merge, --skip-mx\n"
SKILL_NEW = "Modes: auto, force, status, project. Flags: --auto-merge, --merge (deprecated alias of --auto-merge), --skip-mx\n"

REF_OLD = "- --merge: Auto-merge PR and clean up branch after sync\n"
REF_NEW = "- --auto-merge: Opt in to auto-merge the PR and clean up the branch after sync\n"

QGC_ARGS_OLD = "  - Flag: --merge\n"
QGC_ARGS_NEW = "  - Flag: --auto-merge\n"

QGC_FLAGS_OLD = "- --merge: After sync, auto-merge PR and clean up branch. Worktree/branch environment is auto-detected from git context.\n"
QGC_FLAGS_NEW = ("- --auto-merge: After sync, auto-merge PR and clean up branch (opt-in; merge conditions per "
                 "`manager-git.md` § PR Auto-Merge). Worktree/branch environment is auto-detected from git context.\n"
                 "- --merge: deprecated alias of --auto-merge (logs a warning).\n")

SYNC_USAGE_OLD = "/moai sync [mode] [--pr] [--merge] [--skip-mx]\n"
SYNC_USAGE_NEW = "/moai sync [mode] [--pr] [--auto-merge] [--skip-mx]\n"

SYNC_FLAGS_OLD = "**Flags**: `--pr` (PR 생성) | `--merge` (deprecated, auto-merge) | `--skip-mx` (MX 검증 스킵)\n"
SYNC_FLAGS_NEW = "**Flags**: `--pr` (PR 생성) | `--auto-merge` (auto-merge 옵트인) | `--merge` (deprecated alias of `--auto-merge`) | `--skip-mx` (MX 검증 스킵)\n"

EDITS = [
    (".claude/skills/moai/workflows/sync/delivery.md", [(DL_OLD, DL_NEW), (NEXT_OLD, NEXT_NEW)]),
    (".claude/skills/moai/workflows/sync/doc-execution.md", [(DE_OLD, DE_NEW)]),
    (".claude/skills/moai/SKILL.md", [(SKILL_OLD, SKILL_NEW)]),
    (".claude/skills/moai/references/reference.md", [(REF_OLD, REF_NEW)]),
    (".claude/skills/moai/workflows/sync/quality-gates-context.md", [(QGC_ARGS_OLD, QGC_ARGS_NEW), (QGC_FLAGS_OLD, QGC_FLAGS_NEW)]),
    (".claude/skills/moai/workflows/sync.md", [(SYNC_USAGE_OLD, SYNC_USAGE_NEW), (SYNC_FLAGS_OLD, SYNC_FLAGS_NEW)]),
]


def main() -> int:
    dry = len(sys.argv) > 1 and sys.argv[1] == "--dry"
    staged = {}
    ok = True
    for rel, pairs in EDITS:
        for path in (rel, T + rel):
            with open(path, encoding="utf-8") as f:
                text = f.read()
            for old, new in pairs:
                n = text.count(old)
                print(f"{path}: count={n} for {old[:40]!r}")
                if n != 1:
                    ok = False
                    continue
                text = text.replace(old, new, 1)
            staged[path] = text
    if not ok:
        print("ABORT: a pattern did not match exactly once; nothing written")
        return 1
    if dry:
        print("dry-run: all patterns matched exactly once")
        return 0
    for path, text in staged.items():
        with open(path, "w", encoding="utf-8") as f:
            f.write(text)
    print(f"applied to {len(staged)} files")
    return 0


if __name__ == "__main__":
    sys.exit(main())
