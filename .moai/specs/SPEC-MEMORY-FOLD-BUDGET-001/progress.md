# progress.md — SPEC-MEMORY-FOLD-BUDGET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-04
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + progress.md + decision-index.md (conditional, `interview.decision_gate: on`) + fixtures/store-A/ (synthetic, read-only acceptance input)
- req_count: 15
- ac_count: 15
- baseline_tree: 2f492df19

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build: `./bin/moai`, built from this tree and invoked by path, commit stamped `2f492df19` through `-ldflags` (`./bin/moai version` prints `v3.1.3   2f492df19   built 2026-10-04`); measured tree HEAD: `2f492df19`. The Go VCS stamp of a plain `go build` in this worktree named another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so it was not trusted.
- `./bin/moai spec lint SPEC-MEMORY-FOLD-BUDGET-001` → stdout `✓ No findings — all SPEC documents are valid`, exit 0.
- `./bin/moai spec lint --strict SPEC-MEMORY-FOLD-BUDGET-001` → same stdout, exit 0.
- positive control (a deliberately bad scratch spec, `phase: plan`, uppercase `If … then`, no AC, no exclusions) → exit 1, `1 error(s), 4 warning(s)` — the lint can fail.
- `moai spec audit` (MCP, filter this SPEC, `project_root` this worktree) → era `V3R6`, one INFO `EraAutoDetected`, `modern_era_clean: 1`, no drift.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
