#!/bin/bash
# MoAI statusline wrapper (#2383, #2871) - renders with the Rust moai-cli.
# Looks in the HOME state folder where the install is (`~/.moai` after the move, `~/.moai-cli`
# before it), then for the name `moai-cli` on PATH. No PATH lookup of `moai`: the Go `moai`
# name leaves PATH after PR #2374 (#2383).
# stdin is handed straight to the exec'd process; this script creates no temporary
# file (an exec'd process never runs an EXIT trap, so a temp buffer would leak).

# A regular file that is executable: `-x` alone is true for a directory, which exec
# cannot run (exit 126, no fallback) - #2503.
# Only an absolute path counts: with `.` or an empty entry in PATH, `command -v` answers the bare
# name `moai-cli` (or a shell function of that name), and exec would then run whatever sits in
# the project folder the status line is started from - #3008 audit B1.
for MOAI_CLI in "$HOME/.moai/path/moai-cli" "$HOME/.moai-cli/path/moai-cli" "$(command -v moai-cli)"; do
	case "$MOAI_CLI" in /*) ;; *) continue ;; esac
	if [ -f "$MOAI_CLI" ] && [ -x "$MOAI_CLI" ]; then
		exec "$MOAI_CLI" statusline
	fi
done

# Fallback while the Go build is still installed: the same Go render as today.
MOAI_GO="$HOME/go/bin/moai.old"
if [ -f "$MOAI_GO" ] && [ -x "$MOAI_GO" ]; then
	exec "$MOAI_GO" statusline
fi

# Nothing installed - exit silently (Claude Code handles a missing statusline).
exit 0
