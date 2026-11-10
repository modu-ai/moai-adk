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
40 lowercase hex characters, each as a whole whitespace-separated token (39 or
41 hex, or a suffix such as `-dirty`, fail); otherwise it must start with
`not_run` followed by a reason containing a letter or digit. A body with no `ci:` line is still accepted,
because citing CI is allowed only for merge-only deltas. This checker reads
text only: it cannot tell whether the run exists or whether its head equals
the PR head; the auditor reads the run back with `gh api`. The independent audit stays in
place — this tool only pulls the audit evidence forward, before the PR is
opened.

Re-audits (Issue #2388) require explicit `--reaudit` (or check(body, reaudit=True));
neither template fields nor a claim epoch enables this mode. In the exact
`## 독립 review` section, up to the next Markdown heading, require exactly one
line for each field (whitespace and Markdown list markers are accepted):

  - 수리 확인 범위: F1 path/boundary repaired and touched scope
  - 새 지적 분류: 없음

The classification can instead be `차단: <concrete reason>` or
`후속 Issue #<positive ASCII integer>: <concrete reason>`. Angle-bracket text is
documentation notation, not an accepted value. Empty, duplicate, punctuation-only,
TODO/TBD/not_run and angle-placeholder values fail. Contract violations, data
damage and security findings are blocking; other new findings belong in a
follow-up Issue. This is a text-completeness check only: it cannot judge repair
truth, classification policy compliance, or whether the follow-up Issue exists.
Initial audits keep the existing checks and output; the AUDIT verdict format
and CI requirements are unchanged.

Usage:
  python3 -I tools/check_pr_self_audit.py --file BODY.md
  python3 -I tools/check_pr_self_audit.py < BODY.md
  python3 -I tools/check_pr_self_audit.py --reaudit < BODY.md
  python3 -I tools/check_pr_self_audit.py --reaudit --file BODY.md
  python3 -I tools/check_pr_self_audit.py --file BODY.md --reaudit

Exit codes: 0 = section complete, 1 = section/items missing or unmeasured,
2 = usage.
"""

import re
import sys
from pathlib import Path

SECTION = "## self-audit"
TAGS = ("mutation", "clone", "host", "receipt")
CI_TAG = "ci:"
# Keyed fields are whole whitespace-separated tokens: whitespace (or the line
# edge) on both sides, so `rerun=7`, `.run=7`, a 41-hex string or a suffix such
# as `<40hex>-dirty` never passes as `run=` or `head=`.
CI_RUN = re.compile(r"(?<!\S)run=\d+(?!\S)")
CI_HEAD = re.compile(r"(?<!\S)head=[0-9a-f]{40}(?!\S)")
# `not_run` followed by a reason that has at least one letter or digit.
CI_NOT_RUN = re.compile(r"^not_run\b(?=.*\w)")


def fail(missing):
    print("self-audit: blocked missing=" + ",".join(missing))
    raise SystemExit(1)


def strip_marker(line: str) -> str:
    """Drop leading whitespace and markdown list markers before tag matching."""
    return line.strip().lstrip("-*").strip().lower()


def check(body: str, *, reaudit: bool = False) -> None:
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
    if reaudit:
        check_reaudit(lines)
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


def meaningful(value: str) -> bool:
    """Reject unfilled evidence without attempting to judge its truth."""
    return (
        any(char.isalnum() for char in value)
        and not re.search(r"(?<![A-Za-z0-9])(?:TODO|TBD|not_run)(?![A-Za-z0-9])",
                          value, re.IGNORECASE)
        and not re.search(r"<[^>]*>", value)
    )


def check_reaudit(lines) -> None:
    """Require one repair scope and one explicit new-finding classification."""
    starts = [i for i, line in enumerate(lines) if line.strip() == "## 독립 review"]
    if not starts:
        fail(["## 독립 review"])
    fields = {"수리 확인 범위": [], "새 지적 분류": []}
    for line in lines[starts[0] + 1:]:
        if re.match(r"^\s*#{1,6}(?:\s|$)", line):
            break
        match = re.fullmatch(
            r"\s*(?:(?:[-*+]|[0-9]+[.)])\s+)?"
            r"(수리 확인 범위|새 지적 분류)\s*:\s*(.*)", line,
        )
        if match:
            fields[match[1]].append(match[2].strip())
    missing = []
    scope = fields["수리 확인 범위"]
    if len(scope) != 1 or not meaningful(scope[0]):
        missing.append("reaudit:scope")
    findings = fields["새 지적 분류"]
    valid_findings = False
    if len(findings) == 1 and meaningful(findings[0]):
        if findings[0] == "없음":
            valid_findings = True
        else:
            match = re.fullmatch(
                r"(?:차단|후속\s+Issue\s+#[1-9][0-9]*)\s*:\s*(.*)",
                findings[0],
            )
            valid_findings = bool(match and meaningful(match[1]))
    if not valid_findings:
        missing.append("reaudit:findings")
    if missing:
        fail(missing)


def usage() -> None:
    print(__doc__)
    raise SystemExit(2)


def main() -> None:
    args = [a for a in sys.argv[1:]]
    reaudit = False
    if args and args[0] == "--reaudit":
        reaudit = True
        args.pop(0)
    elif args and args[-1] == "--reaudit":
        reaudit = True
        args.pop()
    if args and args[0] == "--file" and len(args) == 2 and not args[1].startswith("--"):
        body = Path(args[1]).read_text(encoding="utf-8")
    elif args:
        usage()
    elif sys.stdin.isatty():
        usage()
    else:
        body = sys.stdin.read()
    check(body, reaudit=reaudit)


if __name__ == "__main__":
    main()
