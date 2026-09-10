#!/usr/bin/env python3
"""D1 corpus cleanup — remove the `status:` line from non-spec.md SPEC artifacts.

SPEC-ARTIFACT-STATELESS-001 M3 (REQ-AST-001-008 / -011).

Predicate — deliberately IDENTICAL to the lint rule
(`internal/spec/lint_artifact_status.go` `frontmatterStatusLine`), so the
checker and its cleanup select the same set:

  1. The frontmatter block opens ONLY at line 1 (a `---` later in the body is
     body text, not a fence).
  2. `---` is matched as a PREFIX, so a `----` rule line closes the block.
  3. A field line is `status:` as a prefix — no whitespace required after the
     colon, which is WIDER than the counting predicate `^status:[[:space:]]`
     used to size the corpus. The widening is measured, not assumed: the corpus
     carries zero `status:<non-space>` lines today, so the two select the same
     set now and can diverge only on a line a future author writes.

Scope guards, in code rather than in prose:

  - Only the four governed artifact names are opened. `spec.md` and
    `progress.md` are never opened at all (REQ-AST-001-011), so no exclusion
    can be forgotten.
  - Only lines inside the leading frontmatter block are removed. Body text is
    never rewritten.
  - Nothing but whole `status:` lines is removed — every other frontmatter
    field survives byte-identically (D1, not D2).

Usage:
    python3 .moai/reports/t357/t357_d1_clean.py            # apply
    python3 .moai/reports/t357/t357_d1_clean.py --dry-run  # report only
"""

import pathlib
import sys

ARTIFACTS = ("plan.md", "acceptance.md", "design.md", "research.md")


def strip_status(text: str) -> tuple[str, int]:
    """Return the text with frontmatter `status:` lines removed, and the count."""
    lines = text.split("\n")
    if not lines or not lines[0].startswith("---"):
        return text, 0
    kept = [lines[0]]
    removed = 0
    i = 1
    while i < len(lines):
        line = lines[i]
        if line.startswith("---"):
            break
        if line.startswith("status:"):
            removed += 1
        else:
            kept.append(line)
        i += 1
    if removed == 0:
        return text, 0
    kept.extend(lines[i:])
    return "\n".join(kept), removed


def main() -> int:
    dry_run = "--dry-run" in sys.argv
    root = pathlib.Path(".moai/specs")
    if not root.is_dir():
        print("FAIL — .moai/specs not found; run from the project root", file=sys.stderr)
        return 2

    files = 0
    lines = 0
    for spec_dir in sorted(root.glob("SPEC-*")):
        if not spec_dir.is_dir():
            continue
        for name in ARTIFACTS:
            path = spec_dir / name
            if not path.is_file():
                continue
            original = path.read_text(encoding="utf-8")
            cleaned, removed = strip_status(original)
            if removed == 0:
                continue
            files += 1
            lines += removed
            print(f"{path}  (-{removed})")
            if not dry_run:
                path.write_text(cleaned, encoding="utf-8")

    verb = "would clean" if dry_run else "cleaned"
    print(f"{verb}: {files} files, {lines} status lines removed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
