#!/usr/bin/env python3
"""Fail when the text MoAI ships gains one more maintainer-only token (SPEC-20261121-shipped-text-ratchet).

The shipped text (`skills/moai/**`, `agents/*.md`, `src/help.rs`) is read by every user, so it
should not carry what only this repository's maintainers need: card numbers, issue numbers,
links to this repository's tracker, the name of the maintainers' private rules file. It already
holds many, so "none allowed" would fail at once. This is a ratchet: the count may fall, it may
not rise. The counts a branch may not exceed are in a checked-in baseline.

  python3 -I tools/shipped_text_ratchet.py check      # exit 0 pass, 1 a count rose, 2 refused
  python3 -I tools/shipped_text_ratchet.py baseline   # print the current counts as a baseline

Options (both commands): `--root DIR` (default: the repository this file is in), `--tokens FILE`
(default `tools/improve/shipped-text-tokens.json`: the shipped path globs and the token
patterns). `check` also takes `--baseline FILE` (default `tools/improve/shipped-text-baseline.json`).
`baseline` prints the baseline file's exact shape on stdout; a person redirects it into the file.
Neither command writes anything.

`check` fails (exit 1) for two rules and names each break:
  - a token's total is above its baseline total;
  - a file holds a token more often than the baseline records for that file (a file or a
    token the baseline leaves out counts as 0). This catches a token that moves to another
    file while the total stays equal.
A count below its baseline is not a failure: it is listed under `headroom`, so the next change
can lower the baseline. A shipped file that is not valid UTF-8 is listed under `skipped` and
not counted.

Refusals (exit 2, stdout empty, one line on stderr starting `shipped_text_ratchet: `): a missing
or unreadable file, JSON that does not parse, a shape other than the documented one, a version
other than 1, an invalid pattern, a pattern that can match the empty string, a baseline that
names a token the token file lacks, a duplicate token id.

Token file shape: {"version": 1, "paths": [glob...], "tokens": [{"id", "pattern", "why"?}...]}.
A glob is relative to the root; `*` stays inside one folder, `**` crosses folders.
Baseline shape:   {"version": 1, "tokens": {id: total}, "files": {path: {id: count}}}.

Written for Python 3.9 and the standard library only: the first `python3` on a maintainer's PATH
can be the system 3.9.
"""

import argparse
import json
import os
import re
import sys
from pathlib import Path

PREFIX = "shipped_text_ratchet: "
REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_TOKENS = REPO_ROOT / "tools" / "improve" / "shipped-text-tokens.json"
DEFAULT_BASELINE = REPO_ROOT / "tools" / "improve" / "shipped-text-baseline.json"
VERSION = 1
# A pattern is refused when it matches the empty string anywhere in these probes; a match that
# turns out empty while a file is read is refused too.
EMPTY_PROBES = ("", "a", " ", "0", "-", "#", ".", "\n", "a b", "C-1 x")


class Refusal(Exception):
    """The input is not usable; the message is the one line the command prints."""


def refuse(message):
    raise Refusal(message)


def load_json(path, what):
    try:
        text = Path(path).read_text(encoding="utf-8")
    except OSError as err:
        refuse("%s %s cannot be read: %s" % (what, path, err.strerror or err))
    except UnicodeDecodeError:
        refuse("%s %s is not UTF-8" % (what, path))
    try:
        return json.loads(text)
    except ValueError as err:
        refuse("%s %s is not JSON: %s" % (what, path, err))


def is_count(value):
    return isinstance(value, int) and not isinstance(value, bool) and value >= 0


def check_keys(data, expected, what):
    if not isinstance(data, dict):
        refuse("%s must be a JSON object" % what)
    unknown = sorted(set(data) - set(expected))
    missing = sorted(set(expected) - set(data))
    if unknown or missing:
        refuse("%s has unknown keys %s and lacks keys %s" % (what, unknown, missing))
    if data["version"] != VERSION or isinstance(data["version"], bool):
        refuse("%s has version %r, expected %d" % (what, data["version"], VERSION))


def load_tokens(path):
    data = load_json(path, "token file")
    check_keys(data, ("version", "paths", "tokens"), "token file")
    paths = data["paths"]
    if not (isinstance(paths, list) and paths and all(isinstance(p, str) and p for p in paths)):
        refuse("token file: paths must be a non-empty list of non-empty strings")
    tokens = data["tokens"]
    if not (isinstance(tokens, list) and tokens):
        refuse("token file: tokens must be a non-empty list")
    compiled = []
    seen = set()
    for entry in tokens:
        if not isinstance(entry, dict) or not {"id", "pattern"} <= set(entry):
            refuse("token file: every token needs an id and a pattern")
        if set(entry) - {"id", "pattern", "why"}:
            refuse("token file: token %r has unknown keys" % (entry.get("id"),))
        token_id, pattern = entry["id"], entry["pattern"]
        if not (isinstance(token_id, str) and token_id and isinstance(pattern, str) and pattern):
            refuse("token file: token %r needs a non-empty id and pattern" % (token_id,))
        if token_id in seen:
            refuse("token file: token id %s appears twice" % token_id)
        seen.add(token_id)
        try:
            rx = re.compile(pattern)
        except re.error as err:
            refuse("token %s: invalid pattern: %s" % (token_id, err))
        for probe in EMPTY_PROBES:
            found = rx.search(probe)
            if found and found.end() == found.start():
                refuse("token %s: the pattern can match the empty string" % token_id)
        compiled.append((token_id, rx))
    return paths, compiled


def load_baseline(path, token_ids):
    data = load_json(path, "baseline")
    check_keys(data, ("version", "tokens", "files"), "baseline")
    totals, files = data["tokens"], data["files"]
    if not isinstance(totals, dict) or not isinstance(files, dict):
        refuse("baseline: tokens and files must be JSON objects")
    for token_id, count in totals.items():
        if token_id not in token_ids:
            refuse("baseline names the token %s, which the token file lacks" % token_id)
        if not is_count(count):
            refuse("baseline: the total of %s is not a non-negative integer" % token_id)
    for rel, row in files.items():
        if not isinstance(row, dict):
            refuse("baseline: the entry of %s must be an object" % rel)
        for token_id, count in row.items():
            if token_id not in token_ids:
                refuse("baseline names the token %s in %s, which the token file lacks" % (token_id, rel))
            if not is_count(count):
                refuse("baseline: %s in %s is not a non-negative integer" % (token_id, rel))
    return totals, files


def glob_regex(glob):
    out = []
    i = 0
    while i < len(glob):
        if glob.startswith("**", i):
            out.append(".*")
            i += 2
        elif glob[i] == "*":
            out.append("[^/]*")
            i += 1
        elif glob[i] == "?":
            out.append("[^/]")
            i += 1
        else:
            out.append(re.escape(glob[i]))
            i += 1
    return re.compile("".join(out))


def literal_prefix(glob):
    parts = []
    for part in glob.split("/"):
        if any(ch in part for ch in "*?"):
            break
        parts.append(part)
    return "/".join(parts)


def shipped_files(root, globs):
    found = set()
    for glob in globs:
        rx = glob_regex(glob)
        start = Path(root) / literal_prefix(glob)
        if start.is_file() and not start.is_symlink():
            rel = start.relative_to(root).as_posix()
            if rx.fullmatch(rel):
                found.add(rel)
            continue
        if not start.is_dir():
            continue
        for dirpath, dirnames, filenames in os.walk(str(start)):
            dirnames[:] = sorted(d for d in dirnames if d != ".git")
            for name in filenames:
                path = Path(dirpath) / name
                if path.is_symlink() or not path.is_file():
                    continue
                rel = path.relative_to(root).as_posix()
                if rx.fullmatch(rel):
                    found.add(rel)
    return sorted(found)


def count_tokens(root, globs, tokens):
    totals = {token_id: 0 for token_id, _ in tokens}
    per_file = {}
    skipped = []
    for rel in shipped_files(root, globs):
        try:
            text = (Path(root) / rel).read_bytes().decode("utf-8")
        except UnicodeDecodeError:
            skipped.append(rel)
            continue
        except OSError as err:
            refuse("%s cannot be read: %s" % (rel, err.strerror or err))
        row = {}
        for token_id, rx in tokens:
            n = 0
            for found in rx.finditer(text):
                if found.end() == found.start():
                    refuse("token %s: an empty match in %s" % (token_id, rel))
                n += 1
            if n:
                row[token_id] = n
                totals[token_id] += n
        if row:
            per_file[rel] = row
    return totals, per_file, skipped


def compare(tokens, totals, per_file, base_totals, base_files):
    failures, headroom = [], []
    for token_id, _ in tokens:
        count, base = totals[token_id], base_totals.get(token_id, 0)
        row = {"kind": "total", "token": token_id, "count": count, "baseline": base}
        if count > base:
            failures.append(row)
        elif count < base:
            headroom.append(row)
    for rel in sorted(set(per_file) | set(base_files)):
        for token_id, _ in tokens:
            count = per_file.get(rel, {}).get(token_id, 0)
            base = base_files.get(rel, {}).get(token_id, 0)
            row = {"kind": "file", "file": rel, "token": token_id, "count": count, "baseline": base}
            if count > base:
                failures.append(row)
            elif count < base:
                headroom.append(row)
    return failures, headroom


def emit(data):
    sys.stdout.write(json.dumps(data, indent=2, sort_keys=True, ensure_ascii=False) + "\n")


def run(args):
    root = Path(args.root)
    if not root.is_dir():
        refuse("root %s is not a folder" % root)
    globs, tokens = load_tokens(args.tokens)
    if args.command == "check":
        base_totals, base_files = load_baseline(args.baseline, {t for t, _ in tokens})
    totals, per_file, skipped = count_tokens(root, globs, tokens)
    if args.command == "baseline":
        emit({"version": VERSION, "tokens": totals, "files": per_file})
        return 0
    failures, headroom = compare(tokens, totals, per_file, base_totals, base_files)
    emit({"ok": not failures, "failures": failures, "headroom": headroom, "skipped": skipped})
    return 1 if failures else 0


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    sub = parser.add_subparsers(dest="command", required=True)
    for name in ("check", "baseline"):
        one = sub.add_parser(name)
        one.add_argument("--root", default=str(REPO_ROOT))
        one.add_argument("--tokens", default=str(DEFAULT_TOKENS))
        if name == "check":
            one.add_argument("--baseline", default=str(DEFAULT_BASELINE))
    args = parser.parse_args(argv)
    try:
        return run(args)
    except Refusal as err:
        sys.stderr.write(PREFIX + str(err) + "\n")
        return 2


if __name__ == "__main__":
    sys.exit(main())
