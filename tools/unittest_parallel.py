#!/usr/bin/env python3
"""Run the Python test modules of a folder in parallel, one process per module.

SPEC-20261120-parallel-python-tests. `python3 tools/unittest_parallel.py -v -s tests -p '*_test.py'`
takes the options of `python3 -m unittest discover` that the repository uses (`-v`, `-s`, `-p`)
and runs every matching module in its own `python3 -m unittest discover -p <module file>`
process, a few at a time. A module file bigger than 32 KiB is split by test class: unittest
itself lists the classes the module's tests belong to, and each class runs in its own process
(`python3 -m unittest <module>.<Class>`), so one long module does not decide the whole run's
length. A process gets its own temporary folder as `HOME` and `TMPDIR`; the caller's
`CARGO_HOME` and `RUSTUP_HOME` stay (or, when unset, keep the caller's defaults), so a test
that runs cargo still finds the toolchain. Each process's output is printed as one block when
it ends, so the lines `name (module.Class.name) ... ok` keep their shape, and the run ends
with the summary of a serial run: `Ran N tests in S.SSSs`, then `OK` or `FAILED (...)`. The
exit code is 0 only when every process passed. Standard library only.
"""

import argparse
import concurrent.futures
import fnmatch
import os
import re
import shutil
import subprocess
import sys
import tempfile
import threading
import time

RAN = re.compile(r"^Ran (\d+) tests? in [0-9.]+s$", re.M)
RESULT = re.compile(r"^(OK|FAILED)(?: \(([^)]*)\))?$", re.M)
# Keys of the parenthesis `unittest` prints, in its order.
KEYS = ("failures", "errors", "skipped", "expected failures", "unexpected successes")
# A module file bigger than this is split by test class.
SPLIT_BYTES = 32 * 1024
# Prints the test classes of a module, one per line, in the order its tests come; exits 3 when a
# test belongs to another module (a `load_tests` hook) or the module does not import.
LIST_CLASSES = """
import sys, unittest
sys.path.insert(0, sys.argv[1])
name = sys.argv[2]
def walk(suite):
    for t in suite:
        if isinstance(t, unittest.TestSuite):
            yield from walk(t)
        else:
            yield t
seen = []
for t in walk(unittest.defaultTestLoader.loadTestsFromName(name)):
    if type(t).__module__ != name:
        raise SystemExit(3)
    if type(t).__name__ not in seen:
        seen.append(type(t).__name__)
print("\\n".join(seen))
"""

OUT_LOCK = threading.Lock()


def say(text):
    """One flushed write to stderr, where `unittest` writes its report."""
    with OUT_LOCK:
        sys.stderr.write(text)
        sys.stderr.flush()


def default_jobs():
    return max(1, min(4, os.cpu_count() or 1))


def job_count(option):
    """`-j`, else `MOAI_UNITTEST_JOBS`, else the smaller of four and the processor count."""
    if option is not None:
        value, source = option, "-j"
    elif os.environ.get("MOAI_UNITTEST_JOBS"):
        value, source = os.environ["MOAI_UNITTEST_JOBS"], "MOAI_UNITTEST_JOBS"
    else:
        return default_jobs()
    try:
        jobs = int(value)
    except ValueError:
        jobs = 0
    if jobs < 1:
        say(f"unittest_parallel: {source} wants a whole number of at least 1, got {value!r}\n")
        sys.exit(2)
    return jobs


def modules(start, pattern):
    """File names in `start` that match `pattern`, the biggest file first, then by name."""
    try:
        names = os.listdir(start)
    except OSError as err:
        say(f"unittest_parallel: cannot read {start}: {err}\n")
        sys.exit(2)
    found = sorted(
        n for n in names if fnmatch.fnmatch(n, pattern) and os.path.isfile(os.path.join(start, n))
    )
    # The biggest files start first: a long module started last would be the whole run's tail.
    return sorted(found, key=lambda n: -os.path.getsize(os.path.join(start, n)))


def process_env(root):
    """The environment of one process: its own HOME and TMPDIR, the caller's toolchain homes."""
    env = dict(os.environ)
    caller_home = env.get("HOME", "")
    if caller_home:
        env.setdefault("CARGO_HOME", os.path.join(caller_home, ".cargo"))
        env.setdefault("RUSTUP_HOME", os.path.join(caller_home, ".rustup"))
    home = os.path.join(root, "home")
    tmp = os.path.join(root, "tmp")
    os.makedirs(home)
    os.makedirs(tmp)
    env["HOME"] = home
    env["TMPDIR"] = tmp
    return env


def classes_of(start, name):
    """The test classes of a big module, or `None` when it is run whole."""
    if os.path.getsize(os.path.join(start, name)) <= SPLIT_BYTES:
        return None
    root = tempfile.mkdtemp(prefix="moai-unittest-")
    try:
        proc = subprocess.run(
            [sys.executable, "-c", LIST_CLASSES, start, os.path.splitext(name)[0]],
            env=process_env(root),
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
        )
    finally:
        shutil.rmtree(root, ignore_errors=True)
    found = proc.stdout.split() if proc.returncode == 0 else []
    return found if len(found) >= 2 else None


def run_unit(start, label, argv, verbose):
    """Runs one process; returns (label, exit code, output, seconds)."""
    root = tempfile.mkdtemp(prefix="moai-unittest-")
    began = time.monotonic()
    try:
        env = process_env(root)
        cmd = [sys.executable, "-m", "unittest"]
        if argv[0] == "discover":
            cmd.append("discover")
            if verbose:
                cmd.append("-v")
            cmd += ["-s", start, "-p", argv[1]]
        else:
            # `<module>.<Class>`: the start folder reaches the import path like discover's does.
            if verbose:
                cmd.append("-v")
            cmd.append(argv[1])
            env["PYTHONPATH"] = os.pathsep.join(
                [start] + ([env["PYTHONPATH"]] if env.get("PYTHONPATH") else [])
            )
        proc = subprocess.run(
            cmd,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            errors="replace",
        )
        return label, proc.returncode, proc.stdout, time.monotonic() - began
    finally:
        shutil.rmtree(root, ignore_errors=True)


def split_report(output):
    """(body, ran, counts): the output before `unittest`'s closing report, the number of tests
    it ran and the numbers in its parenthesis, or `None` for a run that printed no report."""
    rule = "\n" + "-" * 70 + "\n"
    at = output.rfind(rule + "Ran ")
    if at < 0:
        return output, None, {}
    tail = output[at + len(rule):]
    ran = RAN.search(tail)
    result = RESULT.search(tail)
    counts = {}
    if result and result.group(2):
        for part in result.group(2).split(", "):
            key, _, number = part.partition("=")
            if number.isdigit():
                counts[key] = int(number)
    return output[:at].rstrip("\n"), int(ran.group(1)) if ran else None, counts


def plan(start, names, jobs):
    """The processes to run: one per module, one per class for a big module."""
    with concurrent.futures.ThreadPoolExecutor(max_workers=jobs) as pool:
        listed = list(pool.map(lambda n: classes_of(start, n), names))
    units = []
    for name, classes in zip(names, listed):
        if classes is None:
            units.append((name, ["discover", name]))
        else:
            module = os.path.splitext(name)[0]
            units.extend((f"{name}::{c}", ["name", f"{module}.{c}"]) for c in classes)
    return units


def main():
    parser = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    parser.add_argument("-v", "--verbose", action="store_true")
    parser.add_argument("-s", "--start-directory", default=".")
    parser.add_argument("-p", "--pattern", default="test*.py")
    parser.add_argument("-j", "--jobs", default=None)
    args = parser.parse_args()
    jobs = job_count(args.jobs)
    names = modules(args.start_directory, args.pattern)
    if not names:
        say(f"unittest_parallel: no test module matched {args.pattern!r} in {args.start_directory}\n")
        return 1

    began = time.monotonic()
    units = plan(args.start_directory, names, jobs)
    total = 0
    sums = {key: 0 for key in KEYS}
    failed = []
    timings = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=jobs) as pool:
        futures = [
            pool.submit(run_unit, args.start_directory, label, argv, args.verbose)
            for label, argv in units
        ]
        for future in concurrent.futures.as_completed(futures):
            label, code, output, seconds = future.result()
            body, ran, counts = split_report(output)
            timings.append((seconds, label))
            # `NO TESTS RAN` (exit 5 on new Pythons) is a module with no cases, not a failure.
            empty = code == 5 and "NO TESTS RAN" in output
            ok = code == 0 or empty
            if ran is None and not empty:
                # No report at all: the process died or printed nothing a run prints.
                counts = {"errors": 1}
                body = (
                    f"{output.rstrip()}\n" if output.strip() else ""
                ) + f"unittest_parallel: {label} exited {code} without a unittest report"
                ok = False
            total += ran or 0
            for key in KEYS:
                sums[key] += counts.get(key, 0)
            if not ok:
                failed.append(label)
            if body.strip() and (args.verbose or not ok):
                say(body.rstrip("\n") + "\n")
            if not ok and not body.strip():
                say(f"unittest_parallel: {label} failed with exit {code}\n")

    wall = time.monotonic() - began
    slowest = sorted(timings, reverse=True)[:5]
    say("\nmodules: " + str(len(names)) + ", processes: " + str(len(units)) + ", jobs: "
        + str(jobs) + ", slowest: " + ", ".join(f"{n} {s:.1f}s" for s, n in slowest) + "\n")
    if failed:
        say("failed: " + ", ".join(sorted(failed)) + "\n")
    infos = [f"{key}={sums[key]}" for key in KEYS if sums[key]]
    if failed and not (sums["failures"] or sums["errors"]):
        infos.insert(0, f"errors={len(failed)}")
    verdict = "FAILED" if failed else "OK"
    tail = f" ({', '.join(infos)})" if infos else ""
    say("-" * 70 + "\n" + f"Ran {total} tests in {wall:.3f}s\n\n{verdict}{tail}\n")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
