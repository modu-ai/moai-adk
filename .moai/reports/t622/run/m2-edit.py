"""M2 line-level edit for card t622 (delivery.md, both copies).

Each old string must occur exactly once in both the local and the template
copy; run with --dry first. No whole-file copy.
"""
import sys

T = "internal/template/templates/"
REL = ".claude/skills/moai/workflows/sync/delivery.md"

A_OLD = """3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge
"""
A_NEW = """3. If all checks pass: Execute `gh pr merge --<merge_method> --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

`<merge_method>` is resolved from `git_strategy.<mode>.merge_method` for the active mode (`squash` | `merge` | `rebase`; default `squash`).
"""

B_OLD = "3. If passing and mergeable: Execute `gh pr merge --squash --delete-branch`\n"
B_NEW = "3. If passing and mergeable: Execute `gh pr merge --<merge_method> --delete-branch`\n"


def main() -> int:
    dry = len(sys.argv) > 1 and sys.argv[1] == "--dry"
    staged = {}
    ok = True
    for path in (REL, T + REL):
        with open(path, encoding="utf-8") as f:
            text = f.read()
        for old, new in ((A_OLD, A_NEW), (B_OLD, B_NEW)):
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
