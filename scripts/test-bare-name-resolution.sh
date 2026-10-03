#!/bin/sh
# test-bare-name-resolution.sh — the self-test + negative-control wrapper for
# check-bare-name-resolution.sh (SPEC-INIT-SHRINK-001 AC-020 (b), plan M1).
#
# Runs the gate's --self-check mode (isolation scrub + negative control,
# protected-set hash, output line shape — NO real tool runtime) and prints
# its lines verbatim. Exit 0 iff every case passed (RESULT ... fail=0).
# Safe under `go test`; the full measurement is never started here.

set -u

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
gate="$script_dir/check-bare-name-resolution.sh"

if [ ! -f "$gate" ]; then
    echo "FAIL self-test: gate script missing at $gate"
    echo "RESULT pass=0 fail=1"
    exit 1
fi

exec sh "$gate" --self-check
