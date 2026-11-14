#!/usr/bin/env python3
"""Check or rebuild `skills/moai/references/source-manifest.json` (Issue #2837).

The manifest pairs every file under `skills/moai/` with its sha256 and a hand-written
`origin`. When two branches change the same skill file, its sha256 line conflicts on merge
even when the skill file itself merges cleanly (38 of 65 manifest conflicts since
2026-11-08). Pick neither side: resolve the skill files, then run `write`.

  python3 -I tools/source_manifest.py check   # exit 0 when every entry matches, 1 otherwise
  python3 -I tools/source_manifest.py write   # recompute the sha256 values in place

`write` also reads a manifest that still holds git conflict markers: each side is parsed
as JSON, the `files` maps are joined (our order first, then entries only the other side
has) and every sha256 is recomputed from the files on disk. It writes nothing and exits 2
when a person has to decide: an entry whose `origin` differs between the sides, a
top-level value other than `files` that differs, a skill file that still holds conflict
markers, a new file with no entry (its `origin` is the author's to write), a link anywhere from the repository root to the
manifest or in the package, or a file that is not a regular file (a FIFO would block).
Entries of deleted files are dropped and reported. The file is written in its existing
form, `json.dumps(indent=2, ensure_ascii=False)` plus a newline, and only when it changes.
"""
import argparse
import hashlib
import json
import os
import stat
import subprocess
import sys
import tempfile
from pathlib import Path

PACKAGE = Path("skills") / "moai"
MANIFEST = "references/source-manifest.json"
MARKERS = ("<<<<<<< ", "||||||| ", "=======", ">>>>>>> ")


class NeedsPerson(Exception):
    """A case the tool will not decide; nothing is written."""


def has_markers(text):
    return any(
        line.startswith(m) and (m != "=======" or line.rstrip("\r\n") == m)
        for line in text.splitlines()
        for m in MARKERS
    )


def sides(text):
    """`(ours, theirs)` of a text with git conflict markers (diff3 base sections dropped)."""
    ours, theirs, state = [], [], None
    for line in text.splitlines(keepends=True):
        bare = line.rstrip("\r\n")
        if line.startswith("<<<<<<< ") and state is None:
            state = "ours"
        elif line.startswith("||||||| ") and state == "ours":
            state = "base"
        elif bare == "=======" and state in ("ours", "base"):
            state = "theirs"
        elif line.startswith(">>>>>>> ") and state == "theirs":
            state = None
        elif state is None:
            ours.append(line)
            theirs.append(line)
        elif state == "ours":
            ours.append(line)
        elif state == "theirs":
            theirs.append(line)
    if state is not None:
        raise NeedsPerson("the manifest has an unclosed conflict block")
    return "".join(ours), "".join(theirs)


def parse(text, label):
    try:
        data = json.loads(text)
    except ValueError as e:
        raise NeedsPerson(f"{label} of the manifest is not JSON: {e}") from None
    if not isinstance(data, dict) or not isinstance(data.get("files"), dict):
        raise NeedsPerson(f"{label} of the manifest has no `files` object")
    return data


MISSING = object()


def pick(what, base, ours, theirs):
    """Three-way choice of one value: equal sides, or the side that changed from the base
    (`MISSING` stands for an absent key). Without a base only equal sides are accepted."""
    if ours == theirs:
        return ours
    if base is not None:
        if ours == base:
            return theirs
        if theirs == base:
            return ours
    raise NeedsPerson(f"{what} differs between the sides")


def merge_entry(rel, base, ours, theirs):
    """One `files` entry; `sha256` is recomputed later, so only `origin` (and any other
    key) is chosen. `base` is `None` when no common version is known."""
    if ours is MISSING or theirs is MISSING:
        # Added on one side, or deleted on one side: the file on disk decides later.
        return theirs if ours is MISSING else ours
    out = {}
    for key in list(ours) + [k for k in theirs if k not in ours]:
        if key == "sha256":
            out[key] = ours.get(key) or theirs.get(key)
            continue
        b = None if base is None else base.get(key, MISSING)
        o, t = ours.get(key, MISSING), theirs.get(key, MISSING)
        if all(isinstance(x, dict) for x in (o, t)) and (b is None or isinstance(b, dict)):
            # `origin`: one key at a time, so one side's `kind` and the other's `reason` join.
            value = {}
            for sub in list(o) + [k for k in t if k not in o]:
                bs = None if b is None else b.get(sub, MISSING)
                v = pick(f"`{rel}`: {key}.{sub}", bs, o.get(sub, MISSING), t.get(sub, MISSING))
                if v is not MISSING:
                    value[sub] = v
        else:
            value = pick(f"`{rel}`: {key}", b, o, t)
        if value is not MISSING:
            out[key] = value
    return out


def joined(text, stages=None):
    """The manifest as one JSON object. A conflicted one is joined from both sides, three-way
    against the common version when it is known (`stages` from the git index, or a diff3 base
    section)."""
    if stages is None and not has_markers(text):
        return parse(text, "the text")
    if stages is not None:
        base_text, ours_text, theirs_text = stages
    else:
        ours_text, theirs_text = sides(text)
        base_text = base_section(text)
    ours, theirs = parse(ours_text, "our side"), parse(theirs_text, "their side")
    base = parse(base_text, "the common version") if base_text is not None else None
    out = {}
    for key in list(ours) + [k for k in theirs if k not in ours]:
        if key == "files":
            continue
        b = None if base is None else base.get(key, MISSING)
        value = pick(f"top-level `{key}`", b, ours.get(key, MISSING), theirs.get(key, MISSING))
        if value is not MISSING:
            out[key] = value
    files = {}
    for rel in list(ours["files"]) + [r for r in theirs["files"] if r not in ours["files"]]:
        b = None if base is None else base["files"].get(rel)
        files[rel] = merge_entry(
            rel, b, ours["files"].get(rel, MISSING), theirs["files"].get(rel, MISSING)
        )
    # Keep the top-level key order of our side, `files` in its place.
    ordered = {}
    for key in list(ours) + [k for k in out if k not in ours]:
        if key == "files":
            ordered["files"] = files
        elif key in out:
            ordered[key] = out[key]
    return ordered


def base_section(text):
    """The common version of a diff3-style conflicted text, or `None` when a block has no
    `|||||||` section."""
    base, state, every_block_has_base = [], None, True
    for line in text.splitlines(keepends=True):
        bare = line.rstrip("\r\n")
        if line.startswith("<<<<<<< ") and state is None:
            state, seen_base = "ours", False
        elif line.startswith("||||||| ") and state == "ours":
            state, seen_base = "base", True
        elif bare == "=======" and state in ("ours", "base"):
            every_block_has_base &= seen_base
            state = "theirs"
        elif line.startswith(">>>>>>> ") and state == "theirs":
            state = None
        elif state is None or state == "base":
            base.append(line)
    return "".join(base) if every_block_has_base else None


def index_stages(root):
    """`(base, ours, theirs)` texts of the manifest from the git index while a merge has it
    unresolved; `None` when it is not unmerged (or this is not a git checkout)."""
    rel = (PACKAGE / MANIFEST).as_posix()
    try:
        out = subprocess.run(
            ["git", "-C", str(root), "ls-files", "-u", "-z", "--", rel],
            capture_output=True, check=True,
        ).stdout
    except (OSError, subprocess.CalledProcessError):
        return None
    blobs = {}
    for rec in out.split(b"\0"):
        if not rec:
            continue
        meta, _ = rec.split(b"\t", 1)
        _, oid, stage = meta.decode().split()
        blobs[int(stage)] = oid
    if 2 not in blobs or 3 not in blobs:
        return None

    def text(stage):
        if stage not in blobs:
            return None
        return subprocess.run(
            ["git", "-C", str(root), "cat-file", "blob", blobs[stage]],
            capture_output=True, check=True,
        ).stdout.decode("utf-8")

    return text(1), text(2), text(3)


def package_files(package):
    """Every file under the package except the manifest, as `/`-separated relative paths."""
    out = []
    for dirpath, dirnames, filenames in os.walk(package):
        base = Path(dirpath)
        for name in sorted(dirnames):
            if (base / name).is_symlink():
                raise NeedsPerson(f"{(base / name).relative_to(package).as_posix()} is a link")
        for name in filenames:
            path = base / name
            rel = path.relative_to(package).as_posix()
            mode = os.lstat(path).st_mode
            if stat.S_ISLNK(mode):
                raise NeedsPerson(f"{rel} is a link")
            # A FIFO or device would block or never end a read (#2849 audit B1).
            if not stat.S_ISREG(mode):
                raise NeedsPerson(f"{rel} is not a regular file")
            if rel != MANIFEST:
                out.append(rel)
    return sorted(out)


def guard(root):
    """Refuse a package path that is not plain: every folder from the repository root down to
    the manifest's must be a real folder and the manifest a regular file, never a link that
    would send a write outside the repository (#2849 audit B2)."""
    here = Path(root)
    parts = (PACKAGE / MANIFEST).parts
    for i, part in enumerate(parts):
        here = here / part
        rel = Path(*parts[: i + 1]).as_posix()
        try:
            mode = os.lstat(here).st_mode
        except FileNotFoundError:
            raise NeedsPerson(f"{rel} is missing") from None
        if stat.S_ISLNK(mode):
            raise NeedsPerson(f"{rel} is a link")
        last = i == len(parts) - 1
        if (last and not stat.S_ISREG(mode)) or (not last and not stat.S_ISDIR(mode)):
            raise NeedsPerson(f"{rel} is not a regular {'file' if last else 'folder'}")


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rebuild(root, stages=None):
    """`(manifest_object, report)` with every sha256 recomputed; raises `NeedsPerson`.
    `stages` defaults to the git index's unmerged versions of the manifest, if any."""
    guard(root)
    package = Path(root) / PACKAGE
    text = (package / MANIFEST).read_text(encoding="utf-8")
    if stages is None:
        stages = index_stages(root)
    data = joined(text, stages)
    present = package_files(package)
    # In a skill file only the opening and closing markers count: a bare `=======` line is
    # also a Markdown heading underline.
    unresolved = [
        rel for rel in present
        if any(
            line.startswith(("<<<<<<< ", ">>>>>>> "))
            for line in (package / rel).read_text(encoding="utf-8", errors="replace").splitlines()
        )
    ]
    if unresolved:
        raise NeedsPerson("conflict markers remain in: " + ", ".join(unresolved))
    missing = [rel for rel in present if rel not in data["files"]]
    if missing:
        raise NeedsPerson("no entry (write its origin first): " + ", ".join(missing))
    files, dropped, changed = {}, [], []
    for rel, entry in data["files"].items():
        if rel not in present:
            dropped.append(rel)
            continue
        new = dict(entry)
        new["sha256"] = sha256(package / rel)
        if new["sha256"] != entry.get("sha256"):
            changed.append(rel)
        files[rel] = new
    data["files"] = files
    report = {"conflicted": stages is not None or has_markers(text), "changed": changed, "dropped": dropped}
    return data, report


def render(data):
    return json.dumps(data, indent=2, ensure_ascii=False) + "\n"


def check(root):
    guard(root)
    package = Path(root) / PACKAGE
    text = (package / MANIFEST).read_text(encoding="utf-8")
    data, report = rebuild(root)
    report["ok"] = render(data) == text
    return report


def write(root, stages=None):
    path = Path(root) / PACKAGE / MANIFEST
    data, report = rebuild(root, stages)
    new = render(data)
    report["written"] = new != path.read_text(encoding="utf-8")
    if report["written"]:
        mode = stat.S_IMODE(os.lstat(path).st_mode)
        fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=".source-manifest.", suffix=".tmp")
        try:
            with os.fdopen(fd, "w", encoding="utf-8", newline="") as f:
                f.write(new)
            # mkstemp makes 0600; keep the manifest's own mode (#2849 audit N1).
            os.chmod(tmp, mode)
            os.replace(tmp, path)
        except BaseException:
            Path(tmp).unlink(missing_ok=True)
            raise
    return report


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("command", choices=["check", "write"])
    ap.add_argument("--root", default=str(Path(__file__).resolve().parent.parent),
                    help="repository root (default: the one holding this tool)")
    args = ap.parse_args(argv)
    try:
        report = check(args.root) if args.command == "check" else write(args.root)
    except NeedsPerson as e:
        print(json.dumps({"ok": False, "needs_person": str(e)}, ensure_ascii=False))
        return 2
    if args.command == "write":
        report["ok"] = True
    print(json.dumps(report, ensure_ascii=False))
    return 0 if report["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
