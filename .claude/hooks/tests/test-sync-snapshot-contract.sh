#!/bin/bash
set -euo pipefail

doc="$(cd "$(dirname "$0")/../.." && pwd)/skills/moai/workflows/sync/doc-execution.md"
snapshot_line="$(grep -n 'sync_snapshot_id' "$doc" | head -1 | cut -d: -f1)"
fanout_line="$(grep -n 'drafter fan-out launches CONCURRENTLY' "$doc" | head -1 | cut -d: -f1)"

test -n "$snapshot_line"
test -n "$fanout_line"
test "$snapshot_line" -lt "$fanout_line"
grep -q 'HEAD SHA, porcelain-v2 status, diff hash, SPEC artifact hashes' "$doc"
grep -q 'does not require a future Phase 11 output' "$doc"
grep -q 'missing or stale snapshot is a blocker' "$doc"
echo 'PASS: sync drafters consume one immutable snapshot before parallel launch'
