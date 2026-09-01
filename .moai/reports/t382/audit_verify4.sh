#!/bin/sh
# Lint the non-terminal V3R5 bucket (lint terminalStatusEnum sense) under --strict.
set -u
LIST="SPEC-ASTGREP-BREADTH-001 SPEC-ASTGREP-LANG16-001 SPEC-CLOCAL-AUDIT-001 SPEC-CONFIG-DEAD-SWEEP-001 SPEC-ERA-H3-NARROWING-001 SPEC-GITFLOW-DOCTRINE-ALIGN-001 SPEC-GLM-EFFORT-REBALANCE-001 SPEC-INTERNAL-ARCH-001 SPEC-KANBAN-BOOTSTRAP-001 SPEC-KANBAN-PR-CARD-TRACEABILITY-001 SPEC-KANBAN-QUEUE-PR-SYNC-001 SPEC-KANBAN-TODO-CLI-001 SPEC-KANBAN-WORKTREE-001 SPEC-SEC-SCAN-SURFACE-001 SPEC-V3R5-INIT-WIZARD-EXPANSION-001"
FILES=""
for s in $LIST; do FILES="$FILES .moai/specs/$s/spec.md"; done
./bin/moai spec lint --strict --json $FILES > /tmp/t382lint_strict.json 2>/tmp/t382lint_strict.err
echo "strict rc=$?"
./bin/moai spec lint --json $FILES > /tmp/t382lint_plain.json 2>/dev/null
echo "plain rc=$?"
