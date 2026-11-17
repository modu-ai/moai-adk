#!/usr/bin/env python3
"""Git merge driver for `skills/moai/references/source-manifest.json` (card C-74).

`.gitattributes` names the driver `moai-manifest` for the manifest. Register it once per clone
(this writes only the repository's own `.git/config`, shared by its worktrees):

  python3 -I tools/merge_driver_manifest.py --install   # idempotent
  python3 -I tools/merge_driver_manifest.py --check     # exit 0 when registered as below

Without the registration git merges the manifest as text, as before. The registered command
runs this file from the checkout being merged; where the checkout has no such file (an older
branch) or no `python3`, it runs git's own text merge instead.

Git calls the driver as `%O %A %B %P` before it writes any merged file to the working tree
(measured with git 2.54 for merge, rebase, cherry-pick and merge-tree: the working tree and
the index still hold our side, and drivers run one path at a time in reverse path order), so
the driver never reads the package from disk. It merges the three manifests like
`source_manifest.py write` does (each `origin` three-way) and takes each sha256 three-way:
equal sides, or the side that changed from the common version. For a file both sides changed,
it finds the three versions of that file by their sha256 among the file's past versions
(`git log --all`), merges them with `git merge-file --diff-algorithm=histogram` (the algorithm
of git's own `ort` merge) and records the hash of the clean result.

Anything else is left as a normal conflict (git's own text merge of the manifest, exit 1),
and a message names the reason: a skill file with a real conflict, a version not found, a file
deleted on one side and changed on the other, an `origin` changed on both sides, a skill file
with its own `merge` attribute. Then resolve the skill files and run
`python3 -I tools/source_manifest.py write`. The driver trusts each side's sha256 values, so a
side committed without `write` gives a wrong hash; `source_manifest.py check` (CI) fails on it.
"""
import argparse
import hashlib
import importlib.util
import json
import os
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
NAME = "moai-manifest"
SCRIPT = "tools/merge_driver_manifest.py"
COMMAND = (
    f"if test -f {SCRIPT} && command -v python3 >/dev/null 2>&1; "
    f"then python3 -I {SCRIPT} %O %A %B %P; "
    "else git merge-file -L ours -L base -L theirs %A %O %B; fi"
)
# Pseudo-refs that may name a side of the merge in progress, besides every ref (`--all`).
HEADS = ("HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD")


def load_source_manifest():
    """`tools/source_manifest.py` by its path: `-I` keeps this folder off `sys.path`. No
    `__pycache__` is written into the checkout being merged."""
    sys.dont_write_bytecode = True
    spec =importlib.util.spec_from_file_location("source_manifest", HERE / "source_manifest.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


sm = load_source_manifest()


class Unresolved(Exception):
    """The driver will not decide; git's text merge of the manifest stands."""


def git(*args, input=None):
    out = subprocess.run(["git", *args], input=input, capture_output=True)
    if out.returncode != 0:
        raise Unresolved(f"git {args[0]} failed: {out.stderr.decode(errors='replace').strip()}")
    return out.stdout


def merge_files(base, ours, theirs):
    """`(files, todo)`: the `files` map merged three-way, and `(rel, base, ours, theirs)`
    sha256 values of each file both sides changed, whose sha256 is still `None`."""
    bf, of, tf = base["files"], ours["files"], theirs["files"]
    files, todo = {}, []
    for rel in list(of) + [r for r in tf if r not in of]:
        b = bf.get(rel, sm.MISSING)
        o, t = of.get(rel, sm.MISSING), tf.get(rel, sm.MISSING)
        if o is sm.MISSING or t is sm.MISSING:
            side = t if o is sm.MISSING else o
            if b is sm.MISSING:
                files[rel] = side  # added on one side
            elif side != b:
                raise Unresolved(f"`{rel}` was deleted or renamed on one side and changed on the other")
            continue  # deleted on one side, unchanged on the other
        entry = sm.merge_entry(rel, None if b is sm.MISSING else b, o, t)
        bs = None if b is sm.MISSING else b.get("sha256")
        os_, ts = o.get("sha256"), t.get("sha256")
        if os_ == ts:
            entry["sha256"] = os_
        elif bs is not None and os_ == bs:
            entry["sha256"] = ts
        elif bs is not None and ts == bs:
            entry["sha256"] = os_
        elif bs is None:
            raise Unresolved(f"`{rel}` was added on both sides with different contents")
        else:
            entry["sha256"] = None
            todo.append((rel, bs, os_, ts))
        files[rel] = entry
    return files, todo


def candidate_revs():
    revs = ["--all"]
    names = list(HEADS) + [k[len("GITHEAD_"):] for k in os.environ if k.startswith("GITHEAD_")]
    for name in names:
        if not re.fullmatch(r"[A-Z_]+|[0-9a-f]{40}|[0-9a-f]{64}", name):
            continue
        out = subprocess.run(["git", "rev-parse", "--verify", "-q", f"{name}^{{commit}}"],
                             capture_output=True, text=True)
        if out.returncode == 0:
            revs.append(out.stdout.strip())
    return revs


def blobs_by_sha256(paths):
    """`{sha256: blob id}` of every past version of `paths` reachable from any ref or head."""
    raw = git("log", *candidate_revs(), "--diff-merges=separate", "--format=", "--raw",
              "--no-abbrev", "--no-renames", "--",
              *[f":(top,literal){p}" for p in paths]).decode()
    oids = set()
    for line in raw.splitlines():
        fields = line.split("\t", 1)[0].split()
        if len(fields) >= 4 and line.startswith(":"):
            oids.update(o for o in fields[2:4] if o.strip("0"))
    if not oids:
        return {}
    order = sorted(oids)
    data = git("cat-file", "--batch", input="".join(o + "\n" for o in order).encode())
    found, pos = {}, 0
    for _ in order:
        header_end = data.index(b"\n", pos)
        oid, kind, size = data[pos:header_end].decode().split()
        body = data[header_end + 1:header_end + 1 + int(size)]
        pos = header_end + 1 + int(size) + 1
        if kind == "blob":
            found[hashlib.sha256(body).hexdigest()] = oid
    return found


def fill_merged_hashes(package, todo):
    """Merge each file both sides changed from its three versions; set the clean result's
    sha256 or raise `Unresolved`."""
    paths = [f"{package}/{rel}" for rel, *_ in todo]
    top = git("rev-parse", "--show-toplevel").decode().strip()
    attrs = git("-C", top, "check-attr", "merge", "--", *paths).decode()
    for line in attrs.splitlines():
        if not line.endswith(": merge: unspecified"):
            raise Unresolved(f"a skill file has its own merge attribute: {line}")
    blobs = blobs_by_sha256(paths)
    shas = {}
    for rel, bs, os_, ts in todo:
        oids = [blobs.get(s) for s in (os_, bs, ts)]
        if None in oids:
            raise Unresolved(f"a version of `{rel}` named by the manifests is not in the history")
        out = subprocess.run(["git", "merge-file", "-p", "--object-id",
                              "--diff-algorithm=histogram", *oids], capture_output=True)
        if out.returncode != 0:
            raise Unresolved(f"`{rel}` does not merge cleanly")
        shas[rel] = hashlib.sha256(out.stdout).hexdigest()
    return shas


def merged_manifest(base_text, ours_text, theirs_text, path):
    """The merged manifest text, or `Unresolved`."""
    if not path.endswith("/" + sm.MANIFEST):
        raise Unresolved(f"{path} is not a package manifest")
    package = path[: -len("/" + sm.MANIFEST)]
    try:
        data = sm.joined(None, stages=(base_text, ours_text, theirs_text))
        sides = [sm.parse(t, label) for t, label in
                 ((base_text, "the common version"), (ours_text, "our side"),
                  (theirs_text, "their side"))]
        files, todo = merge_files(*sides)
    except sm.NeedsPerson as e:
        raise Unresolved(str(e)) from None
    if todo:
        for rel, sha in fill_merged_hashes(package, todo).items():
            files[rel]["sha256"] = sha
    data["files"] = files
    return sm.render(data)


def drive(base, ours, theirs, path):
    """The driver: write the merged manifest over `ours` and return 0, or leave git's text
    merge there and return 1."""
    try:
        texts = [Path(p).read_text(encoding="utf-8") for p in (base, ours, theirs)]
        merged = merged_manifest(*texts, path)
    except Exception as e:  # any surprise leaves git's own conflict, never our side alone
        print(f"{NAME}: {path} left as a conflict: {e}. Resolve the skill files, then run "
              f"`python3 -I tools/source_manifest.py write`.", file=sys.stderr)
        out = subprocess.run(["git", "merge-file", "-L", "ours", "-L", "base", "-L", "theirs",
                              ours, base, theirs])
        return 0 if out.returncode == 0 else 1
    with open(ours, "w", encoding="utf-8", newline="") as f:
        f.write(merged)
    return 0


def registered(root):
    out = subprocess.run(["git", "-C", root, "config", "--local", "--get-all",
                          f"merge.{NAME}.driver"], capture_output=True, text=True)
    values = out.stdout.splitlines()
    return values == [COMMAND], values


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("--install", action="store_true",
                    help="register the driver in this repository's .git/config")
    ap.add_argument("--check", action="store_true", help="report whether it is registered")
    ap.add_argument("files", nargs="*", metavar="PATH",
                    help="the four paths git passes when it runs the driver (%%O %%A %%B %%P)")
    ap.add_argument("--root", default=str(HERE.parent),
                    help="repository root (default: the one holding this tool)")
    args = ap.parse_args(argv)
    if args.files or not (args.install or args.check):
        if len(args.files) != 4 or args.install or args.check:
            ap.error("give --install, --check, or exactly the four driver paths %O %A %B %P")
        return drive(*args.files)
    if args.install:
        out = subprocess.run(["git", "-C", args.root, "config", "--local", "--replace-all",
                              f"merge.{NAME}.driver", COMMAND], capture_output=True, text=True)
        if out.returncode != 0:
            print(json.dumps({"ok": False, "error": out.stderr.strip()}))
            return 2
    ok, values = registered(args.root)
    print(json.dumps({"ok": ok, "registered": ok, "driver": values, "expected": COMMAND}))
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
