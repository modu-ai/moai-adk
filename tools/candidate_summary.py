#!/usr/bin/env python3
"""Summary and verdict of one candidate verification (#2554 PR-A).

`.github/workflows/candidate.yml` runs the check set on five targets for one pinned commit.
Its last job calls this script with the `needs` context as JSON in $NEEDS and the run's own
commit in $GITHUB_SHA. The script prints one Markdown summary on stdout (the workflow appends
it to the step summary) and exits 0 only when every one of the five targets

  - ended `success`,
  - reports the same commit the run was pinned to, and
  - ran tests: `tests_run` is a positive integer and `tests_passed` equals it.

A target that only compiled has no tests, so it cannot pass. Anything unreadable fails: the
script never raises on a malformed input, it reports FAIL.
"""

import json
import os
import re
import sys

TARGETS = ("linux-x86_64", "linux-arm64", "macos-arm64", "macos-x86_64", "windows-x86_64")
_SHA = re.compile(r"[0-9a-f]{40}")
_NUMBER = re.compile(r"[0-9]+")


def clean(value):
    """A short, single-line, table-safe rendering of an untrusted value."""
    text = str(value)[:60]
    return "".join(c if c.isprintable() and c != "|" else "?" for c in text)


def is_sha(value):
    return isinstance(value, str) and _SHA.fullmatch(value) is not None


def number(value):
    """A non-negative integer from a digits-only string, else None."""
    if isinstance(value, str) and _NUMBER.fullmatch(value):
        return int(value)
    return None


def outputs_of(entry):
    outputs = entry.get("outputs") if isinstance(entry, dict) else None
    return outputs if isinstance(outputs, dict) else {}


def judge(needs, expected):
    """(ok, markdown) for the `needs` context and the commit the run was pinned to."""
    problems = []
    if not is_sha(expected):
        problems.append("the run's commit is not a 40-character lowercase hex SHA")
    needs = needs if isinstance(needs, dict) else {}

    pin = needs.get("pin")
    pin_sha = outputs_of(pin).get("sha")
    if not isinstance(pin, dict) or pin.get("result") != "success":
        problems.append("pin: did not succeed")
    elif pin_sha != expected:
        problems.append(f"pin: reports {clean(pin_sha)}, the run is {clean(expected)}")

    rows = []
    for target in TARGETS:
        entry = needs.get(target)
        if not isinstance(entry, dict):
            problems.append(f"{target}: missing")
            rows.append((target, "missing", "-", "-", "-", "-"))
            continue
        result = entry.get("result")
        outputs = outputs_of(entry)
        sha = outputs.get("sha")
        run = number(outputs.get("tests_run"))
        passed = number(outputs.get("tests_passed"))
        skipped = outputs.get("tests_skipped")
        if result != "success":
            problems.append(f"{target}: result {clean(result)!r}, not success")
        if sha != expected:
            problems.append(f"{target}: checked out {clean(sha)}, the run is {clean(expected)}")
        if run is None or run == 0:
            problems.append(
                f"{target}: no tests ran (tests_run={clean(outputs.get('tests_run'))!r})"
                " - compiling is not a pass"
            )
        elif passed != run:
            problems.append(f"{target}: {clean(outputs.get('tests_passed'))} of {run} tests passed")
        rows.append(
            (
                target,
                clean(result),
                clean(sha),
                clean(outputs.get("tests_run")),
                clean(outputs.get("tests_passed")),
                clean(skipped),
            )
        )

    ok = not problems
    heading = clean(expected) if is_sha(expected) else "(unreadable)"
    lines = [f"## candidate {heading}", ""]
    lines += [
        "| target | result | commit | tests run | passed | skipped |",
        "| --- | --- | --- | --- | --- | --- |",
    ]
    lines += ["| " + " | ".join(row) + " |" for row in rows]
    lines.append("")
    if ok:
        lines.append("verdict: PASS - five targets, one commit, tests run on each")
    else:
        lines.append("verdict: FAIL")
        lines.append("")
        lines += [f"- {p}" for p in problems]
    return ok, "\n".join(lines) + "\n"


def main():
    expected = os.environ.get("GITHUB_SHA", "")
    raw = os.environ.get("NEEDS")
    if raw is None:
        sys.stdout.write("## candidate\n\nverdict: FAIL\n\n- NEEDS is not set\n")
        return 1
    try:
        needs = json.loads(raw)
    except ValueError:
        sys.stdout.write("## candidate\n\nverdict: FAIL\n\n- NEEDS is not JSON\n")
        return 1
    ok, text = judge(needs, expected)
    sys.stdout.write(text)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
