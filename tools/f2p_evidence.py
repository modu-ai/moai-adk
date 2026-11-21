#!/usr/bin/env python3
"""Fail-to-pass evidence for the merge train and the audits (SPEC-20261121-fail-to-pass-evidence).

A test that the head adds or changes should fail on the base before it passes on the head.
This tool shows the base side: it lists the Rust integration tests (`tests/*.rs`, top level) that
the head adds or changes against the base, runs them on the base source with the head's test
files, and prints one JSON line whose `summary` reads `f2p=<n> flagged=<m>`: `n` tests failed on
the base (a test file that does not build there counts, with the reason `build_error`), `m`
passed there and so prove nothing, unless `--reason` records why (a pure refactor has no test
to fail). Exit 0 when nothing is flagged, 1 when something is, 2 on a refusal (stderr starts
`f2p_evidence: `, stdout empty).

  python3 -I tools/f2p_evidence.py --dir . --base <rev> --head <rev> [--reason <text>]
  python3 -I tools/f2p_evidence.py --dir . --base <rev> --head <rev> --list-only

The run builds the base, which is heavy: call it under the heavy-slot command, and only in the
train and in audits. The base is extracted with `git archive` into a temporary folder that is
removed afterwards; the repository is only read. `--cargo <program>` replaces `cargo` (the tests
use a fake one) and `--target-dir <dir>` becomes `CARGO_TARGET_DIR` for the base build. Limits:
Python tests and unit tests inside `src/` are not looked at, a `#[test]` made by a macro is not
seen, and tests in sub-folders of `tests/` (modules of another file) are not listed. Written
for Python 3.9.
"""
import argparse
import io
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile

PREFIX = "f2p_evidence: "
TEST_FILE = re.compile(r"^tests/[^/]+\.rs$")
VERDICT = re.compile(r"^test (\S+) \.\.\. (ok|FAILED|ignored)\b")
CHAR_LITERAL = re.compile(r"'(?:\\[^']*|[^\\'])'")


class Refusal(Exception):
    pass


def refuse(message):
    raise Refusal(message)


def clean_env():
    return {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}


def git(directory, *args, binary=False):
    proc = subprocess.run(["git"] + list(args), cwd=directory, capture_output=True, env=clean_env())
    if proc.returncode != 0:
        refuse("git %s failed: %s" % (" ".join(args[:2]), proc.stderr.decode("utf-8", "replace").strip()))
    return proc.stdout if binary else proc.stdout.decode("utf-8", "replace")


def skip_string(text, i):
    """The index after the string literal that starts at text[i] (a quote or a raw string)."""
    if text[i] == "r":
        j = i + 1
        hashes = 0
        while j < len(text) and text[j] == "#":
            hashes += 1
            j += 1
        end = '"' + "#" * hashes
        k = text.find(end, j + 1)
        return len(text) if k < 0 else k + len(end)
    j = i + 1
    while j < len(text):
        if text[j] == "\\":
            j += 2
            continue
        if text[j] == '"':
            return j + 1
        j += 1
    return len(text)


def skip_atom(text, i):
    """If a string, char literal or comment starts at i, the index after it, else i."""
    c = text[i]
    if c == '"':
        return skip_string(text, i)
    if c == "r" and re.match(r'r#*"', text[i:i + 8]) and (i == 0 or not (text[i - 1].isalnum() or text[i - 1] == "_")):
        return skip_string(text, i)
    if c == "'":
        m = CHAR_LITERAL.match(text, i)
        return m.end() if m else i
    if text.startswith("//", i):
        k = text.find("\n", i)
        return len(text) if k < 0 else k
    if text.startswith("/*", i):
        k = text.find("*/", i + 2)
        return len(text) if k < 0 else k + 2
    return i


def matching_brace(text, start):
    """The index of the brace that closes the one at text[start], or -1."""
    depth = 0
    i = start
    while i < len(text):
        j = skip_atom(text, i)
        if j != i:
            i = j
            continue
        if text[i] == "{":
            depth += 1
        elif text[i] == "}":
            depth -= 1
            if depth == 0:
                return i
        i += 1
    return -1


def skip_attributes(text, i):
    """The index after any attribute lines `#[...]` and white space starting at i."""
    while True:
        while i < len(text) and text[i].isspace():
            i += 1
        if text.startswith("#[", i):
            depth = 0
            while i < len(text):
                j = skip_atom(text, i)
                if j != i:
                    i = j
                    continue
                if text[i] == "[":
                    depth += 1
                elif text[i] == "]":
                    depth -= 1
                    if depth == 0:
                        i += 1
                        break
                i += 1
            continue
        return i


def parse_tests(text):
    """name -> the normalized text of each `#[test]` function of a Rust file (attributes to the closing brace)."""
    found = {}
    for m in re.finditer(r"#\[test\]", text):
        i = skip_attributes(text, m.end())
        fn = re.compile(r"(?:pub\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+([A-Za-z_][A-Za-z0-9_]*)").match(text, i)
        if not fn:
            continue
        brace = text.find("{", fn.end())
        if brace < 0:
            continue
        close = matching_brace(text, brace)
        if close < 0:
            continue
        found[fn.group(1)] = " ".join(text[m.start():close + 1].split())
    return found


def show(directory, rev, path):
    proc = subprocess.run(["git", "show", "%s:%s" % (rev, path)], cwd=directory, capture_output=True, env=clean_env())
    return proc.stdout.decode("utf-8", "replace") if proc.returncode == 0 else None


def listed_tests(directory, base, head):
    changed = git(directory, "diff", "--name-only", "--no-renames", base, head, "--", "tests").split("\n")
    out = []
    for path in sorted(p for p in changed if TEST_FILE.match(p)):
        head_text = show(directory, head, path)
        if head_text is None:
            continue
        base_text = show(directory, base, path)
        before = parse_tests(base_text) if base_text is not None else {}
        for name, text in sorted(parse_tests(head_text).items()):
            if name not in before:
                out.append({"file": path, "name": name, "kind": "new"})
            elif before[name] != text:
                out.append({"file": path, "name": name, "kind": "changed"})
    return out


def validate(args):
    if not os.path.isdir(args.dir):
        refuse("%s is not a directory" % args.dir)
    for label, rev in (("base", args.base), ("head", args.head)):
        proc = subprocess.run(["git", "rev-parse", "--verify", "--quiet", rev + "^{commit}"],
                              cwd=args.dir, capture_output=True, env=clean_env())
        if proc.returncode != 0:
            refuse("unknown %s revision %r" % (label, rev))
    anc = subprocess.run(["git", "merge-base", "--is-ancestor", args.base, args.head],
                         cwd=args.dir, capture_output=True, env=clean_env())
    if anc.returncode != 0:
        refuse("the base %s is not an ancestor of the head %s" % (args.base, args.head))
    if args.reason is not None and not args.reason.strip():
        refuse("--reason must not be empty")
    if not args.list_only and shutil.which(args.cargo) is None:
        refuse("the cargo program %r was not found" % args.cargo)


def extract_base(directory, base, target):
    data = git(directory, "archive", "--format=tar", base, binary=True)
    with tarfile.open(fileobj=io.BytesIO(data)) as tar:
        for member in tar.getmembers():
            name = member.name
            if name.startswith("/") or ".." in name.split("/"):
                refuse("the archive holds an unsafe path %r" % name)
        tar.extractall(target)


def touch_all(root):
    for folder, _dirs, files in os.walk(root):
        for f in files:
            try:
                os.utime(os.path.join(folder, f), None)
            except OSError:
                pass


def run_file(args, root, stem, tests):
    """name -> (result, reason) for the tests of one test file, run on the extracted base."""
    argv = [shutil.which(args.cargo), "test", "--test", stem, "--", "--exact"] + [t["name"] for t in tests]
    env = dict(os.environ)
    if args.target_dir:
        env["CARGO_TARGET_DIR"] = os.path.abspath(args.target_dir)
    proc = subprocess.run(argv, cwd=root, capture_output=True, text=True, env=env)
    verdicts = {}
    for line in proc.stdout.splitlines():
        m = VERDICT.match(line)
        if m:
            verdicts[m.group(1)] = m.group(2)
    if not verdicts and proc.returncode != 0 and re.search(r"error(\[E\d+\])?:|could not compile", proc.stderr + proc.stdout):
        return {t["name"]: ("fail", "build_error") for t in tests}
    out = {}
    for t in tests:
        word = verdicts.get(t["name"])
        if word is None:
            word = next((v for k, v in verdicts.items() if k.endswith("::" + t["name"])), None)
        if word not in ("ok", "FAILED"):
            refuse("no verdict for %s in %s (cargo exit %d)" % (t["name"], t["file"], proc.returncode))
        out[t["name"]] = ("pass", None) if word == "ok" else ("fail", None)
    return out


def main(argv):
    ap = argparse.ArgumentParser(description="Fail-to-pass evidence: show the new and changed tests failing on the base.")
    ap.add_argument("--dir", required=True)
    ap.add_argument("--base", required=True)
    ap.add_argument("--head", required=True)
    ap.add_argument("--cargo", default="cargo")
    ap.add_argument("--target-dir")
    ap.add_argument("--reason")
    ap.add_argument("--list-only", action="store_true")
    args = ap.parse_args(argv)
    try:
        validate(args)
        tests = listed_tests(args.dir, args.base, args.head)
        report = {"base": args.base, "head": args.head}
        if args.list_only:
            report.update({"mode": "list", "tests": tests, "summary": "listed=%d" % len(tests)})
            sys.stdout.write(json.dumps(report, ensure_ascii=False) + "\n")
            return 0
        report["mode"] = "run"
        results = {}
        if tests:
            root = tempfile.mkdtemp(prefix="f2p-evidence-")
            try:
                extract_base(args.dir, args.base, root)
                by_file = {}
                for t in tests:
                    by_file.setdefault(t["file"], []).append(t)
                for path, group in sorted(by_file.items()):
                    dest = os.path.join(root, path)
                    os.makedirs(os.path.dirname(dest), exist_ok=True)
                    with open(dest, "w", encoding="utf-8") as f:
                        f.write(show(args.dir, args.head, path) or "")
                touch_all(root)
                for path, group in sorted(by_file.items()):
                    stem = os.path.basename(path)[:-3]
                    for name, verdict in run_file(args, root, stem, group).items():
                        results[(path, name)] = verdict
            finally:
                shutil.rmtree(root, ignore_errors=True)
        rows = []
        failed = passed = 0
        for t in tests:
            result, reason = results[(t["file"], t["name"])]
            row = dict(t, base=result)
            if reason:
                row["reason"] = reason
            rows.append(row)
            failed += result == "fail"
            passed += result == "pass"
        flagged = 0 if args.reason is not None else passed
        report["tests"] = rows
        report["summary"] = "f2p=%d flagged=%d" % (failed, flagged)
        if args.reason is not None:
            report["reason"] = args.reason.strip()
        sys.stdout.write(json.dumps(report, ensure_ascii=False) + "\n")
        return 1 if flagged else 0
    except Refusal as err:
        sys.stderr.write(PREFIX + str(err) + "\n")
        return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
