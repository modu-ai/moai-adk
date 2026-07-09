---
id: SPEC-OBSERVE-HYGIENE-001
title: "Observation Sink Hygiene — Consume-or-Document for Write-Only Logs, Dormant Gates, and Unexercised State"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
era: V3R6
tier: S
related_specs: [SPEC-HARNESS-RATCHET-REWIRE-001, SPEC-TOKEN-BUDGET-STOP-001]
tags: "observation-sinks, log-hygiene, pruning, sync-gate, dormancy-annotation, write-only-logs, workflow-reflex"
---

# SPEC-OBSERVE-HYGIENE-001 — Observation Sink Hygiene

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **6 of 6**.

- **Dependency notes**: No blocking dependency. Boundary contracts with two siblings: harness learner internals are owned by SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6); `.moai/state/verify/` persistence is owned by SPEC-TOKEN-BUDGET-STOP-001 (plan committed 93c38003b) — both cross-referenced, neither duplicated.
- **Tier**: S (minimal envelope) — see plan.md §A.4 for evidence (file count exceeds the nominal band only via template mirrors + test files; LOC stays inside the S band).
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| H4 | MED | Write-only observation sinks with no reader: status-transition-audit.log (~389KB), fact-force-skip.log (~73KB, growing), task-metrics.jsonl (stale since 2026-05-30), hundreds of zero-byte trace-*.jsonl files |
| H5 | MED | status-transition-ownership.sh writes audit lines but its promised exit-2 enforcement is "reserved for future" — neither teeth nor a consumer |
| H6 | LOW | sync-phase-quality-gate.sh blocking arm dormant behind MOAI_SYNC_GATE_BLOCKING=1, although go vet/go build are fast, deterministic, once-per-commit sentinel-guarded — the lowest-risk real block candidate |
| L7 | LOW | Doctrine-only / unexercised surfaces: loop-snapshots dir never materialized while loop.md promises snapshot-based resume; sunset.yaml dormant with zero in-file annotation; harness.yaml model_upgrade_review trigger env never set |

---

## User Story

**As a** maintainer of the MoAI observation layer,
**I want** every observation sink to be exactly one of (a) consumed by a mechanical reader, (b) documented as intentionally write-only, or (c) retired/pruned,
**so that** the logs directory stops accumulating unread promises — an audit log someone will trust because `moai spec audit` actually reads it, telemetry that rotates instead of fossilizing, a sync gate whose cheapest deterministic checks actually block, and config blocks that say out loud that they are dormant.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read (byte sizes and file counts are same-day measurements and drift with use). Line numbers indicative; content anchors are authoritative.

### GAP-1 — Write-only sinks with no reader (H4)

- **Measured source**: `ls -la .moai/logs/` → `status-transition-audit.log` 397,935 bytes (appended by `.claude/hooks/moai/status-transition-ownership.sh`, observed append at line ~79); `fact-force-skip.log` 75,128 bytes (writer `.claude/hooks/moai/gateguard-fact-force.sh`, advisory per its own header); `task-metrics.jsonl` 467,600 bytes, mtime 2026-05-30 12:56 (writer `internal/hook/post_tool.go` Agent/Task subagent-metrics path, SPEC-MONITOR-001 comment near line 166); trace files: 643 `trace-*.jsonl`, of which 314 are zero-byte (writer `internal/hook/trace/writer.go` — per-session file, size-based rotation only per its REQ-OBS-006 comment; no age-based pruning anywhere).
- **Observed pattern**: `grep -rn "status-transition-audit" --include='*.go' internal/ cmd/` → 0 (no Go consumer; the only reference is doc prose in agent-common-protocol.md). No consumer found for fact-force-skip.log or task-metrics.jsonl either. Nothing prunes zero-byte traces or ages out telemetry. Four sinks, zero readers, unbounded growth or silent staleness.

### GAP-2 — Audit hook has neither teeth nor a consumer (H5)

- **Measured source**: `.claude/hooks/moai/status-transition-ownership.sh` line ~68: "Advisory hook (never blocks; exit 2 reserved for future ownership-mismatch enforcement)".
- **Observed pattern**: The hook diligently logs every SPEC-artifact Write/Edit transition site, the promised enforcement is deferred, and no tool reads the log — the observation is pure cost today.

### GAP-3 — Cheapest deterministic gate checks are dormant (H6)

- **Measured source**: `.claude/hooks/moai/sync-phase-quality-gate.sh` — blocking mode opt-in at line ~325 (`MOAI_SYNC_GATE_BLOCKING:-0`); go steps at lines ~218-220 (`go vet ./...`, `go build ./...`); the hook's own header notes advisory stdout-JSON vs blocking semantics.
- **Observed pattern**: `go vet`/`go build` are fast, deterministic, and the hook is once-per-commit sentinel-guarded — the lowest-risk candidates for a real default block — yet the blocking arm requires an env var nobody sets.

### GAP-4 — Unexercised state and unannotated dormancy (L7)

- **Measured source**: `ls .moai/cache/loop-snapshots` → No such file or directory, while `.claude/skills/moai/workflows/loop.md` references the snapshot dir and `--resume` re-entrancy at 8+ places (observed lines 5, 60, 75, 82, 150, 198, 201, 257); `.moai/config/sections/sunset.yaml` → typed `SunsetConfig` exists but DORMANT — "struct defined but no runtime hot path (REQ-MIG003-006)" per `internal/config/audit_loader_completeness_test.go:23`, with a Go-side once-per-session DORMANT notice (REQ-MIG003-018, `internal/config/loader.go:97`) — and the YAML file itself carries ZERO comments; `.moai/config/sections/harness.yaml` `model_upgrade_review:` (line ~77) — a consumer EXISTS (`emitModelUpgradeReminder`, `internal/cli/harness_validate.go`, REQ-HRN-001-016) but it triggers only on `CLAUDE_MODEL_PREVIOUS`/`CLAUDE_MODEL` env vars for which `grep -rn "CLAUDE_MODEL_PREVIOUS"` finds no setter anywhere → dormant-in-practice at the trigger layer. (Correction vs the audit brief: the brief said "no mechanical consumer"; measurement shows consumer code exists and the dormancy sits in the never-set trigger env.)
- **Observed pattern**: loop.md promises snapshot-based resume that has never once produced a file; two config blocks are dormant with no in-file annotation telling a reader so.

### Aggregate defect claim

**The observation layer writes promises nobody reads: four reader-less sinks, an audit log with deferred teeth, a dormant deterministic gate, and unannotated dormant config.** This SPEC applies consume-or-document-or-retire per sink: one new consumer (spec-audit cross-check), one pruning/retention policy, two write-only documentations, one gate-promotion decision, one best-effort doc marker, and two dormancy annotations.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the spec audit engine", "the session-end hook path", "the sync gate", "the config file"). No legacy `IF/THEN` modality. Per-sink disposition is exactly ONE of: (a) consumer, (b) documented write-only, (c) retire/prune.

### REQ-OBH-001 — Event-driven (When) — status-transition-audit.log consumer [(a) recommended]

**When** `moai spec audit` runs and `.moai/logs/status-transition-audit.log` is present, the spec audit engine SHALL consume the log as a cross-check input — correlating logged transition sites against the Status Transition Ownership Matrix owners and emitting INFO-severity findings for ownership mismatches (absent or unparseable log degrades gracefully to no findings, never an error). Fallback direction (b) — documenting the log as intentionally write-only — is presented in plan.md §D D1; the plan recommends (a).

### REQ-OBH-002 — Event-driven (When) — telemetry pruning and retention [(c)]

**When** a session ends (the SessionEnd hook path in `internal/hook`), the session-end hook path SHALL prune zero-byte `trace-*.jsonl` files under `.moai/logs/` and apply a documented age-based retention policy to non-empty `trace-*.jsonl` files (threshold per plan.md §D D2); the `task-metrics.jsonl` staleness SHALL be dispositioned in the same milestone — root-cause note plus either writer repair, documented write-only status, or retirement (plan.md §D D2 presents the directions).

### REQ-OBH-003 — Ubiquitous — fact-force-skip.log documented write-only [(b)]

The fact-force skip log SHALL be documented as intentionally write-only (a local audit trail with no mechanical consumer, pruned at the operator's discretion) — a one-line statement in the `gateguard-fact-force.sh` header and in the log's referencing doctrine, closing the "is something supposed to read this?" ambiguity at the lowest cost.

### REQ-OBH-004 — Capability gate (Where) — sync-gate deterministic-check promotion decision

**Where** the sync-phase quality gate's language detection resolves to the Go toolchain (the `go vet ./...` + `go build ./...` steps), the sync gate SHALL apply the Kickoff-decided blocking default for exactly those two deterministic checks — plan.md §D D3 presents both directions (promote to default-blocking vs keep opt-in) with a recommendation to promote, given determinism, speed, once-per-commit sentinel guarding, and the preserved anti-death-spiral recovery carve-out (runtime-recovery-doctrine.md §4 SHOULD guidance unchanged); test/coverage checks remain advisory regardless of the decision.

### REQ-OBH-005 — Ubiquitous — loop-snapshots best-effort marker [(b)]

The loop workflow doc (loop.md snapshot/resume sections) SHALL carry a best-effort marker stating that snapshot persistence and `--resume` re-entrancy are best-effort orchestrator behaviors with no mechanical writer guarantee (the snapshot dir has never materialized) — documentation-only in this SPEC; mechanical persistence, if ever built, belongs to the loop-verdict persistence line (SPEC-LOOP-VERDICT-CONTRACT-001 follow-ups), not here.

### REQ-OBH-006 — Ubiquitous — dormancy annotations in config [(b)]

The config files SHALL carry one-line dormancy annotations at the block level: `sunset.yaml` (dormant — typed struct loaded, no runtime hot path; Go-side DORMANT notice exists) and `harness.yaml` `model_upgrade_review` (dormant-in-practice — consumer exists but its `CLAUDE_MODEL_PREVIOUS`/`CLAUDE_MODEL` trigger env vars have no setter) — live + template mirrors, comment-only edits with zero semantic change.

---

## Constraints

1. **Sibling scope boundaries (HARD)** — `.moai/state/verify/` persistence is SPEC-TOKEN-BUDGET-STOP-001's scope (cross-ref only); harness learner internals (usage-log, classifier, proposals) are SPEC-HARNESS-RATCHET-REWIRE-001's scope. This SPEC touches neither.
2. **Hook contract preservation (HARD)** — Claude Code hook exit-code/stdout-JSON semantics (block decisions ride exit-0 stdout) and the runtime-recovery §4 carve-out are preserved; REQ-OBH-004 changes a default, not the signaling mechanism. status-transition-ownership.sh's reserved exit-2 enforcement stays reserved (activation is out of scope).
3. **Comment-only config edits** — REQ-OBH-006 changes zero YAML semantics; loaders and tests must be bit-identical in behavior.
4. **Prefer document-or-prune** — per the Tier S lean directive, new consumers are built only where mandated (REQ-OBH-001); every other sink takes the (b)/(c) route.
5. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — .moai/state/verify/ implementation

- The verification-evidence persistence directory and its writer — owned by SPEC-TOKEN-BUDGET-STOP-001 (plan committed 93c38003b). This SPEC only cross-references it as a doctrine-committed, not-yet-materialized surface; no duplication.

### Out of Scope — harness learner internals

- `.moai/harness/usage-log.jsonl`, tier classifier, proposals pipeline, lessons inbox — owned by SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6).

### Out of Scope — status-transition exit-2 enforcement activation

- Turning the "reserved for future" blocking arm of status-transition-ownership.sh into live exit-2/stdout-block enforcement. REQ-OBH-001 gives the log a READER; giving the hook TEETH is a separate risk decision for a follow-up SPEC.

### Out of Scope — mechanical loop-snapshot persistence

- Building a Go/hook writer for `.moai/cache/loop-snapshots/`. REQ-OBH-005 is documentation-only; mechanical persistence belongs to the loop-verdict line if anywhere.

### Out of Scope — recovery-signal carve-out mechanization

- Parsing `stopReason` in hooks to mechanically enforce runtime-recovery-doctrine.md §4 — explicitly deferred by that doctrine to a future runtime-layer SPEC; this SPEC only preserves the carve-out.

### Out of Scope — new consumers beyond the spec-audit cross-check

- Dashboards, log analyzers, or Go readers for fact-force-skip.log, task-metrics.jsonl, or trace files. Document-or-prune is the chosen posture for all sinks except REQ-OBH-001's.

---

## Cross-References

- **EXTEND base (Go)**: `internal/hook` session-end path (pruning, REQ-OBH-002); `internal/spec` audit engine + `internal/cli/spec_audit.go` surface (REQ-OBH-001 consumer).
- **EXTEND base (hook/sh)**: `.claude/hooks/moai/gateguard-fact-force.sh` (header doc line); `.claude/hooks/moai/sync-phase-quality-gate.sh` (blocking-default per D3). Both mirrored in templates (verified 2026-07-09).
- **EXTEND base (config/doc)**: `.moai/config/sections/sunset.yaml` + `harness.yaml` (annotations; mirrors verified); `.claude/skills/moai/workflows/loop.md` snapshot sections (best-effort marker; mirror verified; sibling SPEC-LOOP-VERDICT-CONTRACT-001 shares this file — sequencing per plan.md §B).
- **Referenced (unmodified)**: `.claude/hooks/moai/status-transition-ownership.sh` (log producer; reserved enforcement stays reserved); `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (the correlation target for REQ-OBH-001); `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 (preserved carve-out); `internal/hook/trace/writer.go` (existing size rotation, REQ-OBS-006); `internal/cli/harness_validate.go` `emitModelUpgradeReminder` (the dormant-in-practice consumer named by REQ-OBH-006).
- **Epic**: Workflow-Reflex 6 of 6. Siblings: SPEC-HARNESS-RATCHET-REWIRE-001 (1), SPEC-MODEL-ROUTING-WIRE-001 (2), SPEC-LOOP-VERDICT-CONTRACT-001 (3), SPEC-ADVISOR-RUNG-001 (4), SPEC-CADENCE-BRIDGE-001 (5).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 6 of 6. Consume-or-document-or-retire across 6 sink groups; spec-audit consumer + telemetry pruning + gate-promotion decision + dormancy annotations. Tier S. Corrects the audit brief's model_upgrade_review claim (consumer exists; trigger env dormant). |
