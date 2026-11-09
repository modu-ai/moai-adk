#!/usr/bin/env python3
"""PR self-audit section checker (Issue #2279, scope 3).

Every lane PR must carry a `## Self-Audit` section with four measured lines
(read-only checker; it judges text, it never posts anything):

  mutation: what guard was mutated and that the mutation FAILED a check
  clone:    the check set run from a fresh clone (symlink paths like /tmp cannot
            leak pre-seeded state in)
  host:     external hosts exercised through a fake executable (stub), not the
            network
  receipt:  how a false-pass receipt was ruled out (digest/reference checked)

A PR body without the section, missing any of the four lines, or a mutation
line that does not report FAILED exits 1. The independent audit stays in
place — this tool only pulls the audit evidence forward, before the PR is
opened.

Usage:
  python3 -I tools/check_pr_self_audit.py --file BODY.md
  python3 -I tools/check_pr_self_audit.py < BODY.md

Exit codes: 0 = section complete, 1 = section/items missing or unmeasured,
2 = usage.
"""

import sys
from pathlib import Path

SECTION = "## self-audit"
TAGS = ("mutation", "clone", "host", "receipt")


def fail(missing):
    print("self-audit: blocked missing=" + ",".join(missing))
    raise SystemExit(1)


def strip_marker(line: str) -> str:
    """Drop leading whitespace and markdown list markers before tag matching."""
    return line.strip().lstrip("-*").strip().lower()


def check(body: str) -> None:
    lines = body.splitlines()
    starts = [i for i, l in enumerate(lines) if l.strip().lower() == SECTION]
    if not starts:
        fail([SECTION])
    start = starts[0]
    # The section runs to the next markdown heading of any level.
    section = []
    for line in lines[start + 1:]:
        if line.startswith("#"):
            break
        section.append(line)
    missing = []
    for tag in TAGS:
        hit = [l for l in section if strip_marker(l).startswith(tag + ":")]
        if not hit or not hit[0].split(":", 1)[1].strip():
            missing.append(tag)
    if missing:
        fail(missing)
    # The mutation line must report the mutation actually failing a check
    # (measured, not asserted): the word FAILED must appear on it.
    mut = [l for l in section if strip_marker(l).startswith("mutation:")][0]
    if "failed" not in mut.lower():
        fail(["mutation:FAILED"])
    print("self-audit: pass (mutation/clone/host/receipt measured)")


def usage() -> None:
    print(__doc__)
    raise SystemExit(2)


def main() -> None:
    args = [a for a in sys.argv[1:]]
    if args and args[0] == "--file" and len(args) == 2:
        body = Path(args[1]).read_text(encoding="utf-8")
    elif args:
        usage()
    elif sys.stdin.isatty():
        usage()
    else:
        body = sys.stdin.read()
    check(body)


if __name__ == "__main__":
    main()
