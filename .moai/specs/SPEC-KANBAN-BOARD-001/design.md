---
id: SPEC-KANBAN-BOARD-001
title: "Design — six-column kanban board model with a single-origin board state store"
version: "0.4.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, design, decisions, rejected-alternatives, sole-writer, atomicity"
tier: L
---

## §A. What this file is

The decisions this SPEC rests on, each with the alternative that was rejected and the **measured** reason it was rejected. It is a Tier L artifact added at v0.2.0 with the promotion; `spec.md` carries the requirements those decisions produce, and `research.md` carries the raw measurements both cite.

A decision recorded without its rejected alternative is not a decision — it is a preference, and the next reader re-litigates it for free. Every section below therefore names what was *not* chosen, and why the choice was forced rather than merely preferred.

---

## §B. The store

### B.1 One origin, and the origin is the primary checkout

**Decided.** The board state lives in a single file beneath the primary checkout's `.moai/state/kanban-board/`. Every session, in whatever worktree, resolves that one path.

**Rejected — `.moai/state/kanban/`, which is v0.2.0's own choice and is already occupied.** `SPEC-KANBAN-RENAME-001` `REQ-KR-009` puts the **session record** there, resolved beneath each tree's own root — measured, `internal/factory/record.go` carries `stateDirSegments = []string{".moai", "state", "factory"}` joined beneath `projectRoot`. Per-tree is correct for a session record and wrong for a board, so the two occupants would have shared a name while differing by an invisible resolution rule, and an implementer following `REQ-KR-009`'s instruction to reuse the existing path constant would land the board in each worktree — AP-1 through the front door. The board moved rather than the record: the record's constant, its per-tree semantics, and its migration decision are all correct for what it stores, so amending the sibling would have traded a correct design for a name.

**Rejected — the column in SPEC frontmatter** (the predecessor's design), on three measured grounds:

- A worktree's `.moai/state/` is private to that worktree. `.gitignore` carries `.moai/state/` and `**/.moai/state/`, and an ignored directory is not shared between checkouts. A board each session reads from its own tree is six boards that each look coherent.
- `.moai/specs/` is tracked — **2,528** files measured. With `enforce_admins: true` on this repository, main-direct push is blocked, so six column transitions per card become six pull requests, for a value that changes several times an hour.
- A card's worktree branch is cut early and merged late, while the lead's column writes land on the primary checkout in between. Merging the card restores a stale column. This is not a locking problem: it is what putting mutable state in a versioned file does when a long-lived branch forked from it.

**The cost, accepted rather than hidden.** The board is gitignored, so there is no per-transition commit, no `git log` of column moves, and no bisectable history. The rejected alternative bought that history at six pull requests per card plus a merge-restores-stale-column defect. Where a durable movement trail is later wanted it is a separate artifact under `.moai/state/kanban-board/` — not a frontmatter field.

**Why the primary checkout is the origin, rather than a fixed absolute path or an environment variable.** A configured path is a second thing to keep in sync and a new way for two sessions to disagree; an environment variable is per-process and silently absent. The repository already knows where its own primary checkout is, and every session can ask it. The origin is derived, not declared.

### B.2 Resolving it: reuse the existing discriminant, never the bare form

**Decided.** The board root is the parent of the **absolute** common git directory, obtained by reusing the probe `internal/hook/branch_guard.go` already carries in `isPrimaryCheckout`: `git rev-parse --path-format=absolute --git-dir --git-common-dir`, falling back for git older than 2.31 to `--absolute-git-dir` plus a `--git-common-dir` result normalized against the project directory. The symbol is named; the line anchors live in `research.md`, because a normative citation of a line number in a file this SPEC does not own goes stale on that file's next edit while nothing announces it.

**Rejected — the parent of the bare `git rev-parse --git-common-dir`.** This was v0.1.0's rule, and it is wrong in exactly one place. Measured: from a worktree the bare form returns an absolute path; from the primary checkout it returns the **relative** `.git`. "The parent of `.git`" is not a path, so the rule resolves the board root to the parent of the *current directory* in the primary checkout — the one checkout the whole single-origin design points at.

The failure shape is what makes it worth a design entry rather than a typo fix: it is **asymmetric and passes the obvious test**. Every worktree exercise succeeds, and worktrees are where the feature is developed. Only a run from the primary checkout exposes it, which is why `AC-KB-002` requires that run as its positive control rather than as a nicety.

**Rejected — writing a second probe.** The discriminant exists, is in service, carries the version fallback, and has been exercised. A second probe would have to rediscover the 2.31 flag floor and the Apple Git rejection — and the naive rediscovery *is* the defect above.

### B.3 Reuse is an extraction, because the code shape does not permit anything else

**Decided.** The resolution is **extracted** into `internal/core/git` and returns a **path**; both consumers then call the extracted symbol.

**Rejected — "reuse it where it is", which is what v0.2.0 actually mandated.** Measured, the symbol is `isPrimaryCheckout(projectDir string) (bool, error)`: **unexported**, in package `hook`, returning a boolean discriminant. A board package cannot call it, and no exported path-resolving helper exists to call instead. So the v0.2.0 rule required a copy in the same breath as forbidding one — an instruction with no satisfiable reading, which a run-phase resolves by picking whichever half it notices first. This is the defect class `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018` was written for, and its disposition is taken rather than re-derived.

**Rejected — `internal/worktree` as the target,** which is the name-shaped guess. Its `doc.go` declares "working tree state guard primitives", which a git-directory resolution is not, and the sibling's own audit caught it making exactly this substitution. `internal/core/git` was chosen by reading its `doc.go` — "Git repository operations for MoAI-ADK", a read-only repository surface — and confirming it imports neither `internal/hook` nor `internal/cli`, so the extraction creates no cycle and no existing caller's contract changes.

**Why the return type changes, and why that is the point.** `isPrimaryCheckout` compares two paths for **equality**, and an equality is insensitive to an offset shared by both operands: a normalization error shifting both identically leaves every existing test green. The board consumes a *path*, where the same error is a board root one directory wrong. The existing caller's confidence therefore does not transfer, which is why `AC-KB-002` judges the resolved path against a separately recorded value instead of inheriting it.

### B.4 The board reads `status` from the card's branch

**Decided.** A card's `status` is read from the card's **branch**, by blob read and without a checkout; a card with no live worktree is read from the primary checkout. *Which* branch, where a card has several, is §B.4a — this section decides only that the branch side is read at all, and was written as though the answer were unique.

**Rejected — the primary checkout for every card**, which is what v0.2.0 left implied by saying nothing. It is wrong for the whole interval that matters. `status` transitions ride commits on the card's branch, inside a worktree that `SPEC-KANBAN-WORKTREE-001` `REQ-KW-005` keeps for the card's `run`, `review`, and `sync` sessions and `REQ-KW-007` holds until both pull requests merge. The primary-side copy therefore still reads `draft` while the card sits in `run`, `(run, draft)` is outside the compatibility table, and `REQ-KB-008` refuses to dispatch — **every card, on the normal path**. A safety mechanism firing on the case it was built to permit is worse than no mechanism, because the refusal is indistinguishable from a genuine one.

That this went unstated is not an oversight at the margin: §B.1's third rejection ground is *a card's branch is cut early and merged late, so a merge restores stale state*. The argument was in the document and was not turned on the value the board reconciles against.

**Rejected — reading inside each card's worktree.** It would see uncommitted work, and it is per-tree resolution, which §B.1 rejects for the board and would re-admit here through the read side.

**Rejected — re-deriving the branch name from the SPEC identifier.** `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` already resolves a card's branch **by observation and by the SPEC identifier the branch carries, never by prefix**, and records that the synthesized prefix diverges from the repository's convention. A second derivation here would disagree with it on exactly the cards where the convention was not followed. The contract is consumed **by name and by citation**, which is why that sibling stays in `related_specs:` and is deliberately *not* a `dependencies:` entry — the promotion was made and reversed inside v0.3.0, because it closed a cycle against the sibling's own landing dependency on this SPEC (`spec.md` §A.4a, `plan.md` AP-27). The sentence asserting the promotion survived that reversal in this file and is corrected at v0.4.0.

### B.4a Which branch, when the card has several

**Decided.** The branch is the one the card's **worktree reports**; where no worktree is live, the primary checkout. Worktree liveness is the selector, and the board searches the branch set for nothing.

§B.4 said *the card's branch*, and cards have several. Measured, 3 of the 29 SPEC identifiers appearing on branches carry two or more, and `SPEC-NAVIGATOR-SYNC-003` carries `draft`, `in-progress` and `completed` across three of them at once (`research.md` §O.2). Nothing reconciles them, because nothing in this family deletes a card's branches.

**Rejected — the most advanced stage wins.** The reading that arrives first, and it fails three separate ways. It needs the type prefix to be a stage ladder, and measurement says otherwise: only one of the three multi-branch cards carries a `plan`/`feat`/`sync` triple, the others carrying `docs`/`docs`/`feat` and `fix`/`chore`, over which no order exists. It is forbidden at the source — `REQ-KW-019` refuses to select among matching branches and names recency *and any other tiebreak* explicitly — so adopting it here would silently override a sibling's decision rather than make a local choice. And it inverts the live case: `SPEC-CODEX-PHASE2-001`'s worktree holds `feat/…-run` at `in-progress`, so a card mid-`sync` with a `sync/` branch present would be read from a tree its work is not happening in.

**Rejected — falling back when the card has no branch**, which is what §B.4 left standing. Branches are never deleted, so the condition never becomes true, and an implementation keyed on it searches the branch set for every disposed card — resolving `draft` off a retained `plan/` branch for a card sitting in `done`, which pairs outside the compatibility table and is refused. This is the v0.2.0 defect reflected: v0.2.0 read the primary checkout and broke `run`; the branch-existence key breaks `sync` and `done`. Keying on liveness is correct at both ends, because a worktree spans creation to both-pull-requests-merged and the measurement shows the primary-side value stale for exactly the one card that has one.

**The accepted consequence.** Two residuals are refused rather than guessed (`REQ-KB-024`): a worktree reporting no branch — a detached `HEAD`, present in this repository today — and a `REQ-KW-019` refusal reaching the board. Both render the card in its recorded column with `status` **unresolved**, not dispatchable, candidates surfaced. The alternative is a substituted enum member, and the one an implementation acquires for free is the zero value, which reports `draft` and dispatches.

**The accepted consequence.** A branch read observes **committed** state, so an uncommitted transition is not yet a transition. Stated rather than left implicit, because the alternative reading would require the per-tree read just rejected.

---

## §C. Ownership: who writes, how, and under what exclusion

Four decisions, deliberately separate — the writer, the write, the exclusion, and the exclusion's own exit. Collapsing any of them is the failure mode; each fails on its own.

### C.1 One writer: the lead

**Decided.** Exactly one role — the session occupying the `lead` role — writes the board state file. Every other session reads it. The implementation enforces this; it does not merely say it.

**Restored, not invented.** The predecessor's `REQ-KM-044` bundled the `column:` storage mechanism with the ownership rule. §B.1 rejects the mechanism. Nothing in that rejection touches ownership — relocating a store says nothing about who may write it — yet the split deleted both. Measured at v0.1.0: `grep -rc 'sole writer\|single writer'` reported **0** in every file of all three sibling SPECs.

**Rejected — leaving it to the bootstrap sibling.** That is what happened, and it produced a hole with a boundary drawn around each side: this SPEC said it "names no actor", the sibling says it "decides who is told about a card and by what message". Being told about a card is not writing it. Write authority belongs with the store's owner, so it is stated here; election belongs with topology, so it stays in `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004`.

**Rejected — documenting the rule instead of enforcing it.** A rule that lives only in prose is the rule that was already lost once. `AC-KB-017` demands both an enumeration of write call sites and a runtime refusal, because an enumeration establishes what the code contains, not what it refuses.

**The runtime half needs a readable role, and that contract is now supplied.** Refusing a write from a non-`lead` session presupposes the board can ask what role the caller occupies — a runtime value, which `REQ-KB-004`'s *session identifier* is not. At v0.2.0 no SPEC in the family owned it; two auditors reached the gap from opposite ends. `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` now carries it, widened at that SPEC's v0.3.0 into a **role declaration** distinct from the launch label and resolvable from a session that is not the `lead`. This design **consumes** it by name and decides nothing about it — not its carrier, not its derivation, not its lifetime. **Rejected — deriving the role from the session identifier or the launch label**: the sibling's own text rules the label out (one role maps to two-or-more possible labels, chosen by the operator at launch), and a derivation here would silently re-open the question that sibling just closed.

### C.2 Atomic writes: same-directory temp, then rename

**Decided.** Every board write goes to a temporary file created **in the target's own directory**, then renames over the target, reusing `internal/atomicfile`.

**Rejected — writing in place.** A reader can then observe a prefix, which is precisely the partially-written file the unknown state of §D exists to detect. Preventing the condition is cheaper than detecting it, and both are kept: prevention here, detection there.

**Rejected — a temp file in the system temp directory.** This is the tempting version, because `os.CreateTemp("")` is the shorter call and the rename still happens. `rename(2)` is atomic only **within one filesystem**; across a device boundary the operation degrades into a copy and the torn write returns. The code looks atomic, and a test asserting "a rename occurred" passes. `AC-KB-018` therefore checks directory equality **statically** as a separate half.

**Rejected — a second atomic-write primitive.** `internal/atomicfile` exists, carries the Windows handling (where a rename fails while a handle is open on the destination), and `internal/verify/store.go` `Save` is a working same-directory caller to copy. A new one would re-derive the same platform edge cases.

### C.3 Board-wide exclusion, and why the card lock cannot do it

**Decided.** Every board mutation is serialized beneath an advisory lock scoped to the **whole board**, held across the entire read-modify-write — read, decide, write — obtained from `internal/spec/lock.go` and its platform counterparts.

**Rejected — the card-scoped lock.** This is the most likely wrong turn, because the card lock is already there and already correct for its own purpose. It cannot make a board-wide invariant hold. With WIP 2 and two concurrent assignments of two **different** cards: each takes its own card's lock, each reads "1 card in `run`", each finds the bound satisfied, each writes. The board ends at WIP 3, and neither writer misbehaved under its own lock.

The consequence is stated plainly because it is what makes this decision load-bearing rather than defensive: **`REQ-KB-009`'s WIP-2 bound is only sound beneath board-wide exclusion.** At v0.1.0 it was a claim the design could not deliver.

The sibling's `REQ-KW-013` card lock is untouched and stays correct for holder assignment, where card scope is the right granularity. This is an addition. What is asserted is narrow: a card lock is *insufficient for a board mutation*.

**Rejected — `internal/lockfile`.** It is a `map[string]*sync.Mutex` on Windows, and its own package comment records that concurrent writes across **different** OS processes are not protected and forbids upgrading it to `LockFileEx`. Sessions are distinct OS processes, so it would hold nothing in production while passing every same-process test. `SPEC-KANBAN-WORKTREE-001` selected the `internal/spec/lock.go` family for the same measured reason; this reuses that selection rather than re-deciding it.

**Rejected — a third locking mechanism.** Two exist; one is right. Adding a third multiplies the platform matrix for no capability.

### C.3a The lock can outlive its holder, and v0.2.0 did not notice

**Decided.** The board-wide lock artifact records the identity of the process that created it, and a **bounded clearing operation** removes it only when that process is positively observed absent — an explicit operator-visible act, never a step the acquire path takes, and never conditioned on age.

**Rejected — reusing the lock family unmodified**, which is what `REQ-KB-019` mandated on its own. That family's Windows substrate states in its own header that stale-lock detection is a post-MVP enhancement and that cleanup is manual. A `lead` killed mid-mutation therefore leaves an artifact blocking **every** future board mutation, with no operation defined that could change the answer — §D's brick, reintroduced by the revision that removed it. `plan.md` AP-22 was written against the corrupt-file path and never turned on the lock, which is the same blind spot as §B.4's, in a different section.

**The asymmetry is recorded rather than smoothed, because it decides who finds this.** Measured: the Unix substrate holds `flock(2)` on an open descriptor, released by the kernel on process exit, so a killed holder leaves an inert file — the repository's own `.moai/state/` carries fourteen such orphaned zero-length `spec-close-*.lock` artifacts, none of which blocks anything. The Windows substrate is atomic-create-exclusive with no release, so there the artifact **is** the lock. A defect that never reproduces on the developer's machine and always reproduces on a user's is not a smaller defect; it is a later one.

**Rejected — clearing on age.** A long mutation is indistinguishable from a dead one by elapsed time alone, and the threshold that makes the clear safe makes it useless.

**Rejected — the naive check-then-unlink**, which is the hazard the sibling's audit found in its own version of this operation. Between reading the recorded identity and removing the artifact, the lock may be legitimately released by its owner and **re-acquired by a live process**; the clearer then unlinks a valid lock and two writers enter the critical section — a repair that causes the exact concurrency the lock prevents, with every step correct in isolation. The removal is therefore conditioned on a re-read of the recorded identity immediately before the unlink, so a recreate observed at that point aborts the clear.

**Rejected — asserting that the removal happens only if the artifact is unchanged.** This is what v0.3.0's requirement text claimed, and it is not available. `unlink(2)` — and the `os.Remove` wrapping it — takes a **path**, resolves the name at call time, and removes whatever the name then denotes; the descriptor the identity was read through has no bearing on which file is unlinked. There is no portable handle-based form: no POSIX call takes one, `funlinkat(2)` is FreeBSD-only and absent on Linux and Darwin, and the Windows delete-on-close disposition is set by the *holder* at open time — which is the process this operation exists to clean up after and which is by hypothesis gone. An implementer reading a determinism claim writes the two-step code regardless, believes it atomic, and stops looking, so the claim is worse than the gap it papers over.

**Rejected — acquiring the lock exclusively and unlinking beneath that exclusion.** The mechanism that comes closest, and it does not reach either. On the Unix substrate `flock(2)` coordinates only among processes that also take the lock; it does not prevent the path being rebound, so the unlink is unprotected in exactly the way that matters. On Windows the substrate is atomic-create-exclusive, so acquiring the artifact to clear it is indistinguishable from acquiring it to hold it, and a clearer that succeeds has become a second writer. The portability cost is also asymmetric in the wrong direction — the platform where the mechanism is least expressible is the only platform where the defect occurs at all (§C.3a).

**Accepted instead — a time-of-check-to-time-of-use mitigation, named as one.** The re-read narrows the interval from the span of an operator's inspection to two adjacent statements; it does not empty it. Two things make the residual acceptable where the original window was not. It is entered only by a clear already running against an artifact whose recorded owner was observed dead, so the re-acquisition it can race must have begun after that observation. And the operation is explicit, operator-invoked, and reports what it removed, so a rare bad outcome is attributable rather than silent.

### C.4 Why C.1 and C.3 are not the same decision

A reader may reasonably ask why a sole writer needs a lock at all. Three answers, each a live failure if the two are collapsed:

- Sole ownership removes cross-**role** contention, not contention within the writer. A lead performing two mutations concurrently is unserialized without the lock.
- A **reader** observing an in-flight read-modify-write has no guarantee of a coherent board from ownership alone. Atomicity (C.2) covers the file; the lock covers the multi-step decision.
- A second lead process during a handoff, or a stale one, is exactly the case ownership assumes away and the lock does not.

Sole ownership keeps the set of writers **knowable**; the lock makes the read-modify-write **indivisible**. Different properties.

---

## §D. Failure and recovery

**Decided.** A partially-written or unreadable board state file is reported as **unknown**, every dispatch is refused, and no empty board is presented. From that state, a bounded recovery — the sole writer reconstructing or replacing the file under the board-wide lock — returns the board to a readable state, invoked as an explicit operator-visible act.

**Rejected — treating an unreadable file as an empty board.** The empty-board path already exists and an empty board is a valid board, which is what makes this tempting. It is the wrong direction: an empty board presented as accurate reports zero cards in `run`, admitting new cards past a WIP limit whose contents are unknown, onto work that may already be in flight. A refusal is loud and costs a re-read; a false empty board is silent and costs two sessions on one card.

**Rejected — the refusal with no exit** (v0.1.0's actual shipped shape). "Fail safe" and "fail permanently" are different claims, and only the first was argued. With no atomicity requirement and no recovery operation, one session killed mid-write left the board refusing every dispatch forever, with no operation defined that could ever change the answer.

**Rejected — silent auto-repair on read.** The obvious exit, and worse than the brick. It discards the evidence that something killed a writer, and it does so on a file whose contents are *by definition* unknown. It is the same objection `spec.md` §A.4 raises against silently reconciling a column/status disagreement, sharpened: there the board would overwrite a value it could read, here one it could not.

This is why `AC-KB-020` asserts the **repeated read** as its load-bearing half. A criterion checking only that recovery works passes a silent auto-repair perfectly; a criterion checking only that reads stay unknown passes an implementation with no recovery at all. Only the pair distinguishes a bounded exit from both.

**Decided — absent is a third state, and it is an empty board.** A state file that does not exist is neither partially written nor unreadable, and v0.2.0 assigned it to nobody. **Rejected — folding absent into unknown**: the board would brick on first use, before any card exists, making the recovery operation the only way to start a board. **Rejected — folding unreadable into absent**, or reaching the empty-board result through the absent path: that is the door this section closes, re-opened one substitution away, and from the outside the two are indistinguishable. What makes the split safe is that it rests on the operating system's own answer to whether a file exists rather than on a judgment about content. The **sole writer** creates the state directory on the absent path, since the directory is gitignored and so cannot arrive with the repository, and the reused `internal/spec/lock.go` already demonstrates the shape by creating `.moai/state/` before taking its lock.

**Decided — what "bounded" means, since v0.2.0 asserted it without defining it.** Bounded **in extent**: the recovery touches the board state file alone — no frontmatter, no worktree, no card it can still read. Bounded **in effect**: one invocation, one definite verdict, no re-entry and no retry.

**Rejected — an unqualified "replacement of the state file".** Read literally, that is permission to write an empty board — after which reads succeed, `run` reports zero cards, and new cards are admitted onto work that may be in flight. It is this section's own rejected alternative, reached through a door labelled *explicit*, and `AC-KB-020` as written asserted only that the board leaves the unknown state, which such an implementation satisfies perfectly. So a recovery that cannot reconstruct a card **records what it could not recover, durably, and surfaces it**. An operator who invoked a recovery is entitled to know whether they got their board back or a new one.

---

## §E. The column

### E.1 Explicit, never derived

**Decided.** The board reads a recorded column. It computes nothing.

**Rejected — deriving the column from `status` plus the `progress.md` §E markers** that `internal/spec/era.go` already parses. It is blind exactly where it is needed: `§E.3` is written at the **end** of run-phase, and the `review` column is the interval *before* it is written. The derivation cannot separate `run` from `review` — the collision it was brought in to solve. Rejected in the predecessor, rejected again here, and `plan.md` AP-3 names the re-entry route (a helper that "recovers" a missing column by inference).

### E.2 Six columns, fixed

**Decided.** `backlog` → `plan` → `run` → `review` → `sync` → `done`, a closed enumeration.

**Rejected — a `test` column.** Unit tests belong to the TDD/DDD cycle inside `run`; a separate column encodes test-after.

**Rejected — a held or blocked column.** The unheld state already exists and serves both of its causes — a card waiting for a free session, and a card whose holder was released after its session died. One field, two causes; a seventh column would be a new column for a state the board already has.

**Rejected — an operator-extensible column set.** Every column carries a legality rule and a status-compatibility row; an extensible set makes both unspecifiable.

### E.3 The status vocabulary does not reach, and the board does not extend it

**Decided.** Where a column has no `status` counterpart — `review` most sharply — the gap is resolved at the board layer. The canonical 8-value enum is not extended.

**Rejected — adding a `reviewed` status.** The board does not own that enum, and the absence is a fact about the vocabulary rather than a defect in it. Extending an enum owned elsewhere to fit a local model is how two sources of truth begin.

**Decided — the four values the six-row table does not carry, and they are not one case.** `planned` is a lifecycle value (17 files measured) that collides where `draft` already collides, so it is admitted in `backlog` and `plan` and nowhere else; admitting it to `run` would assert work started under a status meaning it had not. `archived`, `superseded`, and `rejected` (42 files) are out-of-lifecycle terminals for which **no card is created at all** — a different statement from "illegal in every column", and stated as a property so the absence is not read as the `planned` defect a second time. `done` is the wrong home for them: `done` means worked and finished, and a rejected SPEC was never worked.

---

## §F. WIP and the session count

**Decided.** The `run` column admits at most 2 cards. The deployed `run`-session count defaults to 1 and is raisable to 2 by configuration. Neither is derived from the other.

**Rejected — gating admission on a free session.** It reads as prudence, and it is the conflation arriving by the back door: the effective WIP becomes the session count, and the board appears to honor a limit it has silently stopped enforcing. With WIP 2 and one session, the second card enters `run` and waits there **unheld** — a legal steady state, dispatched the moment a session frees.

**Rejected — one knob.** They answer different questions: how much work may be in flight, and how much capacity is deployed. Collapsing them makes the first unspecifiable.

This is the decision `C.3` makes deliverable. Without board-wide exclusion the bound is not enforceable at all, which would have made the two knobs one knob by accident rather than by design.

---

## §G. Out of Scope

### Out of Scope — decisions this file does not make

- How the `lead` role is elected, assigned, or handed over. §C.1 names the owner of the write; `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004` elects it.
- The transport by which a card's move is communicated, and the dispatch protocol. Both are the bootstrap sibling's.
- Worktree lifecycle, stall detection, holder release, and orphan classification (`SPEC-KANBAN-WORKTREE-001`). §E.2 defines the unheld state one of them resolves to; it defines no detector.
- The internal design of `internal/atomicfile` and `internal/spec/lock.go`. §C.2 and §C.3 decide to reuse them, not how they work.

### Out of Scope — rendering

- Any web view, TUI, or dashboard. The read-only web board is a separate line of work, and no decision here constrains it beyond the store's location.

---

## §H. Cross-references

- `spec.md` §A.3, §A.4a, §A.6, §A.7 — the requirements these decisions produce.
- `research.md` — the commands and observed outputs every measurement above cites.
- `plan.md` §E (D1 … D5) and §G (AP-16 … AP-26) — the same decisions as an execution order, and the routes back into each rejected alternative.
- `acceptance.md` AC-KB-002, AC-KB-012, AC-KB-017 … AC-KB-024 — the criteria that decide them.
- `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` (branch resolution, consumed by §B.4), `REQ-KW-005` / `REQ-KW-007` (the worktree lifetime that makes the primary-side `status` stale), `REQ-KW-014` (the stale-lock shape §C.3a ports), `REQ-KW-018` (the extraction disposition §B.3 takes).
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the role declaration §C.1's runtime half consumes.
- `SPEC-KANBAN-RENAME-001` `REQ-KR-009` — the session-record path §B.1 separates from, and does not amend.
