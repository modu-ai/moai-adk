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
  python3 -I tools/build_folder.py seed --from <name>
                                           start this worktree's folder as a copy-on-write clone
                                           of the cargo output of `<HOME>/.moai/build/<name>` (an
                                           audit starts warm); where a clone is not possible the
                                           answer is `{"seeded": false, "reason": ...}` and no
                                           byte is copied: build cold. BUILD_FOLDER_NO_COW=1
                                           switches cloning off; the source is only read
  python3 -I tools/build_folder.py remove --preview <optimize-plan JSON> [--moai <bin>] <name>...
                                           delete the cargo output of folders the user picked
                                           from the preview; every name must be a `trim` row of
                                           that preview and of a fresh `moai optimize-plan` run
                                           on its integration_ref (from this repository)

`<project>` is the primary checkout's folder name, `<worktree name>` the worktree's folder name;
the primary checkout and a worktree whose name already starts with `<project>-` use the name
alone. HOME must be absolute; on Windows USERPROFILE stands in when HOME is unset (the rule of
`moai lane-env`). A link anywhere on the folder's path, an unwritable HOME, or a
record of another worktree refuses: stdout stays empty and the exit code is 1 (2 for usage).
Use `e=$(python3 -I tools/build_folder.py env) && eval "$e"`: a bare eval of an empty output
would go on.
"""
import errno
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


def home(environ=None, windows=None):
    """HOME, or on Windows USERPROFILE when HOME is unset — the rule of `moai lane-env` and
    the rest of MoAI (`key_vault::home_from`). It must be absolute."""
    environ = os.environ if environ is None else environ
    windows = os.name == "nt" if windows is None else windows
    value = environ.get("HOME", "")
    if not value and windows:
        value = environ.get("USERPROFILE", "")
        if not value:
            raise Refused("neither HOME nor USERPROFILE is set")
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


def ensure_folder():
    """This worktree's build folder, made 0700 with its owner record written; returns the
    folder. A folder of another worktree refuses."""
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
    return folder


def cmd_env():
    folder = ensure_folder()
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


RMTREE_HOOK = "onexc" if sys.version_info >= (3, 12) else "onerror"  # `onerror` is deprecated in 3.12


def make_writable_and_retry(func, path, _exc):
    """rmtree error hook: a clone keeps the read-only modes of its source, so a folder or file the
    owner cannot write is made writable for the owner and the call is tried once more."""
    parent = os.path.dirname(path)
    os.chmod(parent, os.stat(parent).st_mode | 0o700)
    if os.path.lexists(path) and not os.path.islink(path):
        os.chmod(path, os.stat(path).st_mode | 0o700)
    func(path)


def empty_cargo(folder):
    """Delete the cargo output of `folder`; the folder and its record go only when nothing else
    is left. Returns (removed names, kept names)."""
    removed, kept = [], []
    for entry in sorted(folder.iterdir()):
        if entry.name == RECORD:
            continue
        if is_cargo_entry(entry):
            if entry.is_dir():
                shutil.rmtree(entry, **{RMTREE_HOOK: make_writable_and_retry})
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


COW_OFF = "BUILD_FOLDER_NO_COW"  # set to 1: never clone, build cold
CLONE_NOFOLLOW = 0x0001  # <sys/clonefile.h>
CARGO_HOME_NOTE = ("dependencies are reused only when CARGO_HOME is the same path as the lane's "
                   "(the build info records registry paths); the workspace crate always rebuilds")


def cargo_entries(folder):
    """The cargo output entries of `folder`, by name (never the record, never a link)."""
    return [e for e in sorted(folder.iterdir()) if e.name != RECORD and is_cargo_entry(e)]


def hold_source(folder):
    """A shared hold on the cargo lock of every profile folder of the source, so a lane that starts
    to build while the clone is made waits for it. Returns the open lock files, or None when a cargo
    run already holds one (the source is being built). The caller closes the files. Nothing is
    written to the source."""
    import fcntl
    held = []
    for entry in cargo_entries(folder):
        lock = entry / ".cargo-lock"
        if not entry.is_dir() or not os.path.isfile(lock) or os.path.islink(lock):
            continue
        fd = os.open(lock, os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0))
        try:
            fcntl.flock(fd, fcntl.LOCK_SH | fcntl.LOCK_NB)
        except OSError:
            os.close(fd)
            for other in held:
                os.close(other)
            return None
        held.append(fd)
    return held


def clone_entry(src, dst):
    """Clone one entry copy-on-write. Returns the method name, or raises OSError. There is no
    other way to copy here: where a clone fails the answer is `seeded: false`."""
    if sys.platform == "darwin":
        import ctypes
        libc = ctypes.CDLL(None, use_errno=True)
        libc.clonefile.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_uint32]
        if libc.clonefile(os.fsencode(src), os.fsencode(dst), CLONE_NOFOLLOW) != 0:
            err = ctypes.get_errno()
            raise OSError(err, os.strerror(err))
        return "clonefile"
    if sys.platform.startswith("linux"):
        out = subprocess.run(["cp", "--reflink=always", "-a", "--", str(src), str(dst)],
                             capture_output=True, text=True)
        if out.returncode != 0:
            text = out.stderr.strip() or "cp --reflink=always failed"
            # cp gives an exit code, not an errno: read its message for "cannot clone here".
            low = text.lower()
            number = errno.ENOTSUP if "not supported" in low or "cross-device" in low else errno.EIO
            raise OSError(number, text)
        return "reflink"
    raise OSError(errno.ENOTSUP, "no copy-on-write clone on this platform")


def cmd_seed(name):
    """Seed this worktree's build folder from the lane folder `name` by copy-on-write clone."""
    if safe(name) != name or "/" in name:
        raise Refused(f"{name!r} is not one folder name")
    root = home()
    missing = False
    for part in (".moai", "build"):
        root = root / part
        if not os.path.lexists(root):
            missing = True  # no lane has built here yet: nothing to clone from, not an error
        elif not real_dir(root, create=False):
            raise Refused(f"{root} is a link or not a folder; nothing was seeded")
    _, _, own_name = owner()
    if name == own_name:
        raise Refused(f"{name} is this worktree's own build folder")
    source = root / name
    if os.path.islink(source) or (os.path.lexists(source) and not os.path.isdir(source)):
        raise Refused(f"{source} is a link or not a folder; it is not followed")
    folder = ensure_folder()
    # A name that differs only in letter case reaches the own folder on a case-insensitive disk.
    if os.path.isdir(source) and os.path.samefile(source, folder):
        raise Refused(f"{name} is this worktree's own build folder")

    def answer(seeded, reason=None, **more):
        doc = {"seeded": seeded, "folder": str(folder), "from": str(source), **more}
        if reason:
            doc["reason"] = reason
        print(json.dumps(doc))
        return 0

    if os.environ.get(COW_OFF) == "1":
        return answer(False, "disabled")
    if missing or not os.path.isdir(source):
        return answer(False, "source_missing")
    if cargo_entries(folder):
        return answer(False, "already_built")
    held = []
    made, method = [], None
    try:
        held = hold_source(source)
        if held is None:
            return answer(False, "source_building")
        wanted = cargo_entries(source)
        if not wanted:
            return answer(False, "source_empty")
        # Nothing of the own folder is overwritten or taken back: stop before the first clone when
        # a path is already taken (evidence, a link, a half-built folder).
        taken = [e.name for e in wanted if os.path.lexists(folder / e.name)]
        if taken:
            return answer(False, "target_in_the_way", detail=", ".join(taken))
        for entry in wanted:
            made.append(folder / entry.name)  # the path did not exist before this run
            method = clone_entry(entry, folder / entry.name)
    except OSError as e:
        # Take back only what this run made: every path in `made` was free just before.
        for path in made:
            if path.is_dir() and not path.is_symlink():
                shutil.rmtree(path, ignore_errors=True)
            elif os.path.lexists(path):
                path.unlink()
        reason = "unsupported" if e.errno in (errno.ENOTSUP, errno.EXDEV) else "clone_failed"
        return answer(False, reason, detail=str(e))
    finally:
        for fd in held or []:
            os.close(fd)
    return answer(True, method=method, entries=[p.name for p in made], note=CARGO_HOME_NOTE)


def trim_paths(doc):
    """The `agent_build_folder` rows marked `trim` in an optimize-plan answer (or its result)."""
    result = doc.get("result") if isinstance(doc, dict) and "result" in doc else doc
    rows = result.get("rows", []) if isinstance(result, dict) else []
    return {r.get("path") for r in rows if isinstance(r, dict)
            and r.get("kind") == "agent_build_folder" and r.get("action") == "trim"}


def fresh_preview(moai, ref):
    """Audit N1: the preview as it is now, from `moai optimize-plan` on the same integration
    ref, so a hand-written or stale preview never decides a deletion alone."""
    project = git(".", "rev-parse", "--show-toplevel")
    out = subprocess.run([moai, "optimize-plan", project, "--integration-ref", ref],
                         capture_output=True, text=True)
    try:
        doc = json.loads(out.stdout)
    except ValueError as e:
        raise Refused(f"fresh preview unreadable (exit {out.returncode}): {e}") from e
    if out.returncode != 0 or not doc.get("ok"):
        raise Refused(f"fresh preview failed (exit {out.returncode}): {out.stdout.strip()}")
    return trim_paths(doc)


def cmd_remove(preview_path, moai, names):
    try:
        doc = json.loads(Path(preview_path).read_text())
    except (OSError, ValueError) as e:
        raise Refused(f"preview unreadable: {e}") from e
    result = doc.get("result") if isinstance(doc, dict) and "result" in doc else doc
    ref = result.get("integration_ref") if isinstance(result, dict) else None
    if not isinstance(ref, str) or not ref or ref.startswith("-"):
        raise Refused("the preview names no integration_ref; nothing deleted")
    picked_ok = trim_paths(doc)
    # Audit B1: lstat every step, `.moai` and `build` included, never following a link.
    root = home()
    for part in (".moai", "build"):
        root = root / part
        if not real_dir(root, create=False):
            raise Refused(f"{root} is missing; nothing deleted")
    folders = []
    for name in names:
        if safe(name) != name or "/" in name:
            raise Refused(f"{name!r} is not one folder name")
        folder = root / name
        if str(folder) not in picked_ok:
            raise Refused(f"{name} is not a `trim` row of the preview; nothing deleted")
        if not real_dir(folder, create=False):
            raise Refused(f"{folder} is missing; nothing deleted")
        if read_record(folder) is None:
            raise Refused(f"{folder} has no owner record now; nothing deleted")
        folders.append(folder)
    fresh = fresh_preview(moai, ref)
    stale = [str(f) for f in folders if str(f) not in fresh]
    if stale:
        raise Refused(f"not a `trim` row of the fresh preview: {', '.join(stale)}; nothing deleted")
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
    elif argv[:2] == ["seed", "--from"] and len(argv) == 3:
        action = lambda: cmd_seed(argv[2])  # noqa: E731
    elif argv[:2] == ["remove", "--preview"] and len(argv) >= 4:
        rest, moai = argv[3:], "moai"
        if rest[:1] == ["--moai"] and len(rest) >= 3:
            rest, moai = rest[2:], rest[1]
        action = lambda: cmd_remove(argv[2], moai, rest)  # noqa: E731
    else:
        print("usage: build_folder.py env | clean | seed --from <lane folder name> | "
              "remove --preview <file> [--moai <bin>] <name>...", file=sys.stderr)
        return 2
    try:
        return action()
    except Refused as e:
        print(f"build_folder: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
