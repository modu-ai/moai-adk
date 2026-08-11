---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Progress — Kanban session topology, bootstrap, and dispatch"
version: "0.5.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, progress, plan-phase, audit-repair"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## §E.1 Plan-phase Audit-Ready Signal

- Tier: L. Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.
- Requirements: 25 (`REQ-KS-001` … `REQ-KS-025`) — at the Tier L cap of 25, unchanged at v0.2.0 and again at v0.3.0. Acceptance criteria: 30 (`AC-KS-001` … `AC-KS-030`) — **five over the Tier L cap of 25**.
- **Ceiling overflow, reported not absorbed.** Tier L is the top tier, so no promotion is available. The excess is `AC-KS-026` … `AC-KS-028` (three of the seven observations `AC-KS-024` previously reported through a single verdict), `AC-KS-029` (the refusal observation REQ-KS-013 lacked), and `AC-KS-030` (the role-declaration observation added at v0.3.0). Re-bundling them to fit the cap would restore the defects these revisions repair. The disposition — carry the excess, split this SPEC further, or accept a bundled criterion — is the orchestrator's decision, not this document's. Precedent: `SPEC-KANBAN-WORKTREE-001` v0.2.0 carries a two-requirement overflow on the same terms.
- **The requirement cap bound at v0.3.0 and held, and bound again at v0.5.0.** The role-resolution repair needed a runtime contract that no SPEC in the family owned; with 25 of 25 requirements used, it was authored as a widening of `REQ-KS-006` rather than as a twenty-sixth requirement. At v0.5.0 the same move closed the vacancy-dispatch gap through `REQ-KS-019`, after v0.4.0 had released it on the ground that no widening was available — a judgement the v0.5.0 entry below re-examines and reverses for half the area. Where a widening is genuinely unavailable the finding is reported to the orchestrator instead, which is what remains true of the recovery half.
- Split from the superseded `SPEC-KANBAN-MULTISESSION-001` (59 requirements, plan-audit FAIL 0.87), alongside `SPEC-KANBAN-BOARD-001` and `SPEC-KANBAN-WORKTREE-001`.
- Three corrections landed with the split and each is argued in `spec.md`: the dead `TestNew_NoAskUserQuestion` guard citation (§A.6), the missing baseline recording window (§A.7), and the rejected `column:` frontmatter field (§A.11).
- Unresolved `[NEEDS CLARIFICATION]` markers: none.

### v0.5.0 — plan-audit delta repair (one finding), no requirement and no criterion added

The confirming audit of v0.4.0 returned PASS-WITH-DEBT 0.885 with MP-6 passing. Its one major finding is that v0.4.0's defect (4) released more of the vacancy area than the ceiling actually forced.

1. **D1 — a widening was available after all, for the dispatch half.** v0.4.0 released the area on the ground that owning it "would need a new obligation, not a widening", because detection is a runtime-lifecycle act with no natural host. That treats detection as separate from selection. It is not, at dispatch time: the lead resolves the owning role's occupant on **every** dispatch, so the vacancy is the null result of the lookup `REQ-KS-019` already governs — same lookup, same moment. `REQ-KS-019` is widened in place with a **where**-clause (no dispatch into a vacancy; surface the unoccupied role and the waiting card), the same in-place move v0.3.0 made on `REQ-KS-006`.
2. **The criterion judges the report, not the refusal.** The refusal half is entailed by REQ-KS-019's existing "dispatch **only** to the session whose declared role owns the column", so a do-nothing implementation satisfies it — and that do-nothing reading is the failure the clause exists to rule out. `AC-KS-019` gains a conjunct that fails a run in which nothing is addressed **and** nothing is surfaced; the reasoning is at `spec.md` §D.12, including why the neighbouring not-dispatchable refusal correctly carries no surfacing duty (the board reports that verdict itself; it reports a vacancy nowhere).
3. **The exclusion is narrowed, not deleted.** Still unowned: **recovery** (re-establishing an occupant — re-quorum, targeted relaunch, or an operator act), and **detection independent of a dispatch attempt**. The second is stated as a limit rather than smoothed over: a role vacated while its column holds no dispatchable card goes unreported, and is also not stalling anything, so the reported set and the costly set coincide. `plan.md` §B.6 and its §E decision row are rewritten; the three-way decision recorded there for the next revision is now two-way.
4. **Budget, re-measured in this worktree after the edit.** `grep -cE '^\*\*REQ-KS-[0-9]{3}\*\*' spec.md` → **25**; `grep -cE '^\*\*AC-KS-[0-9]{3}\*\*' acceptance.md` → **30**. Both unchanged from v0.4.0; the five-criterion overflow remains v0.3.0's and is not widened here.
5. **Artifact versions.** `spec.md`, `plan.md`, `acceptance.md` and this file move to `0.5.0`; each was edited. `design.md` stays at `0.3.0` and `research.md` at `0.4.0` — neither was edited. Measured: `grep -cin 'vacan' design.md research.md` returns **0** in both. `grep -c 'KS-019'` returns **0** for `design.md` and **1** for `research.md` — the single hit is §D line 112, which quotes REQ-KS-019's declared-role clause ("the session whose **declared role** owns the card's column") while counting the four consumers of a runtime role lookup. That clause is carried through this revision verbatim; the where-clause is appended after it and changes no word of it, so the citation stays exact and the file needs no edit.
- Unresolved `[NEEDS CLARIFICATION]` markers: none.

### v0.4.0 — plan-audit delta repair (D1-D4), under an explicit override

The audit exhausted its three attempts on this SPEC and closed on a must-pass failure (MP-6). The repair proceeds under an explicit user override. Four defects, closed with **no requirement and no criterion added** — the SPEC is at the requirement ceiling and already five criteria over it, so every closure is an amendment in place. Three of the four share one shape: a sibling's rule restated instead of cited, an implementation named instead of the guarantee stated, and a disclaimer written instead of an adjacent criterion read.

| ID | Defect | Closure |
|---|---|---|
| D1 (critical) | §A.11 carried its own copy of `REQ-KB-005` — `.moai/state/kanban/` as the store, "the parent of `git rev-parse --git-common-dir`" as the resolution — and the sibling changed both underneath it, moving the store to `.moai/state/kanban-board/` (its §A.3(e)) and forbidding the bare probe standing alone. Measured before the repair: `grep -rc 'kanban-board' .` returned **0** in all six files, so this SPEC was naming `SPEC-KANBAN-RENAME-001` `REQ-KR-009`'s **session-record** path as the board's. The stale probe reached the execution gates, where it fails in the direction that hides — `git rev-parse --git-common-dir` prints an absolute path from a worktree and a repository-relative `.git` from the primary checkout, so `AC-KS-001` passed or failed on where the preflight happened to run | restatement deleted, citation in its place, at `spec.md` §A.11 (rewritten, with the two-checkout measurement recorded), `AC-KS-001` (discriminant named by citation; the "whichever checkout" clause added and the criterion run from both), `plan.md` §C check 5 (probe replaced with the `--path-format=absolute` form `REQ-KB-005` prescribes, run from both checkouts) and constraint C8, and `research.md` §D |
| D2 (major — the MP-6 must-pass failure) | normative text named `syscall.Exec` as the mechanism by which a per-session backend reaches the launched session. Measured, `launcher.go:791` calls `execOrSpawnClaude`, which is build-tagged two ways: `launch_exec_posix.go` (`//go:build !windows`) calls `syscall.Exec`, and `launch_exec_windows.go` (`//go:build windows`) spawns an `exec.Command` child with `child.Env = env`, its own comment recording that `syscall.Exec` returns `syscall.EWINDOWS` there. The conclusion drawn — the constructed environment reaches the backend, so interleaved launches give each worker its own — holds on both platforms; only the mechanism was wrong, and naming it prescribed on Windows the exact call that file exists to avoid | rewritten on the **environment guarantee** rather than the call: `spec.md` §A.8 (with the platform split measured and recorded), the §A.8 consequence bullet ("already been exec'd" → already been launched), `REQ-KS-003`, and `AC-KS-003` plus a new note explaining why a `syscall.Exec` observation would be unsatisfiable on Windows against conforming code. The POSIX call now appears only as one platform's implementation, beside the Windows one, at `spec.md` §E (new cross-reference for the two build-tagged files), `plan.md` §C check 8 (probes both definitions) and §C's M0 narration, and `research.md` §J.1 |
| D3 (minor, blocking) | §D.11 stated that quorum accounting "is REQ-KS-012's criterion and is not re-checked here" — wrong twice. `AC-KS-030`'s **fourth conjunct** performs exactly that check, three lines above in the same document; and `AC-KS-012`, the criterion it pointed at, carries no declaration-related clause at all. Reading all three, `AC-KS-012` is **correct as written**: the bound and the expiry are REQ-KS-012's subject, while the accounting *key* (declared roles, not labels) is REQ-KS-006's. The defect is the disclaimer alone | §D.11's third paragraph rewritten to record the split and name `AC-KS-030`'s fourth conjunct as the check; a pointer added beneath `AC-KS-012` stating what it deliberately does not decide and where that is decided. Three surfaces now agree; no criterion changed its subject |
| D4 (major) | the seventh unowned area. Quorum is bootstrap-scoped — `REQ-KS-007` waits for it, `REQ-KS-012` bounds that wait, neither is evaluated again — while both siblings' §C hand "the quorum bound" to this SPEC by name and this SPEC's §C carried no clause answering that. A session dying after bootstrap leaves its role unoccupied; `REQ-KW-011` releases the card's holder and leaves the column unchanged, `REQ-KW-012` makes a clean orphan "immediately re-dispatchable", and `REQ-KS-019` then dispatches to the session whose declared role owns that column — the dead one. Neither sibling requirement is wrong; the missing step is observing the vacancy | **released explicitly** rather than owned, the ceiling forbidding a new requirement: a `spec.md` §C exclusions entry, headed for the vacated-role case, naming what is unowned, why it cannot be owned here, the symptom a reader will hit (the §A.5 silent stall, arriving after bootstrap instead of at it), the check to run first (resolve `REQ-KS-006` declarations across the launched set), and the operator workaround (§A.5's relaunch-and-re-run); `plan.md` §B.6 out-of-band note with the three-way decision the next revision faces; a §E decision row. Shape matched to `plan.md` §B.4's treatment of the unowned `backlog → plan` admission |

Counts re-measured after the edits, in this worktree:

```
$ grep -cE '^\*\*REQ-KS-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KS-[0-9]{3}\*\*' acceptance.md
30
```

Against the Tier L ceiling of 25 and 25: **requirements exactly at the ceiling with nothing to spare, criteria five over.** Both figures are unchanged from v0.3.0 — no requirement was added and the criteria overflow was not widened. Both sequences are contiguous with no duplicates (`REQ-KS-001` … `REQ-KS-025`, `AC-KS-001` … `AC-KS-030`), confirmed by the unique-prefix count matching the total in each file. The ceiling is what forced D4 into an exclusion rather than a requirement, and that constraint is stated at the point of release rather than left for a reader to infer.

**One premise of the repair brief was contradicted by measurement and is recorded rather than smoothed.** The brief located `syscall.Exec` in "the normative text of §A.8, `REQ-KS-003`, and `AC-KS-003`". Measured, `grep -rc 'syscall.Exec' .` before the repair returned 6 occurrences distributed as `research.md` 2, `spec.md` 3, `plan.md` 1 — and **zero** in `acceptance.md`. Neither `REQ-KS-003` nor `AC-KS-003` named the call; both said "the process environment it execs into", which is a milder form of the same defect (the verb is still POSIX-shaped and Windows spawns rather than execs) and both were repaired on that ground. The three `spec.md` occurrences were §A.8's normative prose, a §E cross-reference, and the v0.3.0 HISTORY entry. The classification of all six, and the disposition of the HISTORY occurrence, is in the deliverable notes below.

**The v0.3.0 HISTORY entry is left unamended.** Its D2 narration says the launcher "execs into `claude`", which is the POSIX reading this revision corrects. It is a record of what was concluded at v0.3.0 and rewriting it would erase the error rather than close it; the correction is stated in the v0.4.0 entry instead, which is where a reader looking for the current position goes.

**Artifact versions.** `spec.md`, `plan.md`, `acceptance.md`, `research.md` and this file move to `0.4.0`; each was edited. `design.md` stays at `0.3.0` — it was read for both D1 and D2 and needed neither. Its §F describes board state as having "a single origin under the primary checkout, resolved through the common git directory" — a citation-shaped sentence naming neither a path nor a probe form, so D1 does not reach it; `grep -n 'git-common-dir' design.md` returns nothing and `grep -c 'syscall.Exec' design.md` returns **0**, so D2 does not either.

### v0.3.0 — plan-audit repair (0.857 against a Tier L threshold of 0.85, narrow-delta FAIL)

The audit verified all four v0.2.0 repairs closed with zero regressions and sampled nine of nine citations as exact; the FAIL rested on three blocking findings, all delta-closable. Three closed, one optional finding taken, two optional findings declined, no requirement added.

1. **D1 — runtime role resolution was owned by no SPEC in the family.** Found independently here and in `SPEC-KANBAN-BOARD-001`, from opposite directions. Four requirements across three SPECs consume a runtime role lookup — `REQ-KS-019` ("declared role"), `REQ-KB-017`, `REQ-KW-007`, `REQ-KW-011` — and each defers to `REQ-KS-004`, which elects the role set and says nothing about occupancy; `REQ-KB-004` records a session identifier, not a role. The one derivable candidate is ruled out by this SPEC's own text: `REQ-KS-014` requires distinct labels while `AC-KS-013` emits two commands per unconfigured worker role and `REQ-KS-005` permits two `run` sessions, so a role maps to two-or-more labels. `REQ-KS-006` is widened in place — addressability **plus** role declaration, distinct from the label, resolvable from a non-`lead` session — with `spec.md` §A.13, §D.11, `AC-KS-030`, `design.md` §B.0, `research.md` §D and §G.8d-f.
2. **D2 — the backend decision rested on an unverified premise, and the premise is half wrong.** §A.8 assumed `moai cc` / `moai glm` are per-session backend selectors and had measured nothing about them. Measured (`research.md` §J): the selector half **holds** — the backend enters the process environment and the launcher execs into `claude`, which inherits it — while each launcher additionally mutates project-global state the other undoes (`team_mode`, tmux session env, `settings.local.json`, `worker-` worktree removal). Consequence stated at §A.8, `REQ-KS-003` extended to re-measure the launcher surface, `plan.md` §C gains checks 8 and 9, two derived hazards recorded (AP-16; `worker-` prefix reserved, `plan.md` §B.5).
3. **D3 — the dispatch cycle contradicted this SPEC's own role table.** §A.4 omitted `review` in the one place the protocol sequence is written, against `REQ-KB-003`'s fixed order. Cycle corrected to `plan → run → review → sync`; `sync → done` stated as the lead's terminal evidence-read write; `backlog → plan` stated as an operator act, with the unowned admission *mechanism* recorded at `plan.md` §B.4 rather than left silent.
4. **D4 (optional, taken) — `REQ-KS-018` asserted `REQ-KB-017` in normative voice.** Recast so the normative force sits on the dispatch this SPEC owns; `AC-KS-018` follows.
5. **Declined, with reasons.** `REQ-KS-024`'s Template-First ordering clause (unobservable after the fact) and `AC-KS-019`'s missing positive control (exhaustiveness legitimately delegated to `AC-KB-017`, which carries one).

- Unresolved `[NEEDS CLARIFICATION]` markers: none. The two gaps this revision surfaced — the operator-admission mechanism and the declaration's carrier — are recorded as an out-of-band note (`plan.md` §B.4) and a deliberate run-phase decision (REQ-KS-006) respectively, neither being a question this SPEC needs answered to be implementable.

### v0.2.0 — plan-audit repair (0.84, narrow-delta FAIL)

Four defects closed, one sibling-consistency change, no requirement renumbered and no v0.1.0 citation altered.

1. **D1 — `AC-KS-009` scanned the wrong access direction.** It looked for `os.Getenv("` while REQ-KS-009's governed surface writes: measured, `internal/cli/factory.go` carries `os.Setenv` at 93/95/110, `os.LookupEnv` at 107, `os.Unsetenv` at 113, and `grep -c 'os.Getenv('` returns **0**. The scan was vacuous, not merely weak. Now enumerates all five access forms with a positive control in each direction (`spec.md` §D.10, `research.md` §H).
2. **D2 — three criteria compared against an unrecorded baseline.** `AC-KS-022`, `AC-KS-023` and `AC-KS-024` demanded equivalence to a "before" nothing obliged anyone to capture. REQ-KS-011's recorded-artifact-plus-provenance shape is generalized in place to REQ-KS-022 / 023 / 024, and M1 grows from one recording to four (`spec.md` §D.9). `AC-KS-020` was named by the audit alongside these and is **not** amended — it asserts three absences with positive controls, not an equivalence.
3. **D3 — REQ-KS-013's impossibility claim had no falsifier.** New `spec.md` §A.12 enumerates the three channels a `lead` backend can arrive through and the refusal each must produce; `AC-KS-029` observes all three. The fourth channel — a hand-launched lead — is recorded as unclosable rather than smoothed over.
4. **D4 — undisclosed compression in REQ-KS-024, seven observations in AC-KS-024.** The absorption of the predecessor's `REQ-KM-037` / `REQ-KM-038` / `REQ-KM-039` is disclosed at `spec.md` §B.8, §E and `research.md` §I; the criterion is split into four (`spec.md` §D.7).
5. **C1 — sole-writer consistency with `SPEC-KANBAN-BOARD-001` v0.2.0.** That sibling restored `REQ-KB-017` (the `lead` is the board's sole writer) after the split deleted it; this SPEC's §C disclaimer was half of what let the loss hide. Write authority is now deferred **by name** at `spec.md` §A.4, §A.11, §C, REQ-KS-004, REQ-KS-018, REQ-KS-019, `plan.md` constraint C8a and AP-12, `design.md` §B.1 / §D.4 / §F, and `research.md` §D. No part of REQ-KB-017 is duplicated.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
