---
id: SPEC-HARNESS-TOKEN-OPT-001
title: "Harness Token/Time Optimization — paths-scoping, SSOT consolidation, and A9 default inversion"
version: "1.0.0"
status: completed
created: "2026-08-11"
updated: "2026-08-11"
author: GOOS
priority: P1
phase: "v3.2 target"
module: "harness-rules"
lifecycle: spec-anchored
tier: M
tags: "token-opt,paths-scope,ssot,template-mirror"
related_specs: []
---

# SPEC-HARNESS-TOKEN-OPT-001 — Harness Token/Time Optimization

## HISTORY

- 2026-08-11 — plan-phase artifacts authored (spec/plan/acceptance/progress). Source: 5-lens parallel audit (6 agents, 108 tool calls, 58 findings) that analyzed the MoAI PLAN>RUN>SYNC workflow for token and time waste; result ~18,400 tokens/turn recoverable. User approved "apply all" scope (P0+P1, 7 recommendations).
- 2026-08-11 — plan-audit iteration 1 FAIL (0.62); defects D1-D7 addressed. D1 `lifecycle: spec-anchored` + `tier: M`. D2 user confirmed "보존 우선 (default-to-preserve)" classification policy via AskUserQuestion; IK classification table appended to plan.md §F.M3. D3/D4 acceptance sentinels repointed to canonical text. D5 IK baseline reconciled to measured 45/12 everywhere. D7 M6 pre-existing byte-drift note added.

## §A. User Story

As a MoAI orchestrator session consuming an ever-growing always-loaded rule corpus, I waste ~18,400 tokens/turn loading doctrine that the current turn does not need (run/sync-phase completion patterns loaded during plan-phase; BAS scanner reference loaded on every session; goal-directive deep-dives loaded when no goal is armed; **45 measured occurrences of one mandate — Implementation Kickoff Approval — duplicated across 12 files**, per `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md | wc -l` run 2026-08-11). I also waste 30-120s of wall-clock per run-phase completion re-executing test/lint/vet/cover commands whose results the manager-develop §E evidence already attributes to the current HEAD.

This SPEC collapses both wastes without weakening any load-bearing invariant. The recovery comes from three structural moves: (1) `paths:` frontmatter restrictions that convert always-loaded files into path-matched lazy files; (2) SSOT consolidation that replaces N duplicate restatements with a 1-line cross-reference; (3) A9 default inversion that makes the attributable diff-check the happy path and re-execution the explicit fallback.

## §B. Scope (WHAT/WHY, not HOW)

### In Scope

- Add `paths:` YAML frontmatter to `verification-batch-pattern.md` and `nav-tokens.md` so they load only on path-match.
- Split `goal-directive.md` into an always-loaded stub (What It Is + Goal-Presentation Timing arm-only invariant + T1-T4 summaries) and a lazy companion `goal-directive-detail.md` with `paths:` scoped to goal state + goal workflow skill.
- Move `session-handoff.md` §Diet Constraints (AP-D-001..005 + 9-item pre-emit checklist) and §V0 Abort Gate Doctrine to the lazy sidecar (`session-handoff-examples.md`, already lazy via `paths:`, or a new `session-handoff-diet.md`).
- Consolidate Implementation Kickoff Approval restatements: designate `orchestration-mode-selection.md` §E as the single SSOT; replace redundant restatements across 9+ files with a 1-line cross-reference.
- Invert the A9 attributable diff-check default in `agent-common-protocol.md` §Parallel Execution: consume §E evidence on three-way match (default), fall back to re-execution on any mismatch.
- Consolidate `CLAUDE.local.md` §18-27 stub tail into a single `## References` section; move §5 (Version Management) and §7 (Hook Development) to `.moai/docs/` or path-scoped rules.
- Mirror every template-track change to `internal/template/templates/` and verify §25 template-neutrality.

### Out of Scope

### Out of Scope — Code path changes

- No change to manager-develop §E E1-E8 self-verification triple (verbatim command + observed output + baseline-attribution) — only the orchestrator-side diff-check DEFAULT that consumes §E is inverted.
- No change to sync-auditor 4-dimension weights (Functionality 40 / Security 25 / Craft 20 / Consistency 15) or Security HARD threshold.
- No change to AskUserQuestion channel monopoly or `ToolSearch(query: "select:AskUserQuestion")` preload mandate.

### Out of Scope — Structural rewrites of safety-critical doctrine

- No rewrite of `verification-claim-integrity.md` §1 (no-unobserved-claim invariant, all four binding surfaces), §2 (baseline-attribution), or §3 (5-section evidence report format). Only §5/§6 Worked Examples MAY move lazy in a follow-up SPEC.
- No rewrite of the session-handoff 6-block Canonical Format + Cut-line Marker Spec (the paste-ready resume contract).
- No rewrite of the Implementation Kickoff Approval gate semantics (mandatory, score-independent, ordering-invariant). The mandate itself is load-bearing; only its duplicate restatements are cut.
- No rewrite of the A9 fallback-to-re-execution contract (any-mismatch → re-execute, never silent skip; mismatch reason logged; VCI §1.1 invariant holds on every path).

### Out of Scope — Audit recommendations not in the P0+P1 set

- P2+ recommendations deferred: cross-file ALWAYS-LOADED inventory sweep, skill preload cap reduction, agent prompt compression pass, hook timeout tuning, manager-develop §E template refactor. Tracked as follow-up SPECs.
- No new SPEC creation tooling, lint engine extensions, or runtime mechanical enforcers for the paths: restrictions (the `paths:` loader already enloads them; this is content-only).

## §C. Functional Requirements (GEARS)

### REQ-HTO-001 — verification-batch-pattern.md paths: restriction + A9 thin pointer

**Where** the orchestrator reaches run/sync-phase completion verification, **When** the session touches `.moai/specs/**` or run/sync workflow skill files, the `verification-batch-pattern.md` rule shall load via `paths:` frontmatter match and shall NOT be always-loaded.

**Additionally**, the rule body shall compress the A9 attributable diff-check section to a thin pointer (~5 lines) that defers to `agent-common-protocol.md` §Parallel Execution → Attributable diff-check doctrinal switch as the SSOT, breaking the circular cross-reference (currently each file restates the other).

### REQ-HTO-002 — nav-tokens.md paths: restriction

**Where** a session touches `.moai/project/*.md`, `.moai/docs/**/*.md`, or Go source files (`**/*.go`), the `nav-tokens.md` rule shall load via `paths:` frontmatter match and shall NOT be always-loaded. BAS scanner surface is niche; most sessions never touch it.

### REQ-HTO-003 — goal-directive.md split

The `goal-directive.md` rule shall be split into:
- an always-loaded stub (~2K tokens) containing: "What It Is" section, the Goal-Presentation Timing arm-only invariant, the Hard Preconditions summary, and T1-T4 trigger one-liners; and
- a lazy companion `goal-directive-detail.md` with `paths:` scoped to goal state files (`.moai/state/goal/**`) and the goal workflow skill tree, containing: the "Comparing Autonomous-Continuation Approaches" table, "Writing an Effective Condition" detail, condition templates, "MoAI Integration Notes", and the "Native `/goal` Prohibition" provenance section.

**When** no goal is armed, the orchestrator session shall pay only the always-loaded stub cost (minority of sessions arm a goal).

### REQ-HTO-004 — session-handoff.md Diet Constraints lazy move

The `session-handoff.md` rule shall keep always-loaded: the 6-block Canonical Format, the Cut-line Marker Specification, the Localization Table (en/ko inline), the 5 Triggers, the Emission-Time Save Obligation, and the Auto-Injected Resume Flow invariants.

The rule shall move to the lazy sidecar (`session-handoff-examples.md` or a new `session-handoff-diet.md`, both `paths:`-scoped to `**/session-handoff.md`): §Diet Constraints full AP-D-001..005 catalogue with the 9-item pre-emit checklist, and §V0 Abort Gate Doctrine full detail. The always-loaded file shall retain a 2-concrete-example inline summary plus a pointer to the sidecar.

### REQ-HTO-005 — Implementation Kickoff Approval SSOT consolidation

The `orchestration-mode-selection.md` rule (§intro line 16 + §E Anti-Patterns) shall be designated as the single SSOT for the Implementation Kickoff Approval mandatory-restoration invariant (mandatory, score-independent, ordering-invariant — gate before any autonomy).

**When** any other rule file, agent definition, or CLAUDE surface restates the Implementation Kickoff Approval mandate, that restatement shall be replaced by a 1-line cross-reference: `Per the Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E).`

**While** the consolidation proceeds, the mandate itself shall remain load-bearing — ONLY redundant restatements are cut, NEVER the mandate. The canonical §E statement and at most one canonical restatement per file (where the file's own logic depends on the gate) MAY be preserved.

### REQ-HTO-006 — A9 attributable diff-check default inversion

**Where** the orchestrator composes the run/sync-phase completion verification batch, **When** the three-way attribution match holds (snapshot key == §E-cited HEAD SHA AND command match AND output match), the orchestrator shall CONSUME the §E evidence and shall NOT re-execute the corresponding verification command.

**When** any of the three attribution axes mismatches, the orchestrator shall fall back to re-execution of the affected dimension, log the mismatch reason (`snapshot_key_drift` / `command_drift` / `missing_section_e` / `output_drift`), and the fallback-to-re-execution contract (any-mismatch → re-execute, never silent skip; VCI §1.1 invariant holds on every path) shall be preserved unchanged.

### REQ-HTO-007 — CLAUDE.local.md consolidation (local-only, no template mirror)

The `CLAUDE.local.md` file shall consolidate §18-27 stub tail ("See: .moai/docs/X.md" pointers, ~11,840 bytes) into a single ~10-line `## References` section, and shall move §5 (Version Management, ~92 lines) and §7 (Hook Development, ~57 lines) to path-scoped rules under `.claude/rules/moai/development/` or to `.moai/docs/`.

**While** `CLAUDE.local.md` is explicitly excluded from templates per CLAUDE.local.md §2 (user-private, runtime-managed), no mirror to `internal/template/templates/` shall occur for this REQ.

### REQ-HTO-008 — Template mirror + make build

**When** a target rule file exists in both `.claude/rules/moai/` and `internal/template/templates/.claude/rules/moai/`, every edit made by REQ-HTO-001 through REQ-HTO-006 shall be mirrored verbatim to the template source.

**After** the mirror, the build shall regenerate embedded files via `make build` (`//go:embed all:templates` in `internal/template/embed.go`), and the regenerated `internal/template/catalog.yaml` shall be committed alongside the template source.

**While** CLAUDE.local.md (REQ-HTO-007) is local-only, it shall NOT be mirrored.

### REQ-HTO-009 — §25 template-neutrality verification

**After** the template mirror, the template source files shall pass the §25 template-neutrality CI guard (`.github/workflows/template-neutrality-check.yaml` + `internal/template/internal_content_leak_test.go`).

The template source shall NOT contain: internal SPEC IDs (e.g., `SPEC-HARNESS-TOKEN-OPT-001`), REQ tokens (e.g., `REQ-HTO-001`), audit citations (e.g., "Audit N Finding AX"), internal work dates, commit SHAs, archive/memory paths, or macOS-bias absolute paths.

### REQ-HTO-010 — do_not_touch preservation (verbatim grep sentinels)

**While** all other REQs execute, the following safety-critical content shall be preserved verbatim in the edited rule files, verified by grep sentinels at run-phase completion:

- `verification-claim-integrity.md` §1 no-unobserved-claim invariant (all four binding surfaces), §2 baseline-attribution, §3 5-section evidence report format.
- AskUserQuestion channel monopoly statement + `ToolSearch(query: "select:AskUserQuestion")` preload mandate, in `askuser-protocol.md` and `moai-constitution.md`.
- Implementation Kickoff Approval gate semantics (mandatory, score-independent, ordering-invariant — gate before any autonomy), in `orchestration-mode-selection.md` §E.
- A9 fallback-to-re-execution contract (any-mismatch → re-execute, never silent skip; mismatch reason logged; VCI §1.1 invariant holds on every path).
- sync-auditor 4-dimension weights (Functionality 40 / Security 25 / Craft 20 / Consistency 15) + Security HARD threshold.
- manager-develop §E E1-E8 self-verification triple (verbatim command + observed output + baseline-attribution).
- session-handoff 6-block Canonical Format + Cut-line Marker Spec.

## §D. Constraints

- **Token recovery target**: ≥ 18,000 tokens/turn recoverable in the always-loaded set, measured by `wc -c` delta on the always-loaded subset of edited files (lazy-moved content counts as full savings when not loaded).
- **Wall-clock recovery target**: 30-120s per run-phase completion attributable to A9 diff-check consume-path (REQ-HTO-006), verified via orchestrator self-timing.
- **Parity invariant**: local `.claude/rules/moai/<file>` byte-content shall be byte-identical to `internal/template/templates/.claude/rules/moai/<file>` for every mirrored file (CI guard: `internal/template/split_namespace_test.go` pattern; manual verify via `diff`).
- **Neutrality invariant**: template source shall contain zero SPEC-ID / REQ-token / SHA / internal-date / audit-citation leakage (CI guard: `.github/workflows/template-neutrality-check.yaml`).
- **No new dependencies**: this SPEC introduces no new Go code, no new external libraries, no new hook scripts. Pure content-only rule edits + one default inversion in an existing rule body.

## §E. Acceptance Criteria Summary

See `acceptance.md` for the canonical Given-When-Then enumeration (AC-HTO-001 through AC-HTO-018, 18 criteria across 7 milestones + template + neutrality + do_not_touch).

## §F. Risks

- **R1 — Goal-directive split reachability**: the lazy companion must be reachable when a goal is armed. Mitigation: `paths:` glob includes `.moai/state/goal/**` (written by SessionStart + `moai goal arm`) AND the goal workflow skill tree — at least one of which is touched whenever a goal context arises.
- **R2 — IK SSOT consolidation over-cut**: a restatement that LOOKS redundant may be load-bearing for a file's internal logic. Mitigation: M3 enumerates each of the ~53 occurrences, classifies as "canonical §E" / "load-bearing local restatement" / "redundant restatement", and only the third class is cut. plan-auditor reviews the classification before M3 merges.
- **R3 — A9 default inversion silently weakens VCI §1.1**: if the diff-check predicates on a stale snapshot key, the consume-path could mark a dimension PASS without fresh evidence. Mitigation: REQ-HTO-006 preserves the fallback contract unchanged (any-mismatch → re-execute); the three-way match is strict-AND; the mismatch reason is logged.
- **R4 — Template-neutrality slip**: a SPEC-ID or REQ-token leaking into the template source during mirror. Mitigation: REQ-HTO-009 CI guard + pre-PR self-check (`.moai/docs/template-internal-isolation-doctrine.md` §25.3 5-item checklist).
- **R5 — CLAUDE.local.md §5/§7 move breaks a §2 Template-First expectation**: §5 and §7 are referenced by other sections of CLAUDE.local.md. Mitigation: §5 and §7 destinations are `.moai/docs/version-management.md` and `.moai/docs/hook-development.md` (or path-scoped rules); CLAUDE.local.md retains a 1-line pointer under the new `## References` section.

## §G. Dependencies

- No upstream SPEC dependency. This SPEC is self-contained content + default-inversion work.
- Downstream: the audit's P2+ recommendations (deferred) become follow-up SPECs.

## §H. Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution → Attributable diff-check doctrinal switch (the A9 site REQ-HTO-006 edits).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (the IK SSOT REQ-HTO-005 designates).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields (frontmatter schema REQs validated against).
- `.moai/docs/template-internal-isolation-doctrine.md` §25 + §25.1 + §25.3 (the neutrality catalogue REQ-HTO-009 enforces).
- CLAUDE.local.md §2 [HARD] Template-First Rule (the mirror obligation REQ-HTO-008 satisfies).
