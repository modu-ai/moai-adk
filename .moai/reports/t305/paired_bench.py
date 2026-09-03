"""Paired, interleaved wall-clock comparison of the baseline and changed
`moai statusline` binaries.

The bench-before / bench-after series were taken minutes apart on a loaded
machine (load 9.49 vs 5.66), so their absolute millisecond figures are not
comparable. This alternates the two binaries within one run, so whatever load
the machine is under lands on both arms and largely cancels — the same
interleaving discipline the repo's own latency-budget test uses.
"""

import statistics
import subprocess
import sys
import time

PAYLOAD = open(".moai/reports/t305/payload.json", "rb").read()
BINARIES = {"base": "/tmp/t305_moai_base", "after": "/tmp/t305_moai_after"}
N = 25


def one(path: str) -> float:
    start = time.perf_counter()
    subprocess.run([path, "statusline"], input=PAYLOAD,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)
    return (time.perf_counter() - start) * 1000.0


def main() -> int:
    # One warm-up per arm so page-cache effects do not land on the first sample.
    for path in BINARIES.values():
        one(path)

    samples = {name: [] for name in BINARIES}
    for _ in range(N):
        for name, path in BINARIES.items():
            samples[name].append(one(path))

    for name in BINARIES:
        s = sorted(samples[name])
        print(f"{name:6s} n={len(s)} median={statistics.median(s):8.2f}ms "
              f"p95={s[int(len(s) * 0.95) - 1]:8.2f}ms min={s[0]:8.2f}ms max={s[-1]:8.2f}ms")

    base_med = statistics.median(samples["base"])
    after_med = statistics.median(samples["after"])
    print(f"delta  median {base_med - after_med:+.2f}ms "
          f"({(base_med - after_med) / base_med * 100:.1f}% of baseline)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
