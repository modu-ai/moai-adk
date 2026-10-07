#!/bin/sh
# t1534 repro E27 — positive control for E24. Runs the repo's validator
# (unmodified, by committed path) from a cwd whose .github/required-checks.yml
# is IDENTICAL to the phantom fixture except the context value: the phantom
# `Test (windows-latest)` is replaced by the genuinely published
# `Test (ubuntu-latest)` (ledger E13).
#
# Expectation (both pre- and post-repair): exit 0. Post-M2 this is the
# two-directional discriminator — the publishability dimension must reject the
# phantom fixture (E24) while passing this control; a context-blind
# implementation cannot separate the two fixtures, because their workflow sets
# are identical and only the context value differs.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
# Anchor on the repo root — no ../ climbing (path-traversal scanner finding,
# card t1534 amendment 3). Override with MOAI_T1534_REPO_ROOT if detached.
REPO_ROOT="${MOAI_T1534_REPO_ROOT:-$(git -C "$HERE" rev-parse --show-toplevel 2>/dev/null || true)}"
if [ -z "$REPO_ROOT" ]; then
	echo "run this script from inside the repository (git rev-parse --show-toplevel failed)" >&2
	exit 9
fi
cd "$HERE/phantom-control"
sh "$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
