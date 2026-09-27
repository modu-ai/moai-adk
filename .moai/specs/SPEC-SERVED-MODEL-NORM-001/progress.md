# Progress — SPEC-SERVED-MODEL-NORM-001 (card t1287)

## §A Premise (CLAUDE.local.md §30, three steps)

- Base: local develop `37dc766b9` (worktree `.claude/worktrees/t1287`, branch `WT-served-model-norm`).
- Premise alive. Binary built from this tree, `moai doctor --check "Served Model" --verbose`: `swept 2334 subagent transcripts: ok 1295, served_drift 982, unknown 11, unmapped 46`.
- Drift shapes: `expected=claude-opus-5[1m] served=[claude-opus-5]` ×295, `expected=claude-fable-5-1[1m] served=[claude-fable-5-1]` ×2, `expected=inherit served=[claude-sonnet-5]` ×1 → 298 false drifts (t1282 recorded 292). `expected=opus[1m] served=[glm-5.3-flash]` ×125 are true drifts.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-28

## §F Phase 4 Mode Selection

Decision: direct — Tier S, one function plus one test table, lane-executed under §31 autonomous kickoff.

## §E.2 Run-phase Evidence

- RED: `go test ./internal/hook/ -run TestServedModel_Classify -count=1` → 3 FAIL (e, f, h: `verdict = "served_drift"`); control g passed.
- GREEN: `go test ./internal/hook/ -run TestServedModel -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 5.241s`.
- Sweep with fixed binary: `swept 2335 …: ok 1592, served_drift 684, unknown 12, unmapped 47`; drift rows with `expected=…[1m]` or `expected=inherit`: 0.
