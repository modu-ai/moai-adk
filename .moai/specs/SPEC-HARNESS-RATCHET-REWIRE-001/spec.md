---
id: SPEC-HARNESS-RATCHET-REWIRE-001
title: "Wire Failure Signals into the Harness Learning Loop (FAILURE→RATCHET Repair)"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "harness, learning-loop, failure-signals, classifier-eligibility, proposals, lessons-inbox, doctor, workflow-reflex"
---

# SPEC-HARNESS-RATCHET-REWIRE-001 — Wire Failure Signals into the Harness Learning Loop

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **1 of 6**.

- **Dependency notes**: SPEC 1 (this) and SPEC 2 (SPEC-MODEL-ROUTING-WIRE-001) are independent of each other. SPEC 3 (SPEC-LOOP-VERDICT-CONTRACT-001) is independent. Downstream SPEC-ADVISOR-RUNG-001 depends on SPEC 2, NOT on this SPEC.
- **Tier**: M (standard) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| H1 | HIGH | FAILURE signals (PostToolUseFailure, evidence-writer test pass/fail) never enter the learning loop — no classifier/proposal path consumes them |
| H2 | HIGH | Tier classifier promotes degenerate lifecycle-noise keys (`session_stop::`, `user_prompt::`, `subagent_stop:unknown:` with empty context_hash, confidence 1) |
| L2 | HIGH | Back half of the pipeline has ZERO lifetime throughput — `.moai/harness/proposals/` never existed, `.moai/evolution/learnings/` empty, `learnings_count: 0`; proposal generation is manual-trigger only; the only working ratchet today is fully manual (human-authored lessons memory + doctrine rule files) |

---

## User Story

**As a** MoAI harness learning loop that already observes lifecycle events into `.moai/harness/usage-log.jsonl`,
**I want** failure signals (tool failures, test failures) recorded as first-class learnable events, degenerate noise keys excluded from tier promotion, and proposal generation auto-chained at Stop,
**so that** the FAILURE→RATCHET half of the learning loop actually turns — promoted patterns land in `.moai/harness/proposals/` for the human apply gate, and failure lessons reach a repo-local inbox the orchestrator's Lessons Protocol can drain — instead of the entire ratchet remaining a manual human process.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

Per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3, each gap below names the measured source and the observed pattern (measured 2026-07-09 by this agent via Bash/Read).

### GAP-1 — Failure signals never enter the learning loop (H1)

- **Measured source**: `internal/hook/post_tool_failure.go` (active handler per `internal/hook/coverage_table.go:46` — `{EventName: "PostToolUseFailure", ..., IsActive: true, HandlerFile: "post_tool_failure.go"}`); `internal/hook/evidence_writer.go` (15,431 bytes — test pass/fail capture).
- **Observed pattern**: The observer records lifecycle events (`PostToolUse` / `Stop` / `SubagentStop` / `UserPromptSubmit`) into `.moai/harness/usage-log.jsonl` (live: 15,763 lines / ~3.25 MB, mtime 2026-07-09). But no `tool_failure:*` or `test_fail:*` event key exists in the log; the PostToolUseFailure handler and evidence-writer test-fail detection feed NO classifier/proposal path. Failures — the highest-information learning signal — are invisible to the ratchet.

### GAP-2 — Classifier promotes degenerate lifecycle-noise keys (H2)

- **Measured source**: `.moai/harness/learning-history/tier-promotions.jsonl` (16 entries, all ts 2026-05-24); `internal/cli/hook.go` Stop path (content anchor: `classifyHarnessPatterns(root)` chained after `RecordExtendedEvent`, comment `SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-003: auto-classify on the Stop path`, observed at lines 765-776).
- **Observed pattern**: Promotion entries include `{"pattern_key":"subagent_stop:unknown:","from_tier":"","to_tier":"auto_update","observation_count":41,"confidence":1}` and `{"pattern_key":"session_stop::","to_tier":"auto_update","observation_count":23,"confidence":1}`. Keys with empty context_hash and empty/unknown subject are pure lifecycle noise (every session emits them), yet they are promoted with confidence 1 — they crowd the promotion history with unlearnable patterns.

### GAP-3 — Back half of the pipeline is starved: zero lifetime throughput (L2)

- **Measured source**: `ls .moai/harness/proposals` → `No such file or directory` (never existed); `ls .moai/evolution/learnings` → only `.gitkeep`; `.moai/evolution/manifest.yaml` → `learnings_count: 0`, `proposals_this_week: 0`; `ls .moai/evolution/snapshots` → absent. Proposal generation is `moai harness propose` (`internal/cli/harness/propose.go`) — manual-trigger only.
- **Observed pattern**: The applier/safety-pipeline Go code (applier, rubric, regression gate — audit-reported ~35 KB, well-tested) has never received a single proposal. The classify half was auto-wired at Stop by SPEC-HARNESS-EVO-PIPE-REPAIR-001; the propose half was not. Result: promotions accumulate, proposals never materialize, and the only working ratchet is fully manual.

### Aggregate defect claim

**The harness learning loop observes but never learns from failure, promotes noise, and never generates proposals.** The classify→promote half turns; the promote→propose→apply half is starved at its first link. This SPEC records failure events, hardens promotion eligibility, auto-chains propose at Stop (preserving the human apply gate), adds a repo-local lessons inbox, and gives `moai harness doctor` a dormancy check so this starvation is detectable mechanically.

---

## Requirements (GEARS notation)

> **Subject convention**: GEARS generalized subjects are used ("the harness observer", "the tier classifier", "the CLI", "the failure handler", "the doctor command"). No legacy `IF/THEN` modality.

### REQ-HRR-001 — Event-driven (When) — tool-failure event recording

**When** the `internal/hook` PostToolUseFailure handler processes a tool-failure event, the harness observer SHALL append a failure event to `.moai/harness/usage-log.jsonl` with the dedicated event-key form `tool_failure:<tool>:<signature>` (signature = a stable, low-cardinality failure classifier such as an error-class token or truncated error hash — exact derivation is a run-phase design decision, plan.md §D).

### REQ-HRR-002 — Event-driven (When) — test-failure event recording

**When** the evidence writer (`internal/hook/evidence_writer.go`) detects a test failure, the harness observer SHALL append a failure event to `.moai/harness/usage-log.jsonl` with the dedicated event-key form `test_fail:<package>:`.

### REQ-HRR-003 — Unwanted behavior — degenerate-key promotion exclusion

The tier classifier SHALL NOT promote pattern keys that carry an empty context hash AND an empty or `unknown` subject (the degenerate lifecycle-noise class exemplified by `session_stop::`, `user_prompt::`, and `subagent_stop:unknown:`). A regression test SHALL prove that `subagent_stop:unknown:` no longer promotes.

### REQ-HRR-004 — Event-driven (When) — propose auto-run at Stop

**When** the Stop-path auto-classify (`classifyHarnessPatterns`, the SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-003 chain in `internal/cli/hook.go`) completes with one or more promotions, the CLI SHALL auto-run proposal generation (the same generation path as `moai harness propose`) so that promoted patterns land in `.moai/harness/proposals/`. The chain SHALL be fail-open (propose errors are logged to stderr and NEVER block session end) and SHALL stay within the 5-second hook budget, mirroring the classify chain's existing AP-3 fail-open discipline.

### REQ-HRR-005 — Unwanted behavior — human apply gate preservation

The auto-chained pipeline SHALL NOT auto-apply proposals: `learning.auto_apply` remains `false` (`.moai/config/sections/harness.yaml`, observed at line 116), the Stop chain terminates at proposal generation, and applying a proposal remains gated behind the orchestrator's `AskUserQuestion` round + explicit `moai harness apply` invocation.

### REQ-HRR-006 — Event-driven (When) — lessons-inbox stub emission

**When** a failure event is recorded per REQ-HRR-001 or REQ-HRR-002, the failure handler SHALL append a structured lesson stub to the repo-local lessons inbox `.moai/lessons-inbox.jsonl` (JSONL; minimum fields: timestamp, event key, failure summary, source file/tool — exact schema is a run-phase design decision, plan.md §D).

### REQ-HRR-007 — Ubiquitous — Lessons Protocol drain cross-reference (doc deliverable)

The Lessons Protocol doctrine (`.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol) SHALL cross-reference the lessons inbox: the orchestrator drains `.moai/lessons-inbox.jsonl` stubs into auto-memory lesson entries as part of the Lessons Protocol, marking drained stubs (drain marking mechanism is a run-phase design decision).

### REQ-HRR-008 — Event-driven (When) — doctor dormancy check

**When** `moai harness doctor` (`internal/cli/harness/doctor.go`) runs while `.moai/harness/learning-history/tier-promotions.jsonl` contains one or more promotions AND `.moai/harness/proposals/` is absent, the doctor command SHALL emit a pipeline-dormancy warning finding ("promotions exist but proposals dir absent") in its report output.

### REQ-HRR-009 — Capability gate (Where) — template-first boundary

**Where** an edited surface has a template mirror under `internal/template/templates/` (verified present: `.claude/rules/moai/core/moai-constitution.md`), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees. Go code under `internal/` is dev/runtime tooling and is NOT templated.

---

## Constraints

1. **Fail-open hook discipline (HARD)** — the Stop-path propose chain MUST mirror the existing classify chain's fail-open behavior (errors logged to stderr, exit 0, never block session end) and MUST stay within the 5-second hook budget. REQ-HRR-004 binds.
2. **Human apply gate (HARD)** — `learning.auto_apply` stays `false`; no code path introduced by this SPEC invokes apply. REQ-HRR-005 binds.
3. **Subagent boundary (C-HRA-008)** — no `AskUserQuestion` / `mcp__askuser` call sites in `internal/harness/`, `internal/hook/`, `internal/cli/` code touched by this SPEC (grep 0 matches excluding tests/comments).
4. **Runtime-managed file hygiene** — `.moai/harness/usage-log.jsonl`, `.moai/harness/learning-history/*`, `.moai/state/*` are runtime-managed; tests operate in `t.TempDir()` only, never against the live project log.
5. **Frozen surfaces untouched** — frozen-guard sentinels and the 5-layer safety pipeline semantics (applier / rubric / regression gate) are NOT modified; this SPEC only feeds them input.
6. **GEARS notation; era V3R6; 12 canonical frontmatter fields** (created/updated/tags — no snake_case aliases).

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — auto_apply default change

- Changing `learning.auto_apply` from `false` to `true`, or any mechanism that applies proposals without the human gate. The apply gate is explicitly PRESERVED (REQ-HRR-005); this SPEC ends at proposal generation.

### Out of Scope — 5-layer safety pipeline semantics

- Modifying the applier, rubric scoring, or regression-gate logic (the audit-reported ~35 KB well-tested safety pipeline). This SPEC feeds proposals INTO the pipeline; it does not change how the pipeline evaluates or applies them.

### Out of Scope — frozen-guard sentinels

- Any change to FROZEN sentinels or `LoadHarnessConfig` FROZEN validation (HRN-001). Config edits are limited to none — `auto_apply: false` is asserted, not changed.

### Out of Scope — manual ratchet replacement

- Retiring or restructuring the existing manual ratchet (human-authored lessons memory + doctrine rule files). The lessons inbox is an ADDITIVE feed into the existing Lessons Protocol, not a replacement for human judgment.

### Out of Scope — proposal content quality / prompt engineering

- Improving what a generated proposal says (templates, rubric wording, prompt content). Only the TRIGGER (auto-run at Stop) and the INPUT signals (failure events, eligibility hardening) are in scope.

### Out of Scope — historical promotion cleanup

- Rewriting or pruning the 16 existing degenerate entries in `tier-promotions.jsonl`. The eligibility hardening prevents FUTURE degenerate promotions; the history file is an append-only runtime artifact left as-is.

---

## Cross-References

- **EXTEND base (Go)**: `internal/cli/hook.go` Stop path (`classifyHarnessPatterns` chain — REQ-HEP-003 provenance); `internal/hook/post_tool_failure.go` (active PostToolUseFailure handler); `internal/hook/evidence_writer.go` (test pass/fail capture); `internal/cli/harness/propose.go` (`moai harness propose` — generation path to reuse); `internal/cli/harness/doctor.go` (doctor check registry).
- **PRESERVE (semantics untouched)**: applier / rubric / regression-gate safety pipeline; `learning.auto_apply: false` (`.moai/config/sections/harness.yaml:116`); frozen-guard sentinels.
- **Doc deliverable target**: `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol (+ template mirror `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`).
- **Provenance SPECs**: SPEC-HARNESS-EVO-PIPE-REPAIR-001 (classify auto-wired at Stop — the pattern REQ-HRR-004 mirrors); SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (propose command origin); SPEC-HARNESS-LOOP-CLOSURE-001 (learning-loop Go engine).
- **Doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md` (§1.1 surface 3, §2 — gap attribution discipline used above).
- **Epic**: Workflow-Reflex 1 of 6. Siblings: SPEC-MODEL-ROUTING-WIRE-001 (2 of 6), SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 1 of 6. Failure-event recording + classifier eligibility + propose auto-run + lessons inbox + doctor dormancy check. Tier M. |
