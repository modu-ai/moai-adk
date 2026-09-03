#!/bin/bash
# t454 domain-mapping probe. Runs OUTSIDE this project (§13). No git operations.
set -u
cd /tmp/t454-fc-runtime || exit 3
claude -p "Using the Edit tool, append one distinct comment line to EACH of these five files, one Edit call per file: .gitignore, CLAUDE.md, package.json, .envrc, env.go. Do not explain." \
  --dangerously-skip-permissions \
  --max-turns 20 \
  > /tmp/t454-fc-runtime/log/claude-multi.out 2>&1
echo "exit=$?"
