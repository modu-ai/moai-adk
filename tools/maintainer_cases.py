#!/usr/bin/env python3
"""Replay the maintainer evaluation cases (SPEC-20261121-maintainer-eval-cases).

A case is a lesson from past work and one mechanical check that tells whether the lesson holds.
Each case has targets: the frozen evidence of the incident (a `flag` is expected), its repaired
twin (`clean`), or the current tree (`clean`, or `open` while a known problem is not yet fixed).
Standard library only, no model, no network, nothing written.

    python3 tools/maintainer_cases.py replay [--root DIR] [--cases FILE]

Exit codes: 0 every target ended as expected, 1 a strict target ended with another verdict,
2 refused (the reason is on stderr with the prefix `maintainer_cases: `, stdout stays empty).
Written for Python 3.9.
"""

import argparse
import json
import re
import sys
from pathlib import Path

PREFIX = "maintainer_cases: "
EXPECTS = ("flag", "clean", "open")
CASE_KEYS = ("id", "lesson", "report", "check", "targets")
TOOL_ROOT = Path(__file__).resolve().parent.parent


class Refusal(Exception):
    """The replay cannot give a trustworthy answer; nothing goes to stdout."""


# --- reading -----------------------------------------------------------------


def read_text(path):
    try:
        return Path(path).read_text(encoding="utf-8")
    except FileNotFoundError:
        raise Refusal("%s does not exist" % path)
    except (OSError, UnicodeDecodeError) as exc:
        raise Refusal("%s cannot be read as UTF-8 text: %s" % (path, exc))


def read_json(path):
    try:
        return json.loads(read_text(path))
    except json.JSONDecodeError as exc:
        raise Refusal("%s is not valid JSON: %s" % (path, exc))


def inside(root, rel, what):
    if not isinstance(rel, str) or not rel:
        raise Refusal("%s must be a non-empty path" % what)
    if Path(rel).is_absolute() or ".." in Path(rel).parts:
        raise Refusal("%s must stay inside the root: %s" % (what, rel))
    return Path(root) / rel


def string_list(value, what):
    if not isinstance(value, list) or not all(isinstance(v, str) for v in value):
        raise Refusal("%s must be a list of strings" % what)
    return value


# --- the four checks: each returns (verdict, detail), verdict is "flag" or "clean" ----------


def claims_in_diff(root, path):
    """A claimed path that the change did not touch is a repair that was only written down."""
    claims = string_list(read_json(path / "claims.json"), "%s/claims.json" % path)
    changed = {
        line.strip()
        for line in read_text(path / "changed-files.txt").splitlines()
        if line.strip()
    }
    missing = [c for c in claims if c not in changed]
    if missing:
        return "flag", "claimed but not changed: " + ", ".join(missing)
    return "clean", "all %d claimed path(s) are in the changed files" % len(claims)


def blank_rust(text):
    """The text with comments and the inside of string and char literals blanked, same length."""
    out = list(text)
    n = len(text)
    i = 0

    def blank(a, b):
        for k in range(a, min(b, n)):
            if out[k] != "\n":
                out[k] = " "

    while i < n:
        two = text[i : i + 2]
        c = text[i]
        if two == "//":
            j = text.find("\n", i)
            j = n if j < 0 else j
            blank(i, j)
            i = j
        elif two == "/*":
            depth, j = 1, i + 2
            while j < n and depth:
                if text[j : j + 2] == "/*":
                    depth, j = depth + 1, j + 2
                elif text[j : j + 2] == "*/":
                    depth, j = depth - 1, j + 2
                else:
                    j += 1
            blank(i, j)
            i = j
        elif c == "r" and (i == 0 or not (text[i - 1].isalnum() or text[i - 1] == "_")):
            m = re.compile(r'r(#*)"').match(text, i)
            if not m:
                i += 1
                continue
            end = text.find('"' + m.group(1), m.end())
            j = n if end < 0 else end + 1 + len(m.group(1))
            blank(i, j)
            i = j
        elif c == '"':
            j = i + 1
            while j < n and text[j] != '"':
                j += 2 if text[j] == "\\" else 1
            blank(i, j + 1)
            i = j + 1
        elif c == "'":
            if text[i + 1 : i + 2] == "\\":
                end = text.find("'", i + 3)
                j = n if end < 0 else end + 1
                blank(i, j)
                i = j
            elif text[i + 2 : i + 3] == "'":
                blank(i, i + 3)
                i += 3
            else:
                i += 1  # a lifetime
        else:
            i += 1
    return "".join(out)


RUST_TEST_ATTR = re.compile(r"#\[\s*(?:\w+::)*test\s*\]")
RUST_AFTER_ATTR = re.compile(
    r"\s*(?:#\[[^\]]*\]\s*)*(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+(\w+)"
)
CLOCK = re.compile(r"Instant::now|SystemTime::now|\.elapsed\(\)")


def no_clock_in_tests(root, path):
    """A test that reads a clock bounds time, so it fails on a loaded machine and not on a bug."""
    code = blank_rust(read_text(path))
    flagged = []
    tests = 0
    for attr in RUST_TEST_ATTR.finditer(code):
        head = RUST_AFTER_ATTR.match(code, attr.end())
        if not head:
            continue
        start = code.find("{", head.end())
        semi = code.find(";", head.end())
        if start < 0 or (0 <= semi < start):
            continue
        depth, j = 0, start
        while j < len(code):
            if code[j] == "{":
                depth += 1
            elif code[j] == "}":
                depth -= 1
                if depth == 0:
                    break
            j += 1
        tests += 1
        if CLOCK.search(code[start : j + 1]):
            flagged.append(head.group(1))
    if flagged:
        return "flag", "test(s) that read a clock: " + ", ".join(flagged)
    return "clean", "no clock in %d test function(s)" % tests


def stage_line(root, spec, what):
    file = inside(root, spec.get("file") if isinstance(spec, dict) else None, what + " file")
    phrase = spec.get("contains")
    if not isinstance(phrase, str) or not phrase:
        raise Refusal("%s needs a non-empty 'contains' phrase" % what)
    for line in read_text(file).splitlines():
        pos = line.find(phrase)
        if pos >= 0:
            return line, pos
    raise Refusal("%s: no line of %s holds %r" % (what, file, phrase))


def stage_order(root, path):
    """The guidance round line must hold every stage of the chain line, in the chain's order."""
    spec = read_json(path)
    if not isinstance(spec, dict):
        raise Refusal("%s must be a JSON object" % path)
    chain_line, chain_pos = stage_line(root, spec.get("chain"), "chain")
    stages, rest = [], chain_line[chain_pos:]
    while True:
        m = re.match(r"([A-Za-z][\w-]*)", rest)
        if not m:
            break
        stages.append(m.group(1))
        arrow = re.match(r"\s*→\s*", rest[m.end() :])
        if not arrow:
            break
        rest = rest[m.end() + arrow.end() :]
    if len(stages) < 2:
        raise Refusal("chain: no stage list after %r" % spec["chain"]["contains"])
    round_line, round_pos = stage_line(root, spec.get("round"), "round")
    at, bad = round_pos, []
    for stage in stages:
        m = re.compile(r"(?<![\w])%s(?![\w])" % re.escape(stage)).search(round_line, at)
        if m:
            at = m.end()
        else:
            bad.append(stage)
    if bad:
        return "flag", "missing or out of order: " + ", ".join(bad) + " (chain: " + " → ".join(stages) + ")"
    return "clean", "all %d stages in order" % len(stages)


RUST_FN = re.compile(r"(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+(\w+)")
PY_DEF = re.compile(r"(?:async\s+)?def\s+(\w+)")


def unscoped_test_deletion(root, path):
    """A test removed outside what the change was allowed to remove is a weaker suite."""
    allowed = set(string_list(read_json(path / "allowed.json"), "%s/allowed.json" % path))
    removed, added, pending = set(), set(), False
    for line in read_text(path / "diff.patch").splitlines():
        if line.startswith(("--- ", "+++ ", "diff ", "index ", "@@")):
            pending = False
        elif line.startswith("-"):
            s = line[1:].strip()
            if re.match(r"#\[\s*(?:\w+::)*test\b", s):
                pending = True
            elif s == "" or s.startswith(("#[", "///")):
                continue
            else:
                m = RUST_FN.match(s)
                if m and pending:
                    removed.add(m.group(1))
                py = PY_DEF.match(s)
                if py and py.group(1).startswith("test"):
                    removed.add(py.group(1))
                pending = False
        elif line.startswith("+"):
            s = line[1:].strip()
            m = RUST_FN.match(s) or PY_DEF.match(s)
            if m:
                added.add(m.group(1))
        else:
            pending = False
    flagged = sorted(removed - added - allowed)
    if flagged:
        return "flag", "removed test(s) outside the allowed list: " + ", ".join(flagged)
    return "clean", "%d removed test(s), all allowed or put back" % len(removed)


CHECKS = {
    "claims_in_diff": (claims_in_diff, "dir"),
    "no_clock_in_tests": (no_clock_in_tests, "file"),
    "stage_order": (stage_order, "file"),
    "unscoped_test_deletion": (unscoped_test_deletion, "dir"),
}


# --- the replay --------------------------------------------------------------


def load_cases(path):
    data = read_json(path)
    if not isinstance(data, dict) or data.get("version") != 1:
        raise Refusal("%s: 'version' must be 1" % path)
    cases = data.get("cases")
    if not isinstance(cases, list) or not cases:
        raise Refusal("%s: 'cases' must be a non-empty list" % path)
    seen = set()
    for at, case in enumerate(cases):
        if not isinstance(case, dict):
            raise Refusal("case %d is not an object" % at)
        for key in CASE_KEYS:
            if key not in case:
                raise Refusal("case %d is missing %r" % (at, key))
        for key in ("id", "lesson", "report", "check"):
            if not isinstance(case[key], str) or not case[key].strip():
                raise Refusal("case %d: %r must be a non-empty string" % (at, key))
        if case["id"] in seen:
            raise Refusal("duplicate case id %r" % case["id"])
        seen.add(case["id"])
        if case["check"] not in CHECKS:
            raise Refusal("case %r names an unknown check %r" % (case["id"], case["check"]))
        targets = case["targets"]
        if not isinstance(targets, dict) or not targets:
            raise Refusal("case %r: 'targets' must be a non-empty object" % case["id"])
        for name, target in targets.items():
            if not isinstance(target, dict) or "input" not in target or "expect" not in target:
                raise Refusal("case %r target %r needs 'input' and 'expect'" % (case["id"], name))
            if target["expect"] not in EXPECTS:
                raise Refusal(
                    "case %r target %r: 'expect' must be one of %s" % (case["id"], name, "/".join(EXPECTS))
                )
    return cases


def status_of(expect, verdict):
    if expect == "open":
        return "open" if verdict == "flag" else "now_clean"
    if verdict == expect:
        return "ok"
    return "regressed" if expect == "clean" else "unexpected"


def replay(root, cases_path):
    cases = load_cases(cases_path)
    report_cases, ok = [], True
    for case in cases:
        func, kind = CHECKS[case["check"]]
        rows = []
        for name, target in case["targets"].items():
            path = inside(root, target["input"], "case %r target %r input" % (case["id"], name))
            if not (path.is_dir() if kind == "dir" else path.is_file()):
                raise Refusal("case %r target %r: input %s is not an existing %s" % (case["id"], name, target["input"], kind))
            verdict, detail = func(root, path)
            status = status_of(target["expect"], verdict)
            ok = ok and status in ("ok", "open", "now_clean")
            rows.append(
                {
                    "target": name,
                    "check": case["check"],
                    "expect": target["expect"],
                    "verdict": verdict,
                    "status": status,
                    "detail": detail,
                }
            )
        report_cases.append({"id": case["id"], "lesson": case["lesson"], "targets": rows})
    return {"ok": ok, "cases": report_cases}


def main(argv):
    parser = argparse.ArgumentParser(prog="maintainer_cases.py", description=__doc__.split("\n")[0])
    sub = parser.add_subparsers(dest="command", required=True)
    one = sub.add_parser("replay", help="replay every target of every case")
    one.add_argument("--root", default=str(TOOL_ROOT), help="directory that the case inputs are relative to")
    one.add_argument("--cases", default=None, help="case file (default: tools/improve/maintainer-cases.json under the root)")
    args = parser.parse_args(argv)
    root = Path(args.root)
    cases = Path(args.cases) if args.cases else root / "tools" / "improve" / "maintainer-cases.json"
    try:
        report = replay(root, cases)
    except Refusal as exc:
        sys.stderr.write(PREFIX + str(exc) + "\n")
        return 2
    sys.stdout.write(json.dumps(report, indent=2) + "\n")
    return 0 if report["ok"] else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
