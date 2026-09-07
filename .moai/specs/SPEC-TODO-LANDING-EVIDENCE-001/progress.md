# Progress — SPEC-TODO-LANDING-EVIDENCE-001

Card: **t359** · Worktree: `.claude/worktrees/t359` · Branch: `WT-landing-evidence`
Tier: **L** (5 artifacts) · 21 requirements · 21 acceptance criteria · **v0.3.1**

## §E.1 Plan-phase Audit-Ready Signal

**Trajectory**: iter-1 FAIL 0.80 → iter-2 **PASS-WITH-DEBT 0.89** → iter-3 **PASS-WITH-DEBT 0.92**
(Tier L threshold 0.85). All seven must-pass criteria pass in every round; all thirteen iteration-1
defects and all seven iteration-2 defects RESOLVED; no stagnation, no scope reduction, no score
regression. Reports: `.moai/reports/t359/plan-audit.md` @ `2f1c36151`, `plan-audit-iter2.md` @
`8a1ae5b70`, `plan-audit-iter3.md` @ `c0cfb2520`. The audit ceiling of three iterations is spent —
there will be no confirming re-audit, which is why iteration 3's blocking items were closed as a
documentation revision (v0.3.1) rather than carried forward.

**v0.3.1 — iteration-3 closure (3 blocking, all closed here; nothing carries into run-phase):**

| Defect | Disposition |
|---|---|
| F1 (minor) | AC-TLE-020's boundary RED was unconditional in text but fires only when the fixture ref's history mentions **exactly one** of the two card ids. The Given now pins that premise, the RED hangs on it, and both degenerate cases (neither id mentioned / both mentioned) are stated |
| F2 (minor) | AC-TLE-018 clause (c)'s second conjunct (frozen switch's accepted-version set equals the live one) **withdrawn as unrealisable** — the live set is control flow at `backlog_sqlite.go:286-298`, not extractable, and its cheapest runnable reading misses the drift it named. The DDL byte-identity conjunct, which carries the independent RED, is untouched |
| F3 (minor) | `research.md` §R.10.1's causal sentence attributed non-determinism to `grep -m1`, which has no such property (`/usr/bin/grep -m1`: 6/6 argument-order runs at HEAD `c0cfb2520`). The claim is **withdrawn** and replaced with the shell-function / parallel-grep cause; the block is labelled one capture of a non-deterministic ordering, and `spec.md` §A.3b's twin block now carries the same note |

**Run-phase debt from the plan-audit: none.** F1-F3 were the only blocking items and are closed
above. The unchanged residuals below are recorded limits of the SPEC, not audit debt.

**v0.3.0 — iteration-2 debt closure (3 blocking + 4 non-blocking, all addressed):**

| Defect | Disposition |
|---|---|
| E1 (major) | AC-TLE-020's boundary clause could not fail — it varied card *text* while the predicate keys on card *id*. Replaced with one `--sha` against **two card ids**, asserted on an accepting and a refusing branch |
| E2 (major) | AC-TLE-018 (b) was a coverage assertion, not a detector; the "fail independently" claim **withdrawn**. New clause (c) asserts the frozen replica equals the live `backlogDDL` + accepted-version set, supplying (b)'s independent RED. §G gains forward drift |
| E3 (minor) | AC-TLE-020 case (d) added — the cannot-be-run branch REQ-TLE-020 binds, with the write-refuses vs read-degrades contrast made explicit |
| E4 (non-blocking) | status block re-pasted with full paths; the non-argument output ordering disclosed |
| E5 (non-blocking) | `plan.md` §B renumbered 1-5 |
| E6 (non-blocking) | AC-TLE-021 re-renders and counts its own row; `7` demoted to a note |
| E7 (non-blocking) | §G's "nothing flags it" **withdrawn**; the commit-subject mitigation recorded as considered-and-not-built, with record-time capture named as the only §D-compatible shape |

**Counts**: unchanged at 21 requirements / 21 criteria (Tier L ceiling 25/25). This delta strengthened
two criteria and added one case; it added no requirement.

**Preserved from v0.2.0** (iteration-2-confirmed, deliberately not touched): D5's three plants,
reproduced by the auditor and measured independent (`ITEMS-ONLY GUARD TRIPS=false`,
`ARCHIVED GUARD TRIPS=true`; the name-set assertion does not separate 019c while the tuple assertion
does) — so D13 is genuinely closed rather than recorded; the "record checks itself" argument for
reachability at the requirement level; and the guard-ownership decay hand-off sitting inside
AC-TLE-019 where run-phase reads it.

**Open for the Implementation Kickoff Approval gate** — three decisions, none a defect:
1. The stored shape (§B.1 — one JSON-bearing column versus four scalar columns).
2. The seventh-column contract change (§G, inherited from half A).
3. **New at v0.3.0**: whether to capture the delivering commit's subject at record time (§G, E7).
   It would add a seventh fact to the record and change REQ-TLE-016's cell format, so it is a design
   decision rather than a debt fix — surfaced here rather than folded in.

**Open optional item, held by the operator** — **F4** from iteration 3, classed optional by the
audit and deliberately not fixed at v0.3.1: AC-TLE-020's case (d) groups its two sub-conditions
under one `unrunnable` stderr classification, while REQ-TLE-020 asks the stderr to name which check
failed. The two differ — a missing `git` makes **both** checks unrunnable, whereas an unresolvable
`--ref` leaves the existence check runnable and defeats only reachability. Case (d) stays
distinguishable from (b) and (c); it is only undistinguished internally. Fixing it is one added
clause requiring (d)'s stderr to name the unrunnable check. The operator holds this; it is not
run-phase debt unless they ask for it.

**Not resolved** (recorded, not closed): no criterion runs a genuinely older binary against a
post-change database — the backward divergence between the frozen replica and a released build is
undetected by construction (§G); the `--sha` validation commands are named but were not executed in
this tree (`research.md` §R.10.5); a reachable-but-wrong operator SHA stays machine-undetectable, and
the human-noticing mitigation is recorded but not built.

**Carried into run-phase as an observation duty, not as debt**: F1's fix is a stated premise, and a
stated premise is not a built fixture. With the audit ceiling spent there is no confirming re-audit,
so the only place the premise can be checked is M3's own RED observation — plant the card-token leak
and watch the boundary clause actually red. Until that is observed, treat AC-TLE-020's boundary
clause as asserted rather than demonstrated. The same discipline generalises from the audit's
closing note: across three rounds the recurring defect was a clause that could not fail, so every
newly written assertion should be accompanied by the mutation that reds it before it is accepted.

## §E.2 Run-phase Evidence

### M1 — the stored shape

Verbatim command + output pairs: **`.moai/reports/t359/m1-evidence.md`**. This section carries the
verdict and its attribution; that file carries the material.

**Claim.** M1 is complete. The schema-freeze guard now pins exact ordered
`(name, type, notnull, dflt_value)` tuple sequences on `items` AND `archived_items`, asserted per
table; `landing TEXT` (nullable, no default) is added to both by an idempotent `ALTER` at engine
open, placed after `backlogDDL` and before the `schema_version` switch, with idempotence decided by
reading `pragma_table_info` rather than by catching SQLite's `duplicate column name`. The DDL const
is deliberately unedited. `schema_version` stays `"1"`; the `items.state` CHECK is byte-identical.

| Criterion | Status | Evidence |
|---|---|---|
| AC-TLE-001 — `items` column shape | **PASS** | `TestBacklogLanding_ItemsColumnShape` |
| AC-TLE-002 — `archived_items` column shape, asserted independently | **PASS** | `TestBacklogLanding_ArchivedItemsColumnShape` |
| AC-TLE-003 — migration idempotent, rewrites nothing | **PASS** | `TestBacklogLanding_MigrationIsIdempotent`; its stated RED (unconditional `ALTER`) also observed |
| AC-TLE-019a — `items` plant | **PASS** (guard FAILED under plant, reverted) | evidence file § Step 5 |
| AC-TLE-019b — `archived_items` plant, alone | **PASS** (guard FAILED under plant, reverted) | evidence file § Step 5 |
| AC-TLE-019c — tuple plant (`TEXT NOT NULL DEFAULT ''`) | **PASS** (guard FAILED under plant, reverted) | evidence file § Step 5 |
| AC-TLE-004 | **NOT CLAIMED by M1** — see Gaps | — |

**Evidence.** Four observations carry the milestone, in this order:

1. **Step 0 — the decayed baseline was re-measured, not cited.** AC-TLE-019's decay note warns the
   guard is owned by `SPEC-TODO-ARCHIVE-QUERY-001`. Both plants left the CURRENT (unextended) guard
   GREEN in this tree at `903bcc03c`, so the column-blindness premise holds and the SPEC's recorded
   RED was not relied on.
2. **Step 1 — the extended guard was written against the PRE-change tuples and observed GREEN**
   before `landing` existed (`plan.md` §B.1 ordering).
3. **Step 2 — the same guard FAILED on both tables** once the `ALTER` landed and before the
   expectations were updated. This is the demonstration that the assertion is not merely written to
   match what was built.
4. **Step 5 — each §D.2 plant was applied alone, observed FAILING, and reverted**, with the
   before/after SHA-1 of `backlog_sqlite.go` recorded per plant (all three return to
   `01872b7bc1050b33aa1ee9eb3c326c726d4207c6`). No mutant survived.

**Baseline-attribution.**

- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch `WT-landing-evidence`,
  confirmed by `git rev-parse --show-toplevel`. (`/Users/goos/moai/moai-adk-go` is the SAME tree
  under a second spelling — same `dev:inode`, and git resolves it to the uppercase path — so the
  real hazard is worktree → primary drift. The discriminant is that the primary sits at a different
  HEAD and branch.)
- **Pre-change baseline HEAD `903bcc03c`**, measured by the lane before any edit:
  `go test ./internal/kanban/... -count=1` → rc=0; `go test ./internal/cli/... -count=1 -timeout 600s`
  → rc=0. Any red after the M1 edits is therefore attributable to them.
- **Post-M1 HEAD `3bcb0c33a`** (3 files, +374/−3): `gofmt -l internal/kanban/` → no output;
  `go vet ./internal/kanban/...` → rc=0; `go test ./internal/kanban/... -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/kanban 137.728s`.
- **Follow-up HEAD `2dacb1d83`** (SQL-finding response, no behavior change):
  `go vet ./internal/kanban/...` → rc=0; `go test ./internal/kanban/... -count=1` →
  `ok ... 137.483s`.

**Gaps** — what was explicitly NOT observed:

- **AC-TLE-004 is NOT claimed by M1.** Its Given-When requires `moai todo landed`, which does not
  exist until M3. It was not attempted and nothing here asserts it (`plan.md` §F M1.4).
- **`internal/cli` was NOT re-run after the M1 edits.** It was green at baseline `903bcc03c`; only
  `internal/kanban` was re-measured post-change, per the lane-local verification scope. CI owns the
  full-suite verdict.
- **Step 0 establishes column-blindness only at `903bcc03c`, in THIS tree.** It says nothing about
  the guard's state in any other tree, on develop, or at any other commit.
- **The CWD attribution for the pre-correction batches is a discriminant argument, not a direct
  per-call observation.** No `pwd` was taken before each early batch; the argument is that the two
  trees differ in HEAD and branch (`WT-landing-evidence` vs `main`), and every bare `git` call in
  those batches returned the worktree's values. A batch that produced no git output carries no
  independent proof of where it ran. `pwd` was measured directly only after the tree-identity
  correction landed.
- **No non-darwin build or test run.** darwin/arm64 only; `ALTER TABLE` and `pragma_table_info` were
  not exercised on another GOOS.
- **`golangci-lint` was not run** — only `gofmt -l` and `go vet`.
- **No JSON⇄SQLite round trip or parity comparison was exercised** (M5 / AC-TLE-017 scope).
- **Per-plant hashes cover `backlog_sqlite.go` only**; no plant modified a test file, and the
  tree-level `git status --short` is the substitute evidence.

**Residual-risk** — what could still be wrong despite the above:

- **The column is live in the schema but carried by no read or write path.** `backlog_migrate.go`
  uses explicit column lists that do not mention `landing`, so once M2/M3 begin writing values, a
  JSON export → SQLite import round trip would drop them silently until M5 closes it.
- **Concurrent first-open is unproven.** Two processes opening a pre-change database at once could
  both see the column absent and both issue the `ALTER`; the loser fails its open.
  `SetMaxOpenConns(1)` and `_txlock=immediate` do not serialize across processes. Not exercised, not
  covered by any criterion.
- **The guard is now a maintenance coupling.** Any future column on either table breaks
  `TestTodoHistoryAddsNoSchemaChange` by design; a hurried reader may update the expected string
  rather than ask why it fired.
- **Steps 2 and 3 share one commit** (`plan.md` §F M1.3), so the intermediate FAILURE exists in no
  commit — it is a transcript observation, captured live and in order. The evidence file is its only
  durable carrier.
- **Two `pragma_table_info` queries are added to every engine open.** Open latency was not measured
  before or after.
- **The SQL finding was real but the site was not dangerous — two claims that must not collapse into
  one.** A PostToolUse check reported string-concatenated SQL. It was NOT a false positive: a real
  concatenation existed (`` `SELECT count(*) FROM ` + table `` in `rowCount`) and was removed in
  `2dacb1d83`. It was also NOT dangerous: the helper was test-local, its `table` parameter had one
  caller passing the literal `"items"`, and no runtime value could reach it. The removal's
  justification is simplification — the parameter earned nothing — not security. Recording only the
  first claim would make a bookkeeping error about where the reader had looked read as a
  narrowly-averted incident.
- **A lane worker overrode an orchestrator instruction on its own reading of the evidence.** The lead
  instructed that the finding was a false positive and the code be left as it was; the lane had
  already changed it, judged on re-reading that the finding pointed at a site the lead had not
  inspected, and reported the divergence rather than conforming. The lead retracted the
  false-positive call on review and accepted the deviation. Recorded because a good outcome is
  exactly the circumstance in which the fact of a deviation gets smoothed away — and the next
  deviation, with a worse outcome, would then have no baseline to be judged against. Full record:
  `.moai/reports/t359/m1-evidence.md` § Disposition.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
