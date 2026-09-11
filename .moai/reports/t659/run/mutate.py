#!/usr/bin/env python3
"""Mutant runner for card t659 (SPEC-CON-AMEND-APPLY-001).

Usage: mutate.py <specs.json> <mutant-id> [<mutant-id> ...]

Each spec entry: {"id": ..., "edits": [[file, old, new], ...],
"run": [go test argv after "go"], "expect_runs": N, "out": path}.

For one mutant: apply every edit (each `old` must occur exactly once), run the
killing selector, write the verbatim output to `out`, restore every edited file
byte for byte, and print one summary line. The mutant is KILLED only when the
selector reached its tests (top-level `=== RUN` count >= expect_runs) and the
run failed. Restoration is checked against the pre-mutation bytes.
"""
import json
import re
import subprocess
import sys


def main() -> int:
    specs = {s["id"]: s for s in json.load(open(sys.argv[1]))}
    status = 0
    for mid in sys.argv[2:]:
        spec = specs[mid]
        originals = {}
        try:
            for path, old, new in spec["edits"]:
                if path not in originals:
                    originals[path] = open(path, "rb").read()
                src = open(path).read()
                if src.count(old) != 1:
                    raise SystemExit(f"{mid}: edit anchor occurs {src.count(old)} times in {path}")
                open(path, "w").write(src.replace(old, new))
            proc = subprocess.run(["go"] + spec["run"], capture_output=True, text=True)
            out = proc.stdout + proc.stderr
        finally:
            for path, data in originals.items():
                open(path, "wb").write(data)
        for path, data in originals.items():
            if open(path, "rb").read() != data:
                raise SystemExit(f"{mid}: restore of {path} failed")
        with open(spec["out"], "w") as fh:
            fh.write(f"# mutant {mid}: {spec.get('desc', '')}\n")
            fh.write(f"# edits: {json.dumps([[p, o, n] for p, o, n in spec['edits']])}\n")
            fh.write(f"# $ go {' '.join(spec['run'])}\n# exit={proc.returncode}\n")
            fh.write(out)
        runs = len(re.findall(r"(?m)^=== RUN   [A-Za-z0-9_]+$", out))
        build_failed = "[build failed]" in out or "[setup failed]" in out
        killed = proc.returncode != 0 and runs >= spec["expect_runs"] and not build_failed
        verdict = "KILLED" if killed else ("BUILD-FAIL" if build_failed else "SURVIVED-OR-UNREACHED")
        fails = re.findall(r"(?m)^\s*--- FAIL: (\S+)", out)
        print(f"{mid}: {verdict} exit={proc.returncode} top_level_runs={runs} "
              f"(expect>={spec['expect_runs']}) fails={fails}")
        if not killed:
            status = 1
    return status


if __name__ == "__main__":
    sys.exit(main())
