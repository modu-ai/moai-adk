#!/bin/sh
# MoAI name-switch shim, installed 2026-11-10 by the ADK leader on the user's instruction
# ("use the name moai for the Rust build, keep Go as moai.old").
#
# `moai` now means the Rust build. The Go build is kept as moai.old and still serves the
# subcommands only it has: hooks, the status line, the MCP server and the harness commands that
# running Claude Code sessions call. Without this routing the Rust binary would answer
# `moai hook ...` with exit 2, which Claude Code treats as "block this tool call".
#
# v2 (SPEC-20261119-switch-script-v2, C-33 S2a). Differs from v1 in three words: `statusline` goes
# to Rust, and `spec` and `session` go to Rust except for the actions only the Go build has.
#
# v3 (SPEC-20261119-switch-script-v3, C-136). Differs from v2 in one word: `bundle` goes to Rust
# (the M-Flow actions `ready`, `check`, `body`, `build`, `verify` and its help) except for the
# actions only the Go build has: `add`, `remove` and `list` of the opt-in skill bundles.
#
# Undo: put back the copy the apply step saved beside this file:
#   mv "<this folder>/moai.before-<stamp>" "<this folder>/moai"
RUST="${MOAI_RUST_BIN:-$HOME/.moai-cli/path/moai-cli}"
GO="${MOAI_GO_BIN:-$HOME/go/bin/moai.old}"

case "${1:-}" in
  # Launcher words always go to Rust.
  cc|glm|codex|omp|-l|-a|--leader|--agent)
    exec "$RUST" "$@" ;;
  # v2: `statusline` is no longer listed below, so it reaches Rust. `spec` and `session` are split
  # by action: the actions only the Go build has stay with Go (Go-generation projects still call
  # `moai spec lint`, `moai spec audit` and `moai session list`), every other call is Rust's.
  # v3: `bundle` is split the same way: Go keeps `add`, `remove` and `list`, Rust gets the rest.
  spec|session|bundle)
    case "$1:${2:-}" in
      spec:status|spec:drift|spec:view|spec:lint|spec:audit|spec:archive|spec:ceiling) ;;
      session:register|session:heartbeat|session:deregister|session:list|session:purge|session:current|session:doctor) ;;
      bundle:add|bundle:remove|bundle:list) ;;
      *) exec "$RUST" "$@" ;;
    esac
    if [ -x "$GO" ]; then
      exec "$GO" "$@"
    fi
    echo "moai: '$1' belongs to the Go build, which was not found at $GO" >&2
    exit 1 ;;
  # Subcommands only the Go build has.
  af|agent|ast-edit|ast-grep|chain|ci-verdict|clean|completion|config|constitution|contract|decision|doctor|epic|factory|feedback|fetch|gate|github|goal|graph|gtd|handoff|harness|hook|inbox|init|integration|inventory|jev|jevmeasure|loop|lsp|mcp|mcp-server|memory|migrate|migration|moai|mx|paths|plan|pr|preference|profile|research|slot|state|status|telemetry|that|todo|tokens|tool-policy|verify|version|web|workflow|worktree)
    # #3021: a long-lived MCP server must keep its own audit jobs; with the shim on PATH it always
    # reads as stale and would hand writes to a one-shot process that loses them.
    case "$1" in mcp-server) export MOAI_MCP_ONESHOT=1 ;; esac
    if [ -x "$GO" ]; then
      exec "$GO" "$@"
    fi
    echo "moai: '$1' belongs to the Go build, which was not found at $GO" >&2
    exit 1 ;;
esac
exec "$RUST" "$@"
