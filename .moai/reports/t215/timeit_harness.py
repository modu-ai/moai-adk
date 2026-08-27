#!/usr/bin/env python3
"""External wall-clock timing harness for the moai statusline chain.

Runs each candidate command N times sequentially with the fixture JSON on
stdin (file redirect), records per-run wall time, and prints median / p95 /
min / max in milliseconds plus a trimmed summary. One python process drives
all runs so harness startup never contaminates individual samples.
"""
import json
import statistics
import subprocess
import sys
import time

FIXTURE = "bin/fixture.json"


def run_series(cmd, n):
    times = []
    with open(FIXTURE, "rb") as f:
        payload = f.read()
    for _ in range(n):
        t0 = time.perf_counter()
        subprocess.run(
            cmd,
            input=payload,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000.0)
    return times


def summarize(name, times):
    srt = sorted(times)
    med = statistics.median(srt)
    p95 = srt[int(len(srt) * 0.95) - 1] if len(srt) >= 20 else srt[-1]
    print(
        f"{name}: N={len(srt)} median={med:.1f}ms p95={p95:.1f}ms "
        f"min={srt[0]:.1f}ms max={srt[-1]:.1f}ms mean={statistics.mean(srt):.1f}ms"
    )
    return {"name": name, "median": round(med, 1), "p95": round(p95, 1), "all": [round(t, 1) for t in srt]}


def main():
    with open(sys.argv[1]) as f:
        spec = json.load(f)
    n = int(sys.argv[2]) if len(sys.argv) > 2 else 40
    results = []
    # Warm-up: OS page cache for binary + dyld + file cache (2 unrecorded runs)
    for entry in spec:
        run_series(entry["cmd"], 2)
    for entry in spec:
        times = run_series(entry["cmd"], n)
        results.append(summarize(entry["name"], times))
    with open("bin/timing_results.json", "w") as f:
        json.dump(results, f, indent=2)


if __name__ == "__main__":
    main()
