---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Acceptance criteria — Kanban session topology, bootstrap, and dispatch"
version: "0.5.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, bootstrap, dispatch, acceptance, given-when-then, recorded-baseline, sole-writer"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## §A. How these criteria are judged

Every criterion below is binary: it names a command or an observation, and a result that either holds or does not. "Looks correct" is never a verdict.

### A.1 Command hygiene (binding on every criterion here)

- An **absence** claim is never believed on its own. Each absence-grep is paired with a positive control — a deliberately introduced instance the same command reports — establishing that the command can fire. An absence-grep that has never been shown to fire is evidence of nothing.
- No `grep -E` pattern carries a raw `|` inside a markdown table cell. Commands needing alternation are written on their own fenced line.
- A guard cited by name is first confirmed to have exactly one `func` definition. This rule exists because a predecessor cited a guard that does not exist (`spec.md` §A.6).
- A scan is written for the direction the code actually accesses, not for the direction that is easiest to grep. A scan that cannot fire against the shape of the surface it governs reports a clean result on something it never examined — measured at `spec.md` §D.10, where the entry switch's existing environment mutation carries five env-access call sites and **zero** occurrences of `os.Getenv(`.
- An **equivalence** claim ("identical to before", "unchanged in meaning") is judged against a **durable artifact recorded before the change** and re-readable at verification time. Where no such artifact exists the criterion is unfalsifiable, because at verification time the original is gone and only assertion remains (`spec.md` §D.9).

### A.2 Positive controls required

Criteria AC-KS-008, AC-KS-009, AC-KS-010, AC-KS-015, AC-KS-016, AC-KS-020, AC-KS-025 and AC-KS-028 each assert that something is *not* present. Each carries an explicit positive control and fails if the control does not trip the check.

### A.3 Recorded baselines required

Criteria AC-KS-011, AC-KS-022, AC-KS-023 and AC-KS-024 each compare an implementation against a prior state. Each names the artifact holding that prior state and the point before which it was recorded, and each judges the artifact's **provenance** as well as its content — a baseline taken after the change agrees with the change it was meant to detect, and its passing comparison is a tautology.

---

## §B. Preconditions

**AC-KS-001** — *Given* a working tree at the base branch, *when* the M0 preflight of `plan.md` §C is run, *then* the rename package, the board sibling's card-record holder field, the messaging doctrine and its template mirror are all reported present, the guard-anchor definition count is exactly 1, and the single-origin discriminant `SPEC-KANBAN-BOARD-001` `REQ-KB-005` prescribes resolves to the primary checkout's git directory **from the checkout the preflight is run in, whichever that is**; and *when* any one of those checks reports absence, *then* the run halts with the absence named and no file is edited. (REQ-KS-002)

The discriminant is named by citation rather than written out, and the "whichever that is" clause is the whole of what this criterion adds. Measured, the bare `git rev-parse --git-common-dir` this criterion carried through v0.3.0 prints `/Users/goos/MoAI/moai-adk-go/.git` from a worktree and `.git` from the primary checkout, so as written it passed from a worktree and failed from the primary — deciding the check on where the preflight happened to run rather than on whether the discriminant works. `REQ-KB-005` owns the probe form; this criterion asserts only that it resolves, and it is run from both checkouts (`spec.md` §A.11).

**AC-KS-002** — *Given* the messaging doctrine as it actually landed, *when* its role-boundary-dispatch clause is diffed against the three conditions `spec.md` §A.9 was authored against, *then* either the clause is unchanged and the diff is empty, or the delta is surfaced as a blocker recorded in `progress.md` before any dispatch-protocol file is created. (REQ-KS-002)

**AC-KS-003** — *Given* the run-phase runtime, *when* the five substrate properties of `spec.md` §A.3 are exercised, *then* each of the five has a recorded result attributed to a command run in this session, and no property is carried forward from the prior session's measurement; and *when* the launcher surface of `spec.md` §A.8 is measured, *then* two further results are recorded — that the selected backend reaches the launched session through the environment the launcher constructs for it, on the launch path the host platform takes, and that each launcher additionally mutates project-global state the other undoes — each attributed to a command run in this session. (REQ-KS-003)

The launcher half is here because its absence is what let the backend decision stand on an unmeasured premise. Both halves are required: recording only the per-session mechanism would re-assert the assumption that the launchers do nothing else, and recording only the project-global mutation would suggest per-worker backends do not work, when measurement says they do (`research.md` §J).

The first result is judged on the **environment guarantee**, not on the name of the call that delivers it. `execOrSpawnClaude` has two build-tagged definitions — `syscall.Exec` on POSIX, an `exec.Command` child with an explicit `child.Env` on Windows — so a criterion satisfied only by observing `syscall.Exec` would be unsatisfiable on Windows against code that meets the requirement exactly (`spec.md` §A.8). What is recorded is that the constructed environment reaches the launched backend on the path this host takes, with the definition that path resolves to named in the record.

---

## §C. Topology

**AC-KS-004** — *Given* a loaded topology, *when* its role set is enumerated, *then* it contains exactly `lead`, `plan`, `run`, `review`, `sync`; the `lead` entry owns no column and has no card-work entry point; and no session entry exists for `backlog` or `done`. (REQ-KS-004)

**AC-KS-005** — *Given* topology configurations declaring a `run`-session count of 1, 2, 0, and 3, *when* each is loaded, *then* 1 and 2 load successfully, 0 and 3 are each rejected with a named error, the rejected loads produce no clamped value anywhere in the resulting configuration, and the default with the key absent is 1. (REQ-KS-005)

**AC-KS-006** (REQ-KS-006, label half) — *Given* a bootstrap run over a five-role topology, *when* the emitted guidance is parsed, *then* every emitted launch command carries a label argument, and the label set across all emitted commands has no duplicate.

**AC-KS-030** (REQ-KS-006, role-declaration half) — *Given* a launched session, *when* its declared role is resolved **from the lead**, *then* it resolves to exactly one of the five roles; *when* the same declaration is resolved **from a session that is not the lead**, *then* it resolves to the same value; *given* an unconfigured worker role whose two emitted commands were each launched in turn, *when* the declaration is resolved for each, *then* the declared role is identical across both while the labels differ; and *given* the launched set, *when* the lead accounts for quorum, *then* the accounting is over declared roles and a role is reported answered only where a session declares it. All four hold, or the criterion fails.

Separate from AC-KS-006 rather than folded into it, because the content of the repair is that the declaration and the label are different data (`spec.md` §A.13). One verdict over both would let the declaration half break while the label half passed — which is the bundling defect the v0.2.0 pass removed from `AC-KS-024`, reintroduced at a different slot.

The non-lead resolution is the row most likely to be dropped and the most expensive to drop: an implementation holding the declaration where only the lead can read it satisfies this SPEC's own dispatch routing and silently defeats `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011`, both of which resolve the `lead` occupant from a session that is not it. The two-label row is what makes the non-derivation observable: an implementation that populated the declaration by parsing the label passes every other row and fails this one.

---

## §D. Entry switch, configuration, and the prompt boundary

**AC-KS-007** — *Given* a project with a topology configuration present, *when* the entry switch runs, *then* it prints bootstrap guidance and enters the quorum poll; and *given* the same project with the configuration absent, *when* the entry switch runs, *then* it enters the single-session chain and prints no bootstrap guidance. (REQ-KS-007)

**AC-KS-008** — *Given* the shipped surface, *when* the configuration file is located, *then* it resolves beneath `.moai/config/`; the `moai init` question flow contains a question that writes it; and a grep of the entry-switch argument parser for any topology parameter returns zero hits, with a positive control confirming the grep can fire. (REQ-KS-008)

**AC-KS-009** — *Given* the source files this SPEC authors, *when* every environment-access call site in them is enumerated **across all five access forms** — the reads `os.Getenv` and `os.LookupEnv`, and the writes `os.Setenv`, `os.Unsetenv` and a test's `Setenv` — *then* each site's key argument is an identifier resolving to a constant declared in `internal/config/envkeys.go`, no site passes a string literal, and every new environment-variable name has such a constant; and *given* a deliberately introduced literal-keyed access inserted **once as a write and once as a read**, *when* the same enumeration runs, *then* it reports both. (REQ-KS-009)

Both directions are required, and the write direction is the one that carries the criterion. REQ-KS-009 binds every call site, and the surface it governs writes far more than it reads: measured in this worktree, `internal/cli/factory.go` — the existing entry-switch mutation this SPEC extends — carries `os.Setenv` at lines 93, 95 and 110, `os.LookupEnv` at 107 and `os.Unsetenv` at 113, while `grep -c 'os.Getenv('` over it returns **0**. The v0.1.0 criterion scanned for `os.Getenv("` alone, so against that shape of code it was vacuous rather than merely weak — a violation authored as `os.Setenv("MOAI_KANBAN", …)` passed it untouched, and the positive control it lacked would have exposed that immediately. Reads remain in scope because the requirement covers them, not because the read scan was sufficient; narrowing this criterion back to `os.Getenv` would re-open the defect (`spec.md` §D.10).

**AC-KS-010** — *Given* the bootstrap source file, *when* the new static guard is run, *then* it reports no reference to `AskUserQuestion` and none to `mcp__askuser`; and *given* a deliberately inserted reference to each, *when* the same guard is re-run, *then* it fails on both — establishing the guard can fire. The guard's own function definition count is exactly 1, and the anchor it was written from was confirmed to have exactly 1 definition before it was copied. (REQ-KS-010)

---

## §E. The backward-compatibility baseline

**AC-KS-011** — *Given* the recorded baseline artifact, *when* its contents are inspected, *then* every identifier in it is a post-rename identifier and no retired identifier appears — establishing it was recorded against the post-rename tree; and *given* a project with no topology configuration, *when* the entry switch's parse result, environment mutation, and exec argument vector are compared table-driven against that baseline, *then* all three are equal. A baseline failing the first half is discarded and re-recorded rather than reconciled, and the second half's result against a discarded baseline is not counted. (REQ-KS-011)

---

## §F. Bootstrap, quorum, and the emitted command set

**AC-KS-012** — *Given* a topology whose roles all answer within the bound, *when* the bootstrap runs, *then* it proceeds and exits zero; and *given* the same topology with one role deliberately absent and the bound configured short, *when* the bootstrap runs, *then* it exits non-zero, its output names the absent role and the answering roles in separate groups, and no dispatch was sent. The configured bound with the key absent reads 300 seconds, and the bound is not expressible as an entry-switch argument. (REQ-KS-012)

This criterion decides the bound and the expiry behaviour, and deliberately says nothing about **what "answered" is counted over**. That key is REQ-KS-006's — quorum is accounted over declared roles, not over launch labels — and it is decided by the fourth conjunct of `AC-KS-030` (`spec.md` §D.11). The two are kept apart because they fail apart: a bound that never fires and an accounting that counts labels are different defects, and an implementation can pass either while failing the other.

**AC-KS-013** — *Given* a topology configuration naming a backend for two of the four worker roles, *when* the guidance is emitted and parsed into a command set, *then* the `lead` contributes exactly one command in the fixed Claude-backed form; each of the two configured worker roles contributes exactly one command in its configured backend; and each of the two unconfigured worker roles contributes exactly two commands whose backends are exactly the two supported ones. (REQ-KS-013)

**AC-KS-014** — *Given* the emitted guidance, *when* each line is parsed as an executable invocation, *then* every line parses, and the multiset of labels across all lines contains no duplicate. Judgment is on the parse result, not on reading the text. (REQ-KS-014)

**AC-KS-015** — *Given* the emitted guidance for every topology permutation exercised by AC-KS-013, *when* it is scanned for the rejected mixed backend, *then* zero lines name it; *given* a deliberately injected mixed-backend line, *when* the same scan runs, *then* it reports the line. And *given* an attempt to enter the board under the mixed backend, *when* it runs, *then* it is rejected with the `KANBAN_MODE_UNSUPPORTED_BACKEND` sentinel unchanged from its pre-change text. (REQ-KS-015)

**AC-KS-029** (REQ-KS-013, refusal half) — *Given* a topology configuration naming a backend for the `lead` role, *when* it is loaded, *then* the load fails with a named error and no configuration value is produced; *given* a backend supplied as an entry-switch argument, *when* the launcher parses it, *then* the parse fails because no such parameter is defined; and *given* one of this SPEC's environment variables set to a backend value, *when* the guidance is emitted, *then* the `lead` line is byte-identical to the line emitted with that variable unset. Each of the three is decided separately and all three must hold.

The configured channel is the load-bearing row. A key that is silently dropped and a key that is silently honored are indistinguishable from outside, so "no error and no effect" is not evidence of refusal — only the named error distinguishes a configuration the system rejected from one it never read. The environment row is the converse case and is checked by equality against the unset-variable output rather than by an error, because ignoring is the correct response there.

This criterion deliberately makes **no claim** about a `lead` session launched by hand under the other launcher. The entry switch prints guidance; it does not supervise what an operator subsequently types, and no observation available here could decide that case. What is verified is that the system offers no path to a non-Claude lead — not that no such session can exist (`spec.md` §A.12, §D.8). Recording the boundary is the point: a criterion claiming to close it would be claiming an observation nobody can make.

---

## §G. Dispatch protocol

**AC-KS-016** — *Given* the dispatch payload type, *when* its fields are enumerated, *then* the field set is exactly a SPEC identifier, a file path, and a contract section reference, and no field has a free-text type; and *given* a deliberately added free-text field, *when* the same field-set check runs, *then* it fails — establishing the check can fire. (REQ-KS-016)

**AC-KS-017** — *Given* a peer discovered through the discovery tool, *when* a dispatch is addressed, *then* the send carries both the peer's label and its short reference; and *given* a send refused for want of a reference, *when* the lead handles the refusal, *then* it re-sends using the reference the refusal supplied, and does not report the peer unreachable. (REQ-KS-017)

**AC-KS-018** — *Given* a dispatched card, *when* the same scenario is run three times — the reply arriving promptly, arriving after a delay, and never arriving — *then* the board state is byte-identical across all three runs, the card advances only when its `progress.md` carries the corresponding evidence, no run blocks waiting on a reply, and no worker session wrote board state in any of the three runs — the worker sessions writing only their own `progress.md`. That only the `lead` may write is `AC-KB-017`'s verdict, not this one's; what is checked here is that the dispatch protocol produced no worker write. (REQ-KS-018)

**AC-KS-019** — *Given* a card in a column owned by a role, *when* the lead selects a dispatch target, *then* the target is the session declaring that role — "declaring" in the sense REQ-KS-006 defines and AC-KS-030 observes, which until v0.3.0 was a phrase this criterion used and nothing defined; *given* an idle session declaring a different role, *when* the same selection runs, *then* that session is not selected; and *given* a card the board reports not dispatchable, *when* the lead runs, *then* no session is addressed for it; and *given* a **dispatchable** card whose column's owning role is declared by no session in the launched set, *when* the lead selects a dispatch target, *then* no session is addressed for that card **and** the lead surfaces both the unoccupied role and the card awaiting it — a run in which nothing is addressed and nothing is surfaced **fails** this criterion, the refusal alone being entailed by the first clause above and the silence being the failure shape the clause was added to convert (`spec.md` §D.12); and *given* a session that has received a dispatch, *when* the board-write paths reachable from it are enumerated, *then* none is — the dispatch conferring no write authority, per the deferral of `spec.md` §C to `REQ-KB-017`. Whether that enumeration is exhaustive is `AC-KB-017`'s question, not this one's; what is checked here is that the dispatch protocol adds no path. (REQ-KS-019)

**AC-KS-020** — *Given* the dispatch surface, *when* it is scanned for an operator-decision path routed through a peer, for treatment of a peer reply as approval for a gated action, and for a request that a peer perform an action the lead may not perform, *then* all three scans return zero hits; and *given* a deliberately introduced instance of each, *when* the same scans run, *then* each reports its instance. (REQ-KS-020)

**AC-KS-021** — *Given* a card already in progress, *when* a dispatch for it is re-sent, *then* the board state after the re-send is byte-identical to the state before it, and the card's holder, column, and last-transition instant are all unchanged. A criterion satisfied only by the absence of an error does not count. (REQ-KS-021)

---

## §H. The coder session's internal chain

**AC-KS-022** — *Given* the chain-behavior baseline artifact recorded before this SPEC's first edit to the chain — holding the verify exit gate's severity partition, rung attribute and re-entry ceiling, the revision dedup predicate's inputs and matching rule, and the goal preset's arming rules — *when* its provenance is checked, *then* it is present, readable, and its recording point precedes the first chain edit; and *given* a coder session driving a card, *when* its chain is exercised, *then* the gate fires at the exit of the card's run-phase producing behavior equal to the recorded partition, attribute and ceiling; the predicate's inputs and matching rule equal the recorded ones; the preset drives one session over one card with the recorded arming rules; and a scan for a second chaining mechanism returns zero hits. (REQ-KS-022)

Without the artifact this criterion cannot fail. "Producing the same observable behavior as the single-session mode" names a comparand that, by the time the comparison is made, exists only in the verifier's memory of code that has since been edited — and a criterion whose comparand is a recollection is satisfied by every implementation, including one that quietly widened the re-entry ceiling. The artifact is what converts the claim into an observation (`spec.md` §D.9).

**AC-KS-023** — *Given* the human-gate inventory recorded before this SPEC's first edit — each gate's identity, its position in the ordering, and what it gates — *when* its provenance is checked, *then* it is present, readable, and its recording point precedes that first edit; and *when* the gates are enumerated after the change and compared row by row against it, *then* the count, the ordering, and each gate's semantics are equal to the recorded inventory, and no row is present after that is absent before. (REQ-KS-023)

This is the equivalence claim whose silent violation is most expensive, because a dropped gate produces no error and no failing test — it produces work shipped past a review nobody decided to remove. An enumeration compared against nothing would report that a gate exists without establishing that it is the same gate in the same place; the recorded inventory is what makes "unchanged" a decidable word.

---

## §I. Mirror, neutrality, and verification

> **Why this is four criteria.** `REQ-KS-024` absorbs three separable predecessor requirements (`spec.md` §B.8), and at v0.1.0 a single criterion returned one verdict for seven observations — so six could be broken while it passed. The four below partition them along the lines that fail independently (`spec.md` §D.7).

**AC-KS-024** (REQ-KS-024, delta-preservation half) — *Given* the pair-classification artifact recorded before this SPEC edited any mirrored pair — each touched pair's class and its pre-change `diff` — *when* its provenance is checked, *then* it is present, readable, and its recording point precedes the first template edit; and *when* each pair's post-change `diff` is compared against its recorded one with the change's own token substitutions applied, *then* they are equal for every pair, a pair recorded byte-identical is still byte-identical, and a pair recorded sanitized still retains exactly the content its template side strips. Decided table-driven, one row per touched pair, each row naming its recorded class.

The two class outcomes stay in one criterion because they are one property evaluated per pair, not two properties: each row asserts *this pair's recorded relationship still holds*, and a sanitized pair that has become byte-identical fails its own row rather than a separate criterion. What the v0.1.0 form lacked was the comparand — "the `diff` taken before the change" with nothing obliging anyone to keep it. Pair classification is time-varying and each class is read from the artifact recorded at run-phase, never from `spec.md` or `plan.md`.

**AC-KS-026** (REQ-KS-024, neutrality half) — *Given* the completed change, *when* `internal/template/internal_content_leak_test.go` is run and, separately, `.github/workflows/template-neutrality-check.yaml` is run, *then* each exits zero, and **each exit code is recorded separately**. An aggregate report ("neutrality passed") does not satisfy this criterion, because the two are independent authorities on different triggers and one passing is no evidence about the other. Neither is reimplemented: the guards' own exit codes are the verdict, and a hand-rolled regex without their exemption lists is a false-failure machine.

**AC-KS-027** (REQ-KS-024, catalog half) — *Given* the template edits, *when* `make build` is run and the working tree inspected, *then* `internal/template/catalog.yaml` differs from its pre-change content, that regenerated file is present in the commit, and no template file was edited in a commit that omits it. A stale-but-committed catalog is the failure this checks: it is invisible to the neutrality guards and to the test suite, and surfaces only as a template that ships without its entry.

**AC-KS-028** (REQ-KS-024, content-isolation half) — *Given* every file authored or modified under `internal/template/templates/`, *when* each is scanned for a SPEC identifier, a REQ or AC token, an internal date, and a commit SHA, *then* all four scans return zero hits; and *given* one deliberately inserted instance of each of the four, *when* the same scans run, *then* each is reported. The four are enumerated rather than aggregated because they are four patterns and a scan silently missing one would report clean.

**AC-KS-025** — *Given* the completed implementation, *when* the full test suite is run — not an affected-packages subset — *then* it passes, and the command and its exit code are recorded verbatim. Every occurrence of `factory` introduced by this SPEC's own authored files is zero, checked by grep with a positive control. (REQ-KS-001, REQ-KS-025)

---

## §J. Traceability

| Requirement | Criterion |
|---|---|
| REQ-KS-001 | AC-KS-025 |
| REQ-KS-002 | AC-KS-001, AC-KS-002 |
| REQ-KS-003 | AC-KS-003 |
| REQ-KS-004 | AC-KS-004 |
| REQ-KS-005 | AC-KS-005 |
| REQ-KS-006 | AC-KS-006, AC-KS-030 |
| REQ-KS-007 | AC-KS-007 |
| REQ-KS-008 | AC-KS-008 |
| REQ-KS-009 | AC-KS-009 |
| REQ-KS-010 | AC-KS-010 |
| REQ-KS-011 | AC-KS-011 |
| REQ-KS-012 | AC-KS-012 |
| REQ-KS-013 | AC-KS-013, AC-KS-029 |
| REQ-KS-014 | AC-KS-014 |
| REQ-KS-015 | AC-KS-015 |
| REQ-KS-016 | AC-KS-016 |
| REQ-KS-017 | AC-KS-017 |
| REQ-KS-018 | AC-KS-018 |
| REQ-KS-019 | AC-KS-019 |
| REQ-KS-020 | AC-KS-020 |
| REQ-KS-021 | AC-KS-021 |
| REQ-KS-022 | AC-KS-022 |
| REQ-KS-023 | AC-KS-023 |
| REQ-KS-024 | AC-KS-024, AC-KS-026, AC-KS-027, AC-KS-028 |
| REQ-KS-025 | AC-KS-025 |

**Reconciliation.** 25 requirements, `REQ-KS-001` … `REQ-KS-025`, each appearing exactly once in the left column. 30 criteria, `AC-KS-001` … `AC-KS-030`, each appearing at least once in the right column. Every requirement has at least one criterion; every criterion names at least one requirement.

**Ceiling.** The criterion count exceeds the Tier L ceiling of 25 by five, and Tier L is the top tier, so no promotion is available. The five are `AC-KS-026` … `AC-KS-028` (three of the seven observations `AC-KS-024` previously bundled), `AC-KS-029` (REQ-KS-013's missing refusal observation), and `AC-KS-030` (the role-declaration observation added at v0.3.0). Re-bundling them to fit the ceiling would restore exactly the defects the v0.2.0 and v0.3.0 audit repairs closed — `AC-KS-030` in particular exists because the declaration and the label are different data, so folding it back into `AC-KS-006` would erase the distinction the repair establishes. The excess is reported here rather than hidden; the disposition is the orchestrator's decision. The requirement count is unchanged at 25 — every v0.2.0 and v0.3.0 repair is an amendment in place, the role-resolution repair being authored as a widening of `REQ-KS-006` because the ceiling left no other shape available.

---

## §K. Exclusions

### Out of Scope — criteria deliberately not written

- A criterion asserting the board's WIP limit, its column↔status check, or its single-origin state resolution. Those are `SPEC-KANBAN-BOARD-001`'s criteria; duplicating them here would create two places for one contract to drift.
- A criterion asserting worktree creation, disposal, orphan classification, or holder-lock contention. Those are `SPEC-KANBAN-WORKTREE-001`'s.
- A criterion citing `TestNew_NoAskUserQuestion`. It has no definition in this tree (`spec.md` §A.6), so a criterion naming it would be unsatisfiable — which is worse than shipping no criterion at all, because it fails for a reason unrelated to the property it claims to check.
- A criterion asserting that a `column` field is written to SPEC frontmatter. That field is rejected; `REQ-KB-007` forbids the board from writing SPEC frontmatter at all.
- A criterion asserting that only the `lead` writes board state, that a board write is atomic, or that a board mutation holds the board-wide lock. Those are `SPEC-KANBAN-BOARD-001`'s (`AC-KB-017`, `AC-KB-018`, `AC-KB-019`). This SPEC's criteria establish that the protocol *confers* no board-write authority on a dispatch recipient (AC-KS-019) and that a card's progression is reported through `progress.md` (AC-KS-018) — the deferral of §C, checked from this side without restating the sibling's rule.
- A criterion asserting that a `lead` session launched by hand under the other launcher is refused. No observation available to this SPEC decides it, and AC-KS-029 records the boundary instead of claiming it.
- A criterion asserting how the role declaration is carried — in the launch command, the session registry, or the peer-discovery output. REQ-KS-006 fixes the contract and deliberately fixes no carrier, so AC-KS-030 resolves the declaration through whatever carrier the run-phase chose and asserts nothing about which one it is.
- A criterion asserting the mechanism by which an operator admits a card from `backlog` into `plan`. `spec.md` §A.4 states that the admission is an operator act rather than a dispatch; the mechanism is owned by no SPEC in this family (`plan.md` §B.4), and a criterion here would be asserting a contract that does not yet exist.
- A criterion measuring how long any milestone takes. Ordering is by reversibility; there are no durations to assert.
