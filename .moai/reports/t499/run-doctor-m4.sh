#!/bin/sh
# M4 real-binary confirmation for SPEC-CODEX-PARTIAL-WIRING-001 (card t499).
#
# Runs `moai doctor` against a half-wired scratch project in both PATH
# branches. The cwd is set inside this script because `moai doctor` reads the
# project root from the process cwd (internal/cli/doctor.go:244) and the
# harness does not carry a cwd across shell invocations.
set -u
BIN=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499/bin/moai
SCRATCH=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/72de2805-bbc4-4f5e-a911-e5b1150d316d/scratchpad/t499-m4
PROJ="$SCRATCH/plain"
OUT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499/.moai/reports/t499

cd "$PROJ" || exit 99

"$BIN" doctor > "$OUT/doctor-halfwired-codex-present.txt" 2>&1
echo "codex-present rc=$?"

PATH=/usr/bin:/bin "$BIN" doctor > "$OUT/doctor-halfwired-codex-absent.txt" 2>&1
echo "codex-absent rc=$?"
