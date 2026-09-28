---
id: SPEC-MANAGER-TODO-001
title: "Rename and repurpose mission-governor into manager-todo — todo-queue management, Jev decision wiring, dispatch ownership, and /moai:todo --auto serial mode"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli/todo.go, .claude/agents/moai/, internal/template/templates/.claude/agents/moai/"
lifecycle: spec-anchored
tags: "agent-catalog, todo, kanban, jev, gtd, codex-roles, docs-i18n"
tier: L
---

# SPEC-MANAGER-TODO-001 — manager-todo (mission-governor rename/repurpose + /moai:todo --auto)

## HISTORY

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase draft. Operator directive 2026-09-29: rename and repurpose the `mission-governor` agent into `manager-todo`; add `/moai:todo --auto` serial processing; codify the Jev decision boundary; absorb the GTD auto-mission judgment as a sub-role; full live-reference refresh. | manager-spec |

## A. Summary

The read-only judgment agent `mission-governor` (`.claude/agents/moai/mission-governor.md`, frontmatter `tools: Read, Grep, Glob, Skill`, `permissionMode: plan`, `memory: project`) is renamed to `manager-todo` and repurposed into the dedicated todo-queue management agent: todo-queue management + Jev decision wiring + card dispatch/management ownership + the new `/moai:todo --auto` serial processing mode. The read-only sealed-snapshot judgment contract of the former mission-governor continues as a sub-role of manager-todo; no capability is orphaned.

The retained-agent ceiling is unaffected: this SPEC renames one MoAI-custom agent in place. The catalog stays exactly **13 retained agents (12 MoAI-custom + 1 Anthropic built-in `Explore`)** — no net addition, no ceiling revision.

## B. Requirements (GEARS)

Requirement domain prefixes: REN (rename/repurpose), REF (reference sweep refresh), AUTO (`/moai:todo --auto`), JEV (Jev boundary), GTD (auto-mission absorption), DOC (docs/locales).

### B.1 Agent rename/repurpose (M1)

- **REQ-MT-001** (Ubiquitous): The agent definition `manager-todo` shall exist at `.claude/agents/moai/manager-todo.md` (C1, hand-edited) and `internal/template/templates/.claude/agents/moai/manager-todo.md` (C2, hand-edited), carrying the todo-queue management mission, the Jev decision boundary (§B.4), the card dispatch/management ownership scope, and the read-only sealed-snapshot judgment sub-role; its frontmatter shall satisfy the agent-authoring format rules (`tools:` CSV string, no `model:`/`effort:` fields, valid `permissionMode` enum value).

- **REQ-MT-002** (Event): When `make agents-emit` runs after REQ-MT-001's C2 edit, the emitter shall produce `internal/template/templates/.codex/agents/moai/manager-todo.toml` from C2, `make agents-emit-check` shall pass with zero drift, and no `*.toml` file under `internal/template/templates/.codex/agents/moai/` shall be hand-edited.

- **REQ-MT-003** (Event): When the rename lands, all three mission-governor copies shall be removed — C1 `.claude/agents/moai/mission-governor.md` and C2 `internal/template/templates/.claude/agents/moai/mission-governor.md` by deletion, C3 `internal/template/templates/.codex/agents/moai/mission-governor.toml` by emitter regeneration — and `git grep -i mission.governor` over tracked files shall return only hits whose file appears in the research.md §B baseline with an explicit frozen/historical disposition.

- **REQ-MT-004** (Ubiquitous): The retained-agent catalog enumeration shall remain 13 retained (12 MoAI-custom + built-in `Explore`) and shall name `manager-todo` in place of `mission-governor` in every catalog enumeration surface: CLAUDE.md §4 (and its template mirror), `agent-authoring.md` § Agent Categories (± mirror), `agent-patterns.md` § static-agent enumerations (± mirror), `internal/harness/rosterguard`, `internal/harness/delegationmap` `retainedCatalog`, `internal/template/retained_agents.go`, and `internal/template/catalog.yaml`.

### B.2 Reference sweep refresh (M2)

- **REQ-MT-005** (Ubiquitous): Every live tracked-file reference to `mission-governor` recorded in research.md §B baseline (262 hits, file:line enumerated) shall be either updated to `manager-todo` (or to manager-todo-equivalent prose in the file's locale) or explicitly dispositioned in the baseline table as frozen (historical testdata fixtures) or historical (changelog/report prose recording past events); a post-change run of `git grep -in mission.governor -- . ':(exclude).moai/reports' ':(exclude).moai/specs'` shall return only hits whose file:line sits in the baseline's explicitly-frozen/historical disposition list, and zero hits anywhere else.

- **REQ-MT-006** (Event): When the codex_audit role registry is refreshed, `internal/cli/codex_audit_mcp.go` (tool descriptions), `internal/cli/codex_role_fingerprint.go`, `internal/harness/rosterguard`, and `internal/template/agentemit/agents-codex.yaml` shall name the post-rename role set with no stale `mission-governor` entry outside the explicitly-frozen disposition list, and the disposition chosen for the read-only role roster (design decision D-3) shall be applied consistently across all four surfaces.

### B.3 `/moai:todo --auto` serial mode (M3)

- **REQ-MT-007** (Event): When `/moai:todo --auto` is invoked and at least one pickup target exists (per REQ-MT-008), the command shall process cards strictly serially — exactly one card at a time, in the cycle accept → process to completion → accept next — and shall never process two cards concurrently.

- **REQ-MT-008** (Event): When a `--auto` cycle selects its next card, the command shall select in this order: (a) unfinished already-picked cards whose owning session is measured dead/ended — ownership judged by the session registry (entry absent, or present with a dead/stale PID confirmed by process probe) AND an `lsof` cwd measurement showing no live process whose working directory sits inside the owning session's tree — then (b) the next unpicked card in queue order.

- **REQ-MT-009** (Ubiquitous — prohibition): A picked card whose owning session is measured alive (registry liveness or an `lsof` cwd hit) shall never be taken over by `--auto`, absolutely, including when no other pickup target exists and the queue is otherwise empty.

- **REQ-MT-010** (Ubiquitous): The pickup predicate shall be expressed only in the current card-state vocabulary — pickup targets are cards in state `queued`, plus dead-owner `picked` cards — never as a negation of a terminal state, so that a future card state (for example a `hold` state) is excluded from pickup automatically without revision of this SPEC.

- **REQ-MT-011** (Event): When a card reaches completion inside a `--auto` run, the command shall emit the /clear guidance — the instruction directing the operator to clear the session, naming the completed card and the next step — before the cycle accepts the next card; this emission is unconditional for every completed card.

- **REQ-MT-012** (Event): Where `/moai:todo --auto` is invoked, the invocation itself shall constitute the operator's batch approval for serially consuming the queue in order, and the command shall reconcile with the "promotion is the operator's act" doctrine by deriving promotion authority solely from that operator-issued invocation — it shall not self-promote, reorder, admit, or drop cards beyond the queue order the operator approved.

- **REQ-MT-013** (Ubiquitous — prohibition): `--auto` shall create no factory lease, claim no factory slot, and take no dependency on unlanded factory-dispatch code; its integration with the parallel factory work is a serial-merge-window seam only (research.md §D), never a build or run dependency.

### B.4 Jev decision boundary (M4)

- **REQ-MT-014** (Ubiquitous): Where the Jev local scripts (`scripts/jev/`) are available, `manager-todo` and the `--auto` cycle may consult Jev (`triage.sh`, `route.sh`) for dispatch order and priority judgment as a display-only signal, and shall treat Jev output as judgment input for the lead — never as authority.

- **REQ-MT-015** (Ubiquitous — prohibition): Jev output shall never be the basis of a queue mutation, a completion verdict, a merge approval, or any third-grade decision (operator/lead owned); a Jev absent-key or absent-network degradation shall exit the consultation as a labelled non-finding and the cycle shall proceed on lead judgment alone.

### B.5 GTD auto-mission absorption (M4)

- **REQ-MT-016** (Event): When the GTD auto-mission workflow dispatches its judgment role, it shall dispatch `manager-todo` carrying the read-only sealed-snapshot judgment as a sub-role, the goal workflow text (`.claude/skills/moai/workflows/goal.md` and its template mirror) shall name `manager-todo` for that role, and `moai goal run --governor-receipt` shall continue to accept a decision receipt of the same schema so no receipt-consuming code path breaks.

- **REQ-MT-017** (Event): When the codex_audit role registry is updated for the rename, no GTD-judgment capability shall be orphaned: the disposition of the read-only role roster entry (design decision D-3) shall be applied, and any GTD-judgment capability that the disposition removes from the Codex path shall be recorded in research.md §C as a reported contract conflict for the lead rather than silently dropped.

### B.6 Docs and locales (M5)

- **REQ-MT-018** (Event): When any reference refresh touches `docs-site/content/**` or the README set, the change shall land in all 4 locales (en/ko/ja/zh) with file-existence and section parity, and every edit under `internal/template/templates/**` shall pass the template-neutrality guard (`.github/workflows/template-neutrality-check.yaml` C1-C8 classes) and the sibling content-leak tests.

## C. Constraints

- C-1: The rename is in-place — the retained-agent ceiling (12 MoAI-custom + 1 built-in) is unchanged and no catalog-count surface may drift from 13/12.
- C-2: C3 (`internal/template/templates/.codex/agents/moai/*.toml`) is machine-emitted only; hand-editing it is prohibited at every milestone.
- C-3: All template-mirrored content obeys the 16-programming-language neutrality contract and the template internal-content isolation doctrine (no card ids of other cards, no internal dates, no commit SHAs, no macOS-bias paths, no CLAUDE.local references in `internal/template/templates/**`).
- C-4: No dependency on unlanded code from the parallel factory card; the todo CLI delta overlap is absorbed through the serial merge window at run time.
- C-5: The live todo queue store is the home-directory SQLite db (`~/.moai/db/<project-key>/todo/backlog.db` per `internal/cli/todo.go`); no SPEC prose may reintroduce the stale `.moai/state/kanban/backlog.json` claim.
- C-6: Plan-phase only — this SPEC touches no production code; run-phase owns all implementation.

## D. Acceptance Criteria

All acceptance criteria are enumerated, with Given-When-Then scenarios and mechanical check commands, in `acceptance.md` (AC-MT-001 … AC-MT-022). Every REQ above maps to at least one AC; the mapping table lives in `acceptance.md` §D.

## Out of Scope

### Out of Scope — factory dispatch integration

- Creating, acquiring, or consuming a factory lease or slot from the `--auto` cycle (follow-up-card candidate only, after the parallel factory card lands).
- Any change to the factory dispatch code path owned by the parallel factory SPEC.

### Out of Scope — new card states

- Introducing a `hold` or any new card state; REQ-MT-010 only requires the pickup predicate to be forward-compatible with one.

### Out of Scope — queue store migration

- Any change to the queue store location, schema, or the SQLite runtime store behavior.

### Out of Scope — Jev productization

- Wiring `scripts/jev/` into the shipped product, templates, or any user project; Jev remains local-only, and this SPEC codifies a boundary, not a dependency.

### Out of Scope — historical artifact rewrites

- Editing `.moai/reports/**`, other SPECs' artifacts, `CHANGELOG.md` history rows, frozen testdata (`internal/cli/testdata/codex-rollouts-t1171/**`), or codemaps content; those are dispositioned frozen/historical in the research.md §B baseline, and codemaps regenerate at the next `/moai codemaps` run.
