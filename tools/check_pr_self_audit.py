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

Zero-selected evidence (Issue #2403, item 5). In the section and in every
markdown table row of the body:

  - a line that reports `0 passed` (a run that selected no test) must also say
    `not_run` on the same line; otherwise it is blocked as zero-selected;
  - a `cargo test` command that passes a test-name filter must select exactly
    one target (`--lib`, `--bin X`, `--test X`, ...), because cargo applies the
    filter to every target of the call; a filtered call with two or more
    targets, or with none (cargo then runs several by default), is blocked as
    filter-multi-target. Only commands written in a code span (`...`) are
    read; prose that mentions cargo test is not a command.
    Quoted option values remain single POSIX shell words. Named target glob
    patterns (`*`, `?`, `[]`) may select several targets, including in `--test=X`
    form, so they cannot establish single-target filtered evidence. No Cargo
    command or GitHub request is executed by this checker.

Usage:
  python3 -I tools/check_pr_self_audit.py --file BODY.md
  python3 -I tools/check_pr_self_audit.py < BODY.md

Exit codes: 0 = section complete, 1 = section/items missing or unmeasured,
2 = usage.
"""

import re
import shlex
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


# `0 passed` as a whole count: not preceded by another digit, so `10 passed`
# and `20 passed` are not zero.
ZERO_PASSED = re.compile(r"(?<![0-9])0 passed\b")
# cargo test options that take a value as the next token (the value is not a
# test-name filter).
CARGO_VALUE_OPTS = {
    "--test", "--bin", "--example", "--bench", "-p", "--package", "-F",
    "--features", "--manifest-path", "--target", "--target-dir", "-j",
    "--jobs", "--profile", "--color", "--message-format", "--config", "-Z",
    "--exclude",
}
# Target selectors and how many targets each selects.
CARGO_TARGETS_ONE = {"--lib", "--doc", "--test", "--bin", "--example", "--bench"}
CARGO_TARGETS_MANY = {"--tests", "--bins", "--examples", "--benches", "--all-targets"}
# Test-harness options after `--` that take a value as the next token.
HARNESS_VALUE_OPTS = {"--skip", "--test-threads", "--format", "--color", "-Z", "--logfile", "--report-time"}


def cargo_test_commands(line: str):
    """The token lists of the `cargo test` commands written in code spans on one
    line: each runs from `test` to the end of its span or a shell separator.
    Prose such as "cargo test 671 passed" outside a code span is not a command."""
    out = []
    for span in re.findall(r"`([^`]+)`", line):
        if not re.search(r"\bcargo\b", span):
            continue
        lexer = shlex.shlex(span, posix=True, punctuation_chars=";&|()")
        lexer.whitespace_split = True
        lexer.commenters = ""
        words = list(lexer)
        i = 0
        while i < len(words):
            if words[i] != "cargo":
                i += 1
                continue
            test = i + 1
            if test < len(words) and words[test].startswith("+"):
                test += 1
            if test == len(words) or words[test] != "test":
                i += 1
                continue
            end = test + 1
            expects_value = False
            while end < len(words):
                token = words[end]
                if not expects_value and token in {";", "||", "&&", "|", ")"}:
                    break
                expects_value = not expects_value and (
                    token in CARGO_VALUE_OPTS or token in HARNESS_VALUE_OPTS
                )
                end += 1
            out.append(words[test + 1:end])
            i = end
    return out


def filtered_target_count(tokens):
    """(has a name filter, number of targets selected; None = cargo's default
    set, which is more than one)."""
    has_filter = False
    targets = 0
    many = False
    after_dashdash = False
    skip_next = None
    for tok in tokens:
        if skip_next:
            if skip_next in CARGO_TARGETS_ONE and any(c in tok for c in "*?["):
                many = True
            skip_next = None
            continue
        if after_dashdash:
            if tok in HARNESS_VALUE_OPTS:
                skip_next = tok
            elif not tok.startswith("-"):
                has_filter = True
            continue
        if tok == "--":
            after_dashdash = True
            continue
        name = tok.split("=", 1)[0]
        if name in CARGO_TARGETS_ONE:
            targets += 1
            if "=" in tok and any(c in tok.split("=", 1)[1] for c in "*?["):
                many = True
        elif name in CARGO_TARGETS_MANY:
            many = True
        if name in CARGO_VALUE_OPTS and "=" not in tok:
            skip_next = name
        elif not tok.startswith("-"):
            has_filter = True
    if many:
        return has_filter, 2
    return has_filter, (targets if targets else None)


def check_zero_selected(section, lines) -> None:
    """Block zero-selected test evidence in the section and in table rows."""
    scope = list(section) + [ln for ln in lines if ln.lstrip().startswith("|")]
    for line in scope:
        if ZERO_PASSED.search(line) and "not_run" not in line:
            fail(["zero-selected"])
        for tokens in cargo_test_commands(line):
            has_filter, count = filtered_target_count(tokens)
            if has_filter and (count is None or count > 1):
                fail(["filter-multi-target"])


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
    check_zero_selected(section, lines)
    # The mutation line must report the mutation actually failing a check
    # (measured, not asserted): the word FAILED must appear on it.
    mut = [ln for ln in section if strip_marker(ln).startswith("mutation:")][0]
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
