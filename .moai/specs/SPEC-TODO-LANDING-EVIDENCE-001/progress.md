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
  transcribing it. `acceptance.md` was not edited — see Gaps.
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

- **`acceptance.md` AC-TLE-014 still reads `.moai/state/kanban/`.** The test is correct and derives
  the path; the criterion prose is stale. `acceptance.md` body content is manager-spec's artifact,
  not this milestone's, so it is reported rather than edited.
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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
