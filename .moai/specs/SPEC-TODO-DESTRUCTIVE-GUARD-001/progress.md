# Progress — SPEC-TODO-DESTRUCTIVE-GUARD-001

Card: t330 · Tier M · Route A (Tier S/M default)

## §E.1 Plan-phase Audit-Ready Signal

| Item | Value |
|---|---|
| SPEC ID | SPEC-TODO-DESTRUCTIVE-GUARD-001 (regex self-check: `PASS`) |
| Tier | M — 10 files, est. 400-700 LOC (`plan.md` §A) |
| Artifacts | spec.md + plan.md + acceptance.md (Tier M set) + this file |
| Requirements | 16 (at ceiling 16) |
| Acceptance criteria | 16 (at ceiling 16) |
| Base tree | `812ee01fc`, branch `WT-todo-destructive-guard` |
| Status | `draft` |

Decisions settled at plan-phase, each from a cited measurement:

- **Decision 1** — additive archive, not a fourth `BacklogState`. `spec.md` §B.1 (M1/M2/M3); `plan.md` §B. Confirmed independently by plan-audit iter1; not reopened.
- **Decision 2** — t330 owns the reversal and the refusal seam; t331 owns the persisted landing-state field. `spec.md` §B.2; `plan.md` §C.
- **Decision 3** — the archive is deliberately **included** in `export-json` (the downgrade route), with a stderr disclosure of the downgrade loss; live-queue readers exclude it. `spec.md` §C.5.

Measurements recorded at plan-phase that contradict the framing the card was dispatched with:

1. The existing landed primitive fails in **both** reachable modes, and correcting the ref moves the failure rather than removing it. As shipped (`LandedRef = "origin/main"`, `prlink_landed.go:28`) it answers **false** for every develop-integrated card — `origin/main` names t306 in 0 commits — so a default-on refusal would block everything. After the obvious ref correction it answers **true** on any mention — `origin/develop` names t306 in 13 commits, the earliest being the run commit `3cb258d62` — so it would have passed the premature `done` silently (`spec.md` §A.4). The opt-in ruling rests on the predicate answering the wrong question, not on the ref.
2. `LandedRef` is stale under the develop-integration git-flow. Not fixed here — shared with `moai todo pr`; declared out of scope in `spec.md` §D and left as a candidate follow-up card.
3. There is no live JSON engine. `openEngine` (`backlog_store.go:437-455`) falls through to the SQLite engine on every path, so a "both live backends" comparison is unreachable (`plan.md` §D).
4. The `moai todo` surface is **15** verbs, not 14 — the doctrine table omits `why` (`todo.go:137-141`).
5. `--expect` is carried by `next`, `edit`, `drop`, `undrop` — **not** by `move`.

### plan-audit history

| Iteration | Verdict | Score | Disposition |
|---|---|---|---|
| iter1 | FAIL | 0.75 (Tier M threshold 0.80) | 5 blocking defects (D1-D5) + 3 non-blocking (D6-D8); all addressed — see below |
| iter2 | PASS-WITH-DEBT | 0.9375 (monotonic +0.1875, no dimension regressed) | Clarity 0.75→1.00, Testability 0.50→0.75, Traceability 0.75→1.00; 7/7 MUST-PASS. Decision 3 reviewed and approved. 3 debts (S1, N1, N2) + 2 optional (N3, N4) — all 5 landed in iter3 |

iter3 debt closure: S1 (the two `acceptance.md` citations the D7 sweep missed, plus the false "four refreshed" claim corrected above) · N1 (AC-TDG-015 captures stdout/stderr separately — a merged `2>&1` would pass a disclosure printed to the machine-read stdout line) · N2 (AC-TDG-007 asserts the exact output `t1: no findings`, since `todo_why.go:34-35` echoes the id and defeats a grep) · N3 (`move` declares four flags, not two) · N4 (softened "the only point we control" — the exported artifact is also ours and could carry the warning; not built, not foreclosed).

D1 (spec.md §A.4 rewritten with both modes) · D2/D4 (Decision 3: REQ-TDG-015, AC-TDG-015, M5 rewritten to budget the disclosure) · D3 (REQ-TDG-006, AC-TDG-006, §A.3 rewritten to the reachable configuration) · D5 (15-verb list re-derived from `AddCommand`) · D6 (`--expect` set corrected — and `move` removed, which the defect report itself carried wrongly) · D7 (citations refreshed in spec.md and plan.md; t306 count 10→13, my original was `head -10` truncation) · D8 (REQ-TDG-016, AC-TDG-016: restore empties the entry).

> Correction (iter3): the iter2 line above originally read "four citations refreshed", which was an unobserved completion claim — two of the four (`todo.go:341` and `todo.go:352-354`) survived untouched in `acceptance.md` because the D7 sweep covered `spec.md` and `plan.md` only. Caught by iter2 audit as S1 and landed in iter3. The failure is the one D7 itself named, and it survived because the report read finished.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
