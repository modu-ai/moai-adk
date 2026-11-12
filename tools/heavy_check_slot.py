#!/usr/bin/env python3
"""Limit how many heavy checks run at once on a shared machine (Issue #2491).

    heavy_check_slot.py acquire --max N --name LANE [--dir DIR] -- COMMAND...
    heavy_check_slot.py status [--dir DIR]

`acquire` waits for one of N slots, runs COMMAND, and returns COMMAND's exit code
(128 + signal number when COMMAND ends on a signal). While waiting it prints one line
each time its place in the queue changes. `status` lists the slots held now. It only
reads; it never ends or reorders another lane's work.

A slot is a file `slot-<i>` with an exclusive `fcntl.flock` lock. The lock lives on
the open file, so the descriptor is also handed to COMMAND: the slot stays held while
either this wrapper or COMMAND is alive, and the operating system frees it when both
are gone, even after SIGKILL. No stale slot is left behind to clean up.

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
import signal
import subprocess
import sys
import time

try:
    import fcntl
except ImportError:  # pragma: no cover - non-unix
    fcntl = None

POLL_SECONDS = 0.2
FORWARDED = (signal.SIGINT, signal.SIGTERM, signal.SIGHUP)


def default_dir():
    base = os.environ.get("XDG_CACHE_HOME") or os.path.join(os.path.expanduser("~"), ".cache")
    return os.path.join(base, "moai-cli", "heavy-slots")


def try_lock(path, mode):
    """Open `path` and take `mode` without waiting; return the fd or None if held."""
    fd = os.open(path, os.O_RDWR | os.O_CREAT, 0o600)
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
        with open(path, "rb") as f:
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
    """1-based place of `me` among live waiters, oldest first (by file name)."""
    live = []
    for name in sorted(os.listdir(waiting)):
        path = os.path.join(waiting, name)
        if name == me or is_held(path):
            live.append(name)
    return live.index(me) + 1, len(live)


def acquire(folder, maximum, name, command):
    waiting = os.path.join(folder, "waiting")
    os.makedirs(waiting, mode=0o700, exist_ok=True)
    slot, fd = take_slot(folder, maximum)
    if fd is None:
        me = f"{time.time_ns():020d}-{os.getpid()}"
        wait_path = os.path.join(waiting, me)
        wait_fd = try_lock(wait_path, fcntl.LOCK_EX)
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
    slots = []
    if os.path.isdir(folder):
        for name in sorted(os.listdir(folder)):
            path = os.path.join(folder, name)
            if name.startswith("slot-") and is_held(path):
                slots.append(read_record(path))
    print(json.dumps({"folder": folder, "held": slots}, indent=1, sort_keys=True))
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = parser.add_subparsers(dest="op", required=True)
    a = sub.add_parser("acquire", help="wait for a slot, run COMMAND, release")
    a.add_argument("--max", type=int, required=True)
    a.add_argument("--name", required=True)
    a.add_argument("--dir", default=None)
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
    if args.max < 1 or not command:
        parser.error("acquire needs --max >= 1 and a command after --")
    os.makedirs(folder, mode=0o700, exist_ok=True)
    return acquire(folder, args.max, args.name, command)


if __name__ == "__main__":
    sys.exit(main())
