"""Paired, interleaved wall-clock comparison of the pre- and post-t456
`moai statusline` render path.

Adapted from .moai/reports/t305/paired_bench.py — same discipline, same reason.
Absolute millisecond figures taken minutes apart are not comparable on a loaded
machine (this one sat at load 16.5 while t305 measured at 6.68), so the two
binaries alternate within a single run and whatever load is present lands on
both arms.

Both arms use a NON-self-invocable binary name, so neither spawns a refresh
child: the claim under test is that the render path did not get slower, and a
fire-and-forget child the render never waits on would only add noise. The cold
-> warm child spawn is measured separately, by subprocess census rather than by
clock.
"""

import statistics
import subprocess
import sys
import time

PAYLOAD = open(".moai/reports/t456/payload-wt.json", "rb").read()
BINARIES = {"base": "/tmp/t456_moai_base", "after": "./bin/moai-t456"}
N = 25


def one(path: str) -> float:
    start = time.perf_counter()
    subprocess.run([path, "statusline"], input=PAYLOAD,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)
    return (time.perf_counter() - start) * 1000.0


def main() -> int:
    samples = {k: [] for k in BINARIES}
    for _ in range(N):
        for name, path in BINARIES.items():
            samples[name].append(one(path))
    for name in BINARIES:
        s = sorted(samples[name])
        print(f"{name:6s} n={len(s)} median={statistics.median(s):8.2f}ms "
              f"p95={s[int(0.95 * len(s)) - 1]:8.2f}ms "
              f"min={s[0]:8.2f}ms max={s[-1]:8.2f}ms")
    b = statistics.median(samples["base"])
    a = statistics.median(samples["after"])
    print(f"delta  median {a - b:+.2f}ms ({(a - b) / b * 100:+.1f}% of baseline)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
