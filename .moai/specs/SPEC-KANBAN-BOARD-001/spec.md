---
id: SPEC-KANBAN-BOARD-001
title: "Six-column kanban board model with a single-origin board state store"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, column, wip-limit, state-store, single-origin, worktree-safe, sole-writer, atomicity, branch-side-status, stale-lock, role-declaration"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001]
related_specs: [SPEC-KANBAN-WORKTREE-001, SPEC-KANBAN-BOOTSTRAP-001, SPEC-FACTORY-MODE-001]
---

## HISTORY

- **v0.3.0** (2026-08-11) — Plan-audit repair at the layer the v0.2.0 repairs exposed. The six prior repairs verified closed with no regression; the 0.825 shortfall rested entirely beneath them, and every remaining defect was delta-closable. Eight repairs, and the first is the one that broke the normal path. **(1) No SPEC decided which tree the board reads a card's `status` from**, and the omission made every card inconsistent: `status` transitions are written on the card's *branch*, inside the worktree that `SPEC-KANBAN-WORKTREE-001` `REQ-KW-005` keeps alive until both pull requests merge, so the primary checkout's copy still reads `draft` while the card sits in `run` — a pairing outside §A.4, which `REQ-KB-008` refuses to dispatch. The irony is exact: §A.3(3) already argued that a card's branch is cut early and merged late, and never turned that argument on the status it reconciles against. `REQ-KB-020` decides it — the card's branch where one exists, the primary checkout where none does (§A.4a) — consuming `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` for branch identification as a **contract** dependency stated in requirement text, with that sibling kept in `related_specs:`. It was briefly promoted to `dependencies:` during this revision, which closed a cycle — that sibling already declares this SPEC on a genuine landing need — and the promotion is reversed; §A.4a records why the absent edge is a decision rather than an oversight. **(2) The board-wide lock re-created the brick `REQ-KB-013` was amended to remove**: the reused `internal/spec/lock.go` family performs no stale-lock detection on its Windows substrate, by its own header, so a `lead` killed mid-mutation leaves an artifact blocking every future board mutation — `REQ-KB-023`, ported from the sibling's `REQ-KW-014` shape and hardened against the check-then-unlink race that sibling's own audit found. **(3) An **absent** state file was owned by nobody** — neither "partially written" nor "unreadable" — and nothing created the store's directory: `REQ-KB-021`. **(4) The sanctioned recovery could produce exactly the empty board §A.6 forbids**, and "bounded" was asserted without a bound: `REQ-KB-022`. **(5) `REQ-KB-005` mandated reuse of a discriminant that is unexported and returns a boolean, not a path** — a reuse the code shape does not permit — and hard-coded two line numbers into normative text; both repaired, the first by porting `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018`'s extraction disposition. **(6) The board state path collided with `SPEC-KANBAN-RENAME-001` `REQ-KR-009`'s session-record path** — one directory name, two occupants, two different resolution rules — and the board's is moved to `.moai/state/kanban-board/` (§A.3e); the sibling's path is deliberately untouched. **(7) `REQ-KB-017`'s runtime half presupposed a readable caller role**, which `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` now supplies at its v0.3.0 and which this SPEC consumes by name rather than re-deriving. **(8) Three criteria repaired**: `AC-KB-012` varied a knob this SPEC excludes, `AC-KB-017` asked a static enumeration to establish a runtime value, and `AC-KB-002` never entered the fallback branch it requires.
- **v0.2.0** (2026-08-10) — Plan-audit repair, and **Tier M → Tier L** promotion. The promotion exists to make room: two independent auditors found that the predecessor's `REQ-KM-044` bundled two separable things — the mechanism of writing a `column:` field into SPEC frontmatter, and the ownership rule that *the lead session is the sole writer of board state* — and that the split deleted **both**, when only the first was rejected by §A.3. The second changed storage location, not owner, and should have survived. Measured loss: `grep -rc 'sole writer'` and `'single writer'` over all three sibling SPECs report **0** everywhere, and `grep -c 'atomic'` over this `spec.md` reported **0**. Restoring it costs three requirements (`REQ-KB-017` … `REQ-KB-019`) and their criteria, which puts both counts past the Tier M ceiling of 16 — hence Tier L, and hence `design.md` and `research.md`. Four further repairs ride along: the primary-checkout path-resolution rule of §A.3 and REQ-KB-005 was **wrong in the binding document** (the bare `--git-common-dir` form returns a relative `.git` in the primary checkout, so "the parent of it" is not a path); §A.4's `sync` row omitted the legal `completed` pairing; `status: planned` was unhandled by the compatibility table while 17 SPECs carry it; and two acceptance criteria did not verify what their requirement claimed. `REQ-KB-013` is amended with a bounded recovery path — as written it locked the board permanently on one partial write.
- **v0.1.0** (2026-08-10) — Initial plan-phase authoring. First of three SPECs split out of the superseded `SPEC-KANBAN-MULTISESSION-001` (59 requirements, 61 criteria), which failed a plan audit at 0.87 for structural reasons. This SPEC carries the board model only. The predecessor's column/status reconciliation (its §A.4) and its WIP-versus-session-count separation (its §A.5) are carried forward; its `column`-in-frontmatter decision is **rejected and replaced** by a single-origin state store (§A.3), on measured grounds recorded there.

---

## §A. Context

### A.1 What this SPEC is, and what the other two are

The kanban system has three separable concerns. Splitting them is the point of this SPEC's existence: the predecessor tried to carry all three and became unauditable.

| SPEC | Owns |
|---|---|
| **this one** | the six columns, the card record, the board state store, the column↔status consistency check, the WIP limit, and the unheld state |
| `SPEC-KANBAN-WORKTREE-001` | worktree creation and disposal, orphan classification, stall detection, holder release on session death, assignment mutual exclusion |
| `SPEC-KANBAN-BOOTSTRAP-001` | preflight, session roles and topology, bootstrap and the entry switch, configuration, quorum, the dispatch protocol, backend selection, the coder-session internal chain |

The boundary is stated again as an exclusion (§C) so that an auditor reading this SPEC alone does not read the omissions as gaps. This SPEC defines *what a card is and where it sits*; it defines neither *who moves it* nor *where the work happens*.

Every identifier below is written in its **post-`SPEC-KANBAN-RENAME-001`** form. That rename is a `dependencies:` entry and a hard gate (REQ-KB-002).

### A.2 The six columns, and where the status vocabulary does not reach

The six columns are fixed and decided: `backlog` → `plan` → `run` → `review` → `sync` → `done`.

They were intended to map onto the SPEC frontmatter `status` lifecycle. Measured against the canonical 8-value enum (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`, `archived`, `rejected`) and its Status Transition Ownership Matrix, the mapping is **not a bijection**. Two collisions and one absence:

| Column | Status counterpart | Verdict |
|---|---|---|
| `backlog` | *none* — a backlogged item may have no `spec.md` at all, and one that does carries `status: draft` | **absent / colliding.** `(none) → draft` is written by manager-spec at plan-phase artifact creation, so `draft` marks a SPEC whose plan artifacts already exist. A backlog entry that has not been planned has no status to read. |
| `plan` | `draft` | **colliding with `backlog`.** Both read `draft` once artifacts exist. |
| `run` | `in-progress` | exact (`draft → in-progress`, manager-develop, first run-phase commit) |
| `review` | *none* — the verify exit gate fires **inside** run-phase, before the run→sync chain, so the status is still `in-progress` | **absent, colliding with `run`.** There is no `reviewed` status and this SPEC does not invent one. |
| `sync` | `implemented` | exact-ish. The matrix records `in-progress → implemented → completed` as a single manager-docs sync commit, so `implemented` is observable but short-lived. |
| `done` | `completed` | exact |

Four of the eight enum values are absent from that table, and their absence has two different causes. Both are stated as properties, because an auditor reading only the six rows above reads four unmapped values as the same defect twice.

**`planned` is a lifecycle value, and it collides where `draft` already collides.** It is not hypothetical: measured at authoring time, `grep -rlE '^status: planned\s*$' .moai/specs/` reports **17** files. `planned` is a member of the canonical 8-value enum, and the v0.1.0 table admitted it in no column — which made every one of those 17 SPECs illegal in all six columns and, by REQ-KB-008, permanently undispatchable. That is a defect in the table, not in the SPECs.

The reading the schema supports is that `planned` marks a SPEC whose plan artifacts exist and whose implementation has not started — the same span `draft` covers, reached by a legacy-optional path the active V3R6 flow does not write. It therefore collides with `backlog` and `plan` in exactly the way `draft` already does, and it is admitted in exactly those two columns (§A.4). It is admitted nowhere else: a `planned` SPEC in `run` would assert that work started under a status meaning it had not.

**`archived`, `superseded`, and `rejected` are not board cards at all.** Measured: `grep -rlE '^status: (archived|superseded|rejected)\s*$' .moai/specs/` reports **42** files. These are out-of-lifecycle terminals — a SPEC that was retired, replaced, or declined. They are not work in flight, and `done` is the wrong home for them: `done` means *this card was worked and finished*, and a rejected SPEC was never worked. The board therefore admits them to **no column**, which is a different statement from "they are illegal in every column": a card is never created for them, so REQ-KB-008 is never reached and no inconsistency is reported. Stated as a property so the absence reads as a boundary rather than as the `planned` defect a second time (REQ-KB-008).

**The column is explicit, never derived.** The derived reading — computing the column from `status` plus the `progress.md` §E markers `internal/spec/era.go` already parses — is rejected and stays rejected. It is blind exactly where it is needed: `§E.3` is written at the *end* of run-phase, and the `review` column is the interval before it is written, so the derivation cannot separate `run` from `review` — the collision it was brought in to solve.

That operator decision is preserved unchanged from the predecessor. What changed is only *where the explicit value lives* — §A.3.

A `test` column was considered and **rejected**: unit tests belong to the TDD/DDD cycle inside `run`, and a separate column would encode test-after.

### A.3 The board state has one origin: the primary checkout

The predecessor put the card's column in the SPEC's `spec.md` frontmatter, as a new optional `column:` field owned by the lead session. That decision is **rejected and removed**, on three measured grounds.

**(1) A worktree's `.moai/state/` is private to that worktree.** `.gitignore` ignores `.moai/state/` (line 275, with `**/.moai/state/` at line 207 catching the nested SPEC-local form). An ignored directory is not shared between checkouts, so every worktree carries its own copy of anything beneath it. A board that each session reads out of its own tree is six boards.

**(2) `.moai/specs/` is git-tracked, so a frontmatter write is a committed change.** Measured: `git ls-files .moai/specs/ | wc -l` reports **2,528** tracked files. With `enforce_admins: true` on this repository, main-direct push is blocked, so six column transitions per card would require six pull requests — for a field that changes several times an hour on an active board.

**(3) A card's branch is cut before the lead's later column writes.** A card's worktree branch is created early and merged late; the lead's `column` writes land on the primary checkout in between. Merging the card's branch therefore restores a stale column, and the board silently regresses. This is not a race that better locking fixes — it is a consequence of putting mutable board state in a versioned file that a long-lived branch forked from.

**The replacement.** The board state is a file beneath the **primary checkout's** `.moai/state/kanban-board/`, and every session — whichever worktree it runs in — resolves that one path rather than its own tree's copy. The directory name is not the obvious one, and §A.3(e) records why it could not be.

The resolution mechanism is measured and already in service, but **the bare `--git-common-dir` form must never be used alone**, because it is correct from inside a worktree and wrong in the primary checkout. Both checkouts were measured at authoring time, and the primary case is the one that breaks:

```
# primary checkout — /Users/goos/MoAI/moai-adk-go
$ git rev-parse --git-common-dir                         → .git                              # RELATIVE
$ git rev-parse --path-format=absolute --git-common-dir   → /Users/goos/MoAI/moai-adk-go/.git
$ git rev-parse --absolute-git-dir                        → /Users/goos/MoAI/moai-adk-go/.git

# worktree — /Users/goos/.moai/worktrees/kanban
$ git rev-parse --git-dir                                 → /Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban
$ git rev-parse --git-common-dir                          → /Users/goos/MoAI/moai-adk-go/.git   # absolute here
$ git rev-parse --path-format=absolute --git-common-dir    → /Users/goos/MoAI/moai-adk-go/.git
```

"The parent of `.git`" is not a path. A rule phrased against the bare form therefore resolves the board root correctly from every worktree and incorrectly from the primary checkout — which is the one checkout the whole single-origin design points at. v0.1.0 carried exactly that phrasing in §A.3 and in REQ-KB-005, and it is repaired here.

**The correct rule is the one already in service** at `internal/hook/branch_guard.go`, in the function `isPrimaryCheckout`: a single probe `git rev-parse --path-format=absolute --git-dir --git-common-dir` emitting both paths absolute, in argument order — `--path-format=absolute` applies to every path flag that follows it, which is why the two flags share one call. Older git, and Apple Git, may reject that flag: `--path-format=absolute` requires git 2.31 (March 2021), so a fallback is required, and the existing one is `--absolute-git-dir` plus a `--git-common-dir` result normalized against the project directory when it is not already absolute. The board root is the parent of the **absolute** common directory so obtained. The function is cited **by name**, not by line number: v0.2.0 pinned lines 178 and 190 in normative text, and a requirement binding run-phase to line numbers in a file it does not own goes stale on the next edit. The line numbers survive in `research.md`, where the re-measurement discipline that governs every other figure applies to them too.

**But "reuse rather than re-derive" mandated something the code shape does not permit, and that is repaired here.** Measured: `isPrimaryCheckout(projectDir string) (bool, error)` is **unexported**, lives in package `hook`, and returns a *discriminant* — a boolean — not a path. There is no exported path-resolving helper to reuse, so v0.2.0's rule simultaneously mandated a copy and forbade one. The sibling settled this class already: `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018` requires that where a reused behaviour lives in an unexported symbol, the system either **extract** it into a package importing neither consumer, or **implement it as a contract the SPEC owns**. This SPEC takes the extraction branch and names its target: `internal/core/git`, whose `doc.go` declares it to be "Git repository operations for MoAI-ADK" — a read-only repository surface, which a git-directory resolution is — and which imports neither `internal/hook` nor `internal/cli`, so no cycle is created and neither existing caller's contract changes. The target was chosen by reading that `doc.go` rather than by resemblance of name: the sibling's own audit found it had named `internal/worktree`, whose `doc.go` scopes it to "working tree state guard primitives" and which its §C excluded. `internal/core/git` is excluded by no §C here.

Extracting it also buys something the boolean form cannot give: the extracted resolver returns a **path**, and a path is the thing whose correctness the existing caller's tests cannot speak to. `isPrimaryCheckout` compares two values for equality, and an equality is insensitive to an offset shared by both sides — a normalization error that shifted both paths identically would leave every existing test green while resolving the board root one directory wrong. That is why `AC-KB-002` judges the resolved path against a recorded value rather than inheriting the caller's confidence.

Five consequences follow, and none is glossed.

**(a) No `column` field is added to SPEC frontmatter.** The predecessor required registering `column` in the § Optional Fields table of `spec-frontmatter-schema.md` and its template mirror, and recorded a lint probe establishing that no lint rule was needed. **All of that is gone** — there is no new frontmatter field, so there is no schema registration, no template-mirror edit to that schema, and no lint question to answer. An auditor arriving from the predecessor will look for these; they are absent by decision, not by omission.

**(b) The column is still explicit, never derived.** §A.2's argument against derivation stands in full and is unweakened by the relocation. The board reads a recorded value; it computes nothing. Only the storage location moved.

**(c) Column history is not in git — the accepted cost.** The board state is gitignored, so there is no per-transition commit, no `git log` of column moves, and no bisectable column history. This is a real loss and it is accepted deliberately: the alternative purchased that history at the price of six pull requests per card and a merge-restores-stale-column defect. Where a durable audit trail of a card's movement is later wanted, it is a separate artifact under `.moai/state/kanban-board/` — not a frontmatter field, and not in this SPEC.

**(d) Two stores, disjoint domains, no overlap.** "Two sources of truth" is exactly what an auditor should challenge, so the disjointness is stated as a property rather than left as an implication:

| Store | Owns | Never carries |
|---|---|---|
| `.moai/state/kanban-board/` (primary checkout) | the column, the holder, the last-transition instant | `status`, artifact content, anything the SPEC lifecycle owns |
| `.moai/specs/SPEC-*/` (git-tracked) | `status`, the artifacts, the lifecycle | the column, the holder, the board's own record of anything |

Neither field is ever computed from the other, and neither store ever writes into the other's domain. The board reads `status`; it never writes it (REQ-KB-007). Nothing in `.moai/specs/` reads the board.

**(e) The board's directory is `kanban-board`, not `kanban`, because `kanban` is already taken by a different occupant under a different resolution rule.** `SPEC-KANBAN-RENAME-001` `REQ-KR-009` places the **session record** at `.moai/state/kanban/`, expressed through the existing path-segment constant — measured, `internal/factory/record.go` carries `stateDirSegments = []string{".moai", "state", "factory"}`, joined beneath `projectRoot`, which in a worktree is that worktree's own root. A session record is *deliberately* per-tree: it is session-scoped, best-effort, and `REQ-KR-010` records that an orphaned one is inert.

v0.2.0 put the board state in that same directory while resolving it at the **primary checkout**. One directory name, two occupants, two contradictory resolution rules, and no SPEC in the family stating the coexistence — a reader of either SPEC alone would conclude the directory means what their SPEC says it means, and an implementer reusing the existing path constant (as `REQ-KR-009` instructs) would land the board in each worktree's own tree, which is AP-1 arriving through the front door.

The paths are therefore **separated**. The session record keeps `.moai/state/kanban/` unchanged — `SPEC-KANBAN-RENAME-001` is not amended, and deliberately so: its constant, its migration decision, and its per-tree semantics are all correct for what it stores. The board takes `.moai/state/kanban-board/`, resolved at the primary checkout, through its own path-segment constant. Measured at authoring time, `.moai/state/` carries no entry named `kanban` or `kanban-board`, so the new name collides with nothing already there. The cost is that two adjacent directories differ by a suffix and a reader must know which is which; the alternative was two occupants sharing a name and differing by an invisible resolution rule, which is the failure this replaces.

### A.4 Where the two stores disagree

Because the board and the frontmatter answer different questions over the same card, some disagreement is not merely possible but *normal* — `backlog`/`plan` both correspond to `draft`, and `run`/`review` both to `in-progress`. A compatibility table therefore decides which pairings are legal. It is carried forward from the predecessor unchanged:

| `column` | Legal `status` values | Note |
|---|---|---|
| `backlog` | *no spec.md*, or `draft`, or `planned` | a card admitted before planning has no frontmatter at all; `planned` collides here exactly as `draft` does (§A.2) |
| `plan` | *no spec.md*, or `draft`, or `planned` | plan artifacts appear partway through this column; `planned` is admitted for the same reason |
| `run` | `in-progress` | `planned` is **not** admitted: it asserts work has not started |
| `review` | `in-progress` | the verify gate fires inside run-phase |
| `sync` | `in-progress`, `implemented`, `completed` | the matrix records `in-progress → implemented → completed` as **one** manager-docs sync commit, so `implemented` is transient and `completed` is the value usually observable while the card sits in `sync`. Omitting it — as v0.1.0 did — marks essentially every card inconsistent in the moment before it reaches `done`, and REQ-KB-008 then refuses to dispatch it |
| `done` | `completed` | |

The `sync` row is the one that changed at v0.2.0, and the reason is worth keeping visible: the omission was not a rare edge. Because the three-step transition rides a single commit, a board polling `status` during `sync` reads `completed` most of the time, so the illegal-pair path — designed for genuine disagreement — would have fired on the normal case and blocked the card one column short of `done`.

`archived`, `superseded`, and `rejected` appear in no row and are not omissions: no card is created for them at all (§A.2), so this table is never consulted for them.

A pair inside this table is consistent, even where it is ambiguous in the derived direction. A pair **outside** it — `column: done` against `status: draft`, say — is an inconsistency, and the board resolves it in the safe direction: it marks the card inconsistent, refuses to dispatch it, and surfaces both values. **It repairs neither.**

The refusal to repair is the load-bearing half. Repairing `status` would mean the board writing a lifecycle transition it does not own (REQ-KB-007). Repairing `column` would mean overwriting the record of an actor that observed something the board did not. A board that quietly reconciles a disagreement destroys the only evidence that something upstream is wrong.

What changed from the predecessor is only the read: the board now reads the column from its own store and the `status` from frontmatter, rather than reading both out of one file. The table, the illegal-pair behavior, and the no-repair rule are unchanged.

### A.4a Which tree the `status` is read from

The table above decides which pairings are legal. It says nothing about **where the `status` half is read**, and until v0.3.0 neither did any SPEC in the family — a grep across all four returns no statement of the read location. The omission is not a gap at the margin; it breaks the normal path for every card.

The board state is single-origin under the primary checkout (§A.3). A card's `status` transitions are not. §A.2 records that `draft → in-progress` is written by manager-develop on the **first run-phase commit**, and that commit lands on the card's **branch**, inside the card's worktree — `SPEC-KANBAN-WORKTREE-001` `REQ-KW-005` puts the card's `run`, `review`, and `sync` sessions in one worktree, and its `REQ-KW-007` holds that worktree until **both** the run and the sync pull requests have merged, which is after the card reaches `done`.

So for the whole interval a card sits in `run`, the primary checkout's copy of its `spec.md` still reads `draft`. `(run, draft)` is outside the table. `REQ-KB-008` therefore marks the card inconsistent and refuses to dispatch it — **every card, on the normal path**, with the board's own safety mechanism firing on the case it was built to let through.

The irony is worth keeping visible rather than quietly fixing, because it is the same blind spot twice. §A.3(3) rejected the `column:` frontmatter field partly on the ground that *a card's branch is cut early and merged late*, so a merge restores stale state. That argument is correct and it was never turned on the `status` the board reconciles against — the document argued the hazard in one paragraph and walked into it two sections later.

**The rule: the board reads a card's `status` from the card's branch, without checking it out.** Measured, a branch-side blob read requires no worktree and no checkout:

```
$ git show origin/docs/SPEC-CODEX-PHASE2-001-fork-resolution:.moai/specs/SPEC-CODEX-PHASE2-001/spec.md | grep -m1 '^status:'
status: draft

$ git show spec-kanban:internal/factory/record.go | head -1
// Package factory implements the state record that carries Factory Mode's
```

Both runs are from the primary checkout, and the second reads a **local** branch that a live worktree holds — refs are shared across every checkout of one repository, so no fetch and no remote round-trip is involved.

Two halves, because a card without a branch is not an edge case:

- **A card whose branch exists** — anything from the moment its worktree is created onward — has its `status` read from that branch. The branch is identified by `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`, which resolves a card's branch **by observation and by the SPEC identifier the branch carries, rather than by prefix**. That contract is consumed by name; this SPEC re-derives no branch naming.

**Why that consumption is not a declared dependency, stated here because the natural move is to declare it.** A requirement consuming a sibling's requirement looks like a dependency, and the first revision of this section made it one — promoting `SPEC-KANBAN-WORKTREE-001` out of `related_specs:` — which closed a cycle, because that sibling already names this SPEC in its own `dependencies:`. The sibling's author found the cycle, declined to break it by deleting its own edge, and recorded the analysis in its §A.4.0 with `AC-KW-001` written to observe it. The resolution is theirs and it is adopted here.

The two consumptions are of **different kinds**, and only one of them is a dependency in the sense the field means:

| Edge | Kind | Discharged by |
|---|---|---|
| the sibling → this SPEC (`REQ-KW-002` needs the card record's holder field) | **landing** — until the field exists in code, `REQ-KW-011`'s release path has nothing to release and cannot be implemented at all | the board's code having landed. Correctly declared, and it stays declared. |
| this SPEC → the sibling (`REQ-KB-020` needs `REQ-KW-003`'s identification rule) | **contract** — a rule that is readable from the sibling document today | citation. No code of that SPEC's need have landed for `REQ-KB-020` to be implementable. |

`dependencies:` declares what must land first. A contract consumption imposes no such ordering, so declaring it buys nothing and costs the cycle. It is therefore carried in requirement text — the identical shape `REQ-KB-017` already uses for `REQ-KS-006`, adopted in the same revision and for the same reason.

**The absence of that edge is a decision, not an oversight, and re-adding it is an error.** A later editor who notices `REQ-KB-020` naming a sibling requirement will reach for `dependencies:`; this paragraph exists to stop that. The mirror of the sibling's own rule applies: neither SPEC resolves a cycle by deleting an edge that records a real prerequisite. Here the edge to delete would be the sibling's landing dependency on this SPEC, and deleting it would leave the family with an undeclared prerequisite and no record of why — which is precisely the failure the split has already produced twice (§A.7, and the runtime-role gap §A.4a's own repair consumes).
- **A card with no branch yet** — `backlog` and `plan` cards, which have not been cut — has its `status` read from the **primary checkout**, which is the correct and only source. The same fallback covers the tail case: where a card's branch has been deleted after merge, the merged content is on the base branch in the primary checkout, and the two sources have converged by construction.

**A transition that has not been committed is not yet a transition.** A branch-side read observes committed state, so an edit sitting in a worktree's working tree is invisible to the board until it is committed. This is stated rather than left to inference because the alternative reading — that the board should somehow see uncommitted work — would require the board to read *inside* every worktree, which is precisely the per-tree resolution §A.3 rejects. The board observes what the repository records; a writer who has not committed has not yet told the repository anything, and the board is right to reflect that.

### A.5 The WIP limit and the session count are two knobs

The `run` column admits at most **2** cards concurrently — that is a property of the board, and it is this SPEC's. The number of deployed `run` sessions defaults to **1** and is raisable to **2** by configuration — that is a deployment knob, and it belongs to `SPEC-KANBAN-BOOTSTRAP-001`. This SPEC names the second knob only to forbid deriving one from the other (REQ-KB-010).

**With WIP 2 and one coder session, the second card enters `run` and waits there, unheld.** Admission is *not* gated on a session being free. An unheld card in `run` is a legal steady state, not an error and not a stall: it is dispatched the moment a session frees up.

The alternative — gating admission on a free session — reads as prudence and is the conflation arriving by the back door: it would make the effective WIP equal the session count, which is precisely what REQ-KB-010 exists to forbid. The two knobs would then be one knob wearing two names, and the board would appear to honor a WIP limit it had silently stopped enforcing.

"Unheld" is consequently a state the board must represent, and two independent causes converge on the same field: a card waiting for a free session (here) and a card whose holder was released after a session died (`SPEC-KANBAN-WORKTREE-001`) are the *same board state*. One field, two causes — which is why the holder is an emptiable value on the card rather than a separate column, and why no held or blocked column is introduced.

### A.6 Failing in the safe direction

A board state file can be found partially written or unreadable — a session killed mid-write, a truncated file, a corrupt encoding. The board resolves this by reporting the board as **unknown** and refusing every dispatch.

The tempting alternative is to treat an unreadable file as an empty board, because an empty board is a valid board and the code path already exists. It is the wrong direction: an empty board presented as accurate reports zero cards in `run`, which admits new cards past a WIP limit whose contents are unknown, onto work that may already be in flight. A refusal is loud and costs a re-read; a false empty board is silent and costs concurrent sessions working one card.

**But a refusal with no exit is a brick.** v0.1.0 stopped at the refusal, and that is a second defect wearing the first one's clothes: it required neither write atomicity nor any way out of the unknown state, so one session killed mid-write would have left the board refusing every dispatch **forever**, with no operation defined that could ever change the answer. "Fail safe" and "fail permanently" are not the same claim, and only the first was argued.

Two additions close it, and they are deliberately of different kinds. The first is **preventive**: every board write is atomic, so the partially-written file the unknown state exists to detect should not be producible in the first place (§A.7, REQ-KB-018). The second is **corrective**: a bounded, explicit recovery path — the sole writer, holding the board-wide lock, reconstructing or replacing the state file — through which the board leaves the unknown state (REQ-KB-013).

The recovery is **operator-visible and never automatic**. A board that silently self-repairs on the next read is a board that discards the evidence that something killed a session mid-write, which is the same objection §A.4 raises against silently reconciling a column/status disagreement — and it is worse here, because the repair would run on a file whose contents are by definition unknown. Recovery is therefore an act somebody performs and can see, not a fallback the read path takes. The refusal itself is unweakened: until recovery runs, every dispatch is still refused and no empty board is ever presented.

**Absent is a third case, and v0.2.0 assigned it to nobody.** `REQ-KB-013` binds a state file "partially written or unreadable". A file that **does not exist** is neither, and a grep across all four SPECs in the family returns no statement of it. Both readings the silence permits are wrong:

- Folding absent into *unknown* bricks the board on **first use**, before any card exists. The recovery operation would then be the only way to start a board at all — a first-run path that requires an explicit repair of something that was never damaged.
- Folding absent into *empty* opens, under another name, the exact door §A.6 spent this section closing. "An unreadable file is not an empty board" and "a missing file is an empty board" are one substitution apart, and an implementation that reaches the empty-board path through either route is indistinguishable from the outside.

The distinction is therefore decided rather than inferred: an **absent** state file is a legitimately empty board — no cards, dispatch permitted, no recovery required — and the store's directory is created on that path. An **unreadable** one is unknown, and the refusal stands. What separates them is not a heuristic but the operating system's own answer to whether the file exists; the two are distinguishable without judgment, which is why they can be assigned different behavior safely.

The directory nobody created is the same omission one layer down. `REQ-KB-018`'s same-directory temp file requires the board state directory to exist; no requirement created it; and it is gitignored (§A.3(1)), so it cannot ship in the repository and be there on arrival. The **sole writer** creates it — consistent with `REQ-KB-017`, since the writer is the only actor permitted to bring the store into existence, and the reused `internal/spec/lock.go` already demonstrates the shape by creating `.moai/state/` before taking its lock.

**And "bounded" needs a bound.** `REQ-KB-013` calls the recovery bounded, and v0.2.0 defined no bound anywhere — an assertion doing the work of a specification. The bound is stated in two parts. **In extent**: the recovery touches the board state file and nothing else — it writes no SPEC frontmatter (REQ-KB-007 holds through recovery as everywhere else), removes no worktree, and moves no card that it can still read. **In effect**: it terminates in one invocation with a definite verdict — recovered, or not recoverable — and never re-enters itself or retries.

The second half is the one with teeth, because it constrains what a recovery may *produce*. A "replacement of the state file" is, read literally, permission to write an empty board — after which reads succeed, the board reports zero cards in `run`, and new cards are admitted onto work that may be in flight. That is the harm this section argues against, reached through a door labelled *explicit*. So a recovery that cannot reconstruct a card **records what it could not recover, durably, and surfaces it to the operator** rather than silently dropping it. An operator who invoked a recovery and got a working board back is entitled to know whether they got their board back or a new one.

### A.7 Who writes the board, and under what exclusion

The predecessor's `REQ-KM-044` carried two things in one requirement: *where* the column is recorded, and *who is allowed to write it*. §A.3 rejects the first — the `column:` frontmatter field is gone, on measured grounds. The second was never rejected by that argument at all: changing a store's location says nothing about its ownership. It was nevertheless deleted with it, and the loss is measurable — at v0.1.0, `grep -rc 'sole writer'` and `grep -rc 'single writer'` over all three sibling SPECs reported **0** in every file, and `grep -c 'atomic'` over this document reported **0**.

Both siblings then disclaimed the gap toward each other. This SPEC's §C said it "names no actor"; `SPEC-KANBAN-BOOTSTRAP-001` §C says it "decides who is told about a card and by what message". Being told about a card is not writing the card. Nobody owned the write, which is not a boundary — it is a hole with a boundary drawn around each side of it.

Three properties close it, and they are separable because each fails independently.

**(1) One writer.** Exactly one role — the session occupying the `lead` role — writes the board state file. Every other session reads it. This is a property the implementation *enforces*, not one it documents: a rule that only a comment states is a rule the second worker session breaks without noticing. Who the lead is, and how it comes to be the lead, stays with `SPEC-KANBAN-BOOTSTRAP-001` (`REQ-KS-004`); this SPEC names the owner of the write, not the election that produces it.

**(2) Atomic writes.** Every board write goes to a temporary file **in the same directory** as the target, followed by an atomic rename over it. Same-directory is not a stylistic preference: `rename(2)` is atomic only within one filesystem, so a temp file in the system temp directory can land on a different device and degrade the rename into a copy — which is precisely the partially-observable write the atomicity was bought to prevent. A reader therefore observes either the whole previous board or the whole new one, never a prefix.

**(3) Board-wide exclusion.** The lock covering a board mutation spans the *entire* read-modify-write of the **whole board** — read, decide, write — not a card. For board mutations this **supersedes** card scope, and the reason is REQ-KB-009: a card-scoped lock cannot make a board-wide invariant hold. With WIP 2, two concurrent assignments of two *different* cards each take their own card's lock, each read "1 card in `run`", each pass the check, and each write — and the board ends at WIP 3, with neither writer having done anything wrong under its own lock. The WIP limit of REQ-KB-009 is therefore **only sound under board-wide exclusion**; at v0.1.0 it was a bound the design could not deliver.

The sibling's `REQ-KW-013` card-scoped lock remains valid for its own purpose — holder assignment on a single card, where card scope is exactly the right granularity. This is an addition, not a replacement. What is asserted is narrower and stronger: a card-scoped lock is *insufficient for a board mutation*, and reaching for the one already there is the failure mode this exists to name.

**(3a) The lock the board reuses can outlive its holder, and v0.2.0 re-created the brick it had just removed.** `REQ-KB-019` mandates reuse of the `internal/spec/lock.go` family. That family's Windows substrate says of itself, in its own header:

> Stale lock detection (PID + timestamp embedded) is a post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring manual `del .moai/state/spec-close-*.lock`.

A `lead` killed while holding the board-wide lock therefore leaves an artifact that blocks **every** future board mutation, with no operation defined that could change the answer. That is precisely the shape §A.6 names — a refusal with no exit — arriving through the lock the same revision introduced. AP-22 was written against the corrupt-file path and never turned on the lock; the same blind spot as §A.4a, in a different section.

The hazard is **platform-asymmetric**, and the asymmetry is recorded rather than smoothed over, because it decides how likely this is to be noticed. Measured: the Unix substrate holds `flock(2)` on an open descriptor and releases it when the descriptor closes, which the kernel does on process exit — so a killed `lead` on Unix leaves a lock *file* but no lock, and the next mutation acquires cleanly. The repository's own `.moai/state/` carries fourteen such orphaned `spec-close-*.lock` files, all zero-length, all harmless. On Windows the substrate is atomic-create-exclusive with no such release, so the artifact **is** the lock and the block is permanent. A defect that never reproduces on the developer's machine and always reproduces on a user's is not a lesser defect.

The sibling solved this shape for its own lock: `SPEC-KANBAN-WORKTREE-001` `REQ-KW-014` records the creating process's identity in the artifact and provides a bounded clearing operation conditioned on that process being **positively observed absent**, as an explicit operator-visible act. That shape is ported here (`REQ-KB-023`) rather than copied — a different lock, a different scope, and one hazard the sibling's own audit surfaced in exactly this clearing act, which is not reproduced.

**The hazard not to reproduce is a check-then-unlink race.** Between reading the artifact's recorded identity and removing it, the lock can be legitimately released by its owner and **re-acquired by a live process**. The clearer then unlinks a valid lock, and two sessions enter the critical section — the clearing operation having caused precisely the concurrency the lock existed to prevent, while every step of it looked correct in isolation. The removal is therefore conditioned not only on the recorded process being absent but on the artifact still being **the same artifact that was inspected**, so a recreate between the two steps aborts the clear rather than completing it.

Properties (1) and (3) are related but not the same, and collapsing them is a real hazard. A sole writer removes cross-*role* contention; it does not remove contention within the writer, nor guarantee that a reader observing an in-flight mutation sees a coherent board, nor survive a second lead process appearing during a handoff. The lock is what makes the read-modify-write indivisible; sole ownership is what keeps the set of writers knowable. Each without the other leaves a live failure.

---

## §B. Requirements (GEARS)

> Requirement count: 23 (`REQ-KB-001` … `REQ-KB-023`). Acceptance criteria: 24 (`AC-KB-001` … `AC-KB-024`). Tier L ceiling: 25 requirements, 25 acceptance criteria — both are met with two and one to spare respectively. Promoted from Tier M at v0.2.0 to admit `REQ-KB-017` … `REQ-KB-019`, which the Tier M ceiling of 16 could not hold; grown again at v0.3.0 by `REQ-KB-020` … `REQ-KB-023`. Each of those four was authored as a separate requirement rather than folded into a neighbour, because folding separable obligations together is the F1 defect this SPEC exists to have repaired and the ceiling is not a reason to repeat it. Tier L is the top tier, so a further finding of this size would be reported rather than absorbed.

### B.1 Preconditions

**REQ-KB-001** — The implementation shall write every renamed identifier in its post-`SPEC-KANBAN-RENAME-001` form, and shall introduce no occurrence of `factory` in any identifier, path, environment variable, sentinel, preset name, or prose it authors.

**REQ-KB-002** — The implementer shall verify at preflight that the rename prerequisite has landed on the base branch, and **when** the renamed package is found absent, the implementer shall halt and surface the absence rather than proceeding against an unlanded prerequisite or performing the rename itself.

### B.2 The six columns and the card

**REQ-KB-003** — The board shall carry exactly six columns, ordered `backlog`, `plan`, `run`, `review`, `sync`, `done`, expressed as a closed enumeration in `internal/kanban` with no seventh value and no operator-extensible column set.

**REQ-KB-004** — The board shall record, for each card, the SPEC identifier the card targets, the column the card occupies, the session identifier currently holding the card, and the instant of the card's last column transition; and a card held by no session shall record an empty holder rather than a synthesized one.

### B.3 Single-origin board state

**REQ-KB-005** — The board state shall persist beneath the **primary checkout's** `.moai/state/kanban-board/`, expressed through a path-segment constant of its own and distinct from the session-record directory `.moai/state/kanban/` that `SPEC-KANBAN-RENAME-001` `REQ-KR-009` resolves beneath each tree's own root, which this SPEC shall not reuse, relocate, or amend; the board root shall be resolved by every session as the parent of the **absolute** common git directory obtained by the resolution `internal/hook/branch_guard.go` carries in `isPrimaryCheckout` — the single probe `git rev-parse --path-format=absolute --git-dir --git-common-dir`, falling back for git older than 2.31 to `--absolute-git-dir` together with a `--git-common-dir` result normalized against the project directory when it is not already absolute; the bare `--git-common-dir` form shall not be used alone, because it returns a repository-relative path in the primary checkout; and no session shall read or write a board state path resolved relative to its own working tree.

Because that resolution lives in an **unexported** symbol of a package the board must not import, and returns a boolean discriminant rather than a path, the reuse shall take the extraction disposition of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018`: the implementation shall extract it into a package importing neither consumer — `internal/core/git`, whose `doc.go` declares its scope to be git repository operations and which imports neither `internal/hook` nor `internal/cli` — rather than copying it, shall have the extracted symbol return the resolved **path**, and shall change no existing caller's contract. The implementer shall confirm the target's declared purpose by reading its `doc.go` and confirm it is excluded by no §C of this SPEC, rather than selecting it by resemblance of name.

**REQ-KB-006** — The board shall read each card's column as a recorded value, and shall not derive, infer, or recover it from a SPEC's frontmatter `status`, from any `progress.md` marker, or from any other observable.

**REQ-KB-007** — The board shall write no field of any SPEC's frontmatter: it shall add no `column` field, shall write no `status` value, shall not extend the canonical 8-value status enum, and shall write no `status` transition that the Status Transition Ownership Matrix assigns to another agent.

### B.4 Column-to-status consistency

**REQ-KB-008** — The board shall accept every `(column, status)` pairing the compatibility table of §A.4 admits, and **when** a pairing outside that table is observed, the board shall mark the card inconsistent, report it as not dispatchable, and surface both values — repairing neither the column nor the status.

**REQ-KB-020** — The board shall read a card's frontmatter `status` from the card's **branch**, without checking that branch out, identifying the branch by `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` — by observation and by the SPEC identifier the branch carries — rather than by re-deriving a branch name here; **where** a card has no branch, the board shall read its `status` from the **primary checkout**, which covers both a card not yet cut and a card whose branch was deleted after merge; the board shall read committed state only, so that a transition written but not yet committed is not yet observable and is not treated as one; and the board shall not read a card's `status` from the primary checkout while that card's branch exists, since a card's `status` transitions are written on its branch and its worktree survives until both of its pull requests merge, which would make the primary-side value stale for the whole interval the card spends in `run`. This consumption is a **contract** dependency and not a landing one: what is consumed is `REQ-KW-003`'s identification *rule* — observe the branch the worktree reports, recognize it by the SPEC identifier it carries, re-derive no name — which is readable from that document (§A.2, §A.2.1) with none of that SPEC's code having landed, so it is discharged by citation exactly as `REQ-KB-017` discharges `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`. The sibling therefore stays in `related_specs:` and shall not be promoted to `dependencies:`, because that sibling declares this SPEC among its own `dependencies:` on a genuine landing need and the reverse edge would close a cycle (§A.4a).

### B.5 WIP, admission, and the unheld state

**REQ-KB-009** — The board shall admit at most two cards to the `run` column concurrently, and **when** a transition into `run` would make the third, the board shall refuse the transition with a named error and leave the board unchanged rather than silently queueing it.

**REQ-KB-010** — The board shall treat the `run` WIP limit and the deployed coder-session count as independent values, and shall derive neither from the other.

**REQ-KB-011** — The board shall admit a card into `run` whenever the WIP limit permits it, without regard to whether a session is free to hold it, and a card in `run` with no holder shall be a legal steady state rather than an error or a stall.

**REQ-KB-012** — The board shall treat `backlog` as a queue with no owning session and `done` as terminal with no owning session, and shall report a card in either column as not dispatchable.

### B.6 Failing in the safe direction

**REQ-KB-013** — **When** the board state is found partially written or unreadable, the board shall report the board as unknown and refuse every dispatch, and shall not present an empty board as an accurate one; and the board shall provide a bounded recovery operation by which the unknown state is left, being a reconstruction or replacement of the state file performed by the sole writer of REQ-KB-017 while holding the board-wide lock of REQ-KB-019, invoked as an explicit operator-visible act and never as a silent repair taken by the read path — so that the unknown state is escapable rather than terminal, while the refusal to dispatch and the refusal to present an empty board both hold unchanged until the recovery has run.

**REQ-KB-021** — The board shall distinguish an **absent** state file from an **unreadable** one: **when** the state file does not exist, the board shall report a legitimately empty board, permit dispatch, and require no recovery, and the sole writer of REQ-KB-017 shall create the state directory on that path, since the directory is gitignored and therefore cannot be present on arrival; and **when** the state file exists but cannot be read or parsed, the board shall report unknown under REQ-KB-013 and shall not reach the empty-board result through the absent path.

**REQ-KB-022** — The recovery operation of REQ-KB-013 shall be bounded in extent and in effect: in extent, it shall modify the board state file alone, writing no SPEC frontmatter, removing no worktree, and moving no card it can still read; in effect, it shall terminate in a single invocation with a definite verdict and shall neither re-enter itself nor retry; and **when** it cannot reconstruct part or all of the prior board, it shall record what it could not recover in a durable artifact and surface that record to the operator, rather than presenting the replacement it produced as the board that was lost.

### B.7 Who writes the board, and under what exclusion

**REQ-KB-017** — The board state file shall be written by exactly one role, the session occupying the `lead` role, and every other session shall read it and shall not write it; the implementation shall enforce this rather than document it, so that no write path to the board state file is reachable from a session in any other role; and this SPEC shall define no part of how the `lead` role is elected or assigned, which belongs to `SPEC-KANBAN-BOOTSTRAP-001` (`REQ-KS-004`). The runtime half of this requirement — refusing a board write from a session in another role — presupposes that the caller's role is readable at runtime; that contract is `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, which requires each session to **declare the role it occupies** as a datum distinct from its launch label and resolvable by a session that is not the `lead`, and it is consumed here by name: this SPEC shall define no role-declaration mechanism, shall derive no role from a session identifier or a launch label, and shall not restate that contract's content.

**REQ-KB-018** — Every write of the board state file shall be performed as a write to a temporary file created **in the same directory as the target file**, followed by an atomic rename of that temporary file over the target, reusing the repository's existing `internal/atomicfile` replacement primitive rather than adding a second one; the temporary file shall not be created in the system temporary directory or in any other directory, because `rename(2)` is atomic only within a single filesystem and a cross-device rename degrades into a copy — reintroducing the partially-observable write this requirement exists to prevent; and no reader shall be able to observe a partially written board.

**REQ-KB-019** — Every mutation of the board state shall be serialized beneath an advisory lock whose scope is the **whole board**, held across the entire read-modify-write — the read of the board, the admission or transition decision, and the write — and the board shall obtain that lock by reusing the repository's existing cross-process per-scope lock pattern (`internal/spec/lock.go` and its platform counterparts) rather than adding a third locking mechanism; for a board mutation this board-wide scope **supersedes** card scope, and the WIP bound of REQ-KB-009 shall be enforced only beneath it, since two concurrent mutations of two different cards each holding only that card's lock would each observe the bound satisfied and each write, admitting a third card to `run`; and the card-scoped lock of the sibling's `REQ-KW-013` shall remain in force for holder assignment, this requirement adding an exclusion rather than replacing one.

**REQ-KB-023** — Because the lock family REQ-KB-019 reuses performs no stale-lock detection on its Windows substrate, so that an artifact left by a killed `lead` would block every subsequent board mutation permanently, the board shall record in the board-wide lock artifact the identity of the process that created it, and shall provide a bounded clearing operation that removes such an artifact **only when** that recorded process is positively observed absent — never on the mere age of the artifact, and never as a step the acquire path takes on its own — the clearing being an explicit operator-visible act that reports what it removed; and the removal shall additionally be conditioned on the artifact still being the same artifact that was inspected, so that a lock legitimately released and **re-acquired by a live process** between the inspection and the removal aborts the clear rather than being unlinked, which would admit two writers to the critical section the lock exists to hold.

### B.8 Mirror, neutrality, and verification

**REQ-KB-014** — The implementer shall edit template source under `internal/template/templates/` before its local counterpart, shall run `make build`, and shall commit the regenerated `internal/template/catalog.yaml`; and **while** applying a change to a mirrored pair, shall preserve that pair's measured relationship, so that a pair measured byte-identical remains byte-identical and a sanitized pair retains exactly the content its template side strips. A sanitized pair becoming byte-identical is a failure, not a convergence.

**REQ-KB-015** — No file authored or modified under `internal/template/templates/` shall contain a SPEC identifier, a REQ or AC token, an internal date, or a commit SHA.

**REQ-KB-016** — The verification shall run the full test suite rather than an affected-packages subset, because a prior run-phase in this repository missed a cross-cutting template guard by testing narrowly.

---

## §C. Exclusions

### Out of Scope — the worktree sibling

- Worktree creation, disposal, the disposal gating contract, refused-removal handling, and orphan classification. All belong to `SPEC-KANBAN-WORKTREE-001`.
- Stall detection, the stall threshold and its default, and the release of a card's holder when its session dies. This SPEC defines the *unheld state* the release resolves to (§A.5, REQ-KB-011); it defines no detector and no release trigger.
- Mutual exclusion of card assignment across sessions.

### Out of Scope — the bootstrap sibling

- Preflight beyond the rename gate of REQ-KB-002, the session roles and topology, the bootstrap surface and the entry switch, the topology configuration and its `moai init` question, the quorum bound, the dispatch protocol, backend selection, and the coder-session internal chain. All belong to `SPEC-KANBAN-BOOTSTRAP-001`.
- The deployed coder-session count and its configuration. Named here only as the knob REQ-KB-010 forbids deriving the WIP limit from.
- **How the `lead` role is elected, assigned, or handed over**, and by what message a card's move is communicated. This SPEC defines which moves are legal and — since v0.2.0 — **who is permitted to write the board** (REQ-KB-017: the `lead`, solely). It names **no transport** and elects no lead.

- **The role-declaration mechanism.** `REQ-KB-017`'s runtime refusal consumes `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the declaration exists, is distinct from the launch label, and is resolvable from a session that is not the `lead`. Whether it rides the launch command, the session registry, or the peer-discovery output is that SPEC's run-phase decision and is not re-decided here, nor is the contract restated. This SPEC is a **consumer** of it, which is why `SPEC-KANBAN-BOOTSTRAP-001` remains in `related_specs:` and not in `dependencies:`: that sibling already carries this SPEC in its own `dependencies:`, so the reverse edge would state a cycle. A runtime dependency in requirement text is the shape the sibling used for the same situation, and it is the shape used here.

  This line replaces the v0.1.0 wording "it names no actor and no transport", which is now false as stated and was the disclaimer half of the gap §A.7 describes: this SPEC disclaimed the actor, `SPEC-KANBAN-BOOTSTRAP-001` §C disclaimed everything but notification, and the write ended up owned by neither. Write authority is this SPEC's; election and transport remain the sibling's.

### Out of Scope — the frontmatter and the status vocabulary

- A `column` field in SPEC frontmatter, its registration in the schema's § Optional Fields table, and its template mirror. Rejected and removed at v0.1.0 on measured grounds (§A.3); the predecessor's schema-registration requirement and its lint probe are withdrawn with it.
- Adding a `reviewed` status, or any other value, to the canonical 8-value `status` enum. The `review` column's absence of a status counterpart is a measured fact about the vocabulary (§A.2), resolved at the board layer, never by extending an enum this SPEC does not own.
- Reassigning any row of the Status Transition Ownership Matrix.
- Amending `SPEC-KANBAN-RENAME-001` `REQ-KR-009`, its `.moai/state/kanban/` session-record path, or the path-segment constant that expresses it. §A.3(e) resolves the collision by moving **this SPEC's** store to `.moai/state/kanban-board/`; the sibling's path is deliberately unchanged, because per-tree resolution is correct for what it stores.
- Deriving a card's branch name, or any branch-naming scheme. `REQ-KB-020` consumes `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`'s resolution by name and re-derives nothing. That consumption is a **contract** dependency carried in requirement text; promoting the sibling into this SPEC's `dependencies:` is out of scope and would close a cycle (§A.4a).
- Amending `SPEC-KANBAN-WORKTREE-001`'s `dependencies:` — in particular, removing its entry for this SPEC. That entry records a real **landing** prerequisite (`REQ-KW-002` needs the card record's holder field to exist), and deleting it to break a cycle would leave the family with an undeclared prerequisite and no record of why. The cycle is resolved on **this** side, by not declaring the contract edge.

### Out of Scope — a seventh column

- A `test` column. Rejected: unit tests belong to the cycle inside `run`, and a separate column would encode test-after.
- An operator-extensible column set. The six are fixed (REQ-KB-003).
- A held or blocked column for a card with no holder. The unheld state already exists and serves both of its causes (§A.5).

### Out of Scope — the board as a product surface

- A web view, a TUI, a live dashboard, or any rendering of the board. The read-only web board is a separate line of work.
- A durable git-visible history of column transitions. Its absence is the accepted cost of §A.3(c), not a deferred feature.

---

## §D. Verification surfaces

### D.1 The single-origin rule is an absence claim (REQ-KB-005)

That every session resolves the *primary* checkout is two claims, and they are checked separately because one of them is negative.

**The positive half** is a resolution equality, and it is *measured* rather than reasoned about: from inside a worktree, the resolved board root must equal a known primary-checkout path. Reasoning from git's documented behavior of `--git-common-dir` would establish a fact about the documentation, not about the tree the code will actually run in.

**The negative half** — that *no* path resolves relative to the current tree — is an absence claim over every call site, so it is checked as one: the tree is scanned for a board-state path constructed from the working directory, the repository root as the session sees it, or any other tree-relative anchor, and must yield none. A scan asserting an absence with no demonstrated ability to fire is indistinguishable from a broken command, so a **positive control** is required: a deliberately introduced tree-relative read that the same scan reports, run once and recorded.

### D.2 The compatibility table is checked on its illegal rows (REQ-KB-008)

The table of §A.4 is exercised table-driven over **every** pairing — the legal rows and the illegal ones alike. Each legal pairing must be accepted; each illegal pairing must mark the card inconsistent, report it not dispatchable, and leave both the board record's column and the SPEC's `status` byte-unchanged on disk.

A test asserting only the legal rows fails this surface. It never demonstrates that an illegal pairing is refused, and refusal is the entire behavior — an implementation that accepts everything passes a legal-rows-only test perfectly.

### D.3 The board writes no frontmatter (REQ-KB-007)

Also an absence claim, and checked in the same shape: the board package is scanned for any write to a SPEC frontmatter key, with a positive control establishing the scan can fire. Additionally, no string literal outside the canonical 8-value enum may be introduced as a status, and the enum is read from `spec-frontmatter-schema.md` rather than from a copy in this SPEC — a copied enum is a second source of truth that begins drifting on the day it is written.

### D.4 The WIP refusal must be conditional (REQ-KB-009)

A refusal that fires unconditionally passes a naive refusal test. The criterion therefore pairs the refusal with a **positive control**: with one card in `run`, the same transition succeeds. Only the pair establishes that the bound is 2 rather than 0.

### D.5 The sole-writer rule is an absence claim (REQ-KB-017)

Checked in the same shape as §D.1 and §D.3, because it is the same kind of claim: *no* write path to the board state file is reachable from a session in a role other than `lead`. The tree is scanned over every call site that opens, creates, truncates, or renames onto the board state path, and each must be reachable only from the writer the requirement names.

A **positive control** is required and is not optional here: a deliberately introduced board-state write from a non-`lead` call site must be reported by the same scan, run once and recorded. Without it the scan's zero result is indistinguishable from a scan that cannot fire — and the whole defect this requirement repairs is a rule that existed in prose while nothing enforced it.

### D.6 The board-wide lock is checked across processes (REQ-KB-019)

The lock must be exercised by a concurrency test using **separate processes**, not separate goroutines. The distinction is not pedantry: an in-process mutex passes a same-process test perfectly and provides nothing in production, where the `lead` and the worker sessions are distinct OS processes.

The repository already carries the demonstration. `internal/lockfile/lockfile_windows.go` is a `map[string]*sync.Mutex` whose own package comment states that concurrent writes across **different** OS processes are *not* protected, and that the limitation is preserved deliberately — "do NOT silently 'upgrade' this to LockFileEx". Its sibling `SPEC-KANBAN-WORKTREE-001` selected the `internal/spec/lock.go` family over it for exactly that reason, and REQ-KB-019 reuses the same family. A test that would pass against the in-process implementation is therefore a test that does not measure this requirement.

The scenario is the one §A.7(3) names, run for real: two processes concurrently transition two **different** cards into `run` against a board already holding one. Exactly one must succeed; the final board must hold two cards in `run`, never three. A card-scoped lock passes nothing here — both processes take different locks — which is what makes this the criterion that distinguishes board scope from card scope.

### D.7 Atomicity is observed from a reader, not asserted from the writer (REQ-KB-018)

Two checks, because the requirement has two failure modes.

The **same-directory** property is checked statically: the temporary file's directory must equal the target's directory, so a cross-device rename is unreachable. Asserting only that a rename occurs misses this entirely — a temp file in the system temp directory still renames, it just stops being atomic.

The **atomicity** property is checked from a concurrent reader: while writes proceed, a reader repeatedly reading the board state observes only whole boards — every read is either the complete prior state or the complete new one, never a prefix and never a parse failure. The reader is the right vantage point because the writer cannot observe its own torn write.

### D.8 The recovery path is bounded and does not weaken the refusal (REQ-KB-013)

Three properties, and the third is the one that is easy to lose. First, from the unknown state, the recovery operation returns the board to a readable state. Second, until it runs, every dispatch is still refused and no empty board is presented — the exit does not soften the failure mode it exits from. Third, the recovery is **not** reached by the read path: a read of an unreadable board repeated any number of times must still report unknown, never silently repair. The third is what separates a recovery from a fallback, and a criterion omitting it would pass an implementation that self-repairs on read.

### D.9 Mirror delta preservation (REQ-KB-014)

For each mirrored pair this SPEC touches, the `diff` taken before the change must equal the `diff` taken after, once the change's own token substitutions are applied to it. The classification of each pair is **time-varying** and is re-measured at run-phase rather than trusted from this document.

### D.10 Neutrality (REQ-KB-015)

`internal/template/internal_content_leak_test.go` and `.github/workflows/template-neutrality-check.yaml` are the mechanical authority, and their exit codes are the verdict. This SPEC adds one directed check but does not reimplement the guard's regex — a hand-rolled reimplementation without the guard's exemption list is a false-failure machine.

### D.11 The status read is judged on the divergence, not on the agreement (REQ-KB-020)

The two trees agree for most of a card's life, and a criterion built on an agreeing pair cannot fail: an implementation reading the primary checkout passes it exactly as one reading the branch does. The surface is therefore the **interval where they differ**, which is the interval the defect lived in — a card in `run` whose branch-side `status` is `in-progress` while the primary checkout's copy still reads `draft`. The board must report that card consistent and dispatchable. Reporting it inconsistent is the v0.2.0 behavior, on the normal path, for every card.

The second half is the no-branch case: a card in `backlog` or `plan` reads its `status` from the primary checkout and is judged by the same table. A criterion covering only the first half passes an implementation that has no fallback and fails on the first unplanned card.

### D.12 The stale-lock clear is judged on what it refuses (REQ-KB-023)

Three observations, and the first two are the ones an implementer would write unprompted: a lock artifact left by a dead process is cleared, and a lock artifact held by a live process is not. Both are necessary and neither is sufficient, because both are satisfied by an implementation that reads the recorded identity once and then unlinks.

The third is the race, and it is the reason this is a verification surface rather than a note: between the inspection and the removal, the artifact may be released by its owner and **re-acquired by a live process**. The criterion constructs exactly that interleaving and requires the clear to abort. An implementation passing the first two observations and failing this one has a clearing operation that causes the concurrency the lock prevents — a repair that is worse than the brick it repairs, and invisible to any test that does not construct the window.

### D.13 Absent and unreadable are distinguished by observation (REQ-KB-021)

Both states are checked, and they are checked **against each other** rather than separately, because the defect is a conflation and a conflation is only visible in the pair. From an absent file the board reports an empty board and permits dispatch; from an unreadable file of the same size it reports unknown and refuses. A test exercising one alone passes an implementation that treats both the same way, in whichever direction it happens to have chosen — and both directions are wrong (§A.6). The directory-creation half rides here: after the absent path, the state directory exists.

---

## §E. Cross-references

- `SPEC-KANBAN-RENAME-001` — the prerequisite rename. Its `dependencies:` entry is this SPEC's blocking gate (REQ-KB-002).
- `SPEC-KANBAN-WORKTREE-001` — the sibling owning worktree lifecycle, stall detection, holder release, and assignment exclusion. It consumes this SPEC's unheld state; this SPEC defines no part of it. **Deliberately in `related_specs:` and not in `dependencies:`** — the absence is a decision, not an oversight, and re-promoting it is an error. `REQ-KB-020` consumes its `REQ-KW-003` (a card's branch, resolved by observation and by the SPEC identifier the branch carries) as a **contract** dependency: the identification rule is readable from that document now, so nothing of that SPEC's need land first, and the consumption is carried in `REQ-KB-020`'s own text. The reverse edge would close a cycle, because that sibling declares this SPEC among its own `dependencies:` on a **landing** need (`REQ-KW-002` requires the card record's holder field to exist in code). That edge is the correctly-declared one and stays; see §A.4a for the two-kinds analysis and the sibling's §A.4.0 for the same finding from its side. Its `REQ-KW-005` (one worktree per card, serving `run`, `review`, and `sync`) and `REQ-KW-007` (disposal only after both pull requests merge) are the two facts that make the primary-side `status` stale for a card's whole run interval (§A.4a); its `REQ-KW-014` is the stale-lock shape `REQ-KB-023` ports; and its `REQ-KW-018` is the extraction disposition `REQ-KB-005` takes.
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the role-declaration contract `REQ-KB-017`'s runtime half consumes: each session declares the role it occupies, as a datum distinct from its launch label and resolvable by a session that is not the `lead`. Widened into that shape at that SPEC's v0.3.0, which closed the same runtime-role-resolution gap this SPEC's audit reached from the other side. Consumed by name; not restated, and not in `dependencies:` — that sibling lists this SPEC in its own, so the reverse edge would be a cycle.
- `SPEC-KANBAN-RENAME-001` `REQ-KR-009` — the **session-record** path `.moai/state/kanban/`, resolved beneath each tree's own root through `internal/factory/record.go`'s existing path-segment constant. It is the occupant this SPEC's board state was colliding with, and it is deliberately unchanged; the board moved instead (§A.3(e)).
- `internal/core/git` — the extraction target `REQ-KB-005` names for the git-directory resolution, chosen by reading its `doc.go` ("Git repository operations for MoAI-ADK") and confirming it imports neither `internal/hook` nor `internal/cli`.
- `internal/spec/lock_windows.go` — the substrate whose own header records that stale-lock detection is a post-MVP enhancement and that cleanup is manual. It is the measured reason `REQ-KB-023` exists, and the reason §A.7(3a) records the hazard as platform-asymmetric: the Unix substrate releases its `flock(2)` on process exit, so the artifact there is inert.
- `SPEC-KANBAN-BOOTSTRAP-001` — the sibling owning preflight, topology, bootstrap, configuration, quorum, dispatch, and backend selection. It consumes this SPEC's board model; this SPEC defines no part of it.
- `SPEC-KANBAN-MULTISESSION-001` — the superseded 59-requirement predecessor this SPEC is split out of. Its §A.4 and §A.5 are carried forward; its `column`-in-frontmatter decision is rejected (§A.3).
- `SPEC-FACTORY-MODE-001` — the closed SPEC that delivered the single-session chain the kanban system extends. Preserved as a historical record; not amended.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum + § Status Transition Ownership Matrix — the vocabulary §A.2 and §A.4 reconcile against, and the matrix REQ-KB-007 refuses to write into.
- `internal/hook/branch_guard.go` — the existing consumer of the git-dir versus git-common-dir discriminant REQ-KB-005 reuses, in its unexported `isPrimaryCheckout`. Cited by symbol name; the line anchors and the fallback-forcing test pattern (`execCommand` injection, whose own comment records that direct invocation of the fallback is a vacuous pass) are recorded in `research.md`, where the re-measurement discipline applies.
- `internal/atomicfile` (`replace.go`, `replace_unix.go`, `replace_windows.go`) — the existing write-temp-then-rename primitive REQ-KB-018 reuses rather than reimplementing. Its package comment records that POSIX `rename(2)` replaces the destination atomically while other processes hold it open, and that Windows needs the platform handling that package already carries.
- `internal/spec/lock.go`, `lock_unix.go`, `lock_windows.go` — the cross-process per-scope lock pattern REQ-KB-019 reuses, the same family the worktree sibling's `REQ-KW-013` selected.
- `internal/lockfile/lockfile_windows.go` — the in-process mutex REQ-KB-019 does **not** use, cited because its own package comment states the cross-process gap and forbids upgrading it. It is the measured reason §D.6 requires separate processes.
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004` — the topology requirement that defines the `lead` role REQ-KB-017 names as the sole writer. This SPEC names the owner; that one elects it.
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — the doctrine that discriminant serves, named here to record that this SPEC adds a consumer rather than a mechanism.
- `.claude/rules/moai/workflow/cross-session-messaging.md` — lands ahead as its own pull request. Context only: the preflight requirement for it belongs to `SPEC-KANBAN-BOOTSTRAP-001`, not here.
- `CLAUDE.local.md` §2 (Template-First), §14 (env constants), §25 (Template Internal-Content Isolation).
