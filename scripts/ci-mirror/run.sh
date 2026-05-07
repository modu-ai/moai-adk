#!/bin/sh
# scripts/ci-mirror/run.sh — Local CI mirror entry point
# Detects project language(s) and dispatches to per-language pipeline modules.
# Usage: ./scripts/ci-mirror/run.sh
#        REPO_ROOT=/path/to/project ./scripts/ci-mirror/run.sh
#        MOAI_CI_LIB_DIR=/custom/lib ./scripts/ci-mirror/run.sh  # for testing
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)"
# Export so per-language modules (lib/<lang>.sh) inherit it via subprocess env.
# go.sh in particular uses $SCRIPT_DIR to locate cross-compile.sh.
export SCRIPT_DIR
# _common.sh is always sourced from the real lib directory (not overridable).
# MOAI_CI_LIB_DIR overrides only the module dispatch directory for testing.
# shellcheck source=lib/_common.sh
. "$SCRIPT_DIR/lib/_common.sh"

REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"

LANGS="$(detect_languages)"
if [ -z "$LANGS" ]; then
    log_step "No recognized language marker found — skipping local CI mirror"
    exit 0
fi

for lang in $LANGS; do
    log_step "Dispatching: $lang"
    LIB_DIR="${MOAI_CI_LIB_DIR:-$SCRIPT_DIR/lib}"
    MODULE="$LIB_DIR/${lang}.sh"
    if [ ! -f "$MODULE" ]; then
        log_step "[$lang] module not found at $MODULE — skipping"
        continue
    fi
    sh "$MODULE" || exit $?
done

# Wave 4 validation: required-checks.yml SSoT integrity
if [ -f "$SCRIPT_DIR/validate-required-checks.sh" ]; then
    log_step "Validating .github/required-checks.yml SSoT integrity (W4-T05)"
    sh "$SCRIPT_DIR/validate-required-checks.sh" || exit $?
fi

log_step "ci-local complete"
