#!/usr/bin/env bash
# SPEC-V3R3-UPDATE-CLEANUP-001 — NFR-UPC-P1 benchstat overhead gate.
# Usage: ./scripts/benchmark-overhead-check.sh <bench-report.txt>
#
# Reads benchstat output and fails (exit 1) if:
#   delta_pct > 5% AND p_value < 0.05
#
# If delta_pct ≤ 5%, the test passes regardless of p-value.
# If p_value ≥ 0.05 (not statistically significant), the test passes.
#
# Official gate OS: ubuntu-latest (CI). macOS/Windows results are informational.

set -euo pipefail

REPORT="${1:-bench-report.txt}"

if [[ ! -f "$REPORT" ]]; then
  echo "[ERROR] benchmark report not found: $REPORT" >&2
  exit 1
fi

echo "=== NFR-UPC-P1 Benchmark Overhead Gate ==="
echo "Report: $REPORT"
echo ""
cat "$REPORT"
echo ""

# Extract delta_pct and p-value from benchstat output.
# benchstat output example:
#   name                         old ns/op   new ns/op   delta
#   UpdateCleanup_Baseline       8.2ms ± 2%  8.5ms ± 1%  +3.66%  (p=0.008 n=10+10)
#
# We parse the last numeric column for delta% and the p-value.

FAIL=0

while IFS= read -r line; do
  # Skip header and empty lines
  [[ -z "$line" ]] && continue
  [[ "$line" == name* ]] && continue

  # Look for lines containing delta % and p-value
  if [[ "$line" =~ ([+-][0-9]+(\.[0-9]+)?)\%[[:space:]]+\(p=([0-9]+\.[0-9]+) ]]; then
    delta_pct="${BASH_REMATCH[1]}"
    p_value="${BASH_REMATCH[3]}"

    echo "Detected: delta_pct=${delta_pct}% p_value=${p_value}"

    # Remove leading + sign for comparison
    abs_delta="${delta_pct#+}"
    # abs_delta may start with - for improvement; use absolute value
    abs_delta="${abs_delta#-}"

    # Compare: delta_pct > 5 AND p_value < 0.05 → FAIL
    gate_fail=$(awk -v d="$abs_delta" -v p="$p_value" 'BEGIN {
      if (d > 5 && p < 0.05) { print "1" } else { print "0" }
    }')

    if [[ "$gate_fail" == "1" ]]; then
      echo "[FAIL] Atomic write overhead exceeds 5% (delta=${delta_pct}%, p=${p_value})" >&2
      FAIL=1
    else
      echo "[PASS] Overhead within acceptable range (delta=${delta_pct}%, p=${p_value})"
    fi
  fi
done < "$REPORT"

if [[ $FAIL -eq 1 ]]; then
  echo ""
  echo "[ERROR] NFR-UPC-P1 gate FAILED: atomic write overhead > 5% (p < 0.05)" >&2
  exit 1
fi

echo ""
echo "[OK] NFR-UPC-P1 gate PASSED"
exit 0
