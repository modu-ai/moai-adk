#!/usr/bin/env python3
"""Per-call latency of `moai-cli hook-dispatch` (Issue #2391, cost: P50/P95 per call).

A measuring tool, not a test: it runs each hook duty SAMPLES times against a throwaway git
repository with its own HOME, and prints one JSON object with P50/P95/max per host and case,
labelled with the binary, its build profile (guessed from the path), the sample count and the
load average at start and end. No target is decided yet (HOLD), so nothing here passes or
fails on a number; wall-clock values depend on machine load and must be read with the label.

Usage:
  python3 -I tools/hook_latency.py --bin PATH [--samples N]

Exit codes: 0 = measured, 1 = a run could not start, 2 = usage.
"""

import json
import math
import os
import subprocess
import sys
import tempfile
import time
from pathlib import Path

CASES = {
    "session-check": ("session-check", {"hook_event_name": "SessionStart", "source": "startup"}),
    "tool-preflight-allow": ("tool-preflight", {"hook_event_name": "PreToolUse", "tool_name": "Bash",
                                                "tool_input": {"command": "ls"}}),
    "tool-preflight-deny": ("tool-preflight", {"hook_event_name": "PreToolUse", "tool_name": "Bash",
                                               "tool_input": {"command": "git commit -m x"}}),
    "tool-preflight-commit-token": ("tool-preflight", {"hook_event_name": "PreToolUse", "tool_name": "Bash",
                                                       "tool_input": {"command": "git commit -m 'x Moai-Claim: 1#5'"}}),
    "tool-observe": ("tool-observe", {"hook_event_name": "PostToolUse", "tool_name": "Read",
                                      "tool_input": {"file_path": "a.txt"}, "tool_response": "x"}),
    "turn-validation": ("turn-validation", {"hook_event_name": "Stop", "stop_hook_active": False}),
}
HOSTS = ("claude", "codex")


def percentile(values, p):
    """Nearest-rank percentile: the smallest value with at least p% of the values at or below
    it. `p` is in (0, 100]; an empty list or an out-of-range p is an error."""
    if not values:
        raise ValueError("no values")
    if not 0 < p <= 100:
        raise ValueError("p must be in (0, 100]")
    ordered = sorted(values)
    rank = math.ceil(p / 100 * len(ordered))
    return ordered[rank - 1]


def summarize(durations_ms):
    return {
        "p50_ms": round(percentile(durations_ms, 50), 3),
        "p95_ms": round(percentile(durations_ms, 95), 3),
        "max_ms": round(max(durations_ms), 3),
    }


def profile_of(binary):
    parts = Path(binary).parts
    for name in ("release", "debug"):
        if name in parts:
            return name
    return "unknown"


def git(root, *args):
    env = dict(os.environ, GIT_CONFIG_GLOBAL="/dev/null", GIT_CONFIG_SYSTEM="/dev/null",
               HOME=str(root / "home"))
    subprocess.run(["git", *args], cwd=root, env=env, check=True, capture_output=True)


def fixture(base):
    root = Path(base).resolve()
    (root / "home").mkdir()
    git(root, "init", "-q", "-b", "issue/1-bench")
    git(root, "config", "user.name", "bench")
    git(root, "config", "user.email", "bench@example.invalid")
    (root / "a.txt").write_text("a\n")
    git(root, "add", "a.txt")
    git(root, "commit", "-q", "-m", "init")
    (root / ".moai-cli").mkdir()
    return root


def measure(binary, root, host, event, payload, samples):
    env = dict(os.environ, HOME=str(root / "home"), GIT_CONFIG_GLOBAL="/dev/null",
               GIT_CONFIG_SYSTEM="/dev/null")
    env.pop("MOAI_FORGE_GH_BIN", None)
    data = json.dumps(dict(payload, cwd=str(root), session_id="bench")).encode()
    durations, codes = [], set()
    for _ in range(samples):
        start = time.perf_counter_ns()
        out = subprocess.run([binary, "hook-dispatch", "--host", host, "--event", event],
                             input=data, cwd=root, env=env, capture_output=True)
        durations.append((time.perf_counter_ns() - start) / 1e6)
        codes.add(out.returncode)
    return dict(summarize(durations), exit_codes=sorted(codes))


def usage():
    print(__doc__)
    raise SystemExit(2)


def main(argv):
    args = list(argv)
    if "--bin" not in args or args.index("--bin") + 1 >= len(args):
        usage()
    # Each run's working directory is the fixture repository, so a relative path would no
    # longer resolve there: fix it to an absolute path first.
    binary = str(Path(args[args.index("--bin") + 1]).resolve())
    samples = 50
    if "--samples" in args:
        try:
            samples = int(args[args.index("--samples") + 1])
        except (IndexError, ValueError):
            usage()
        if samples < 1:
            usage()
    load_start = os.getloadavg()
    results = {}
    with tempfile.TemporaryDirectory() as base:
        root = fixture(base)
        try:
            for host in HOSTS:
                results[host] = {name: measure(binary, root, host, event, payload, samples)
                                 for name, (event, payload) in CASES.items()}
        except OSError as e:
            print(json.dumps({"error": f"run failed: {e.strerror}"}))
            raise SystemExit(1)
    print(json.dumps({
        "binary": binary,
        "build_profile": profile_of(binary),
        "samples": samples,
        "load_avg_start": [round(x, 2) for x in load_start],
        "load_avg_end": [round(x, 2) for x in os.getloadavg()],
        "target": "HOLD: no P50/P95 target is decided",
        "results": results,
    }, indent=2))


if __name__ == "__main__":
    main(sys.argv[1:])
