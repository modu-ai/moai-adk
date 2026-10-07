#!/bin/sh
# t1534 gate round 4 — P2: matrix.exclude subtraction is missing from the
# validator's publishable-name set. The fixture declares a product matrix
# whose exclude removes `Test (windows-latest, 1.23)`, while the required
# SSoT demands exactly that excluded name — GitHub will NEVER publish it.
#
# RED expectation (pre-repair): exit 0 — the validator counts the excluded
# combination as publishable and the phantom context passes silently (the
# false-green class this SPEC exists for).
# GREEN expectation (post-repair): non-zero naming the excluded phantom.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="${MOAI_T1534_REPO_ROOT:-$(git -C "$HERE" rev-parse --show-toplevel 2>/dev/null || true)}"
if [ -z "$REPO_ROOT" ]; then
	echo "run this script from inside the repository (git rev-parse --show-toplevel failed)" >&2
	exit 9
fi
cd "$HERE/matrix-exclude"
sh "$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
