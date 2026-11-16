#!/usr/bin/env python3
"""Can a sandbox write the ledger's file operations? (Issue #3068 P8, AC9.)

The local ledger lives at `<HOME>/.moai/projects/<project id>/ledger.db`: a SQLite file
(journal_mode=DELETE), a hard link that keeps a damaged copy, a `.recover.lock` advisory file
lock, and a rename. A lane that runs under a sandbox (`codex sandbox`, an audit sub-agent) may
not be allowed to do these. This tool does exactly those operations, one recorded step each,
inside a folder the caller names, and then removes everything it made.

Usage:
  python3 -I tools/sandbox_write_probe.py --base ABSOLUTE_DIR [--project-id ID]

It writes only under `<base>/.moai/projects/<project id>/`. The default project id is
`probe-` plus 8 random hex digits; an id is 1-64 characters of [A-Za-z0-9._-], not `.` or `..`.

Stdout is one JSON object and the exit code is 0 whenever the probe ran, even when steps
failed (a failed step is a measurement):
  {"tool": "sandbox_write_probe", "version": 1, "base", "project_id", "python", "uid",
   "steps": [{"name", "ok", "errno", "error"}, ...], "all_ok", "left_behind",
   "writes_outside_base": false}
A step that cannot run because an earlier one it needs failed is `ok: false`, `errno: null`
and an error starting with `skipped:`. `cleanup` removes file by file and then folder by
folder (never a recursive delete); whatever it could not remove is listed in `left_behind`.

It refuses, with exit 2, one JSON line `{"refused": "<word>", "detail": ...}` and nothing
created, when:
  usage             an unknown or repeated argument, or a missing value
  base_missing      `--base` is missing or empty
  base_not_absolute `--base` is not an absolute path
  root              `--base` is `/`
  base_not_a_directory  `--base` is not an existing directory
  symlink           `--base` itself is a symbolic link
  real_home         `--base` resolves to the current user's real home directory (the
                    account database entry of the uid, never `$HOME`) or to anything inside
                    its `.moai`, `.moai-cli`, `.claude`, `.codex` or `go` folder; the real
                    home cannot be looked up
  moai_exists       `<base>/.moai` already exists (also as a link or file)
  bad_project_id    `--project-id` is not a valid id

It never points at the real `<HOME>/.moai` by itself: that measurement is guarded (Issue
#3068 plan, Q4) and is described in docs/design-review/appendix/ledger_sandbox_probe.md. It
reads no environment variable, prints no secret and makes no network call. Standard library
only; Python 3.9 or newer.
"""

import fcntl
import json
import os
import pwd
import re
import secrets
import sqlite3
import stat
import sys
from pathlib import Path

TOOL = "sandbox_write_probe"
VERSION = 1
KNOWN = ("--base", "--project-id")
PROTECTED = (".moai", ".moai-cli", ".claude", ".codex", "go")
ID_PATTERN = re.compile(r"^[A-Za-z0-9._-]{1,64}$")
LEDGER = "ledger.db"
DAMAGED = "ledger.db.damaged-probe"
RECOVERING = "ledger.db.recovering"
NEXT = "ledger.db.next"
LOCK = ".recover.lock"


class Refused(Exception):
    """A refusal: exit 2, nothing created. `word` is the machine-readable reason."""

    def __init__(self, word, detail):
        super().__init__(detail)
        self.word = word
        self.detail = detail


def real_home():
    """The current user's home directory from the account database, not from `$HOME`."""
    return pwd.getpwuid(os.getuid()).pw_dir


def parse_args(argv):
    opts = {}
    i = 0
    while i < len(argv):
        name = argv[i]
        if name not in KNOWN:
            raise Refused("usage", f"unknown argument {name!r}; usage: --base ABSOLUTE_DIR "
                                   "[--project-id ID]")
        if i + 1 >= len(argv):
            raise Refused("usage", f"{name} needs a value")
        if name in opts:
            raise Refused("usage", f"{name} given twice")
        opts[name] = argv[i + 1]
        i += 2
    return opts


def same_entry(a, b):
    """The same file-system entry (so a case-insensitive volume cannot hide `Go` for `go`)."""
    try:
        return os.path.samefile(a, b)
    except OSError:
        return False


def inside_protected(real_base, real_home_dir):
    """True when `real_base` is the home or a protected folder of it, or lies inside one."""
    if real_base == real_home_dir or same_entry(real_base, real_home_dir):
        return True
    folders = [os.path.join(real_home_dir, name) for name in PROTECTED]
    here = real_base
    while True:
        for folder in folders:
            if here == folder or same_entry(here, folder):
                return True
        parent = os.path.dirname(here)
        if parent == here:
            return False
        here = parent


def check_base(base, home_lookup):
    """The normalized base, or `Refused`. Checks run in the order of the module docstring."""
    if not base or "\0" in base:
        raise Refused("base_missing", "--base is required and must not be empty")
    if not os.path.isabs(base):
        raise Refused("base_not_absolute", f"--base must be an absolute path, got {base!r}")
    norm = os.path.normpath(base)
    if norm.startswith("//"):
        norm = "/" + norm.lstrip("/")
    if norm == "/" or os.path.realpath(base) == "/":
        raise Refused("root", "--base must not be /")
    try:
        mode = os.lstat(norm).st_mode
    except OSError as e:
        raise Refused("base_not_a_directory", f"cannot inspect --base: {e.strerror}") from e
    if stat.S_ISLNK(mode):
        raise Refused("symlink", "--base is a symbolic link; name the real folder")
    if not stat.S_ISDIR(mode):
        raise Refused("base_not_a_directory", "--base is not a directory")
    try:
        home = home_lookup()
    except (KeyError, OSError) as e:
        raise Refused("real_home", "the real home directory cannot be looked up") from e
    if not home or not os.path.isabs(home):
        raise Refused("real_home", "the real home directory cannot be looked up")
    if inside_protected(os.path.realpath(norm), os.path.realpath(home)):
        raise Refused("real_home", "--base is the real home or inside one of its state folders")
    if os.path.lexists(os.path.join(norm, ".moai")):
        raise Refused("moai_exists", "<base>/.moai already exists; the probe needs a base "
                                     "without one")
    return norm


def check_project_id(raw):
    if raw is None:
        return "probe-" + secrets.token_hex(4)
    if not ID_PATTERN.match(raw) or raw in (".", ".."):
        raise Refused("bad_project_id", "--project-id must be 1-64 characters of "
                                        "[A-Za-z0-9._-], not . or ..")
    return raw


class Probe:
    """One run: the steps, and the record of every path made so cleanup removes only those."""

    def __init__(self, base, project_id):
        self.base = base
        self.project_id = project_id
        self.moai = os.path.join(base, ".moai")
        self.projects = os.path.join(self.moai, "projects")
        self.project = os.path.join(self.projects, project_id)
        self.files = []
        self.dirs = []
        self.steps = []
        self.failed = set()

    def path(self, name):
        return os.path.join(self.project, name)

    def step(self, name, needs, fn):
        for need in needs:
            if need in self.failed:
                self.record(name, False, None, f"skipped: {need} failed")
                self.failed.add(name)
                return
        try:
            fn()
        except OSError as e:
            self.record(name, False, e.errno, e.strerror or str(e))
            self.failed.add(name)
        except Exception as e:  # sqlite3.Error and a wrong read-back are a result, not a crash
            self.record(name, False, None, f"{type(e).__name__}: {e}")
            self.failed.add(name)
        else:
            self.record(name, True, None, None)

    def record(self, name, ok, errno_, error):
        self.steps.append({"name": name, "ok": ok, "errno": errno_, "error": error})

    def make_file(self, path, flags=os.O_WRONLY, data=b""):
        fd = os.open(path, flags | os.O_CREAT | os.O_EXCL, 0o600)
        self.files.append(path)
        try:
            if data:
                os.write(fd, data)
        finally:
            os.close(fd)

    def keep_if_exists(self, path):
        if os.path.lexists(path) and path not in self.files:
            self.files.append(path)

    def mkdir_tree(self):
        for folder in (self.moai, self.projects, self.project):
            os.mkdir(folder, 0o700)
            self.dirs.append(folder)

    def create_file(self):
        self.make_file(self.path(LEDGER))

    def sqlite_write(self):
        conn = sqlite3.connect(self.path(LEDGER))
        try:
            mode = conn.execute("PRAGMA journal_mode=DELETE").fetchone()[0]
            if str(mode).lower() != "delete":
                raise RuntimeError(f"journal_mode is {mode!r}, not delete")
            conn.execute("CREATE TABLE probe (id INTEGER PRIMARY KEY, v TEXT)")
            conn.execute("INSERT INTO probe (v) VALUES ('ledger-probe')")
            conn.commit()
        finally:
            conn.close()
            self.keep_if_exists(self.path(LEDGER + "-journal"))

    def sqlite_read(self):
        conn = sqlite3.connect(Path(self.path(LEDGER)).as_uri() + "?mode=ro", uri=True)
        try:
            row = conn.execute("SELECT v FROM probe").fetchone()
        finally:
            conn.close()
        if row != ("ledger-probe",):
            raise RuntimeError(f"read back {row!r}")

    def hard_link(self):
        os.link(self.path(LEDGER), self.path(DAMAGED))
        self.files.append(self.path(DAMAGED))
        links = os.stat(self.path(LEDGER)).st_nlink
        if links != 2:
            raise RuntimeError(f"link count is {links}, not 2")

    def rename(self):
        self.make_file(self.path(NEXT), data=b"old")
        self.make_file(self.path(RECOVERING), data=b"new")
        os.rename(self.path(RECOVERING), self.path(NEXT))
        self.files.remove(self.path(RECOVERING))
        with open(self.path(NEXT), "rb") as f:
            if f.read() != b"new":
                raise RuntimeError("the rename did not replace the target")

    def advisory_lock(self):
        fd = os.open(self.path(LOCK), os.O_RDWR | os.O_CREAT, 0o600)
        self.files.append(self.path(LOCK))
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
            fcntl.flock(fd, fcntl.LOCK_UN)
        finally:
            os.close(fd)

    def fsync_dir(self):
        fd = os.open(self.project, os.O_RDONLY)
        try:
            os.fsync(fd)
        finally:
            os.close(fd)

    def cleanup(self):
        """Unlink what was made, file by file, then rmdir folder by folder."""
        problems = []
        for path in reversed(self.files):
            try:
                os.unlink(path)
            except FileNotFoundError:
                pass
            except OSError as e:
                problems.append(e)
        for folder in reversed(self.dirs):
            try:
                os.rmdir(folder)
            except FileNotFoundError:
                pass
            except OSError as e:
                problems.append(e)
        self.left_behind = [p for p in self.files + self.dirs if os.path.lexists(p)]
        if self.left_behind:
            raise problems[0] if problems else RuntimeError("left behind")

    def run(self):
        self.step("mkdir_tree", [], self.mkdir_tree)
        self.step("create_file", ["mkdir_tree"], self.create_file)
        self.step("sqlite_write", ["create_file"], self.sqlite_write)
        self.step("sqlite_read", ["sqlite_write"], self.sqlite_read)
        self.step("hard_link", ["create_file"], self.hard_link)
        self.step("rename", ["mkdir_tree"], self.rename)
        self.step("advisory_lock", ["mkdir_tree"], self.advisory_lock)
        self.step("fsync_dir", ["mkdir_tree"], self.fsync_dir)
        self.left_behind = []
        self.step("cleanup", [], self.cleanup)
        return {
            "tool": TOOL,
            "version": VERSION,
            "base": self.base,
            "project_id": self.project_id,
            "python": sys.version.split()[0],
            "uid": os.getuid(),
            "steps": self.steps,
            "all_ok": all(s["ok"] for s in self.steps),
            "left_behind": self.left_behind,
            "writes_outside_base": False,
        }


def execute(argv, home_lookup=real_home):
    """(exit code, JSON-able dict). A refusal happens before anything is created."""
    try:
        opts = parse_args(argv)
        base = check_base(opts.get("--base"), home_lookup)
        project_id = check_project_id(opts.get("--project-id"))
    except Refused as e:
        return 2, {"refused": e.word, "detail": e.detail}
    return 0, Probe(base, project_id).run()


def main(argv):
    code, out = execute(argv)
    sys.stdout.write(json.dumps(out, sort_keys=True) + "\n")
    return code


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
