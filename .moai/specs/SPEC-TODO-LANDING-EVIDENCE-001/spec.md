---
id: SPEC-TODO-LANDING-EVIDENCE-001
title: "A card that knows its own landing state, half B — the evidence store: one additive column, an operator verb that records, and an attribution rule that survives REQ-1.10"
version: "0.3.1"
status: completed
created: 2026-09-03
updated: 2026-09-08
author: manager-spec (card t359)
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban, internal/cli, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, cli, landing-evidence, sqlite, add-column, attribution, operator-act"
tier: L
depends_on:
  - SPEC-TODO-LANDING-STATE-001
related_specs:
  - SPEC-KANBAN-TODO-CLI-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-TODO-ANALYSIS-001
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-TODO-ARCHIVE-QUERY-001
---

# SPEC: A card that knows its own landing state — half B, the evidence store

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-03 | Initial plan-phase authoring (card t359), measured in worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, at HEAD `e50964ad3`. Half A (`SPEC-TODO-LANDING-STATE-001`, card t331) is landed and `completed`; this SPEC builds the storage axis its §B.2 and §D handed over. Two premises the card carried were re-measured before being built on: the `REQ-TODO-013` reading (§B.1) and the claimed disagreement between the two SPECs carrying it (§A.3 — **the disagreement does not exist**; the card's premise is falsified here rather than reconciled). The `REQ-1.10` tension is resolved on the provenance axis without amending, weakening, or reversing that requirement (§B.3). *This row described both SPECs as `completed`; v0.2.0 measured them and corrected it — see D1 below and §A.3b.* |
| 0.2.0 | 2026-09-03 | **Plan-audit iteration 1 remediation** (FAIL 0.80 vs the Tier L threshold 0.85; 9 blocking, all seven must-pass criteria passed — a score-driven FAIL concentrated in Testability 0.75). **D1**: three `completed` claims about two SPECs that measurably read `in-progress` are corrected, and the correction is followed through — §A.3b re-grounds why those requirements bind on two *measured* properties (the requirement is live in the tree; an `in-progress` SPEC makes reversal MORE disruptive, not less) instead of on a lifecycle field that was not read. **D2**: the ground for declining `--sha` validation was factually wrong — a referential-integrity check is not the card-token grep and makes no attribution claim. The false ground is withdrawn (§B.3.1), and the check is ADDED as REQ-TLE-020 / AC-TLE-020 rather than re-declined; §G now states plainly what a reachable-but-wrong operator typo costs. **D3**: §C.7's false claim that AC-TLE-015/016 verify the doctrine text is withdrawn and split into what is mechanically verified (mirror parity + stated column count, new REQ-TLE-021 / AC-TLE-021) and what is a DoD item with no criterion (the prose). **D4**: AC-TLE-015 now pins fields 1-5, not only the count and the two it adds. **D5**: AC-TLE-019 gains independent `archived_items` and tuple-drift plants (019a/b/c) — a guard extended for `items` alone previously satisfied it in full. **D6**: AC-TLE-014 excludes `.git/` (the `todo pr` git subprocess may write there during a read) and gains a non-empty + queue-present positive control so the exclusion cannot hollow it out. **D7**: the downgrade gap stops being deferred to a criterion that does not close it — AC-TLE-018 gains a reconstructed pre-change open path, and §G keeps REQ-TLE-018 listed as an argued claim with a partial demonstration. **D8**: §E maps REQ-TLE-004 → M3 (its criterion needs the M3 verb). **D9**: AC-TLE-005's Given gains the `spec_id` its Then asserts on. Non-blocking: **D10** redundant prompt-guard conjunct dropped in favour of the inherited `todo*.go` guard; **D11** §R.8's mis-labelled awk block re-pasted verbatim; **D12** REQ-TLE-001 reduced to the observable shape with the statement form left to `design.md` §2; **D13** the guard's remaining type/nullability blindness is closed by the REQ-TLE-019 tuple assertion rather than merely recorded. Counts: 19 → **21 requirements / 21 criteria** (Tier L ceiling 25/25). No source file was modified. |
| 0.3.0 | 2026-09-03 | **Plan-audit iteration 2 delta** (PASS-WITH-DEBT 0.89 vs the Tier L threshold 0.85; trajectory 0.80 → 0.89, all thirteen iteration-1 defects resolved). A scoped three-item debt closure, not a full round — no decision is reopened. **E1**: AC-TLE-020's attribution-boundary clause could not fail. It varied the card's *text* (a rename) while the predicate it guards against keys on the card *id* — `LandedGrepArgs` builds `` `--grep=\b` + cardID + `\b` `` (`prlink_landed.go:96-108`) and no registered `todo` verb changes an id (`todo.go:148-152`), so an implementation leaking the card token passed with the leak intact. Replaced with **one `--sha` recorded against two different card ids**, asserted on both an accepting and a refusing branch. **E2**: AC-TLE-018 clause (b) was a coverage assertion, not a detector — its only listed RED (the version bump) is already caught by clause (a), and no permitted mutation redded (b) alone; the v0.2.0 claim that the two "fail independently" is **withdrawn**. New clause (c) asserts the frozen replica equals the live `backlogDDL` and accepted-version set, which supplies (b)'s independent RED (edit the live const, leave the copy — only (c) fails). §G gains **forward drift** as a hazard distinct from the released-binary divergence it already named. **E3**: AC-TLE-020 gains case (d), the cannot-be-run branch REQ-TLE-020 binds, with the write-refuses / read-degrades contrast stated so the `todo_pr.go` fail-open habit is not carried across. Non-blocking: **E4** the status block is re-pasted with full paths and its ordering disclosed; **E5** `plan.md` §B renumbered 1-5; **E6** AC-TLE-021 now re-renders and counts its own row rather than citing another test's runtime value, with `7` demoted to a note; **E7** §G's "nothing flags it" is **withdrawn as an overstatement** — rendering the commit's own subject beside the SHA attributes nothing and is exactly what lets a human notice, so it is recorded as considered-and-not-built, with the record-time capture named as the only shape compatible with §D's no-new-read-cost exclusion. Counts unchanged at **21 requirements / 21 criteria**. No source file was modified. |
| 0.3.1 | 2026-09-03 | **Plan-audit iteration 3 closure — documentation only** (PASS-WITH-DEBT 0.92 vs the Tier L threshold 0.85; trajectory 0.80 → 0.89 → 0.92, seven of seven must-pass criteria, no regression). Iteration 3 is the last permitted round, so its three blocking items are closed here rather than carried into run-phase; no decision is reopened and no requirement or criterion is added. **F1**: AC-TLE-020's attribution-boundary RED was written unconditionally but fires only when the fixture ref's history mentions **exactly one** of the two card ids — mentioning neither makes both invocations refuse alike (the leak surfaces at condition (a) instead), mentioning both lets the leak pass all three clauses. The Given now pins that premise and the RED hangs on it, with both degenerate cases stated. **F2**: AC-TLE-018 clause (c)'s **second** conjunct — that the frozen switch's accepted-version set equals the live one — is **withdrawn as unrealisable**, not softened or restated more loosely. The live set is control flow at `backlog_sqlite.go:286-298`, not a structure a test can extract, and its cheapest runnable reading (comparing the `backlogSchemaVersion` const) misses the very drift it named, since adding a `case` changes no const; the criterion's RED list never carried a mutation that redded it, which is the same absence seen from the other side. The first conjunct — DDL byte-identity, which has its own stated RED — is untouched, so clause (c) still supplies (b)'s independent failure mode. **F3**: `research.md` §R.10.1's v0.3.0 disclosure attributed the block's ordering to `grep -m1` resolving files out of argument order. `grep -m1` has no such property — `/usr/bin/grep -m1` emits in argument order, 6/6 identical runs measured at HEAD `c0cfb2520`; the variance belongs to this shell's `grep`, a function wrapping a parallel grep whose output order differs between runs of the identical command. The causal claim is **withdrawn**, the block is labelled one capture of a non-deterministic ordering, and `spec.md` §A.3b's twin block — re-pasted at v0.3.0 with full paths but no note — carries the same disclosure so the two surfaces agree. The four status values were never in dispute and match re-measurement. Not fixed: the audit's optional F4 (case (d)'s two sub-conditions share one stderr classification), which the operator holds. Counts unchanged at **21 requirements / 21 criteria**. No source file was modified. |

> **Provenance discipline.** Every `file:line` citation in this document was measured at HEAD
> `e50964ad3` in the worktree `.claude/worktrees/t359`. Where a claim rests on a command rather
> than on a file, the command and its verbatim output are recorded in `research.md` §R and in
> `.moai/reports/t359/`. Nothing here is carried over from another tree or another time; a figure
> that could not be measured in this tree is recorded in §G Gaps rather than stated.

---

## §A Context

### A.1 What half A left standing, and what it deliberately did not build

Half A repaired the landed **predicate**: the ref it asks about is now resolved from the project's
configured integration branch rather than compiled to `origin/main`, the answer became three-valued
(`landed` / `not-landed` / `unknown`), and the verdict reaches stdout where a caller reads it.

Half A stored nothing, and said so as a ruling rather than an omission
(`SPEC-TODO-LANDING-STATE-001` §B.2, `spec.md:408-421`):

> Ruled: **nothing is persisted.** This SPEC changes what the landed question asks and how its
> answer is reported; it stores no landing evidence, adds no field to the queue record, adds no
> table, and renders no observed commit.
>
> **Card t359 owns the storage axis** … **t359 depends on this SPEC landing first**: an evidence
> record written from a predicate that asks the wrong ref stores wrong evidence permanently, and a
> stored observation costs far more to correct than a live answer.

That precondition is met. `SPEC-TODO-LANDING-STATE-001` reads `status: completed`, and
`git merge-base --is-ancestor c9f712232 develop` returns `rc=0` — measured in this worktree, output
in `research.md` §R.1.

### A.2 Why a live answer is not enough

The live answer is computed at the instant it is asked, from a ref that moves. That is the right
shape for a *question*, and the wrong shape for a *record*:

- A lane that observed a landing and closed its worktree leaves nothing behind. The next reader
  re-asks the question against a ref that has moved by hundreds of commits, and gets an answer about
  a different history than the one the lane saw.
- The answer cannot distinguish a **run** landing from a **sync** landing — half A says so plainly
  in the doctrine and records the gap as needing "evidence this SPEC does not collect (card t359)"
  (`SPEC-TODO-LANDING-STATE-001` §D).
- An `unknown` answer is not preserved anywhere, so the fact that the question *went unasked* on a
  given day is unrecoverable.

Evidence is the missing half: a record of what was observed, when, against which ref, and on whose
authority.

### A.3 The card's D1 premise, re-measured — the two SPECs do NOT disagree

The dispatching card states that prior reasoning misattributed `REQ-TODO-013`, and instructs that
if the requirement's own text permits `ADD COLUMN`, the design is to be built on that reading. It
further instructs reconciling a record in `SPEC-TODO-ANALYSIS-001` (`status: completed`) "that judged the
opposite".

**The first half of the premise is confirmed. The second half is falsified.**

`REQ-TODO-013` (`SPEC-KANBAN-TODO-CLI-001/spec.md:59`) reads, verbatim:

> **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record shape
> — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with
> `state ∈ {queued, picked, dropped}` — **changing it only additively** (the high-water mark, per
> REQ-TODO-009).

It does not freeze the field set. It constrains the *manner* of change to additive, and names its
own precedent for one.

`SPEC-TODO-ANALYSIS-001/spec.md:51` — the record the card asked to be reconciled — reads, verbatim:

> 항목당 5필드(`id`, `text`, `added_at`, `spec_id`, `state`)는 SPEC-KANBAN-TODO-CLI-001 REQ-TODO-013
> 이 고정한 계약이다. 다만 그 SPEC의 §E는 **자기 범위**의 out-of-scope 선언이지 영구 동결이 아니고,
> 같은 REQ가 "additively" 변경을 허용한다 — `last_seq` 최상위 필드가 그 선례다.

Translated to the claim it makes: the five fields are the contract REQ-TODO-013 fixes; **but** that
SPEC's §E is a scope-local out-of-scope declaration rather than a permanent freeze, and the same REQ
permits additive change — with the top-level `last_seq` field as the precedent.

That is the same reading this SPEC builds on. There is no disagreement between the two SPECs to
reconcile, and no edit to either is required or performed. What existed was a
**summary** that dropped the qualifying clause after the first sentence. §B.1 records the reading
once, with the verbatim text beside it, so the next reader does not re-derive it from a summary.

The `last_seq` precedent is present in the tree: `internal/kanban/backlog_store.go:187-193` carries
`LastSeq int \`json:"last_seq"\`` as a top-level field of `BacklogRecord`, alongside the five-field
`BacklogItem` at `:65-71`.

### A.3b The measured status of every SPEC this one leans on, and why status is not the reason they bind

Version 0.1.0 of this document called two of them `completed`. That was inherited from the card
text and never measured, and it is wrong. Measured in this tree:

```
$ grep -m1 '^status:' .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md \
    .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md \
    .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md \
    .moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md
.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:status: in-progress
.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:status: in-progress
.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md:status: completed
.moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md:status: completed
```

The block is **one capture of a non-deterministic ordering**: `grep` in this shell is a function
wrapping a parallel grep, whose output order varies between runs of the identical command, so the
lines do not follow the argument order and a reader re-running it will often see a different
sequence. `/usr/bin/grep -m1` emits in argument order deterministically; the variance is the
wrapper's. The four values are what the criterion rests on, and they are stable under
re-measurement. (`research.md` §R.10.1 carries the same block and the same note.)

So `REQ-TODO-013` (§A.3), `REQ-1.10` (§B.3), and `REQ-2.1` (§A.4, §B.2) all come from SPECs that
read `in-progress`, not `completed`.

**This changes the argument, and the corrected argument is the stronger one.** Version 0.1.0
leaned on "a completed SPEC's requirement cannot be quietly reversed", which made the binding a
property of a *lifecycle field* — the weakest available ground, and one that was not even true.
The two properties actually relied on are these, and both were measured rather than read off a
frontmatter line:

- **The requirement is live in the tree.** REQ-2.1's prohibition is a stated property of the file
  that implements it (`internal/cli/todo_pr.go:1-15`), and REQ-1.10's shape is visible in
  `PRLinkOutcome` carrying no delivering-commit field at all
  (`internal/kanban/prlink.go:101-114`: `CardID`, `Kind`, `PRs`, `PRState`, `Confidence`).
  Reversing either is a change to shipped behaviour that other code already depends on, whatever
  its SPEC's frontmatter says.
- **`in-progress` makes reversal *more* disruptive, not less.** A SPEC still being built out has
  remaining milestones that may rest on the requirement, and its consumer set is still moving. A
  closed SPEC at least has a stable, enumerable set of things that depend on it. "It is only
  in-progress, so the requirement is soft" is precisely backwards, and is recorded here so the
  inference is not available to a later reader.

Neither property depends on a status, which is why this SPEC no longer cites one as a reason.

### A.4 The constraint the card did not name, and which decides the design

`SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-2.1 (`spec.md:259-261`) is [HARD] — and binds for the reasons
in §A.3b, not because of its lifecycle status:

> **REQ-2.1** — [HARD] The read surface shall leave `.moai/state/kanban/backlog.json`
> byte-identical across an invocation, and shall write no field, no `findings[]` entry, and no
> timestamp.

`moai todo pr` **is** that read surface, and the file states the property of itself
(`internal/cli/todo_pr.go:1-15`: "It reads, and it writes NOTHING — not a field, not a findings
entry, not a timestamp, not a cache, not a lock").

So the surface that *detects* a landing may not be the surface that *records* one. Who writes the
evidence, and when, is therefore not an implementation detail — it is the load-bearing design
question, and §B.2 answers it.

### A.5 The schema-freeze guard is column-blind — measured, not inferred

`internal/kanban/backlog_schema_freeze_test.go` (`TestTodoHistoryAddsNoSchemaChange`) asserts four
things: the exact table set, that the stored `items` SQL *contains* the three-state CHECK, the exact
non-auto index set, and `schema_version == "1"`. **None of the four is a column-set assertion.**

This was measured rather than read off: a temporary probe planted
`ALTER TABLE items ADD COLUMN bogus TEXT` and re-ran all four assertions. All four stayed green
(`.moai/reports/t359/measure-guard-column-blind.txt`):

```
A1 table set     = "archived_findings archived_items findings items meta" -> match=true
A2 CHECK present = true
A3 index set     = "idx_items_state" -> match=true
A4 schema_version = "1" -> match=true
COLUMN SET (unasserted by the guard) = "seq id text added_at spec_id state bogus"
```

The mechanism is visible in the same output: SQLite appends an added column *after* the closing
constraint, so the stored SQL becomes `… CHECK (state IN ('queued','picked','dropped'))\n, bogus TEXT)`
and a `strings.Contains` test on the CHECK clause still matches.

The consequence is the reason REQ-TLE-019 exists: **this SPEC's own column would land silently.**
An unextended guard is how the change after this one drifts. The probe file was deleted immediately
after the measurement; `git status --short` reports no modification under `internal/`.

### A.6 What the store looks like today

Measured at HEAD `e50964ad3`:

| Table | Columns (`PRAGMA table_info`, measured) | Source |
|---|---|---|
| `items` | `seq id text added_at spec_id state` | `internal/kanban/backlog_sqlite.go:107-114` |
| `archived_items` | `seq id text added_at spec_id state position` | `internal/kanban/backlog_sqlite.go:124-132` |

Every production statement naming `items` columns names them **explicitly** — there is no
`SELECT *` on the path. `grep -rn 'SELECT .*FROM items\|INSERT INTO items' internal/kanban/*.go`
excluding tests returns exactly four lines, of which two carry the column list
(`backlog_migrate.go:60` SELECT, `:276` INSERT) and two are aggregates (`:340`, `:351`). This is what
makes REQ-TLE-018's downgrade claim buildable rather than hopeful.

`ALTER TABLE` appears **nowhere** in `internal/kanban/` or `internal/cli/` today
(`grep -rn "ALTER TABLE" internal/kanban/ internal/cli/` → `rc=1`, no output). This change
introduces the project's first one, which is why REQ-TLE-003 states the idempotence rule rather than
assuming the reader will supply it.

---

## §B Decisions

### B.1 Decision 1 — the storage is one additive nullable column, and it holds a record, not a scalar

Ruled: **`ALTER TABLE items ADD COLUMN landing TEXT`** — nullable, no `DEFAULT`, no `CHECK`, no
index, and no `schema_version` bump. `archived_items` gains the same column.

Grounded on `REQ-TODO-013`'s own text (§A.3): additive change is permitted, and a nullable column
that no existing statement names is the most additive change available — every row already in the
field keeps its bytes, and every statement already written keeps working.

**Why one column holding a record, rather than four scalar columns.** The evidence is four-to-six
facts (§B.4). Expressed as separate columns they would be four `ALTER`s, four entries in the
migration's column lists, and four fields in the parity comparison — and the *next* fact would be a
fifth `ALTER` against every operator queue in the field. Expressed as one nullable TEXT column
holding a JSON object, the shape extends by adding a key, which is the same additive move
`REQ-TODO-013` already blesses and which `last_seq` already set a precedent for. The Go boundary
decodes it into a typed struct, so the value is stringly-typed in SQLite only, never in the program.

**Why not a fifth table.** A table is the right shape for a one-to-many relation, and this is
one-to-at-most-one: a card carries the latest landing observation or none. `findings` is a table
because a card carries many findings. Adding a table here would also break the guard's table-set
assertion — deliberately, in a change whose whole point is that it must be *cheap* on an existing
queue.

**Why not a fourth `state` value.** Forbidden by the card, and the reason is in the tree:
`internal/kanban/backlog_sqlite.go:95-97` records that "SQLite cannot ALTER a CHECK constraint, so
admitting a fourth state would need a table rebuild on every operator queue in the field."
`ALTER TABLE ADD COLUMN` needs no rebuild; changing the CHECK does. The two are not comparable
costs, and REQ-TLE-004 keeps the CHECK exactly as it is.

### B.2 Decision 2 — the writer is a new operator verb, because the read surface may not write

Ruled: **`moai todo landed <id> [--sha <sha>] [--ref <ref>] | --clear`** writes the evidence, as a
single locked write, and it is an operator act in the sense `todo.md` already fixes for `add`,
`relate`, `drop`, and the pick.

This answers the question §A.4 raises. Three candidates were considered:

| Candidate | Verdict |
|---|---|
| `moai todo pr` records what it observed | **Rejected.** REQ-2.1 is [HARD] and forbids the read surface writing a field or a timestamp. Recording on a read would also make every read a write, so a reader's glance would mutate the queue — the precise shape `todo.md`'s operator-only rule exists to prevent. |
| `moai todo pr --record` | **Rejected.** The flag does not change what the surface is. REQ-2.1 binds the surface, not the invocation, and a write path reachable by a flag is a write path. |
| A new explicit verb | **Adopted.** The operator asserts the landing; the queue records the assertion. `todo pr` stays byte-identical, keeps its one `gh` query, and gains no write path. |

**The verb records; it never transitions.** This is the card's first [HARD] constraint and half A's
§B.4 non-goal, and REQ-TLE-008 makes it mechanical rather than aspirational: the verb writes the
`landing` column and touches no other column of any row. A `picked` card that records a landing
stays `picked` until the operator runs `done`. The failure this prevents is concrete and was already
paid once: an automatic close on a **run** landing closes a card whose sync commit is still unpushed
in a lane's worktree.

`todo.md:59-63` is the [HARD] rule the verb sits inside, verbatim:

> `edit`, `move`, `drop`, `undrop`, `done`, and `undone` are operator acts, exactly like `add` and
> the pick. … The queue records the operator's intent; it does not curate it.

### B.3 Decision 3 — REQ-1.10 is preserved, and the resolution is provenance, not permission

`SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 (`spec.md:251-255`) reads, verbatim:

> **REQ-1.10** — The resolver shall not name, return, or otherwise claim which commit delivered a
> card. The `landed` outcome is a boolean fact about `origin/main` and nothing more. (Grounds: §C.2
> — a card's first matching commit may be another card's report commit that merely mentions it, so
> any "first match is the delivering commit" reading attributes wrongly.)

Ruled: **REQ-1.10 stands, unamended and unreversed.** No requirement in this SPEC weakens it, and
the resolver's behaviour is unchanged.

The naive reading of this card's axis — "get permission to show a SHA" — is refused, because the
grounds REQ-1.10 states are still true: first-match attribution mis-attributes, and a stored
mis-attribution is *worse* than a rendered one, since it outlives the invocation that made it. The
correct question is the one the dispatch posed: **what makes a stored SHA a correct attribution
where a resolved one is not?**

The answer is **provenance, not permission**:

- A *resolved* SHA is inferred by a predicate known to mis-attribute. The machine would be making a
  claim it has no basis for. REQ-1.10 forbids exactly this, and REQ-TLE-012 forbids it again at the
  storage boundary — the system may never derive a delivering SHA from the card-token grep.
- An *operator-asserted* SHA is a claim by the party with the authority to make it, arriving through
  the same channel as the card's own text. The queue records operator intent; a SHA the operator
  typed is intent, not inference. REQ-1.10 binds *the resolver*; it does not forbid the operator
  from stating a fact and the queue from recording it.
- Where no operator SHA was supplied, the record still narrows to something true: a **ref position**
  — the head SHA of the ref at the observation instant. That is a fact about where the ref stood,
  and REQ-TLE-013 requires it be carried and rendered as such. It is never presented as a delivering
  commit, and REQ-TLE-016 requires the two to be distinguishable by a machine, not merely by a
  careful human reading a column.

So the design never claims delivery on the machine's authority. Where it names a commit, the
operator named it; where the operator did not, it names a ref position and says that is what it is.

#### B.3.1 Provenance is not enough on its own — the SHA is checked for referential integrity

Provenance answers *who claimed it*. It does not answer *whether the claim names a real commit*,
and version 0.1.0 of this document conflated the two: it declined all validation of `--sha` on the
ground that "validating it against the grep predicate would reintroduce exactly the inference
REQ-1.10 forbids". **That ground is false and is withdrawn.** Two different things were collapsed
into one:

| Check | What it asks | Does it attribute? |
|---|---|---|
| the card-token grep predicate | *which* commit delivered card `tX` | **Yes** — and wrongly, per REQ-1.10's own grounds |
| `git rev-parse --verify <sha>^{commit}` | does this object exist, and is it a commit | No |
| `git merge-base --is-ancestor <sha> <ref>` | is that commit reachable from this ref | No |

The second and third are **referential-integrity** checks. They make no claim about which card a
commit delivered — they cannot, because the card id is not an input to either. Refusing them
alongside the grep predicate was a category error, and its consequence was that `sha_source:
operator` became the only thing standing between the store and an arbitrary string: the marker
recorded who typed it, and nothing recorded whether it named anything at all.

Ruled: **a supplied `--sha` is validated for existence and reachability before it is stored**
(REQ-TLE-020). Three reasons, in order of weight:

1. **This SPEC's own §B.3 argument demands it.** "A stored mis-attribution is worse than a rendered
   one, since it outlives the invocation that made it" is a property of **storage**, not of
   authorship. An operator typo produces exactly the durable wrong record that sentence calls
   worse. An argument that condemns machine mis-attribution and then waves through operator
   mis-typing is not an argument; it is a preference.
2. **The check costs nothing that is not already being paid.** It runs in the *write* verb, never
   on a read path, so the §D no-new-read-cost exclusion is untouched. The verb already invokes git
   to read `ref_head` (REQ-TLE-005), so the marginal cost is two local invocations on a path the
   operator entered deliberately.
3. **The record is otherwise internally incoherent.** A record asserts "this commit delivered the
   card, observed against this ref". A SHA unreachable from that ref contradicts the record's own
   `ref` field regardless of who typed it. Validating reachability is checking the record against
   itself — not checking the operator against the machine.

What the check explicitly does **not** establish is stated in REQ-TLE-020 itself and repeated here
because it is the clause that keeps REQ-1.10 intact: it asserts **existence and reachability, never
delivery**. A reachable commit that the operator named in error is stored, and is indistinguishable
from a correct one. §G records what that residual costs.

### B.4 Decision 4 — what a record carries, and why the SPEC status is observed at record time

A record carries six facts:

| Fact | Why it is in the record |
|---|---|
| `ref` | The answer is meaningless without the ref it was asked about — half A's whole root cause. |
| `ref_head` | Where that ref stood at the observation instant, so a later reader can re-derive the history the observer saw. Marked as a ref position (REQ-TLE-013). |
| `observed_at` | Evidence is a fact about an instant. A reader judges staleness from it rather than from the record's mere existence. |
| `sha` (optional) | The delivering commit, when and only when the operator supplied it. |
| `sha_source` | `operator` when a SHA is present; absent otherwise. This is the field a machine keys on to tell an assertion from an observation (REQ-TLE-016). |
| `spec_status` | The SPEC's frontmatter `status` at record time, when the card carries a `spec_id`. |

**The SPEC status is what separates a run landing from a sync landing** — the fourth item half A
handed over. It is read **at record time by the writing verb**, not on the read path, for two
reasons. It keeps `todo pr` at its measured cost (one `gh` query, no per-card file I/O — the
property `internal/cli/todo_pr.go:8-13` justifies at length and half A's §D protects). And a status
read on the read path would answer about *today's* SPEC file, which is a different question from
"what did the observer see"; the record wants the latter.

The consequence is that a stored `spec_status` can go stale. That is accepted, not overlooked: the
record carries `observed_at` beside it precisely so staleness is visible, and §G records it as a
residual risk. An unreadable or absent SPEC stores the status as unknown (REQ-TLE-010) — never a
fabricated one, which would be an unobserved claim in persistent form.

### B.5 Decision 5 — a record is replaceable and clearable

Ruled: re-recording **replaces**; `--clear` removes; a card holds at most one record.

Half A's own argument for splitting the axis was that "a stored observation costs far more to
correct than a live answer". A store with no correction path takes that cost and makes it permanent
— a mistyped `--sha` would be unfixable except by editing the database by hand. Replace-and-clear is
the cheapest correction path that keeps the verb an operator act, and it makes retention a
non-problem: with at most one record per card there is nothing to prune, which is why retention is
recorded as out of scope in §D rather than designed.

### B.6 Decision 6 — the evidence travels into the archive

`archived_items` gains the same column (REQ-TLE-002). Without it, a card's evidence is destroyed by
`done` — at exactly the moment the evidence becomes historically interesting, and by a verb whose
whole purpose is closing a card whose work landed. The archive-parity comment already in the tree
(`internal/kanban/backlog_migrate.go:604-608`) makes the same argument for archived rows generally:
omitting them "would let the migration drop exactly the rows this SPEC exists to preserve, while
still reporting success."

---

## §C Requirements

Twenty-one requirements (Tier L ceiling 25). Every one carries at least one acceptance criterion in
`acceptance.md`, and §E maps them both ways.

### C.1 Storage

- **REQ-TLE-001** (Ubiquitous) The backlog store shall carry landing evidence in one column named
  `landing` on the `items` table, of type `TEXT`, nullable, with no `CHECK` constraint, no
  `DEFAULT`, no index, and no `schema_version` bump. (The statement form that achieves this is
  `design.md` §2's; the behavioural obligation not to rebuild or rewrite is REQ-TLE-003's.)
- **REQ-TLE-002** (Ubiquitous) The `archived_items` table shall carry a `landing` column of the same
  shape, so a card's evidence survives its close.
- **REQ-TLE-003** (Event-driven) **When** the engine opens a backlog database, it shall add each
  missing `landing` column idempotently — deciding from the table's own column metadata rather than
  from a failed statement — and shall rebuild no table and rewrite no existing row.
- **REQ-TLE-004** (Ubiquitous) The `items.state` CHECK shall admit exactly `queued`, `picked`, and
  `dropped`; no evidence write shall alter any card's `state`.

### C.2 The record

- **REQ-TLE-005** (Ubiquitous) A landing evidence record shall carry the ref the observation was
  made against, that ref's head SHA at the observation instant, the observation instant itself, the
  provenance of any commit SHA it names, the operator-supplied delivering commit SHA when one was
  supplied, and the SPEC status observed at record time.
- **REQ-TLE-006** (Ubiquitous) A card with no landing evidence shall store SQL `NULL`, and every
  surface shall render its evidence as absent — never as `not-landed`, and never as a present record
  with empty fields.

### C.3 The recording verb

- **REQ-TLE-007** (Event-driven) **When** the operator runs `moai todo landed <id> [--sha <sha>]
  [--ref <ref>]`, the command shall write exactly one evidence record for the addressed card as a
  single locked write, shall prompt for nothing, and shall follow the `internal/cli` conventions
  (structured stdout under `--json`, human-readable stderr, exit 0/1/2).
- **REQ-TLE-008** (Ubiquitous) The recording verb shall not change any card's `state`, position,
  text, or `spec_id`, and shall not close, archive, drop, re-order, or admit a card.
- **REQ-TLE-009** (Event-driven) **When** a record already exists for the addressed card, a further
  `landed` shall replace it rather than accumulate beside it; **and when** the operator runs
  `moai todo landed <id> --clear`, the command shall remove the record, returning the column to
  `NULL`.
- **REQ-TLE-010** (Capability gate) **Where** the addressed card carries a `spec_id`, the recording
  verb shall read that SPEC's frontmatter `status` at record time and store what it read; **when**
  the SPEC document or its status cannot be read, it shall store the status as unknown rather than
  inferring, defaulting, or omitting one.

### C.4 Attribution

- **REQ-TLE-011** (Ubiquitous) The landed resolver shall not name, return, or otherwise claim which
  commit delivered a card; `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 is preserved unchanged and no
  requirement in this SPEC amends it.
- **REQ-TLE-012** (Ubiquitous) A stored delivering commit SHA shall be stored only when the operator
  supplied it, and the record shall mark it operator-asserted; the system shall never derive a
  delivering SHA from the card-token grep predicate or from any other match set.
- **REQ-TLE-013** (Ubiquitous) The ref head SHA a record carries shall be marked as an observed ref
  position and shall not be presented, keyed, or rendered as a delivering commit.
- **REQ-TLE-020** (Event-driven) **When** the operator supplies `--sha`, the recording verb shall
  store it only after resolving it to an existing commit object and confirming that commit is
  reachable from the record's `ref`, storing the full resolved SHA rather than the supplied form;
  **when** either check fails or cannot be run, the verb shall write nothing and exit 1, naming
  which check failed. This asserts **existence and reachability, never delivery** — the card id is
  not an input to either check, so neither can and neither does attribute a commit to a card
  (§B.3.1).

### C.5 Read surfaces

- **REQ-TLE-014** (Ubiquitous) `moai todo pr` shall remain a read surface: it shall leave the
  project byte-identical across an invocation and shall write no field, no finding, no timestamp,
  and no cache — `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-2.1 preserved.
- **REQ-TLE-015** (Event-driven) **When** `moai todo pr` renders its rows, each row shall carry a
  seventh tab-separated column holding the stored evidence, placed before the card-text tail so the
  card text remains the last field; **and** the `--json` output shall carry the same record under
  its own key.
- **REQ-TLE-016** (Ubiquitous) The rendered evidence shall distinguish an operator-asserted
  delivering commit from an observed ref position by a marker a consumer can key on, not by the SHA
  value alone.

### C.6 Compatibility and the guard

- **REQ-TLE-017** (Ubiquitous) The JSON⇄SQLite round trip shall preserve the landing evidence for
  live and archived cards alike, and the migration's parity verification shall compare it, so a
  downgrade-then-upgrade cycle cannot drop it while still reporting success.
- **REQ-TLE-018** (Ubiquitous) A binary predating this change shall continue to open and serve a
  database carrying the `landing` columns: `schema_version` shall remain `"1"` and every production
  statement shall name its columns explicitly.
- **REQ-TLE-019** (Ubiquitous) The schema-freeze guard shall assert, for `items` and for
  `archived_items` independently, the exact ordered sequence of `(name, type, notnull, dflt_value)`
  column tuples — so a further column, a removed column, a reordering, or a type / nullability /
  default change to an existing column is each a deliberate act rather than a silent one.

### C.7 Doctrine

- **REQ-TLE-021** (Ubiquitous) The two doctrine surfaces —
  `.claude/skills/moai/workflows/todo.md` and its template mirror at
  `internal/template/templates/.claude/skills/moai/workflows/todo.md` — shall carry byte-identical
  `moai todo` verb-table and `todo pr` rows, and the column count those rows state shall equal the
  count the rendered row actually carries.

**What has a criterion behind it, and what does not.** Version 0.1.0 claimed AC-TLE-015 and
AC-TLE-016 "verify what those files must state". That was false: AC-TLE-015 asserts a field split
and a JSON key, AC-TLE-016 asserts a marker survives SHA substitution, and neither opens
`todo.md`. The claim is withdrawn and replaced by an accurate split:

- **Mechanically verified** (AC-TLE-021): mirror parity between the two files, and agreement
  between the column count the doctrine *states* and the count the surface *emits*. Both are
  cross-artifact comparisons with a reachable red — edit one file only, or leave the row saying
  six columns, and the criterion fails.
- **Not verified by any criterion** (`acceptance.md` §D.3 Definition of Done only): the doctrine
  *prose* — the verb's description and the evidence-is-not-a-transition sentence. This is
  deliberate. A criterion asserting that a sentence is present is satisfied by pasting the
  sentence, which measures nothing; the behaviour that sentence describes is verified where it can
  actually fail, by AC-TLE-008. Recording this as an uncovered DoD item is the honest position and
  is repeated in §G.

---

## §D Exclusions

This SPEC deliberately does not build the following. Each is out of scope for a stated reason.

### Out of Scope — automatic state transitions

- Any automatic close, archive, drop, re-order, or re-state on a recorded or detected landing. This
  is the card's first [HARD] constraint and half A's §B.4 non-goal; REQ-TLE-008 makes it mechanical,
  and it is restated here so a reader who reads only §D still meets it.
- Any change to **who** may issue a queue-mutating verb. The [HARD] operator-only doctrine at
  `todo.md:59-63` stands unchanged.
- Turning `moai todo pr`, `moai todo list`, or any other read verb into a write path.

### Out of Scope — a fourth state value

- Any fourth `items.state` value, including `landed`. `internal/kanban/backlog_sqlite.go:95-97`
  records that SQLite cannot ALTER a CHECK constraint, so a fourth value forces a table rebuild on
  every operator queue in the field. REQ-TLE-004 keeps the three-value CHECK exactly as it is.
- Any `schema_version` bump. `internal/kanban/backlog_sqlite.go:293-297` rejects any value but `"1"`
  as `ErrBacklogCorrupt`, so a bump is a downgrade break rather than a feature.

### Out of Scope — reversing or amending REQ-1.10

- Any change to the resolver's behaviour, including naming a delivering commit, selecting among the
  several commits a card's grep predicate matches, or ranking them. §B.3 resolves the tension on the
  provenance axis precisely so this reversal is not needed, and REQ-TLE-011 restates the prohibition
  as binding here.

### Out of Scope — evidence history and its lifecycle

- More than one record per card: no observation history, no append log, no per-phase series. §B.5's
  replace-and-clear shape is what makes this cheap.
- Retention, pruning, compaction, or expiry of evidence. With at most one record per card there is
  nothing to prune; this is a consequence of §B.5 rather than a separate decision.
- Any automatic staleness check, warning, or refresh of a stored `spec_status`. The record carries
  `observed_at`; judging staleness from it is the reader's act.

### Out of Scope — new cost on the read paths

- Any new subprocess, network call, or per-card file read on `moai todo pr`, `moai todo list`,
  `moai todo next`, or any other read verb. The SPEC-status read happens once, in the writing verb
  (§B.4).
- Any change to the one-`gh`-query-per-invocation budget
  (`internal/cli/todo_pr.go:8-13`).

### Out of Scope — adjacent surfaces half A left open

- Flipping `--require-landed` to default-on, or reversing the proceed-on-unanswerable policy. Half A
  §D excludes both; storing evidence does not close either question.
- Making the live predicate phase-aware. This SPEC stores an observed phase; it does not teach the
  resolver to compute one.
- `tier` / `depends_on` normalization on the card record, excluded by half A §D on the grounds that
  its precondition is a doctrine change that is not open.

---

## §E Traceability

Every requirement has at least one acceptance criterion; every criterion traces to exactly one
requirement. Criterion bodies, their RED conditions, and the mutants that establish them live in
`acceptance.md`.

| Requirement | Criterion | Milestone |
|---|---|---|
| REQ-TLE-001 | AC-TLE-001 | M1 |
| REQ-TLE-002 | AC-TLE-002 | M1 |
| REQ-TLE-003 | AC-TLE-003 | M1 |
| REQ-TLE-004 | AC-TLE-004 | M3 |
| REQ-TLE-005 | AC-TLE-005 | M2 |
| REQ-TLE-006 | AC-TLE-006 | M2 |
| REQ-TLE-007 | AC-TLE-007 | M3 |
| REQ-TLE-008 | AC-TLE-008 | M3 |
| REQ-TLE-009 | AC-TLE-009 | M3 |
| REQ-TLE-010 | AC-TLE-010 | M3 |
| REQ-TLE-011 | AC-TLE-011 | M2 |
| REQ-TLE-012 | AC-TLE-012 | M3 |
| REQ-TLE-013 | AC-TLE-013 | M2 |
| REQ-TLE-014 | AC-TLE-014 | M4 |
| REQ-TLE-015 | AC-TLE-015 | M4 |
| REQ-TLE-016 | AC-TLE-016 | M4 |
| REQ-TLE-017 | AC-TLE-017 | M5 |
| REQ-TLE-018 | AC-TLE-018 | M5 |
| REQ-TLE-019 | AC-TLE-019 (a/b/c) | M1 |
| REQ-TLE-020 | AC-TLE-020 | M3 |
| REQ-TLE-021 | AC-TLE-021 | M5 |

---

## §F Cross-references

- `SPEC-TODO-LANDING-STATE-001` — half A. `depends_on`; its §B.2 and §D hand over this axis.
- `SPEC-KANBAN-TODO-CLI-001` REQ-TODO-013 — the additive-change permission this SPEC builds on
  (§A.3, §B.1); REQ-TODO-014 — the no-prompt rule REQ-TLE-007 inherits.
- `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 (preserved, §B.3) and REQ-2.1 (preserved, §A.4/§B.2).
- `SPEC-TODO-ANALYSIS-001` §A.4 — the record the card asked to reconcile; measured as agreeing
  (§A.3).
- `SPEC-TODO-DESTRUCTIVE-GUARD-001` — the archive tables REQ-TLE-002 extends.
- `SPEC-TODO-ARCHIVE-QUERY-001` — owner of the schema-freeze guard REQ-TLE-019 extends.
- `.claude/skills/moai/workflows/todo.md` — the operator-act doctrine (`:59-63`) and the `todo pr`
  outcome list (`:57`) both surfaces must state.

---

## §G Gaps and residual risk

Recorded explicitly, per the evidence discipline: what was **not** measured, and what could still be
wrong despite what was.

**Not measured in this tree:**

- No `moai todo landed` verb exists yet, so nothing about its runtime behaviour was measured — every
  claim about it is a design intent to be verified in run-phase against `acceptance.md`.
- **No pre-change binary is ever run against a post-change database, and no criterion in this SPEC
  closes that.** Version 0.1.0 said the execution "is AC-TLE-018's job in run-phase", which was a
  parking spot rather than a gap: AC-TLE-018 does not do that job. What AC-TLE-018 demonstrates,
  after the D7 strengthening, is that (a) the pre-change *statements* still succeed verbatim against
  a post-change database, and (b) the pre-change *open path* — the DDL const, the `schemaVersion`
  read, and the version switch **as they stand at HEAD `e50964ad3`** — reconstructed in-test, opens
  a post-change database without error. What it does **not** demonstrate is a genuinely older
  build: the reconstruction is compiled from today's source, so a divergence between the reconstruction
  and a real released binary would be invisible to it. REQ-TLE-018 therefore remains **an argued
  claim with a partial demonstration**, not a demonstrated one, and it stays in this list after
  AC-TLE-018 passes.
- **Forward drift of the frozen replica — a distinct hazard from the released-binary one, and the
  only one of the two that is now detected.** AC-TLE-018 clause (b) holds a copy of `backlogDDL`,
  the `schemaVersion` read, and the version switch frozen at HEAD `e50964ad3`. Nothing in the
  language stops the live `backlogDDL` from being edited later by a different card; when that
  happens the replica stops representing the current pre-change path, while the test keeps passing
  and keeps reporting that it exercises it. v0.2.0 named only the *backward* divergence (replica vs
  a real released binary) and was silent on this *forward* one. Clause (c) now asserts the frozen
  copy equals the live const, so forward drift fails loudly and becomes a deliberate act rather than
  a silent decay. The backward divergence remains undetected by construction and stays in this list.
- The migration parity path (`backlog_migrate.go:585-630`) was read, not exercised. Whether its
  comparison is reachable for every archived row was not measured here.
- No external consumer of `moai todo pr` was enumerated. Half A recorded that they cannot be
  grepped, and that remains true; this SPEC inherits the residual rather than closing it.
- SQLite version behaviour for `ALTER TABLE ADD COLUMN` was observed on this machine only, via the
  probe. Behaviour on the Windows and Linux CI matrices was not measured.

**Residual risk:**

- **A stored `spec_status` goes stale** (§B.4). Accepted with `observed_at` as the mitigation; a
  reader who ignores the timestamp can still misread an old record as current.
- **The seventh column is a machine-readable contract change.** A consumer doing `cut -f6` now gets
  the evidence where it previously got the queue state. This is the second such change on this
  surface — half A went five columns to six and recorded that external consumers cannot be
  enumerated — so the risk is inherited and compounded, not new. The card text staying last is the
  only mitigation available, and it protects only tail-readers.
- **An operator can assert a SHA that is reachable and still wrong.** REQ-TLE-020 closes the
  cheap half of this — a SHA that names no commit, or names one unreachable from the record's `ref`,
  is refused at write time and nothing is stored. What survives is the expensive half: a **typo that
  happens to land on another reachable commit** is stored, is marked `operator`, and is
  indistinguishable from a correct record by any check the machine can run — because telling them
  apart is exactly the card-to-commit attribution REQ-1.10 forbids the machine to attempt. Concretely,
  what such a typo costs: the record renders on `todo pr` as an operator-asserted delivering commit
  and is believed, and it is corrected only when a human notices and runs `--clear` or re-records
  (§B.5).

  **A mitigation exists and was considered — v0.2.0's "nothing flags it" overstated the absence and
  is withdrawn.** The machine cannot *detect* the error (that half stands), but it can make a human
  likely to: **render the named commit's own subject line beside the SHA**. A commit's subject is a
  fact about that commit, in the same non-attributing class as `ref_head` — it says nothing about
  which card the commit delivered — and it is exactly what lets an operator who typed one SHA and
  meant another see that the row describes the wrong change.

  It is **not built here**, for a reason that constrains how a follow-up must build it. Reading the
  subject at render time would put a new `git` subprocess on `todo pr`, which §D excludes and which
  half A's cost argument protects. The only shape compatible with that exclusion is to capture the
  subject **at record time**, alongside `ref_head` and `spec_status` (§B.4) — which adds a seventh
  fact to the record and changes REQ-TLE-016's cell format. That is a change to the record shape,
  i.e. a design decision, not a defect fix, so it is surfaced at the Implementation Kickoff Approval
  gate rather than folded into a debt-closing revision. Recorded here so the next reader neither
  re-derives it nor reaches for the render-time `git` call that §D forbids.
- **The guard sees column tuples, not the whole schema.** REQ-TLE-019's
  `(name, type, notnull, dflt_value)` assertion catches an added, removed, reordered, retyped, or
  re-nullabled column. It does not catch a changed CHECK expression beyond the substring the
  existing assertion already pins, a changed primary key, or a trigger. The axis §A.5 was written
  to close is closed; the rest of the schema surface is not, and no criterion here claims it is.
- **The doctrine prose has no criterion behind it** (§C.7). AC-TLE-021 verifies mirror parity and
  the stated column count; the sentence describing the verb, and the evidence-is-not-a-transition
  sentence, are Definition-of-Done items only. The behaviour those sentences describe is verified
  by AC-TLE-008; the wording is not.
- **This SPEC carries 21 of Tier L's 25 requirements** (up from 19 at v0.1.0: REQ-TLE-020 for the
  D2 validation, REQ-TLE-021 for the D3 mirror parity). Four of headroom remain. The axis is not
  infinitely extensible: a follow-up that adds observation history or a phase-aware predicate should
  be its own card rather than an amendment here.
