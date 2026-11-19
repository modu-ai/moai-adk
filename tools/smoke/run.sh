#!/usr/bin/env bash
# The smoke harness the lanes and the auditors share (SPEC-20261119-smoke-harness).
#
#   tools/smoke/run.sh <area> [--bin PATH] [--repo ROOT] [--keep]
#   area: install | update | project | launcher | all
#
# Boots the built binary (default: $CARGO_TARGET_DIR/debug/moai-cli, else target/debug/moai-cli)
# in an empty temporary HOME and a temporary git project, prints one trace line per step and
# ends with `smoke: <area> pass|fail steps=<n> ...` (exit 0 / 1; 2 = usage). It never reads or
# writes the real HOME. Details: tools/smoke/smoke.py.
set -u
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec python3 -I "$here/smoke.py" "$@"
