#!/usr/bin/env python3
"""Entry count of `<HOME>/.moai` (Issue #3068 P2, AC8: the entry count of `~/.moai` is the same
before and after a verification run, and a read-only command creates no folder).

Usage:
  python3 -I tools/home_entry_count.py --home ABSOLUTE_HOME_PATH
  python3 -I tools/home_entry_count.py --home ABSOLUTE_HOME_PATH --compare BEFORE.json

Without `--compare` it prints one JSON object to stdout and exits 0:
  {"home": ..., "moai": {"state": "dir"|"absent"|"file"|"link"|"other", "count": N,
   "names": [...]}, "depth": 1, "contents_read": false}
`names` are the entries directly inside `<HOME>/.moai`, sorted: depth 1, nothing below them is
listed, no file is opened or read, and a link at `<HOME>/.moai` is reported, not followed.

With `--compare BEFORE.json` (a file written from an earlier run of this tool for the same
HOME) it prints the same object plus a `compare` object (`unchanged`, `state`, `count`,
`added`, `removed`; the pairs are [before, after]) and exits 0 when the state, the count and
the names are all unchanged, 1 when any differs. The names are compared, not only the count: a
rename keeps the count and still exits 1. File contents and modification times are not
compared.

Exit 2, with a message on stderr and nothing on stdout, for a usage error, a `--home` that is
not an absolute path of an existing folder, an unreadable `<HOME>/.moai`, and a BEFORE file
that is missing, malformed or taken for another HOME. The HOME is never defaulted: this tool
never defaults to the real HOME, never reads `$HOME`, `~` or any other environment value, so
nobody runs it on the real HOME by accident.

Procedure for AC8's real-HOME half (not part of any test, not run by CI): after the user
confirms, the leader or operator runs it with the real HOME path before and after a
verification run, saving the first output, and `--compare`s the second run to it. Tests run it
on temporary HOME folders only (tests/home_entry_count_test.py).

Python standard library only.
"""

import json
import os
import stat
import sys

KNOWN = ("--home", "--compare")
DEPTH = 1


class Refused(Exception):
    """A usage error or an unusable input: exit 2."""


def parse_args(argv):
    opts = {}
    i = 0
    while i < len(argv):
        name = argv[i]
        if name not in KNOWN:
            raise Refused(f"unknown argument {name!r}; usage: --home PATH [--compare FILE]")
        if i + 1 >= len(argv):
            raise Refused(f"{name} needs a value")
        if name in opts:
            raise Refused(f"{name} given twice")
        opts[name] = argv[i + 1]
        i += 2
    if "--home" not in opts:
        raise Refused("--home PATH is required: the HOME is never defaulted or read from the "
                      "environment")
    return opts


def check_home(path):
    if not path or "\0" in path:
        raise Refused("--home is empty")
    if not os.path.isabs(path):
        raise Refused(f"--home must be an absolute path, got {path!r}")
    if not os.path.isdir(path):
        raise Refused(f"--home is not an existing folder: {path!r}")
    return os.path.normpath(path)


def measure(home):
    """The state, count and sorted names of the entries directly inside `<home>/.moai`."""
    moai = os.path.join(home, ".moai")
    try:
        mode = os.lstat(moai).st_mode
    except FileNotFoundError:
        return {"state": "absent", "count": 0, "names": []}
    except OSError as e:
        raise Refused(f"cannot inspect {moai!r}: {e.strerror}") from e
    if stat.S_ISLNK(mode):
        return {"state": "link", "count": 0, "names": []}
    if stat.S_ISREG(mode):
        return {"state": "file", "count": 0, "names": []}
    if not stat.S_ISDIR(mode):
        return {"state": "other", "count": 0, "names": []}
    try:
        names = sorted(os.listdir(moai))
    except OSError as e:
        raise Refused(f"cannot list {moai!r}: {e.strerror}") from e
    return {"state": "dir", "count": len(names), "names": names}


def report(home):
    return {"home": home, "moai": measure(home), "depth": DEPTH, "contents_read": False}


def read_before(path, home):
    try:
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
    except (OSError, ValueError) as e:
        raise Refused(f"cannot read the BEFORE file {path!r}: {e}") from e
    moai = data.get("moai") if isinstance(data, dict) else None
    ok = (isinstance(moai, dict) and isinstance(data.get("home"), str)
          and isinstance(moai.get("state"), str) and isinstance(moai.get("count"), int)
          and isinstance(moai.get("names"), list)
          and all(isinstance(n, str) for n in moai["names"])
          and moai["count"] == len(moai["names"]))
    if not ok:
        raise Refused(f"the BEFORE file {path!r} is not an output of this tool")
    if os.path.normpath(data["home"]) != home:
        raise Refused(f"the BEFORE file {path!r} was taken for another HOME")
    return moai


def compare(before, after):
    added = sorted(set(after["names"]) - set(before["names"]))
    removed = sorted(set(before["names"]) - set(after["names"]))
    unchanged = (before["state"] == after["state"] and before["count"] == after["count"]
                 and not added and not removed)
    return {"unchanged": unchanged, "state": [before["state"], after["state"]],
            "count": [before["count"], after["count"]], "added": added, "removed": removed}


def main(argv):
    try:
        opts = parse_args(argv)
        home = check_home(opts["--home"])
        before = read_before(opts["--compare"], home) if "--compare" in opts else None
        out = report(home)
    except Refused as e:
        sys.stderr.write(f"home_entry_count: {e}\n")
        return 2
    code = 0
    if before is not None:
        out["compare"] = compare(before, out["moai"])
        code = 0 if out["compare"]["unchanged"] else 1
    sys.stdout.write(json.dumps(out, sort_keys=True) + "\n")
    return code


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
