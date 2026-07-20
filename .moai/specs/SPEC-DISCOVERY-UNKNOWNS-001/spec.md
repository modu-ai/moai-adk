---
id: SPEC-DISCOVERY-UNKNOWNS-001
title: "Unknowns framework Tier-1 for Context-First Discovery — Blind Spot Pass + decision-reversibility ordering + 4-quadrant lens"
version: "0.1.0"
status: completed
created: 2026-07-05
updated: 2026-07-13
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tier: M
tags: "discovery, unknowns, blind-spot, gears, askuser, planning, context-first, doc"
---

# SPEC-DISCOVERY-UNKNOWNS-001 — Unknowns framework Tier-1 for Context-First Discovery

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-05 | 0.1.0 | Initial plan-phase draft. Tier M. Applies the Tier-1 subset of the "A Field Guide to Fable: Finding Your Unknowns" framework (Thariq/@trq212) to moai-adk's Context-First Discovery (CLAUDE.md §7 Rule 5). Three genuinely-absent enhancements only (T1 Blind Spot Pass, T2 decision-reversibility plan ordering, T3 unknowns 4-quadrant lens); interview + reference discovery are already covered by AskUserQuestion Socratic rounds + Explore/Context7 and are explicitly NOT re-added. All three gaps grep-verified against the current tree (§A.2). Doc/rule/agent-body level only; no Go code, no new agents/skills/subsystems. | manager-spec |

## §A. Context and Intent

### §A.1 Why this enhancement

The article's thesis: the **map** (the prompts and context you hand an agent) is not the **territory** (the real codebase and its constraints); the gap between them is **unknowns**, and the core skill of agentic coding is reducing and planning for unknowns. This is the same root problem moai-adk's **Context-First Discovery** already addresses (CLAUDE.md §7 Rule 5 + `.claude/rules/moai/core/askuser-protocol.md` § Ambiguity Triggers and Exceptions). moai-adk supports Fable 5 (`claude-fable-5`) and targets long-horizon autonomous work, so the article's taxonomy lens and phase-specific unknown-discovery techniques are directly relevant.

The article contributes three items moai-adk's Discovery system currently lacks. This SPEC adapts exactly those three — **adaptation, not transplant**: the article assumes an interactive IDE pairing loop, whereas moai-adk is an orchestrator whose user-facing channel is `AskUserQuestion` and whose read-only exploration is `Agent(Explore)`. Each enhancement is re-expressed in the orchestrator model.

### §A.2 Verified gaps (the problem baseline)

All three gaps were grep-verified against the current working tree before authoring:

1. **No Blind Spot Pass technique.** `grep -rniE 'blind ?spot|unknown[ -]unknown' .claude/rules .claude/agents` returns exactly one incidental match — `.claude/agents/moai/manager-git.md:79` `"(only when known test blind spots)"`, which is PR test-scenario phrasing, NOT a Discovery technique. There is no pre-plan "help me find my unknown-unknowns" capability anywhere in the rule/agent surface.
2. **No unknowns taxonomy in Discovery.** `grep -niE 'unknown|blind' .claude/rules/moai/core/askuser-protocol.md` returns zero matches. The SSOT for Context-First Discovery frames ambiguity by *detection signal* (pronouns, multi-interpretable verbs, unclear boundaries, state conflict) — never by *user blind spot*.
3. **Plans ordered by execution, not by decision-reversibility.** `grep -rniE 'reversib|most likely to change|change[ -]likelihood|decision.*first' .claude/agents/moai/manager-spec.md` returns zero matches. Neither `manager-spec.md` nor the plan.md section skeleton leads with the decisions most likely to change. The article's insight: lead with data-model / type-interface / UX-flow decisions (highest change-likelihood) and bury mechanical refactoring at the bottom, so human review focuses on what actually matters.

### §A.3 Adaptation contract (what is deliberately NOT re-added)

The article names four unknown-reduction activities: **Interviews**, **References**, **Blind Spot Pass**, and **plan ordering by reversibility**. Two of the four already exist in moai-adk and MUST NOT be re-added:

| Article activity | moai-adk existing mechanism | This SPEC |
|------------------|-----------------------------|-----------|
| Interviews (ask the human to fill gaps) | AskUserQuestion Socratic interview rounds (`askuser-protocol.md` § Socratic Interview Structure) | NOT re-added |
| References (pull external/library docs) | `Agent(Explore)` + Context7 MCP + WebFetch (JIT docs) | NOT re-added |
| Blind Spot Pass (surface unknown-unknowns) | **absent** (§A.2 gap 1) | **T1 — added** |
| Plan ordering by reversibility | **absent** (§A.2 gap 3) | **T2 — added** |
| Unknowns taxonomy lens | **absent** (§A.2 gap 2) | **T3 — added** |

### §A.4 Affected surfaces (Template-First paired edits)

Every surface below is template-mirrored under `internal/template/templates/`. Per CLAUDE.local.md §2 (Template-First), each content edit is applied to BOTH the local file AND its mirror, followed by `make build`. Baseline parity checked at plan-time: `CLAUDE.md` and `spec-workflow.md` are local==template identical; `askuser-protocol.md` diverges by ~1 hunk and `manager-spec.md` by ~5 hunks / 10 lines (all §25-neutrality strips of internal dates / SPEC-IDs / REQ tokens — verified via live diff). Neither divergence touches this SPEC's T1/T2 edit anchors (the manager-spec `Step 4` / `**plan.md**` T2 anchor is confirmed untouched), so the paired edits target identical anchors in each surface.

| # | Surface (local) | Template mirror | Enhancement(s) | Parity regime (CI-enforced) |
|---|-----------------|-----------------|----------------|------------------------------|
| 1 | `.claude/rules/moai/core/askuser-protocol.md` | `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` | T1 (new Blind Spot Pass subsection) + T3 (4-quadrant lens in § Ambiguity Triggers) | **sanitized-pair** — `TestSanitizedPairParity` (`sanitizedPairPaths`, sentinel `SANITIZED_PAIR_PARITY_DRIFT`); T1+T3 doctrine MUST propagate to BOTH trees |
| 2 | `.claude/agents/moai/manager-spec.md` | `internal/template/templates/.claude/agents/moai/manager-spec.md` | T2 (plan.md decision-reversibility ordering guidance) | sanitized mirror, **NOT in any parity registry** — grep-token parity only (AC-DU-012) + `TestTemplateNoInternalContentLeak` cleanliness |
| 3 | `CLAUDE.md` (§7 Rule 1 + §7 Rule 5) | `internal/template/templates/CLAUDE.md` | T1 wire (Rule 5) + T2 (Rule 1) + T3 one-line lens framing (Rule 5) | **none** — grep-token parity only (AC-DU-012) |
| 4 | `.claude/rules/moai/workflow/spec-workflow.md` (Plan Phase) | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | T1 wire (plan-phase entry reference) | **byte-identical** — `TestRuleTemplateMirrorDrift` (`workflowOptMirroredPaths`, sentinel `RULE_TEMPLATE_MIRROR_DRIFT`); local↔mirror MUST be byte-for-byte identical |

> **Parity-regime note (audit D2)**: surfaces 1 (sanitized-pair) and 4 (byte-identical) are enrolled in ENFORCED CI parity tests that this SPEC's `make build` + grep-token checks do NOT run. A whitespace-only mirror drift (surface 4) or a partial/absent doctrine propagation (surface 1) passes the grep-token / `make build` ACs but REDS main on the Route A push. AC-DU-017 runs `TestRuleTemplateMirrorDrift` + `TestSanitizedPairParity` after the mirror edits to catch exactly this. Surfaces 2-3 have no enforced parity test — their mirror parity is grep-token only (AC-DU-012).

## §B. Goals (in-scope) — exactly three Tier-1 enhancements

- **T1. Blind Spot Pass** — a new, OPTIONAL pre-plan Discovery technique defined in `askuser-protocol.md`, wired by a one-line reference from `spec-workflow.md` Plan Phase and CLAUDE.md §7 Rule 5.
- **T2. Plan ordering by decision-reversibility** — plan.md authoring guidance (manager-spec.md) and Approach-First reports (CLAUDE.md §7 Rule 1) lead with the decisions most likely to change; mechanical steps deferred to the bottom. Section/milestone-ordering rule only.
- **T3. Unknowns 4-quadrant lens** — the Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns taxonomy added as an explicit lens in `askuser-protocol.md` § Ambiguity Triggers, with a single one-line framing + SSOT pointer in CLAUDE.md §7 Rule 5.

## §C. Requirements (GEARS)

### §C.1 T1 — Blind Spot Pass

- **REQ-DU-001** (Ubiquitous): The `askuser-protocol` rule **shall** define a named "Blind Spot Pass" Discovery technique as its own subsection.
- **REQ-DU-002** (Capability gate): **Where** the user is working in an unfamiliar domain (a new subsystem, unfamiliar design/library work) and unknown-unknowns are suspected, the orchestrator **shall** run a Blind Spot Pass (at orchestrator discretion — the suspicion-trigger is a judgment call, not an automatic gate; see REQ-DU-004 for the optionality invariant) before plan-phase entry. The "shall run" obligation is conditioned on the orchestrator's suspicion of unknown-unknowns; it does NOT make the pass a mandatory gate on every unfamiliar-domain plan entry.
- **REQ-DU-003** (Event-driven): **When** the orchestrator runs a Blind Spot Pass, the orchestrator **shall** spawn `Agent(Explore)` in read-only mode to scan the relevant domain and then surface the user's likely unknown-unknowns through an `AskUserQuestion` round so the user can react and prompt better.
- **REQ-DU-004** (Ubiquitous): The Blind Spot Pass **shall** be optional — triggered only when unknown-unknowns are suspected — and **shall not** be a mandatory gate.
- **REQ-DU-005** (Unwanted behavior): The Blind Spot Pass **shall not** have `Agent(Explore)` or any subagent prompt the user directly; unknown-unknowns **shall** be surfaced only through the orchestrator's `AskUserQuestion` channel (preserving the asymmetric orchestrator–subagent boundary).
- **REQ-DU-006** (Ubiquitous): The `spec-workflow` Plan Phase section and CLAUDE.md §7 Rule 5 **shall** each carry a one-line reference to the Blind Spot Pass technique pointing to the `askuser-protocol` SSOT.

### §C.2 T2 — Plan ordering by decision-reversibility

- **REQ-DU-007** (Ubiquitous): The `manager-spec` plan.md authoring guidance **shall** instruct that plan.md milestones/sections lead with the decisions most likely to change (data-model changes, new type interfaces, user-facing/UX flows) and defer mechanical/refactoring steps to the bottom.
- **REQ-DU-008** (Ubiquitous): CLAUDE.md §7 Rule 1 (Approach-First Development) **shall** instruct that Approach-First reports present the highest change-likelihood decisions first.
- **REQ-DU-009** (Unwanted behavior): The decision-reversibility ordering **shall** be a section/milestone-ordering rule only and **shall not** introduce new tooling, machinery, agents, or schema.

### §C.3 T3 — Unknowns 4-quadrant lens

- **REQ-DU-010** (Ubiquitous): The `askuser-protocol` § Ambiguity Triggers section **shall** present the Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns taxonomy as an explicit lens.
- **REQ-DU-011** (Event-driven): **When** the 4-quadrant lens identifies suspected Unknown-Unknowns, the lens **shall** direct the orchestrator to run a Blind Spot Pass (per REQ-DU-002).
- **REQ-DU-012** (Ubiquitous): CLAUDE.md §7 Rule 5 **shall** carry a single one-line framing of the 4-quadrant lens plus an SSOT pointer to `askuser-protocol`, and **shall not** expand CLAUDE.md §7 Rule 5 beyond that framing + pointer (CLAUDE.md is under active diet).

### §C.4 Cross-cutting constraints

- **REQ-DU-013** (Unwanted behavior): The SPEC **shall not** introduce Go source changes, new agents, new skills, or new subsystems; all changes **shall** be confined to documentation, rule, and agent-body markdown.
- **REQ-DU-014** (Capability gate): **Where** a target file has a template mirror under `internal/template/templates/`, every content edit **shall** be applied to both the local file and its template mirror, and `make build` **shall** be run after the mirror edits so the embedded template FS is regenerated.
- **REQ-DU-015** (Unwanted behavior): The template-mirror content **shall not** leak the internal SPEC ID, internal dates, commit SHAs, or audit citations (§25 neutrality); only the generic unknowns-framework prose **shall** ship in templates.

## §D. Acceptance Criteria (summary)

The canonical AC enumeration is the SSOT in `acceptance.md` (AC-DU-001 .. AC-DU-017 — 17 ACs). Mapping to the 15 REQs: REQ-DU-002 is covered by AC-DU-016 (its unfamiliar-domain trigger + before-plan-phase-entry timing); REQ-DU-009 and REQ-DU-013 share the no-machinery / no-Go-change checks AC-DU-008 + AC-DU-015; REQ-DU-014 is covered by AC-DU-012 (per-file mirror token) + AC-DU-013 (`make build` exit-0) + AC-DU-017 (the enforced CI parity guards `TestRuleTemplateMirrorDrift` + `TestSanitizedPairParity`); all other REQs map 1:1 to a single AC. Each AC is independently testable via grep-verifiable section presence, template-mirror parity check, `make build` exit-0, an enforced parity `go test` (exit-0), or a no-Go-change assertion. See `acceptance.md` § D AC Matrix + § Given-When-Then scenarios.

## §E. Out of Scope (exclusions)

The items below are explicitly **out of scope** for this SPEC. Each is expressed as an `### Out of Scope — <topic>` sub-heading with bullet exclusions.

### Out of Scope — Go code and new machinery

- No changes to Go source under `internal/`, `pkg/`, or `cmd/`.
- No new lint rule, no new CLI subcommand, no new hook script, no schema/frontmatter field.
- No new agent file, no new skill, no new subsystem. The three enhancements are documentation/rule/agent-body prose only.

### Out of Scope — re-adding already-covered Discovery activities

- Interviews are already covered by AskUserQuestion Socratic interview rounds — NOT re-added.
- Reference/library-doc discovery is already covered by `Agent(Explore)` + Context7 MCP + WebFetch — NOT re-added.
- This SPEC adds only the three genuinely-absent items (T1, T2, T3) per §A.3.

### Out of Scope — mandatory-gate and full-taxonomy expansion

- The Blind Spot Pass is NOT promoted to a mandatory plan-phase gate; it stays optional (triggered only on suspected unknown-unknowns).
- The 4-quadrant taxonomy is NOT expanded into a full standalone rule file or a CLAUDE.md section block; CLAUDE.md §7 Rule 5 receives only a one-line framing + SSOT pointer (diet-respecting).
- No Tier-2/Tier-3 items from the broader article (automated blind-spot scoring, unknowns-metrics dashboards, etc.).

### Out of Scope — interactive-IDE transplant

- The article's interactive human-in-IDE pairing loop is NOT transplanted verbatim; the orchestrator model (orchestrator + Explore surfacing to the user via AskUserQuestion) is the adaptation target.

## §F. Cross-References

- `.claude/rules/moai/core/askuser-protocol.md` — Context-First Discovery SSOT (T1 host + T3 host).
- `CLAUDE.md` §7 Rule 1 (Approach-First) + Rule 5 (Context-First Discovery) — T2 (Rule 1) + T1/T3 wires (Rule 5).
- `.claude/agents/moai/manager-spec.md` § Step 4 / plan.md authoring guidance — T2 host.
- `.claude/rules/moai/workflow/spec-workflow.md` § Plan Phase — T1 wire host.
- `CLAUDE.local.md` §2 (Template-First) + §25 (Template Internal-Content Isolation) — paired-edit + neutrality constraints (REQ-DU-014, REQ-DU-015).
- Source article: "A Field Guide to Fable: Finding Your Unknowns" (Thariq / @trq212) — the framework being adapted (generic; provenance kept out of shipped template content per §25).
