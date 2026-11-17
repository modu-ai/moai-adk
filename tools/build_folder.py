#!/usr/bin/env python3
"""One cargo build folder per worktree, outside the worktree (card C-82).

User decision 2026-11-17 ("빌드 폴더를 한 곳으로 모으기"): agent and audit builds put
CARGO_TARGET_DIR in `<HOME>/.moai/build/<project>-<worktree name>/` (one folder per worktree,
never shared between worktrees), always with CARGO_INCREMENTAL=0, and delete the folder's
cargo output when the task ends; skills/moai/references/internal/build-folder.md is the rule.

  python3 -I tools/build_folder.py env     print two `export` lines for this worktree and write
                                           the folder's owner record (.moai-build.json)
  python3 -I tools/build_folder.py clean   delete this worktree's cargo output; files that are
                                           not cargo output (evidence) stay and are listed
  python3 -I tools/build_folder.py remove --preview <optimize-plan JSON> <name>...
                                           delete the cargo output of folders the user picked
                                           from the preview; every name must be a `trim` row

`<project>` is the primary checkout's folder name, `<worktree name>` the worktree's folder name;
the primary checkout and a worktree whose name already starts with `<project>-` use the name
alone. HOME must be absolute. A link anywhere on the folder's path, an unwritable HOME, or a
record of another worktree refuses: stdout stays empty and the exit code is 1 (2 for usage).
Use `e=$(python3 -I tools/build_folder.py env) && eval "$e"`: a bare eval of an empty output
would go on.
"""
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
from pathlib import Path

RECORD = ".moai-build.json"  # src/optimize.rs AGENT_BUILD_RECORD
RECORD_MAX = 4096  # src/optimize.rs AGENT_BUILD_RECORD_MAX
CARGO_FILES = {"CACHEDIR.TAG", ".rustc_info.json"}  # src/worktree_gc.rs split_target


class Refused(Exception):
    pass


def git(cwd, *args):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    for key in ("GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM"):
        if key in os.environ:
            env[key] = os.environ[key]
    out = subprocess.run(["git", *args], cwd=cwd, env=env, capture_output=True, text=True)
    if out.returncode != 0:
        raise Refused(f"git {' '.join(args)} failed: {out.stderr.strip()}")
    return out.stdout.strip()


def home():
    value = os.environ.get("HOME", "")
    if not value or not os.path.isabs(value):
        raise Refused(f"HOME must be an absolute folder, got {value!r}")
    return Path(value)


def safe(name):
    cleaned = re.sub(r"[^A-Za-z0-9._-]", "_", name)
    if cleaned in ("", ".", ".."):
        raise Refused(f"no usable folder name from {name!r}")
    return cleaned


def owner():
    """(git common dir, worktree root, folder name) of the current folder's worktree."""
    worktree = Path(os.path.realpath(git(".", "rev-parse", "--show-toplevel")))
    common = Path(git(".", "rev-parse", "--git-common-dir"))
    if not common.is_absolute():
        common = Path.cwd() / common
    common = Path(os.path.realpath(common))
    primary = common.parent if common.name == ".git" else common
    project = primary.name[:-4] if primary.name.endswith(".git") else primary.name
    project, name = safe(project), safe(worktree.name)
    if worktree != primary and name != project and not name.startswith(project + "-"):
        name = f"{project}-{name}"
    elif worktree == primary:
        name = project
    return common, worktree, name


def real_dir(path, create):
    """`path` is a real folder (created 0700 when `create`); a link or a file refuses."""
    try:
        st = os.lstat(path)
    except FileNotFoundError:
        if not create:
            return False
        try:
            os.mkdir(path, 0o700)
        except PermissionError as e:
            raise Refused(f"{path.parent} is not writable: {e}") from e
        except FileExistsError:
            pass
        st = os.lstat(path)
    if os.path.islink(path):
        raise Refused(f"{path} is a link; it is not followed")
    if not os.path.isdir(path) or st is None:
        raise Refused(f"{path} is not a folder")
    return True


def folder_path(create):
    h = home()
    common, worktree, name = owner()
    path = h
    for part in (".moai", "build", name):
        path = path / part
        if not real_dir(path, create):
            return common, worktree, path, False
    return common, worktree, path, True


def read_record(folder):
    rec = folder / RECORD
    try:
        st = os.lstat(rec)
    except FileNotFoundError:
        return None
    if os.path.islink(rec) or not os.path.isfile(rec) or st.st_size > RECORD_MAX:
        raise Refused(f"{rec} is not a regular file of at most {RECORD_MAX} bytes")
    try:
        data = json.loads(rec.read_text())
    except (ValueError, UnicodeDecodeError) as e:
        raise Refused(f"{rec} is not a record: {e}") from e
    return data if isinstance(data, dict) else None


def cmd_env():
    common, worktree, folder, _ = folder_path(create=True)
    record = {"schema_version": 1, "git_common_dir": str(common), "worktree": str(worktree)}
    old = read_record(folder)
    if old is not None and old.get("worktree") not in (None, str(worktree)):
        raise Refused(f"{folder} belongs to the worktree {old.get('worktree')}")
    if old != record:
        tmp = folder / (RECORD + ".tmp")
        if os.path.lexists(tmp):
            raise Refused(f"{tmp} is in the way")
        fd = os.open(tmp, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, "w") as f:
            json.dump(record, f)
        os.replace(tmp, folder / RECORD)
    print(f"export CARGO_TARGET_DIR={shlex.quote(str(folder))}")
    print("export CARGO_INCREMENTAL=0")
    return 0


def is_cargo_entry(path):
    """The rule of src/worktree_gc.rs split_target: what cargo wrote into a target folder."""
    if os.path.islink(path):
        return False
    if os.path.isdir(path):
        return (path.name == "tmp" or os.path.isfile(path / ".cargo-lock")
                or os.path.isfile(path / "CACHEDIR.TAG"))
    return os.path.isfile(path) and path.name in CARGO_FILES


def empty_cargo(folder):
    """Delete the cargo output of `folder`; the folder and its record go only when nothing else
    is left. Returns (removed names, kept names)."""
    removed, kept = [], []
    for entry in sorted(folder.iterdir()):
        if entry.name == RECORD:
            continue
        if is_cargo_entry(entry):
            if entry.is_dir():
                shutil.rmtree(entry)
            else:
                entry.unlink()
            removed.append(entry.name)
        else:
            kept.append(entry.name)
    if not kept:
        rec = folder / RECORD
        if os.path.lexists(rec):
            rec.unlink()
        folder.rmdir()
    return removed, kept


def cmd_clean():
    _, worktree, folder, exists = folder_path(create=False)
    if not exists:
        print(json.dumps({"folder": str(folder), "removed": [], "kept": [], "note": "no folder"}))
        return 0
    record = read_record(folder)
    if record is None or record.get("worktree") != str(worktree):
        raise Refused(f"{folder} has no owner record of this worktree; nothing deleted")
    removed, kept = empty_cargo(folder)
    print(json.dumps({"folder": str(folder), "removed": removed, "kept": kept}))
    return 0


def cmd_remove(preview_path, names):
    root = home() / ".moai" / "build"
    if os.path.islink(root) or not os.path.isdir(root):
        raise Refused(f"{root} is not a real folder")
    try:
        doc = json.loads(Path(preview_path).read_text())
    except (OSError, ValueError) as e:
        raise Refused(f"preview unreadable: {e}") from e
    rows = (doc.get("result") or doc).get("rows", []) if isinstance(doc, dict) else []
    picked_ok = {r.get("path") for r in rows if isinstance(r, dict)
                 and r.get("kind") == "agent_build_folder" and r.get("action") == "trim"}
    folders = []
    for name in names:
        if safe(name) != name or "/" in name:
            raise Refused(f"{name!r} is not one folder name")
        folder = root / name
        if str(folder) not in picked_ok:
            raise Refused(f"{name} is not a `trim` row of the preview; nothing deleted")
        if os.path.islink(folder) or not os.path.isdir(folder):
            raise Refused(f"{folder} is not a real folder now; nothing deleted")
        if read_record(folder) is None:
            raise Refused(f"{folder} has no owner record now; nothing deleted")
        folders.append(folder)
    report = []
    for folder in folders:
        removed, kept = empty_cargo(folder)
        report.append({"folder": str(folder), "removed": removed, "kept": kept})
    print(json.dumps(report))
    return 0


def main(argv):
    if argv[:1] == ["env"] and len(argv) == 1:
        action = cmd_env
    elif argv[:1] == ["clean"] and len(argv) == 1:
        action = cmd_clean
    elif argv[:2] == ["remove", "--preview"] and len(argv) >= 4:
        action = lambda: cmd_remove(argv[2], argv[3:])  # noqa: E731
    else:
        print("usage: build_folder.py env | clean | remove --preview <file> <name>...",
              file=sys.stderr)
        return 2
    try:
        return action()
    except Refused as e:
        print(f"build_folder: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
