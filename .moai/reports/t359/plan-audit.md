# SPEC Review Report: SPEC-TODO-LANDING-EVIDENCE-001

Card: **t359** · Iteration: **1/3** · Tier **L** (PASS threshold **0.85**)
Tree: worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD `b2d30deb2`
Auditor: plan-auditor (Claude anchor only; `audit_model` multi-backend fan-out not requested)

**Verdict: FAIL**
**Overall Score: 0.80** (Tier L threshold 0.85)
**Blocking defects: 9** (7 major, 2 minor) · non-blocking: 4

Reasoning context ignored per M1 Context Isolation. The dispatch's hunt list was used to
choose where to spend measurement effort; every finding below is grounded in a command run in
this tree, not in the dispatch's framing. Where the dispatch supplied a premise (the
`REQ-TODO-013` reading, the `SPEC-TODO-ANALYSIS-001` agreement, the AC count, the clean tree),
it was re-measured rather than accepted.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-TLE-[0-9]*' spec.md | sort -u` yields
  exactly `REQ-TLE-001` … `REQ-TLE-019` — 19 ids, sequential, no gap, no duplicate, uniform
  3-digit zero-padding. `grep -c '^- \*\*REQ-TLE-'` = 19, so every id is also a definition.
- **[PASS] MP-2 GEARS format compliance — judged against the REQUIREMENT layer (`spec.md` §C),
  not the verification layer.** All 19 `REQ-TLE-xxx` entries match a GEARS pattern:
  Ubiquitous (001, 002, 004, 005, 006, 011, 012, 013, 014, 016, 017, 018, 019), Event-driven
  (003, 007, 009, 015), Where+When compound (010), Unwanted-as-`shall not` (008, and the second
  clause of 004/012). `grep -Ein 'should|may |reasonable' ` over the 19 REQ lines → `rc=1`.
  `IF/THEN` scan → `rc=1`. The 19 Given-When-Then entries in `acceptance.md` are `AC-TLE-xxx`
  verification-layer entries and are graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present with correct types
  (`spec.md:2-14`): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created`/`updated: 2026-09-03` (ISO), `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). `grep -nE
  '^(created_at|updated_at|labels|spec_id):'` → `rc=1`; no rejected snake_case alias.
  `tier: L` present as the optional 13th.
- **[N/A] MP-4 Section 22 language neutrality.** Single-language SPEC — the entire surface is
  Go (`internal/kanban`, `internal/cli`) plus two markdown doctrine files. No multi-language
  tooling claim is made. Auto-passes per the MP-4 N/A rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Six referenced SPEC ids extracted; all six
  directories exist. Measured statuses: `SPEC-TODO-LANDING-STATE-001` `completed`,
  `SPEC-TODO-ANALYSIS-001` `completed`, `SPEC-TODO-ARCHIVE-QUERY-001` `completed`,
  `SPEC-TODO-DESTRUCTIVE-GUARD-001` `completed`, `SPEC-KANBAN-QUEUE-PR-SYNC-001` `in-progress`,
  `SPEC-KANBAN-TODO-CLI-001` `in-progress`. **None is `retired`, `superseded`, or `archived`**,
  so no D7 BLOCKING finding is emitted. (The two `in-progress` statuses ARE misdescribed by the
  SPEC as `completed` — that is defect D1 below, a §1 unobserved-claim defect, not a D7 trigger.)
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` → `0`. D8 is
  auto-PASS per D8-4; no build-tag or exemption clause is required.
- **[PASS] MP-7 clarification gate.** `grep -rn '\[NEEDS CLARIFICATION' ` over all five
  artifacts → `rc=1`, no match. No clarification marker is open.

No must-pass criterion fails. The FAIL verdict is score-driven.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 (above-band: the requirement layer itself is single-interpretation) | 19 REQs each carry one unambiguous reading and no weasel term; deductions are in the supporting prose — D1 (`completed` × 3 for two `in-progress` SPECs), D3 (`spec.md:403-409` §C.7 names two ACs that verify something else), D8 (`spec.md` §E maps REQ-TLE-004 → M1 while its AC needs the M3 verb), D9 (`acceptance.md:61-70` Given omits the `spec_id` its Then asserts on) |
| Completeness | 0.85 | 0.75 (above-band: five Tier L artifacts, six `### Out of Scope —` H3s with bullets, an explicit §G) | HISTORY / §A / §B / §C / §D / §E / §F / §G all present; six out-of-scope H3s at `spec.md:417,426,434,441,450,458`, each with `-` bullets; §G separates *not measured* from *residual risk*. Deductions: D3 (a coverage claim with no coverage behind it) and D7 (a §G gap deferred to a criterion that does not close it) |
| Testability | 0.75 | 0.75 | 19 binary Given-When-Then criteria, each with an explicit RED, a five-entry planted-mutant register (`acceptance.md` §D.2) and a §D.1 severity table; zero weasel words (`grep -Ein 'appropriate\|adequate\|reasonable\|proper'` → `rc=1`). Four criteria are nonetheless weaker than their requirements: D4 (AC-TLE-015), D5 (AC-TLE-019), D6 (AC-TLE-014 is not reliably binary), D7 (AC-TLE-018) |
| Traceability | 0.80 | 0.75 (above-band: a genuine bidirectional 19↔19 table) | `spec.md` §E maps every REQ to exactly one AC and every AC to exactly one REQ; counts measured 19 / 19 / 19 (REQ ids, AC headings, AC ids). Deductions: D8 (one wrong milestone cell), D5 (REQ-TLE-019's `archived_items` half traces to an AC that never exercises it), D3 (§C.7 asserts a trace that does not exist) |

Aggregate (harmonic mean of the four): **0.7984 → 0.80**. Arithmetic mean is also 0.80.
Tier L threshold is **0.85**. **0.80 < 0.85 → FAIL.**

---

## Hunt 1 — the D2/D3 resolution: does `sha_source: operator` earn its keep?

**Verdict: the escape is genuine, but one supporting argument is false, and that argument is the
one doing the work the evidence does not support.**

The pins are accurate. `SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251-255` reads verbatim as quoted,
including the grounds clause; `:259-261` carries REQ-2.1 `[HARD]` verbatim.
`.claude/skills/moai/workflows/todo.md:59-63` carries the operator-act `[HARD]` verbatim,
including "The queue records the operator's intent; it does not curate it."

**What survives adversarial pressure.** REQ-1.10's stated grounds are specifically about
*machine inference* — "a card's first matching commit may be another card's report commit that
merely mentions it". An operator-typed `--sha` involves no inference by the machine, and the
queue already stores unvalidated operator assertions of exactly this class: `spec_id` is
operator-supplied, unchecked against `.moai/specs/`, and persisted. So the SPEC is not inventing
a new category to slip a SHA through; it is placing the SHA in a category the landed doctrine
already defines and the store already implements. REQ-TLE-011 restates the resolver prohibition,
REQ-TLE-012 restates it at the storage boundary, and both carry planted-mutant REDs
(`acceptance.md` §D.2 rows 3 and 4) rather than doc-grep assertions. The asymmetry the
resolution rests on — the machine may not claim what it cannot know, the operator may — is
stated, not assumed.

**Where it does not survive.** The SPEC's own §B.3 argument against a resolved SHA is that
"a stored mis-attribution is *worse* than a rendered one, since it outlives the invocation that
made it". That "worse" property belongs to **storage**, not to authorship: an operator typo
produces precisely the durable wrong attribution the SPEC just called worse. §G owns this in one
paragraph — and then declines the only cheap remedy on a ground that is **factually wrong**:

> nothing validates that a supplied SHA exists on the named ref, and REQ-TLE-012 deliberately
> does not add such a check, because validating it against the grep predicate would reintroduce
> exactly the inference REQ-1.10 forbids.

Validating that an object exists (`git cat-file -e <sha>^{commit}`) or that it is an ancestor of
the named ref (`git merge-base --is-ancestor <sha> <ref>`) is **not** the card-token grep
predicate and makes **no** attribution claim. It answers "does this commit exist on this ref",
never "did this commit deliver this card". The SPEC conflates a referential-integrity check with
the forbidden inference and refuses both together. The result is that `sha_source: operator` is
the *only* thing standing between the store and an arbitrary string — the marker records who
typed it, and nothing at all records whether it names a commit. That is more weight than the
provenance argument was built to carry, and it is defect **D2**.

The fix is small and does not touch REQ-1.10: either add a requirement that a supplied `--sha`
must resolve to a commit reachable from the named ref (rejecting with exit 1 otherwise, and
saying explicitly that this asserts existence and never delivery), or — if the check is
genuinely unwanted — replace the false ground in §G with the true one.

---

## Hunt 2 — AC-TLE-019's "already red": reproduced, and extended

**Verdict: the author's measurement is correct and I reproduced it independently. The proposed
exact-column-set assertion closes the `items` half. The `archived_items` half is stated in the
requirement and left untrippable by the criterion.**

I did not read the evidence file as evidence. I wrote my own probe
(`internal/kanban/zz_audit_t359_probe_test.go`, since deleted), planted a bogus column on
**both** tables, and re-ran the guard's four assertions verbatim:

```
$ go test ./internal/kanban/ -run 'TestTodoHistoryAddsNoSchemaChange|TestAuditT359ColumnBlind' -count=1 -v
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
=== RUN   TestAuditT359ColumnBlind
    A1 table set     = "archived_findings archived_items findings items meta" -> match=true
    A2 CHECK present = true
    A2 stored items SQL = CREATE TABLE items (
      seq      INTEGER PRIMARY KEY,
      ...
      state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
    , bogus TEXT)
    A3 index set     = "idx_items_state" -> match=true
    A4 schema_version = "1" -> match=true
    COLUMN SET items          = "seq id text added_at spec_id state bogus"
    COLUMN SET archived_items = "seq id text added_at spec_id state position bogus2"
--- PASS: TestAuditT359ColumnBlind (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.554s
```

Confirmed independently: all four assertions stay GREEN with a planted column on either table,
and the stated mechanism is visible in the A2 output — SQLite appends `, bogus TEXT)` **after**
the closing `CHECK (...)`, so the guard's `strings.Contains(itemsSQL, wantCheck)` at
`backlog_schema_freeze_test.go:62-65` still matches. Reading the guard confirms the structural
claim: its four assertions are a table-name set (`:50-53`), a `strings.Contains` on the CHECK
(`:62-65`), an index-name set (`:84-86`), and `schema_version` (`:94-96`). **None reads
`pragma_table_info`.** The SPEC's §A.5 is accurate and its §A.6 column tables match
`backlog_sqlite.go:107-114` and `:124-132` exactly.

**Does REQ-TLE-019's assertion close it?** For `items`, yes: an exact column-name sequence is
tripped by any addition, removal, or reordering, and AC-TLE-001 additionally pins `landing`'s
type / `notnull` / `dflt_value`. For `archived_items`, the *requirement* says so
(REQ-TLE-019: "the exact column set of `items` **and of `archived_items`**") but **AC-TLE-019's
Given-When plants only on `items`** ("an extra column is planted (`ALTER TABLE items ADD COLUMN
bogus TEXT`)"). A run-phase implementation that extends the guard for `items` alone satisfies
AC-TLE-019 in full — plant on `items` → fails; no plant → passes — while leaving exactly half
of REQ-TLE-019 unbuilt and undetected. This is the same half-applied-migration hazard the author
deliberately guarded against by splitting AC-TLE-001 from AC-TLE-002 ("The two criteria are
deliberately separate so a half-applied migration is caught"), and then did not guard against
here. Defect **D5**.

A residual the SPEC does not record: a column-**name**-set assertion catches additions and
removals, not a type or `NOT NULL` change to an existing column. AC-TLE-001 pins the shape of
`landing` only. That is defect **D13**, non-blocking.

---

## Findings on the remaining hunts

**Hunt 3 — mutant register and the "cannot fail" class.** The five registered mutants each
genuinely trip their criterion. The one I expected to be vacuous is not: AC-TLE-007's RED
("writing outside the lock → the concurrency assertion loses one of the two records") holds
because `BacklogStore.Mutate` (`internal/kanban/backlog_store.go:638-665`) is a whole-record
read-modify-write — `readRecord` → callback → `normalizeBacklogRecord` → `writeRecord` — so two
unlocked concurrent mutations on *different* cards do lose one, exactly as claimed. AC-TLE-011
and AC-TLE-012 are non-vacuous by construction (three known SHAs, containment over the rendered
outcome, plus a positive control that the answer is one of the three landing values).
AC-TLE-016's SHA-substitution step is the strongest criterion in the set: it is specifically
built so the assertion cannot pass on the SHA values alone. Of the fourteen non-mutant criteria
I found four holes — D4, D6, D7, D9 — plus one redundancy, D10. None of the fourteen is
satisfied-before-implementation in the way the preceding card's criterion was.

**Hunt 4 — who writes the evidence.** The three-candidate table (`spec.md` §B.2) is sound and
its rejection grounds check out: REQ-2.1 is `[HARD]` and binds the *surface*
(`SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:259-261`), and `internal/cli/todo_pr.go:1-15` states the
writes-nothing property of the file itself, so `--record` really would be a write path on a
surface forbidden to have one. Nothing machine-driven can reach the new verb by construction any
more than it can reach `add` or `drop` — the boundary is doctrinal, inherited unchanged, and §D
says so ("Any change to **who** may issue a queue-mutating verb … stands unchanged"). But the
doctrine's only landing point in this SPEC is two prose files, and §C.7's claim that AC-TLE-015
and AC-TLE-016 verify what those files state is **false** (D3): AC-TLE-015 asserts a seven-field
split and a JSON key; AC-TLE-016 asserts a marker survives SHA substitution. Neither reads
`todo.md`. The doctrine text is a `§D.3 Definition of Done` item and nothing more.

**Hunt 5 — contract-change accumulation.** The inherited residual is stated honestly: `todo.md`
today reads "six tab-separated columns — card id, outcome, pull requests, confidence, queue
state, card text", confirming the 5→6 predecessor change, and §G names this as the second and
records that consumers still cannot be enumerated. What the SPEC does *not* do is protect the
positions it is disturbing: AC-TLE-015 pins field 7 (text) and field 6 (evidence) and the field
count, and leaves fields 1-5 entirely unasserted. Given that the SPEC's own stated hazard is a
consumer keying on a column position, a criterion that leaves five of seven positions free is
under-specified for its own risk (D4). A restatement is not enough on the second change; one
extra assertion is.

**Hunt 6 — tier and budget.** 19 REQs and 19 ACs against the Tier L ceilings of 25/25
(`spec-workflow.md:145-151`) — 6 of headroom, and the headroom means there was no pressure to
merge. I checked for merged requirements anyway: REQ-TLE-005 (six facts), REQ-TLE-009
(replace + clear), REQ-TLE-015 (column + JSON key) and REQ-TLE-003 (idempotent + no rebuild +
no rewrite) are each multi-clause, but each is covered clause-by-clause by its criterion, so
none is a budget-driven merge. No padding in the other direction either. The Tier L claim
survives: `spec-workflow.md:142` allows "> 1000 LOC **or constitutional**", and a schema change
to the persisted store shape reaches every operator queue in the field — the file count
(~14-17 across `plan.md` §F deliverables plus `backlog_store.go`, which `design.md` §5 requires)
is borderline against the "> 15 files" disjunct, but the constitutional disjunct is met
independently and the 5-artifact set is complete.

**Hunt 7 — the declared gaps.** Three of five are genuinely deferred; one is a parking spot; one
is cosmetic.
- *Migration parity read-not-executed* — genuinely deferred; AC-TLE-017 requires the parity check
  itself to **FAIL** under the drop mutant, which is the right shape and does close it.
- *`ALTER TABLE ADD COLUMN` observed on this machine only* — genuinely deferred; CI owns the
  matrix and the SPEC says so.
- *External `todo pr` consumers un-enumerable* — genuinely inherited and correctly not claimed
  closed (though see D4).
- *The downgrade claim* — **a parking spot (D7).** §G says the pre-change-binary execution "is
  AC-TLE-018's job in run-phase". AC-TLE-018 does not do that job: it executes the pre-change
  *statements* verbatim against a post-change database and reads `schema_version`. That is a good
  test of the statement surface and it is not a pre-change binary; the open path
  (`backlog_sqlite.go:278-299`) — DDL exec, `schemaVersion`, the version switch — is never
  exercised by an old build. The gap is real, but the SPEC has told itself it is closed.
- *Four drifted coordinates re-measured* — verified. I re-opened every `file:line` in `spec.md`,
  `design.md`, `plan.md`, and `research.md` at its address in this tree, including all four named
  in `progress.md` (`backlog_store.go:65-71` / `:187-193`, `backlog_sqlite.go:124-132`,
  `backlog_migrate.go:604-608` / `:585-590`). **Every pin resolves to the cited content.** I also
  reproduced §R.6 (four statements, all naming their columns), §R.6's `SELECT *` scan (`rc=1`),
  §R.7's `ALTER TABLE` scan (`rc=1` after my probe was deleted), §R.5, and §R.9. One cosmetic
  lapse: §R.8's block labelled `awk NR>=292 && NR<=298` shows five lines, not seven — the leading
  `}` (292) and trailing `}` (298) are absent, so the "verbatim output" is not verbatim (D11).

**Where the SPEC is right, briefly.** The §A.3 premise falsification is correct and I verified
both halves: `SPEC-KANBAN-TODO-CLI-001/spec.md:59` permits additive change and
`SPEC-TODO-ANALYSIS-001/spec.md:51` says the same thing, not the opposite — the two do not
disagree and no completed SPEC needed editing. The five-vs-six-vs-seven column reasoning, the
one-column-holding-a-record versus four-scalar-columns argument, the refusal to bump
`schema_version` (grounded on the live rejection at `backlog_sqlite.go:293-297`), the refusal of
a fourth `state` value (grounded on the in-tree comment at `:95-97`), and the ALTER-placement
argument in `design.md` §2 (after the DDL, before the version switch — which matches the actual
control flow at `:279-298`) are all sound and correctly grounded. The plan's reversibility
ordering is real, not decorative: M1's decision genuinely invalidates M2-M5 if overturned.

---

## Defects Found

**D1** — `.moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/spec.md:L121` (§A.4), `:L242` (§B.3), `:L224` (§B.2),
`:L81`/`:L110-111` (§A.3) — Two referenced SPECs are asserted to be `completed`; both measurably read
`in-progress`. §A.4: "`SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-2.1 … is [HARD] and is *completed*";
§B.3: "REQ-1.10 … is completed"; §A.3 frames `SPEC-KANBAN-TODO-CLI-001` and
`SPEC-TODO-ANALYSIS-001` as "two completed SPECs". Measured:
`grep -m1 '^status:' .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md` → `status: in-progress`;
same for `SPEC-KANBAN-TODO-CLI-001`. The requirements bind either way, so the design is not
invalidated — but the SPEC opens by claiming every figure in it was measured in this tree, and
this one was not. — Severity: **major** — Class: **blocking** — Required fix: replace
"completed" with the measured status at each of the three sites, or state the property actually
relied on (the requirement is landed and binding) rather than a status that was not read.

**D2** — `spec.md:L543-547` (§G "Residual risk", 3rd bullet) — The stated ground for declining any
validation of an operator-supplied `--sha` is false. "validating it against the grep predicate
would reintroduce exactly the inference REQ-1.10 forbids" conflates a referential-integrity
check with an attribution inference: `git cat-file -e <sha>^{commit}` and
`git merge-base --is-ancestor <sha> <ref>` make no claim about which card a commit delivered.
The consequence is that the store accepts an arbitrary string as a delivering SHA, while §B.3's
own argument ("a stored mis-attribution is worse than a rendered one") applies to it in full.
— Severity: **major** — Class: **blocking** — Required fix: add a requirement that `--sha` must
resolve to a commit reachable from the named ref, rejecting with exit 1 otherwise and stating
explicitly that this asserts existence and never delivery (with a matching AC); OR keep the
refusal and replace the false ground in §G with the true one.

**D3** — `spec.md:L403-409` (§C.7) — "`acceptance.md` AC-TLE-015 and AC-TLE-016 verify what those
files must state" is false. AC-TLE-015 asserts a seven-field tab split, `field 7 == card text`,
`field 6 == evidence`, and a `--json` key; AC-TLE-016 asserts a marker survives SHA substitution.
Neither reads `.claude/skills/moai/workflows/todo.md` or its template mirror. The doctrine text
— including the evidence-is-not-a-transition sentence, the single most load-bearing user-facing
claim in the SPEC — is covered by `acceptance.md` §D.3 Definition of Done only, which is not a
criterion. — Severity: **major** — Class: **blocking** — Required fix: correct §C.7 to state
that the doctrine prose is a DoD item with no criterion behind it (the "a grep for a sentence is
a bad criterion" reasoning is sound and should stay), or add a criterion that asserts the two
files agree with each other and with the rendered column count.

**D4** — `acceptance.md:L176-186` (AC-TLE-015) — The criterion pins the field count, field 7
(card text) and field 6 (evidence), and asserts nothing about fields 1-5. The SPEC's own §G
residual risk is that a consumer keys on a column position ("A consumer doing `cut -f6` now gets
the evidence where it previously got the queue state") on a surface undergoing its **second**
contract change. A run-phase implementation that reorders or relabels id / outcome / pull
requests / confidence / queue state while inserting evidence at 6 passes this criterion unchanged.
— Severity: **major** — Class: **blocking** — Required fix: extend AC-TLE-015's Then to assert
fields 1-5 equal their pre-change values for the same fixture (id, outcome, pull requests,
confidence, queue state), so the criterion pins the whole contract it is changing.

**D5** — `acceptance.md:L226-240` (AC-TLE-019) vs `spec.md` REQ-TLE-019 — The requirement binds
the exact column set "of `items` **and of `archived_items`**"; the criterion plants only
`ALTER TABLE items ADD COLUMN bogus TEXT`. A guard extended for `items` alone satisfies
AC-TLE-019 completely while leaving the `archived_items` half unbuilt — measured blind in this
tree (`COLUMN SET archived_items = "seq id text added_at spec_id state position bogus2"`, all
four guard assertions GREEN). This is the identical half-applied hazard the author deliberately
split AC-TLE-001 from AC-TLE-002 to catch. — Severity: **major** — Class: **blocking** —
Required fix: add the second plant to AC-TLE-019 (`ALTER TABLE archived_items ADD COLUMN bogus2
TEXT` → guard FAILS), asserted independently of the `items` plant; or split it into
AC-TLE-019a/b mirroring the AC-TLE-001/002 split.

**D6** — `acceptance.md:L163-174` (AC-TLE-014) — "hashes **every file under the project root**"
is not reliably binary. `moai todo pr` shells out to `git` in the project working directory
(`internal/kanban/prlink_landed.go:154` via the `todoRunCommand` seam at
`internal/cli/todo_pr.go:57-65`), and the Given requires a git repo for the landed question to be
askable at all. Git may write inside `.git/` during a read — commit-graph writes, `gc.log`,
opportunistic ref packing — so the two hashes can differ for reasons that have nothing to do with
the property being asserted. Widening the scope was the right correction to half A's D1; leaving
`.git/` in it is a new flake surface. — Severity: **major** — Class: **blocking** — Required fix:
state the exclusion set in the criterion — hash every file under the project root **except**
`.git/` — and add a positive control that the hash set is non-empty and includes the queue
directory, so the exclusion cannot hollow the assertion out.

**D7** — `spec.md:L522-526` (§G "Not measured", 2nd bullet) vs `acceptance.md:L211-224`
(AC-TLE-018) — §G defers the downgrade demonstration with "That execution is AC-TLE-018's job in
run-phase". AC-TLE-018 executes the pre-change **statements** against a post-change database; it
never runs a pre-change **binary**, so the pre-change open path (`backlog_sqlite.go:278-299`:
`ExecContext(backlogDDL)` → `schemaVersion` → the version switch) is not exercised by any old
build. The gap is real and correctly identified; it is then handed to a criterion that does not
close it, which converts a declared gap into an assumed closure. — Severity: **major** —
Class: **blocking** — Required fix: either state in §G that REQ-TLE-018 remains an argued claim
after AC-TLE-018 passes, naming what AC-TLE-018 does and does not demonstrate; or strengthen
AC-TLE-018 to build the pre-change engine open path in-test (the DDL const plus the version
switch as they stand at HEAD) and run it against the post-change database.

**D8** — `spec.md` §E traceability table, row `REQ-TLE-004 | AC-TLE-004 | M1` — AC-TLE-004's
Given-When requires `moai todo landed` to run ("**When** `moai todo landed` records evidence for
each"), and that verb does not exist until M3 (`plan.md` §F M3). `plan.md` M1.4 acknowledges this
by writing "AC-TLE-004 **in part**", and M3.2 claims it again ("state untouched (AC-TLE-004)").
The §E table states M1 flatly, so the milestone map says a criterion completes one milestone
before its subject exists. — Severity: **minor** — Class: **blocking** — Required fix: change the
§E milestone cell for REQ-TLE-004 to M3, or split the criterion into the M1 half (the CHECK is
untouched by the ALTER) and the M3 half (no write moves a card).

**D9** — `acceptance.md:L61-70` (AC-TLE-005) — The Given supplies "a card, a resolvable ref whose
head SHA is known to the test, and an operator-supplied `--sha`"; the Then asserts all six fields
including `spec_status`. Per `design.md` §1, `spec_status` is "present when the card carries a
`spec_id`" — which the Given does not provide. As written the criterion either cannot assert its
sixth field or requires the implementer to infer a fixture the criterion does not name.
— Severity: **minor** — Class: **blocking** — Required fix: add "carrying a `spec_id` that
resolves to a fixture SPEC" to AC-TLE-005's Given.

**D10** — `acceptance.md:L83-92` (AC-TLE-007) — The conjunct `grep -rn AskUserQuestion
internal/cli/todo_landed*.go returns no match` adds no coverage: the landed
`TestTodoCmd_NoAskUserQuestion` (`internal/cli/todo_test.go:451-483`) already globs
`todo*.go` and therefore covers `todo_landed.go` automatically the moment it exists, with a
`scanned < 2` positive control the proposed grep lacks. Separately, `todoPromptGuard`
(`todo_test.go:485-492`) scans for two literal tokens, so neither form verifies REQ-TLE-007's
"shall prompt for nothing" against an actual stdin read (`bufio.NewReader(os.Stdin)`,
`fmt.Scan`) — measured absent from `internal/cli/todo*.go` today (`rc=1`), but nothing asserts it
stays absent. — Severity: **minor** — Class: **non-blocking** — Required fix: drop the redundant
grep conjunct and rely on the inherited guard; optionally extend `todoPromptGuard`'s token list
to cover stdin reads, which would benefit the whole todo surface rather than this verb alone.

**D11** — `research.md` §R.8 — The block labelled `$ awk 'NR>=292 && NR<=298'
internal/kanban/backlog_sqlite.go` shows five lines; the range holds seven. Measured: line 292 is
`}` and line 298 is `}`, both absent from the quoted output. The substantive claim (the version
gate rejects any value but `"1"`) is correct and the `:293-297` citation in `spec.md` §D is
exact. — Severity: **minor** — Class: **non-blocking** — Required fix: re-paste the block
verbatim or narrow the stated range to `NR>=293 && NR<=297`, so the document's own verbatim
discipline holds.

**D12** — `spec.md` §C.1 REQ-TLE-001 — The requirement prescribes the mechanism ("added by
`ALTER TABLE ... ADD COLUMN`") alongside the outcome. `design.md` §2 already owns that mechanism
and argues it; the behavioural obligation (no table rebuild, no row rewrite) is REQ-TLE-003's.
This is HOW in the requirement layer. Low impact — the storage shape genuinely is the external
contract here, which is why this is not blocking. — Severity: **minor** — Class: **non-blocking**
— Required fix: if touched, reduce REQ-TLE-001 to the observable shape and leave the statement
form to `design.md` §2.

**D13** — `spec.md` §G — REQ-TLE-019's exact-column-**set** assertion catches additions,
removals, and reordering; it does not catch a type, `NOT NULL`, or `DEFAULT` change to an
existing column, and AC-TLE-001 pins the shape of `landing` only. The guard therefore remains
partly blind on the axis §A.5 was written to close, and §G does not record this.
— Severity: **minor** — Class: **non-blocking** — Required fix: record the residual in §G, or
extend the guard assertion to compare `(name, type, notnull, dflt_value)` tuples rather than
names alone.

---

## Recommendation

FAIL at 0.80 against the Tier L threshold of 0.85. All seven must-pass criteria pass; the
shortfall is the Testability and Traceability bands, and it is concentrated in seven blocking
defects that are all small, local edits. This SPEC is closer to PASS than the score suggests —
its reasoning is unusually well grounded and every one of its ~30 `file:line` citations resolved
correctly when I re-opened it. What it does not yet do is make its criteria as strong as its
arguments.

Ordered fixes for iteration 2, cheapest-verifiable first:

1. **D1** — correct the three `completed` claims to the measured `in-progress`, or restate the
   property actually relied on. One-line edits at `spec.md` §A.3, §A.4, §B.3.
2. **D8** — change the §E milestone cell for REQ-TLE-004 from M1 to M3.
3. **D9** — add `carrying a spec_id` to AC-TLE-005's Given.
4. **D3** — correct §C.7's coverage claim; do not invent a doc-grep criterion to make it true.
5. **D5** — add the `archived_items` plant to AC-TLE-019, asserted independently.
6. **D4** — extend AC-TLE-015 to pin fields 1-5.
7. **D6** — name AC-TLE-014's exclusion set (`.git/`) and add a non-empty positive control.
8. **D7** — either downgrade §G's claim about AC-TLE-018 or strengthen AC-TLE-018 to exercise
   the pre-change open path.
9. **D2** — the substantive one. Decide whether `--sha` is validated for existence on the named
   ref. Either answer is defensible; the current text refuses the check on a ground that does not
   hold, and that ground would otherwise propagate into run-phase as settled reasoning.

The four non-blocking findings (D10-D13) are surfaced for the orchestrator's discretion and
should not by themselves drive a revision round.

---

## Gaps — what this audit did NOT observe

Stated explicitly so the empty set is not implied.

- **No run-phase behaviour was verified**, because none exists: `moai todo landed`,
  `internal/kanban/landing_evidence.go`, and the extended guard are all unwritten. Every judgment
  about them is a judgment about the *criterion*, never about an implementation.
- **I did not run `go test ./internal/cli/...`.** My probe was scoped to
  `./internal/kanban/ -run 'TestTodoHistoryAddsNoSchemaChange|TestAuditT359ColumnBlind'`. The
  pre-edit baseline over both packages is `plan.md` §C's run-phase-entry obligation and I did not
  pre-empt it; `go test ./...` is prohibited on this machine and was not attempted.
- **I did not re-run `git merge-base --is-ancestor c9f712232 develop`.** The half-A landing
  precondition is asserted in `research.md` §R.1 and `plan.md` §C already requires it re-read at
  run-phase entry because it decays. I read `SPEC-TODO-LANDING-STATE-001` `status: completed`
  directly and relied on that alone.
- **I did not enumerate external `moai todo pr` consumers.** Half A recorded that they cannot be
  grepped; I did not attempt to falsify that, so D4's severity rests on the SPEC's own risk
  statement rather than on a measured consumer.
- **I did not measure `ALTER TABLE ADD COLUMN` behaviour off this machine.** My reproduction is
  darwin/arm64 only, the same limitation §G records.
- **No cross-model backend was consulted.** `mcp__moai__audit_multi` / `codex_audit` /
  `glm_audit` were not invoked; this is a Claude-anchor-only verdict.
- **File-count arithmetic for the Tier L "> 15 files" disjunct is approximate** (~14-17 from
  `plan.md` §F plus `design.md` §5's `BacklogItem` change). I did not resolve it precisely
  because the "or constitutional" disjunct is independently satisfied.

## Residual risk

- Seven of the nine blocking defects are criterion-strength defects. Fixing them makes the
  criteria harder to pass, so iteration 2 should be expected to raise Testability rather than to
  reveal new design problems — and a score that rises without the criteria being genuinely
  strengthened would be the signal to look for.
- D2 is the one defect where a wrong decision is expensive after run-phase: a store that has
  accepted unvalidated SHAs cannot retroactively validate them, and `--clear` only helps an
  operator who already knows the value is wrong.
- My probe measured guard blindness at HEAD `b2d30deb2` in this worktree. The guard is owned by
  `SPEC-TODO-ARCHIVE-QUERY-001`; if that SPEC's owner extends it independently on `develop`, my
  measurement and REQ-TLE-019 both decay and the run-phase must re-measure rather than cite this
  report.
