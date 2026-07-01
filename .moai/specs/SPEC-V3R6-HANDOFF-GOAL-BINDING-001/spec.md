---
id: SPEC-V3R6-HANDOFF-GOAL-BINDING-001
title: "Bind Claude Code /goal into session-handoff resume Block 1"
version: "0.1.0"
status: completed
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "session-handoff, goal-directive, resume-message, paste-ready, documentation"
tier: M
---

# SPEC-V3R6-HANDOFF-GOAL-BINDING-001 — Bind `/goal` into session-handoff resume Block 1

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-01 | manager-spec | Initial plan-phase draft — bind Claude Code native `/goal` into the paste-ready resume Block 1 as a purpose-conditional re-set line, mirroring the existing `/effort ultracode` mechanism. |

## §A. Context and Motivation

Claude Code's native `/goal <condition>` command sets a session-scoped **completion condition** that keeps Claude working across turns until a fast evaluator confirms the condition holds (`https://code.claude.com/docs/en/goal`). Two verified ground-truth properties make it structurally analogous to the existing `/effort ultracode` handoff concern:

1. `/clear` removes an active goal. A goal is restored on `--resume`/`--continue` (baselines reset), but a goal is **NOT** restored by the `ultrathink.` opener of a paste-ready resume message — it must be explicitly re-issued after `/clear`. This is the *exact* discipline already documented for `/effort ultracode` (`.claude/rules/moai/workflow/dynamic-workflows.md` + `session-handoff.md` Block 1 Field-by-Field Specification).
2. The evaluator judges the condition against what Claude has surfaced in the conversation; it does not run commands or read files. Therefore a `/goal` line is only useful in the next session when the next SPEC has a **mechanically verifiable** end-state the evaluator can judge (e.g. `go test ./... exits 0 AND golangci-lint clean, or stop after N turns`).

Today the paste-ready resume message re-sets `/effort ultracode` conditionally but has no parallel affordance to re-arm a `/goal` autonomous-continuation loop after `/clear`. A session resumed for a run-phase SPEC with a clear machine-verifiable completion condition silently loses the `/goal` loop — the resumed session reverts to per-turn STOP prompts even though the prior session had established an autonomous-continuation condition.

This SPEC binds `/goal` into the resume-message structure using the identical conditional-emit mechanism already proven for `/effort ultracode`. Scope is deliberately narrow: a single conditional line in Block 1, emitted only for run-phase SPECs with verifiable completion conditions.

## §B. Requirements (GEARS)

Notation: GEARS (Generalized EARS). Subject `<subject>` generalized per the canonical GEARS guide. `shall`/`shall not` are normative.

### REQ-HGB-001 — Conditional `/goal` re-set line in Block 1 (Ubiquitous)

The session-handoff paste-ready resume message Block 1 **shall** carry a purpose-conditional `/goal <completion-condition>` re-set line, authored with the same conditional-emit mechanism as the existing `/effort ultracode` re-set line, positioned within Block 1 alongside the `ultrathink.` opener and the `/effort ultracode` line.

### REQ-HGB-002 — Run-phase AND verifiable emit trigger (Compound: Where + When)

**Where** the next SPEC's phase is run-phase **When** the next SPEC declares a mechanically verifiable completion condition (a machine-checkable end-state such as `go test ./... exits 0`, a lint-clean state, or a bounded `stop after N turns` clause), the resume-message renderer **shall** emit the `/goal <completion-condition>` line in Block 1.

The renderer **shall not** emit the `/goal` line for a plan-phase or sync-phase next SPEC, nor for any next SPEC lacking a machine-verifiable end-state. The default on ambiguity **shall** be to omit the line — the identical default already binding the `/effort ultracode` line.

### REQ-HGB-003 — Diet single-line constraint (Ubiquitous)

The `/goal` re-set line **shall** be a single conditional line, mirroring the single `/effort ultracode` comment line, and **shall not** expand into a multi-line block. The completion-condition text **shall** remain on one line; detailed acceptance-criteria bindings **shall** reside in the SPEC's `acceptance.md`, not inline in the paste-ready message. This binds the `/goal` line to the session-handoff Diet Constraints doctrine ("next-session minimum executable context").

### REQ-HGB-004 — Implementation Kickoff Approval not bypassed (Unwanted behavior)

The presence of a `/goal` line in a resume message **shall not** authorize autonomous run-phase entry. The Implementation Kickoff Approval human gate (orchestrator `AskUserQuestion` per `CLAUDE.local.md §19.1` and `.claude/rules/moai/workflow/goal-directive.md`) **shall** remain required before run-phase entry, independent of whether a `/goal` line is present in the resume. The binding text **shall** state this invariant explicitly so the `/goal` line cannot be read as pre-authorizing run-phase.

### REQ-HGB-005 — SSOT↔render-surface parity (Ubiquitous)

The `/goal` binding **shall** be present identically in both the SSOT (`.claude/rules/moai/workflow/session-handoff.md`) and the render surface (`.claude/output-styles/moai/moai.md §8`), preserving the existing bidirectional drift-mitigation sentinel. Adding the binding **shall not** introduce a new pre-emit self-check concern-name qualifier beyond the existing three (`paste-ready budget`, `localization render`, `session-handoff template completeness`); the `/goal` self-check item **shall** fit within the existing qualifiers.

### REQ-HGB-006 — Localization parity (Ubiquitous)

The `/goal` line **shall** be treated as a command literal preserved verbatim across all locales (en/ko/ja/zh), the same as the `ultrathink.` and `/effort ultracode` literals. Any surrounding header or comment label **shall** follow the existing localization contract, and the Localization Table locale-column count (4 columns: en/ko/ja/zh) **shall** remain intact in both the SSOT and the render surface.

### REQ-HGB-007 — Omission anti-pattern (Ubiquitous)

The session-handoff Anti-Patterns list **shall** include an anti-pattern describing the omission of the `/goal` line when the next SPEC has a verifiable run-phase completion condition, mirroring the existing `/effort ultracode` omission anti-pattern (the resumed session silently loses autonomous-continuation because `/goal` is NOT restored by `ultrathink.`).

### REQ-HGB-008 — goal-directive.md cross-reference (Ubiquitous)

The `.claude/rules/moai/workflow/goal-directive.md` MoAI Integration Notes `ultrathink.` resume-pairing bullet **shall** be strengthened to explicitly state the Block 1 conditional `/goal` re-set binding and **shall** cross-reference `session-handoff.md`.

### REQ-HGB-009 — Template-First mirror parity (Capability gate)

**Where** an edited rule/style file has a template mirror under `internal/template/templates/`, the identical `/goal` binding **shall** be applied to the template copy, and `make build` **shall** re-embed the templates. All four candidate target files are verified MIRRORED (both live and template copies present), so the binding **shall** be mirrored for every file this SPEC edits.

The template Example copy of `session-handoff.md` **shall** use a language-neutral completion-condition phrasing (e.g. "the SPEC's test suite passes AND lint is clean, or stop after N turns") and **shall not** introduce a language-specific token (such as `go test`, `pytest`, or `cargo`), preserving that template file's currently-clean language-neutrality baseline per `CLAUDE.local.md §15/§25`. The illustrative `go test` phrasing remains permitted only on non-template SPEC surfaces (this spec.md, acceptance.md) and on the `moai.md` / `goal-directive.md` siblings that already carry it.

### REQ-HGB-010 — context-window-management.md non-contradiction (State-driven)

**While** editing the resume-message binding, the resume-message format references in `.claude/rules/moai/workflow/context-window-management.md` **shall** remain consistent with the new `/goal` binding. That file references the resume format at the reference level; the run phase **shall** confirm no contradiction is introduced and **shall** edit it only if a contradiction is found.

## §C. Exclusions

This SPEC is deliberately narrow. The following are out of scope.

### Out of Scope — `/loop` and scheduled-task patterns

- Adding `/loop`, `/moai loop`, or scheduled-task (`https://code.claude.com/docs/en/scheduled-tasks`) patterns to the resume message. The user explicitly deferred `/loop`; only the `/goal` conditional line is in scope. The `/loop` vs `/goal` distinction is background context only.

### Out of Scope — runtime `/goal` mechanics and evaluator changes

- Any change to how Claude Code's native `/goal` command, its evaluator model, or its condition-judging behavior works. This SPEC only binds a *reference to* `/goal` into a documentation template; it does not implement, wrap, or alter the native command.
- Any Go code (`internal/`, `pkg/`, `cmd/`) change. This is a documentation-rule alignment SPEC; the only build interaction is `make build` re-embedding the edited template files (no source-code edit).

### Out of Scope — autonomous run-phase authorization

- Weakening, bypassing, or auto-satisfying the Implementation Kickoff Approval human gate. REQ-HGB-004 forbids this; the `/goal` line is a continuation-loop convenience, never a run-phase pre-authorization. Any interpretation that a `/goal` line authorizes autonomous `/moai run` entry is explicitly excluded.

### Out of Scope — new locales or Localization Table restructuring

- Adding locales beyond the existing four (en/ko/ja/zh) or restructuring the Localization Table. REQ-HGB-006 requires preserving the existing 4-column parity; expanding the locale set is a separate concern.

### Out of Scope — mechanical drift-lint enforcement

- Adding a Go lint rule or hook to mechanically enforce SSOT↔render parity for the `/goal` binding. The existing drift-mitigation is a documentation sentinel (mitigation + visibility, not mechanical prevention); a mechanical enforcement layer is deferred per the existing session-handoff SSOT-align follow-up and is not introduced here.

## §D. Cross-References

- `.claude/rules/moai/workflow/session-handoff.md` — SSOT for the 6-block resume format, Block 1 Field-by-Field Spec, Diet Constraints, Anti-Patterns.
- `.claude/output-styles/moai/moai.md §8` — render surface for the 6-block skeleton + pre-emit self-check.
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` autonomous-continuation guidance; `ultrathink.` resume-pairing bullet.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the `/effort ultracode` NOT-restored-by-`ultrathink.` precedent this SPEC mirrors.
- `.claude/rules/moai/workflow/context-window-management.md` — resume-message format reference (non-contradiction target).
- `CLAUDE.local.md §19.1` — Implementation Kickoff Approval human-gate invariant (REQ-HGB-004).
- Official docs (verified this session): `https://code.claude.com/docs/en/goal`, `https://code.claude.com/docs/en/scheduled-tasks`.
