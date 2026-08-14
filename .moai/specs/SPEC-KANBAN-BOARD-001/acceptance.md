---
id: SPEC-KANBAN-BOARD-001
title: "Acceptance criteria — six-column kanban board model with a single-origin board state store"
version: "0.6.2"
status: draft
created: 2026-08-10
updated: 2026-08-14
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, acceptance, given-when-then, verification, sole-writer, atomicity, role-carrier"
tier: L
---

## §A. How these criteria are judged

Every criterion below is `Given … When … Then …`, binary, and names the command or the mechanism that decides it. A criterion whose command cannot fail is not a criterion.

### A.1 Command hygiene (binding on every criterion here)

Five rules, each paid for by a prior failure in this repository.

1. **Never read `$?` after a pipe.** `cmd | tail` makes `$?` belong to `tail`. Redirect to a log, read `rc` from the command, then count `^FAIL` across the **whole** log — not a tail, which truncates exactly the region failures live in.
2. **Never iterate an undefined shell array.** Write the literal `for` list.
3. **`spec lint` is invoked per file.** A `*.md` glob is unsatisfiable for a multi-artifact SPEC here: `DuplicateSPECIDRule.CheckAll` treats each path as a separate SPEC, so siblings fail with `ParseFailure` or `DuplicateSPECID`.
4. **A table cell `\|` is a literal pipe, not an alternation.** Any criterion whose pattern is transcribed out of a markdown table must be re-authored before it is run, or it passes vacuously.
5. **Do not reimplement an existing guard's regex.** A guard reimplemented without its exemption list is a false-failure machine. Where a mechanical guard exists (neutrality, catalog parity), run **that guard** and treat its exit code as the verdict; a supplementary grep is an early-warning aid, never the authority.

6. **A working-tree check does not observe a branch.** `test -d`, `test -f`, and every other filesystem predicate read the *local working tree*, so they are satisfied by an uncommitted local edit. Where a criterion claims something has **landed on the base branch**, it must read the branch — `git cat-file`, `git ls-tree` against `origin/main` — not the tree the implementer is standing in. AC-KB-001 was rewritten at v0.2.0 for exactly this reason.

### A.2 Positive controls

Where a criterion asserts an absence, or asserts that a refusal fires, a **positive control** is required: a construction under which the same command reports the opposite, run once and recorded. An absence-check with no demonstrated ability to fire is indistinguishable from a broken command — and so is a refusal that fires unconditionally.

### A.3 Independence claims need a varied knob

A criterion asserting that two values are **independent** is unfalsifiable if it only reads them once in one configuration: an implementation deriving one from the other satisfies it whenever the two happen to agree. Such a criterion must **vary one knob while holding the other fixed** and observe that the dependent value does not move — in both directions, since a one-directional check passes an implementation that derives the other way. AC-KB-012 was rewritten on this rule at v0.2.0 and rewritten again at v0.3.0, when the knob one of its directions varied turned out to belong to an **excluded** sibling and so might not exist to vary at this SPEC's run-phase. Where a direction's knob is outside this SPEC's scope, the falsifiable form of the independence claim is an **absence** over this SPEC's own inputs — the board reads no such value — which is observable at a boundary this SPEC owns.

### A.4 Concurrency criteria run in separate processes

Where a criterion judges an exclusion between sessions, it must use **separate OS processes**. Sessions are distinct processes, and an in-process mutex satisfies a goroutine-based test perfectly while protecting nothing in production — `internal/lockfile/lockfile_windows.go` is the repository's own worked example, its package comment stating that cross-process writes are unprotected and forbidding an upgrade. A goroutine test therefore measures the test harness, not the requirement.

---

## §B. Preconditions

**AC-KB-001** (REQ-KB-001, REQ-KB-002) — *Given* the run-phase tree, *when* the **base branch** is queried for the renamed package — `git ls-tree -d --name-only origin/main internal/kanban` reporting the path, and `git ls-tree -d --name-only origin/main internal/factory` reporting nothing — *then* both hold; and *when* the added lines of this SPEC's diff are scanned for the token `factory` (case-insensitive, excluding paths under `.moai/specs/`), *then* zero matches are reported. **Positive control**: run against a revision predating the rename, the same two queries report the opposite — recorded once.

REQ-KB-002 requires the prerequisite to have **landed on the base branch**, and the v0.1.0 form of this criterion checked `test -d internal/kanban`, which does not observe a branch at all. A directory predicate reads the local working tree, so an uncommitted local rename — or a rename present only in the implementer's worktree, which is precisely the state measured at plan time — satisfies it while the prerequisite has landed nowhere. The gate then passes on the one tree it was written to fail on, and every later identifier is authored against a prerequisite that may never merge. `git ls-tree` against `origin/main` cannot be satisfied by uncommitted local state (§A.1 rule 6).

---

## §C. Single-origin board state

**AC-KB-002** (REQ-KB-005) — *Given* a session running inside a worktree, *when* the board root is resolved, *then* it equals the primary checkout's path recorded by `plan.md` §C commands 2 and 2b — the parent of the **absolute** common git directory — and the board file resolves beneath that root's `.moai/state/kanban-board/` through **this package's own** path-segment constant, distinct from the session record's per-tree `stateDirSegments` — which after `SPEC-KANBAN-RENAME-001` landed is `internal/kanban/record.go`'s `{".moai", "state", "kanban"}`, re-measured at v0.5.0 and **not** the pre-rename `internal/factory/record.go` value earlier revisions cited. Decided by comparing the resolved value against the recorded one, and by the property scan below.

**The scan is a property scan, not a name scan, because the rename made a name scan undecidable.** Through v0.5.0 this criterion was decided by "scanning the board package for a reference to the session-record constant, which must yield none". That was executable when the constant lived in `internal/factory`. It is not now: this SPEC's `module:` is `internal/kanban`, and `grep -c 'stateDirSegments' internal/kanban/record.go` reports **3** — the board package *is* the package the session-record constant is defined in, so a name scan over it yields matches by construction and can never reach zero. A criterion whose command cannot report the passing outcome is not a criterion (§A.1), and this one had silently become one. v0.5.0 recorded the same-package fact and propagated it to `plan.md` AP-24 and the M1 persistence bullet, but not to the criterion whose decision procedure it broke — cross-file drift of the shape v0.4.0 repaired, running the other way.

The observable that survives the collocation is **use**, not mention: *no board-state path is constructed from the session record's `stateDirSegments`.* The scan therefore enumerates every construction of a board-state path and requires each to be built from this SPEC's own path-segment constant, with the session-record constant reachable from none of them; equivalently, the name scan may be retained if it is scoped to the board's own files **excluding** `record.go`, which is the session record's file and not the board's. Either form is acceptable and the run-phase records which it used. **Positive control**: a board-state path deliberately constructed from `stateDirSegments` is reported by whichever form was chosen, run once and recorded — without it the scan's zero is again indistinguishable from a scan that cannot fire, which is the state this repair exists to leave.

Placing the board in a distinct sub-package would also restore decidability and is not required here; `REQ-KB-005` binds the constant and the resolution, not the package layout, and a criterion that mandated a layout would be deciding a run-phase question this SPEC has no measurement to decide.

**Positive control, and it is the load-bearing half here**: the same resolution run **from the primary checkout itself** yields that same path. Running it only from a worktree would pass an implementation built on the bare `--git-common-dir`, which returns the relative `.git` in the primary checkout and resolves the board root to the parent of the *current directory* — measured, `plan.md` §C command 2. The resolution must be a function of the repository, not of the caller's location, and only the primary-checkout run demonstrates that.

**Second half — the fallback branch must be entered, and it is not entered by running the criterion.** The primary probe succeeds on every git at or above 2.31; measured, this repository runs 2.50.1, so both runs above take the `--path-format=absolute` path and the older-git fallback executes **zero times**. An unexercised branch in a resolver is not a cosmetic gap here: its output is a *path*, and a wrong path is a wrong board root, which is the failure the whole single-origin design exists to prevent. The criterion therefore runs a **third** resolution with the primary probe forced to fail, and requires the same board root.

It is forced through the existing mechanism, not around it: `internal/hook/branch_guard.go` carries a package-level `execCommand` indirection whose own comment records that this exists so an older-git host rejecting `--path-format=absolute` can be simulated, and the test file's comment records that **direct invocation of the fallback is INSUFFICIENT — a vacuous pass that bypasses the dispatcher**. Calling the fallback directly therefore does not satisfy this half.

The reason the borrowed code carries no evidence about this is worth stating, because it is not that the borrowed normalization is visibly wrong. `isPrimaryCheckout` compares two paths for **equality**, and equality is insensitive to an offset shared by both operands — a normalization error shifting both identically leaves every existing test green. The board takes the **parent** of one of them as a path, where the same error is silently fatal. The existing caller's green suite is therefore not evidence for this consumer's use, which is why the resolution is judged against a recorded value rather than trusted.

**AC-KB-003** (REQ-KB-005, absence half) — *Given* the implementation, *when* it is scanned for a board-state path constructed from the working directory, from the session's own tree root, or from any other tree-relative anchor, *then* none is found. **Positive control**: a deliberately introduced tree-relative board-state read is reported by the same scan, run once and recorded, so the absence check is demonstrably able to fire. A scan reporting zero matches without this control does not satisfy this criterion.

**AC-KB-004** (REQ-KB-006) — *Given* a card whose recorded column is `review` while its `progress.md` lacks the `§E.3` marker, *when* the board reads the card, *then* the column is `review`, taken from the record and not computed; and *given* the same card with its recorded column mutated to `run` and nothing else changed, *then* the board's answer changes accordingly. The second half is the load-bearing one: a derived implementation could not change its answer from that mutation alone.

**AC-KB-005** (REQ-KB-007) — *Given* the board package, *when* it is scanned for a write to any SPEC frontmatter key, *then* none is found — no `column` key is introduced anywhere, and no `status` value is written; and *when* it is scanned for status string literals, *then* every one is drawn from the canonical 8-value enum read from `spec-frontmatter-schema.md` rather than from a copy in this SPEC. **Positive control**: a deliberately introduced frontmatter write is reported by the same scan.

**AC-KB-006** (REQ-KB-013, detection half) — *Given* a truncated or unparseable board state file, *when* the board is read, *then* the read reports the board as unknown and every card is reported not dispatchable; an empty board is **not** returned. **Positive control**: a well-formed file of the same size reads successfully, so the failure path is conditional on the content rather than on the size.

**AC-KB-020** (REQ-KB-013, REQ-KB-022) — *Given* a board in the unknown state, *when* the read is repeated any number of times, *then* every read still reports unknown and still refuses every dispatch — the read path performs no repair; and *when* the recovery operation is invoked explicitly by the sole writer of REQ-KB-017 while holding the board-wide lock of REQ-KB-019, *then* the board leaves the unknown state and subsequent reads succeed; and *given* a recovery that could not reconstruct one or more cards, *when* it returns, *then* a durable record naming what could not be recovered exists and is surfaced to the operator, and the recovery is observed to have modified the board state file alone, to have returned a single definite verdict, and not to have re-entered or retried.

The repeated-read half is the load-bearing one for the *auto-repair* failure and is the reason this is a separate criterion rather than a clause on AC-KB-006. An implementation that silently self-repairs on the second read satisfies "the board leaves the unknown state" perfectly while violating the requirement — it discards the evidence of whatever killed the writer, on a file whose contents are by definition unknown (`spec.md` §D.8). Conversely, an implementation with no recovery at all satisfies the repeated-read half; only the pair distinguishes a bounded exit from both a brick and a silent auto-repair.

The third clause closes a different hole, and it is the one v0.2.0 left open. `REQ-KB-013` permits the recovery to be "a reconstruction **or replacement** of the state file", and a replacement of an unreadable file *is* an empty board — after which reads succeed, `run` reports zero cards, and new cards are admitted onto work that may be in flight. That is the harm §A.6 argues against, reached through the door labelled *explicit*, and the v0.2.0 form of this criterion asserted only that the board leaves the unknown state and that subsequent reads succeed — which such an implementation satisfies perfectly. **Positive control**: a recovery that discards an unreadable board without recording anything is reported by this criterion; run once and recorded.

---

**AC-KB-024** (REQ-KB-021) — *Given* a board root whose state file **does not exist**, *when* the board is read, *then* it reports a legitimately empty board, dispatch is permitted, no recovery is required, and after the read the state directory exists; and *given* a state file that exists and is unreadable, *when* the board is read, *then* it reports unknown and refuses every dispatch. Decided by a single table-driven test asserting **both** rows, so that the two are observed to differ.

The pairing is the criterion. The defect is a conflation, and a conflation is invisible when either row is asserted alone: a test covering only the absent row passes an implementation that also reports empty for an unreadable file (§A.6's rejected direction), and a test covering only the unreadable row passes an implementation that bricks on first use, before any card exists, making recovery the only way to start a board. Both readings were available under v0.2.0, which bound only "partially written or unreadable" and named the absent case nowhere in the family. **Positive control**: an implementation returning the same verdict for both inputs is reported by this criterion, whichever verdict it returns.

## §D. The six columns and the card

**AC-KB-007** (REQ-KB-003) — *Given* the board package, *when* the column enumeration is exercised over every declared value, *then* exactly six values exist in the order `backlog`, `plan`, `run`, `review`, `sync`, `done`, and no constructor accepts a value outside the set. Decided by `go test ./internal/kanban/ -run Column`.

**AC-KB-008** (REQ-KB-004) — *Given* a card, *when* it is persisted and re-read, *then* the SPEC identifier, the column, the holding session identifier, and the last-transition instant all round-trip unchanged; and *given* a card held by no session, *then* it round-trips with an empty holder rather than a synthesized one.

---

## §E. Column-to-status consistency

**AC-KB-009** (REQ-KB-008) — *Given* the `(column, status)` compatibility table of `spec.md` §A.4 **as revised at v0.2.0**, *when* a table-driven test exercises **every** pairing of the six columns against every enum value — the legal rows and the illegal ones alike — *then* each legal pairing is accepted, and each illegal pairing marks the card inconsistent, reports it not dispatchable, surfaces both values, and leaves the recorded column and the SPEC's `status` byte-unchanged on disk. A test asserting only the legal rows fails this criterion, because it never demonstrates that an illegal pairing is refused; an implementation accepting every pairing would pass it.

Four pairings are called out because the v0.1.0 table decided each of them wrongly or not at all, and a test transcribed from that table would re-encode the error: `(sync, completed)` is **legal** — the `in-progress → implemented → completed` sync commit makes `completed` the value usually observable during `sync`, so refusing it would block essentially every card one column short of `done`; `(backlog, planned)` and `(plan, planned)` are **legal**; and `(run, planned)` is **illegal**, since `planned` asserts work has not started.

**AC-KB-022** (REQ-KB-020) — Four observations over **one** fixture, whose branch set is the measured shape of `SPEC-NAVIGATOR-SYNC-003` (`spec.md` §A.4b): three branches matching the card by `REQ-KW-003`'s exact-token rule, carrying `status: draft`, `status: in-progress` and `status: completed` respectively, with the primary checkout carrying `status: completed`.

*Given* that fixture with a **live worktree reporting the `in-progress` branch** and the card in the `run` column, *when* the board reconciles the card, *then* it reads exactly `in-progress` — not `completed` from the more advanced branch, not `draft` from the less advanced one, and not `completed` from the primary checkout — and reports the card **consistent** and dispatchable. *Given* the identical branch set with **no live worktree** and the card in `done`, *then* the board reads `completed` from the primary checkout and none of the three branches supplies the value, so the card is consistent rather than pairing as `(done, draft)`. *Given* a card with no branch at all — a `backlog` or `plan` card — *then* its `status` is read from the primary checkout and judged by the same table. *Given* a transition written into a worktree's working tree but not committed, *then* the board does not observe it.

The first two observations run against the **same** branch set and differ only in worktree liveness, which is what makes them a pair rather than two tests: an implementation selecting the source on branch *existence* passes neither, and one selecting by stage passes the second and fails the first. A fixture configuring a single branch decides none of this — a search over one candidate and an observation of one worktree return the same answer — which is why the multiplicity is in the fixture rather than in the prose (`spec.md` §D.11).

**Traversal refusal conjunct (added in place at run-phase review, operator decision #1 — the amendment-in-place pattern, not a new criterion).** The SPEC identifier is interpolated into a worktree path, a `git show` ref, and the primary checkout's `spec.md` path before any of the four observations can run, so the read/persistence surface itself is part of this criterion's shape: *given* a traversal-shaped specID — containing `..`, a path separator, or an absolute path — *when* it is passed to the board's read or persistence entry point, *then* it is REFUSED before reaching any path or ref join, and no out-of-tree read occurs; the refusal routes through the repository's shared sanitizer (`internal/cli/specid.ValidateSpecID`), the same guard the CLI applies at its SPEC-ID boundaries, and the canonical-id positive control holds (the refusal is conditional on shape, never unconditional). A traversal value is not persistable through the write surface either.

**AC-KB-025** (REQ-KB-024) — *Given* a card whose worktree exists and reports **no branch** — a detached `HEAD`, the shape the repository currently holds at `.claude/worktrees/rc-build` — *when* the board reconciles the card, *then* the card's `status` reads **unresolved**, the card still occupies the column the board recorded for it, it is reported not dispatchable, every candidate the resolution found is surfaced, and no member of the canonical 8-value enum is reported in its place. *Given* the same card, *then* it is **not** reported inconsistent under REQ-KB-008, and the test asserts that distinction by observing which of the two outcomes was reported rather than by observing that dispatch was refused — both outcomes refuse dispatch, so a criterion resting on the refusal alone cannot tell them apart. *And given* the reconciliation of that card, *when* it returns, *then* the board state file and the SPEC's `spec.md` are both byte-unchanged on disk.

The observation that carries this criterion is the **absence of a substituted value**. An implementation that cannot resolve a source and returns the zero value of its status type reports `draft`, which pairs legally with `backlog` and `plan` and would be dispatched — the silent-default failure, indistinguishable from a correct read at every surface except this one. Asserting `unresolved` positively is what excludes it; asserting only that dispatch was refused does not, since the illegal-pairing path refuses too.

**The divergent pair is the whole criterion, and it must be constructed rather than found.** The two trees agree for most of a card's life, and while they agree an implementation reading the primary checkout is indistinguishable from one reading the branch — so a criterion built on an agreeing card passes both and measures nothing. The interval where they differ is the interval the defect lived in: `status` transitions ride commits on the card's branch, and `SPEC-KANBAN-WORKTREE-001` `REQ-KW-005` and `REQ-KW-007` keep that worktree alive until both pull requests have merged, so the primary-side copy reads `draft` for the entire time the card sits in `run`. Under v0.2.0 that pairs as `(run, draft)` — outside the compatibility table — and `REQ-KB-008` marks it inconsistent and refuses to dispatch it. Every card. On the normal path.

**Positive control**: the same card evaluated against a primary-checkout read is reported inconsistent by the table, and this criterion reports the failure — which is what demonstrates that the branch-side read is doing the work rather than an incidentally-agreeing pair. The no-branch half needs its own row and is not implied by the first: an implementation with no fallback fails on the first unplanned card and passes every branch-side assertion. Decided without a checkout: a blob read against the card's branch, whose ref is shared across every checkout of the repository.

Additionally, *given* this SPEC's frontmatter and `SPEC-KANBAN-WORKTREE-001`'s, *when* each `dependencies:` line is read, *then* they do **not** both name each other: this SPEC's does not name that sibling, and that sibling's does name this SPEC. The pair is the observation — the first half alone would also be satisfied by a SPEC that had simply forgotten the relationship, and the second half is what establishes that the omission is a **refused cycle** rather than an oversight, the consumption of `REQ-KW-003` being a contract dependency discharged in `REQ-KB-020`'s own text (`spec.md` §A.4a). This is the mirror of the sibling's `AC-KW-001`, whose corresponding observation **failed** at its v0.3.0 authoring time because this SPEC then declared the reverse edge; that failure is resolved by this criterion holding.

**AC-KB-010** (REQ-KB-008, collision half) — *Given* a SPEC whose frontmatter `status` is `draft`, *when* the board reconciles it, *then* the reconciliation does not by itself decide between the `backlog` and `plan` columns; and *given* `status: in-progress`, *then* it does not by itself decide between `run` and `review`. Decided by a table-driven test asserting each collision is reported as ambiguous rather than resolved.

**AC-KB-021** (REQ-KB-008, status-coverage half) — *Given* the canonical 8-value status enum read from `spec-frontmatter-schema.md`, *when* every value is checked against the compatibility table, *then* `planned` is admitted in `backlog` and `plan` and rejected everywhere else; and *given* a SPEC whose `status` is `archived`, `superseded`, or `rejected`, *when* the board builds its card set, *then* no card is created for it in any column — which is distinct from creating a card and reporting it inconsistent, and the test asserts the distinction by observing card **absence** rather than an inconsistency report.

The scale is why this is a criterion rather than a note. Measured at authoring time, `grep -rlE '^status: planned\s*$' .moai/specs/` reports **17** files and `grep -rlE '^status: (archived|superseded|rejected)\s*$' .moai/specs/` reports **42**. Under the v0.1.0 table none of those 59 SPECs was legal in any column, so every one of them would have been marked inconsistent and made permanently undispatchable by REQ-KB-008 — a table omission presenting as a board-wide failure. **Positive control**: re-running this criterion against the v0.1.0 table reports the failure for `planned`, so the criterion is demonstrably able to fire.

---

## §F. WIP, admission, and the unheld state

**AC-KB-011** (REQ-KB-009) — *Given* a board with two cards in `run`, *when* a third transition into `run` is attempted, *then* it is refused with a named error and the board is byte-unchanged. **Positive control**: with one card in `run`, the same transition succeeds — so the refusal is demonstrably conditional, and the bound is 2 rather than 0.

**AC-KB-012** (REQ-KB-010) — *Given* a board under test, *when* the WIP bound is varied while every other input is held fixed, *then* the number of cards admitted to `run` tracks the WIP bound; and *given* the reverse direction, *when* the board's **inputs** are enumerated, *then* no deployed-session count appears among them — the admission path reads no session-count value, from configuration, from the topology, or from any observable, and the scan yields none. Decided by a table-driven test over the first direction and an absence scan over the second.

**Why the reverse direction is an absence claim rather than a varied knob.** v0.2.0 required varying "the deployed coder-session count across every value the topology admits (1 and 2)". Both the count and the topology are `SPEC-KANBAN-BOOTSTRAP-001`'s (`REQ-KS-005`) and are excluded by this SPEC's §C — so at this SPEC's run-phase the knob may not exist to vary, and a criterion that cannot be executed is not a criterion. It is also unnecessary: what independence *means* on this side is that the board never consults the count, and "never consults" is directly observable at the board's own boundary, which this SPEC does own. The stronger observation replaces a dependency on an excluded sibling with a claim about this package.

The forward direction stays a varied knob, because it is falsifiable on its own: v0.1.0 asserted only that the WIP limit *reads* 2, which observes a constant and cannot fail — an implementation deriving the limit from a session count and clamping it to 2 passes such a check perfectly (§A.3).

**Positive control, both halves.** Forward: a deliberately introduced derivation — the admitted-card count computed from a session count — makes the admitted count stop tracking the WIP bound, and the table-driven half reports it. Reverse: a deliberately introduced read of a session-count value in the admission path is reported by the absence scan. Each run once and recorded, so neither half is a scan that has never been shown to fire.

Where the session-count knob **is** reachable at run-phase — the sibling having landed — the end-to-end observation may be added: with the count at 1, two cards are admitted to `run` and one is unheld. It is recorded as a supplementary confirmation, never as this criterion's verdict, so that this criterion's executability does not depend on a sibling's state.

**AC-KB-013** (REQ-KB-011) — *Given* a board with one card already held in `run` and no session free, *when* a second card is admitted to `run`, *then* the admission succeeds and the card is recorded unheld rather than refused; and *when* that card is later read, *then* it is reported as a valid steady state rather than as an error or a stall. **Positive control**: a third admission is refused by the WIP limit, so admission is bounded by WIP and not by session availability.

**AC-KB-014** (REQ-KB-012) — *Given* a card in `backlog` and a card in `done`, *when* each is evaluated for dispatchability, *then* both are reported not dispatchable, because neither column has an owning session. **Positive control**: a card in `plan` is reported dispatchable, so the refusal is column-conditional.

---

## §G. Who writes the board, and under what exclusion

**AC-KB-017** (REQ-KB-017) — Two halves, and they are separated because they observe different kinds of thing.

**Static half.** *Given* the implementation, *when* every call site that opens, creates, truncates, or renames onto the board state path is enumerated, *then* **every one of them reaches the file through exactly one guarded entry point** — there is no second write path, and no call site writes the board state file directly. **Positive control**: a deliberately introduced direct write to the board state path, bypassing that entry point, is reported by the same enumeration, run once and recorded.

**Runtime half.** *Given* a session whose **declared role** — read through `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, never derived from a session identifier or a launch label — is not `lead`, *when* it attempts a board write through that entry point, *then* the write is refused and the board is byte-unchanged; and *given* a session declaring `lead`, *then* the same write succeeds. Both directions are required: a guard that refuses unconditionally satisfies the first alone.

**Carrier half (REQ-KB-025), added at v0.5.0, repaired at v0.6.0.** The runtime half above reads a declaration; this half decides that the thing it reads satisfies the contract rather than a narrowed form of it, and it is a separate observation because the runtime half cannot fail on the difference. *Given* the declaration the run-phase established, *when* it is resolved **from a session that is not the `lead`**, *then* it resolves to the same value the subject session declares; and *when* it is resolved **by the `lead`, for a session that is not the `lead`**, *then* it likewise resolves to that session's declared role. **Both directions are required, and the second was missing through v0.5.0.** They are different properties of a carrier, not one property observed twice: the first is workers-reading-lead, which `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` need; the second is lead-reading-workers, which is the key `REQ-KS-006` requires for dispatch selection (`REQ-KS-019`) and quorum accounting (`REQ-KS-012`). A carrier exposing only the `lead`'s own declaration satisfies the first, satisfies every other observation in this criterion, and fails the second — and it was the carrier v0.5.0's enumerated form of `REQ-KB-025` licensed. **Positive control**: a lead-only carrier is reported by the second direction and by no other observation here, run once and recorded. *Given* the same declaration, *when* it is compared against the session's launch label, *then* the two are distinct data and neither is computed from the other — decided by launching one role under two different labels and observing that the declared role is identical while the labels differ, which is the sibling's own `AC-KS-030` shape and is reproduced here because this SPEC now owns the carrier those observations run against. *Given* the implementation, *when* it is scanned for a declaration surface, *then* exactly one is found. **Positive control**: a deliberately introduced second declaration surface is reported by the same scan, run once and recorded.

**Why the carrier half cannot be folded into the runtime half.** The runtime half asks only what role *the calling session* occupies, so a session-private carrier passes it completely — every board-side observation succeeds, and the clauses that fail are never reached. Those clauses are what two other SPECs need: `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` resolve the `lead`'s occupant from a session that is not the `lead`, and the sibling's own research records a lead-private declaration as the rejected alternative for exactly this reason — it "satisfies every dispatch-routing test here and breaks two gates over there — a failure that surfaces in another SPEC's criteria, which is the kind least likely to be caught by the SPEC that caused it". This half is that catch, placed on the side that now owns the carrier.

**What the uniqueness scan can and cannot establish, stated so its passing is not over-read.** It runs at this SPEC's M1, and `SPEC-KANBAN-BOOTSTRAP-001` declares this SPEC among its own `dependencies:` — so at the moment this scan runs, that sibling has not landed and cannot have contributed a second declaration. The scan therefore decides that *this SPEC's own implementation* defines exactly one, and it structurally cannot decide the cross-SPEC uniqueness `REQ-KB-025` also asserts. That second question is decided on the sibling, by `AC-KS-030`'s adoption conjunct, which is where the implementer able to break it is reading. Recorded rather than left implicit, because a scan reporting one declaration would otherwise read as having cleared a fork it never had the opportunity to observe.

**What this criterion does not decide.** Which carrier was chosen. `REQ-KS-006` fixes none and `REQ-KB-025` fixes none; the observations resolve the declaration through whatever carrier the run-phase established and assert its **four** — workers-read-lead, lead-reads-workers, label-distinctness, and uniqueness. (This read "three properties" through v0.6.0, a count left over from before the lead-reads-workers direction was added; corrected at v0.6.1. A criterion that miscounts its own observations invites a run-phase to run one fewer.) A criterion asserting the choice would be fixing a run-phase decision on preference — the move `REQ-KS-006` itself declines — and would also break the adoption path, since a later landing of that sibling must be able to supply the carrier without failing a criterion written here.

**Why the split, and why v0.2.0's single form was not decidable.** It asked a *static* enumeration to establish that each write site "is reachable only from the session occupying the `lead` role". Role occupancy is a **runtime** value: the same code path executes in every session, so statically there is no `lead` call site and no non-`lead` call site to tell apart — and the stated positive control, "a board-state write from a non-`lead` call site", names something that does not exist statically at all. A criterion whose positive control cannot be constructed cannot fire, which is exactly the state §A.2 forbids. So the static half now asserts what static analysis can actually decide (one entry point, no bypass) and role enforcement moves entirely to the runtime half, where it is decided by varying the declared role.

The positive controls are mandatory here rather than customary. The defect this requirement repairs is a rule that existed in prose while nothing enforced it (`spec.md` §A.7), so a scan reporting zero that has never been shown to fire would reproduce exactly that state — a clean result standing in for an absent enforcement.

**AC-KB-018** (REQ-KB-018) — *Given* a board write, *when* the temporary file's directory is compared to the target file's directory, *then* they are equal; and *given* a reader repeatedly reading the board state while writes proceed, *when* each read is evaluated, *then* every read yields either the complete prior board or the complete new board — never a prefix, and never a parse failure.

Both halves are required and neither implies the other. Asserting only that a rename occurs passes an implementation whose temp file sits in the system temp directory: the rename still happens, it merely stops being atomic once it crosses a device boundary and degrades into a copy. Asserting only the directory equality establishes that a torn write is *unlikely*, not that a reader never observes one. The reader is the correct vantage point for the second half because a writer cannot observe its own torn write. Decided against `internal/atomicfile`'s `Replace` being the rename step, per REQ-KB-018's reuse constraint.

**AC-KB-019** (REQ-KB-019) — *Given* a board already holding one card in `run`, *when* **two separate OS processes** concurrently transition two **different** cards into `run`, *then* exactly one succeeds, the other is refused with the named WIP error, and the final board holds two cards in `run` — never three. **Positive control**: with the board holding zero cards in `run`, the same two concurrent transitions both succeed and the board holds two, so the refusal is conditional on the bound rather than on concurrency itself.

Two properties of this criterion carry its weight. **Separate processes**, because sessions are distinct OS processes and an in-process mutex passes a goroutine test while protecting nothing — `internal/lockfile/lockfile_windows.go` states that gap in its own package comment and forbids upgrading it, which is why REQ-KB-019 reuses the `internal/spec/lock.go` family instead (§A.4). **Two different cards**, because that is the case a card-scoped lock cannot cover: each process takes its own card's lock, each reads one card in `run`, each finds the bound satisfied, and both write. A test transitioning the *same* card twice is passed by a card lock and measures nothing about board scope.

**AC-KB-023** (REQ-KB-023) — Three observations, run in **separate OS processes** per §A.4, and the third is the one that decides the criterion.

*Given* a board-wide lock artifact left behind by a process that has terminated, *when* the clearing operation is invoked, *then* the artifact is removed, the removal is reported, and a subsequent board mutation acquires the lock. *Given* an artifact whose recorded process is **live and holding the lock**, *when* the same operation is invoked, *then* nothing is removed and the operation reports the refusal — the age of the artifact having no bearing on either outcome. *Given* an artifact inspected by the clearing operation and then, **before the pre-removal re-read**, released by its owner and re-acquired by a live process, *when* the removal step proceeds, *then* the re-read observes the changed identity, the clear aborts, and the re-acquired lock survives.

The first two observations are what an implementer writes unprompted, and both are satisfied by a clear that reads the recorded identity once and then unlinks. The third is the check-then-unlink race, and an implementation failing it has a *repair* that admits two writers to the critical section the lock exists to hold — the operation causing precisely the concurrency it was added alongside. It is invisible to any test that does not construct the window, which is why the window is constructed here rather than argued about.

**What this criterion does not decide, stated so its passing is not over-read.** It decides the **mitigation** — that a mismatch observed at the re-read aborts the clear — and it is satisfiable, because the re-acquisition is interleaved before the re-read. It does **not** decide that no window remains. A residual interval survives between the re-read and the unlink and cannot be closed at this layer: the removal primitive takes a path, resolves it at call time, and no portable handle-based form exists (`spec.md` §A.7(3a)). No observation is written against that residual, because a test that fails to hit a microsecond window is not evidence the window is absent, and a criterion asserting elimination would be asserting an outcome the platform cannot deliver. An implementation passing all three observations has the mitigation `REQ-KB-023` requires, and still carries the residual `REQ-KB-023` names.

**Positive control**: with the recorded process still live, the clear is refused — so removal is conditional on absence rather than unconditional. **Platform note, recorded rather than waived**: measured, the Unix substrate releases its `flock(2)` on process exit, so an artifact left there is inert and the first observation is trivially satisfiable without any clearing operation at all. The requirement exists for the Windows substrate, whose own header records that stale-lock cleanup is manual. A result obtained only on Unix is recorded as such and does not by itself establish the requirement.

## §H. Template mirror, neutrality, and verification

**AC-KB-015** (REQ-KB-014, REQ-KB-015) — *Given* a commit touching `internal/template/templates/`, *when* the committed tree is checked, *then* `internal/template/catalog.yaml` is regenerated and committed in the same change, decided by the repository's existing catalog-parity check whose exit code is the verdict; *when* each touched mirrored pair's pre-change `diff` (with this SPEC's own token substitutions applied) is compared to its post-change `diff`, *then* they are equal, and a pair measured sanitized before and byte-identical after **fails**; and *when* the repository's neutrality guard is run over every authored template file, *then* it exits 0. The guard is `internal/template/internal_content_leak_test.go` plus the neutrality workflow; its regex is **not** reimplemented here (§A.1 rule 5). Each pair's classification is re-measured at run-phase, not read from `spec.md`.

**AC-KB-016** (REQ-KB-016) — *Given* the run-phase tree, *when* the full suite is run with output redirected to a log, *then* `rc` is 0 and `grep -c '^FAIL'` over the **whole** log is 0; and *when* `spec lint` is invoked **per file** over the literal list `spec.md plan.md acceptance.md design.md research.md progress.md` — the full Tier L artifact set, since a list omitting `design.md` and `research.md` leaves two authored files unlinted — *then* each invocation exits 0. A count taken over a tail does not satisfy this criterion, and neither does a `*.md` glob (§A.1 rules 1 and 3).

---

## §I. Definition of Done

- All **25** criteria (`AC-KB-001` … `AC-KB-025`) executed, with each one's command and verbatim output recorded in `progress.md` §E.2. (This line read "All 24 criteria" through v0.4.0, a count left over from before `AC-KB-025` was added; corrected at v0.5.0. A Definition of Done naming fewer criteria than exist is a gate that passes with one unexecuted.)
- The full suite green per AC-KB-016; `spec lint` clean per AC-KB-016; the neutrality and catalog guards green per AC-KB-015.
- Every criterion carrying a **positive control** having had that control run once and recorded. An absence-check with no demonstrated ability to fire is not evidence of absence, and a refusal never shown to be conditional is not evidence of a bound.
- The three concurrency criteria — AC-KB-018, AC-KB-019, and AC-KB-023 — run with **separate OS processes**, not goroutines (§A.4). A goroutine result does not satisfy any of them.
- AC-KB-002's fallback-forced half run through the `execCommand` indirection, not by invoking the fallback directly; AC-KB-023 recorded with the platform it ran on, since the Unix substrate satisfies its first observation trivially.
- The five decisions of `plan.md` §E implemented as decided. They are settled, not open — a run-phase that re-opens one records the reason as a blocker rather than choosing differently in place.

## §J. Traceability

Every requirement maps to at least one criterion, and every criterion to at least one requirement.

| REQ | AC | Milestone |
|---|---|---|
| REQ-KB-001 | AC-KB-001 | M0–M3 (standing) |
| REQ-KB-002 | AC-KB-001 | M0 |
| REQ-KB-003 | AC-KB-007 | M2 |
| REQ-KB-004 | AC-KB-008 | M2 |
| REQ-KB-005 | AC-KB-002, AC-KB-003 | M1 |
| REQ-KB-006 | AC-KB-004 | M1 |
| REQ-KB-007 | AC-KB-005 | M1 |
| REQ-KB-008 | AC-KB-009, AC-KB-010, AC-KB-021 | M2 |
| REQ-KB-009 | AC-KB-011, AC-KB-019 | M2 |
| REQ-KB-010 | AC-KB-012 | M2 |
| REQ-KB-011 | AC-KB-013 | M2 |
| REQ-KB-012 | AC-KB-014 | M2 |
| REQ-KB-013 | AC-KB-006, AC-KB-020 | M1 |
| REQ-KB-014 | AC-KB-015 | M3 |
| REQ-KB-015 | AC-KB-015 | M3 |
| REQ-KB-016 | AC-KB-016 | M3 |
| REQ-KB-017 | AC-KB-017 | M1 |
| REQ-KB-018 | AC-KB-018 | M1 |
| REQ-KB-019 | AC-KB-019 | M1 |
| REQ-KB-020 | AC-KB-022 | M2 |
| REQ-KB-021 | AC-KB-024 | M1 |
| REQ-KB-022 | AC-KB-020 | M1 |
| REQ-KB-023 | AC-KB-023 | M1 |
| REQ-KB-024 | AC-KB-025 | M2 |
| REQ-KB-025 | AC-KB-017 | M1 |

**Reconciliation.** 25 requirements, `REQ-KB-001` … `REQ-KB-025`, each appearing exactly once in the left column. 25 criteria, `AC-KB-001` … `AC-KB-025`, each appearing at least once in the middle column. Against the Tier L ceiling of 25 and 25: **both are now at the ceiling and neither has headroom.** Reported rather than smoothed, because Tier L is the top tier — any further finding requiring a new requirement *or* a new observation has to be reported to the orchestrator rather than absorbed, and re-bundling an existing entry to make room would restore the F1 defect the v0.2.0 repair closed.

The v0.5.0 addition is `REQ-KB-025` alone — the last spare requirement — and **no criterion was added, retired, or merged**: its carrier half was authored into `AC-KB-017`, which already reads the declaration in its runtime half, so the two observations run against one subject and one fixture. `AC-KB-017` therefore now serves two requirements. They are kept on one criterion because both are decided by varying the same declaration through the same entry point; they are separate *requirements* because an implementation can satisfy `REQ-KB-017`'s refusal completely on a session-private carrier, which is what `REQ-KB-025` exists to exclude and what no board-side observation can see. The v0.4.0 additions were `REQ-KB-024` and `AC-KB-025`; the other v0.4.0 repairs were made by amending `REQ-KB-020`, `REQ-KB-023`, `AC-KB-022` and `AC-KB-023` in place, which is why two defects cost one requirement between them.

`AC-KB-020` serves two requirements: `REQ-KB-013` (the exit exists) and `REQ-KB-022` (the exit is bounded and records what it lost). They are kept on one criterion because the second is a property of the operation the first defines and both are decided by one invocation; they are separate *requirements* because an implementation can satisfy the first while failing the second, which is exactly what an unrecorded replacement does.

Three criteria serve more than one requirement: `AC-KB-001` (the rename gate and the identifier form, which one preflight decides), `AC-KB-015` (the mirror pair and its neutrality, which one commit decides), and `AC-KB-017` (the sole-writer refusal and the carrier it reads, which one declaration decides — added at v0.5.0). Four requirements take more than one criterion, each because it carries independently-failable halves: `REQ-KB-005` (the resolution and the absence of tree-relative paths), `REQ-KB-008` (the illegal-pair refusal, the collision ambiguity, and the status-value coverage), `REQ-KB-013` (detecting the unknown state and escaping it — a criterion covering only the first passes a brick, and one covering only the second passes a silent auto-repair), and `REQ-KB-009`, which appears against `AC-KB-019` because the WIP bound is enforceable only beneath the board-wide lock: `AC-KB-011` establishes the bound single-threaded, `AC-KB-019` establishes it holds under concurrency, and an implementation passing only the first admits a third card whenever two sessions act at once.

Four milestones, M0 … M3, with M1 grown at v0.2.0 to carry the sole-writer trio ahead of the M2 admission they make sound, and again at v0.3.0 to carry `REQ-KB-021` … `REQ-KB-023`. `REQ-KB-020` lands in M2 rather than M1, ahead of the compatibility table it feeds: the table is unusable without it, since a primary-side read pairs every card in `run` as `(run, draft)`. `REQ-KB-024` lands in M2 beside it and not after it, because it is the same resolution's other exit — an implementation that lands the resolution without it has no defined behavior on a card it cannot resolve, and the behavior it acquires by default is the silent `draft` `AC-KB-025` exists to exclude.

## §K. Out of Scope

### Out of Scope — what these criteria do not judge

- The correctness of `SPEC-KANBAN-RENAME-001`'s rename mapping. AC-KB-001 checks only that it landed on the base branch.
- How the `lead` role is elected, assigned, or handed over. Who becomes the `lead` is `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004`, and no criterion here judges it.
- **Which** carrier the role declaration rides — the launch command, the session registry, or the peer-discovery output. `REQ-KS-006` fixes none and `REQ-KB-025` fixes none; `AC-KB-017`'s carrier half resolves the declaration through whatever carrier the run-phase established and judges its **four** contract properties — workers-read-lead, lead-reads-workers, label-distinctness, and uniqueness — never the choice. (This read "three contract properties" through v0.6.1, the same leftover corrected at line 165 and in `spec.md` §D.14; all three surfaces agree at v0.6.2.) **Narrowed at v0.5.0**: through v0.4.0 this line excluded "by what mechanism a session's role declaration is carried" outright, which was correct while `SPEC-KANBAN-BOOTSTRAP-001` was in scope and left the carrier judged by nobody once it was not. The carrier's *existence and properties* are now judged (`REQ-KB-025`); its *identity* is not.
- The correctness of the role-declaration **contract** itself — that a role does not determine a label and a label does not determine a role. That argument is `SPEC-KANBAN-BOOTSTRAP-001`'s (`REQ-KS-013`, `REQ-KS-005`, `REQ-KS-014`) and is consumed as a conclusion here, never re-derived or re-tested.
- The correctness of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`'s branch resolution. AC-KB-022 judges that the board reads the branch that resolution names, not that the resolution names the right branch.
- `SPEC-KANBAN-RENAME-001`'s `.moai/state/kanban/` session-record path. AC-KB-002 judges only that the board does not use it.
- The correctness of `internal/atomicfile` and `internal/spec/lock.go` themselves. AC-KB-018 and AC-KB-019 judge this board's use of them, not the primitives, which carry their own tests.
- Anything owned by `SPEC-KANBAN-WORKTREE-001` or `SPEC-KANBAN-BOOTSTRAP-001` — worktree lifecycle, stall detection, holder release, topology, bootstrap, dispatch, backend selection. No criterion here judges a session's behavior; they judge the board's.
- Any board rendering surface, and the absence of a git-visible column history — the latter is an accepted cost per `spec.md` §A.3(c), not a defect to be measured.
