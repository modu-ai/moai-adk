#!/bin/sh
# t1534 repro E24 — runs the repo's validator (unmodified, by committed path)
# from a cwd whose .github/required-checks.yml is VALID YAML that carries a
# phantom required context ("Test (windows-latest)" — a name no workflow
# publishes on main-targeting PRs; ledger E21/E13).
#
# RED expectation (pre-repair): exit 0 — no validator dimension inspects
# required-context publishability, so the phantom passes silently.
# GREEN expectation (post-M2): non-zero naming the phantom context.
set -eu
HERE="$(cd "$(dirname "$0")" && pwd)"
# Anchor on the repo root — no ../ climbing (path-traversal scanner finding,
# card t1534 amendment 3). Override with MOAI_T1534_REPO_ROOT if detached.
REPO_ROOT="${MOAI_T1534_REPO_ROOT:-$(git -C "$HERE" rev-parse --show-toplevel 2>/dev/null || true)}"
if [ -z "$REPO_ROOT" ]; then
	echo "run this script from inside the repository (git rev-parse --show-toplevel failed)" >&2
	exit 9
fi
cd "$HERE/phantom"
sh "$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
