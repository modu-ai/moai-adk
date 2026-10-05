#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
hook="$repo_root/.claude/hooks/moai/sync-phase-quality-gate.sh"
docs="$repo_root/.claude/skills/moai/workflows/sync/quality-gates-quality.md"
grep -q 'not a vulnerability scan' "$docs"
grep -q 'Dependency manifest-change observation' "$hook"

root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT
git -C "$root" init -q
git -C "$root" config user.email test@example.com
git -C "$root" config user.name test
printf 'module example.test\n\ngo 1.23\n' > "$root/go.mod"
printf 'package main\n' > "$root/main.go"
git -C "$root" add go.mod main.go
git -C "$root" commit -qm 'init'
printf 'module example.test\n\ngo 1.24\n' > "$root/go.mod"
printf 'package main\n\nvar changed = true\n' > "$root/main.go"
git -C "$root" add go.mod main.go
git -C "$root" commit -qm 'docs(t611): sync-phase dependency observation fixture'

output=$(cd "$root" && CLAUDE_PROJECT_DIR="$root" bash "$hook" <<< '{}')
grep -q 'deps_modified=1' <<< "$output"
grep -q 'deps_modified=1' "$root/.moai/logs/sync-quality-gate.log"
printf 'PASS: manifest observation is explicit and reports changed dependency files\n'
