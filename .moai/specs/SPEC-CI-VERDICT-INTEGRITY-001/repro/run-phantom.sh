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
cd "$HERE/phantom"
sh "$HERE/../../../../scripts/ci-mirror/validate-required-checks.sh"
