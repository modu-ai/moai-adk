# t1894 — plan hold (Tier L ceiling hit at 3 rounds, hold-and-split)

Card t1894, SPEC-CARD-MODE-JUDGMENT-001, branch `WT-fail-safe-serial`. The plan audit ceiling for Tier L (3 rounds) is reached with a FAIL on the latest round. By `harness.plan_audit_ceiling_policy` (`hold-and-split`) the lane makes no further correction and no further audit. `moai spec ceiling SPEC-CARD-MODE-JUDGMENT-001 --evidence .moai/reports/t1894` (read-only, run by the lane after round 3) printed: `ceiling HIT at 3 rounds - latest verdict "FAIL" (admitted=false) - disposition "hold"` and `split required by plan_audit_ceiling_policy.on_final_hit=hold-and-split`. The outcome record (`--record`) was NOT written by the lane: it writes into the SPEC's progress record, which belongs to the plan specialist.

## The three rounds

| Round | Report file | Audited HEAD | Score | Blocking | Receipt | Required codex |
|---|---|---|---|---|---|---|
| 0 | `plan-audit.md` | `cc109a75b6` | 0.65 | 9 (D1-D9), MP-9 failed | `rcpt-39ba8947c14db9823fd47f7e` | fail |
| 1 | `plan-audit-iter1.md` | `4834700468` | 0.79 | 4 (D16-D19) | `rcpt-2adbd9564960a6f7da294042` | fail |
| 2 | `plan-audit-iter2.md` | `cb71f2b475` | 0.88 | 1 (D23) | `rcpt-55ffb196ab252308a7c4a3be` | fail |

The score rose every round and is now above the Tier L threshold 0.85; the verdict is FAIL because the required codex gate failed and one blocking defect stands. Each round closed the previous round's defects (D1-D22 closed) and each new blocking defect was a residual of the previous fix, not a repeat.

## What is open (from `plan-audit-iter2.md`, read by the lane)

- **D23 (BLOCKING).** The path half of AC-MJ-020 (`closed-world-check.py:96`, `p not in allowed and not p.startswith('.moai/')`) exempts everything under `.moai/`, so a widening written into a tracked projection of the changed documents (`.moai/policies/**`, `.moai/workflows/gtd.md`) passes every release-blocking criterion, while four sentences of the SPEC claim no other path may change (acceptance.md AC-MJ-020 title and AC-MJ-006, spec.md R-8, design.md D-10). The auditor and the independent codex leg both reproduced it. Fix named by the auditor: narrow the exemption to the paths a run legitimately writes (the SPEC directory, `.moai/reports/`, `.moai/state/`, `.moai/logs/`), add a known failing input cell, re-pin the two SHA-256 values in E-020d; or keep the exemption and rewrite the four statements to say `.moai/` is not covered.
- **D25 (SHOULD-FIX, optional).** The order-blind line comparison does not see a moved code fence (closing fence of C moved behind the new Mode judgment paragraph); codex reproduced it as P2. Fix: a fifth number in `lines` (fence lines per file against the base count).
- **D24 (MINOR).** Control cell E-013c does not reproduce under the shell-function grep (ugrep) but does under `/usr/bin/grep`; add one line to section 0.
- **Q7 (FOUNDER, empty verdict).** What establishes "the operator has delegated dispatch". Independent of the audit: it blocks the AUTONOMOUS form of the plan to run Kickoff until the operator answers or states that the ruling's phrase is the whole condition. No requirement or milestone of the SPEC waits on it.

## Options for the leader / operator (the lane does not choose)

A. **One scoped correction round plus one final audit under an operator ceiling exception** (same form as the exceptions already granted on other cards). Scope: D23 (plus D25 and the D24 line in the same edit): `closed-world-check.py`, acceptance.md AC-MJ-020 E-020b/E-020d, and the wording of AC-MJ-006, spec.md R-8, design.md D-10/D-11. The fixed probes for before/after are named in the iteration 2 report's Operational Notes. The final audit file would be `plan-audit-iter3.md`.
B. **Split the SPEC along the card's four scope items** (proposal required by the policy): (1) mode judgment step and the issuance procedure (`--classification-file`, `--spec` at pick; scope items 1 and 2, no Go-guard interaction beyond the ledger rows they pin), (2) the delegated-dispatch promotion and free-lane definition (scope item 3 with REQ-MJ-007/017, carries the Q3 ruling d314 and the open Q7), (3) the renewal-rule scoping a672 (scope item 4, REQ-MJ-010 with the ten cron sites and the stub budget). The closed-world check (the source of D16, D23 and D25) would then apply to smaller file sets. This is a larger change than A and re-opens plan audits for three SPECs.
C. **Accept D23 and D25 as dispositioned residual risk.** Requires the operator to waive the required codex gate for this finding class; the lane cannot do that and does not recommend it without an explicit decision.

Lane recommendation: A. D23 is one narrow path predicate plus four sentences, D25 is one extra number, the score is above the threshold, and every earlier defect stays closed. It is a recommendation, not a decision: the exception is the operator's.

## Preserved

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`, branch `WT-fail-safe-serial`. Nothing deleted. See `.moai/reports/t1894/handoff.md` for the resume procedure.
