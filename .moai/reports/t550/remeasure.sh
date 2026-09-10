#!/usr/bin/env bash
# t550 post-change re-measurement — the SAME instrument as
# .moai/reports/t550/baseline.md: a binary built from this tree, seven fixtures
# differing only in their linter config file, npx/npm/node shadowed by exit-0
# stubs on PATH (so nothing reaches the network), and the gate's own run summary
# read as the instrument. No git commands here; branch/HEAD state is read
# separately as plain commands.
set -u

WT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t550
SP=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/cbea09a7-48ea-46ea-bb4f-bfbba105cfdb/scratchpad
BIN="$SP/moai-after"
STUB="$SP/stubs"
FX="$SP/fx-after"

cd "$WT" || exit 1

go build -o "$BIN" ./cmd/moai || exit 1
echo "go build -o \$SP/moai-after ./cmd/moai: exit 0"

rm -rf "$STUB" "$FX"
mkdir -p "$STUB"
for b in npx npm node; do
  printf '#!/bin/sh\nexit 0\n' > "$STUB/$b"
  chmod +x "$STUB/$b"
done

NAMES=(eslintprj biomeprj oxlintprj ox2 ox3 ox4 bareprj)
CFGS=(eslint.config.js biome.json .oxlintrc.json .oxlintrc.jsonc oxlint.config.ts oxlint.config.mts "")

for i in "${!NAMES[@]}"; do
  d="$FX/${NAMES[$i]}"
  mkdir -p "$d"
  printf '{"name":"%s"}\n' "${NAMES[$i]}" > "$d/package.json"
  printf 'export const x = 1;\n' > "$d/index.js"
  if [ -n "${CFGS[$i]}" ]; then printf '{}\n' > "$d/${CFGS[$i]}"; fi
done

for i in "${!NAMES[@]}"; do
  d="$FX/${NAMES[$i]}"
  echo "===== fixture ${NAMES[$i]} (linter config: ${CFGS[$i]:-none — CONTROL}) ====="
  CLAUDE_PROJECT_DIR="$d" PATH="$STUB:/usr/bin:/bin:/usr/sbin:/sbin" "$BIN" gate 2>&1
  echo "exit=$?"
done
