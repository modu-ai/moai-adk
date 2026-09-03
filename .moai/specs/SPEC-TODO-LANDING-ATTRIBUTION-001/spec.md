---
id: SPEC-TODO-LANDING-ATTRIBUTION-001
title: "The landed verdict: an attribution-position predicate, and a ref chain that asks the branch this repository actually integrates on"
version: "0.2.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec (card t472)
priority: P1
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, landing-state, attribution, git-flow, integration-branch, false-positive"
tier: M
related_specs:
  - SPEC-TODO-LANDING-STATE-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC: The landed verdict — attribution position, then the right ref

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.2.0 | 2026-09-03 | Plan-audit iteration-1 remediation, measured in this tree at HEAD `e227871b4` against `origin/develop` `7835148d3` (5,837 subjects). D1 (BLOCKING) closed: §A.4 form 3 was an occurrence test and is replaced by two positional shapes (3a, 3b) plus an explicit non-attribution rule for absorb-direction merges; `MUT-MERGE-ANY-TOKEN` added to the mutant set. D2 citation corrected and its residual re-stated as measured-disjoint. D4/D5/D6/D7/D8/D9 dispositions recorded in `acceptance.md`, `plan.md`, and `progress.md`. Requirement and criterion counts unchanged at 12/12. |
| 0.1.0 | 2026-09-03 | Initial plan-phase authoring (card t472), measured in worktree `.claude/worktrees/t472` at HEAD `4bcac7079` (branch `WT-landed-drift-detect`). Two axes, merged by lead verdict into one SPEC: the attribution predicate (axis F) and the ref chain plus its disclosure (axes A+B). Every figure below is carried from this lane's own committed measurements at `.moai/reports/t472/premise-recheck.md` and `.moai/reports/t472/axis-bf-measurement.md`, or re-run in this tree and cited beside the command. |

> **Provenance discipline.** Every `file:line` citation is measured at tree `4bcac7079`. Every figure
> **about** a moving ref (`origin/main`, `origin/develop`, `refs/remotes/origin/HEAD`) travels with the
> command that produced it and is re-measured rather than re-cited by a later reader
> (`verification-claim-integrity.md` §2.1, remedy R4).

---

## §A Context

### A.1 What is wrong, in one sentence per axis

**Axis F — the predicate answers a question nobody asked.** `LandedGrepArgs`
(`internal/kanban/prlink_landed.go:96-109`, tree `4bcac7079`) builds

    git log <ref> --perl-regexp --grep=\btNNN\b --oneline

and `--grep` matches the **whole commit message**. A commit that merely *mentions* a card in its
body therefore reads as that card having landed. The file's own comment already knows the hazard —
"a card's first matching commit may be another card's report commit that merely mentions it"
(`prlink_landed.go:139-140`) — but that knowledge was spent suppressing SHA output, not on the
match itself.

**Axes A+B — the verdict asks the wrong branch, and never says which.** `LandedRefFor`
(`prlink_landed.go:74-80`) reads `git_strategy.worktree_base_branch` from the root the queue hangs
from, which `todoLandedRef` (`internal/cli/todo.go:81-90` — doc comment `:81-87`, func `:88-90`) deliberately resolves to the **primary
checkout** — "the queue and the integration branch are properties of one repository, not of
whichever worktree the command happens to run in". That primary checkout is parked on `main`, whose
copy of the key is empty, so resolution falls to the hardcoded `DefaultLandedRef = "origin/main"`
(`prlink_landed.go:41`). The fallback is taken with no run-time notice, and the verdict line
`done <id> landing=<verdict>` (`internal/cli/todo.go:508`) does not name the ref that answered.

### A.2 The measured size of axis F

Measured in this tree with a binary built from it (`go build ./cmd/moai`), not the installed build
— VCI §2.2.

| Population | Landed verdicts | True positives | False positives |
|---|---|---|---|
| Live queue against `origin/main` (today's ref), 57 rows | 2 (t237, t312) | 0 | **2** |
| Live cards against `origin/develop` (the ref axis A would select), 31 cards | 9 | 2 (t401, t440) | **7** |

Precision of the shipped predicate on the population it currently marks landed: **0/2**.

Reproduction of one false positive, re-run in this tree at `4bcac7079`:

    git log origin/main --perl-regexp --grep='\bt237\b' --oneline
      539349c5b docs(t230): t230 sync-audit evidence ... (#1649)
      32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

Both commits are **t230's**. `32d2221fa`'s body reads "the card about to change the hook body is
t237/#1641. Re-measured: the issue is OPEN" — a statement that t237 has *not* landed, read as
evidence that it has. Commits whose subject attributes t237: zero.

Control (the comparison is not vacuous), re-measured at `origin/develop` `7835148d3` under the
**repaired** §A.4 enumeration (forms 1, 2, 3a, 3b):

    # attributed ids — forms 1/2/3a
    git log origin/develop --format=%s | grep -ohE '^[a-z]+\(t[0-9]+\)!?:|\((card )?t[0-9]+\)$|^Merge card t[0-9]+' \
      | grep -oE 't[0-9]+' | sort -u | wc -l           → 257
    # form 3b adds exactly one id beyond those three   → t412
    # ids mentioned anywhere in a message
    git log origin/develop --format=%B | grep -oE '\bt[0-9]+\b' | sort -u | wc -l   → 379

Attributed = **258**, mentioned = **379**, and `attributed ⊆ mentioned` holds (`comm -23` → 0 lines).
Both operands are non-empty, so the comparison asserts something.

*Correction (plan-audit D9a).* Version 0.1.0 recorded this control as **291**. That figure was
produced by a broader proxy (paren convention + bare token) than §A.4 defines, and does not
reproduce under the SPEC's own enumeration. The figure above is the reproducible one. The control's
purpose — both operands non-empty — held under either.

### A.3 Why matching the subject is not enough

Two of the seven `develop` false positives survive a "match the subject only" repair, because they
are **other cards' subjects mentioning this card**. Re-run in this tree:

    git log origin/develop --perl-regexp --grep='\bt216\b' --oneline    # 3 lines, shown in full
      48c35a4d4 Merge branch 'WT-incremental-rebuild' into develop (card t263)
      673d3d8a0 docs(t263): die-at-exit reproduced (0/5) — remedy sequenced behind t216
      2f170549b fix(hooks): re-wire the navigator SessionStart hook (t243)

    git log origin/develop --perl-regexp --grep='\bt443\b' --format='%h %s'   # 14 lines; 1 shown
      0d26f8a00 chore(catalog): revert sync-auditor hash to develop value — t443 jurisdiction (t461)
      [13 further lines elided — body mentions and other cards' subjects; selection rule: the single
       line whose *subject* contains the queried token is shown, the rest are body-only matches]

None of the three `t216` lines attributes t216: `48c35a4d4` and `673d3d8a0` are **t263**'s and
`2f170549b` is **t243**'s (its body mentions t216). `0d26f8a00` is attributed to **t461**. In the two
subject-occurrence cases (`673d3d8a0`, `0d26f8a00`) the queried card appears in the subject and is
not the card the commit belongs to. The discriminator is therefore not *occurrence*, nor
*occurrence in the subject*, but **attribution position**.

### A.4 The attributing positions — four positional shapes

Every shape below is a **position**, not an occurrence. A card token that appears anywhere else in a
subject — mid-sentence, inside a branch name, inside a dependency or absorb note — attributes
nothing. Counts re-measured at `origin/develop` `7835148d3`, 5,837 subjects, 414 of them merges.

| # | Form | Positional rule | Observed | Example |
|---|---|---|---|---|
| 1 | Conventional-commit scope | `<type>(<card>):` **at subject start** | 290 | `docs(t440): record develop-absorb re-measure evidence` |
| 2 | Trailing parenthetical | `(<card>)` or `(card <card>)` **closing the subject** | 869 | `... refresh catalog moai whole-tree hash (t447)` |
| 3a | Merge, card-led | subject **begins** `Merge card <card>` | 5 | `Merge card t440 (WT-delivery-notice-docs) into develop: ...` |
| 3b | Merge, integration-targeted | the merge's **named target is the branch the landed ref resolves to**, AND the card token lies **inside the subject's trailing parenthetical group** | 76 | `Merge branch 'WT-mx-tag-edges' into develop (card t412 — SPEC-MX-TAG-EDGES-001)` |

**[HARD] The non-attribution rule.** A merge whose named target is a **card worktree branch**
(`WT-…`) attributes **no card**, in any position. Such a merge absorbs work into a card's branch; the
subject of the sentence is a branch, not a card, and its parenthetical is a dependency note or an
absorb record rather than an attribution. Measured: **21** absorb-direction merges carry a card token
inside a trailing parenthetical group and are excluded by this rule alone.

    git log origin/develop --format=%s | grep -E '^Merge ' | grep -E 'into WT-' \
      | grep -oE '\([^()]*\)$' | grep -cE 't[0-9]+'        → 21

**Why 3b, and not "a merge subject naming the card" (plan-audit D1).** Version 0.1.0 wrote form 3 as
*"a merge subject naming the card"* — an **occurrence** test, and therefore the exact reading §A.3
rejects, reproduced inside this SPEC's own enumeration. Measured, the occurrence reading admits
**146** merge subjects against the 5 that form 3a describes — a 29× widening — of which **50** are
caught by no attributing form at all. Two of them state the case:

    Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)
    Merge branch 'develop' into WT-inbox-drain-gap (absorb t280, lane-15 window; includes t239 merge e79c010b8)

The first names **t387 and t386**, as a dependency note; the second names **t280 and t239**, as an
absorb record. Neither is an attribution, and both are excluded by the non-attribution rule above.

**What the repair costs and what it buys, measured.** Under forms 1/2/3a the attributed-id set is
**257**; form 3b adds exactly **one** further id — `t412`, whose only landing evidence is the merge
`b6231290d ... into develop (card t412 — SPEC-MX-TAG-EDGES-001)`, whose trailing group carries text
after the card id and so escapes form 2's `)$` anchor. Set total **258**.

    comm -13 <attributed under 1/2/3a> <attributed under 3b>   → t412   (exactly one line)

One under-count survives and is recorded rather than hidden: **t250** is attributed on `develop` only
by `6786c3fa4 t250: graph freshness ... (#1648)` — a bare `t250:` prefix, which is none of the four
shapes. It will read `not-landed`. That failure direction is **loud** (an operator who knows the card
landed sees `not-landed`), which is the direction `plan.md` §D already accepts for a fifth
convention; the alternative — widening to catch it — is the occurrence reading this SPEC exists to
remove.

### A.5 The ref: the repository already knows the answer

    git symbolic-ref refs/remotes/origin/HEAD   → refs/remotes/origin/develop
    primary .git/HEAD                            → ref: refs/heads/main
    primary .moai/config/sections/git-strategy.yaml:7 → worktree_base_branch: ""

The git-level answer is already `develop`. The landed check ignores it and uses a compiled-in
constant, because the only level it consults — a config file in a checkout parked on `main` — is
empty there. The key's value **is** `develop` on the integration branch (`a9c61cf56`), but that
commit is not an ancestor of `origin/main`, so the value cannot take effect until a release carries
it across.

### A.6 What the four consumers of the key want

All four consumers of `git_strategy.worktree_base_branch` want the **integration** meaning; none
wants the release meaning. One value can therefore answer all four.

| Consumer | Location (`4bcac7079`) | Wanted meaning |
|---|---|---|
| Card-worktree base | `internal/cli/session_worktree.go:215` | integration |
| doctor check | `internal/cli/doctor_worktree_base.go:43` | integration |
| Landed ref | `internal/kanban/prlink_landed.go:75` | integration |
| SessionStart alignment | `internal/hook/worktree_base_branch.go:125` (the write; see below) | integration |

The fourth **writes** `refs/remotes/origin/HEAD` to match the setting. The key is therefore not
inert, and no requirement below may be justified on the premise that changing its value is
mechanism-free.

*Correction (plan-audit D2).* Version 0.1.0 cited `worktree_base_branch.go:156` as the write. Measured
at HEAD `e227871b4`: `:155` is `worktreeBaseBranchReadConfigReal`, a config **read**; the write is
`WorktreeBaseBranchSetHead(configured)` at **`:125`**, backed by `worktreeBaseBranchSetHeadReal`
(`git remote set-head`) at **`:170`**. The corrected citation is `:125` / `:170`.

*And the write and chain level 2 are measurably disjoint.* `RunWorktreeBaseAlignment` gates on the
primary checkout at `:92` and returns at `:97-100` when the configured key is **empty**:

    // internal/hook/worktree_base_branch.go:92-100, HEAD e227871b4
    if !WorktreeBaseBranchInPrimaryCheckout() { return data }
    configured := worktreeBaseBranchReadConfig(projectRoot)
    if configured == "" {
        // REQ-WBR-005: the neutral value performs no git-metadata read at all.
        return data
    }

Chain level 2 fires only when level 1 is **empty**; the writer fires only when level 1 is
**non-empty**, from the same primary-checkout root `LandedRefFor` reads. The two are mutually
exclusive by construction, so **no cycle exists**. §D records what remains.

### A.7 The ordering decision — [HARD]

**Axis F lands before axis A.** Correcting the ref first does not fix a symptom; it **grows the
false-positive population from 2 to 9** (§A.2). Nine cards would then read landed on a ref the
operator trusts, seven of them wrongly. The predicate is repaired first, and the ref chain is laid
on top of a predicate that can bear it. This decision binds the milestone order in `plan.md`.

---

## §B Requirements

Notation: GEARS. Requirement IDs are stable; milestone assignment is in `plan.md`.

### B.1 Milestone 1 — the attribution predicate (axis F)

- **REQ-TLA-001** (Ubiquitous) — The landed predicate shall decide `landed` on **attribution
  position** within a commit's subject line, and shall not decide it on the presence of the card
  token anywhere in the commit message.

- **REQ-TLA-002** (Ubiquitous) — The set of attributing positions shall be exactly the positional
  shapes enumerated in §A.4 — conventional-commit scope (form 1), trailing parenthetical, bare and
  `card`-prefixed (form 2), card-led merge subject (form 3a), and integration-targeted merge subject
  whose trailing parenthetical group carries the card token (form 3b) — and shall include §A.4's
  non-attribution rule: a merge whose named target is a card worktree branch (`WT-…`) shall attribute
  no card. No shape shall be expressed as a bare occurrence test — "the subject contains the token"
  is not a position, and admitting it reintroduces REQ-TLA-001's defect through the enumeration. The
  enumeration and its non-attribution rule shall live in one named place in the implementation, so a
  further form is a reviewable one-place diff.

- **REQ-TLA-003** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  only occurrence of the queried card token lies in the commit body.

- **REQ-TLA-004** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  subject contains the queried card token in a **non-attributing** position, that is, a commit
  attributed to a different card.

- **REQ-TLA-005** (Ubiquitous) — The argv the landed check runs shall be constructed by a single
  exported builder, so a tripwire test asserts against the implementation's own construction rather
  than a transcription of it.

- **When** the landed query cannot be evaluated — no git, no such ref, a query error — the querier
  shall answer `unknown` rather than `not-landed` (**REQ-TLA-006**, event-detected; preserves the
  three-valued contract of `SPEC-TODO-LANDING-STATE-001`).

### B.2 Milestone 2 — the ref chain and its disclosure (axes A+B)

- **REQ-TLA-007** (Ubiquitous) — The landed ref resolver shall resolve through an ordered
  three-level chain: (1) the configured `git_strategy.worktree_base_branch`, (2)
  `refs/remotes/origin/HEAD`, (3) `DefaultLandedRef`.

- **Where** level 1 yields an empty value, the resolver shall consult `refs/remotes/origin/HEAD`
  before reaching the constant (**REQ-TLA-008**, capability gate).

- **When** `refs/remotes/origin/HEAD` cannot be resolved, the resolver shall yield
  `DefaultLandedRef` and shall not fail the invocation (**REQ-TLA-009**, event-detected).

- **REQ-TLA-010** (Ubiquitous) — The `todo done` landing verdict line shall name the ref that
  answered, appended so that the existing `done <id> ` prefix every reader keys off is preserved.

- **While** the answering ref came from a level below the configured key, **when** a landing verdict
  is emitted, the command shall disclose on stderr which chain level supplied the ref
  (**REQ-TLA-011**, compound).

- **REQ-TLA-012** (unwanted) — The landed ref resolution shall not write any repository ref, and in
  particular shall not set `refs/remotes/origin/HEAD`.

---

## §C Constraints

- **C-1** — The three-valued `LandingAnswer` contract (`landed` / `not-landed` / `unknown`) is
  inherited from `SPEC-TODO-LANDING-STATE-001` and is not renegotiated here.
- **C-2** — `--require-landed` remains opt-in, and remains asymmetric: it refuses on `not-landed`
  only, and proceeds on `unknown`. The measured behaviour at `internal/cli/todo.go:555-578` is the
  baseline.
- **C-3** — The landed query performs no network I/O. Level 2 of the chain reads a local ref.
- **C-4** — A project that configures the key keeps its current behaviour exactly: level 1 still
  answers first.
- **C-5** — The queue root resolution (`resolveTodoQueueRoot`, one repository one queue) is not
  changed. The chain is added below the config read, not moved to a different tree.

---

## §D Exclusions

This section states what is **out of scope** for this SPEC and why, so no successor re-derives the
boundary.

### Out of Scope — axis C, landing evidence storage

- `archived_items` carries no landing column. Its measured schema is
  `seq, id, text, added_at, spec_id, state, position` — no landing SHA, no answering ref, no verdict
  instant. The `landing=` line `todo done` prints is stdout-only and is never persisted; all 126
  archived rows are closed with no stored evidence.
- This is recorded here as **measured context only**. The surface belongs to card **t359**, which is
  picked and carries plan-audit iteration-1 redesign items. Authoring requirements against it here
  would duplicate a live card.

### Out of Scope — axis D, the blank card-to-SPEC link

- Re-measured in this tree: `items` 57/57 rows with a blank `spec_id`; `archived_items` 126/126
  blank. The attaching path (`todo next <n> --spec <SPEC-ID>`) exists and is used zero times.
- The card excludes this axis explicitly. It is an operational gap, not a schema defect, and it is
  recorded as context only.

### Out of Scope — axis E, a new detection subcommand

- The detection surface already ships: `moai todo pr` reads the whole live queue (57 rows), emits a
  five-value outcome column, and writes nothing.
- What is missing is that column's **precision**, which REQ-TLA-001..004 supply, and an operational
  routine for running it, which is not code.
- Representative mutant, declared a non-goal: building a new query subcommand. An implementation
  that adds one has not delivered this axis; it has re-delivered `todo pr` under a second name.

### Out of Scope — the level-2 / SessionStart-writer interaction

- Measured disjoint (§A.6): chain level 2 fires only on an empty configured key, and
  `RunWorktreeBaseAlignment`'s write fires only on a non-empty one, from the same primary-checkout
  root. No cycle exists, so nothing here is a requirement.
- What genuinely remains is the converse and it is **not** a coupling: a project that **configures**
  the key never reaches level 2 at all, so the two surfaces never interact on this path. Should a
  future change make the writer fire on an empty key, that disjointness ends — but repairing it then
  is that change's obligation, not this SPEC's.

### Out of Scope — release-wait as the axis-A remedy

- Waiting for `develop` to reach `main` was rejected by lead verdict: its timing is bound to the
  t204 deploy gate, and the closed-but-picked backlog recurs on every release cycle regardless.

### Out of Scope — semantic landing ("has this card's last step landed")

- The predicate answers "has a commit attributed to this card landed on the ref", not "has this
  card's sync step landed". That limit is inherited verbatim from `SPEC-TODO-LANDING-STATE-001` and
  is not narrowed here.

---

## §E Cross-references

- `SPEC-TODO-LANDING-STATE-001` — the ancestor: the ref became resolvable, and the answer became
  three-valued. This SPEC corrects the predicate that ref is asked with, and completes the
  resolution chain.
- `.moai/reports/t472/premise-recheck.md` — the card's original axis-A premise, re-checked in this
  tree and found false.
- `.moai/reports/t472/axis-bf-measurement.md` — the axis B–F measurements and the merged axis A+B
  design section.
- `internal/kanban/prlink_landed.go`, `internal/cli/todo.go`, `internal/cli/todo_pr.go` — the
  implementation surface.
