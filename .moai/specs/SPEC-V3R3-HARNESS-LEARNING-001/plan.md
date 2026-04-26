# SPEC-V3R3-HARNESS-LEARNING-001 — Implementation Plan

## 1. Overview

This plan delivers the user-area dynamic harness self-learning subsystem in five phases. Each phase is priority-ordered (no time estimates per `agent-common-protocol.md` Time Estimation rule). Phases are gated by dependency completion (HARNESS-001 + PROJECT-HARNESS-001) and by inter-phase artifact handoff.

## 2. Architectural Approach

### 2.1 Component Map

```
PostToolUse hook
   └─> internal/harness/observer.go            (Phase 1)
        └─> writes .moai/harness/usage-log.jsonl
                └─> internal/harness/learner.go    (Phase 2)
                     └─> tier classification
                          └─> internal/harness/safety.go      (Phase 3)
                               ├─ L1 Frozen Guard
                               ├─ L2 Canary Check
                               ├─ L3 Contradiction Detector
                               ├─ L4 Rate Limiter
                               └─ L5 Human Oversight (orchestrator)
                                    └─> internal/harness/applier.go    (Phase 4)
                                         └─> writes .moai/harness/, .claude/skills/my-harness-*/, .claude/agents/my-harness/

CLI: /moai harness {status|apply|rollback|disable}    (Phase 4)
Coordinator skill: .claude/skills/moai-harness-learner/SKILL.md   (Phase 4)
```

### 2.2 Data Flow

1. User invokes `/moai <subcommand>` or any `Agent()` call.
2. PostToolUse hook fires → `internal/harness/observer.go` writes one JSONL line.
3. Learner runs on a debounced schedule (every 10 events or 5 minutes, whichever first) → classifies patterns into tiers.
4. Tier 2/3/4 promotions trigger safety evaluation in order: L1 → L2 → L3 → L4 → (Tier 4 only) L5.
5. Approved changes pass through `applier.go` which writes to USER area only and creates a snapshot in `learning-history/snapshots/<ISO-DATE>/`.

### 2.3 Boundary Enforcement

The Frozen Guard (L1) is implemented as a path-prefix matcher in `internal/harness/safety.go`. The matcher is the **first** check in every write path and cannot be disabled by configuration. Test coverage (Phase 5 IT-05) verifies that bypass attempts are rejected.

## 3. Phased Implementation

### Phase 1 — Observer + Log Schema (Priority: P1 High)

**Goal**: Activity collection without behavior changes. No tier classification, no auto-updates.

**Deliverables**:
- `internal/harness/observer.go` — PostToolUse-driven event collector.
- `.moai/harness/usage-log.jsonl` schema definition (see §4).
- `.moai/harness/learning-history/` directory scaffolding.
- Hook wrapper `.claude/hooks/moai/handle-harness-observe.sh`.
- Unit tests: observer write performance (<100ms), JSONL line validity, log retention pruning.

**REQ-IDs covered**: REQ-HL-001, REQ-HL-011 (retention).

**Exit gate**: Observer logs at least 100 real events from a developer session without measurable latency impact (verified via `time` wrapper).

### Phase 2 — Tier Classifier (Priority: P1 High)

**Goal**: Classify patterns into 4 tiers; emit promotion events; no side effects on harness files yet.

**Deliverables**:
- `internal/harness/learner.go` — pattern aggregator + tier classifier.
- `internal/harness/types.go` — `Pattern`, `TierEvent`, `Promotion` data types.
- `.moai/harness/learning-history/tier-promotions.jsonl` writer.
- Unit tests: confidence threshold (0.70), tier boundary correctness, idempotency under repeated runs.

**REQ-IDs covered**: REQ-HL-002, REQ-HL-003 (description-only enrichment), REQ-HL-004 (frontmatter trigger injection — deferred-write mode).

**Exit gate**: Replay of 1,000 synthetic events produces correct tier distribution and no false promotions.

### Phase 3 — 5-Layer Safety Architecture (Priority: P0 Critical)

**Goal**: Implement all five safety layers as composable middleware around the applier.

**Deliverables**:
- `internal/harness/safety/frozen_guard.go` — L1 path matcher with hardcoded MOAI-managed prefixes (no config override).
- `internal/harness/safety/canary.go` — L2 shadow-evaluation engine (effectiveness score delta vs baseline).
- `internal/harness/safety/contradiction.go` — L3 trigger-keyword overlap detector + chaining-rules conflict detector.
- `internal/harness/safety/rate_limit.go` — L4 sliding-window rate limiter (3/week, 24h cooldown).
- `internal/harness/safety/oversight.go` — L5 AskUserQuestion bridge (returns proposal payload to orchestrator skill).
- Frozen Guard violation log writer (`.moai/harness/learning-history/frozen-guard-violations.jsonl`).
- Unit tests for each layer + integration test composing all five.

**REQ-IDs covered**: REQ-HL-006, REQ-HL-007, REQ-HL-008.

**Exit gate**: All safety layer unit tests pass; integration test validates correct ordering (L1 → L2 → L3 → L4 → L5).

### Phase 4 — Auto-Update Applier + CLI + Coordinator Skill (Priority: P1 High)

**Goal**: Wire safety-approved changes to actual file writes; expose CLI; deliver coordinator skill.

**Deliverables**:
- `internal/harness/applier.go` — atomic file writer with snapshot creation.
- `internal/cli/harness.go` — `/moai harness {status, apply, rollback <date>, disable}` subcommand handler.
- `.claude/skills/moai-harness-learner/SKILL.md` — coordinator skill that triggers learner runs and surfaces Tier 4 proposals to the orchestrator.
- `.moai/config/sections/harness.yaml` — `learning:` section with documented defaults.
- `internal/template/templates/.moai/config/sections/harness.yaml` — Template-First mirror.
- CLI integration tests (status output formatting, apply flow, rollback restoration, disable persistence).

**REQ-IDs covered**: REQ-HL-005, REQ-HL-009, REQ-HL-010.

**Exit gate**: `/moai harness status` returns valid output on a fresh project; `/moai harness rollback <date>` correctly restores a snapshot; CLI integration tests pass on all three OS targets (macOS, Linux, Windows).

### Phase 5 — Integration Tests + Documentation (Priority: P1 High)

**Goal**: End-to-end validation across the full pipeline; user-facing documentation.

**Deliverables**:
- IT-01: 100-event session replay → expected tier distribution.
- IT-02: Tier 3 promotion → frontmatter modification with snapshot.
- IT-03: Tier 4 promotion → AskUserQuestion gate → applier write on approval.
- IT-04: Frozen Guard rejection of `.claude/skills/moai-foo/SKILL.md` write attempt.
- IT-05: Rate limiter blocks 4th update within 7 days.
- IT-06: Rollback restores prior state byte-identical.
- IT-07: `learning.enabled: false` disables observer and applier.
- README section in `.moai/harness/README.md` documenting CLI verbs and config keys.

**REQ-IDs covered**: All REQ-HL-001 through REQ-HL-011 (cross-cutting validation).

**Exit gate**: All integration tests pass on CI for ubuntu-latest, macos-latest, windows-latest. Documentation reviewed.

## 4. Data Schemas

### 4.1 `.moai/harness/usage-log.jsonl`

One JSON object per line:

```json
{
  "ts": "2026-04-26T19:30:45Z",
  "event_type": "subcommand_invocation | agent_invocation | spec_id_reference | feedback",
  "subject": "/moai plan | manager-spec | SPEC-V3R3-HARNESS-LEARNING-001 | ...",
  "context_hash": "sha256:abcdef...",
  "tier_increment": 1,
  "session_id": "uuid-v7"
}
```

### 4.2 `.moai/harness/learning-history/tier-promotions.jsonl`

```json
{
  "ts": "2026-04-26T19:35:00Z",
  "pattern_key": "subcommand:/moai plan",
  "from_tier": "heuristic",
  "to_tier": "rule",
  "observation_count": 5,
  "confidence": 0.82
}
```

### 4.3 `.moai/config/sections/harness.yaml` `learning:` section

```yaml
learning:
  enabled: true
  auto_apply: false
  tier_thresholds: [1, 3, 5, 10]
  rate_limit:
    max_per_week: 3
    cooldown_hours: 24
  log_retention_days: 90
  canary:
    score_drop_threshold: 0.10
    sessions_to_evaluate: 3
```

## 5. Technical Approach Decisions

### 5.1 Why JSONL for usage log

- Append-only, no parser state required.
- Compatible with `tail`, `grep`, `jq` for ad-hoc inspection.
- Trivial to prune (line-based) and archive (gzip).

### 5.2 Why path-prefix matcher for Frozen Guard

- Single source of truth: a hardcoded `[]string{".claude/agents/moai/", ".claude/skills/moai-", ".claude/rules/moai/", ".moai/project/brand/"}` slice.
- No config injection point — cannot be bypassed by user mistake or malicious config.
- Test surface is small: enumerate all known violation patterns + golden-path USER area writes.

### 5.3 Why Tier 4 default `auto_apply: false`

- Aligns with design constitution §5 Layer 5 (Human Oversight).
- First-deployment safety: user observes 3 weeks of proposals before opting into automation.
- Documented escape hatch via `/moai harness apply` for one-shot acceptance.

### 5.4 Why snapshot per update (vs git-based rollback)

- User may not have committed harness changes; git rollback would be ambiguous.
- Snapshot in `.moai/harness/learning-history/snapshots/<ISO-DATE>/` is self-contained and orthogonal to git.
- Rollback is a directory-copy operation, not a git operation — no risk of touching unrelated files.

## 6. Dependencies on Other SPECs

This plan **cannot start Phase 1** until the following two SPECs are merged:

1. **SPEC-V3R3-HARNESS-001** — Provides the meta-harness skill that generates `.claude/agents/my-harness/*` and `.claude/skills/my-harness-*/*`. Without these target files, the applier (Phase 4) has nothing to write to.
2. **SPEC-V3R3-PROJECT-HARNESS-001** — Conducts the Socratic interview producing baseline customization in `.moai/harness/main.md`. Without the baseline, the learner (Phase 2) has no anchor to evolve from.

Phase 5 integration tests (IT-04 specifically) reference paths defined by HARNESS-001's skeleton generator.

## 7. Risks (Plan-Level)

| Risk | Severity | Mitigation |
|------|----------|------------|
| Phase 1 hook adds measurable latency | Medium | Performance budget <100ms per event; integration test enforces budget. |
| Phase 3 Canary false-rejects valid changes | Medium | Configurable threshold (default 0.10); rejection events logged with rationale for inspection. |
| Phase 4 CLI subcommand collides with future `/moai` verb | Low | Reserve `harness` verb in `internal/cli/router.go` registration table. |
| Cross-platform JSONL line-ending differences | Low | Always emit `\n` (LF), never `\r\n`; verified by IT-01 on Windows runner. |
| HARNESS-001 / PROJECT-HARNESS-001 slip past v2.17 | High | This SPEC is target_release v2.17.0; if dependencies slip, defer entire SPEC to v2.18 — do NOT partially implement. |

## 8. Verification Strategy

- Unit coverage target: 85% per `quality.yaml` standard.
- Integration test count: ≥ 7 (IT-01 through IT-07 above).
- Cross-platform: ubuntu-latest, macos-latest, windows-latest in CI.
- Manual verification: solo developer runs harness for 1 week, reviews `/moai harness status` output, confirms no Frozen Guard violations.

## 9. Rollout

- Default config ships with `learning.enabled: true` and `learning.auto_apply: false` — observation begins immediately, but no auto-update occurs without explicit user opt-in.
- v2.17.0 release notes MUST highlight the new `/moai harness` subcommand and the opt-in nature of auto-apply.
- Post-release: `manager-docs` adds harness learning section to docs-site (per §17 4-locale rule).
