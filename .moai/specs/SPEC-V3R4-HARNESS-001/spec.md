---
id: SPEC-V3R4-HARNESS-001
version: "0.1.0"
status: draft
created: 2026-05-14
updated: 2026-05-14
author: manager-spec
priority: P0
tags: "harness, self-evolution, cli-retirement, v3r4, foundation, lifecycle-consolidation, breaking"
issue_number: null
title: Unified Self-Evolving Harness Foundation (CLI Retirement + Lifecycle Consolidation)
phase: "v3.0.0 R4 — Phase A — Foundation"
module: ".claude/skills/moai/, .claude/skills/moai-harness-learner/, .claude/skills/moai-meta-harness/, .claude/agents/my-harness/, .moai/harness/, internal/cli/harness.go (deprecation marker only)"
dependencies: []
supersedes:
  - SPEC-V3R3-HARNESS-001
  - SPEC-V3R3-HARNESS-LEARNING-001
  - SPEC-V3R3-PROJECT-HARNESS-001
related_specs:
  - SPEC-V3R3-DESIGN-PIPELINE-001
  - SPEC-AGENCY-ABSORB-001
breaking: true
bc_id:
  - BC-V3R4-HARNESS-001-CLI-RETIREMENT
lifecycle: spec-anchored
related_theme: "Self-Evolving Harness v2 — Foundation Wave"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-HARNESS-001 — Unified Self-Evolving Harness Foundation

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-14 | manager-spec | Initial draft. Foundation SPEC for the self-evolving harness v2 system. Supersedes the three V3R3 harness SPECs (HARNESS-001 meta-skill, HARNESS-LEARNING-001 learning subsystem, PROJECT-HARNESS-001 socratic-interview activation) into a single V3R4 architecture. Retires the `moai harness <verb>` Go CLI subcommand path (BC-V3R4-HARNESS-001-CLI-RETIREMENT). Establishes the slash-command-only lifecycle contract (`/moai:harness`) backed by skill workflow + dedicated subagents + Claude Code hooks. Preserves the 5-Layer Safety architecture and FROZEN zones as immutable. Derives from `/moai brain` IDEA-004 (24 cited sources: revfactory/harness Apache-2.0, Reflexion arXiv:2303.11366, Voyager arXiv:2305.16291, Constitutional AI arXiv:2212.08073, LangGraph reflection production patterns 2026, Anthropic Claude Code Skills/Agents official docs 2026). Foundation for downstream SPEC-V3R4-HARNESS-002 through SPEC-V3R4-HARNESS-008. |

---

## 1. Goal

The MoAI-ADK harness system MUST consolidate three prior V3R3 SPECs into one V3R4 unified architecture whose entire lifecycle (generation, observation, learning, evolution, application) operates exclusively through the `/moai:harness` slash command surface, dedicated subagent execution, and Claude Code hooks — with zero dependency on a Go binary CLI subcommand. The Go CLI verb path (`moai harness status|apply|rollback|disable`) MUST be retired as a public surface; its implementation file remains in the tree as a deprecation marker awaiting downstream removal, but it MUST NOT be registered as a public cobra subcommand. The 5-Layer Safety architecture from the design constitution §5 MUST be preserved verbatim, the FROZEN zones MUST remain immutable, and every Tier-4 evolution application MUST be gated by an orchestrator-issued AskUserQuestion approval round. This foundation SPEC establishes the architectural baseline that seven downstream SPECs (002 through 008) will progressively build upon to introduce multi-event observation, embedding-cluster pattern detection, Reflexion self-critique, principle-based scoring, multi-objective effectiveness measurement, skill-library auto-organization, and (deferred) cross-project lesson federation.

### 1.1 Background

- The current Go CLI command path `moai harness <verb>` is implemented in `internal/cli/harness.go` (the `newHarnessCmd` factory) but is NOT registered as a public subcommand in `internal/cli/root.go`. As of v2.14.0, invoking `moai harness status` from the shell returns cobra's `unknown command` error. The `/moai:harness` slash command was a thin wrapper that historically routed through this CLI path; recent work (PR #908, commit `452aa638f`) replaced it with a skill-only thin wrapper (`Skill("moai") arguments: harness $ARGUMENTS`) that does not invoke the binary. This SPEC formalizes that absence as the new contract and prevents future re-registration via CI guard.
- The three V3R3 harness SPECs were authored separately under different Phase letters (Phase C, Phase D) and ship under different release targets (v2.17.0, v2.19.0). They share the same architectural domain but were written before the V3R4 self-evolution vision crystallized. Each one introduced a partial mechanism — meta-skill generation (HARNESS-001), 4-tier learning ladder with PostToolUse observer (HARNESS-LEARNING-001), socratic interview activation (PROJECT-HARNESS-001) — and the surface area of those three SPECs no longer maps cleanly to the V3R4 unified architecture. Treating them as superseded under one foundation SPEC reduces cognitive load and dependency churn for the seven downstream SPECs.
- The user's `/moai brain` IDEA-004 session (completed 2026-05-14) produced research.md with 24 cited external sources spanning the academic foundations of self-improving agents (Reflexion verbal reinforcement, Voyager skill library, Constitutional AI principle scoring) and industrial patterns (LangGraph reflection 2-3 iteration sweet spot, Anthropic Skills/Agents 2026 namespace merger). The proposal.md output explicitly recommended the eight-SPEC decomposition with this foundation SPEC executing first. The user has locked in five operational decisions captured in §1.2 below.
- The design constitution at `.claude/rules/moai/design/constitution.md` §2 (Frozen vs Evolvable Zones) and §5 (Safety Architecture — 5 Layers) already encodes the safety stack and immutability boundary the V3R4 architecture inherits without modification. This SPEC re-asserts and references those rules but does not modify the constitution file itself.

### 1.2 User-Locked Operational Decisions (from IDEA-004 Phase 1 Discovery)

The following decisions are locked-in and MUST NOT be re-litigated within this foundation SPEC or any downstream SPEC:

1. **CLI retirement**: `moai harness <verb>` is retired as a public surface. The slash command `/moai:harness` is the only supported invocation path.
2. **Autonomy ceiling**: Tier-4 application always requires `AskUserQuestion` approval from the orchestrator. The 5-Layer Safety stack remains in force. No autonomy mechanism may bypass these gates.
3. **Success metric**: weekly Tier-4 application count per project, combined with the Tier-4 reach rate (% of observations promoted through Tier 1 → 2 → 3 → 4 ladder). The aspirational +60% effectiveness figure from revfactory/harness is downgraded to a design-intent reference, not a v1 success threshold.
4. **Conflict resolution**: when Reflexion-style self-critique (downstream SPEC-V3R4-HARNESS-004) and `evaluator-active` scoring disagree on an evolution proposal, evaluator-active retains absolute veto power. Self-critique runs first as a pre-screen to save tokens; evaluator-active runs second as the binding gate.
5. **AskUserQuestion fatigue mitigation**: the initial rate limit is 1 Tier-4 application per project per 7-day rolling window, expandable adaptively based on the user acceptance rate. The adaptive expansion mechanism is deferred to a downstream SPEC; this foundation SPEC ships only the initial conservative limit.

### 1.3 Non-Goals

This SPEC is the FOUNDATION. The following capabilities are explicitly OUT OF SCOPE and are deferred to the named downstream SPECs:

- Multi-event observer expansion across PostToolUse + Stop + SubagentStop + UserPromptSubmit events — deferred to `SPEC-V3R4-HARNESS-002`. This SPEC ships the PostToolUse-only baseline that already exists.
- Embedding-cluster pattern detection replacing the current frequency-count ladder — deferred to `SPEC-V3R4-HARNESS-003`. This SPEC preserves the existing 4-tier observation/heuristic/rule/auto-update ladder unchanged.
- Reflexion-style verbal self-critique loop with hard iteration cap — deferred to `SPEC-V3R4-HARNESS-004`.
- Principle-based self-scoring against the design constitution — deferred to `SPEC-V3R4-HARNESS-005`.
- Multi-objective effectiveness measurement (quality + token cost + latency + iteration count tuple) and auto-rollback on any-axis regression — deferred to `SPEC-V3R4-HARNESS-006`.
- Voyager-style skill library auto-organization with embedding-indexed retrieval — deferred to `SPEC-V3R4-HARNESS-007`.
- Cross-project lesson federation with anonymization namespace isolation — deferred (privacy-sensitive) to `SPEC-V3R4-HARNESS-008`, which will not enter plan-phase until SPECs 002-007 demonstrate single-project track record over multiple release cycles.
- Migration tooling that imports historical `usage-log.jsonl` entries into a new schema — deferred entirely (no SPEC). Users begin v2 with empty state.
- GUI or dashboard for evolution-history inspection — out of scope. The text artifacts under `.moai/harness/` remain the only interface.

---

## 2. Scope

### 2.1 In Scope

- Retire `moai harness <verb>` as a public cobra subcommand. The `newHarnessCmd` factory in `internal/cli/harness.go` MUST NOT be registered into `internal/cli/root.go`. The implementation file remains in the tree as a deprecation marker awaiting downstream physical removal in a follow-up SPEC; this foundation SPEC asserts the non-registration contract only.
- Establish the CLI-free harness lifecycle contract via three artifact families:
  - The `/moai:harness` slash command (`.claude/commands/moai/harness.md`) as a thin routing wrapper that delegates to the `moai` skill workflow with `harness $ARGUMENTS`.
  - The `moai` skill body and its `workflows/harness.md` workflow module owning all lifecycle verbs (status, apply, rollback, disable) without invoking any Go binary.
  - The dedicated subagents (`moai-harness-learner` skill and any orchestrating skill the workflow body chains to) owning per-phase execution; the `moai-meta-harness` skill continues to own the meta-harness 7-Phase workflow for project-specific skill generation.
- Re-assert the V3R4 unified architecture as the consolidation of the three superseded V3R3 SPECs. The supersedes relationship is declared in this SPEC's frontmatter. The three V3R3 SPECs themselves MUST NOT be modified within this PR — their status transition to `superseded` is a follow-up commit owned by `manager-git` after this SPEC merges, per delegation contract.
- Re-assert the 5-Layer Safety pipeline (L1 Frozen Guard, L2 Canary Check, L3 Contradiction Detector, L4 Rate Limiter, L5 Human Oversight) as preserved in full from the design constitution §5. No layer may be removed, bypassed, or weakened by any harness mechanism introduced in this SPEC or any downstream SPEC.
- Re-assert the FROZEN zone immutability contract: the harness system MUST NOT auto-modify any path under `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**`. The L1 Frozen Guard MUST log every violation attempt to a path-prefix-matched audit log.
- Re-assert the AskUserQuestion contract for Tier-4 application: only the MoAI orchestrator may invoke `AskUserQuestion`. Subagents that require user input MUST return a structured blocker report to the orchestrator and MUST NOT prompt the user directly. This contract is declared in `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary and `.claude/rules/moai/core/askuser-protocol.md` and is inherited verbatim by every harness-related subagent.
- Re-assert the artifact storage roots:
  - `.moai/harness/` for project-local harness state (usage log, learning history, snapshots, proposals, configuration overrides).
  - `.claude/agents/my-harness/` for project-generated specialist agent definitions.
  - `.claude/skills/my-harness-*/` for project-generated specialist skill bodies.
  - `.moai/archive/skills/<v>/` for skills archived during prior breaking changes (preserved unchanged from V3R3).
- Declare the AskUserQuestion fatigue-mitigation initial policy: at most one Tier-4 application per project per 7-day rolling window. The expansion mechanism is a downstream SPEC.
- Declare the success-metric exposure contract: weekly Tier-4 application count and Tier-4 reach rate MUST be readable from a machine-parsable file under `.moai/harness/` via the `/moai:harness status` skill verb without invoking any Go binary.

### 2.2 Out of Scope

The seven items listed in §1.3 (Non-Goals). In addition:

- Physically deleting `internal/cli/harness.go` or `internal/cli/harness_test.go`. This foundation SPEC marks them as deprecation candidates only. A follow-up SPEC will remove the files after the downstream SPECs 002 through 008 are merged.
- Modifying the design constitution file. The constitution remains FROZEN. This SPEC references its existing rules.
- Modifying the three superseded V3R3 SPECs. Their status transition is a follow-up `manager-git` commit after this SPEC merges.
- Implementing any new Tier-classifier algorithm, scoring rubric, observation-cluster algorithm, or effectiveness-decay pruning. The existing 4-tier frequency-count ladder is preserved unchanged; replacements live in SPEC-V3R4-HARNESS-003 and beyond.
- Touching `.claude/agents/my-harness/*.md` skill bodies — these are user-area artifacts inviolate to MoAI updates per V3R3-HARNESS-001 REQ-HARNESS-008. The contract is inherited unchanged.
- Networking, telemetry upload, or any cross-machine data exchange. The harness remains strictly local.

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| Solo developer (GOOS-style power-user) | Stable slash-command surface across V3R3 → V3R4 transition; no broken muscle memory for `/moai:harness` invocations; no surprise Tier-4 applications without explicit approval. |
| Project tech lead (2-5 projects) | Predictable harness behavior; uniform AskUserQuestion contract across projects; ability to disable learning per-project via existing `.moai/config/sections/harness.yaml`. |
| MoAI-ADK maintainer | Reduced architectural surface (single V3R4 SPEC family instead of three V3R3 SPECs); clear contract for downstream SPEC authors; verifiable CI guard against accidental CLI re-registration. |
| `plan-auditor` subagent | Single foundation SPEC to audit; clear `supersedes` relationship for impact analysis; verbatim 5-Layer + FROZEN re-assertion gives auditor an unambiguous compliance checklist. |
| `evaluator-active` subagent | Unchanged binding-gate role; conflict resolution with downstream Reflexion self-critique (SPEC-004) explicitly locked-in as veto-retains-evaluator. |
| `manager-git` subagent | Owns the follow-up V3R3 status-transition commit; this SPEC delegates that task explicitly. |
| User upgrading from V3R3 | Slash command continues to work; CLI verb path returns cobra's standard `unknown command` error; no automatic migration of existing `.moai/harness/usage-log.jsonl` (fresh-start documented in release notes). |

---

## 4. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these within this foundation PR is a scope violation:

1. Any embedding-model integration, embedding-cluster algorithm, or vector index — deferred to SPEC-V3R4-HARNESS-003.
2. Any Reflexion-style self-critique loop, episodic memory of natural-language reflections, or iteration-cap mechanism beyond what already exists in the current `moai-harness-learner` skill body — deferred to SPEC-V3R4-HARNESS-004.
3. Any principle-based scoring rubric, constitution-parsing logic, or pre-screen step before AskUserQuestion — deferred to SPEC-V3R4-HARNESS-005.
4. Any multi-objective scoring tuple, quality/cost/latency/iteration tracking, or auto-rollback-on-regression mechanism — deferred to SPEC-V3R4-HARNESS-006.
5. Any embedding-indexed skill library, top-K retrieval logic, or compositional skill reuse system — deferred to SPEC-V3R4-HARNESS-007.
6. Any cross-project lesson sharing, anonymization layer, federation namespace, or opt-in approval flow — deferred to SPEC-V3R4-HARNESS-008.
7. Physical deletion of `internal/cli/harness.go` or `internal/cli/harness_test.go`. They remain in the tree as deprecation markers.
8. Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file).
9. Modification of the three superseded V3R3 SPEC files. Their status transition is a follow-up `manager-git` commit.
10. Modification of any file under `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` — except the existing `moai-meta-harness` and `moai-harness-learner` SKILL.md may receive minor text annotations reaffirming the V3R4 contract (boundary: text-only, no behavioral change). Any annotation MUST be additive and MUST NOT touch frozen frontmatter fields.
11. Migration tooling to convert pre-existing `.moai/harness/usage-log.jsonl` data into any new format. Users continue with whatever state they have on disk; no schema migration occurs in this SPEC.
12. Any GUI, dashboard, web client, or non-terminal interface for harness state inspection.
13. Any network call, telemetry upload, or external API integration.

---

## 5. Requirements (EARS format)

Each requirement is identified by `REQ-HRN-FND-NNN` and uses one of the five EARS patterns: Ubiquitous (system shall always), Event-Driven (when X then system shall Y), State-Driven (while X the system shall Y), Optional Feature (where X exists the system shall Y), Unwanted Behavior (if X then system shall not / shall reject Y). Most lifecycle rules use Event-Driven because the harness is an inherently event-triggered system.

### REQ-HRN-FND-001 (Ubiquitous — CLI invocation retirement)

The system **shall not** expose `moai harness <verb>` as a public cobra subcommand. The Go CLI verb path is retired; all harness lifecycle operations **shall** be reachable exclusively through the `/moai:harness` slash command surface.

### REQ-HRN-FND-002 (Unwanted — CLI re-registration prevention)

**If** any change introduces a registration of `newHarnessCmd` (or any equivalent harness subcommand factory) into `internal/cli/root.go` or any other cobra command tree, **then** the continuous integration system **shall** fail the build with a diagnostic message referencing this SPEC. This guard is enforced by a CI test that asserts the absence of harness subcommand registration in the cobra command tree.

### REQ-HRN-FND-003 (Ubiquitous — slash-command verb coverage)

The `/moai:harness` slash command **shall** support, at minimum, the verbs `status`, `apply`, `rollback`, and `disable`. Each verb **shall** be fully implemented in the `moai` skill workflow body (specifically `.claude/skills/moai/workflows/harness.md` or its successor under the V3R4 contract) and **shall not** invoke any Go binary subcommand.

### REQ-HRN-FND-004 (Event-Driven — Tier 4 AskUserQuestion gate)

**When** the harness lifecycle prepares to apply a Tier-4 evolution proposal to the EVOLVABLE zone, the orchestrator **shall** invoke `AskUserQuestion` with a structured options list (Apply / Modify / Defer / Reject, with the first option marked `(권장)` or `(Recommended)`) **before** any file modification is performed. Subagents engaged in the lifecycle **shall not** invoke `AskUserQuestion`; if they require user input, they **shall** return a structured blocker report to the orchestrator.

### REQ-HRN-FND-005 (Ubiquitous — 5-Layer Safety preservation)

The system **shall** preserve the 5-Layer Safety architecture defined in `.claude/rules/moai/design/constitution.md` §5 in its current form: L1 Frozen Guard, L2 Canary Check (rejection on score-drop > 0.10), L3 Contradiction Detector, L4 Rate Limiter (≤ 3 evolutions per week, ≥ 24h cooldown), L5 Human Oversight (AskUserQuestion at Tier 4). No layer **shall** be removed, bypassed, or weakened by any harness mechanism within this SPEC or any downstream SPEC.

### REQ-HRN-FND-006 (Unwanted — FROZEN zone path-prefix protection)

**If** the harness system attempts to modify any path matching the prefixes `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, or `.moai/project/brand/`, **then** the L1 Frozen Guard **shall** block the operation and **shall** log the attempt to `.moai/harness/learning-history/frozen-guard-violations.jsonl` with at minimum: ISO-8601 timestamp, attempted target path, calling subagent or skill identifier, and rejection rationale. The block **shall not** be bypassable by any configuration setting.

### REQ-HRN-FND-007 (Event-Driven — pre-modification snapshot)

**When** an evolution is approved at the Tier-4 gate and the application step begins, the system **shall** create a snapshot directory at `.moai/harness/learning-history/snapshots/<ISO-DATE>/` containing byte-identical copies of every file the modification will touch, **before** any write occurs. The snapshot manifest **shall** record absolute target paths and a content hash for each file.

### REQ-HRN-FND-008 (Event-Driven — rollback to snapshot)

**When** the user invokes `/moai:harness rollback <date>` and a snapshot exists at `.moai/harness/learning-history/snapshots/<date>/`, the system **shall** restore each file in the snapshot manifest to its byte-identical pre-modification state. **If** no matching snapshot exists, the system **shall** return a diagnostic message and **shall not** modify any file.

### REQ-HRN-FND-009 (State-Driven — observer no-op when disabled)

**While** the configuration key `learning.enabled` in `.moai/config/sections/harness.yaml` resolves to `false`, the PostToolUse observer hook **shall** be a complete no-op: it **shall not** read, write, or append to `.moai/harness/usage-log.jsonl` nor invoke any tier-classification or evolution logic. Disabling **shall not** delete existing log entries.

### REQ-HRN-FND-010 (Ubiquitous — PostToolUse baseline observation)

The system **shall** continue to record observations via the PostToolUse hook into `.moai/harness/usage-log.jsonl` using a JSONL append-only format with at minimum: ISO-8601 timestamp, event_type, subject, and a context hash. The multi-event observer expansion (Stop, SubagentStop, UserPromptSubmit) is deferred to `SPEC-V3R4-HARNESS-002` and is explicitly out of scope for this foundation SPEC.

### REQ-HRN-FND-011 (Ubiquitous — 4-tier ladder preservation)

The system **shall** preserve the existing 4-tier observation ladder with thresholds 1 (Observation), 3 (Heuristic), 5 (Rule), 10 (Auto-update) as defined in `SPEC-V3R3-HARNESS-LEARNING-001` REQ-HL-002. Replacement of the frequency-count classifier with an embedding-cluster classifier is deferred to `SPEC-V3R4-HARNESS-003`. Replacement of the promotion-trigger logic with a Reflexion self-critique loop is deferred to `SPEC-V3R4-HARNESS-004`.

### REQ-HRN-FND-012 (Ubiquitous — AskUserQuestion rate limit)

The system **shall** apply an initial Tier-4 application rate limit of at most one application per project per rolling 7-day window. The orchestrator **shall not** invoke `AskUserQuestion` for a Tier-4 approval more than once within this window unless the user has explicitly raised the limit through a mechanism introduced by a downstream SPEC. The adaptive expansion algorithm based on user acceptance rate is out of scope for this foundation SPEC.

### REQ-HRN-FND-013 (Unwanted — superseded SPEC mutation prevention within this PR)

**If** any commit within the pull request implementing this SPEC modifies the file content of `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`, `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`, or `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md`, **then** the pull request **shall** be considered out of scope and **shall** be rejected. The status transition of the three superseded SPECs to `status: superseded` is the responsibility of a follow-up `manager-git` commit after this SPEC merges.

### REQ-HRN-FND-014 (Event-Driven — frozen-guard violation audit log)

**When** the L1 Frozen Guard rejects a write attempt per REQ-HRN-FND-006, the system **shall** emit a JSONL entry to `.moai/harness/learning-history/frozen-guard-violations.jsonl` and **shall not** raise an error to the user. The entry **shall** include ISO-8601 timestamp and the target path that was rejected. The user **shall** be able to inspect the log via `/moai:harness status` or by reading the file directly.

### REQ-HRN-FND-015 (Ubiquitous — AskUserQuestion subagent prohibition)

The system **shall** enforce the orchestrator-only AskUserQuestion contract from `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary across every harness-related subagent. Subagent definitions in `.claude/agents/moai/` (specifically `moai-harness-learner` and any future V3R4 successor) **shall not** contain calls to `AskUserQuestion`; if a subagent requires user input, it **shall** return a structured blocker report and the orchestrator **shall** run the AskUserQuestion round and re-delegate.

### REQ-HRN-FND-016 (Ubiquitous — success-metric exposure)

The system **shall** expose two telemetry values readable via `/moai:harness status` without invoking any Go binary: the weekly Tier-4 application count per project (most recent 7-day rolling window) and the Tier-4 reach rate (percentage of unique patterns promoted from Tier 1 through Tier 4 across the lifetime of the project's `.moai/harness/usage-log.jsonl`). Both values **shall** be derivable from `.moai/harness/learning-history/tier-promotions.jsonl` and `.moai/harness/usage-log.jsonl` without external computation.

### REQ-HRN-FND-017 (Event-Driven — Reflexion-evaluator conflict resolution placeholder)

**When** a downstream SPEC (specifically SPEC-V3R4-HARNESS-004 introducing Reflexion self-critique and SPEC-V3R4-HARNESS-005 introducing principle-based scoring) reaches a state where the self-critique result and the `evaluator-active` score disagree on a proposal, the system **shall** treat the `evaluator-active` verdict as the binding gate and **shall** treat the Reflexion pre-screen as advisory only. This requirement is a contract assertion that constrains downstream SPECs; this foundation SPEC ships no Reflexion implementation.

### REQ-HRN-FND-018 (Optional Feature — adaptive rate-limit expansion deferral)

**Where** a downstream SPEC introduces an adaptive rate-limit expansion mechanism for Tier-4 applications based on user acceptance rate, the system **shall** preserve the initial floor of one Tier-4 application per project per 7-day rolling window as the minimum guarantee. The expansion **shall** never lower this floor; tightening is permitted.

---

## 6. Acceptance Coverage Map

The acceptance criteria are defined in `acceptance.md` (sibling file). The coverage map below shows every REQ from §5 mapped to at least one AC. Full Given-When-Then scenarios are in `acceptance.md`.

| AC ID | Covers REQ IDs |
|-------|----------------|
| AC-HRN-FND-001 | REQ-HRN-FND-001, REQ-HRN-FND-002 |
| AC-HRN-FND-002 | REQ-HRN-FND-003 |
| AC-HRN-FND-003 | REQ-HRN-FND-004, REQ-HRN-FND-015 |
| AC-HRN-FND-004 | REQ-HRN-FND-005 |
| AC-HRN-FND-005 | REQ-HRN-FND-006, REQ-HRN-FND-014 |
| AC-HRN-FND-006 | REQ-HRN-FND-007, REQ-HRN-FND-008 |
| AC-HRN-FND-007 | REQ-HRN-FND-009 |
| AC-HRN-FND-008 | REQ-HRN-FND-010, REQ-HRN-FND-011 |
| AC-HRN-FND-009 | REQ-HRN-FND-012, REQ-HRN-FND-018 |
| AC-HRN-FND-010 | REQ-HRN-FND-013 |
| AC-HRN-FND-011 | REQ-HRN-FND-016 |
| AC-HRN-FND-012 | REQ-HRN-FND-017 |

Coverage: 18 REQs ↔ 12 ACs, every REQ appears in at least one AC.

---

## 7. Constraints

[HARD] Language: All SPEC artifact content in English (per `.claude/rules/moai/development/coding-standards.md` § Language Policy). Internal reasoning during drafting may use Korean.

[HARD] FROZEN zone preservation: `.claude/rules/moai/design/constitution.md` §2 and §5 are NOT modified by this SPEC.

[HARD] No modification of superseded V3R3 SPECs within this PR (REQ-HRN-FND-013 enforces this).

[HARD] EARS format mandatory for all REQs. Every REQ uses one of the five EARS patterns listed in §5 preamble.

[HARD] Conventional Commits format for all commits originating from this SPEC.

[HARD] No tech-stack implementation assumptions in spec.md. Requirements describe contracts and capabilities, not implementations. Implementation decisions belong to `plan.md`.

[HARD] No emojis in user-facing output (per `.claude/rules/moai/development/coding-standards.md` § Content Restrictions).

[HARD] No time estimates or duration predictions in any artifact (per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation). Use priority labels (P0 / P1 / P2 / P3) and phase ordering instead.

[HARD] Plan-in-main doctrine deviation (this SPEC branches from `feat/cmd-harness-slash-wrapper` instead of `main`): user-authorized in the spawn prompt. Commits are scoped to `.moai/specs/SPEC-V3R4-HARNESS-001/` only; unrelated working-tree drift is NOT bundled.

---

## 8. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Downstream SPECs 002-008 introduce mechanisms that conflict with the FROZEN re-assertions in this SPEC | Medium | This SPEC's REQs are explicit and EARS-formatted; downstream plan-auditor runs will catch conflicts. The FROZEN re-assertion (REQ-HRN-FND-005, -006) is the authoritative contract. |
| CI guard for REQ-HRN-FND-002 (CLI re-registration prevention) is non-trivial to author | Medium | The check is a static text scan of `internal/cli/root.go` plus a cobra-tree assertion in a Go test. Both patterns exist elsewhere in `internal/template/`. Risk is implementation cost, not correctness. |
| User confusion when `moai harness status` returns `unknown command` after upgrade | Low | Release notes for the target release MUST document the migration; the slash command `/moai:harness` continues to work. The thin wrapper at `.claude/commands/moai/harness.md` does NOT invoke the binary. |
| The three superseded V3R3 SPECs continue to be referenced by stale memory entries or legacy docs | Medium | The `supersedes:` field in this SPEC's frontmatter and the follow-up `manager-git` commit transitioning their status are the canonical signals. Downstream SPECs cite this SPEC, not the V3R3 SPECs. |
| Tier-4 rate limit of 1/week is too conservative and frustrates power-users | Medium | The limit is explicit in REQ-HRN-FND-012 with a documented expansion path (REQ-HRN-FND-018, deferred). Power-users can disable learning entirely via `learning.enabled: false` (REQ-HRN-FND-009). |
| Reflexion-evaluator conflict resolution (REQ-HRN-FND-017) is asserted before the Reflexion implementation exists | Low | Intentional — the contract precedes the implementation. SPEC-V3R4-HARNESS-004 will reference this REQ as its binding constraint. |

---

## 9. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| `SPEC-V3R3-HARNESS-001` | Superseded by this SPEC | Meta-harness skill creation and 16-skill BC removal — content remains as historical reference; status transitions to `superseded` via follow-up `manager-git` commit. |
| `SPEC-V3R3-HARNESS-LEARNING-001` | Superseded by this SPEC | 4-tier learning ladder + 5-Layer Safety + `/moai harness` CLI verbs — CLI verb path retired by this SPEC; 4-tier ladder preserved (REQ-HRN-FND-011); 5-Layer Safety preserved verbatim (REQ-HRN-FND-005). |
| `SPEC-V3R3-PROJECT-HARNESS-001` | Superseded by this SPEC | `/moai project` Phase 5+ socratic interview + 5-Layer integration wiring — preserved as runtime behavior; this SPEC formalizes the V3R4 contract those layers operate under. |
| `SPEC-V3R3-DESIGN-PIPELINE-001` | Related (non-blocking) | Design pipeline §11 GAN Loop references the same evaluator-active veto contract that REQ-HRN-FND-017 asserts. |
| `SPEC-AGENCY-ABSORB-001` | Reference | Established the FROZEN/EVOLVABLE zone discipline that this SPEC inherits. |
| `SPEC-V3R4-HARNESS-002` through `SPEC-V3R4-HARNESS-008` | Blocked by this SPEC | Downstream SPECs cannot enter plan-phase until this foundation SPEC merges. Each downstream SPEC references this SPEC's REQ IDs as binding contracts. |

---

## 10. Glossary

- **CLI verb path**: The `moai harness <verb>` subcommand sequence invoked from the terminal. Retired by this SPEC.
- **Slash command path**: The `/moai:harness <verb>` invocation inside a Claude Code session. The only supported path post-V3R4.
- **5-Layer Safety**: The five layers defined in design constitution §5 — Frozen Guard, Canary Check, Contradiction Detector, Rate Limiter, Human Oversight.
- **FROZEN zone**: Path prefixes the harness MUST NOT auto-modify. Defined in REQ-HRN-FND-006 and design constitution §2.
- **EVOLVABLE zone**: Path prefixes the harness MAY modify within safety gates. Specifically `.claude/skills/my-harness-*/`, `.claude/agents/my-harness/`, `.moai/harness/` (excluding learning-history audit logs).
- **Tier 1-4**: Observation classification ladder with thresholds 1 / 3 / 5 / 10. Preserved unchanged from `SPEC-V3R3-HARNESS-LEARNING-001` REQ-HL-002.
- **Reflexion self-critique**: Verbal-reinforcement loop deferred to SPEC-V3R4-HARNESS-004. Referenced in REQ-HRN-FND-017 as advisory pre-screen.
- **Evaluator-active veto**: The binding gate role of `evaluator-active` over harness evolution proposals. Asserted in REQ-HRN-FND-017.
- **Supersedes relationship**: Frontmatter relationship `supersedes:` declaring that this SPEC replaces the listed V3R3 SPECs in the V3R4 architecture. Status transition of the superseded SPECs is a follow-up commit, not bundled with this PR.

---

## 11. References

- Brain workflow artifacts (mandatory upstream context):
  - `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-004/proposal.md` — Vision and 8-SPEC decomposition.
  - `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-004/ideation.md` — Lean Canvas and Critical Evaluation.
  - `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-004/research.md` — 24 cited external sources.
- Superseded V3R3 SPECs (historical reference):
  - `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`
  - `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`
  - `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md`
- Design constitution (FROZEN):
  - `.claude/rules/moai/design/constitution.md` §2 (Frozen vs Evolvable Zones), §5 (Safety Architecture), §11 (GAN Loop).
- Orchestrator-subagent contract (FROZEN):
  - `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
  - `.claude/rules/moai/core/askuser-protocol.md` — Canonical AskUserQuestion protocol.
- Workflow rules:
  - `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
  - `.claude/rules/moai/development/coding-standards.md` § Language Policy + § Content Restrictions.
- External academic sources (cited from IDEA-004 research.md):
  - Reflexion (Shinn et al., NeurIPS 2023, arXiv:2303.11366)
  - Voyager (Wang et al., NVIDIA + Caltech, arXiv:2305.16291)
  - Constitutional AI / RLAIF (Bai et al., Anthropic 2022, arXiv:2212.08073)
  - LangGraph Reflection Pattern (production guidance, 2026)
  - Anthropic Claude Code Skills/Agents documentation (2026)
  - revfactory/harness Apache-2.0 (https://github.com/revfactory/harness)
