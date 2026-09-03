# Progress — SPEC-TODO-LANDING-ATTRIBUTION-001

Card t472. Worktree `.claude/worktrees/t472`, branch `WT-landed-drift-detect`.

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored at tree `4bcac7079`: `spec.md`, `plan.md`, `acceptance.md`, this file.
- Tier M. 12 requirements (REQ-TLA-001..012), 12 acceptance criteria (AC-TLA-001..012).
- Milestone order fixed [HARD]: M1 (axis F, the attribution predicate) before M2 (axes A+B, the ref
  chain and its disclosure) — ground in `spec.md` §A.7 and `plan.md` §A.3.
- Ref-chain option **A3** selected; A1, A2, A4 rejected with recorded grounds in `plan.md` §A.1.
- Evidence base: `.moai/reports/t472/premise-recheck.md` and `.moai/reports/t472/axis-bf-measurement.md`,
  both committed in this tree by this lane; fixtures re-run at `4bcac7079` before citation.
- `status: draft`. Run-phase not entered; no implementation code written.

**Lint evidence.** `moai spec lint` and `moai spec lint --strict`, scoped to this SPEC and run with a
binary built from this tree (`go build ./cmd/moai`, rc=0 — not the installed build, which is 190
commits behind; VCI §2.2), both returned `✓ No findings — all SPEC documents are valid`. The lane
observed both; the parent lane session independently re-ran the `--strict` form from the worktree
root with the same binary and observed the same result, so the pass is confirmed rather than reported.

**Gap — whole-corpus lint not observed, and deliberately not pursued.** Three attempts at a
repository-wide `moai spec lint` produced no usable measurement: the first was piped through
`tail -40`, discarding all but the last 40 lines and leaving the result unattributable, and was in any
case taken before the `(maps REQ-…)` clauses were added; the second and third were killed by the parent
lane session (`rc=143`, then exit `144`) because this machine's load average was 27 while
lanes share it, and each left a 0-byte output file. Per lane direction the item is **closed as a gap,
not resolved** — no further attempt is to be made from this lane. A grep over those empty files
returns zero for any pattern and establishes nothing.

The residual this would have covered is bounded by an **argument, not a measurement**: `git status
--short` returns exactly `?? .moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/` — four new files in a new
directory, no existing file modified — so no mechanism is apparent by which this change could alter
another SPEC's findings. That reasoning is recorded as an argument and must not be cited as a corpus
measurement.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
