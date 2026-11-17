#!/usr/bin/env python3
"""Git merge driver for `skills/moai/references/source-manifest.json` (card C-74).

`.gitattributes` names the driver `moai-manifest` for the manifest. Register it once per clone
(this writes only the repository's own git folder, shared by its worktrees):

  python3 -I tools/merge_driver_manifest.py --install   # idempotent
  python3 -I tools/merge_driver_manifest.py --check     # exit 0 when registered and current

`--install` copies this file and `source_manifest.py` into `<git common dir>/moai-manifest-driver/`
and registers that copy by its absolute path, so a merge or rebase never runs the version of
the script on the branch being merged. When `tools/` changes, `--check` reports `stale` until
`--install` runs again. Without the registration git merges the manifest as text, as before;
when the copy or `python3` is missing, the registered command runs git's own text merge.

Git calls the driver as `%O %A %B %P` before it writes any merged file to the working tree
(measured with git 2.54 for merge, rebase, cherry-pick and merge-tree: the working tree and
the index still hold our side, and drivers run one path at a time in reverse path order), so
the driver never reads the package from disk. It merges the three manifests like
`source_manifest.py write` does (each `origin` three-way) and takes each sha256 three-way:
equal sides, or the side that changed from the common version. For a file both sides changed,
it finds the three versions of that file by their sha256 among the file's past versions
(`git log --all`), merges them with `git merge-file --diff-algorithm=histogram` (the algorithm
of git's own `ort` merge) and records the hash of the clean result.

Anything else leaves the manifest unmerged with git's own text merge in it (exit 1, even when
that text merge is clean), and a message names the reason: a skill file with a real conflict, a version not found, a file
deleted on one side and changed on the other, a folder that disappears on one side while the
other side adds or changes files in it (git may move them by directory-rename detection), an
`origin` changed on both sides, a skill file with its own `merge` attribute, a `diff.algorithm`
other than histogram or `merge.renormalize` in the git config the driver sees. Then resolve the
skill files and run `python3 -I tools/source_manifest.py write`.

Limits, with `source_manifest.py check` (CI) as the safety net: `-X` options given to
`git merge` (`-Xignore-space-change`, `-Xdiff-algorithm=...`) never reach a merge driver, and
the driver trusts each side's sha256 values, so a side committed without `write` gives a wrong
hash. `check` fails on any manifest that does not match the files.
"""
import argparse
import hashlib
import importlib.util
import json
import os
import re
import shlex
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
NAME = "moai-manifest"
COPY_DIR = "moai-manifest-driver"
COPIED = ("merge_driver_manifest.py", "source_manifest.py")
# Pseudo-refs that may name a side of the merge in progress, besides every ref (`--all`).
HEADS = ("HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD")
# Git settings that change how git merges a file in ways `git merge-file` here cannot repeat.
# `-X` options never reach a driver; `check` in CI is the net for those.
ALGORITHMS_OK = ("", "histogram")

sm = None  # tools/source_manifest.py, loaded inside drive() so a failure still leaves markers


def load_source_manifest():
    """`source_manifest.py` next to this file, by its path: `-I` keeps this folder off
    `sys.path`. No `__pycache__` is written next to it."""
    sys.dont_write_bytecode = True
    spec = importlib.util.spec_from_file_location("source_manifest", HERE / "source_manifest.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def command(copy):
    """The registered driver command: the copy by its absolute path, else git's text merge."""
    script = shlex.quote(str(copy / "merge_driver_manifest.py"))
    return (f"if test -f {script} && command -v python3 >/dev/null 2>&1; "
            f"then python3 -I {script} %O %A %B %P; "
            "else git merge-file -L ours -L base -L theirs %A %O %B; fi")


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
    refuse_folder_moves(bf, of, tf)
    return files, todo


def folders(paths):
    return {rel.rsplit("/", i)[0] for rel in paths for i in range(1, rel.count("/") + 1)}


def refuse_folder_moves(bf, of, tf):
    """A folder that disappears on one side while the other side adds or changes a file in it:
    git's directory-rename detection may move that file, which the manifests cannot show."""
    before = folders(bf)
    for side, other in ((of, tf), (tf, of)):
        gone = before - folders(side)
        for rel, entry in other.items():
            if bf.get(rel, sm.MISSING) == entry:
                continue
            for folder in gone:
                if rel.startswith(folder + "/"):
                    raise Unresolved(f"folder `{folder}` disappears on one side while the other "
                                     f"side adds or changes `{rel}` in it")


def refuse_settings_git_merges_differently():
    """Git settings the driver can see that make git's own file merge differ from
    `git merge-file --diff-algorithm=histogram`; `-X` options are invisible to a driver."""
    algorithm = subprocess.run(["git", "config", "--get", "diff.algorithm"],
                               capture_output=True, text=True).stdout.strip().lower()
    if algorithm not in ALGORITHMS_OK:
        raise Unresolved(f"diff.algorithm is {algorithm!r}; the driver repeats only histogram")
    renormalize = subprocess.run(["git", "config", "--type=bool", "--get", "merge.renormalize"],
                                 capture_output=True, text=True).stdout.strip()
    if renormalize == "true":
        raise Unresolved("merge.renormalize is on; the driver cannot repeat renormalized merges")


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
    refuse_settings_git_merges_differently()
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
    global sm
    try:
        if sm is None:
            sm = load_source_manifest()
        texts =[Path(p).read_text(encoding="utf-8") for p in (base, ours, theirs)]
        merged = merged_manifest(*texts, path)
    except Exception as e:  # any surprise leaves git's own conflict, never our side alone
        print(f"{NAME}: {path} left as a conflict: {e}. Resolve the skill files, then run "
              f"`python3 -I tools/source_manifest.py write`.", file=sys.stderr)
        # Git's text merge goes into the file, but the manifest stays unmerged even when that
        # text merge is clean: the driver found a reason it may be wrong.
        subprocess.run(["git", "merge-file", "-L", "ours", "-L", "base", "-L", "theirs",
                        ours, base, theirs])
        return 1
    with open(ours, "w", encoding="utf-8", newline="") as f:
        f.write(merged)
    return 0


def copy_dir(root):
    """`<git common dir>/moai-manifest-driver`: shared by every worktree, in no checkout."""
    out = subprocess.run(["git", "-C", root, "rev-parse", "--path-format=absolute",
                          "--git-common-dir"], capture_output=True, text=True)
    if out.returncode != 0:
        raise RuntimeError(out.stderr.strip() or "not a git repository")
    return Path(out.stdout.strip()) / COPY_DIR


def install(root):
    copy = copy_dir(root)
    if copy.is_symlink():
        raise RuntimeError(f"{copy} is a link")
    copy.mkdir(exist_ok=True)
    for name in COPIED:
        fd, tmp = tempfile.mkstemp(dir=copy, prefix=f".{name}.", suffix=".tmp")
        try:
            with os.fdopen(fd, "wb") as f:
                f.write((HERE / name).read_bytes())
            os.chmod(tmp, 0o644)
            os.replace(tmp, copy / name)
        except BaseException:
            Path(tmp).unlink(missing_ok=True)
            raise
    out = subprocess.run(["git", "-C", root, "config", "--local", "--replace-all",
                          f"merge.{NAME}.driver", command(copy)], capture_output=True, text=True)
    if out.returncode != 0:
        raise RuntimeError(out.stderr.strip())


def status(root):
    copy = copy_dir(root)
    expected = command(copy)
    values = subprocess.run(["git", "-C", root, "config", "--local", "--get-all",
                             f"merge.{NAME}.driver"], capture_output=True, text=True).stdout
    values = values.splitlines()
    report = {"registered": values == [expected], "driver": values, "expected": expected}
    if not values:
        report["status"], report["message"] = "missing", "not registered: run --install"
    elif not report["registered"]:
        report["status"] = "other"
        report["message"] = "another command is registered: run --install to replace it"
    elif any(not (copy / n).is_file() or (copy / n).read_bytes() != (HERE / n).read_bytes()
             for n in COPIED):
        report["status"] = "stale"
        report["message"] = f"the copy in {copy} differs from tools/: run --install again"
    else:
        report["status"], report["message"] = "ok", "registered and current"
    report["ok"] = report["status"] == "ok"
    return report


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("--install", action="store_true",
                    help="copy the driver into the git folder and register that copy")
    ap.add_argument("--check", action="store_true",
                    help="report whether it is registered and the copy is current")
    ap.add_argument("files", nargs="*", metavar="PATH",
                    help="the four paths git passes when it runs the driver (%%O %%A %%B %%P)")
    ap.add_argument("--root", default=str(HERE.parent),
                    help="repository root (default: the one holding this tool)")
    args = ap.parse_args(argv)
    if args.files or not (args.install or args.check):
        if len(args.files) != 4 or args.install or args.check:
            ap.error("give --install, --check, or exactly the four driver paths %O %A %B %P")
        return drive(*args.files)
    try:
        if args.install:
            install(args.root)
        report = status(args.root)
    except (OSError, RuntimeError) as e:
        print(json.dumps({"ok": False, "error": str(e)}))
        return 2
    print(json.dumps(report))
    return 0 if report["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
