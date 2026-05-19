# Progress — SPEC-V3R5-CLAUDE-REFRESH-001

## Pre-Run Baseline (T0 — captured 2026-05-19)

### Lint Baseline
- Total findings: **321**
- ERROR count: **237**
- WARN count: **84**
- Baseline file: `/tmp/lint-baseline-w0.json`

### AskUserQuestion Paraphrase Baseline (N)
- Measured N: **29** lines
- Computed AC-CLR-006 target: **≤ ceil(29 × 0.3) = ≤ 9** lines (post-T6)

### Template Version
- Verified: `Version: 14.0.0 (Agency v3.2 + Harness Design Integration)`
- T7 will bump to: `Version: 14.2.0 (Architecture Truth + W0 Bundle A+B)`

### File Presence Baseline
- `expert-mobile.md` exists: YES (confirmed — will be deleted in T4)

---

## Per-Task Verification

| Task | AC | Status | Notes |
|------|----|--------|-------|
| T0 | AC-CLR-008 (baseline) | PASS | 321 total (237 E + 84 W), N=29 |
| T1 | AC-CLR-001 | PASS | grep count=1 (startup\|resume\|clear\|compact) |
| T2 | AC-CLR-002 | PASS | grep count=1 (Write\|Edit\|MultiEdit) |
| T3 | AC-CLR-003 | PASS | Phase 3: expert-backend=0, manager-develop>=1, dormant in auto-workflow=1 |
| T4 | AC-CLR-004 | PASS | file deleted, grep=0, footnote present |
| T5 | AC-CLR-005 | PASS | max_results=0, select:AskUserQuestion...=1 (in §1 Deferred Tool bullet) |
| T6 | AC-CLR-006 | PASS | N=29→9 lines (≤9 target met exactly), all 3 docs cite askuser-protocol.md |
| T7 | AC-CLR-007 | PASS | Version: 14.2.0=1, Last Updated: 2026-05-18=1, Changes in v14.2.0=1 |
| T8 | AC-CLR-008 (delta) | PASS | NEW_COUNT=0, post-run 312 findings (-9 from expert-mobile removal) |

---

## Final Gate (T8)

- T0 baseline: 321 findings (237 ERROR + 84 WARN)
- T8 post-run total: 312 findings (229 ERROR + 83 WARN)
- T8 NEW_COUNT: **0** (PASS — no new findings introduced by SPEC)
- Note: -9 findings = expert-mobile.md LR-08 drift findings dissolved by T4 deletion
