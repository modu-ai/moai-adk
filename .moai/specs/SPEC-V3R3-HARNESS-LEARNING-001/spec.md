---
id: SPEC-V3R3-HARNESS-LEARNING-001
title: Self-Learning Dynamic Harness — User-area Auto-Evolution from Activity Signals
version: "0.1.0"
status: draft
created: 2026-04-26
updated: 2026-04-26
author: manager-spec
priority: P1 High
phase: "v3.0.0 R3 — Phase D — Adaptive Harness"
module: ".claude/skills/moai-harness-learner/, internal/harness/, .moai/harness/, .claude/skills/my-harness-*/, .claude/agents/my-harness/"
dependencies:
  - SPEC-V3R3-HARNESS-001
  - SPEC-V3R3-PROJECT-HARNESS-001
related_specs:
  - SPEC-DESIGN-CONST-AMEND-001
  - SPEC-AGENCY-ABSORB-001
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "harness, learning, adaptive, auto-evolution, frozen-guard, user-area, observer, v3r3, phase-d"
related_theme: "Phase D — Adaptive Harness"
target_release: v2.17.0
issue_number: null
---

# SPEC-V3R3-HARNESS-LEARNING-001: Self-Learning Dynamic Harness

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. Phase D P1 — User-area dynamic harness self-learning from activity signals (PostToolUse-driven). |

---

## 1. Goal (목적)

User-installed dynamic harness — comprising `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`, and `.moai/harness/` artifacts — MUST evolve autonomously from observable user activity (subcommand frequency, SPEC-ID patterns, commit message trends, agent invocation stats, explicit `/moai feedback`). The moai-managed area (`.claude/agents/moai/`, `.claude/skills/moai-*/`) MUST remain inviolable. Evolution operates under a 4-tier confidence pipeline (Observation → Heuristic → Rule → Auto-update) and a 5-layer safety architecture mirroring the design system constitution §5.

### 1.1 Background

- SPEC-V3R3-HARNESS-001 generates the meta-harness skill that produces the user-area harness skeleton (`my-harness/*`).
- SPEC-V3R3-PROJECT-HARNESS-001 conducts a Socratic interview producing the baseline customization captured in `.moai/harness/main.md` and per-phase extensions.
- Without auto-evolution, the user must manually edit harness skills as their workflow shifts. Patterns observable from `/moai` invocations, commit cadence, and agent usage already encode the evolution signal — the system MUST extract it.
- Design constitution §5 (FROZEN) defines the 5-layer safety architecture (Frozen Guard, Canary Check, Contradiction Detector, Rate Limiter, Human Oversight). This SPEC mirrors §5 verbatim into the user-harness self-learning subsystem.

### 1.2 Non-Goals

- This SPEC does NOT modify any file under `.claude/agents/moai/` or `.claude/skills/moai-*/`. The Frozen Guard explicitly blocks such writes.
- No telemetry leaves the user's machine. Activity logs are local-only with 90-day retention.
- No auto-evolution of `.moai/project/`, `.moai/specs/`, or design constitution.
- No replacement for explicit user customization via `/moai harness apply <change>` — auto-update is additive, not exclusive.

---

## 2. Scope

### 2.1 In Scope

- Activity signal collection (PostToolUse hook) for `/moai` subcommand invocations, SPEC-ID prefixes, commit message verbs, agent invocations, `/moai feedback` content.
- 4-tier confidence ladder with thresholds 1x / 3x / 5x / 10x.
- Auto-update targets restricted to USER area: `.moai/harness/`, `.claude/skills/my-harness-*/`, `.claude/agents/my-harness/`.
- 5-layer safety architecture: Frozen Guard, Canary Check, Contradiction Detector, Rate Limiter, Human Oversight.
- CLI subcommand `/moai harness` with status / apply / rollback / disable verbs.
- Configuration section under `.moai/config/sections/harness.yaml` `learning:` key.
- Template-First mirror in `internal/template/templates/.moai/config/sections/harness.yaml`.

### 2.2 Out of Scope

- Network telemetry, cross-machine learning aggregation.
- Machine-learning models requiring external compute.
- Auto-evolution of moai-managed assets (blocked by Frozen Guard L1).
- Auto-merge of conflicting customizations without human approval.

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| Solo developer | Reduced manual harness tuning, harness adapts to workflow shifts. |
| Team lead | Per-developer harness customization without polluting shared moai-managed assets. |
| Plan-auditor | Verifiable boundary between USER and MOAI areas (Frozen Guard logs). |
| Security-conscious user | All activity data local, retention bounded, opt-out via config. |

---

## 4. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these is a scope violation:

1. Network upload of activity logs or learning artifacts to external services.
2. Modification of any file under `.claude/agents/moai/`, `.claude/skills/moai-*/`, `.claude/rules/moai/`, or `.moai/project/brand/`.
3. Auto-update of design constitution (`.claude/rules/moai/design/constitution.md`) or any FROZEN zone asset.
4. Cross-project learning (each project's `.moai/harness/` is isolated).
5. Replacement of `/moai feedback` GitHub issue creation flow — feedback is an additional input signal, not a substitute.
6. ML model training requiring GPU or external API.
7. Automatic rollback without explicit user invocation of `/moai harness rollback`.

---

## 5. Requirements (EARS format)

### REQ-HL-001 (Ubiquitous — Activity Observer)

The system **shall** record every `/moai` subcommand invocation, agent invocation, SPEC-ID reference, and `/moai feedback` event to `.moai/harness/usage-log.jsonl` as a single JSONL line containing timestamp, event_type, subject, context_hash, and tier_increment. The observer **shall** be implemented as a PostToolUse hook handler that runs in under 100ms per event and never blocks the parent tool call.

### REQ-HL-002 (Event-Driven — Tier Classifier)

**When** the cumulative observation count for a unique pattern reaches one of {1, 3, 5, 10}, the system **shall** classify the pattern into the corresponding tier (Observation, Heuristic, Rule, Auto-update) and append a classification event to `.moai/harness/learning-history/tier-promotions.jsonl`. The classifier **shall** treat patterns with confidence below 0.70 as observations regardless of count.

### REQ-HL-003 (Event-Driven — Tier 2 Heuristic Enrichment)

**When** a pattern reaches Tier 2 (3 observations), the system **shall** enrich the description field of the matching `.claude/skills/my-harness-*/SKILL.md` frontmatter with a heuristic note (one line, prefixed `# heuristic:`) and **shall not** modify any other frontmatter field or body content at this tier.

### REQ-HL-004 (Event-Driven — Tier 3 Rule Injection)

**When** a pattern reaches Tier 3 (5 observations), the system **shall** inject the corresponding trigger keyword into the `triggers` or `keywords` list of the matching harness skill's frontmatter, deduplicate the list, and write the change with a backup snapshot under `.moai/harness/learning-history/snapshots/<ISO-DATE>/`.

### REQ-HL-005 (Event-Driven — Tier 4 Auto-Update Gate)

**When** a pattern reaches Tier 4 (10 observations), the system **shall** generate a proposed change to `.moai/harness/chaining-rules.yaml` and **shall not** apply the change unless `learning.auto_apply` is `true` AND the user approves via `AskUserQuestion`. The default `auto_apply` value is `false`.

### REQ-HL-006 (Unwanted — Frozen Guard)

**If** any auto-update target path matches `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**`, **then** the system **shall** block the write, log the attempted path to `.moai/harness/learning-history/frozen-guard-violations.jsonl`, and emit a non-blocking warning to stderr. Frozen Guard **shall not** be bypassable by configuration.

### REQ-HL-007 (Event-Driven — Canary Check)

**When** the system prepares to apply a Tier 3 or Tier 4 change, it **shall** shadow-evaluate the proposed change against the most recent 3 sessions in `.moai/harness/usage-log.jsonl` and **shall** reject the change if the projected effectiveness score drops by more than 0.10 versus baseline. Rejection events **shall** be logged with rationale.

### REQ-HL-008 (Event-Driven — Contradiction Detector + Rate Limiter)

**When** a proposed auto-update conflicts with an existing user customization (defined as overlapping trigger keywords or contradictory chaining rules), the system **shall** flag both versions and surface them via `AskUserQuestion` rather than silently overriding. Independently, the system **shall** enforce a maximum of 3 auto-updates per 7-day window with a minimum 24-hour cooldown between updates.

### REQ-HL-009 (Ubiquitous — CLI Subcommand)

The system **shall** provide a `/moai harness` CLI subcommand exposing the following verbs:
- `status` — report current learning state, tier distribution, last-update timestamp, rate-limit window remaining.
- `apply` — manually approve and apply the next pending Tier 4 proposal.
- `rollback <date>` — restore harness artifacts from the snapshot at `.moai/harness/learning-history/snapshots/<date>/`.
- `disable` — set `learning.enabled: false` in `.moai/config/sections/harness.yaml`.

### REQ-HL-010 (Ubiquitous — Configuration & Template-First)

The system **shall** read all learning behavior parameters from `.moai/config/sections/harness.yaml` under the `learning:` key with the following sub-keys: `enabled` (default `true`), `auto_apply` (default `false`), `tier_thresholds` (default `[1,3,5,10]`), `rate_limit` (default `{max_per_week: 3, cooldown_hours: 24}`), `log_retention_days` (default `90`). The default config **shall** be mirrored into `internal/template/templates/.moai/config/sections/harness.yaml` so `moai init` deploys it.

### REQ-HL-011 (State-Driven — Log Retention)

**While** `learning.enabled` is `true`, the system **shall** prune `.moai/harness/usage-log.jsonl` entries older than `log_retention_days` on every observer write and **shall** archive pruned entries to `.moai/harness/learning-history/archive/<YYYY-MM>.jsonl.gz` for one additional retention cycle before final deletion.

---

## 6. Acceptance Coverage Map

| AC ID | Covers REQ-IDs |
|-------|----------------|
| AC-01 | REQ-HL-001 |
| AC-02 | REQ-HL-002, REQ-HL-003 |
| AC-03 | REQ-HL-004 |
| AC-04 | REQ-HL-005, REQ-HL-008 (rate limit + human gate) |
| AC-05 | REQ-HL-006 (Frozen Guard) |
| AC-06 | REQ-HL-007 (Canary), REQ-HL-008 (Contradiction) |
| AC-07 | REQ-HL-009 (CLI) |
| AC-08 | REQ-HL-010, REQ-HL-011 (Config + Retention) |

Coverage: 11 REQs ↔ 8 ACs, 100% (every REQ appears in at least one AC).

---

## 7. Constraints

- [HARD] All instruction files in English (per `coding-standards.md` Language Policy).
- [HARD] No file under `.claude/agents/moai/` or `.claude/skills/moai-*/` may be created, modified, or deleted by this subsystem.
- [HARD] Activity log is local-only — no network egress.
- [HARD] Default `auto_apply: false` — Tier 4 always requires explicit human approval on first deployment.
- Template-First rule: every default config file MUST be mirrored under `internal/template/templates/`.
- Log retention: 90 days default, configurable, never unbounded.

---

## 8. Risks

| Risk | Mitigation |
|------|------------|
| Activity log privacy concerns | Local-only storage, 90-day retention, `learning.enabled: false` opt-out, gitignore enforced. |
| Harness drift from auto-updates | Rate limiter (3/week, 24h cooldown), weekly status report via `/moai harness status`. |
| Boundary violation against moai-managed area | Frozen Guard (L1), violation log, immutable to config bypass. |
| User surprise from auto-applied changes | Default `auto_apply: false`, all Tier 4 changes gated by AskUserQuestion. |
| Rollback complexity | Per-update snapshot in `learning-history/snapshots/<ISO-DATE>/`, single-command `/moai harness rollback <date>`. |
| Canary false-rejection | Configurable threshold (default 0.10), rejection events logged for inspection. |

---

## 9. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| SPEC-V3R3-HARNESS-001 | Hard prerequisite | Generates `.claude/agents/my-harness/*` and `.claude/skills/my-harness-*/*` skeleton this SPEC evolves. |
| SPEC-V3R3-PROJECT-HARNESS-001 | Hard prerequisite | Socratic interview produces baseline customization in `.moai/harness/main.md`. |
| SPEC-DESIGN-CONST-AMEND-001 | Reference | Mirrors design constitution §5 safety architecture into user-harness subsystem. |
| SPEC-AGENCY-ABSORB-001 | Reference | Establishes user-area vs moai-managed boundary precedent. |

---

## 10. Glossary

- **USER area**: `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`, `.moai/harness/` — user-customized harness assets, mutable by this subsystem.
- **MOAI-managed area**: `.claude/agents/moai/`, `.claude/skills/moai-*/`, `.claude/rules/moai/` — upstream-managed assets, IMMUTABLE to this subsystem.
- **Tier**: Confidence classification (Observation 1x, Heuristic 3x, Rule 5x, Auto-update 10x).
- **Frozen Guard**: Layer 1 of safety architecture — path-based write blocker.
- **Canary Check**: Layer 2 — shadow-evaluation against recent sessions.
- **Pattern**: Unique combination of (event_type, subject, context_hash) used as the observation key.
