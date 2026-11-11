#!/usr/bin/env python3
"""Run explicit one-line mutations in a retained, independent HEAD clone."""

import argparse
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
    result = subprocess.run(
        ["git", "--no-optional-locks", "-c", "core.fsmonitor=false",
         "-c", "maintenance.auto=false",
         "-c", "gc.auto=0", *args], cwd=root, env=env,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if check and result.returncode:
        raise Blocked("git {}: {}".format(args[0], result.stderr.decode("utf-8", "replace").strip()))
    return result


def digest(data):
    return hashlib.sha256(data).hexdigest()


def source_snapshot(root, env):
    head = git(root, env, "rev-parse", "--verify", "HEAD^{commit}").stdout.strip().decode("ascii")
    index = Path(os.fsdecode(git(root, env, "rev-parse", "--git-path", "index").stdout.rstrip(b"\n")))
    if not index.is_absolute():
        index = root / index
    tracked = {}
    for raw in git(root, env, "ls-files", "-z").stdout.split(b"\0"):
        if not raw:
            continue
        name = os.fsdecode(raw)
        path = root / name
        mode = path.lstat().st_mode
        if stat.S_ISLNK(mode):
            value = (mode, os.readlink(path))
        elif stat.S_ISREG(mode):
            value = (mode, digest(path.read_bytes()))
        elif stat.S_ISDIR(mode):
            value = (mode, "directory")
        else:
            value = (mode, "nonregular")
        tracked[name] = value
    return (head, index.read_bytes(), tracked,
            git(root, env, "status", "--porcelain=v1", "--untracked-files=all").stdout,
            git(root, env, "show-ref", check=False).stdout)


def regular_fd(root, relative, flags=os.O_RDONLY):
    """Traverse only directories and a regular leaf, without following symlinks."""
    if os.name == "nt":
        return windows_regular_fd(root, relative, flags)
    directory = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        parts = PurePosixPath(relative).parts
        for part in parts[:-1]:
            child = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=directory)
            os.close(directory)
            directory = child
        fd = os.open(parts[-1], flags | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=directory)
        if not stat.S_ISREG(os.fstat(fd).st_mode):
            os.close(fd)
            raise Blocked("mutation path is not a regular file: " + relative)
        return fd
    finally:
        os.close(directory)


def windows_regular_fd(root, relative, flags):
    """Open a Windows leaf without following reparse points or junctions."""
    import ctypes
    from ctypes import wintypes
    import msvcrt

    path = Path(root)
    for part in (None, *PurePosixPath(relative).parts):
        if part is not None:
            path = path / part
        info = path.lstat()
        if getattr(info, "st_file_attributes", 0) & 0x400:
            raise Blocked("reparse point in mutation path: " + str(path))
        if path != Path(root) / relative and not stat.S_ISDIR(info.st_mode):
            raise Blocked("non-directory mutation ancestor: " + str(path))
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
            command = ["cargo", "test", "--offline", "--locked", "--" + kind]
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


def run_target(argv, workspace, env, timeout):
    row = {"argv": argv, "exit": None, "stdout": "", "stderr": "",
           "passed": None, "failed": None, "status": "error"}
    problem = None
    interrupted = False
    process = None
    windows_job = None
    try:
        if os.name == "nt":
            # Popen closes the initial thread handle. Keep SIGINT deferred until
            # the suspended child is owned and its primary thread is resumed.
            previous = signal.signal(signal.SIGINT, signal.SIG_IGN)
            try:
                windows_job = _WindowsJob()
                process = subprocess.Popen(
                    argv, cwd=workspace, env=env,
                    creationflags=0x00000004 | subprocess.CREATE_NEW_PROCESS_GROUP,
                    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                process._mutation_job = windows_job
                windows_job.assign_and_resume(process)
            finally:
                signal.signal(signal.SIGINT, previous)
        else:
            process = subprocess.Popen(argv, cwd=workspace, env=env, start_new_session=True,
                                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            stdout, stderr = process.communicate(timeout=timeout)
        except subprocess.TimeoutExpired as exc:
            problem = "target timed out"
            row.update(stdout=(exc.output or b"").decode("utf-8", "replace"),
                       stderr=(exc.stderr or b"").decode("utf-8", "replace"))
            quiesce(process, force=True)
            stdout, stderr = process.communicate()
        except KeyboardInterrupt:
            problem, interrupted = "target interrupted", True
            quiesce(process, force=True)
            stdout, stderr = process.communicate()
        finally:
            quiesce(process)
        row.update(exit=process.returncode, stdout=stdout.decode("utf-8", "replace"),
                   stderr=stderr.decode("utf-8", "replace"))
    except UnsafeProcess as exc:
        row["quiescent"] = False
        problem = str(exc)
    except KeyboardInterrupt:
        if os.name != "nt":
            raise
        # Also cover an interrupt between Windows setup and communicate().
        problem, interrupted = "target interrupted", True
    except OSError as exc:
        if process is not None:
            row["quiescent"] = False
        problem = str(exc)
    finally:
        if windows_job is not None:
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
    summaries = re.findall(r"^test result: (ok|FAILED)\. (\d+) passed; (\d+) failed;[^\r\n]*$",
                           row["stdout"] + "\n" + row["stderr"], re.MULTILINE)
    if summaries:
        summary, passed, failed = summaries[-1]
        row.update(passed=int(passed), failed=int(failed))
        if len(summaries) != 1:
            row["error"] = "multiple libtest summaries for one target"
        elif int(passed) == 0:
            row["status"] = "not_run"
        elif summary == "FAILED" and int(failed) > 0 and row["exit"] != 0:
            row["status"] = "FAILED"
        elif summary == "ok" and int(failed) == 0 and row["exit"] == 0:
            row["status"] = "survived"
    elif row["exit"] == 0:
        row["status"] = "not_run"
    return row, interrupted


def _tracked_files(workspace, env):
    """Compare actual tracked content, independently of Git's index shortcuts."""
    tracked = {}
    for raw in git(workspace, env, "ls-files", "-z").stdout.split(b"\0"):
        if not raw:
            continue
        name = os.fsdecode(raw)
        path = workspace / name
        mode = path.lstat().st_mode
        if stat.S_ISREG(mode):
            data, permissions = read_regular(workspace, name)
            tracked[name] = ("file", permissions, digest(data))
        elif stat.S_ISLNK(mode):
            tracked[name] = ("symlink", os.readlink(path))
        elif stat.S_ISDIR(mode):
            tracked[name] = ("directory", stat.S_IMODE(mode))
        else:
            raise Blocked("tracked path is no longer regular: " + name)
    return tracked


def restore(workspace, env, name, data, mode, index, head, head_bytes, tracked):
    try:
        write_regular(workspace, name, data, mode)
        diff = git(workspace, env, "diff", "--no-ext-diff", "--quiet", "HEAD", "--", check=False)
        current, current_mode = read_regular(workspace, name)
        current_index = git(workspace, env, "ls-files", "--stage", "-v", "-z").stdout
        current_tracked = _tracked_files(workspace, env)
        restored = (diff.returncode == 0 and current == data and current_mode == mode
                    and current_index == index and current_tracked == tracked
                    and (workspace / ".git/HEAD").read_bytes() == head_bytes
                    and git(workspace, env, "rev-parse", "HEAD").stdout.strip().decode() == head
                    and git(workspace, env, "diff", "--cached", "--quiet", "HEAD", check=False).returncode == 0)
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
        if before[3]:
            raise Blocked("source index/worktree must be clean, including untracked files")
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
