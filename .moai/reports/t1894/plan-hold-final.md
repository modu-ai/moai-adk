# t1894 — final plan hold (the 8108 exception round ended FAIL)

Card t1894, SPEC-CARD-MODE-JUDGMENT-001, branch `WT-fail-safe-serial`, tree clean at HEAD `a0bcb8d867240a6b56ea857e1c1e0b7fb8db7503` (this commit plus the verdict commit). The operator ceiling exception `d-20261019T062011Z-8108` granted exactly one correction round (D23, D25, D24) and one final plan audit, with "FAIL = hold". The final audit (iteration 3) returned FAIL. By the contract the lane made no further correction and no further audit; the decision below is the operator's, relayed by the leader.

## The four rounds

| Round | File | Score | Blocking | Receipt |
|---|---|---|---|---|
| 0 | `plan-audit.md` | 0.65 | 9 (D1-D9) | rcpt-39ba8947c14db9823fd47f7e |
| 1 | `plan-audit-iter1.md` | 0.79 | 4 (D16-D19) | rcpt-2adbd9564960a6f7da294042 |
| 2 | `plan-audit-iter2.md` | 0.88 | 1 (D23) | rcpt-55ffb196ab252308a7c4a3be |
| 3 (final, 8108) | `plan-audit-iter3.md` | 0.89 | 1 (D28, new) | rcpt-d02259fabfe1b100e4d83659 |

Score rose every round and is above the Tier L threshold 0.85; every verdict is FAIL because the required codex gate failed. D23 and D24 are closed and D25 closed for its named scope; the residual blocker is new each round and each time it is an edge of the surface the previous fix created.

## D28 (BLOCKING, the open item)

Moving C's closing code fence **together with its preceding `moai graph build…` line** keeps every `lines` number at 0 — the fifth number compares (fence, walked-predecessor) pairs and a joint move preserves the pair — while the Mode judgment subsection lands inside the code block. Three SPEC statements claim a moved fence is reported (acceptance.md AC-MJ-020 prose, design.md D-10 row and Unit bullet, the checker docstring) without disclosing the joint-move case. Same defect form as D16/D23: stated coverage ≠ implemented coverage on release-blocking AC-MJ-020.

Independently reproduced by the lane before recording (both constructions on the auditor's manifest-rebuilt copies under `/private/tmp/t1894-aud/`, outside every tree): (1) the lane's own pair-move mutant — C rows `0 0 0 0 0`, fifth 0; (2) the auditor's fixed rebuild `sim2` already sits in the swallowed configuration (closing fence at line 94, behind the `### Mode judgment` heading at 91 and its paragraph) and the checker prints all-zero, exit 0, on it.

## What the operator decides (the lane does not choose)

- **Another exception round.** The candidate fix is small (a sixth number comparing fence pairs against the base side, or disclosing the joint-move limit in the three statements), but this would be the fourth consecutive round where the newest helper surface yields a new edge — the same pattern the leader named on t1755. Convergence is not guaranteed.
- **Split (the ceiling policy's named path).** Per `plan-hold.md` option B: the closed-world check (source of D16/D23/D25/D28) would apply to smaller file sets, or the mode-judgment scope items could go separately from the delegated-dispatch/ Q7-bearing items.
- **Accept the residual as disclosed risk.** Requires waiving the required codex gate for this finding class (the lane cannot do that; option C of `plan-hold.md`).

## Preserved

Tree and branch preserved; lease held and renewed by the lane while the card is alive (operator instruction: held-card renewal continues). Renewal cron active in the lane session. The auditor's scratch `/private/tmp/t1894-aud/` remains on disk — its `rm` was refused by the permission system (recorded in the verdict's Gaps); the lane's two mutant copies under the same directory were also left in place for the same reason.
