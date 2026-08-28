---
id: SPEC-TODO-LANDING-STATE-001
title: "A card that knows its own landing state, half A — the integration-branch ref correction, a three-valued landed answer, and a guard that says whether it ran"
version: "0.3.0"
status: draft
created: 2026-08-28
updated: 2026-08-29
author: manager-spec (card t331)
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban, internal/cli, internal/config, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, cli, landing-state, integration-branch, git-flow, three-valued-answer"
tier: M
related_specs:
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC: A card that knows its own landing state — half A, the discriminator

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-28 | Initial plan-phase authoring (card t331), measured at tree `3de2f85a2`. Root cause identified and measured: `LandedRef = "origin/main"` while the project integrates on `develop`. Scope boundary against t330 inherited verbatim from `SPEC-TODO-DESTRUCTIVE-GUARD-001` §B.2. One deviation from the dispatching lead's stated storage direction is recorded and justified in §B.2. |
| 0.2.0 | 2026-08-28 | **Scope split by operator ruling** (plan-audit iteration 1, FAIL 0.74). The SPEC carried two SPECs in one document — a discriminator and an evidence store — which is what put it at 26 requirements, above even the Tier L ceiling of 25. The evidence half (landing observations, their storage, the recording verb, an observed commit on `todo pr`, the live SPEC-status read) moved to **card t359**, which depends on this one; §B.2 became a pointer and §D records the whole axis as out of scope. What remains is half A: the ref correction, the three-valued answer, the stdout verdict, and the queue `state` on the read surface — **11 requirements, Tier M (ceiling 16)**. Also in this revision: AC-TLS-008 rebuilt from a name-based grep sweep, which a mutant satisfied, into a behavioural whole-queue byte-identity assertion, with its RED **observed by planting that mutant** and then reverted; seven `file:line` citations re-measured and corrected and every remaining citation re-opened at its address; `origin/develop` figures re-expressed as re-runnable commands carrying the ref's own SHA and an observation instant, because a tree SHA does not pin a moving ref; the three plan-phase clarification markers resolved or retired with the B half; and the "six production sites" count corrected to seven against the grep printed beside it. |
| 0.3.0 | 2026-08-29 | **Scoped delta fix after plan-audit iteration 2** (PASS 0.85; the Tier M iteration ceiling is exhausted, so the four blocking defects route as a delta fix before Implementation Kickoff Approval rather than as a third audit round). D1: AC-TLS-008's assertion is widened from the queue directory to the whole project root, because the auditor planted a cache written **outside** `StateDirForRoot` and the criterion stayed GREEN; the widened form's RED was re-established here by planting that same mutant, observing 4/4 FAIL, and reverting (`acceptance.md` §D.1.a). D2: Table 2's S3 column said hand ancestry was "retired as a decision input" while no requirement delivered and no criterion verified that retirement — the cells now read *unchanged from Table 1*, the C4-C7 collapse is re-argued on the shipped surfaces (S1/S2/S4) alone, and §D records the non-retirement as out of scope. D3: `REQ-TLS-011` restated a limit the doctrine already carries at `todo.md:59-67`, so its criterion could not fail on it; the requirement is sharpened to the two clauses genuinely absent — the limit at the `todo pr` outcome list, and the meaning of `unknown` — and `acceptance.md` §D.2 records the gap as three commands with their verbatim output rather than as a conclusion. D4: the `LandedRef` line-class breakdown summed to 11 against a stated 12; the twelfth is named as a substring false positive at `todo_undone_test.go:266` and the word-boundary form (`grep -rnw` → 11) is stated beside it. Optional D5 (the stdout-token ruling's over-claimed source) and D6 (the five-to-six column contract change) are both addressed in `plan.md`. No source file was modified: `internal/cli/todo_pr.go` reads SHA-1 `a80ca6bdf8cd61f278310befdc400e547aa00d04` before and after the mutant probe. |

> **Provenance discipline.** Every `file:line` citation in this document was measured at tree
> `3de2f85a2` (branch `WT-card-landing-state`, worktree `.claude/worktrees/t331`) and re-opened at
> its address during the 0.2.0 remediation at HEAD `11426a128` and again during the 0.3.0 delta fix
> at HEAD `45cff0f59`. The pins remain valid across all three trees because no cited source file
> changed between them — `git diff --name-only 3de2f85a2 HEAD` at `45cff0f59` returns exactly six
> paths: this SPEC's own four artifacts plus `.moai/reports/t331/plan-audit.md` and
> `plan-audit-iter2.md`. A prior reader of this card cited a path from a stale tree and reported it
> missing, so the tree SHA travels with the citation rather than being assumed.

> **A tree SHA does not pin a moving ref.** Citations into files are pinned to a tree; measurements
> **about** `origin/main` or `origin/develop` are not, because those refs advance independently of
> any tree. Every such figure below therefore travels with the command that produced it, the two
> refs' own SHAs, and the instant it was observed — and is **re-measured** rather than re-cited by a
> later reader (`verification-completeness.md` §4). The `0 329` divergence recorded at 0.1.0 read
> `0 334` at plan-audit and `0 349` at 0.2.0; the direction never changed, only the number.

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
`git_strategy.worktree_base_branch: develop` at `:5`, `git_strategy.manual.develop_branch: develop`
at `:16`).

The figures below are **about the refs, not about the tree**, so each is stated as the command that
produced it. Re-run them rather than re-citing them.

```
$ git rev-parse origin/main origin/develop
48239c7dc7428c8751a04f6321887c2d36123884
c6aa613463e6234155f45ce76666e985a42cd80c
$ git rev-list --count --left-right origin/main...origin/develop
0	349
$ git merge-base --is-ancestor origin/main origin/develop ; echo rc=$?
rc=0
```

**Observed at 2026-08-28T13:15Z**, `origin/main` = `48239c7dc`, `origin/develop` = `c6aa61346`.
`main` is a strict ancestor of `develop` and lags it, and the lag grows: the same command read
`0 329` when this SPEC was first authored. What is load-bearing is the **left column being zero**,
not the right column's value.

Per-card, same instant and same two ref SHAs, using the check's own predicate
(`git log <ref> --perl-regexp --grep='\b<id>\b' --oneline | wc -l`):

| Card | commits on `origin/main` | commits on `origin/develop` | `todo pr` answers | truth |
|---|---|---|---|---|
| t293 | 0 | 9 | `no-link` | landed, sync-closed |
| t310 | 0 | 6 | `no-link` | landed, sync-closed |
| t322 | 0 | 24 | `no-link` | landed, sync-closed |
| t200 | 1 | 1 | `landed` | landed (promoted to `main`) |

t322's develop count read `5` at 0.1.0 and `24` now — the count moved, the zero did not. Only the
zero is the claim.

The three misdispatched cards read `no-link` — *"nobody has started this"* — which is the dispatch
signal that sent the lead to dispatch them again. t200 read `landed` only because its landing
predates the `main`/`develop` divergence.

**So the guard did not fail to detect a landing. It asked a branch that had never seen one.** Every
card integrated on `develop` since the git-flow switch returns a FALSE not-landed, and the card's
`no-link` symptom is that false answer reaching the four-outcome renderer
(`internal/kanban/prlink.go:175-185` @ `3de2f85a2` — the landed leg: no PR hit, `landed.Landed()` at
`:178`, and the `if isLanded` promotion at `:182` never taken ⇒ the outcome stays `no-link`).

Recording this attribution is load-bearing: a reader who fixes only the symptom would add PR-less
card handling and leave every develop-landed card still reading `no-link`.

### A.3 The second surface: ancestry is squash-blind

Where `todo pr` is unavailable or unconvincing, the lead falls back by hand to
`git merge-base --is-ancestor <card-branch-commit> <ref>`. That surface has an independent blindness.
Measured against `origin/develop` = `c6aa61346` at 2026-08-28T13:15Z (a ref probe, so re-run rather
than re-cite):

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
note and returns `nil` (`:420-423`), and on a satisfied query it returns `nil` (`:430`). `done`
proceeds identically on both paths and prints `done <id>` with exit 0.

**The code, not the tests, is what establishes stdout identity.** `todoRequireLanded` has exactly
two `return nil` paths and writes nothing to stdout on either, so the caller's printed line is the
same bytes in both cases. The tests below establish the **exit code**, and one of them is commonly
misread as establishing more than it does:

```
$ go test ./internal/cli/ -count=1 -run 'TestTodoDone_RequireLanded|TestTodoDone_NoLandingQueryWithoutTheFlag' -v 2>&1 | tail -8
=== RUN   TestTodoDone_RequireLandedRefusesWhenNotLanded
--- PASS: TestTodoDone_RequireLandedRefusesWhenNotLanded (0.13s)
=== RUN   TestTodoDone_RequireLandedProceedsWhenInconclusive
--- PASS: TestTodoDone_RequireLandedProceedsWhenInconclusive (0.13s)
=== RUN   TestTodoDone_NoLandingQueryWithoutTheFlag
--- PASS: TestTodoDone_NoLandingQueryWithoutTheFlag (0.20s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.395s
```

(One invocation, three tests. The `-run 'TestTodoDone_RequireLanded'` regex used at 0.1.0 excludes
the third test, so a PASS pair drawn from it could not have included
`TestTodoDone_NoLandingQueryWithoutTheFlag`; the regex above selects all three.)

`TestTodoDone_RequireLandedProceedsWhenInconclusive`
(`internal/cli/todo_undone_test.go:287-303` @ `3de2f85a2`) asserts exit 0 and a stdout prefix of
`done t1` on an unanswerable query (`:297`). `TestTodoDone_NoLandingQueryWithoutTheFlag`
(`:306-335`; the positive control at `:326-331`) asserts exit 0 and a subprocess count with a stub
reporting landed — it **never inspects stdout**. So the two outcomes are **byte-identical on stdout
and on the exit code**, with the stdout half resting on `todo.go:417-431` and the exit-code half on
the two tests; they differ only in a stderr note that no caller is obliged to read.

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

| Owned by t330 (landed) | Owned by the t331 axis |
|---|---|
| the archive storage and `undone` | the persisted landing-state field on the card |
| `--expect` guard on `done` | the phase-aware predicate that reads it |
| the landing-predicate **seam** + opt-in `--require-landed` | flipping the check to default-on, if warranted |
| documenting the predicate's limit | replacing the predicate with one that has no such limit |

t330 §D also records the ref correction as explicitly out of ITS scope — "a separate decision with
its own blast radius". That decision is made here (§B.1).

**This SPEC is the first half of that axis, and only the first half.** The right-hand column above
was originally one card; it is now split in two, and the split is the reason this document carries no
storage design:

| Half | Owner | Scope |
|---|---|---|
| **A — the discriminator** | **this SPEC** | the landed ref is resolved from configuration rather than compiled in; the landed answer becomes three-valued; `moai todo done` says on stdout which answer it got, so "the guard passed" and "the guard did not run" stop reading alike; the queue `state` becomes visible beside the link outcome |
| **B — the evidence** | **card t359** | landing observations persisted on the queue record, the storage shape that holds them, the recording verb, rendering an observed commit on `todo pr`, and the live SPEC-status read that separates a run landing from a sync landing |

B depends on A landing first, and the dependency is not schedule convenience: an evidence record
written from a predicate that asks the wrong ref stores wrong evidence, permanently, and a stored
observation is far more expensive to correct than a live answer. Getting the discriminator right is
the precondition for storing anything at all. Everything in the B row is **out of scope here** and is
listed as such in §D.

One inherited finding constrains both halves and belongs on the record here, because it is the reason
A alone is not the whole answer. t330 §A.4 @ `3de2f85a2` measured that with the ref corrected the
grep predicate is satisfied by a card's **plan-phase** commit — the earliest of t306's 13 develop
commits is `3030df58b`, a plan-phase artifacts commit — so a ref-corrected predicate is structurally
always-true for any card that has reached the integration branch at all. Correcting the ref therefore
fixes the *false negative* (§A.2) and leaves the *false positive* untouched. **A fixes the false
negative and states the residual limit in the doctrine (REQ-TLS-011); closing the false positive is t359's
work**, and this SPEC deliberately does not claim it.

### A.6 The state model — the states, and what each surface says about them

**Row legend.** Each row is a state a card can actually occupy, distinguished by facts that are
separately observable: its queue `state` value, whether commits exist on its branch, whether the
integration branch names it, and whether the card's work is complete.

**Column legend.**

- **S1 — `moai todo list`**: the queue's own `state` column (`queued` / `picked` / `dropped`;
  `internal/kanban/backlog_store.go:53-60` @ `3de2f85a2`).
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
| C11 | **queued**, landed — a duplicate card filed for work that already shipped | `queued` | `no-link` (ref lags) | ANCESTOR or NOT-ANCESTOR | refuses | **no** — reads identically to C1 |

C11 is reachable, not hypothetical: `todo add` refuses only a byte-identical duplicate, so a lead
filing new work for something already done lands here — the card's own originating incident. It is
the same S2 defect as C5 seen from a `queued` card; the corrected S2 answers it correctly, so the
design needed no addition, only the enumeration did.

#### Table 2 — what each surface MUST report after this SPEC

**Scope of this table: the three shipped surfaces.** S1, S2, and S4 are surfaces this SPEC changes or
preserves, and every cell below is a claim about them. **S3 is not a shipped surface** (see the column
legend) and **this SPEC delivers no change to it**: no requirement in §C names ancestry, no
acceptance criterion asserts anything about it, and nothing here retires it, deprecates it, or
removes it from anyone's hand. Its cells therefore read *unchanged from Table 1* — an honest record
of what this SPEC leaves alone, not a claim about a change it does not make. (An earlier draft wrote
"*retired as a decision input*" in every S3 cell. Nothing delivered that retirement, so the wording
asserted an unimplemented column; it is corrected here rather than backed by a new requirement,
because retiring a lead's manual fallback is a doctrine decision this SPEC has no mandate for.)

| # | State | S1 | S2 `todo pr` | S3 ancestry (not shipped; unchanged) | S4 `--require-landed` |
|---|---|---|---|---|---|
| C1 | queued, never started | `queued` | `no-link` | as Table 1 | refuses; stdout `landing=not-landed` |
| C2 | picked, no commits yet | `picked` | `no-link` **+ `picked` is visible in the same row** | as Table 1 | refuses; stdout `landing=not-landed` |
| C3 | picked, unmerged commits | `picked` | `linked` where a PR exists, else `no-link` | as Table 1 | refuses; stdout `landing=not-landed` |
| **C4-C7** | picked and landed, by any route — run-only, fully, via squash, or promoted to `main` | `picked` | **`landed`** — one answer, resting on the resolved ref's history and never on ancestry | as Table 1 (still ANCESTOR for C4/C7, NOT-ANCESTOR for C6, either for C5 — see the collapse note) | passes; stdout `landing=landed` |
| C8 | dropped | `dropped` | unchanged | as Table 1 | unchanged |
| C9 | archived | invisible | invisible | as Table 1 | — |
| C10 | landed ref does not resolve | as stored | **`unknown`**, distinct from `no-link`, naming the unresolved ref | as Table 1 | **proceeds** (policy unchanged) but stdout reads `landing=unknown` |
| C11 | queued, landed | `queued` | `landed`, beside the queue state `queued` | as Table 1 | passes; stdout `landing=landed` |

**C4 through C7 are one row here, and the collapse is argued on the shipped surfaces alone.** In
Table 1 they are four distinct rows because today they genuinely answer differently on S2 and S4 —
C4 and C5 read `no-link`, C6 reads `landed` only if the ref happens to name the card, C7 reads
`landed` by luck of `main`. After this SPEC all four answer `landed` through the same mechanism, and
their S1 and S4 cells are likewise identical, so on **S1, S2, and S4 together** the four are no
longer separately observable and keeping them apart would assert a distinction the shipped tooling
does not make.

**What the collapse does not claim.** S3 still separates them — C6's squash landing is
NOT-ANCESTOR where C7's is ANCESTOR — and that separation survives this SPEC untouched, because a
lead running `git merge-base --is-ancestor` by hand is outside every surface here. The collapsed row
is therefore *observationally single on the shipped surfaces*, not observationally single in the
world. The earlier draft's collapse leaned on S3 being retired; it is re-argued here without that
premise, and the argument is unaffected: S3 was never one of the surfaces this SPEC repairs.

The distinction that collapses is real and is not lost, only unaddressed here: separating a **run**
landing (C4) from a **fully sync-closed** one (C5) needs evidence about *which* commits landed and
what the card's SPEC status is — the false positive t330 §A.4 measured, and card **t359**'s scope
(§A.5). This SPEC's honest claim is narrower and is written into the doctrine by REQ-TLS-011: a
`landed` answer reports that the resolved ref's history names the card, **not** that the card's last
step landed.

**What S2 does NOT gain: the delivering commit's name.** `todo pr` keeps rendering `landed` as a
boolean fact about the resolved ref and names no commit. That is `SPEC-KANBAN-QUEUE-PR-SYNC-001`
REQ-1.10 (`spec.md:251-255` @ `3de2f85a2`), enforced at `prlink_landed.go:62-67` and `prlink.go:42-43`
@ `3de2f85a2`, and this SPEC does not touch it — changing the ref the question is asked about changes
no part of what the answer may carry. Whether an observation record may name commits, and under what
rule, is t359's question, not this one's.

Three things the tables make visible that prose does not:

1. **C1 and C5 collapse onto the same cell today** (`queued`/`picked` differ, but the landed-bearing
   column reads `no-link` for both). That collapse is the misdispatch, and it is what this SPEC
   repairs.
2. **C4 and C5 cannot be separated by a boolean at all.** A single landed answer cannot tell "run
   landed" from "everything landed" — t330 §A.4's finding — so this SPEC does not pretend to, and
   the pair stays merged in Table 2. Separating them needs the evidence record, which is t359's.
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

**M3 — the blast radius is bounded and measured.** `grep -rn 'LandedRef' --include='*.go' internal/`
@ `3de2f85a2` returns **12** lines, which account as: 1 doc comment, 1 declaration, **7 production
uses**, 2 test lines, **and 1 substring false positive** — `internal/cli/todo_undone_test.go:266`,
the function name `TestTodoDone_RequireLandedRefusesWhenNotLanded`, which contains `LandedRef`
inside "Require**LandedRef**uses" and is not a reference to the identifier at all. The false positive
is named rather than absorbed because it is the reason the substring count reads one higher than the
identifier count: the **word-boundary** form
`grep -rnw 'LandedRef' --include='*.go' internal/` returns **11** — the eleven real references — and
a reader who re-runs the unanchored grep and counts 12 has found the collision, not a missed site.
Both figures were re-measured during the 0.3.0 remediation and are stated together for that reason.
The 7 production uses:

| Site | Use |
|---|---|
| `internal/kanban/prlink_landed.go:44` | the `git log` ref operand — the behavioural one |
| `internal/kanban/prlink_landed.go:78` | an error string |
| `internal/cli/todo_pr.go:75`, `:87` | user-facing help text |
| `internal/cli/todo.go:357`, `:399`, `:428` | user-facing help and refusal text |

Plus the declaration at `prlink_landed.go:28` and its comment at `:26-27`, and one test —
`internal/cli/todo_undone_test.go:277-278`, asserting the refusal names the ref.

**Seven production sites and one test** (an earlier draft of this row said six, contradicting the
grep printed beside it). t330 §D deferred this correction on blast-radius grounds; the measured
radius is small and entirely inside two packages.

Four **doc comments** hardcode `origin/main` in prose and go stale the moment the ref resolves —
they are not `LandedRef` references, so the grep above does not find them, and a reader who trusts
them after the change is misled by the comment layer rather than the code: `prlink_landed.go:26-27`
("LandedRef is the ref the landed question is asked about"), `:52` ("origin/main is an existing
local ref" — the exact premise REQ-TLS-001 stops relying on), `prlink.go:31` ("already in
origin/main"), and `prlink.go:42` @ `3de2f85a2`. They are named in `plan.md` M1 so the change
carries them.

**Why not simply re-point the constant at `origin/develop`.** It would fix this repository and break
every downstream one, because the constant ships in the binary while the integration branch is a
per-project fact. The neutral-default rule in M2 exists for exactly this reason.

### B.2 Decision 2 — this SPEC persists nothing; the storage axis is card t359's

Ruled: **nothing is persisted.** This SPEC changes what the landed question asks and how its answer
is reported; it stores no landing evidence, adds no field to the queue record, adds no table, and
renders no observed commit.

**Card t359 owns the storage axis** — landing observations on the queue record, the shape that holds
them, the recording verb, an observed commit on the `todo pr` surface, and the live SPEC-status read
that separates a run landing from a sync landing. **t359 depends on this SPEC landing first**: an
evidence record written from a predicate that asks the wrong ref stores wrong evidence permanently,
and a stored observation costs far more to correct than a live answer. The questions the storage
design must answer — what `REQ-TODO-013` permits, how `SPEC-TODO-ANALYSIS-001` read it, and whether
an observation may name a commit under `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 — belong to t359 and
are deliberately not argued here. §D records the whole axis as out of scope.

### B.4 Explicit NON-GOAL — a landing answer is evidence, never a transition

**The tooling must not close, move, drop, re-order, or re-state a card on its own authority when it
detects a landing.** This is stated here in the SPEC body, not only as a design note, because it is
the most plausible slip for the next person reading this design: *"since we are detecting the landing
anyway, let the tool close the card."* It binds this SPEC even though this SPEC stores nothing — the
detection surfaces it changes are exactly where the slip would be introduced.

The reason it is forbidden is not stylistic. `.claude/skills/moai/workflows/todo.md:53-57` @
`3de2f85a2` is [HARD]:

> `edit`, `move`, `drop`, `undrop`, `done`, and `undone` are operator acts, exactly like `add` and the
> pick. Correct a card's wording, move it, or discard it because the operator said to — never on
> inferred priority, never as tidy-up, never to fold one card into another, and never because a card
> looks stale. **The queue records the operator's intent; it does not curate it.**

A landing answer is a *measurement*. Acting on it converts a measurement into a decision, and the
decision is the operator's. The failure mode this prevents is concrete and worse than the one the
card is about: an auto-close on a **run** landing (state C4) would close a card whose sync commit is
still unpushed in a lane's worktree — precisely the t306 incident that motivated t330, now performed
by the machine and at scale. The ref correction makes the slip *more* attractive, not less, because
after it every develop-landed card starts answering `landed`.

The mechanical form of the non-goal:

- `moai todo pr` remains **read-only** — it writes no field, no finding, no cache, no lock
  (`internal/cli/todo_pr.go:1-16` @ `3de2f85a2` states this as a property of the file). Detection
  does not become a write path. This is asserted behaviourally, not by name-matching: see
  `acceptance.md` AC-TLS-008 and its planted-mutant RED.
- `--require-landed` stays **opt-in**. Flipping it default-on is out of scope (§D) even though this
  SPEC repairs the ref: t330 §A.4's false-positive finding is not closed by the ref correction, and a
  default-on check that passes structurally is worse than no check.

### B.5 Decision 4 — the unanswerable case stays permissive, and stops being silent

Ruled: **the proceed-on-unanswerable policy is NOT reversed.** `todo.md:59-67` @ `3de2f85a2` states
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

**11 requirements, Tier M (ceiling 16).** An earlier draft carried 26, above even the Tier L ceiling
of 25. The overage was not editorial: it was two SPECs in one document. The storage half is now card
t359 (§A.5, §B.2, §D), and what remains here is the discriminator — stated once each, with no
requirement split across two entries to look smaller.

### C.1 The landed ref

- **REQ-TLS-001** — **Where** `git_strategy.worktree_base_branch` carries a non-empty value, the
  landed check shall resolve its ref from that value; **where** it is empty, the landed check shall
  use `origin/main`, byte-identically to the behaviour at `3de2f85a2`. In either case the check shall
  read the ref as the repository holds it, and shall not require the ref to be reachable from the
  session's own working tree.
- **REQ-TLS-002** — Every surface that names the landed ref in user-facing text shall name the
  **resolved** ref, never a compile-time constant.

### C.2 The three-valued answer

- **REQ-TLS-003** — The landed check shall return one of exactly three answers — `landed`,
  `not-landed`, `unknown` — carried in a field a consumer must read; no consumer shall infer
  `unknown` from a zero value, an empty string, or a dropped error.
- **REQ-TLS-004** — **When** the resolved landed ref does not resolve in the repository, the landed
  check shall answer `unknown` and shall not answer `not-landed`; and `moai todo pr` shall render
  that answer as an outcome distinct from `no-link`, naming the ref that did not resolve.

### C.3 Distinguishability on the channel callers read

- **REQ-TLS-005** — `moai todo done` shall emit exactly one landing-verdict token on **stdout** in
  every invocation, including invocations made without `--require-landed`; the verdict shall not be
  carried on stderr alone.
- **REQ-TLS-006** — **While** `--require-landed` is set and the landed answer is `unknown`,
  `moai todo done` shall archive the card, shall not refuse, and shall emit a stdout landing-verdict
  token distinct from the token it emits when the answer is `landed`. The permissive policy at
  `todo.md:59-67` @ `3de2f85a2` is preserved unchanged.

### C.4 The read surface

- **REQ-TLS-007** — `moai todo pr` shall render each card's queue `state` alongside its link outcome,
  so that a `queued` card with no link and a `picked` card with no link are distinguishable rows.

### C.5 Authority

- **REQ-TLS-008** — Detecting a landing shall not change a card's `state`, position, text, or
  `spec_id`, and the tooling shall not close, archive, drop, re-order, or re-state a card on its own
  authority under any circumstance, including a fully-landed card whose SPEC reads `completed`.
- **REQ-TLS-009** — `moai todo pr` shall remain read-only: no field, no finding, no cache, no lock.

### C.6 Doctrine

- **REQ-TLS-010** — `.claude/skills/moai/workflows/todo.md` and its template mirror
  (`internal/template/templates/.claude/skills/moai/workflows/todo.md`) shall be updated in the same
  change to carry the resolved ref, the `unknown` outcome, and the stdout verdict token, and the
  [HARD] operator-only mutation rule shall be restated unweakened.
- **REQ-TLS-011** — The doctrine shall state the landed check's remaining limit **at each surface
  that renders a landing answer**, not only under `--require-landed` where it is stated today.
  Concretely: (a) the `moai todo pr` outcome list shall state that `landed` reports that the resolved
  ref's history names the card, not that the card's last step landed; and (b) the same list shall
  state that `unknown` means the question could not be asked and is **not** evidence of not-landed.
  The existing `--require-landed` limit note stays, unweakened.

  > **Why this is a delta and not a restatement.** The limit exists in the doctrine today, but exactly
  > once and scoped to the opt-in guard — `grep -n 'LAST step' .claude/skills/moai/workflows/todo.md`
  > returns a single line, `60`, inside the paragraph that opens `[HARD] `--require-landed` is OPT-IN
  > and honestly limited`. The `todo pr` outcome row (`:51`) describes `landed` mechanically and
  > states no limit, and `grep -c -i 'unknown'` over the file returns `0` (rc=1) because the outcome
  > does not exist yet. (a) and (b) are therefore both absent today. See `acceptance.md` AC-TLS-009's
  > RED cell for the commands and their verbatim output.

---

## §D Exclusions

This SPEC deliberately does not build the following. Each is out of scope for a stated reason, not by
oversight.

### Out of Scope — landing evidence and its storage (card t359)

This SPEC persists nothing (§B.2). The whole storage axis belongs to **card t359**, which depends on
this SPEC landing first: evidence recorded from a predicate that asks the wrong ref is wrong
permanently, and a stored observation costs far more to correct than a live answer.

- Landing observations on the queue record, in any shape — a top-level array, a per-item field, a
  new table, or a column.
- The verb that would record one, and any question about which verb records it.
- Rendering an observed commit SHA or subject on `todo pr`, and any rule that would select among
  the several commits a card's grep predicate matches (t322 matches 24 on `origin/develop`, §A.2).
- The live SPEC-status read that would separate a run landing from a sync landing.
- Retention, pruning, or compaction of any of the above.
- Everything the storage design must reconcile before it can be built — what `REQ-TODO-013` permits,
  how completed `SPEC-TODO-ANALYSIS-001` read it, and whether an observation may name a commit under
  `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10. Those are t359's questions and are not argued here.

### Out of Scope — queue schema normalization

- `tier` and `depends_on` field normalization on the card record. The dispatching lead excluded it,
  and the reason is recorded here so it is not re-litigated: its stated precondition is a revision of
  the `todo.md` [HARD] rules, which is a doctrine change and is not open. Low schema cost is not by
  itself a reason to do it now; when the precondition is met, a follow-up card pays the same additive
  cost this SPEC would have paid.
- Any fourth `state` value, including `landed`. `internal/kanban/backlog_sqlite.go:113` @ `3de2f85a2`
  constrains `items.state` with `CHECK (state IN ('queued','picked','dropped'))`, and the comment
  above the DDL (`:95-97`) explains that SQLite cannot ALTER a CHECK constraint, so a fourth value
  would force a table rebuild on every operator queue in the field.
- Any new per-item field (`internal/kanban/backlog_store.go:62-63` @ `3de2f85a2` names the five-field
  contract). Whether that contract is extensible at all is t359's question, not this SPEC's — nothing
  here touches the record.
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
- Making the predicate phase-aware (telling a run landing from a sync landing) at all. This SPEC
  answers `landed` for every card the resolved ref names and says so plainly in the doctrine
  (REQ-TLS-011); the distinction needs evidence this SPEC does not collect (card t359).

### Out of Scope — adjacent surfaces

- **Retiring hand ancestry (`git merge-base --is-ancestor`) as a decision input.** The S3 column of
  §A.6 records what the lead's manual fallback reports; this SPEC neither changes it nor withdraws
  it. Retiring a manual practice is a doctrine decision about how a lead works, not a code change on
  any surface here, and no requirement or criterion in this SPEC delivers or verifies one. A
  corrected S2 is expected to make the fallback unnecessary in practice, which is an *effect* of this
  SPEC and not a claim it makes.
- The guard's **manifestation history** — how many cards were misjudged, when, and what each
  misjudgement cost. That is card t347's scope. This SPEC uses four cards as test inputs and claims
  nothing about the population.
- Any change to the `gh` query budget on `todo pr` (one query per invocation,
  `internal/cli/todo_pr.go:8-13` @ `3de2f85a2`), or any new default-on network cost on `todo list`,
  `todo next`, or any other read path.
- `moai doctor`'s worktree-base-branch check (`internal/cli/doctor_worktree_base.go:40-43` @
  `3de2f85a2`). It already reports an unresolvable configured branch; this SPEC consumes the same
  resolver but adds no diagnostic surface.

---

## §E Traceability

Every requirement carries at least one acceptance criterion, and every acceptance criterion traces
to a requirement — 11 requirements, 10 criteria, no orphan on either side.

| Requirement | Acceptance criteria |
|---|---|
| REQ-TLS-001 | AC-TLS-001, AC-TLS-002 |
| REQ-TLS-002 | AC-TLS-003 |
| REQ-TLS-003 | AC-TLS-004 |
| REQ-TLS-004 | AC-TLS-004, AC-TLS-005 |
| REQ-TLS-005 | AC-TLS-006 |
| REQ-TLS-006 | AC-TLS-006, AC-TLS-007 |
| REQ-TLS-007 | AC-TLS-010 |
| REQ-TLS-008 | AC-TLS-008 |
| REQ-TLS-009 | AC-TLS-008 |
| REQ-TLS-010 | AC-TLS-009 |
| REQ-TLS-011 | AC-TLS-009 |

---

## §F Cross-references

- `SPEC-TODO-DESTRUCTIVE-GUARD-001` — the archive, `undone`, `--expect`, and the `--require-landed`
  seam this SPEC consumes; its §B.2 boundary table is inherited verbatim in §A.5.
- `SPEC-KANBAN-QUEUE-PR-SYNC-001` — the resolver and the landed primitive being corrected; its
  REQ-1.10 (the resolver names no delivering commit) is untouched by this SPEC.
- `SPEC-WORKTREE-BASEREF-001` (card t313) — `worktree_base_branch`, its neutral-empty ruling, and the
  resolver reused in §B.1.
- **Card t359** — the landing-evidence half of this axis (storage, recording verb, observed commit
  on `todo pr`, live SPEC status). Depends on this SPEC; see §A.5 and §D.
- `.claude/skills/moai/workflows/todo.md` — the [HARD] operator-only mutation doctrine and the
  `--require-landed` limit note.
- `.claude/rules/moai/development/verification-completeness.md` — the two-cell adoption discipline
  and the mutant probe AC-TLS-008 was rewritten under.
