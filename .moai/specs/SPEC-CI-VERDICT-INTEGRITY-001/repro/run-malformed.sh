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
cd "$HERE/malformed"
sh "$HERE/../../../../scripts/ci-mirror/validate-required-checks.sh"
