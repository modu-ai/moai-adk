#!/usr/bin/env python3
"""The load probe of the M-Flow before/after measurement (SPEC-20261120-mflow-measure).

Run one target command, sample the machine while it runs, and answer with the numbers
the instruction names — the 1-minute load median and max, the CPU user/system
percentages with the utilisation, the memory-pressure level and the swap used, the peak
count of concurrently running compile and test processes, the elapsed wall time, and how
the command ended (its exit code, or the signal that killed it — never folded together).
The answer is one JSON line on stdout; `--out` writes the full document with the sample
series, so the reports table is built from observed numbers only.

    python3 -I tools/mflow_probe.py --label before-1 --base <40hex> --head <40hex> \\
        [--interval 1.0] [--max-load N] [--out run.json] -- <target command...>

The macOS sources are `sysctl -n vm.loadavg`, `sysctl vm.swapusage`,
`sysctl kern.memorystatus_vm_pressure_level`, `vm_stat` and `ps aux` — standard library
only. On another OS the command still runs and its ending is still reported; the
platform-specific fields are null with a note. The probe observes: it never kills,
pauses or reprioritises anything, and it writes nothing under the real HOME except the
caller-named `--out` file.
"""
import argparse
import datetime
import json
import os
import signal
import subprocess
import sys
import time

COMPILE_TEST_WORDS = (
    "rustc", "cargo", "cargo-nextest", "nextest", "unittest", "python3", "moai-cli",
)

# The KST offset of the instruction's timestamps (Korea does not observe DST).
KST = datetime.timezone(datetime.timedelta(hours=9))


def run_text(command):
    """First line of a subprocess's stdout, or None when the source is unreadable."""
    try:
        out = subprocess.run(command, capture_output=True, text=True, timeout=10)
    except (OSError, subprocess.TimeoutExpired):
        return None
    if out.returncode != 0:
        return None
    return out.stdout.strip().splitlines()[0] if out.stdout.strip() else None


def load_one():
    """The 1-minute load from `sysctl -n vm.loadavg` (`{ a, b, c }` on macOS)."""
    line = run_text(["sysctl", "-n", "vm.loadavg"])
    if not line:
        return None
    parts = line.strip("{}").split()
    try:
        return float(parts[0])
    except (IndexError, ValueError):
        return None


def swap_used_bytes():
    """Used swap from `sysctl vm.swapusage` (`total = A used = B free = C`) — the third
    `=` field carries the used figure (`1024.00M`)."""
    line = run_text(["sysctl", "vm.swapusage"])
    if not line or "used" not in line:
        return None
    fields = line.split("=")
    if len(fields) < 3:
        return None
    used = fields[2].strip().split()[0]
    return size_to_bytes(used)


def size_to_bytes(text):
    units = {"K": 2**10, "M": 2**20, "G": 2**30, "T": 2**40}
    number, unit = text[:-1], text[-1].upper()
    try:
        value = float(number)
    except ValueError:
        return None
    return int(value * units.get(unit, 1))


def pressure_level():
    """`sysctl kern.memorystatus_vm_pressure_level`: 1 normal … 5 critical, else None."""
    line = run_text(["sysctl", "-n", "kern.memorystatus_vm_pressure_level"])
    try:
        return int(line)
    except (TypeError, ValueError):
        return None


def top_cpu_split():
    """`top -l 1 -n 0`'s `CPU usage: X% user, Y% sys, Z% idle` line — the user/system
    split the other sources do not give. One call, used sparingly (start and end)."""
    try:
        out = subprocess.run(["top", "-l", "1", "-n", "0"], capture_output=True, text=True,
                             timeout=10)
    except (OSError, subprocess.TimeoutExpired):
        return None
    if out.returncode != 0:
        return None
    line = next((l for l in out.stdout.splitlines() if "CPU usage" in l), None)
    if not line:
        return None
    values = {}
    for chunk in line.split(","):
        chunk = chunk.strip()
        for key in ("user", "sys", "idle"):
            if key in chunk:
                for token in chunk.split():
                    if token.endswith("%"):
                        try:
                            values[key] = float(token[:-1])
                        except ValueError:
                            return None
                        break
    if not {"user", "sys", "idle"} <= set(values):
        return None
    return {"user": values["user"], "system": values["sys"],
            "utilisation": max(0.0, 100.0 - values["idle"])}


def compile_test_process_count():
    """Peak-friendly sample: how many compile/test processes are running right now —
    by command word (rustc, cargo, nextest, unittest, the CLI) from `ps aux`."""
    try:
        out = subprocess.run(["ps", "aux"], capture_output=True, text=True, timeout=10)
    except (OSError, subprocess.TimeoutExpired):
        return None
    if out.returncode != 0:
        return None
    count = 0
    for line in out.stdout.splitlines():
        parts = line.split(None, 10)
        if len(parts) < 11:
            continue
        command = parts[10]
        if any(word in command for word in COMPILE_TEST_WORDS):
            count += 1
    return count


def machine_facts():
    ncpu = run_text(["sysctl", "-n", "hw.ncpu"])
    model = run_text(["sysctl", "-n", "hw.model"])
    facts = {"ncpu": None, "model": None, "os": sys.platform}
    try:
        facts["ncpu"] = int(ncpu)
    except (TypeError, ValueError):
        pass
    facts["model"] = model
    return facts


def version_facts():
    cli = None
    for name in ("moai-cli", "moai"):
        found = None
        for base in os.environ.get("PATH", "").split(os.pathsep):
            candidate = os.path.join(base, name)
            if os.path.isfile(candidate) and os.access(candidate, os.X_OK):
                found = candidate
                break
        if found is not None:
            line = run_text([found, "--version"])
            if line:
                cli = line
                break
    return {
        "git": run_text(["git", "--version"]),
        "rustc": run_text(["rustc", "--version"]),
        "cli": cli,
    }


def now_stamps():
    utc = datetime.datetime.now(datetime.timezone.utc)
    return utc.strftime("%Y-%m-%dT%H:%M:%SZ"), utc.astimezone(KST).strftime(
        "%Y-%m-%dT%H:%M:%S+09:00"
    )


def median(values):
    if not values:
        return None
    ordered = sorted(values)
    middle = len(ordered) // 2
    if len(ordered) % 2:
        return float(ordered[middle])
    return (ordered[middle - 1] + ordered[middle]) / 2


def main(argv):
    parser = argparse.ArgumentParser(add_help=True, description=__doc__.splitlines()[0])
    parser.add_argument("--label", required=True, help="the run's label in the reports table")
    parser.add_argument("--base", required=True, help="the measured base commit (40 hex)")
    parser.add_argument("--head", required=True, help="the measured head commit (40 hex)")
    parser.add_argument("--interval", type=float, default=1.0, help="sample seconds (default 1)")
    parser.add_argument("--max-load", type=float, default=None,
                        help="the low-load threshold; the answer marks a run above it")
    parser.add_argument("--out", default=None, help="write the full document (with the sample series) here")
    parser.add_argument("command", nargs=argparse.REMAINDER,
                        help="the target command after `--`")
    args = parser.parse_args(argv)
    if not args.command:
        print("usage: mflow_probe.py [--label L --base H --head H ...] -- <command...>", file=sys.stderr)
        return 2
    if args.command and args.command[0] == "--":
        args.command = args.command[1:]
    if not args.command:
        print("usage: mflow_probe.py [--label L --base H --head H ...] -- <command...>", file=sys.stderr)
        return 2

    utc, kst = now_stamps()
    start_cpu = top_cpu_split()
    start_load = load_one()
    started = time.monotonic()
    try:
        proc = subprocess.Popen(args.command, cwd=os.getcwd())
    except OSError as exc:
        print(json.dumps({"label": args.label, "error": f"target did not start: {exc}"}))
        return 2
    series = {"load": [], "swap_used_bytes": [], "pressure_level": [],
              "cpu_pcpu": [], "compile_test_processes": []}
    peak = 0
    while True:
        sample_load = load_one()
        sample_swap = swap_used_bytes()
        sample_pressure = pressure_level()
        sample_procs = compile_test_process_count()
        try:
            lines = subprocess.run(["ps", "-A", "-o", "pcpu="], capture_output=True,
                                   text=True, timeout=10).stdout.split()
            sample_pcpu = sum(float(x) for x in lines)
        except (OSError, subprocess.TimeoutExpired, ValueError):
            sample_pcpu = None
        for key, value in (("load", sample_load), ("swap_used_bytes", sample_swap),
                           ("pressure_level", sample_pressure),
                           ("cpu_pcpu", sample_pcpu),
                           ("compile_test_processes", sample_procs)):
            if value is not None:
                series[key].append(value)
        if sample_procs is not None:
            peak = max(peak, sample_procs)
        try:
            code = proc.wait(timeout=args.interval)
            break
        except subprocess.TimeoutExpired:
            continue
    elapsed = time.monotonic() - started
    end_cpu = top_cpu_split()

    ended = {"kind": "exit", "code": code}
    if code < 0:
        ended = {"kind": "signal", "signal": signal.Signals(-code).name}
    load_values = series["load"]
    answer = {
        "label": args.label,
        "utc": utc,
        "kst": kst,
        "machine": machine_facts(),
        "versions": version_facts(),
        "base": args.base,
        "head": args.head,
        "command": args.command,
        "interval_seconds": args.interval,
        "window": "ok" if start_load is not None and args.max_load is not None
                  and start_load <= args.max_load else "above",
        "window_observed_load": start_load,
        "load": {"median": median(load_values), "max": max(load_values, default=None),
                 "samples": len(load_values)},
        "cpu": {
            "user_percent": None if not (start_cpu and end_cpu)
                            else round((start_cpu["user"] + end_cpu["user"]) / 2, 1),
            "system_percent": None if not (start_cpu and end_cpu)
                              else round((start_cpu["system"] + end_cpu["system"]) / 2, 1),
            "utilisation_percent": None if not (start_cpu and end_cpu)
                                   else round((start_cpu["utilisation"] + end_cpu["utilisation"]) / 2, 1),
        },
        "memory": {"pressure_level": median([v for v in series["pressure_level"] if v is not None]),
                   "swap_used_bytes": max(series["swap_used_bytes"], default=None)},
        "cpu_pcpu_max": max(series["cpu_pcpu"], default=None),
        "peak_processes": peak,
        "elapsed_seconds": round(elapsed, 3),
        "ended": ended,
    }
    if args.out:
        with open(args.out, "w", encoding="utf-8") as fh:
            json.dump({"answer": answer, "series": series}, fh, indent=1)
    print(json.dumps(answer))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
