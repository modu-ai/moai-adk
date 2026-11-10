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
line that does not report FAILED exits 1.

Optional CI citation lines (Issue #2385; rules in skills/moai-cli/workflows/
verify.md, section "Check-run rules"):

  ci: rc-ci run=<run id> head=<40-hex head SHA>
  ci: not_run <reason>

Each `ci:` line in the section is checked. A citation must carry the keyed
field `run=` with one or more digits and the keyed field `head=` with exactly
40 lowercase hex characters (39 or 41 fail); otherwise it must start with
`not_run` followed by a reason. A body with no `ci:` line is still accepted,
because citing CI is allowed only for merge-only deltas. This checker reads
text only: it cannot tell whether the run exists or whether its head equals
the PR head; the auditor reads the run back with `gh api`. The independent audit stays in
place — this tool only pulls the audit evidence forward, before the PR is
opened.

Usage:
  python3 -I tools/check_pr_self_audit.py --file BODY.md
  python3 -I tools/check_pr_self_audit.py < BODY.md

Exit codes: 0 = section complete, 1 = section/items missing or unmeasured,
2 = usage.
"""

import re
import sys
from pathlib import Path

SECTION = "## self-audit"
TAGS = ("mutation", "clone", "host", "receipt")
CI_TAG = "ci:"
# Keyed fields, bounded on both sides so `rerun=7` or a 41-hex string never
# passes as `run=` or `head=`.
CI_RUN = re.compile(r"(?<![\w-])run=\d+(?![\w-])")
CI_HEAD = re.compile(r"(?<![\w-])head=[0-9a-f]{40}(?![0-9A-Za-z])")
CI_NOT_RUN = re.compile(r"^not_run\b\s*\S")


def fail(missing):
    print("self-audit: blocked missing=" + ",".join(missing))
    raise SystemExit(1)


def strip_marker(line: str) -> str:
    """Drop leading whitespace and markdown list markers before tag matching."""
    return line.strip().lstrip("-*").strip().lower()


def check(body: str) -> None:
    lines = body.splitlines()
    starts = [i for i, ln in enumerate(lines) if ln.strip().lower() == SECTION]
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
        hit = [ln for ln in section if strip_marker(ln).startswith(tag + ":")]
        if not hit or not hit[0].split(":", 1)[1].strip():
            missing.append(tag)
    if missing:
        fail(missing)
    # The mutation line must report the mutation actually failing a check
    # (measured, not asserted): the word FAILED must appear on it.
    mut = [ln for ln in section if strip_marker(ln).startswith("mutation:")][0]
    if "failed" not in mut.lower():
        fail(["mutation:FAILED"])
    check_ci_lines(section)
    print("self-audit: pass (mutation/clone/host/receipt measured)")


def check_ci_lines(section) -> None:
    """Every `ci:` line is either a keyed citation (run id + 40-hex head) or
    `not_run` with a reason."""
    for line in section:
        if not strip_marker(line).startswith(CI_TAG):
            continue
        value = line.split(":", 1)[1].strip()
        if CI_NOT_RUN.match(value):
            continue
        missing = []
        if not CI_RUN.search(value):
            missing.append("ci:run")
        if not CI_HEAD.search(value):
            missing.append("ci:head")
        if missing:
            fail(missing)


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
