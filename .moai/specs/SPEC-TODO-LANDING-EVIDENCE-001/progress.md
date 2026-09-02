# Progress — SPEC-TODO-LANDING-EVIDENCE-001

Card: **t359** · Worktree: `.claude/worktrees/t359` · Branch: `WT-landing-evidence`
Tier: **L** (5 artifacts) · 21 requirements · 21 acceptance criteria · **v0.2.0**

## §E.1 Plan-phase Audit-Ready Signal

**Iteration 1 verdict**: FAIL 0.80 vs the Tier L threshold 0.85 — 9 blocking (7 major, 2 minor),
4 non-blocking. All seven must-pass criteria PASSED; the shortfall was Testability (0.75) and
Traceability (0.80). Report: `.moai/reports/t359/plan-audit.md` @ `2f1c36151`.

**v0.2.0 remediation — all 9 blocking + all 4 non-blocking addressed:**

| Defect | Disposition |
|---|---|
| D1 (major) | 4 coordinates corrected to the measured `in-progress`; §A.3b added, re-grounding why those requirements bind on two measured properties instead of a lifecycle field |
| D2 (major) | Ground withdrawn as false; validation **added** — REQ-TLE-020 + AC-TLE-020 (existence + reachability, never delivery); §G states the surviving typo cost |
| D3 (major) | §C.7's false coverage claim withdrawn; split into REQ-TLE-021/AC-TLE-021 (mirror parity + column count) and an explicitly uncovered DoD item |
| D4 (major) | AC-TLE-015 now pins fields 1-5 individually |
| D5 (major) | AC-TLE-019 split into 019a/b/c with independent plants (`items`, `archived_items`, tuple drift) |
| D6 (major) | AC-TLE-014 excludes `.git/`, adds a non-empty + queue-present positive control |
| D7 (major) | AC-TLE-018 gains a reconstructed pre-change open path; §G keeps REQ-TLE-018 an argued claim |
| D8 (minor) | §E cell REQ-TLE-004 → M3; `plan.md` M1.4 no longer claims it |
| D9 (minor) | AC-TLE-005's Given gains the `spec_id` |
| D10 (non-blocking) | redundant prompt-guard conjunct dropped; inherited `todo*.go` guard cited with its controls |
| D11 (non-blocking) | §R.8 awk block re-measured and re-pasted verbatim (7 lines, not 5) |
| D12 (non-blocking) | REQ-TLE-001 reduced to the observable shape |
| D13 (non-blocking) | **closed rather than recorded** — REQ-TLE-019 now asserts `(name, type, notnull, dflt_value)` tuples |

**Counts**: 21 requirements / 21 criteria (Tier L ceiling 25/25) — 19/19 at v0.1.0, +2 from D2 and
D3. Sequential `REQ-TLE-001`..`021`, no gaps; 21 AC headings; 21 traceability rows.

**Preserved from v0.1.0** (auditor-confirmed, deliberately not touched): all ~30 `file:line` pins;
§A.3's premise falsification; AC-TLE-007's concurrency RED; AC-TLE-016's SHA-substitution step; the
19-vs-25 headroom showing no merging. Four further pins were re-measured this round and two
corrected (`todo_test.go:451-479` → `:451-480`; `prlink.go:101-114` confirmed exact).

**Still open for the Implementation Kickoff Approval gate**: the stored shape (§B.1 — one
JSON-bearing column versus four scalar columns) and the seventh-column contract change (§G,
inherited from half A). Neither is a defect; both are decisions the operator may wish to rule on.

**Not resolved** (recorded, not closed): no pre-change binary is run against a post-change database
by any criterion (§G); the `--sha` validation commands were named but not executed in this tree
(`research.md` §R.10.5); a reachable-but-wrong operator SHA remains undetectable by construction.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
