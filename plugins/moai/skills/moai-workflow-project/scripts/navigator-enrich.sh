#!/usr/bin/env bash
# navigator-enrich.sh — Project Navigator AST symbol enrichment (the Project Navigator AST enrichment feature).
#
# Sibling to the 001 regeneration script and the 002 audit script. Wraps the
# `moai navigator-enrich` Go entry point which reads 001's capability-map.md
# (header-driven join), walks each row's implementation-path, extracts
# tree-sitter symbols, and writes capability-symbols.{md,json} atomically
# under .moai/project/codemaps/.
#
# Design properties:
#   * Self-contained bash (git + the moai binary; NO jq). The tree-sitter
#     parsing happens inside the Go entry point; this script only resolves the
#     project root, performs the capability gate, and invokes the binary.
#   * Capability gate (the capability-gate REQ): if capability-map.md is
#     absent, emit an info log and exit 0 WITHOUT writing any output file.
#   * Atomic writes (.tmp -> mv) and idempotence live inside the Go entry
#     point; this script inherits both.
#   * Fail-open on every error mode: exit 0 always (never aborts /moai codemaps).
#   * Provenance uses git (commit SHA + committer date), never wall-clock
#     (the governing REQ idempotence).
#
# Boundary: writes ONLY .moai/project/codemaps/{capability-symbols.md,json} and
# appends ONLY to .moai/logs/navigator-astx.log. It NEVER touches 001/002
# outputs under .moai/project/navigator/ or the 002 audit outputs, or any LSEL
# surface.
#
# Optional env vars:
#   CLAUDE_PROJECT_DIR  project root (defaults to $PWD).
#   NAVIGATOR_PRE_RENAME_BARRIER  path; test-only synchronized barrier for the
#       atomic-rename fixture (mirrors 001's pattern).
#
# Exit codes: 0 always (fail-open).
set -u

ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
CAP_MAP="$ROOT/.moai/project/navigator/capability-map.md"
ASTX_LOG="$ROOT/.moai/logs/navigator-astx.log"

mkdir -p "$(dirname "$ASTX_LOG")" 2>/dev/null || true

# Capability gate (the governing REQ vs the governing REQ).
if [ ! -f "$CAP_MAP" ]; then
    printf '%s navigator-astx: capability-map.md absent at %s; skipping AST enrichment\n' \
        "$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo '?')" "$CAP_MAP" >> "$ASTX_LOG" 2>/dev/null || true
    echo "navigator-astx: capability-map.md absent, skipping AST enrichment" >&2
    exit 0
fi

# Resolve the moai binary (PATH first, then $HOME/go/bin, then GOPATH/bin).
MOAI_BIN="$(command -v moai 2>/dev/null || true)"
if [ -z "$MOAI_BIN" ]; then
    if [ -x "$HOME/go/bin/moai" ]; then
        MOAI_BIN="$HOME/go/bin/moai"
    elif [ -n "${GOPATH:-}" ] && [ -x "$GOPATH/bin/moai" ]; then
        MOAI_BIN="$GOPATH/bin/moai"
    else
        printf '%s navigator-astx: moai binary not found on PATH; skipping AST enrichment\n' \
            "$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo '?')" >> "$ASTX_LOG" 2>/dev/null || true
        echo "navigator-astx: moai binary not found, skipping AST enrichment" >&2
        exit 0
    fi
fi

# Invoke the Go entry point. It performs the header-driven join, tree-sitter
# extraction, aggregation, and atomic writes internally. Fail-open: any
# non-zero exit is logged and swallowed.
if ! "$MOAI_BIN" navigator-enrich --project-root "$ROOT" 2>>"$ASTX_LOG"; then
    echo "navigator-astx: enrich entry point reported an error (logged); continuing" >&2
fi

exit 0
