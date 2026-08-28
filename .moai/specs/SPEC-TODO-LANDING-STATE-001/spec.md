---
id: SPEC-TODO-LANDING-STATE-001
title: "A card that knows its own landing state — the integration-branch ref correction, a three-valued landed answer, and landing evidence on the queue record"
version: "0.1.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec (card t331)
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban, internal/cli, internal/config, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, cli, landing-state, integration-branch, git-flow, evidence, additive-schema"
tier: M
related_specs:
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-TODO-SQLITE-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC: A card that knows its own landing state

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-28 | Initial plan-phase authoring (card t331), measured at tree `3de2f85a2`. Root cause identified and measured: `LandedRef = "origin/main"` while the project integrates on `develop`. Scope boundary against t330 inherited verbatim from `SPEC-TODO-DESTRUCTIVE-GUARD-001` §B.2. One deviation from the dispatching lead's stated storage direction is recorded and justified in §B.2. |

> **Provenance discipline.** Every `file:line` citation in this document was measured at tree
> `3de2f85a2` (branch `WT-card-landing-state`, worktree `.claude/worktrees/t331`). A prior reader of
> this card cited a path from a stale tree and reported it missing, so the tree SHA travels with the
> citation rather than being assumed.

> **Named exemplar unavailable.** The dispatch named the t347 / t333 report split as the format
> exemplar for §A.6's state table. Neither `.moai/reports/t347` nor `.moai/reports/t333` exists at
> `3de2f85a2` (t333 is a card in flight; t347 is not started). The state table below is therefore
> **self-describing**: it carries its own row legend and its own column legend, and depends on no
> external format.

---

## §A Context

### A.1 What the operator observed

`picked` conflates two states that need different actions: *in progress* and *finished but not
closed*. On 2026-08-27 the kanban lead dispatched already-landed cards as new work **four times in
one day** (t293, t301, t310, t200), and separately closed 15 cards of which 8 had landed days
earlier.

The operator's card names the symptom precisely: `moai todo pr <id>` answers `no-link` for both a
never-started card and an already-landed one, so the surface that exists to prevent exactly this
misdispatch reports the two indistinguishably.

### A.2 The root cause is not the symptom

The card attributes the blindness to cards carrying no pull request under git-flow. That is a real
contributing condition, but it is **not** why `no-link` came back. The cause is that the landed
question is asked **about the wrong ref**.

`internal/kanban/prlink_landed.go:28` @ `3de2f85a2`:

```go
const LandedRef = "origin/main"
```

The project integrates on `develop` (`.moai/config/sections/git-strategy.yaml` @ `3de2f85a2`:
`worktree_base_branch: develop`, `develop_branch: develop`). Measured at `3de2f85a2`:

```
$ git rev-list --count --left-right origin/main...origin/develop
0	329
```

`main` is a strict ancestor of `develop` and lags it by 329 commits. Per-card, measured at
`3de2f85a2` with the check's own predicate (`git log <ref> --perl-regexp --grep='\b<id>\b' --oneline`):

| Card | commits on `origin/main` | commits on `origin/develop` | `todo pr` answers | truth |
|---|---|---|---|---|
| t293 | 0 | 9 | `no-link` | landed, sync-closed |
| t310 | 0 | 6 | `no-link` | landed, sync-closed |
| t322 | 0 | 5 | `no-link` | landed, sync-closed |
| t200 | 1 | 1 | `landed` | landed (promoted to `main`) |

The three misdispatched cards read `no-link` — *"nobody has started this"* — which is the dispatch
signal that sent the lead to dispatch them again. t200 read `landed` only because its landing
predates the `main`/`develop` divergence.

**So the guard did not fail to detect a landing. It asked a branch that had never seen one.** Every
card integrated on `develop` since the git-flow switch returns a FALSE not-landed, and the card's
`no-link` symptom is that false answer reaching the four-outcome renderer
(`internal/kanban/prlink.go:160-178` @ `3de2f85a2`: no PR hit + `Landed()` false ⇒ `no-link`).

Recording this attribution is load-bearing: a reader who fixes only the symptom would add PR-less
card handling and leave every develop-landed card still reading `no-link`.

### A.3 The second surface: ancestry is squash-blind

Where `todo pr` is unavailable or unconvincing, the lead falls back by hand to
`git merge-base --is-ancestor <card-branch-commit> <ref>`. That surface has an independent blindness.
Measured at `3de2f85a2`:

```
$ git merge-base --is-ancestor 7fc161b36 origin/develop ; echo $?
1
$ git merge-base --is-ancestor 294b4b6ab origin/develop ; echo $?
0
$ git log -1 --format='%h %s' 7fc161b36
7fc161b36 fix(doctor): make the exit code and the constitution row tell the truth (t200)
$ git log -1 --format='%h %s' 294b4b6ab
294b4b6ab fix(doctor): make the exit code and the constitution row tell the truth (t200) (#1612)
```

t200's worktree commit `7fc161b36` is **not** an ancestor of the integration branch; its content
landed as the squash commit `294b4b6ab` (PR #1612). Ancestry answers NOT-LANDED for work that
landed. The squash is the merge method this project configures
(`git_strategy.manual.merge_method: squash` @ `3de2f85a2`), so this is the normal path, not an edge.

Note the asymmetry between the two surfaces, because it decides the design: the **grep** predicate
survives the squash (the squash commit's message retained `(t200)`), while **ancestry** does not.
The grep survives only because this project requires the card id in every commit message
(`AGENTS.md` §3) — a discipline, not a mechanism. A landed answer must therefore rest on the
ref's own history, never on ancestry of a card-branch SHA.

### A.4 The third surface: `--require-landed` cannot say whether it ran

`internal/cli/todo.go:417-431` @ `3de2f85a2` — on an unanswerable query the guard emits a stderr
note and returns `nil`, so `done` proceeds and prints `done <id>` with exit 0. Measured at
`3de2f85a2`:

```
$ go test ./internal/cli/ -run 'TestTodoDone_RequireLanded' -v 2>&1 | tail -5
=== RUN   TestTodoDone_RequireLandedRefusesWhenNotLanded
--- PASS: TestTodoDone_RequireLandedRefusesWhenNotLanded (0.15s)
=== RUN   TestTodoDone_RequireLandedProceedsWhenInconclusive
--- PASS: TestTodoDone_RequireLandedProceedsWhenInconclusive (0.14s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.180s
```

`TestTodoDone_RequireLandedProceedsWhenInconclusive`
(`internal/cli/todo_undone_test.go:287-302` @ `3de2f85a2`) asserts exit 0 and a stdout prefix of
`done t1` on an unanswerable query. `TestTodoDone_NoLandingQueryWithoutTheFlag`
(`:326-331`, PASS at `3de2f85a2`) asserts exit 0 with a stub reporting landed. The two outcomes are
therefore **byte-identical on stdout and on the exit code**; they differ only in a stderr note that
no caller is obliged to read.

That is card t330's F6, handed to this card by lane-5: *"the guard passed it"* and *"the guard did
not run"* are mechanically indistinguishable.

The unanswerable case is not hypothetical here. It is what a misconfigured or unfetched ref produces:

```
$ git log origin/no-such-ref --perl-regexp --grep='\bt331\b' --oneline ; echo rc=$?
fatal: ambiguous argument 'origin/no-such-ref': unknown revision or path not in the working tree.
rc=128
```

### A.5 What this card inherits, and does not re-open

`SPEC-TODO-DESTRUCTIVE-GUARD-001` §B.2 @ `3de2f85a2` states the boundary explicitly, and this SPEC
adopts it verbatim rather than restating it in its own words:

| Owned by t330 (landed) | Owned by t331 (this SPEC) |
|---|---|
| the archive storage and `undone` | the persisted landing-state field on the card |
| `--expect` guard on `done` | the phase-aware predicate that reads it |
| the landing-predicate **seam** + opt-in `--require-landed` | flipping the check to default-on, if warranted |
| documenting the predicate's limit | replacing the predicate with one that has no such limit |

t330 §D also records the ref correction as explicitly out of ITS scope — "a separate decision with
its own blast radius". That decision is made here (§B.1).

One inherited finding constrains this SPEC and must be carried forward, because it is the reason the
ref correction alone is not the whole answer. t330 §A.4 @ `3de2f85a2` measured that with the ref
corrected the grep predicate is satisfied by a card's **plan-phase** commit — the earliest of t306's
13 develop commits is `3030df58b`, a plan-phase artifacts commit — so a ref-corrected predicate is
structurally always-true for any card that has reached the integration branch at all. Correcting the
ref therefore fixes the *false negative* (§A.2) and leaves the *false positive* untouched. Both are
in scope here; §B.3's evidence record is the answer to the second.

### A.6 The state model — the states, and what each surface says about them

**Row legend.** Each row is a state a card can actually occupy, distinguished by facts that are
separately observable: its queue `state` value, whether commits exist on its branch, whether the
integration branch names it, and whether the card's work is complete.

**Column legend.**

- **S1 — `moai todo list`**: the queue's own `state` column (`queued` / `picked` / `dropped`;
  `internal/kanban/backlog_store.go:52-60` @ `3de2f85a2`).
- **S2 — `moai todo pr <id>`**: the four-outcome resolver (`internal/kanban/prlink.go` @
  `3de2f85a2`), whose landed leg greps `LandedRef`.
- **S3 — hand ancestry**: `git merge-base --is-ancestor <card-branch-commit> <ref>`, the lead's
  manual fallback. Not a shipped surface; included because it is the one actually used.
- **S4 — `moai todo done --require-landed`**: the opt-in guard (`internal/cli/todo.go:417` @
  `3de2f85a2`).

#### Table 1 — what each surface reports TODAY (measured at `3de2f85a2`)

| # | State | S1 | S2 `todo pr` | S3 ancestry | S4 `--require-landed` | Correct? |
|---|---|---|---|---|---|---|
| C1 | queued, never started | `queued` | `no-link` | n/a (no commit) | refuses | yes |
| C2 | picked, work under way, **no commits yet** (t338 shape) | `picked` | `no-link` | n/a (no commit) | refuses | **no** — reads identically to C1 |
| C3 | picked, commits on the card branch, unmerged | `picked` | `no-link`, or `linked` where an open PR carries it | NOT-ANCESTOR | refuses | partly — correct only when a PR exists |
| C4 | picked, **run** landed on the integration branch, sync not | `picked` | `no-link` (ref lags) | ANCESTOR (merge) / NOT-ANCESTOR (squash) | refuses | **no** |
| C5 | picked, **fully landed** on the integration branch, not closed (t293, t310, t322) | `picked` | **`no-link`** | ANCESTOR or NOT-ANCESTOR | refuses | **no — this is the misdispatch** |
| C6 | picked, landed **via squash**, card-branch SHA not an ancestor (t200) | `picked` | `landed` iff the ref names the card | **NOT-ANCESTOR** | passes iff named | **no** on S3 |
| C7 | picked, landed and promoted to `main` (t241, t278) | `picked` | `landed` | ANCESTOR | passes | yes (by luck of the ref) |
| C8 | dropped | `dropped` | as its commit history dictates | — | refuses / passes | yes |
| C9 | archived (`done`) | invisible | invisible | — | — | yes |
| C10 | any state, landed ref **does not resolve** | as stored | `no-link` + stderr note | error | **proceeds, exit 0, `done <id>`** | **no — F6** |

#### Table 2 — what each surface MUST report after this SPEC

| # | State | S1 | S2 `todo pr` | S3 ancestry | S4 `--require-landed` |
|---|---|---|---|---|---|
| C1 | queued, never started | `queued` | `no-link` | *retired as a decision input* | refuses; stdout `landing=not-landed` |
| C2 | picked, no commits yet | `picked` | `no-link` **+ `picked` is visible in the same row** | " | refuses; stdout `landing=not-landed` |
| C3 | picked, unmerged commits | `picked` | `linked` where a PR exists, else `no-link` | " | refuses; stdout `landing=not-landed` |
| C4 | picked, run landed, sync not | `picked` | `landed` **with the observed commit named**, and the card's live SPEC status shown as `in-progress` | " | passes; stdout `landing=landed` |
| C5 | picked, fully landed, not closed | `picked` | **`landed`** with the observed commit named, SPEC status `implemented`/`completed` | " | passes; stdout `landing=landed` |
| C6 | landed via squash | `picked` | `landed` — the answer rests on the ref's history, never on ancestry | " | passes; stdout `landing=landed` |
| C7 | landed and promoted | `picked` | `landed` | " | passes; stdout `landing=landed` |
| C8 | dropped | `dropped` | unchanged | " | unchanged |
| C9 | archived | invisible | invisible | " | — |
| C10 | landed ref does not resolve | as stored | **`unknown`**, distinct from `no-link`, naming the unresolved ref | " | **proceeds** (policy unchanged) but stdout reads `landing=unknown` |

Three things the tables make visible that prose does not:

1. **C1 and C5 collapse onto the same cell today** (`queued`/`picked` differ, but the landed-bearing
   column reads `no-link` for both). That collapse is the misdispatch.
2. **C4 and C5 stay distinct only through evidence, never through a boolean.** A single landed
   boolean cannot separate "run landed" from "everything landed" — which is t330 §A.4's finding, and
   the reason §B.3 records observations rather than a flag.
3. **C10 is the row with no honest cell today.** It reads exactly like a card that was checked.

---

## §B Decisions

### B.1 Decision 1 — the landed ref is resolved from configuration, not compiled in

Ruled: `LandedRef` stops being a `const` and becomes a **resolved** value, taken from the project's
configured integration branch, with `origin/main` retained as the fallback when nothing is
configured.

Three measurements support the shape:

**M1 — the resolver already exists.** `config.LoadWorktreeBaseBranch(projectRoot)`
(`internal/config/loader_worktree_base.go:28-35` @ `3de2f85a2`) returns the configured base branch,
and `hook.WorktreeBaseBranchResolvable(branch)` (`internal/hook/worktree_base_branch.go:70-82` @
`3de2f85a2`) reports whether that branch actually resolves. Both landed with
SPEC-WORKTREE-BASEREF-001 (card t313). Nothing new is invented; a second consumer is added to a
primitive that already has three (`internal/cli/session_worktree.go:215`,
`internal/cli/doctor_worktree_base.go:43`, `internal/hook/worktree_base_branch.go:156` @
`3de2f85a2`).

**M2 — the empty-value semantics are already ruled and must be preserved.**
`internal/config/types.go:164-174` @ `3de2f85a2` states that the empty string is the neutral
default meaning "take no action", and that the shipped template ships it empty because naming a
branch there would not be neutral across downstream projects. This SPEC inherits that ruling
exactly: **empty ⇒ `origin/main`**, byte-identical to today's behaviour, so a downstream project
that configures nothing sees no change at all.

**M3 — the blast radius is four call sites and is bounded.** Every consumer of the constant, measured
at `3de2f85a2` by `grep -rn 'LandedRef' --include='*.go' internal/`:

| Site | Use |
|---|---|
| `internal/kanban/prlink_landed.go:44` | the `git log` ref operand — the behavioural one |
| `internal/kanban/prlink_landed.go:78` | an error string |
| `internal/cli/todo_pr.go:75`, `:87` | user-facing help text |
| `internal/cli/todo.go:357`, `:399`, `:428` | user-facing help and refusal text |
| `internal/cli/todo_undone_test.go:277-278` | a test asserting the refusal names the ref |

Six production sites and one test. t330 §D deferred this correction on blast-radius grounds; the
measured radius is small and entirely inside two packages.

**Why not simply re-point the constant at `origin/develop`.** It would fix this repository and break
every downstream one, because the constant ships in the binary while the integration branch is a
per-project fact. The neutral-default rule in M2 exists for exactly this reason.

### B.2 Decision 2 — landing evidence is a new top-level array, not a per-item column

Ruled: landing observations are stored as a **new top-level array on the record and a new SQLite
table**, keyed by card id. The five-field per-item contract is untouched.

**This deviates from the dispatching lead's stated direction and the deviation is deliberate.** The
dispatch said "evidence columns added with `ALTER TABLE ADD COLUMN` are the sanctioned direction".
The deviation rests on a constraint the dispatch did not cite and on a capability an `ALTER TABLE`
cannot supply:

**M1 — per-item fields are frozen.** `internal/kanban/backlog_store.go:42-46` and `:62-64` @
`3de2f85a2`:

```go
// backlogVersion is the record schema version. The schema is ADDITIVE within
// version 1: `last_seq` was appended as a top-level field with no version
// bump, and no per-item field may ever be added (spec.md §E out-of-scope).
...
// BacklogItem is one queued card. The five fields are the frozen per-item
// contract (REQ-TODO-013)
```

An `ALTER TABLE items ADD COLUMN landed_sha` is cheap in SQLite and expensive in the record contract:
the SQLite row and `BacklogItem` are two faces of one schema, and REQ-TODO-013 freezes the second.

**M2 — a new table is free on every existing database, exactly as `archived_*` was.**
`internal/kanban/backlog_sqlite.go:96-100` @ `3de2f85a2` records the precedent and its reason in the
same breath as the CHECK constraint the lead's out-of-scope ruling cites:

```
// The two archived_* tables are the reversal storage ... They are ADDITIVE and cost
// nothing on an existing database: this whole DDL runs on every open and every
// statement is IF NOT EXISTS ... which is precisely why the archive is a pair of
// tables rather than a fourth `state` value. SQLite cannot ALTER a CHECK constraint,
// so admitting a fourth state would need a table rebuild on every operator queue in
// the field.
```

The archive is the third application of a precedent (`last_seq`, `findings`, `archived`); landing
observations are the fourth. It needs no migration code and no `schema_version` bump.

**M3 — a column cannot hold what must be held.** A card has more than one landing: the run landing
and the sync landing are distinct events, and telling them apart is the whole content of §A.5's
inherited false-positive finding (C4 vs C5 in the Table 2). One column per card records one event
and silently overwrites the other; an observation list records both, in order, with the commit that
carried each.

The deviation therefore *strengthens* the lead's own out-of-scope ruling rather than contradicting
it: no `ALTER TABLE`, no CHECK change, no rebuild, and the frozen per-item contract preserved.

### B.3 Decision 3 — the recorded fact is an observation, and SPEC status is read live

Ruled: an observation records **what was seen on the ref**, and nothing derived. Concretely: the card
id, the resolved ref, the observed commit SHA **as it exists on that ref**, the commit subject, and
the observation instant.

Two consequences, both requirements rather than notes:

- **Never a card-branch SHA.** §A.3 measured that the card-branch commit and the landed commit are
  different objects under squash. Recording `7fc161b36` as t200's landing would record a SHA that is
  not on the integration branch at all — an evidence path that does not resolve, which
  `verification-claim-integrity.md` §2 calls an unattributed claim.
- **SPEC status is not copied into the queue.** `kanban.ReadCardStatus(primaryRoot, specID, bases)`
  (`internal/kanban/status_read.go:99` @ `3de2f85a2`) already reads a card's frontmatter `status`
  from the card's branch without a checkout, and tags the source it came from
  (`StatusSourceBranch` / `StatusSourcePrimary` / `StatusSourceUnresolved`, `:44-48` @ `3de2f85a2`).
  The SPEC frontmatter is the SSOT for status; copying it into the queue would create a second
  source that drifts silently. The card's request for "the card's SPEC id and status" is satisfied by
  `spec_id` (already a per-item field) plus a **live read**, not by a stored copy.

### B.4 Explicit NON-GOAL — a landing field is evidence, never a transition

**The tooling must not close, move, drop, re-order, or re-state a card on its own authority when it
detects a landing.** This is stated here in the SPEC body, not only as a design note, because it is
the most plausible slip for the next person reading this design: *"since we are detecting the landing
anyway, let the tool close the card."*

The reason it is forbidden is not stylistic. `.claude/skills/moai/workflows/todo.md:53-58` @
`3de2f85a2` is [HARD]:

> `edit`, `move`, `drop`, `undrop`, `done`, and `undone` are operator acts, exactly like `add` and the
> pick. Correct a card's wording, move it, or discard it because the operator said to — never on
> inferred priority, never as tidy-up, never to fold one card into another, and never because a card
> looks stale. **The queue records the operator's intent; it does not curate it.**

A landing observation is a *measurement*. Acting on it converts a measurement into a decision, and
the decision is the operator's. The failure mode this prevents is concrete and worse than the one
the card is about: an auto-close on a **run** landing (state C4) would close a card whose sync commit
is still unpushed in a lane's worktree — precisely the t306 incident that motivated t330, now
performed by the machine and at scale.

The mechanical form of the non-goal:

- `moai todo pr` remains **read-only** — it writes no field, no finding, no cache, no lock
  (`internal/cli/todo_pr.go:1-16` @ `3de2f85a2` states this as a property of the file). Detection
  does not become a write path.
- Recording an observation happens **only inside an operator-issued verb**, and that verb changes no
  card — the same shape `todo relate` already has ("The verb writes a record and touches neither
  card; `absorbs` does not absorb", `todo.md:47` @ `3de2f85a2`).
- `--require-landed` stays **opt-in**. Flipping it default-on is out of scope (§D) even though this
  SPEC repairs the ref: t330 §A.4's false-positive finding is not closed by the ref correction, and a
  default-on check that passes structurally is worse than no check.

### B.5 Decision 4 — the unanswerable case stays permissive, and stops being silent

Ruled: **the proceed-on-unanswerable policy is NOT reversed.** `todo.md:63-68` @ `3de2f85a2` states
it with its reason — "refusing on the absence of evidence would block every machine that cannot
answer" — and that reason still holds. F6 is not that the guard proceeds; F6 is that a caller cannot
tell that it proceeded.

The repair is therefore on the *reporting* axis, not the *policy* axis:

- the landed answer becomes **three-valued** (`landed` / `not-landed` / `unknown`) rather than a
  `(bool, error)` pair whose error path a caller may drop;
- `moai todo done` emits a landing verdict token on **stdout** in every invocation, so the outcomes
  differ in the channel a caller actually reads.

If a later card wants to reverse the policy, it does so against a surface that can already express
the three answers. This SPEC does not make that decision, and §D records it as excluded.

---

## §C Requirements

### C.1 The landed ref

- **REQ-TLS-001** — The landed check shall ask its question about the repository's configured
  integration branch.
- **REQ-TLS-002** — **Where** `git_strategy.worktree_base_branch` carries a non-empty value, the
  landed check shall resolve its ref from that value; **where** it is empty, the landed check shall
  use `origin/main`, byte-identically to the behaviour at `3de2f85a2`.
- **REQ-TLS-003** — Every surface that names the landed ref in user-facing text shall name the
  **resolved** ref, never a compile-time constant.
- **REQ-TLS-004** — The landed check shall not require the resolved ref to be reachable from the
  session's own working tree; it shall read the ref as the repository holds it.

### C.2 The three-valued answer

- **REQ-TLS-005** — The landed check shall return one of exactly three answers: `landed`,
  `not-landed`, `unknown`.
- **REQ-TLS-006** — **When** the resolved landed ref does not resolve in the repository, the landed
  check shall answer `unknown` and shall not answer `not-landed`.
- **REQ-TLS-007** — No consumer shall infer `unknown` from a zero value, an empty string, or a
  dropped error; the answer shall be carried in a field a consumer must read.
- **REQ-TLS-008** — **Where** the landed answer is `unknown`, `moai todo pr` shall render an
  `unknown` outcome distinct from `no-link`, naming the ref that did not resolve.

### C.3 Distinguishability on the channel callers read

- **REQ-TLS-009** — `moai todo done` shall emit exactly one landing-verdict token on **stdout** in
  every invocation, including invocations made without `--require-landed`.
- **REQ-TLS-010** — **While** `--require-landed` is set and the landed answer is `unknown`,
  `moai todo done` shall archive the card and shall emit a stdout landing-verdict token distinct from
  the token it emits when the answer is `landed`.
- **REQ-TLS-011** — The landing verdict shall not be carried on stderr alone.
- **REQ-TLS-012** — The guard shall not refuse on an `unknown` answer; the permissive policy at
  `todo.md:63-68` @ `3de2f85a2` is preserved unchanged.

### C.4 Landing evidence on the queue record

- **REQ-TLS-013** — The queue record shall carry landing observations as a **new top-level array**;
  the five-field per-item contract (REQ-TODO-013) shall remain untouched.
- **REQ-TLS-014** — A landing observation shall carry: the card id, the resolved ref, the observed
  commit SHA, the observed commit subject, and the observation instant.
- **REQ-TLS-015** — The observed commit SHA shall be a commit that exists **on the resolved ref**; a
  card-branch commit SHA shall never be recorded as a landing SHA.
- **REQ-TLS-016** — **Where** a card carries more than one landing observation, all shall be retained
  in observation order; a later observation shall not overwrite or remove an earlier one.
- **REQ-TLS-017** — The storage shall admit landing observations through `CREATE TABLE IF NOT EXISTS`
  only: no `ALTER TABLE`, no CHECK-constraint change, no `schema_version` bump, no table rebuild.
- **REQ-TLS-018** — A binary that predates landing observations shall continue to open and mutate a
  queue that carries them, and shall leave them intact where it rewrites the record.

### C.5 The card's SPEC status

- **REQ-TLS-019** — The card's SPEC status shall be read live through `kanban.ReadCardStatus`; it
  shall not be copied into the queue record.
- **REQ-TLS-020** — **Where** the live SPEC status cannot be resolved, the surface shall render it as
  unresolved rather than as any status value.

### C.6 Authority

- **REQ-TLS-021** — Detecting or recording a landing shall not change a card's `state`, position,
  text, or `spec_id`.
- **REQ-TLS-022** — The tooling shall not close, archive, drop, re-order, or re-state a card on its
  own authority under any circumstance, including a fully-landed card whose SPEC reads `completed`.
- **REQ-TLS-023** — A landing observation shall be written only inside an operator-issued verb.
- **REQ-TLS-024** — `moai todo pr` shall remain read-only: no field, no finding, no cache, no lock,
  no landing observation.

### C.7 Doctrine

- **REQ-TLS-025** — `.claude/skills/moai/workflows/todo.md` and its template mirror
  (`internal/template/templates/.claude/skills/moai/workflows/todo.md`) shall be updated in the same
  change, and the [HARD] operator-only mutation rule shall be restated unweakened.
- **REQ-TLS-026** — The doctrine shall state the landed check's remaining limit: a `landed` answer
  reports that the ref's history names the card, not that the card's last step landed.

---

## §D Exclusions

This SPEC deliberately does not build the following. Each is out of scope for a stated reason, not by
oversight.

### Out of Scope — queue schema normalization

- `tier` and `depends_on` field normalization on the card record. The dispatching lead excluded it,
  and the reason is recorded here so it is not re-litigated: its stated precondition is a revision of
  the `todo.md` [HARD] rules, which is a doctrine change and is not open. Low schema cost is not by
  itself a reason to do it now; when the precondition is met, a follow-up card pays the same additive
  cost this SPEC would have paid.
- Any fourth `state` value, including `landed`. `internal/kanban/backlog_sqlite.go:96-100` and
  `:105-112` @ `3de2f85a2` constrain `items.state` with
  `CHECK (state IN ('queued','picked','dropped'))` and explain in the comment above the DDL that
  SQLite cannot ALTER a CHECK constraint, so a fourth value would force a table rebuild on every
  operator queue in the field.
- Any new per-item field (frozen by REQ-TODO-013, `internal/kanban/backlog_store.go:42-46` @
  `3de2f85a2`).
- Any `schema_version` bump or migration of existing operator queues.

### Out of Scope — automatic state transitions

- Any automatic close, archive, drop, re-order, or re-state on a detected landing. This is §B.4's
  explicit non-goal and is restated here so a scope reader who reads only §D still meets it.
- Any change to **who** may issue a queue-mutating verb. The [HARD] operator-only doctrine stands
  unchanged.
- Turning `moai todo pr` into a write path.

### Out of Scope — the guard's policy

- Reversing the proceed-on-unanswerable policy. §B.5 preserves it; a later card may revisit it
  against the three-valued surface this SPEC delivers.
- Flipping `--require-landed` to default-on. t330 §A.4's false-positive finding is not closed by the
  ref correction, and that decision belongs with a predicate that can distinguish a run landing from
  a sync landing.
- Making the predicate phase-aware (telling a run landing from a sync landing) **automatically**.
  This SPEC records the observations that make the distinction *readable*; deriving a phase verdict
  from them is a further step and is not taken here.

### Out of Scope — adjacent surfaces

- The guard's **manifestation history** — how many cards were misjudged, when, and what each
  misjudgement cost. That is card t347's scope. This SPEC uses four cards as test inputs and claims
  nothing about the population.
- Retention, pruning, or compaction of landing observations. They grow without bound in this SPEC; a
  retention policy is a separate decision requiring operator input.
- Any change to the `gh` query budget on `todo pr` (one query per invocation,
  `internal/cli/todo_pr.go:8-13` @ `3de2f85a2`), or any new default-on network cost on `todo list`,
  `todo next`, or any other read path.
- `moai doctor`'s worktree-base-branch check (`internal/cli/doctor_worktree_base.go:40-43` @
  `3de2f85a2`). It already reports an unresolvable configured branch; this SPEC consumes the same
  resolver but adds no diagnostic surface.

---

## §E Traceability

| Requirement | Acceptance criteria |
|---|---|
| REQ-TLS-001, REQ-TLS-002 | AC-TLS-001, AC-TLS-002 |
| REQ-TLS-003 | AC-TLS-003 |
| REQ-TLS-004 | AC-TLS-002 |
| REQ-TLS-005, REQ-TLS-006, REQ-TLS-007 | AC-TLS-004 |
| REQ-TLS-008 | AC-TLS-005 |
| REQ-TLS-009, REQ-TLS-010, REQ-TLS-011 | AC-TLS-006 |
| REQ-TLS-012 | AC-TLS-007 |
| REQ-TLS-013, REQ-TLS-014, REQ-TLS-016 | AC-TLS-008 |
| REQ-TLS-015 | AC-TLS-009 |
| REQ-TLS-017, REQ-TLS-018 | AC-TLS-010 |
| REQ-TLS-019, REQ-TLS-020 | AC-TLS-011 |
| REQ-TLS-021, REQ-TLS-023 | AC-TLS-012 |
| REQ-TLS-022 | AC-TLS-013 |
| REQ-TLS-024 | AC-TLS-014 |
| REQ-TLS-025, REQ-TLS-026 | AC-TLS-015 |

---

## §F Cross-references

- `SPEC-TODO-DESTRUCTIVE-GUARD-001` — the archive, `undone`, `--expect`, and the `--require-landed`
  seam this SPEC consumes; its §B.2 boundary table is inherited verbatim in §A.5.
- `SPEC-KANBAN-QUEUE-PR-SYNC-001` — the resolver and the landed primitive being corrected.
- `SPEC-TODO-SQLITE-001` — the storage engine and the additive-table precedent.
- `SPEC-WORKTREE-BASEREF-001` (card t313) — `worktree_base_branch`, its neutral-empty ruling, and the
  resolver reused in §B.1.
- `.claude/skills/moai/workflows/todo.md` — the [HARD] operator-only mutation doctrine and the
  `--require-landed` limit note.
- `.claude/rules/moai/core/verification-claim-integrity.md` — why a recorded SHA that does not
  resolve on the named ref is an unattributed claim.
