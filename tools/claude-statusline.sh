#!/bin/bash
# MoAI statusline wrapper (#2383) - renders with the Rust moai-cli by absolute path.
# No PATH lookup of `moai`: the Go `moai` name leaves PATH after PR #2374 (#2383).
# stdin is handed straight to the exec'd process; this script creates no temporary
# file (an exec'd process never runs an EXIT trap, so a temp buffer would leak).

# A regular file that is executable: `-x` alone is true for a directory, which exec
# cannot run (exit 126, no fallback) - #2503.
MOAI_CLI="$HOME/.moai-cli/path/moai-cli"
if [ -f "$MOAI_CLI" ] && [ -x "$MOAI_CLI" ]; then
	exec "$MOAI_CLI" statusline
fi

# Fallback while the Go build is still installed: the same Go render as today.
MOAI_GO="$HOME/go/bin/moai.old"
if [ -f "$MOAI_GO" ] && [ -x "$MOAI_GO" ]; then
	exec "$MOAI_GO" statusline
fi

# Nothing installed - exit silently (Claude Code handles a missing statusline).
exit 0
