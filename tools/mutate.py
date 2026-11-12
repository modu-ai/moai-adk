#!/usr/bin/env python3
"""Run explicit one-line mutations in a retained, independent HEAD clone."""

import argparse
from contextlib import contextmanager
import hashlib
import json
import math
import os
from pathlib import Path, PurePosixPath
import re
import subprocess
import signal
import stat
import sys
import time
import uuid


class Blocked(Exception):
    pass


class UnsafeProcess(Blocked):
    pass


class Parser(argparse.ArgumentParser):
    def error(self, message):
        raise Blocked(message)


def git(root, env, *args, check=True):
    # This CLI is single-threaded. Change cwd only in the forked Git child:
    # macOS cannot traverse /dev/fd/N as a directory, unlike Linux.
    pinned_cwd = ({"pass_fds": (root,), "preexec_fn": lambda: os.fchdir(root)}
                  if isinstance(root, int) else {})
    result = subprocess.run(
        ["git", "--no-optional-locks", "-c", "core.fsmonitor=false",
         "-c", "maintenance.auto=false",
         "-c", "gc.auto=0", *args],
        cwd=None if isinstance(root, int) else root, env=env,
        **pinned_cwd,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if check and result.returncode:
        raise Blocked("git {}: {}".format(args[0], result.stderr.decode("utf-8", "replace").strip()))
    return result


def digest(data):
    return hashlib.sha256(data).hexdigest()


def _index_entries(raw):
    entries = {}
    for row in raw.split(b"\0"):
        if not row:
            continue
        metadata, name = row.split(b"\t", 1)
        tag, mode, oid, stage = metadata.split()
        name = os.fsdecode(name)
        if name in entries or stage != b"0":
            raise Blocked("source index must contain only stage-zero entries")
        entries[name] = (tag, mode, oid)
    return entries


def _repository_identity(root, env):
    actual = os.fsdecode(git(root, env, "rev-parse", "--show-toplevel").stdout.rstrip(b"\n"))
    expected = os.fstat(root) if isinstance(root, int) else os.stat(root)
    found = os.stat(actual)
    if (expected.st_dev, expected.st_ino) != (found.st_dev, found.st_ino):
        raise Blocked("gitlink must have its own actual Git worktree root")
    git_dir = os.fsdecode(git(root, env, "rev-parse", "--absolute-git-dir").stdout.rstrip(b"\n"))
    info = os.stat(git_dir)
    return info.st_dev, info.st_ino


def source_snapshot(root, env, ancestors=()):
    identity = _repository_identity(root, env)
    if identity in ancestors:
        raise Blocked("recursive gitlink repository identity")
    ancestors = (*ancestors, identity)
    head = git(root, env, "rev-parse", "--verify", "HEAD^{commit}").stdout.strip().decode("ascii")
    metadata = git(root, env, "ls-files", "--stage", "-v", "-z").stdout
    entries = _index_entries(metadata)
    # Read tracked entries before asking Git to inspect working-tree status.
    # An unsafe ancestor aborts this snapshot without opening its descendants.
    tracked = _tracked_files(root, env, entries, ancestors)
    raw_metadata = []
    for name in ("index", "HEAD"):
        path = Path(os.fsdecode(git(root, env, "rev-parse", "--git-path", name).stdout.rstrip(b"\n")))
        if path.is_absolute():
            data, _ = read_regular(Path(path.anchor), path.relative_to(path.anchor).as_posix())
        else:
            data, _ = read_regular(root, path.as_posix())
        raw_metadata.append(data)
    return (head, raw_metadata[0], tracked,
            git(root, env, "status", "--porcelain=v1", "--untracked-files=all",
                "--ignore-submodules=none").stdout,
            git(root, env, "show-ref", check=False).stdout, raw_metadata[1], metadata,
            git(root, env, "ls-tree", "-r", "-z", "--full-tree", "HEAD").stdout)


def require_clean_source(snapshot):
    if snapshot[3]:
        raise Blocked("source index/worktree must be clean, including untracked files")
    entries = _index_entries(snapshot[6])
    committed = {}
    for row in snapshot[7].split(b"\0"):
        if row:
            metadata, name = row.split(b"\t", 1)
            mode, kind, oid = metadata.split()
            committed[os.fsdecode(name)] = (mode, oid)
    if entries.keys() != committed.keys():
        raise Blocked("source index paths differ from HEAD")
    for name, (tag, mode, oid) in entries.items():
        if tag.islower() or tag.upper() == b"S":
            raise Blocked("source index has assume-unchanged/skip-worktree flags: " + name)
        if (mode, oid) != committed[name]:
            raise Blocked("source index differs from HEAD: " + name)
        actual = snapshot[2][name]
        if actual[0] == "file":
            actual_mode = b"100755" if actual[1] & stat.S_IXUSR else b"100644"
            matches = actual_mode == mode and actual[3] == oid
        elif actual[0] == "symlink":
            matches = mode == b"120000" and actual[2] == oid
        elif actual[0] == "gitlink" and mode == b"160000":
            state = actual[2]
            if state[0] == "initialized":
                child = state[1]
                require_clean_source(child)
                matches = child[0].encode("ascii") == oid
            else:
                matches = not state[1]
        else:
            matches = False
        if not matches:
            raise Blocked("working tracked entry differs from HEAD: " + name)


def _relative_parts(relative):
    if (not isinstance(relative, str) or not relative or "\\" in relative
            or "\0" in relative or relative.startswith("/")
            or any(part in ("", ".", "..") for part in relative.split("/"))):
        raise Blocked("tracked path must be relative without traversal")
    return PurePosixPath(relative).parts


@contextmanager
def _posix_parent(root, relative):
    parts = _relative_parts(relative)
    directory = (os.dup(root) if isinstance(root, int) else
                 os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW))
    try:
        for part in parts[:-1]:
            child = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=directory)
            os.close(directory)
            directory = child
        yield directory, parts[-1]
    finally:
        os.close(directory)


def regular_fd(root, relative, flags=os.O_RDONLY):
    """Traverse only directories and a regular leaf, without following symlinks."""
    if os.name == "nt":
        return windows_regular_fd(root, relative, flags)
    with _posix_parent(root, relative) as (directory, leaf):
        fd = os.open(leaf, flags | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=directory)
        if not stat.S_ISREG(os.fstat(fd).st_mode):
            os.close(fd)
            raise Blocked("mutation path is not a regular file: " + relative)
        return fd


@contextmanager
def _windows_parent(root, relative):
    """Pin checked Windows ancestors against replacement while accessing a leaf."""
    import ctypes
    from ctypes import wintypes

    parts = _relative_parts(relative)
    kernel = ctypes.WinDLL("kernel32", use_last_error=True)
    create = kernel.CreateFileW
    create.argtypes = (wintypes.LPCWSTR, wintypes.DWORD, wintypes.DWORD,
                       ctypes.c_void_p, wintypes.DWORD, wintypes.DWORD, wintypes.HANDLE)
    create.restype = wintypes.HANDLE
    close = kernel.CloseHandle
    close.argtypes = (wintypes.HANDLE,)

    class AttributeTag(ctypes.Structure):
        _fields_ = [("attributes", wintypes.DWORD), ("tag", wintypes.DWORD)]

    inspect = kernel.GetFileInformationByHandleEx
    inspect.argtypes = (wintypes.HANDLE, ctypes.c_int, ctypes.c_void_p, wintypes.DWORD)
    handles = []
    path = Path(root)
    try:
        for part in (None, *parts[:-1]):
            if part is not None:
                path = path / part
            # Omit FILE_SHARE_DELETE so a checked component cannot be renamed
            # or replaced before the leaf has been read or opened.
            handle = create(str(path), 0x80, 3, None, 3, 0x02200000, None)
            if handle == ctypes.c_void_p(-1).value:
                raise ctypes.WinError(ctypes.get_last_error())
            handles.append(handle)
            tag = AttributeTag()
            if not inspect(handle, 9, ctypes.byref(tag), ctypes.sizeof(tag)):
                raise ctypes.WinError(ctypes.get_last_error())
            if tag.attributes & 0x400 or not tag.attributes & 0x10:
                raise Blocked("reparse point or non-directory tracked ancestor: " + str(path))
        yield path / parts[-1]
    finally:
        for handle in reversed(handles):
            close(handle)


def windows_regular_fd(root, relative, flags):
    with _windows_parent(root, relative) as path:
        return _windows_leaf_fd(path, flags)


def _windows_leaf_fd(path, flags):
    """Open a Windows leaf without following reparse points or junctions."""
    import ctypes
    from ctypes import wintypes
    import msvcrt

    kernel = ctypes.WinDLL("kernel32", use_last_error=True)
    create = kernel.CreateFileW
    create.argtypes = (wintypes.LPCWSTR, wintypes.DWORD, wintypes.DWORD,
                       ctypes.c_void_p, wintypes.DWORD, wintypes.DWORD, wintypes.HANDLE)
    create.restype = wintypes.HANDLE
    access = 0x40000000 if flags & os.O_WRONLY else 0x80000000
    handle = create(str(path), access, 7, None, 3, 0x00200000, None)
    if handle == ctypes.c_void_p(-1).value:
        raise ctypes.WinError(ctypes.get_last_error())
    close = kernel.CloseHandle
    close.argtypes = (wintypes.HANDLE,)
    try:
        class AttributeTag(ctypes.Structure):
            _fields_ = [("attributes", wintypes.DWORD), ("tag", wintypes.DWORD)]
        tag = AttributeTag()
        inspect = kernel.GetFileInformationByHandleEx
        inspect.argtypes = (wintypes.HANDLE, ctypes.c_int, ctypes.c_void_p, wintypes.DWORD)
        if not inspect(handle, 9, ctypes.byref(tag), ctypes.sizeof(tag)):
            raise ctypes.WinError(ctypes.get_last_error())
        if tag.attributes & (0x400 | 0x10):
            raise Blocked("mutation leaf is a reparse point or directory")
        fd = msvcrt.open_osfhandle(handle, flags | os.O_BINARY)
        handle = None
        if not stat.S_ISREG(os.fstat(fd).st_mode):
            os.close(fd)
            raise Blocked("mutation leaf is not a regular file")
        return fd
    finally:
        if handle is not None:
            close(handle)


def read_regular(root, relative):
    fd = regular_fd(root, relative)
    with os.fdopen(fd, "rb") as handle:
        return handle.read(), stat.S_IMODE(os.fstat(handle.fileno()).st_mode)


def _read_tracked_entry(root, relative):
    """Read one rooted entry, never following a parent or a leaf symlink."""
    def read_file(fd):
        with os.fdopen(fd, "rb") as handle:
            info = os.fstat(handle.fileno())
            if not stat.S_ISREG(info.st_mode):
                raise Blocked("tracked leaf is not regular: " + relative)
            return "file", stat.S_IMODE(info.st_mode), handle.read()

    try:
        if os.name == "nt":
            with _windows_parent(root, relative) as path:
                info = path.lstat()
                reparse = getattr(info, "st_file_attributes", 0) & 0x400
                if stat.S_ISLNK(info.st_mode):
                    if getattr(info, "st_reparse_tag", 0) != 0xA000000C:
                        raise Blocked("unsupported tracked reparse point: " + relative)
                    return "symlink", stat.S_IMODE(info.st_mode), os.fsencode(os.readlink(path))
                if reparse:
                    raise Blocked("reparse point in tracked leaf: " + relative)
                if stat.S_ISREG(info.st_mode):
                    return read_file(_windows_leaf_fd(path, os.O_RDONLY))
                if stat.S_ISDIR(info.st_mode):
                    return "directory", stat.S_IMODE(info.st_mode), None
        else:
            with _posix_parent(root, relative) as (directory, leaf):
                info = os.stat(leaf, dir_fd=directory, follow_symlinks=False)
                if stat.S_ISLNK(info.st_mode):
                    return ("symlink", stat.S_IMODE(info.st_mode),
                            os.fsencode(os.readlink(leaf, dir_fd=directory)))
                if stat.S_ISREG(info.st_mode):
                    fd = os.open(leaf, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK,
                                 dir_fd=directory)
                    return read_file(fd)
                if stat.S_ISDIR(info.st_mode):
                    return "directory", stat.S_IMODE(info.st_mode), None
        raise Blocked("tracked path is not a supported entry: " + relative)
    except OSError as exc:
        raise Blocked("cannot safely read tracked entry {}: {}".format(relative, exc)) from exc


@contextmanager
def _tracked_directory(root, relative):
    """Keep the gitlink and all its parents pinned while inspecting its state."""
    try:
        if os.name == "nt":
            # Treat the gitlink itself as an ancestor, so the existing Windows
            # reader pins and checks it for reparse points as well.
            with _windows_parent(root, relative + "/.git") as path:
                yield path.parent
        else:
            with _posix_parent(root, relative) as (directory, leaf):
                child = os.open(leaf, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW,
                                dir_fd=directory)
                try:
                    # Git receives this inherited descriptor as its cwd; no
                    # pathname traversal can redirect it after validation.
                    yield child
                finally:
                    os.close(child)
    except OSError as exc:
        raise Blocked("cannot safely inspect gitlink {}: {}".format(relative, exc)) from exc


def _gitlink_snapshot(root, relative, env, ancestors):
    with _tracked_directory(root, relative) as child:
        names = tuple(sorted(os.listdir(child)))
        if ".git" not in names:
            # Never let Git discover the parent repository from an
            # uninitialized directory, or overlook bytes planted in it.
            return "uninitialized", names
        kind, _, _ = _read_tracked_entry(child, ".git")
        if kind not in ("file", "directory"):
            raise Blocked("gitlink metadata must not be a symlink: " + relative)
        return "initialized", source_snapshot(child, env, ancestors)


def write_regular(root, relative, data, mode):
    fd = regular_fd(root, relative, os.O_WRONLY)
    with os.fdopen(fd, "wb") as handle:
        handle.truncate(0)
        handle.write(data)
        handle.flush()
        if hasattr(os, "fchmod"):
            os.fchmod(handle.fileno(), mode)
        elif stat.S_IMODE(os.fstat(handle.fileno()).st_mode) != mode:
            raise Blocked("Windows file mode changed unexpectedly")


def validate(spec, root, env):
    if not isinstance(spec, dict) or type(spec.get("schema_version")) is not int or spec["schema_version"] != 1:
        raise Blocked("spec requires schema_version 1")
    timeout = spec.get("timeout_seconds", 300)
    if type(timeout) not in (int, float) or not math.isfinite(timeout) or timeout <= 0:
        raise Blocked("timeout_seconds must be positive and finite")
    mutations = spec.get("mutations")
    if not isinstance(mutations, list) or not mutations:
        raise Blocked("mutations must be a nonempty array")
    entries = {}
    for row in git(root, env, "ls-files", "--stage", "-z").stdout.split(b"\0"):
        if row:
            metadata, name = row.split(b"\t", 1)
            entries[os.fsdecode(name)] = metadata.split()
    prepared = []
    identifiers = set()
    for mutation in mutations:
        if not isinstance(mutation, dict):
            raise Blocked("mutation must be an object")
        identifier = mutation.get("id")
        if not isinstance(identifier, str) or not identifier.strip() or identifier in identifiers:
            raise Blocked("mutation ids must be nonempty and unique")
        identifiers.add(identifier)
        name = mutation.get("file")
        if (not isinstance(name, str) or not name or "\\" in name or "\0" in name
                or name.startswith("/") or any(p in ("", ".", "..") for p in name.split("/"))):
            raise Blocked("file must be a relative path without traversal")
        entry = entries.get(name)
        if not entry or entry[0] not in (b"100644", b"100755") or entry[2] != b"0":
            raise Blocked("file must be a tracked regular file: " + name)
        data, mode = read_regular(root, name)
        committed = git(root, env, "show", "HEAD:" + name).stdout
        if data != committed:
            raise Blocked("working file differs from HEAD: " + name)
        old, new = mutation.get("old"), mutation.get("new")
        if any(not isinstance(v, str) or not v or "\n" in v or "\r" in v or "\0" in v for v in (old, new)) or old == new:
            raise Blocked("old/new must be distinct nonempty single lines")
        if old.lstrip().startswith(("//", "#")):
            raise Blocked("mutation must replace code, not comment/attribute lines")
        old_bytes, new_bytes = old.encode("utf-8"), new.encode("utf-8")
        lines = data.splitlines(keepends=True)
        matches = [i for i, line in enumerate(lines) if line.rstrip(b"\r\n") == old_bytes]
        if data.count(old_bytes) != 1 or len(matches) != 1:
            raise Blocked("old must occur uniquely as a complete line: " + name)
        targets = mutation.get("targets")
        if not isinstance(targets, list) or not targets:
            raise Blocked("targets must be a nonempty array")
        commands = []
        for target in targets:
            if not isinstance(target, dict) or set(target) - {"kind", "name", "filter"}:
                raise Blocked("invalid target selector")
            kind = target.get("kind")
            if kind not in ("lib", "bin", "test", "example", "bench"):
                raise Blocked("invalid target kind")
            command = ["cargo", "test", "--offline", "--locked", "--message-format=json", "--" + kind]
            if kind == "lib":
                if "name" in target:
                    raise Blocked("lib must not have a name")
            else:
                target_name = target.get("name")
                if not isinstance(target_name, str) or not re.fullmatch(r"[A-Za-z0-9_][A-Za-z0-9_.-]*", target_name):
                    raise Blocked("target name must be a simple literal")
                command.append(target_name)
            if "filter" in target:
                value = target["filter"]
                if not isinstance(value, str) or not value.strip() or value.startswith("-") or "\0" in value:
                    raise Blocked("filter must be nonempty and not an option")
                command.append(value)
            commands.append(command)
        at = matches[0]
        ending = lines[at][len(old_bytes):]
        lines[at] = new_bytes + ending
        prepared.append((mutation, data, b"".join(lines), mode, commands))
    return prepared, timeout


class _WindowsJob:
    """Own the child tree even after Cargo exits, without a PID-based fallback."""

    def __init__(self):
        import ctypes
        from ctypes import wintypes

        self.ctypes = ctypes
        self.kernel = ctypes.WinDLL("kernel32", use_last_error=True)
        self.handle = None
        self.process = None
        self.assigned = False
        self.stopped = False
        self.failure = None

        class BasicLimits(ctypes.Structure):
            _fields_ = [
                ("PerProcessUserTimeLimit", ctypes.c_longlong),
                ("PerJobUserTimeLimit", ctypes.c_longlong),
                ("LimitFlags", wintypes.DWORD),
                ("MinimumWorkingSetSize", ctypes.c_size_t),
                ("MaximumWorkingSetSize", ctypes.c_size_t),
                ("ActiveProcessLimit", wintypes.DWORD),
                ("Affinity", ctypes.c_size_t),
                ("PriorityClass", wintypes.DWORD),
                ("SchedulingClass", wintypes.DWORD),
            ]

        class IoCounters(ctypes.Structure):
            _fields_ = [(name, ctypes.c_ulonglong) for name in (
                "ReadOperationCount", "WriteOperationCount", "OtherOperationCount",
                "ReadTransferCount", "WriteTransferCount", "OtherTransferCount")]

        class ExtendedLimits(ctypes.Structure):
            _fields_ = [
                ("BasicLimitInformation", BasicLimits), ("IoInfo", IoCounters),
                ("ProcessMemoryLimit", ctypes.c_size_t),
                ("JobMemoryLimit", ctypes.c_size_t),
                ("PeakProcessMemoryUsed", ctypes.c_size_t),
                ("PeakJobMemoryUsed", ctypes.c_size_t),
            ]

        class Accounting(ctypes.Structure):
            _fields_ = [
                ("TotalUserTime", ctypes.c_longlong),
                ("TotalKernelTime", ctypes.c_longlong),
                ("ThisPeriodTotalUserTime", ctypes.c_longlong),
                ("ThisPeriodTotalKernelTime", ctypes.c_longlong),
                ("TotalPageFaultCount", wintypes.DWORD),
                ("TotalProcesses", wintypes.DWORD),
                ("ActiveProcesses", wintypes.DWORD),
                ("TotalTerminatedProcesses", wintypes.DWORD),
            ]

        class ThreadEntry(ctypes.Structure):
            _fields_ = [
                ("dwSize", wintypes.DWORD), ("cntUsage", wintypes.DWORD),
                ("th32ThreadID", wintypes.DWORD),
                ("th32OwnerProcessID", wintypes.DWORD),
                ("tpBasePri", wintypes.LONG), ("tpDeltaPri", wintypes.LONG),
                ("dwFlags", wintypes.DWORD),
            ]

        self.accounting_type = Accounting
        self.thread_type = ThreadEntry
        signatures = {
            "CreateJobObjectW": ((ctypes.c_void_p, wintypes.LPCWSTR), wintypes.HANDLE),
            "SetInformationJobObject": (
                (wintypes.HANDLE, ctypes.c_int, ctypes.c_void_p, wintypes.DWORD), wintypes.BOOL),
            "QueryInformationJobObject": (
                (wintypes.HANDLE, ctypes.c_int, ctypes.c_void_p, wintypes.DWORD,
                 ctypes.POINTER(wintypes.DWORD)), wintypes.BOOL),
            "AssignProcessToJobObject": ((wintypes.HANDLE, wintypes.HANDLE), wintypes.BOOL),
            "TerminateJobObject": ((wintypes.HANDLE, wintypes.UINT), wintypes.BOOL),
            "TerminateProcess": ((wintypes.HANDLE, wintypes.UINT), wintypes.BOOL),
            "CreateToolhelp32Snapshot": ((wintypes.DWORD, wintypes.DWORD), wintypes.HANDLE),
            "Thread32First": ((wintypes.HANDLE, ctypes.POINTER(ThreadEntry)), wintypes.BOOL),
            "Thread32Next": ((wintypes.HANDLE, ctypes.POINTER(ThreadEntry)), wintypes.BOOL),
            "OpenThread": ((wintypes.DWORD, wintypes.BOOL, wintypes.DWORD), wintypes.HANDLE),
            "ResumeThread": ((wintypes.HANDLE,), wintypes.DWORD),
            "GetProcessIdOfThread": ((wintypes.HANDLE,), wintypes.DWORD),
            "CloseHandle": ((wintypes.HANDLE,), wintypes.BOOL),
        }
        for name, (arguments, result) in signatures.items():
            function = getattr(self.kernel, name)
            function.argtypes, function.restype = arguments, result
        self.handle = self.kernel.CreateJobObjectW(None, None)
        if not self.handle:
            self._error("CreateJobObjectW")
        try:
            limits = ExtendedLimits()
            # No breakaway flags: all descendants must remain in this job.
            limits.BasicLimitInformation.LimitFlags = 0x2000  # KILL_ON_JOB_CLOSE
            if not self.kernel.SetInformationJobObject(
                    self.handle, 9, ctypes.byref(limits), ctypes.sizeof(limits)):
                self._error("SetInformationJobObject")
        except BaseException:
            self.close()
            raise

    def _error(self, operation):
        error = self.ctypes.WinError(self.ctypes.get_last_error())
        raise UnsafeProcess(operation + " failed; child tree not proven stopped: " + str(error))

    def close(self):
        if self.handle is not None:
            if not self.kernel.CloseHandle(self.handle):
                self._error("CloseHandle(job)")
            self.handle = None

    def assign_and_resume(self, process):
        self.process = process
        if not self.kernel.AssignProcessToJobObject(self.handle, int(process._handle)):
            # Nested jobs are supported when the host permits assignment.
            # Never resume a process whose job assignment was refused.
            self._error("AssignProcessToJobObject")
        self.assigned = True
        ctypes = self.ctypes
        snapshot = self.kernel.CreateToolhelp32Snapshot(0x00000004, 0)  # SNAPTHREAD
        if snapshot == ctypes.c_void_p(-1).value:
            self._error("CreateToolhelp32Snapshot")
        try:
            entry = self.thread_type()
            entry.dwSize = ctypes.sizeof(entry)
            found = self.kernel.Thread32First(snapshot, ctypes.byref(entry))
            owned = []
            while found:
                if entry.th32OwnerProcessID == process.pid:
                    owned.append(entry.th32ThreadID)
                entry.dwSize = ctypes.sizeof(entry)
                found = self.kernel.Thread32Next(snapshot, ctypes.byref(entry))
            if ctypes.get_last_error() != 18:  # ERROR_NO_MORE_FILES
                self._error("Thread32 enumeration")
            # CREATE_SUSPENDED has not run any user code: exactly the primary
            # thread must exist. Refuse ambiguous ownership rather than guessing.
            if len(owned) != 1:
                raise UnsafeProcess("cannot identify the suspended owned primary thread")
            # QUERY_LIMITED_INFORMATION also allows ownership verification.
            thread = self.kernel.OpenThread(0x0002 | 0x0800, False, owned[0])
            if not thread:
                self._error("OpenThread")
            try:
                if self.kernel.GetProcessIdOfThread(thread) != process.pid:
                    raise UnsafeProcess("suspended thread ownership changed")
                if self.kernel.ResumeThread(thread) != 1:
                    raise UnsafeProcess("owned primary thread did not have suspension count one")
            finally:
                if not self.kernel.CloseHandle(thread):
                    self._error("CloseHandle(thread)")
        finally:
            if not self.kernel.CloseHandle(snapshot):
                self._error("CloseHandle(snapshot)")

    def quiesce(self):
        if self.stopped:
            return
        if self.failure:
            raise UnsafeProcess(self.failure)
        try:
            if self.assigned:
                if not self.kernel.TerminateJobObject(self.handle, 1):
                    self._error("TerminateJobObject")
                deadline = time.monotonic() + 10
                while True:
                    accounting = self.accounting_type()
                    if not self.kernel.QueryInformationJobObject(
                            self.handle, 1, self.ctypes.byref(accounting),
                            self.ctypes.sizeof(accounting), None):
                        self._error("QueryInformationJobObject")
                    if accounting.ActiveProcesses == 0:
                        break
                    if time.monotonic() >= deadline:
                        raise UnsafeProcess("Windows job still has active children after termination")
                    time.sleep(0.01)
            elif self.process is not None:
                # Assignment failed while the sole owned process was suspended.
                if not self.kernel.TerminateProcess(int(self.process._handle), 1):
                    self._error("TerminateProcess(suspended child)")
            if self.process is not None:
                self.process.wait(timeout=10)
        except (UnsafeProcess, OSError, subprocess.TimeoutExpired) as exc:
            self.failure = "child-tree termination failed: " + str(exc)
        finally:
            try:
                self.close()
            except UnsafeProcess as exc:
                self.failure = str(exc)
            # Do not retain a Popen <-> job cycle (including its process handle
            # and any pipes left open by an assignment/resume failure).
            self.process = None
        if self.failure:
            raise UnsafeProcess(self.failure)
        self.stopped = True


def quiesce(process, force=False):
    """Stop descendants before restoration; failure is explicitly unsafe."""
    previous = signal.signal(signal.SIGINT, signal.SIG_IGN)
    try:
        if os.name == "nt":
            job = getattr(process, "_mutation_job", None)
            if job is None:
                raise UnsafeProcess("Windows child has no owned job; restoration refused")
            job.quiesce()
            return
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            return
        # A descendant can outlive its parent; do not wait only on cargo's PID.
        time.sleep(0.1)
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    finally:
        signal.signal(signal.SIGINT, previous)


def _runtime_result(row):
    """Separate Cargo build evidence from subsequent stdout-only libtest output."""
    selectors = [(kind, row["argv"].index("--" + kind)) for kind in
                 ("lib", "bin", "test", "example", "bench") if "--" + kind in row["argv"]]
    if len(selectors) != 1:
        row["error"] = "target requires exactly one Cargo target selector"
        return
    kind, position = selectors[0]
    name = None
    if kind != "lib":
        if position + 1 >= len(row["argv"]):
            row["error"] = "missing Cargo target name"
            return
        name = row["argv"][position + 1]
    lines = row["stdout"].splitlines()
    builds, artifacts = [], []
    malformed_build = False
    for position, line in enumerate(lines):
        try:
            message = json.loads(line)
        except ValueError:
            continue
        if not isinstance(message, dict):
            continue
        if message.get("reason") == "build-finished":
            success = message.get("success")
            if type(success) is not bool:
                malformed_build = True
            builds.append((position, success))
        elif message.get("reason") == "compiler-artifact":
            artifacts.append((position, message))
    if any(success is False for _, success in builds):
        row["build_success"] = False
        row["error"] = "Cargo reported an unsuccessful build"
        return
    if not builds or malformed_build:
        row["error"] = "missing or invalid Cargo build-finished marker"
        return
    row["build_success"] = True
    boundary = builds[-1][0]
    selected = []
    for position, artifact in artifacts:
        target, profile = artifact.get("target"), artifact.get("profile")
        if position >= boundary or not isinstance(target, dict) or not isinstance(profile, dict):
            continue
        kinds = target.get("kind")
        executable = artifact.get("executable")
        if (not isinstance(kinds, list) or profile.get("test") is not True
                or not isinstance(executable, str) or not executable):
            continue
        matches = (any(value in kinds for value in ("lib", "proc-macro")) if kind == "lib"
                   else kind in kinds and target.get("name") == name)
        if matches:
            selected.append(artifact)
    if len(selected) != 1:
        row["error"] = "expected exactly one selected test compiler-artifact with executable"
        return
    summaries = []
    for line in lines[boundary + 1:]:
        summary = re.fullmatch(r"test result: (ok|FAILED)\. (\d+) passed; (\d+) failed;[^\r\n]*", line)
        if summary:
            summaries.append(summary.groups())
    if not summaries:
        if row["exit"] == 0:
            row["status"] = "not_run"
        else:
            row["error"] = "target exited without a stdout libtest summary after the build"
        return
    if len(summaries) != 1:
        row["error"] = "multiple libtest summaries for one target"
        return
    summary, passed, failed = summaries[0]
    passed, failed = int(passed), int(failed)
    row.update(passed=passed, failed=failed, test_executed=passed + failed > 0)
    if passed == 0:
        row["status"] = "not_run"
    elif summary == "FAILED" and failed > 0 and row["exit"] != 0:
        row["status"] = "FAILED"
    elif summary == "ok" and failed == 0 and row["exit"] == 0:
        row["status"] = "survived"
    else:
        row["error"] = "libtest summary and target exit disagree"


def run_target(argv, workspace, env, timeout, *, clock=None, started=None, spawn=None, stop=None):
    clock = time.monotonic if clock is None else clock
    spawn = subprocess.Popen if spawn is None else spawn
    stop = quiesce if stop is None else stop
    row = {"argv": argv, "exit": None, "stdout": "", "stderr": "",
           "passed": None, "failed": None, "status": "error",
           "build_success": None, "test_executed": False, "timed_out": False}
    problem = None
    interrupted = False
    process = None
    windows_job = None
    stop_attempted = False
    stdout = stderr = None

    def stop_owned(force=False):
        nonlocal stop_attempted
        if stop_attempted:
            return
        stop_attempted = True
        try:
            stop(process, force=force)
        except OSError as exc:
            row["quiescent"] = False
            raise UnsafeProcess("child-tree termination failed: " + str(exc)) from exc
        except BaseException:
            row["quiescent"] = False
            raise
        row["quiescent"] = True

    try:
        if os.name == "nt":
            # Popen closes the initial thread handle. Keep SIGINT deferred until
            # the suspended child is owned and its primary thread is resumed.
            previous = signal.signal(signal.SIGINT, signal.SIG_IGN)
            try:
                windows_job = _WindowsJob()
                process = spawn(
                    argv, cwd=workspace, env=env,
                    creationflags=0x00000004 | subprocess.CREATE_NEW_PROCESS_GROUP,
                    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                process._mutation_job = windows_job
                windows_job.assign_and_resume(process)
            finally:
                signal.signal(signal.SIGINT, previous)
        else:
            process = spawn(argv, cwd=workspace, env=env, start_new_session=True,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        deadline = clock() + timeout
        try:
            if started is not None:
                started(process)
            remaining = deadline - clock()
            if remaining <= 0:
                row["timed_out"] = True
                raise subprocess.TimeoutExpired(argv, timeout)
            stdout, stderr = process.communicate(timeout=remaining)
        except subprocess.TimeoutExpired as exc:
            row["timed_out"] = True
            problem = "target timed out"
            row.update(stdout=(exc.output or b"").decode("utf-8", "replace"),
                       stderr=(exc.stderr or b"").decode("utf-8", "replace"))
            stop_owned(force=True)
            stdout, stderr = process.communicate()
        except KeyboardInterrupt:
            problem, interrupted = "target interrupted", True
            stop_owned(force=True)
            stdout, stderr = process.communicate()
        except OSError as exc:
            problem = str(exc)
            stop_owned(force=True)
            stdout, stderr = process.communicate()
        finally:
            if stdout is not None:
                row.update(stdout=stdout.decode("utf-8", "replace"),
                           stderr=stderr.decode("utf-8", "replace"))
            stop_owned()
    except UnsafeProcess as exc:
        row["quiescent"] = False
        problem = str(exc)
    except KeyboardInterrupt:
        problem, interrupted = "target interrupted", True
    except OSError as exc:
        problem = str(exc)
    finally:
        if process is not None:
            try:
                stop_owned(force=problem is not None)
            except (UnsafeProcess, OSError) as exc:
                row["quiescent"] = False
                problem = str(exc)
            except KeyboardInterrupt:
                row["quiescent"] = False
                problem, interrupted = "target interrupted during child-tree cleanup", True
            row["exit"] = process.returncode
        elif windows_job is not None:
            # A job created before a failed spawn still owns a handle to close.
            previous = signal.signal(signal.SIGINT, signal.SIG_IGN)
            try:
                windows_job.quiesce()
            except UnsafeProcess as exc:
                row["quiescent"] = False
                problem = str(exc)
            finally:
                signal.signal(signal.SIGINT, previous)
    if problem:
        row["error"] = problem
        return row, interrupted
    _runtime_result(row)
    return row, interrupted


def _tracked_files(workspace, env, entries=None, ancestors=()):
    """Snapshot actual entries, independently of Git's index shortcuts."""
    if entries is None:
        entries = _index_entries(git(workspace, env, "ls-files", "--stage", "-v", "-z").stdout)
    if not ancestors:
        ancestors = (_repository_identity(workspace, env),)
    algorithm = git(workspace, env, "rev-parse", "--show-object-format").stdout.strip().decode("ascii")
    if algorithm not in ("sha1", "sha256"):
        raise Blocked("unsupported Git object format: " + algorithm)
    tracked = {}
    for name in entries:
        kind, permissions, data = _read_tracked_entry(workspace, name)
        if data is not None:
            blob = hashlib.new(algorithm)
            blob.update(b"blob " + str(len(data)).encode("ascii") + b"\0")
            blob.update(data)
            oid = blob.hexdigest().encode("ascii")
        if kind == "file":
            tracked[name] = (kind, permissions, digest(data), oid)
        elif kind == "symlink":
            tracked[name] = (kind, os.fsdecode(data), oid)
        elif kind == "directory" and entries[name][1] == b"160000":
            tracked[name] = ("gitlink", permissions,
                             _gitlink_snapshot(workspace, name, env, ancestors))
        else:
            tracked[name] = (kind, permissions)
    return tracked


def restore(workspace, env, name, data, mode, index, head, head_bytes, tracked):
    try:
        write_regular(workspace, name, data, mode)
        diff = git(workspace, env, "diff", "--no-ext-diff", "--ignore-submodules=none",
                   "--quiet", "HEAD", "--", check=False)
        current, current_mode = read_regular(workspace, name)
        current_index = git(workspace, env, "ls-files", "--stage", "-v", "-z").stdout
        current_tracked = _tracked_files(workspace, env)
        restored = (diff.returncode == 0 and current == data and current_mode == mode
                    and current_index == index and current_tracked == tracked
                    and (workspace / ".git/HEAD").read_bytes() == head_bytes
                    and git(workspace, env, "rev-parse", "HEAD").stdout.strip().decode() == head
                    and git(workspace, env, "diff", "--cached", "--ignore-submodules=none",
                            "--quiet", "HEAD", check=False).returncode == 0)
        return restored, 0 if restored else (diff.returncode or 1), None
    except (OSError, Blocked) as exc:
        return False, 1, str(exc)


def execute(args, report):
    if os.name not in ("posix", "nt"):
        raise Blocked("unsupported host process and filesystem semantics")
    root = Path.cwd().resolve()
    scratch = root / "target" / ("mutation-run-" + uuid.uuid4().hex)
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update(HOME=str(scratch / "home"), XDG_CONFIG_HOME=str(scratch / "xdg"),
               XDG_CACHE_HOME=str(scratch / "cache"), XDG_DATA_HOME=str(scratch / "data"),
               XDG_STATE_HOME=str(scratch / "state"), TMPDIR=str(scratch / "tmp"),
               TMP=str(scratch / "tmp"), TEMP=str(scratch / "tmp"),
               CARGO_HOME=env.get("CARGO_HOME") or str(scratch / "cargo-home"),
               RUSTUP_HOME=env.get("RUSTUP_HOME") or str(scratch / "rustup-home"),
               GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=os.devnull,
               GIT_CONFIG_SYSTEM=os.devnull, GIT_OPTIONAL_LOCKS="0", GIT_TERMINAL_PROMPT="0")
    if os.name == "nt":
        env.update(USERPROFILE=str(scratch / "home"), APPDATA=str(scratch / "data"),
                   LOCALAPPDATA=str(scratch / "cache"))
    actual = Path(os.fsdecode(git(root, env, "rev-parse", "--show-toplevel").stdout.rstrip(b"\n"))).resolve()
    if actual != root:
        raise Blocked("run from the actual source Git root")
    before = source_snapshot(root, env)
    report["head"] = before[0]
    try:
        require_clean_source(before)
        value = os.environ.get("CARGO_TARGET_DIR")
        if not value:
            raise Blocked("CARGO_TARGET_DIR is required")
        owned_root = root / "target"
        if owned_root.is_symlink() or owned_root.resolve() != owned_root:
            raise Blocked("source target directory must not redirect elsewhere")
        owned = Path(value).resolve()
        if owned_root not in owned.parents:
            raise Blocked("CARGO_TARGET_DIR must resolve strictly below source target/")
        for artifact in (scratch, owned / "mutation-run-probe"):
            if git(root, env, "check-ignore", "--quiet", "--no-index", "--",
                   str(artifact), check=False).returncode != 0:
                raise Blocked("retained artifact paths must be ignored by source Git")
        with open(args.spec, encoding="utf-8") as handle:
            spec = json.load(handle)
        prepared, timeout = validate(spec, root, env)
        # No target or scratch directory is created until the entire spec passes.
        scratch.mkdir(parents=True)
        for name in ("home", "xdg", "cache", "data", "state", "tmp", "cargo-home", "rustup-home"):
            (scratch / name).mkdir()
        owned.mkdir(parents=True, exist_ok=True)
        cargo_target = owned / ("run-" + uuid.uuid4().hex)
        cargo_target.mkdir()
        workspace = scratch / "workspace"
        report.update(workspace=str(workspace), cargo_target_dir=str(cargo_target))
        env["CARGO_TARGET_DIR"] = str(cargo_target)
        git(root, env, "-c", "protocol.allow=never", "-c", "protocol.file.allow=always",
            "clone", "--depth", "1", "--no-local", "--no-hardlinks", "--no-checkout",
            "--", str(root), str(workspace))
        git(workspace, env, "checkout", "--detach", before[0])
        if (workspace / ".git/objects/info/alternates").exists():
            raise Blocked("clone unexpectedly uses alternate objects")
        clone_index = git(workspace, env, "ls-files", "--stage", "-v", "-z").stdout
        # Fresh detached checkout supplies the HEAD working-tree baseline.
        # Unlike git diff, this content snapshot cannot hide assume-unchanged
        # or skip-worktree edits, and ignores harmless index stat-cache churn.
        clone_tracked = _tracked_files(workspace, env)
        clone_head_bytes = (workspace / ".git/HEAD").read_bytes()
        for mutation, original, replacement, _source_mode, commands in prepared:
            name = mutation["file"]
            # Git preserves executable bits, not all source permission bits.
            mode = clone_tracked[name][1]
            row = {"id": mutation["id"], "file": name, "diff": "", "targets": [],
                   "restored": False, "restore_diff_exit": 1}
            report["mutations"].append(row)
            interrupted = False
            try:
                current, _ = read_regular(workspace, name)
                if current != original:
                    raise Blocked("clone bytes do not match saved HEAD: " + name)
                write_regular(workspace, name, replacement, mode)
                diff = git(workspace, env, "diff", "--no-ext-diff", "--no-textconv", "--no-color",
                           "--no-renames", "-U0", "HEAD", "--", name).stdout.decode("utf-8", "replace")
                row["diff"] = diff
                removed = [line[1:] for line in diff.splitlines() if line.startswith("-") and not line.startswith("---")]
                added = [line[1:] for line in diff.splitlines() if line.startswith("+") and not line.startswith("+++")]
                if removed != [mutation["old"]] or added != [mutation["new"]]:
                    raise Blocked("actual diff is not exactly the specified one-line mutation")
                for argv in commands:
                    target, interrupted = run_target(argv, workspace, env, timeout)
                    row["targets"].append(target)
                    if target["status"] in ("not_run", "error"):
                        report["errors"].append("{}: target {}".format(mutation["id"], target["status"]))
                    if interrupted or target.get("quiescent") is False:
                        break
            finally:
                if any(target.get("quiescent") is False for target in row["targets"]):
                    restored, diff_exit, error = False, 1, "child tree not proven stopped; restoration refused"
                else:
                    restored, diff_exit, error = restore(
                        workspace, env, name, original, mode, clone_index, before[0],
                        clone_head_bytes, clone_tracked)
                row.update(restored=restored, restore_diff_exit=diff_exit)
                if not restored:
                    report["errors"].append("{}: restoration failed{}".format(mutation["id"], ": " + error if error else ""))
                    report["status"] = "blocked"
            if not row["restored"] or interrupted:
                break
        if report["status"] != "blocked":
            report["status"] = "error" if report["errors"] else "completed"
    finally:
        try:
            report["source_unchanged"] = source_snapshot(root, env) == before
        except (OSError, Blocked):
            report["source_unchanged"] = False
        if not report["source_unchanged"]:
            report["errors"].append("source HEAD/index/tracked state/refs changed; source was not repaired")
            report["status"] = "blocked"


def main():
    report = {"schema_version": 1, "head": None, "workspace": None,
              "cargo_target_dir": None, "source_unchanged": False,
              "mutations": [], "status": "error", "errors": []}
    try:
        parser = Parser(description=__doc__, add_help=False)
        parser.add_argument("--spec", required=True)
        args = parser.parse_args()
        execute(args, report)
    except Blocked as exc:
        report["status"] = "blocked"
        report["errors"].append(str(exc))
    except KeyboardInterrupt:
        report["status"] = "error"
        report["errors"].append("interrupted")
    except Exception as exc:
        report["status"] = "error"
        report["errors"].append("{}: {}".format(type(exc).__name__, exc))
    print(json.dumps(report, ensure_ascii=True))
    return 0 if report["status"] == "completed" else 1


if __name__ == "__main__":
    sys.exit(main())
