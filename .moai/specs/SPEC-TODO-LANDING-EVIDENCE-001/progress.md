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

### M2 — the record type and the attribution boundary

Verbatim command + output pairs: **`.moai/reports/t359/m2-evidence.md`**. This section carries the
verdict and its attribution; that file carries the material.

**Claim.** M2 is complete. `internal/kanban/landing_evidence.go` defines the six-fact
`LandingEvidence` record with its JSON encoding and decode, keyed so the OBSERVED ref position
(`ref_head`) and the OPERATOR-ASSERTED delivering commit (`sha`) never share a key and the delivering
key is absent rather than aliased to the head. Absence is SQL `NULL`, reached through a single seam
(`LandingEvidenceValue(nil)` → nil driver value) rather than by per-call-site discipline. The encoder
refuses a SHA without provenance, a provenance without a SHA, and any provenance other than
`operator`. `internal/kanban/prlink_landed_attribution_test.go` locks REQ-1.10 in against a
three-match fixture ref, and its §D.2 mutant was observed failing and reverted. No production file
outside M2's deliverable list was modified: `prlink.go` and `prlink_landed.go` are byte-identical to
their pre-M2 state.

| Criterion | Status | Evidence |
|---|---|---|
| AC-TLE-005 — a record carries all six facts | **PASS** (all six per-field drops observed FAILING, each reverted) | `TestLandingEvidence_CarriesAllSixFacts`; evidence file § Step 7 — six one-line encoder drops, per-drop hash + verbatim failure, all rc=1 |
| AC-TLE-006 — absence is NULL | **PARTIAL (storage half)** | `TestLandingEvidence_AbsenceIsSQLNull`; the "renders as absent" conjunct is M4's |
| AC-TLE-011 — the resolver names no commit | **PASS** (both mutant variants observed FAILING, reverted) | `TestResolver_NamesNoDeliveringCommit`; evidence file § Step 5 (`--oneline`, abbreviated assertion) and § Step 8 (`--format=%H`, full-SHA assertion) |
| AC-TLE-012 — a stored SHA is operator-supplied or absent | **NOT CLAIMED by M2** — belongs to M3 | see Gaps; `spec.md` §E maps it to M3 |
| AC-TLE-013 — ref position keyed as a ref position | **PARTIAL (key half)** | `TestLandingEvidence_RefHeadIsNotADeliveringSHA`; Given is verb-shaped, render conjunct is M4's |

**Evidence.** Six observations carry the milestone, in this order:

1. **Step 1 — the pre-edit baseline was measured on this tree, at `56af37cbb`.**
   `go test ./internal/kanban/... -count=1` → `ok … 136.974s`, rc=0. Any later red is attributable
   to M2's edits.
2. **Step 2 — the tests were written first and the package did not build.** Eleven `undefined:`
   diagnostics naming every M2 symbol. This establishes test-before-implementation; it does NOT
   establish per-field discrimination (Gaps).
3. **Step 5 — the §D.2 mutant was planted, observed FAILING, and reverted.** The resolver carrying
   its first grep match leaked `f6d756b` — the fixture's **third** commit, the integration commit
   that merely inherited the card, not the delivering change. `prlink.go` returns to
   `a2f73970a98f536b1af8f853b167cf349e6ca345`, byte-identical to its pre-plant hash.
4. **Step 6 — final scoped verification.** `gofmt -l internal/kanban/` → no output;
   `go vet ./internal/kanban/...` → rc=0; `go test ./internal/kanban/... -count=1` →
   `ok … 137.564s`, rc=0.
5. **Step 7 — AC-TLE-005's own RED, six per-field encoder drops.** Each field's struct tag replaced
   with `json:"-"` one at a time; all six red at rc=1, each reverted to
   `4709a70f7ad480ce5e9dd887bc84f40b8efc438f`. No drop survived and no drop was adjusted. The
   mechanism is not uniform — five red through the decoder's `Validate`, one through the field
   comparison — which is recorded as a finding rather than smoothed (Gaps).
6. **Step 8 — AC-TLE-011's full-SHA half, a second mutant variant.** The same carry-through
   rendered as `--format=%H` leaked the full 40-hex `accd2f7c21f4a65dbd8e131700b7de1e93696f1b` —
   again the fixture's **third** commit — and the full-SHA assertion at `:179` fired for the first
   time. Reverted; `prlink.go` back to `a2f73970a98f536b1af8f853b167cf349e6ca345`. Post-revert:
   `gofmt` clean, `go vet` rc=0, `go test ./internal/kanban/... -count=1` → `ok … 137.025s`, rc=0.

**Baseline-attribution.**

- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch `WT-landing-evidence`,
  confirmed by `git rev-parse --show-toplevel`. `/Users/goos/moai/moai-adk-go` is the same tree under
  a second spelling; the discriminant is branch + HEAD.
- **Three HEAD values, each a different thing.** Recorded separately because collapsing them would
  make a stale dispatch figure read as a measurement:
  - **`d6420c1bd` — the dispatch value, stale before M2 wrote anything.** The dispatch also stated
    the tree was clean; it was not — M2's first read observed
    `M .moai/reports/t359/m1-evidence.md` STAGED. The lead has since attributed both to its own
    dispatch (a HEAD reading taken before `t359-m1`'s queued turn landed), not to `t359-m1`. Not
    carried forward as a baseline.
  - **`56af37cbb` — the baseline M2 actually measured from.** The rc=0 pre-edit
    `go test ./internal/kanban/... -count=1` above was taken here.
  - **`b48a00285` — the tree M2 committed onto.** The released `draft → in-progress` frontmatter
    transition on `spec.md` (1 file, +2/−2, no body content), landed between M2's measurement and
    its commit. Measurements were NOT re-taken here; the two intervening commits touch only
    `m1-evidence.md` and `spec.md` frontmatter, so no Go source moved — stated rather than glossed.
- **M2 commit `7769bbf91`**, parent `b48a00285`, 5 files, +1180/−0: the three added Go files
  (`internal/kanban/landing_evidence.go`, `internal/kanban/landing_evidence_test.go`,
  `internal/kanban/prlink_landed_attribution_test.go`), this progress record, and
  `.moai/reports/t359/m2-evidence.md`.
- **Two follow-up commits**, both docs-only, no code change: `751c2ab08` (this three-value
  attribution) and the gap-closure commit carrying § Step 7 and § Step 8. The Step 7 and Step 8
  measurements were taken at `751c2ab08`, not at `56af37cbb` — a later tree, named as such. The
  files they mutate (`landing_evidence.go`, `prlink.go`) are byte-identical across both, so the
  measurements describe the same code; that is an argument from the hashes recorded in the evidence
  file, not a re-run of Step 1.

**Gaps** — what was explicitly NOT observed:

- **Five of AC-TLE-005's six field-equality assertions were never REACHED.** All six drops red
  (§ Step 7), but only the `spec_status` drop reds through the field comparison; the other five red
  through `DecodeLandingEvidence`'s `Validate` refusal, which aborts before the comparison runs. The
  criterion's requirement — an encoder emitting a subset cannot pass — is established for all six.
  The literal wording — *that field's assertion fails* — is measured for one. That the other five
  field assertions would themselves discriminate (decode without `Validate`) is a counterfactual, not
  a measurement, and was not run. Narrower than the gap it replaces, and deliberately not deleted.
- **AC-TLE-012 is NOT claimed by M2.** Its Given ("the operator records a landing without `--sha`")
  requires the `moai todo landed` verb, which does not exist until M3; `spec.md` §E maps it to M3
  independently. M2's encoder-level refusal of unpaired provenance narrows M3's reachable states but
  does not satisfy M3's criterion.
- **AC-TLE-013 is claimed only in part.** The distinct-key invariant is verified against a
  **constructed** record, not a **produced** one, and the "rendered form labels it as a ref position"
  conjunct belongs to M4. `Marker()` is supplied as M4's labelling primitive, not as the render.
- **AC-TLE-006's render half was not touched.** `moai todo pr` was never run in M2.
- **`internal/cli` was NOT run in M2.** M2 touches no file there. Its last measured state in this
  card is M1's `903bcc03c` baseline — a carry-over, named as such, not an M2 measurement.
- **No `golangci-lint`, no `-race`, no `-cover`, no non-darwin run.** CI owns the full verdict.

**Residual-risk.**

- The provenance pairing is enforced at the **encode** seam. A writer that bypasses
  `EncodeLandingEvidence` / `LandingEvidenceValue` and writes the column directly would not be
  refused. M3 is the only planned writer and routes through the seam, but nothing mechanical stops a
  future one from not doing so.
- `DecodeLandingEvidence` runs `Validate` on read, so a row written by a future non-conforming writer
  fails to decode rather than being silently accepted. That is the intended direction, but it means a
  malformed stored row becomes a **read** error on a path (`todo pr`) that REQ-2.1 requires to stay
  permissive; M4 must decide how it degrades.
- `observed_at` is validated for non-emptiness only, not for RFC 3339 shape. A malformed instant
  stores and decodes cleanly.

### M3 — the recording verb

**Claim.** `moai todo landed <id> [--sha <sha>] [--ref <ref>] | --clear` exists, is registered on
the `todo` command, and records one card's landing evidence as a single locked write of one column.
AC-TLE-004, 007, 008, 009, 010, 012 and 020 pass. Three mandated mutants were planted, observed
failing, and reverted; the plan-audit's F1 boundary premise was **OBSERVED** rather than left
asserted.

| Criterion | Status | Fired assertion under its mutant / input |
|---|---|---|
| AC-TLE-004 | PASS | `todo_landed_test.go:322`, twice (t1 queued, t2 picked) |
| AC-TLE-007 | PASS | concurrency pair; both records survive |
| AC-TLE-008 | PASS | `todo_landed_test.go:351`, ordered tuple list |
| AC-TLE-009 | PASS | replace + `SELECT landing IS NULL` = 1 after `--clear`; usage refusal is exit 2 |
| AC-TLE-010 | PASS | two-phase, frontmatter changed between phases |
| AC-TLE-012 | PASS-WITH-REPAIR | `:469`, `:472`, and — after a clause repair — `:487` |
| AC-TLE-020 | PASS | four conditions + the two-card boundary at `:619` |

**Evidence.** `.moai/reports/t359/m3-evidence.md` — verbatim command/output pairs for the baseline,
the compile-stage RED, the `-v` proof that eight tests and six subtests actually ran, each mutant's
verbatim failure with before/after hashes, the two other-SPEC guard repairs, and the final scoped
verification. Key results:

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	136.644s

$ go test ./internal/cli/... -count=1 -timeout 600s
cli rc=0

$ go vet ./internal/kanban/... ./internal/cli/...
vet rc=0
```

**Baseline-attribution.** Worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch
`WT-landing-evidence`, HEAD re-read by this agent as its first command: `705838c2a`. The pre-edit
green baseline for BOTH packages was re-established in this tree at that HEAD rather than carried
over — M1's `internal/cli` figure at `903bcc03c` was a carry-over M2 correctly named as such, and
this run replaces it with a measurement.

**AC-TLE-012's clause repair, stated plainly.** The mutant redded at `:469` and `:472` but NOT at
the containment clause AC-TLE-012 names in its own words. Cause: the leak stores git's `--oneline`
abbreviation (7 characters here) and the clause probed the full SHA and a 9-character prefix — so
against the very leak it names, the clause was **vacuous**. The probe is now 7 characters, and the
mutant was re-observed with the repair in place: `:487` then fired, naming commit **#3**, the
newest mention rather than any delivering commit. The test file's hash changed across this repair
(`f61f00c5…` → `bf9efd22…`); the production file returned to its exact pre-plant hash.

**F1 — OBSERVED.** The plan-audit's iteration-3 premise (the boundary clause fires only when the
fixture history mentions exactly one of the two card ids) was checked rather than assumed. The
fixture's non-degeneracy is asserted inside the test, and with the card token routed into
validation the clause fired at `:619` on the ACCEPTING branch: one `--sha`, t1 accepted, t2
refused. The audit ceiling being spent, this was the only remaining place the premise could be
checked, and it held.

**Gaps** (unsoftened; the full known-losses list is `m3-evidence.md` §9):

- **A scope disagreement with `plan.md` §F, resolved in the open.** §F assigns the
  `backlog_migrate.go` SELECT/INSERT to M5, but M3 cannot record without them: `writeRecord`
  re-INSERTs an explicit column list after `DELETE FROM items`, so a record written by any other
  route is erased by the next `Mutate` of any kind — `todo done` and `todo edit` included. M3
  therefore plumbs the `items` read/write and the `BacklogItem` field, and nothing else of M5's
  list. `design.md` §5 already sanctions the field; the milestone boundary is what is in dispute.
- **`landed → done → undone` silently loses the record** — MEASURED, not inferred
  (`m3-evidence.md` §9.4). `writeArchive` carries no `landing` column, which is exactly M5's
  declared site (`:196-201`). Until M5 lands, archiving discards the operator's evidence with no
  signal.
- **`golangci-lint` was not run**; only `go vet` and `gofmt`. The DoD's "project linter" line is
  unperformed for M3.
- **darwin/arm64 only.** No windows or linux build or test. `gitUnrunnable` discriminates on Go's
  own `exec` error text rather than OS text, which is why it is expected to hold cross-platform —
  an argument, not a measurement.
- **The full local suite was never run**, by policy. Every package outside `internal/kanban` and
  `internal/cli` is unmeasured at this HEAD; CI owns that verdict.
- **`gofmt -l internal/kanban/ internal/cli/` reports 28 files**, none M3-touched — a pre-existing
  condition at `705838c2a`, recorded so it is not read as an M3 regression.

**Residual-risk.**

- **The decode-on-read choice can wedge the queue.** `readRecord` now surfaces an undecodable
  `landing` value as a read error naming the card, rather than dropping it to nil. The reasoning is
  data-loss avoidance: nil would be written back as NULL by the next whole-record write, losing the
  operator's record silently. The cost is availability — a corrupt value makes the queue unreadable,
  and `--clear` cannot run because it must read first. M2 already flagged the same tension for M4's
  `todo pr` degradation path; this milestone picks loud over silent and records the trade-off.
- **The seam is still convention, not mechanism, on the archived path.** M3 routes its `items`
  write through `LandingEvidenceValue`, so the encoder's refusals hold there. `writeArchive` does
  not write the column at all yet, so nothing is bypassing the seam — but M5 must route through it
  too, and nothing mechanical requires that.
- **The `gitUnrunnable` discriminant is a string match** on three error phrasings. A future Go
  version, or a wrapper that reformats `exec` errors, could make an unrunnable check classify as a
  failed one — turning a "could not run" into a false "not reachable". The two are distinguished in
  the message and in the criterion, but not by anything stronger than substring matching.
- **The REQ-ABI-006 sweep baseline is line-keyed**, so any edit above line 216 of `todo_landed.go`
  moves the declared coordinate and reds another SPEC's guard. That brittleness is the guard's
  existing design, inherited rather than introduced, but M4 and M5 will both edit this file.
- **`--ref` accepts any string and passes it to git.** It is validated only by whether git resolves
  it. An operator can record evidence against a ref that is not the project's integration branch,
  which is intended (the flag exists for it) but means a stored `ref` must always be read before a
  stored `ref_head` is interpreted.

### M4 — the read surfaces

**Claim.** `moai todo pr` renders seven tab-separated fields with the card text LAST, the landing
evidence in field 6, and the record under a `landing` key on a render-time wrapper that leaves
`kanban.PRLinkOutcome` unwidened. AC-TLE-014, 015 and 016 pass, as do the render conjuncts of
AC-TLE-006 and AC-TLE-013. The one mandated mutant was planted, observed failing at the assertion
the criterion names, and reverted with byte-identical hashes. Two criterion/guard defects were
found and repaired; both are recorded rather than folded away.

| Criterion | Status | Fired assertion, and where |
|---|---|---|
| AC-TLE-014 | PASS | `todo_pr_landing_test.go:387` byte-identity, under a cache planted at `.moai/cache/todo-pr.cache` — outside the queue dir and outside `.git/` |
| AC-TLE-015 | PASS | 7 fields; fields 1-5 measured equal to the pre-change render of the same fixture; `landing` key present only on the evidence-carrying card |
| AC-TLE-016 | PASS | markers `(operator)` / `(ref-head)` distinguish the cells after the ref head is substituted with the asserted SHA |
| AC-TLE-006 (render conjunct) | PASS | field 6 empty for a landed-but-unrecorded card while field 2 still carries the resolver's `landed` |
| AC-TLE-013 (render conjunct) | PASS | observed record renders `(ref-head)`; its JSON carries `ref_head` and NO `sha` key |

**Evidence.** `.moai/reports/t359/m4-evidence.md` — verbatim command + output pairs, including the
mutant's failure text, the before/after SHA-1 pairs, and the emitted seven-field row.

**Baseline-attribution.** Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch
`WT-landing-evidence`, HEAD read at M4 entry as `5edba757cc0fb4b6e6ee9b8bbac53fa4e213fb21` with a
clean status. Post-repair: `go test ./internal/cli/... -count=1 -timeout 600s` exit 0 (17 packages
`ok`, `internal/cli` 412.816s); `go test ./internal/kanban/... -count=1` → `ok ... 137.054s`;
`go vet` on both trees exit 0 with no output; `gofmt -l internal/kanban/ internal/cli/` = 28 both
before and after, the same pre-existing files enumerated in the evidence file's §2.

**Two findings, repaired.**

- **AC-TLE-014's containment clause named `.moai/state/kanban/`**, which is the PRE-RENAME legacy
  directory — `internal/kanban/state_dir.go:37,42` sets `stateDirName = "todo"` and records that
  nothing writes through the legacy name. Transcribed literally the clause could never match, so
  the positive control would fail permanently. This is M3's AC-TLE-012 shape in the opposite,
  louder direction: vacuous-FAILING rather than vacuous-PASSING, which is why it surfaced on the
  first run. Repaired by DERIVING the directory from `kanban.StateDirForRoot(root)` rather than
  transcribing it. `acceptance.md` was not edited by M4 — see Gaps. **The prose half is now CLOSED
  by `4522bd439`** (manager-spec, post-run), which rewrote the criterion to name what the test
  derives — `kanban.StateDirForRoot(root)`, today `.moai/state/todo/` — instead of a transcribed
  spelling, so a later rename cannot stale it again.
  **WRONG WHEN WRITTEN, not decay — measured.** The rename landed `8910c337c` 2026-08-27;
  `acceptance.md` was first authored `b2d30deb2` 2026-09-03 with that rename already an ancestor
  (`git merge-base --is-ancestor` → true). Nothing moved after the clause was written, so this is
  an authoring error in the SPEC rather than the coordinate-decay class of M1's `:359`→`:363` and
  M3's line-keyed REQ-ABI-006 baseline. Sharper: the clause is ABSENT from the original draft and
  was added by `483cea858`, the plan-audit **iter-1 remediation** — the pass whose job is to
  strengthen criteria — naming a directory that had not existed for a week, and it then survived
  iter-2 and iter-3 unchanged.
- **`TestTodoPR_RowCarriesQueueState` (`internal/cli/todo_landing_test.go:101`) pinned SIX
  columns.** It is half A's AC-TLS-010 criterion, and AC-TLE-015 mandates breaking it. The expected
  count was bumped 6 → 7 and the text index 5 → 6 as a visible act in the same change that adds the
  column, NOT loosened to a lower bound — the guard's value is that it fails on any count change.

**Gaps.**

- **[CLOSED by `4522bd439`] `acceptance.md` AC-TLE-014 read `.moai/state/kanban/` at M4.** The test
  was correct and derived the path; the criterion prose was stale. `acceptance.md` body content is
  manager-spec's artifact, not this milestone's, so M4 reported it rather than editing it. It was
  repaired post-run by manager-spec in `4522bd439`: the clause now names what the test derives
  (`kanban.StateDirForRoot(root)`) rather than a transcribed spelling. Recorded, not erased — the
  finding that a criterion named a nonexistent directory, and that three plan-audit iterations
  passed it, is what this milestone established about how the SPEC was reviewed.
- **The malformed-evidence marker is unreachable through the store.** `DecodeLandingEvidence` runs
  `Validate` on read and `backlog_migrate.go:87-92` surfaces the failure as a read error, so no
  queue fixture reaches the render's malformed branch; it is asserted at the helper only. The
  render-side and storage-side halves of this question were decided SEPARATELY: the render marker
  by the lead's disposition for M4, the storage-side read-error behaviour by M3 and now escalated
  by the lead as an operator call. That storage-side half is **NOT settled**; until it lands,
  nothing here may claim what an operator actually sees for a corrupt row.
- **No cross-platform build, no `golangci-lint`, no coverage measurement.** M4 introduces one
  platform-sensitive construct (`filepath.Separator` in the test's queue-directory prefix),
  unverified off darwin.
- **The pre-edit `internal/cli` green was read through a filtered, `head`-bounded window** and its
  exit code was not captured, so it is a known loss rather than an attributable baseline. The same
  window shape hid the `TestTodoPR_RowCarriesQueueState` failure for one cycle; the post-repair run
  in Baseline-attribution is exit-code-based.
- **No end-to-end run of `moai todo landed` followed by `moai todo pr`.** Every fixture record was
  hand-authored, so agreement between the ref M3 resolves and the ref M4 renders is unmeasured.

**Residual-risk.**

- **The seventh column is the second contract change on this surface, and external consumers still
  cannot be enumerated.** A consumer doing `cut -f6` now reads the evidence where it read the card
  text. Fields 1-5 and the text-stays-last property are pinned; nothing can pin a consumer nobody
  can name.
- **A malformed record's `ref` reaches the cell unsanitized except for separators.** The cell strips
  tab / newline / carriage return so a corrupt record cannot split a row, but any other bytes in a
  corrupt `ref` render as-is. Unreachable today (see Gaps).
- **The marker set is disjoint by test, not by type.** `operator` / `ref-head` / `malformed` are
  three independent string constants in two packages; nothing structurally prevents a fourth from
  colliding. `TestFormatLandingEvidence_MarkersAreDisjoint` is the only guard.
- **`--json` backward compatibility rests on Go's field promotion for embedded structs.** The
  pre-change object shape is preserved because `todoPRRow` embeds `PRLinkOutcome`; replacing the
  embed with a named field would silently nest every existing key.

### M5 — compatibility and doctrine

**Claim.** The landing evidence survives the legacy-JSON round trip on BOTH card-bearing tables and
is compared by the migration's own parity verification; a reconstruction of the pre-change open
path still serves a post-change database and the frozen replica is asserted not to have drifted;
the two `todo.md` surfaces agree on the contract rows and the column count they state equals the
number of fields the render emits. AC-TLE-017, 018 and 021 pass. The mandated mutant was planted
twice, observed failing at the parity check both times, and reverted with byte-identical hashes.
The `landed → done → undone` loss M3 measured and left is closed, with the discriminating assertion
observed RED before the fix and GREEN after.

| Criterion | Status | Fired assertion, and where |
|---|---|---|
| AC-TLE-017 | PASS | the PARITY CHECK, twice — `item 0 (t1): landing … != <nil>` under the both-tables drop, `archived 0 (t2): landing … != <nil>` under the archived-only drop; both surfaced as `parity check failed, legacy file left authoritative` |
| AC-TLE-018 | PASS-WITH-FINDING | (a)+(b) red together on the `schema_version` bump; (c) reds ALONE on live-DDL drift while (a) and (b) stay green — (b)'s independent RED, observed. The third stated RED does not fire where the criterion says: see below |
| AC-TLE-021 | PASS | `todo_landed_doc_test.go:67` on mirror drift; `:77` on `states 6 columns; renders 7`, on both surfaces |
| `landed → done → undone` (M3's measured gap) | CLOSED | `backlog_landing_roundtrip_test.go:174` and `:189` red before the archived carry landed, green after |

**Evidence.** `.moai/reports/t359/m5-evidence.md` — verbatim command/output pairs with exit codes,
the two mutants' failure text with before/after hashes, the three AC-TLE-018 plants, both AC-TLE-021
plants, and the known-losses list. Key results:

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.668s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	432.249s
cli rc=0

$ go vet ./internal/kanban/... ./internal/cli/... ./internal/template/...
vet rc=0
```

**Baseline-attribution.** Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch
`WT-landing-evidence`, HEAD re-read by this agent as its first command: `e4cb86920`, clean status.
The pre-edit green for `internal/kanban` and `internal/cli` was re-established in this tree at that
HEAD (`kanban rc=0`, `cli rc=0`) rather than carried over from M4. `internal/template` has NO
pre-edit baseline in this run — it entered scope only when the mirror edit landed, which is why
§8's attribution of its two failures is a derivation rather than a before/after measurement.

**FINDING — AC-TLE-018's third stated RED does not fire at the assertion it names.** The criterion
says a `NOT NULL`-without-default `landing` makes "the verbatim INSERT in (a)" fail. Planted, the
mutant reds — at `backlog_downgrade_test.go:97`, the FIXTURE's `store.Add`, never reaching clause
(a). Cause: SQLite accepts the `ALTER … ADD COLUMN … NOT NULL` on the empty table, and the first
LIVE write then violates the constraint because a card with no record binds NULL through
`LandingEvidenceValue(nil)`. Under this mutation the post-change database cannot be built at all,
so clause (a)'s INSERT is unreachable rather than failing. Recorded as a narrower remaining gap
following M2's precedent; the test was NOT restructured to route the mutation through, because that
converts the finding into a green line. The mutation IS caught; the criterion's stated mechanism
for catching it is not what catches it.

**BLOCKER — an inherited emission drift blocks `make build`, and it is not M5's.** `make build`
exits 2 at its `agents-emit-check` pre-step on `sync-auditor.toml`; the same stale-emission defect
also reds `TestManifestHashFormat` (`CATALOG_HASH_UNSTABLE: sync-auditor`). No sync-auditor file is
modified in this tree (`git status --short` on all three: empty), and `git log` attributes the `.md`
edit to `4244c4a06` (`docs(t386/t387)`) with no following `make agents-emit`. Not fixed here:
regenerating would fold another card's un-emitted artifact into M5's commit, outside this SPEC's
scope envelope. The build BODY was run directly instead — catalog-hash regen rc=0, `go build` rc=0 —
and the regen's `sync-auditor` hash proposal was reverted by hand to its HEAD value (verified
byte-identical to `git show HEAD:internal/template/catalog.yaml`) so the commit carries only the
`moai`-skill hash M5's own mirror edit caused.

**Gaps** (unsoftened; the full known-losses list is `m5-evidence.md` §10):

- **[CLOSED by `4522bd439`] `acceptance.md:180` read `.moai/state/kanban/`** — a directory renamed
  a week before that criterion was first authored. M4 repaired the TEST (deriving the path from
  `kanban.StateDirForRoot`) and correctly left the prose alone; `acceptance.md` body content is
  manager-spec's artifact. M5 carried it forward as a DISCLOSED open SPEC defect; manager-spec then
  repaired it post-run in `4522bd439`, the criterion now naming the resolver
  (`kanban.StateDirForRoot(root)`) rather than a transcribed spelling. The record stands as an
  inheritance the sync-auditor meets already closed, not as a defect that never happened.
- **`internal/template/catalog.yaml` is not in the spec.md frontmatter module list**, yet M5
  modifies one line of it. It is the generated hash of the `moai` skill tree the mirror edit
  changed — a same-SPEC cascade, not scope expansion — but the DoD's "no source file outside the
  module list" line reads against it, so it is declared rather than left to be noticed.
- **`internal/template` was not baselined pre-edit** (above). Its two failures are attributed by
  derivation.
- **The `[HARD]` operator-act paragraph M5 adds is NOT mirror-guarded.** AC-TLE-021 compares two
  extracted ROWS only; prose outside them may drift between the surfaces undetected.
- **No `golangci-lint`, no coverage measurement, no cross-platform build** — unchanged from M3 and
  M4. The DoD's "project linter" line is unperformed for M5 as well.
- **`make build` never completed as a whole**; its `templ-generate` step was never run, so nothing
  here establishes that step is clean.
- **`gofmt -l` reports 38 files across the three trees** — `internal/kanban` 0, `internal/cli` 28
  (unchanged from M4's baseline), `internal/template` 10 (a tree M4 never measured). All
  pre-existing; none M5-touched. Recorded with the split so 28 → 38 is not read as a regression.

**Residual-risk.**

- **The parity comparison is index-keyed.** A migration that swapped two cards' evidence between
  two cards whose other compared fields also swapped would pass. Nothing pins the pairing itself.
- **The archived read now shares the live read's loud-failure trade-off.** An undecodable archived
  `landing` value surfaces as a read error naming the card, which makes the whole queue unreadable
  rather than silently dropping the record. M3 chose loud over silent for `items` and this extends
  that choice to `archived_items`; the availability cost is now paid on both tables.
- **The frozen replica is compiled from today's source.** Clause (c) keeps it honest against
  drift, but a divergence between it and a genuinely older RELEASED binary is invisible to all
  three clauses — `spec.md` §G's standing limit, narrowed here rather than closed.
- **`numberWords` is a finite vocabulary.** A `todo pr` row rewritten as "carries a dozen columns"
  fails loudly rather than silently, which is the right direction, but the doc prose and the test's
  parser are coupled by a hand-maintained map.
- **The seventh column's consumers still cannot be enumerated.** Unchanged from M4: a consumer
  doing `cut -f6` now reads the evidence where it read the card text, and nothing can pin a
  consumer nobody can name.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: 088ff0e63   # M5, the LAST IMPLEMENTATION commit — deliberately not the branch tip.
                           # Two later non-implementation commits exist on WT-landing-evidence:
                           # 4522bd439 (manager-spec, acceptance.md AC-TLE-014 prose repair, no
                           # status transition) and this bookkeeping commit. This field names the
                           # run-phase BOUNDARY; `git log --oneline` names the newest write.
                           # A mismatch between the two is expected, not staleness.
run_status: PASS-WITH-DEBT
ac_pass_count: 21
ac_fail_count: 0
ac_pass_with_debt: 3          # AC-TLE-012 (M3 clause repair), AC-TLE-014 (M4 criterion prose stale), AC-TLE-018 (third stated RED unreachable)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run  # lane does not push; integration is the lead's window
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0   # go vet rc=0; gofmt -l unchanged per tree (kanban 0, cli 28 pre-existing, template 10 pre-existing)
cross_platform_build:
  darwin_arm64: pass
  linux_amd64: not-run
  windows_amd64: not-run
total_run_phase_files: 8       # 3 new tests, 1 production file, 2 doctrine surfaces, catalog.yaml, m5-evidence.md
m1_to_mN_commit_strategy: one commit per milestone on WT-landing-evidence
blocked_gate: make build (agents-emit-check) — INHERITED from 4244c4a06, not M5; see §E.2 M5 BLOCKER
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-08
sync_commit_sha: pending-backfill   # a commit cannot cite its own hash; the real SHA lands in the
                                    # immediately-following backfill commit, which is this section's
                                    # own and is exempt from the ownership crossings.
sync_status: PASS-WITH-DEBT
b12_self_test_a: PASS   # duplicate guard — `grep -c 'SPEC-TODO-LANDING-EVIDENCE-001' CHANGELOG.md` = 0 before the append
b12_self_test_b: PASS   # AC count — 21 live AC-TLE-* identifiers measured in acceptance.md (see below)
b12_self_test_c: PASS   # every path claimed in the CHANGELOG entry verified present by `ls` (rc=0, 20/20)
changelog_entry_position: "[Unreleased] → ### Added, first entry (CHANGELOG.md:12)"
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed   # merged, riding THIS single sync commit
  plan_md: not-applicable    # this SPEC's plan.md carries no status: field
  acceptance_md: not-applicable  # this SPEC's acceptance.md carries no status: field
  progress_md: not-applicable    # this SPEC's progress.md carries no status: field
  updated_field: refreshed to 2026-09-08 in spec.md
canary_compliance_check:
  applicable: false   # this SPEC defines no forward-looking policy that its own sync would test
```

### Claim

The sync-phase deliverables landed: a `[Unreleased] → Added` CHANGELOG entry, this section, and the
merged `in-progress → implemented → completed` frontmatter transition on `spec.md`. The card's
implementation is verified GREEN on the four gates the dispatch names, at this tree and this HEAD.
Three acceptance criteria close PASS-WITH-DEBT and are named below rather than rounded up.

### Evidence

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
$ git rev-parse HEAD
0b19b527765b720b3e64414140f4f5c3bea8901c
$ git branch --show-current
WT-landing-evidence

$ grep -c 'SPEC-TODO-LANDING-EVIDENCE-001' CHANGELOG.md      # BEFORE the append
0                                                            # (grep rc=1 — no match)

$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/acceptance.md | sort -u | wc -l
      22
# 22 DISTINCT tokens, of which AC-TLS-008 is a prose cross-reference to half A
# (acceptance.md:8, "the pattern half A adopted after its own AC-TLS-008 was found
# satisfiable by a mutant") and is NOT a criterion of this SPEC. LIVE count = 21,
# which agrees with §E.3 ac_pass_count.

$ go test ./internal/kanban/... -count=1 ; echo "kanban rc=$?"
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.840s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 600s ; echo "cli rc=$?"
ok  	github.com/modu-ai/moai-adk/internal/cli/update	1.529s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	5.713s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	0.927s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	3.429s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	5.820s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	1.219s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	7.425s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.470s
cli rc=0
# The 8 lines above are the TAIL of a 17-line file. The verdict is the exit code,
# not the window: `grep -c '^FAIL\|--- FAIL'` over the UNFILTERED capture returns 0
# for both packages (`.moai/reports/t359/sync-verify/fail-scan.txt`).
# Captures, exported to a TRACKED path so they resolve at audit time:
#   .moai/reports/t359/sync-verify/{kanban,cli,vet,agents-emit-check}.txt
# The `moai spec lint` capture is 1,080,215 bytes and was NOT exported; only its
# verdict line, exit code, and SPEC-scoped grep are, in
# .moai/reports/t359/sync-verify/spec-lint-summary.txt. That truncation is a
# declared known loss — the full capture is machine-local and reaches no clone.

$ go vet ./internal/kanban/... ./internal/cli/... ; echo "vet rc=$?"
vet rc=0
# (no output)

$ moai spec lint ; echo "lint rc=$?"
0 error(s), 4157 warning(s)
lint rc=0
# `grep -c 'SPEC-TODO-LANDING-EVIDENCE-001'` over the capture = 0: no warning names
# this SPEC. The 4157 warnings are the repository-wide standing backlog.

$ ls <20 paths claimed in the CHANGELOG entry> ; echo "ls rc=$?"
ls rc=0     # all 20 present

$ make agents-emit-check ; echo "rc=$?"
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
rc=2
```

### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD
`0b19b5277` as read by THIS agent at sync entry (`git rev-parse HEAD`, working tree clean,
16 commits on `903bcc03c`). Every figure above was measured in this run against this tree; nothing
is carried over from another package, tree, or point in time. The `internal/cli` timing (~412s at
M4) is quoted as context in the dispatch and is NOT a baseline of this run.

### Gaps — what was NOT observed

**The three PASS-WITH-DEBT criteria, with why:**

- **AC-TLE-012** — M3 found its containment clause VACUOUS against the very leak the criterion
  names. The leak stores git's 7-character `--oneline` abbreviation; the clause probed the full SHA
  and a 9-character prefix, so neither could ever match. It was repaired to 7 and re-observed with
  the mutant still applied. The criterion passes on the repaired clause; it never passed on the one
  the plan phase shipped.
- **AC-TLE-014** — the criterion prose named a directory that had been renamed a week before the
  clause was written. Repaired post-run by manager-spec in `4522bd439`. Sharper, and carried
  deliberately: the clause was **ABSENT from the original draft** and was ADDED by `483cea858`, the
  plan-audit **iter-1 remediation** — the pass whose job is to strengthen criteria — then survived
  iter-2 and iter-3 unchanged. Three audit passes read a clause naming a directory that did not
  exist and none of them measured it.
- **AC-TLE-018** — its third stated RED reds at the FIXTURE (`:97`, `store.Add`), not at clause
  (a)'s verbatim `INSERT`. Under that mutation the post-change database cannot be built at all, so
  clause (a) is UNREACHABLE rather than failing. The mutation IS caught; the criterion's stated
  mechanism is not demonstrated.

**Standing gaps carried from M3 / M4 / M5, unsoftened:**

- `golangci-lint` was never run at any point in this card. Only `go vet` and `gofmt -l`. The DoD's
  "project linter" line is unperformed for M3, M4, and M5.
- No coverage measurement, absolute or delta, for either package at any milestone.
- darwin/arm64 only. No linux and no windows build or test. Cross-platform is CI's verdict.
- The full local suite was never run, by policy (`CLAUDE.local.md` §4). Every package outside
  `internal/kanban`, `internal/cli`, and `internal/template` is UNMEASURED at this HEAD.
- `internal/template/catalog.yaml` sits OUTSIDE the `module:` frontmatter list. It is a generated
  hash regenerated by the doctrine-mirror edit — a same-SPEC cascade, declared here rather than
  hidden.
- M5's `[HARD]` operator-act paragraph is NOT mirror-guarded. AC-TLE-021 compares two extracted
  rows only, so the two doctrine surfaces could drift on that paragraph without any criterion
  noticing.
- The M1-M5 known-losses sections name further unmeasured material (the discarded throwaway probes,
  the never-exercised malformed-marker path through a queue fixture, the unrun end-to-end
  `moai todo landed` → `moai todo pr` sequence against a live repository, `go test -race`,
  concurrent-open behaviour of `ensureLandingColumn`). None of it may be cited as a verdict basis.

**The inherited build blocker:**

`make build` fails at `agents-emit-check` on `.codex/agents/moai/sync-auditor.toml` (reproduced in
this run, rc=2, output above). Measured origin: `4244c4a06` (2026-09-02) edited `sync-auditor.md`
without running `make agents-emit`; the TOML's own last regeneration is the older `a7427f902`. This
card branched at `903bcc03c` (2026-09-03) and INHERITED the failure — it touched no agent
definition (`git diff --stat 903bcc03c HEAD -- .claude/agents/ internal/template/templates/.claude/agents/ internal/template/templates/.codex/agents/`
is EMPTY). `b65e7e5f6` on `origin/develop` regenerates exactly that TOML (verified: the commit's
file list contains `sync-auditor.toml`, and `git merge-base --is-ancestor b65e7e5f6 HEAD` answers
NO, so it is not yet in this branch). The claim that it clears on the integration window's standard
`git merge origin/develop` absorb is therefore WELL-FOUNDED but **NOT MEASURED** — no merge was
performed in this run, and the integration lane must re-measure `make build` on the merged tree
rather than inherit this sentence.

**The docs-site hand-off — a disclosed inheritance, not a discovery:**

`docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` enumerates the `moai todo` verbs by
name and now OMITS `landed`; the four READMEs reference `moai todo` as well. Those paths are
outside this SPEC's `module:` frontmatter, the SPEC's five artifacts mention `docs-site` zero times,
and the DoD's final line forbids modifying any source file outside the module list. Touching them
here would make the sync phase violate the SPEC's own DoD; leaving them stale is a real
user-facing gap. The disposition is a FOLLOW-UP CARD, which is the operator's to issue. The
sync-auditor meets this as an inheritance rather than as a finding.

**The open operator decision — recorded OPEN, not resolved:**

M3 chose to surface an undecodable stored `landing` value as a READ ERROR rather than dropping it to
nil, on data-loss grounds: the encoder is the column's only writer and refuses every invalid shape,
so an undecodable value is external corruption, and dropping it would lose the operator's record
permanently on the next whole-record write-back. The COST is that one corrupt value makes the queue
unreadable, and `--clear` cannot run because it must read first. M5 extended the same choice to
`archived_items`, so the availability cost is now paid on BOTH tables. This sits with the lead as an
operator call. It is OPEN.

### Residual-risk

- **The absorb may not be clean.** `b65e7e5f6` also modifies `internal/template/catalog.yaml`, and
  so does this card. The integration window should expect a conflict on that generated hash and
  resolve it by regenerating rather than by picking a side.
- **The merged tree is unmeasured.** Every green above is this branch's tip in isolation. A card
  verified alone is not a card verified merged; the integration lane re-measures at the tip it will
  merge.
- **`landed` is unexercised as a shipped command.** Every fixture in M3-M5 is in-process; `bin/moai`
  was built but never invoked. A defect that lives only in cobra wiring or in the real binary's
  git-subprocess path would not have been caught by anything in this card.
- **Three criteria pass on repaired or partially-demonstrated clauses.** AC-TLE-012's original
  clause was vacuous, AC-TLE-014's was written against a path that did not exist and survived three
  audit passes, and AC-TLE-018's stated mechanism is undemonstrated. The pattern across all three
  is the same: a criterion can read as strict and measure nothing, and the audit passes did not
  separate the two.
- **The absent linter and the absent coverage figure are the largest unmeasured axes.** Neither has
  been observed at any point in this card, so no statement about lint cleanliness or coverage on
  these packages has any basis here.
