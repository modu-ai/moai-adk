# Progress — SPEC-MOAI-AGENTIC-LOOP-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-07-07
- plan_audit_iter1: FAIL 0.83 (Tier L threshold 0.85; MP 4/4 PASS; Testability 0.70 drag; §D.6 4-artifact deviation ACCEPTED)
- plan_audit_iter1_fixes: D1-D10 + B8/M2 budget clarification applied 2026-07-07 (plan-phase artifact edits only; no skill files touched)
- plan_audit_iter2: PASS-WITH-DEBT 0.89 (all D1-D10 resolved; two literal-level debts)
- plan_audit_iter2_debts: D11 (AC-MAL-016 regex collateral match on harness-level line — threshold-context-tightened, verified 4 in-scope matches / 0 collateral) + D1r (plan.md §E E2 diff form aligned to `--exclude=.moai`, exclusion noted as load-bearing) — resolved 2026-07-07
- plan_status: audit-ready (iter-2 PASS-WITH-DEBT 0.89, debts D11/D1r resolved)
- artifacts: spec.md + plan.md + acceptance.md + progress.md (4 — Tier L artifact-count deviation disclosed in spec.md §D.6)

## §E.2 Run-phase Evidence

- baseline_sha: f7b55e637b3b8d0affb1e24176db84934e3064e1
- baseline_pinned_at: 2026-07-07 (M1 pre-flight — HEAD of main at run-phase start; `git rev-list --count --left-right origin/main...HEAD` → `0 0`)

### M1 pre-flight records

- **orchestration-mode-selection.md mirror divergence characterization (REQ-MAL-042 / AC-MAL-R03)**: pre-flight `diff .claude/rules/moai/workflow/orchestration-mode-selection.md internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` → exactly 1 content hunk (`333c333`, §I.3 "Installation status (honest reference)" paragraph — live copy names maintainer-repo template paths + wrapper + Windows variant; template copy carries the neutral shortened variant) + 1 trailing blank line at template EOF (`334a335`). Classification: pre-existing INTENTIONAL template-neutrality variant, NOT stale content — a full re-sync would either leak maintainer-specific path detail into the shipped template or strip valid detail from the live copy. Resolution: this SPEC's delta (§G.1 crosswalk + §B.1 threshold-SSOT sentence) landed IDENTICALLY in both copies; the pre-existing §I.3 hunk is preserved unchanged in both. Residual divergence documented here as the structured report (in lieu of full re-sync); no silent divergence growth — the delta regions diff clean.
- **Junk-dir cleanup (plan.md §F M1(b))**: `internal/template/templates/.claude/skills/moai/workflows/.moai/` (untracked statusline runtime-state leak; `git ls-files` empty) deleted best-effort at M1. The `--exclude=.moai` flag in parity checks remains the load-bearing mechanism — a live statusline process may re-create the directory.

_(per-AC evidence matrix + verbatim command outputs populated at M6)_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
