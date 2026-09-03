#!/bin/bash
# t454 runtime probe driver. Runs OUTSIDE this project (§13): cwd is /tmp/t454-fc-runtime.
# No git operations of any kind. Exists only because the worktree-isolated session
# cannot change its own cwd, and `claude` derives the project dir from cwd.
set -u
cd /tmp/t454-fc-runtime || exit 3
target="${1:?target file required}"
claude -p "Use the Edit tool to append one line to the file ${target} in this project: a comment line. Do not explain, just make the edit." \
  --dangerously-skip-permissions \
  --max-turns 6 \
  > "/tmp/t454-fc-runtime/log/claude-$(basename "$target").out" 2>&1
echo "exit=$?"
