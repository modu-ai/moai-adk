#!/bin/bash
# RC1 benchmark harness (#2190, REQ-D4-01).
#
# Measures moai-cli internal operations on a fixed fixture with fixed conditions:
#   - the binary from $CARGO_TARGET_DIR (built by the caller; debug/release is
#     recorded by the caller in the report)
#   - WARMUP discarded runs, then SAMPLES timed runs per operation
#   - per run: wall time via time.perf_counter_ns around the child process; max
#     RSS via /usr/bin/time -l (macOS: bytes) in a separate sampling pass
#   - p50/p95 over the SAMPLES (nearest-rank on the sorted list)
#
# Emits one markdown table row per operation on stdout. The caller records the
# environment, commands and iteration counts in the report (REQ-D4-01).
# Comparisons are allowed only against this baseline (no reduction-rate claims).
#
# Usage: benchmark.sh [SAMPLES] [WARMUP]   (defaults: 12 2)
# Requires: the binary built at ${CARGO_TARGET_DIR:-target}/debug/moai-cli.

set -euo pipefail

SAMPLES="${1:-12}"
WARMUP="${2:-2}"
REPO="$(git rev-parse --show-toplevel)"
BIN="${CARGO_TARGET_DIR:-target}/debug/moai-cli"
FIXTURE="$(mktemp -d /tmp/moai-bench-fixture.XXXXXX)"
export FIXTURE

mkdir -p "$FIXTURE/skills/moai/workflows" "$FIXTURE/.moai-cli"
printf '%s\n' '{"schema_version":2,"project_id":"bench","purpose":"Benchmark fixture","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"bench-1"}' \
    > "$FIXTURE/.moai-cli/project.yaml"
printf '%s\n' '{"protocol_version":1,"language":"en","workflow_reference":"skills/moai/workflows","local_file":"none","workflow_id":"bench-flow","native_roles":[],"skill_projections":[]}' \
    > "$FIXTURE/generation-request.json"
printf '%s\n' '{"schema_version":2,"project_id":"bench","purpose":"Benchmark fixture","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"]}' \
    > "$FIXTURE/project-request.json"
trap 'rm -rf "$FIXTURE"' EXIT

# Timing loop for one operation (wall time only): label via $BENCH_LABEL, the
# operation's argv after the binary as positional arguments.
bench_op_time() {
    local stdin_file="$1" workdir="$2"
    shift 2
    BENCH_LABEL="$label" python3 - "$BIN" "$stdin_file" "$workdir" "$SAMPLES" "$WARMUP" "$@" <<'PYEOF'
import math, os, subprocess, sys, time
bin = sys.argv[1]
stdin_file, workdir = sys.argv[2], sys.argv[3]
samples, warmup = int(sys.argv[4]), int(sys.argv[5])
label = os.environ["BENCH_LABEL"]
op_args = sys.argv[6:]
stdin_bytes = open(stdin_file, "rb").read() if stdin_file != "-" else b""
def run_once():
    t0 = time.perf_counter_ns()
    subprocess.run([bin] + op_args, input=stdin_bytes, cwd=workdir,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)
    return time.perf_counter_ns() - t0
for _ in range(warmup):
    run_once()
rows = sorted(run_once() for _ in range(samples))
def pct(p):
    idx = max(0, math.ceil(p * len(rows)) - 1)
    return rows[idx] / 1e6  # ms
print(f"| {label} | {pct(0.5):.1f} | {pct(0.95):.1f} | {rows[0]/1e6:.1f} | {rows[-1]/1e6:.1f} |")
PYEOF
}

# RSS sampling loop for one operation: max RSS per run via /usr/bin/time -l.
bench_op_rss() {
    local stdin_file="$1" workdir="$2"
    shift 2
    BENCH_LABEL="$label" python3 - "$BIN" "$stdin_file" "$workdir" "$SAMPLES" "$WARMUP" "$@" <<'PYEOF'
import math, os, re, subprocess, sys
bin = sys.argv[1]
stdin_file, workdir = sys.argv[2], sys.argv[3]
samples, warmup = int(sys.argv[4]), int(sys.argv[5])
label = os.environ["BENCH_LABEL"]
op_args = sys.argv[6:]
stdin_bytes = open(stdin_file, "rb").read() if stdin_file != "-" else b""
def run_once():
    out = subprocess.run(["/usr/bin/time", "-l"] + [bin] + op_args, input=stdin_bytes,
                         cwd=workdir, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
    m = re.search(rb"(\d+)\s+maximum resident set size", out.stderr)
    return int(m.group(1)) if m else 0
for _ in range(warmup):
    run_once()
rows = sorted(run_once() for _ in range(samples))
def pct(p):
    idx = max(0, math.ceil(p * len(rows)) - 1)
    return rows[idx] / 1024 / 1024  # MiB
print(f"| {label} | {pct(0.5):.1f} | {pct(0.95):.1f} |")
PYEOF
}

run_time_ops() {
    label="project-plan (fixture)"
    echo "$(bench_op_time "$FIXTURE/project-request.json" "$FIXTURE" project-plan "$FIXTURE")"
    label="generation-plan (fixture)"
    echo "$(bench_op_time "$FIXTURE/generation-request.json" "$FIXTURE" generation-plan "$FIXTURE")"
    label="package-inspect (skills/moai)"
    echo "$(bench_op_time "-" "$REPO" package-inspect "$REPO/skills/moai")"
}

run_rss_ops() {
    label="project-plan (fixture)"
    echo "$(bench_op_rss "$FIXTURE/project-request.json" "$FIXTURE" project-plan "$FIXTURE")"
    label="generation-plan (fixture)"
    echo "$(bench_op_rss "$FIXTURE/generation-request.json" "$FIXTURE" generation-plan "$FIXTURE")"
    label="package-inspect (skills/moai)"
    echo "$(bench_op_rss "-" "$REPO" package-inspect "$REPO/skills/moai")"
}

echo "| 연산 | p50 (ms) | p95 (ms) | 최소 (ms) | 최대 (ms) |"
echo "| --- | --- | --- | --- | --- |"
run_time_ops
echo
echo "| 연산 | RSS p50 (MiB) | RSS p95 (MiB) |"
echo "| --- | --- | --- |"
run_rss_ops
