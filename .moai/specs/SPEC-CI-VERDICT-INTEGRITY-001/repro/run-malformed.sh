#!/bin/sh
# t1534 repro E11 — runs the repo's validator (unmodified, by committed path)
# from a cwd whose .github/required-checks.yml is the committed malformed-YAML
# fixture (malformed/.github/required-checks.yml).
#
# RED expectation (pre-repair): yq parse error swallowed by
# `2>/dev/null || true` -> empty lists -> vacuous "All validations passed",
# exit 0. GREEN expectation (post-M2): non-zero naming the parser failure.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
# Anchor on the repo root — no ../ climbing (path-traversal scanner finding,
# card t1534 amendment 3). Override with MOAI_T1534_REPO_ROOT if detached.
REPO_ROOT="${MOAI_T1534_REPO_ROOT:-$(git -C "$HERE" rev-parse --show-toplevel 2>/dev/null || true)}"
if [ -z "$REPO_ROOT" ]; then
	echo "run this script from inside the repository (git rev-parse --show-toplevel failed)" >&2
	exit 9
fi
cd "$HERE/malformed"
sh "$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
