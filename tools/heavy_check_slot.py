#!/usr/bin/env python3
"""Limit how many heavy checks run at once on a shared machine (Issue #2491).

    heavy_check_slot.py acquire [--max N] --name LANE [--dir DIR]
                                [--min-free-gb G] [--max-swap-pct P] -- COMMAND...
    heavy_check_slot.py status [--dir DIR]

Machine limits (Issue #2829, the same rule as the Rust `heavy-slot-acquire`, src/heavy_slot.rs):
when the working folder's file system has less than --min-free-gb free (default 50, decimal
GB = 10^9 bytes) or swap use is above --max-swap-pct (default 80), `acquire` starts nothing
and exits 75 (sysexits EX_TEMPFAIL, "try again later") with the measured values on stderr.
It judges before queueing and again right before COMMAND after a wait, since a wait can eat
the space. A value the system does not give is not a reason to refuse. Nothing is ever
deleted (cleanup is a person's job).

The machine's limits live in one file beside the slots, `<folder>/limits.json` (Issue #2930
R-2): `max`, `min_free_gb`, `max_swap_pct`, each optional (defaults 3, 50, 80). A caller's
flag wins only when it is stricter, so no caller turns a limit off; only the file loosens
one. A limits file that is not a plain file (a symlink, even to a valid file, or a folder),
is over 4096 bytes, does not parse, has another key or sets `max` 0 is refused: `acquire`
exits 125 without running COMMAND or writing anything.
Tests set MOAI_CLI_TEST_MACHINE=free_gb=<n>,swap_pct=<n> (or `@<file>` holding that text,
read on every check) — the same test-only input the Rust debug build reads.

`acquire` waits for one of N slots, runs COMMAND, and returns COMMAND's exit code
(128 + signal number when COMMAND ends on a signal). While waiting it prints one line
each time its place in the queue changes. `status` lists the slots held now. It only
reads; it never ends or reorders another lane's work.

A slot or queue path that is a symlink, FIFO, device, folder or hard link is refused
before anything is written: `acquire` exits 125 (the wrapper's own failure, apart from
COMMAND's codes) without running COMMAND, and `status` lists it under "refused".

A slot is a file `slot-<i>` with an exclusive `fcntl.flock` lock. The lock lives on
the open file, so the descriptor is also handed to COMMAND: the slot stays held while
either this wrapper or COMMAND is alive, and the operating system frees it when both
are gone, even after SIGKILL. No stale slot is left behind to clean up. The record
(name, command, pid) is written just after the lock is taken, so for a moment `status`
can list a held slot with an empty record (#2599).

Every caller on one machine should pass the same --max; a smaller --max only looks at
the first slots. The default folder is ${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots.
Unix only (flock); elsewhere the tool refuses with exit 2.

Sources read 2026-11-12: Python `fcntl.flock` (3.13 built-in help: "See the Unix manual
page for flock(2)") and macOS flock(2): descriptors "duplicated through dup(2) or
fork(2) do not result in multiple instances of a lock, but rather multiple references
to a single lock".
"""

import argparse
import json
import os
import re
import signal
import stat
import subprocess
import sys
import time

try:
    import fcntl
except ImportError:  # pragma: no cover - non-unix
    fcntl = None

POLL_SECONDS = 0.2
# The wrapper's own refusal, kept apart from COMMAND's codes like env(1) and timeout(1) do.
REFUSED = 125
# Too short of disk or too deep in swap to start now (src/heavy_slot.rs BUSY).
BUSY = 75
# The leader's numbers (Issue #2677 comment 6850581277), src/heavy_slot.rs DEFAULT_LIMITS.
DEFAULT_MIN_FREE_GB = 50
DEFAULT_MAX_SWAP_PCT = 80
# Issue #2930 R-2 (src/heavy_slot.rs DEFAULT_MAX, LIMITS_FILE).
DEFAULT_MAX = 3
LIMITS_FILE = "limits.json"
LIMITS_FILE_MAX_BYTES = 4096
LIMIT_KEYS = {"max": 2**64 - 1, "min_free_gb": 2**64 - 1, "max_swap_pct": 2**64 - 1}
FORWARDED = (signal.SIGINT, signal.SIGTERM, signal.SIGHUP)


def _test_machine():
    """(free_bytes, swap_pct) from MOAI_CLI_TEST_MACHINE, or None when it is not set."""
    text = os.environ.get("MOAI_CLI_TEST_MACHINE")
    if text is None:
        return None
    if text.startswith("@"):
        try:
            with open(text[1:]) as f:
                text = f.read()
        except OSError:
            text = ""
    free = swap = None
    for part in text.strip().split(","):
        key, _, value = part.partition("=")
        try:
            if key == "free_gb":
                free = int(value) * 1_000_000_000
            elif key == "swap_pct":
                swap = int(value)
        except ValueError:
            pass
    return free, swap


def free_bytes(folder):
    try:
        s = os.statvfs(folder)
    except OSError:
        return None
    return s.f_bavail * s.f_frsize


def swap_pct():
    """Swap in use, percent of the swap size (0 without swap); None when unknown."""
    if sys.platform == "darwin":
        try:
            out = subprocess.run(["/usr/sbin/sysctl", "-n", "vm.swapusage"], capture_output=True,
                                 text=True, timeout=5).stdout
        except (OSError, subprocess.SubprocessError):
            return None
        fields = dict(re.findall(r"(\w+) = ([\d.]+)M", out))
        try:
            total, used = float(fields["total"]), float(fields["used"])
        except (KeyError, ValueError):
            return None
        return 0 if total == 0 else int(used * 100 // total)
    try:
        with open("/proc/meminfo") as f:
            info = dict(line.split(":", 1) for line in f if ":" in line)
        total = int(info["SwapTotal"].split()[0])
        free = int(info["SwapFree"].split()[0])
    except (OSError, KeyError, ValueError, IndexError):
        return None
    return 0 if total == 0 else (total - free) * 100 // total


def over_limits(folder, min_free_gb, max_swap_pct):
    """Why a heavy run must not start now, or None (same text as src/heavy_slot.rs)."""
    injected = _test_machine()
    free, swap = injected if injected is not None else (free_bytes(folder), swap_pct())
    why = []
    if free is not None and free < min_free_gb * 1_000_000_000:
        why.append(f"disk free {free / 1e9:.1f} GB at {folder} is below {min_free_gb} GB")
    if swap is not None and swap > max_swap_pct:
        why.append(f"swap use {swap}% is above {max_swap_pct}%")
    return "; ".join(why) or None


class BadLimitsFile(Exception):
    """A limits file that cannot be used (exit REFUSED, COMMAND not run)."""


def read_limits_file(folder):
    """The limits file in `folder` as a dict, or None when there is none (src/heavy_slot.rs
    read_limits_file): a plain file of at most 4096 bytes, an object of known keys holding
    whole numbers (or null), `max` at least 1."""
    path = os.path.join(folder, LIMITS_FILE)
    try:
        st = os.lstat(path)
    except FileNotFoundError:
        return None
    except OSError as e:
        raise BadLimitsFile(f"{path}: {e}") from None
    if not stat.S_ISREG(st.st_mode):
        raise BadLimitsFile(f"{path}: not a plain file (symlink, FIFO, device or folder)")
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        with os.fdopen(fd, "rb") as f:
            data = f.read(LIMITS_FILE_MAX_BYTES + 1)
    except OSError as e:
        raise BadLimitsFile(f"{path}: {e}") from None
    if len(data) > LIMITS_FILE_MAX_BYTES:
        raise BadLimitsFile(f"{path}: larger than {LIMITS_FILE_MAX_BYTES} bytes")
    try:
        doc = json.loads(data.decode("utf-8"))
    except ValueError as e:
        raise BadLimitsFile(f"{path}: {e}") from None
    if not isinstance(doc, dict):
        raise BadLimitsFile(f"{path}: not a JSON object")
    for key, value in doc.items():
        if key not in LIMIT_KEYS:
            raise BadLimitsFile(f"{path}: unknown field `{key}`")
        if value is not None and (type(value) is not int or not 0 <= value <= LIMIT_KEYS[key]):
            raise BadLimitsFile(f"{path}: `{key}` must be a whole number from 0")
    if doc.get("max") == 0:
        raise BadLimitsFile(f"{path}: max must be at least 1")
    return doc


def effective(file, caller_max, caller_free, caller_swap):
    """(max, min_free_gb, max_swap_pct, notes): the stricter of the file (or the defaults) and
    the caller's flags, one note per flag a stricter value overrode (src/heavy_slot.rs
    effective)."""
    file = file or {}
    source = LIMITS_FILE if file else "the default"
    base_max = file.get("max") if file.get("max") is not None else DEFAULT_MAX
    base_free = file.get("min_free_gb")
    base_free = DEFAULT_MIN_FREE_GB if base_free is None else base_free
    base_swap = file.get("max_swap_pct")
    base_swap = DEFAULT_MAX_SWAP_PCT if base_swap is None else base_swap
    notes = []

    def pick(flag, given, base, stricter):
        if given is None:
            return base
        if stricter(given, base) or given == base:
            return given
        notes.append(f"heavy-check-slot: {flag} {given} is looser than {source} ({base}); "
                     f"using {base}")
        return base

    maximum = pick("--max", caller_max, base_max, lambda a, b: a < b)
    free = pick("--min-free-gb", caller_free, base_free, lambda a, b: a > b)
    swap = pick("--max-swap-pct", caller_swap, base_swap, lambda a, b: a < b)
    return maximum, free, swap, notes


def machine_line(folder):
    try:
        load = "/".join(f"{x:.2f}" for x in os.getloadavg())
    except OSError:
        load = "unknown"
    free = free_bytes(folder)
    disk = "unknown" if free is None else f"{free / (1 << 30):.1f} GiB"
    return f"load {load}; disk free {disk} at {folder}"


def busy(name, why, folder):
    print(f"heavy-check-slot: {name} not started — {why}; {machine_line(folder)}; COMMAND did "
          f"not run, try again later", file=sys.stderr, flush=True)
    return BUSY


def default_dir():
    base = os.environ.get("XDG_CACHE_HOME") or os.path.join(os.path.expanduser("~"), ".cache")
    return os.path.join(base, "moai-cli", "heavy-slots")


class NotASlotFile(Exception):
    """A slot or queue path that is not a plain file of its own (exit REFUSED)."""


def open_plain(path, flags):
    """Open `path` without following a symlink and only as a regular file with one link.

    A symlink, FIFO, device, folder or hard-linked file at a slot path is refused before
    anything is written, so the tool never truncates a file somewhere else (audit B1 of
    PR #2571). O_NONBLOCK keeps a FIFO from blocking the open itself.
    """
    try:
        fd = os.open(path, flags | os.O_NOFOLLOW | os.O_NONBLOCK, 0o600)
    except OSError as e:
        try:
            plain = stat.S_ISREG(os.lstat(path).st_mode)
        except OSError:
            raise e from None
        if not plain:
            raise NotASlotFile(path) from e
        raise
    try:
        st = os.fstat(fd)
        if not stat.S_ISREG(st.st_mode) or st.st_nlink != 1:
            raise NotASlotFile(path)
        os.set_blocking(fd, True)
    except BaseException:
        os.close(fd)
        raise
    return fd


def try_lock(path, mode, create=os.O_CREAT):
    """Open `path` and take `mode` without waiting; return the fd or None if held."""
    fd = open_plain(path, os.O_RDWR | create)
    try:
        fcntl.flock(fd, mode | fcntl.LOCK_NB)
    except BlockingIOError:
        os.close(fd)
        return None
    except BaseException:
        os.close(fd)
        raise
    return fd


def write_record(fd, record):
    os.ftruncate(fd, 0)
    os.pwrite(fd, json.dumps(record, sort_keys=True).encode() + b"\n", 0)


def read_record(path):
    try:
        fd = open_plain(path, os.O_RDONLY)
    except (OSError, NotASlotFile):
        return {}
    try:
        with os.fdopen(fd, "rb") as f:
            return json.loads(f.read() or b"{}")
    except (OSError, ValueError):
        return {}


def is_held(path):
    """True when another process holds the slot's lock (checked with a shared try-lock)."""
    fd = try_lock(path, fcntl.LOCK_SH)
    if fd is None:
        return True
    os.close(fd)
    return False


def take_slot(folder, maximum):
    for i in range(maximum):
        fd = try_lock(os.path.join(folder, f"slot-{i}"), fcntl.LOCK_EX)
        if fd is not None:
            return i, fd
    return None, None


def queue_place(waiting, me):
    """1-based place of `me` among live waiters, oldest first (by file name).

    Another entry that is not a plain file is no waiter: it is skipped, never written.
    """
    live = []
    for name in sorted(os.listdir(waiting)):
        path = os.path.join(waiting, name)
        try:
            alive = name == me or is_held(path)
        except NotASlotFile:
            alive = False
        if alive:
            live.append(name)
    return live.index(me) + 1, len(live)


def acquire(folder, maximum, name, command, min_free_gb=DEFAULT_MIN_FREE_GB,
            max_swap_pct=DEFAULT_MAX_SWAP_PCT):
    cwd = os.getcwd()
    why = over_limits(cwd, min_free_gb, max_swap_pct)
    if why:
        return busy(name, why, cwd)
    waiting = os.path.join(folder, "waiting")
    os.makedirs(waiting, mode=0o700, exist_ok=True)
    slot, fd = take_slot(folder, maximum)
    waited = fd is None
    if fd is None:
        me = f"{time.time_ns():020d}-{os.getpid()}"
        wait_path = os.path.join(waiting, me)
        wait_fd = try_lock(wait_path, fcntl.LOCK_EX, create=os.O_CREAT | os.O_EXCL)
        last = None
        try:
            while fd is None:
                place = queue_place(waiting, me)
                if place != last:
                    print(f"heavy-check-slot: waiting {place[0]} of {place[1]} for {maximum} slot(s)",
                          flush=True)
                    last = place
                time.sleep(POLL_SECONDS)
                slot, fd = take_slot(folder, maximum)
        finally:
            os.close(wait_fd)
            try:
                os.unlink(wait_path)
            except FileNotFoundError:
                pass
    if waited:
        # The machine may have changed during the wait (#2829 D2): judge it again before
        # starting; closing the slot file frees the slot.
        why = over_limits(cwd, min_free_gb, max_swap_pct)
        if why:
            os.close(fd)
            return busy(name, why, cwd)
    # Handlers go in before the slot is announced and the command starts, so no signal lands in
    # a gap where the wrapper would die and leave the command running; one that arrives before
    # the command exists is held and passed on as soon as it does.
    held = {"child": None, "pending": []}

    def forward(sig, _frame):
        if held["child"] is None:
            held["pending"].append(sig)
        else:
            held["child"].send_signal(sig)

    for sig in FORWARDED:
        signal.signal(sig, forward)
    write_record(fd, {"name": name, "command": command, "pid": os.getpid(),
                      "started_unix": int(time.time()), "slot": slot})
    print(f"heavy-check-slot: slot {slot} taken by {name}", flush=True)
    child = subprocess.Popen(command, pass_fds=(fd,))
    held["child"] = child
    for sig in held["pending"]:
        child.send_signal(sig)
    code = child.wait()
    os.close(fd)
    return 128 - code if code < 0 else code


def status(folder):
    slots, refused = [], []
    if os.path.isdir(folder):
        for name in sorted(os.listdir(folder)):
            path = os.path.join(folder, name)
            if not name.startswith("slot-"):
                continue
            try:
                if is_held(path):
                    slots.append(read_record(path))
            except NotASlotFile:
                refused.append(name)
    print(json.dumps({"folder": folder, "held": slots, "refused": refused}, indent=1,
                     sort_keys=True))
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = parser.add_subparsers(dest="op", required=True)
    a = sub.add_parser("acquire", help="wait for a slot, run COMMAND, release")
    a.add_argument("--max", type=int, default=None,
                   help="slot count; only a value below limits.json's (default 3) takes effect")
    a.add_argument("--name", required=True)
    a.add_argument("--dir", default=None)
    a.add_argument("--min-free-gb", type=int, default=None,
                   help="start nothing below this much free space (decimal GB); only a value "
                        "above limits.json's (default 50) takes effect")
    a.add_argument("--max-swap-pct", type=int, default=None,
                   help="start nothing above this swap use in percent; only a value below "
                        "limits.json's (default 80) takes effect")
    a.add_argument("command", nargs=argparse.REMAINDER)
    s = sub.add_parser("status", help="list held slots (read only)")
    s.add_argument("--dir", default=None)
    args = parser.parse_args(argv)
    if fcntl is None:
        print("heavy-check-slot: fcntl.flock is unavailable on this platform", file=sys.stderr)
        return 2
    folder = args.dir or default_dir()
    if args.op == "status":
        return status(folder)
    command = args.command[1:] if args.command[:1] == ["--"] else args.command
    flags = (args.max, args.min_free_gb, args.max_swap_pct)
    if any(v is not None and v < 0 for v in flags):
        parser.error("--max, --min-free-gb and --max-swap-pct take whole numbers from 0")
    # Issue #2930 R-2: the machine's limits file beside the slots; a flag only tightens it.
    try:
        file = read_limits_file(folder)
    except BadLimitsFile as e:
        print(f"heavy-check-slot: refused {e}; nothing was written and COMMAND did not run",
              file=sys.stderr)
        return REFUSED
    maximum, min_free_gb, max_swap_pct, notes = effective(file, *flags)
    for note in notes:
        print(note, file=sys.stderr)
    if maximum < 1 or not command:
        parser.error("acquire needs --max >= 1 and a command after --")
    os.makedirs(folder, mode=0o700, exist_ok=True)
    try:
        return acquire(folder, maximum, args.name, command, min_free_gb, max_swap_pct)
    except NotASlotFile as e:
        print(f"heavy-check-slot: refused {e}: not a plain file (symlink, FIFO, device, folder "
              f"or hard link); nothing was written and COMMAND did not run — remove it or pass "
              f"another --dir", file=sys.stderr)
        return REFUSED


if __name__ == "__main__":
    sys.exit(main())
