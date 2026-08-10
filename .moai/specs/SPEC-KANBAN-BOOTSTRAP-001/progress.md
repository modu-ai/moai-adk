---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Progress — Kanban session topology, bootstrap, and dispatch"
version: "0.3.0"
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
- **The requirement cap bound at v0.3.0 and held.** The role-resolution repair needed a runtime contract that no SPEC in the family owned; with 25 of 25 requirements used, it was authored as a widening of `REQ-KS-006` rather than as a twenty-sixth requirement. Where a widening had not been available the finding would have been reported to the orchestrator instead.
- Split from the superseded `SPEC-KANBAN-MULTISESSION-001` (59 requirements, plan-audit FAIL 0.87), alongside `SPEC-KANBAN-BOARD-001` and `SPEC-KANBAN-WORKTREE-001`.
- Three corrections landed with the split and each is argued in `spec.md`: the dead `TestNew_NoAskUserQuestion` guard citation (§A.6), the missing baseline recording window (§A.7), and the rejected `column:` frontmatter field (§A.11).
- Unresolved `[NEEDS CLARIFICATION]` markers: none.

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
