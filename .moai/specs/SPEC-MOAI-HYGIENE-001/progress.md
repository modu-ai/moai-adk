# progress.md — SPEC-MOAI-HYGIENE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-MOAI-HYGIENE-001 v0.3.0 authored 2026-10-05 by manager-spec (card t1518, plan phase, worktree `WT-audit-log-gc` @ develop `6643c7bba`).
- Revision history: v0.1.0 (plan-audit iter 1/3 audited `8039ea714`, FAIL 0.73, D1–D14) → v0.2.0 (D1–D14 remediation: MP-8 evidence ledger, REQ consolidation 19 → 16, registry-absent classification, mtime prohibition, re-stat-under-lock, symlink refusal, mode precedence; iter 2/3 audited `58c01b4e1`, FAIL 0.81, D15–D22 blocking + D23–D25 optional) → v0.3.0 (D15 rotation apply-mode-only — one decision in REQ-HYG-004, no exemption clauses; D16 staging artifacts as rotator-owned registry entries + under-lock crash recovery; D17 Windows rotation gated on a verified LockFileEx sidecar lock, else skip+retry; D18 lock class excluded from GC scope — empty-file dating impossible + two-live-locks window, upstream parked at Q7; D19 symlink refusal scoped strictly below the resolved root + swap-after-check residual named; D20 transcript-absent ⇒ unmeasured + `HygieneHeartbeatStaleWindow`; D21 per-class shape/dating/grouping table with M3 field-pinning REDs; D22 runtime `testing.Testing()` root guard + escape-pattern greps + L-C3 control; D23 rotation-count bound; D24 `would-probe` + hash exclusion; D25 reclaim limitation + Q8).
- Tier M artifact set complete: spec.md, plan.md, acceptance.md, progress.md, decision-index.md.
- Evidence basis: 2026-10-04 read-only hygiene audit (`hygiene.md` + `hygiene.json`, scratchpad) + plan-phase code-owner mapping (spec.md §G) + plan-audit iter-1 verdict (`.moai/reports/t1518/plan-audit.md`).
- SPEC ID pre-write check: `PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`, executed in Bash); catalogue dedup confirmed (1,048 existing SPECs, no MOAI-HYGIENE entry).
- MP-8 evidence ledger (acceptance.md §A.1, L-001..L-016 + L-C1/L-C2): every cell carries command + verbatim stdout + exit code + tree SHA `8039ea714c988f4264cd8785405dbb1c410370b6`; 14 RB go-test cells observed RED (12× `FAIL ./internal/hygiene [setup failed]` exit 1; 2× `ok … [no tests to run]` exit 0 — empty-sweep reds), 2 RG grep cells observed structurally red (exit 2, targets absent) with seeded positive controls proving the patterns non-vacuous (5 and 1 hits, exit 0).
- Spec lint after revision (measured 2026-10-05, this tree's build): `go run ./cmd/moai spec lint SPEC-MOAI-HYGIENE-001` → `✓ No findings — all SPEC documents are valid` (16 REQ / 16 AC, all maps lines resolved).
- plan-phase scope held: read-only on `internal/` — no implementation files touched.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

