# SPEC-V3R3-HARNESS-LEARNING-001 — Acceptance Criteria

All acceptance scenarios use Given-When-Then format. Each AC explicitly references the REQ-IDs it covers (REQ traceback). Coverage MUST be 100% — every REQ-HL-NNN must appear in at least one AC.

## AC-01: Activity Observer Records Events with Bounded Latency

**Covers**: REQ-HL-001

**Given** a fresh project with `.moai/harness/usage-log.jsonl` empty and `learning.enabled: true`,
**When** the user invokes `/moai plan "test feature"` followed by `/moai run SPEC-TEST-001`,
**Then**:
- `.moai/harness/usage-log.jsonl` contains exactly two new lines.
- Each line is valid JSON parseable by `jq`.
- Each line contains the fields: `ts` (RFC3339), `event_type`, `subject`, `context_hash` (sha256:...), `tier_increment` (integer), `session_id` (UUID).
- The PostToolUse hook duration measured by `time` is under 100ms per invocation.
- The parent tool call's exit status is unchanged versus a control run with the observer disabled.

### Edge Cases
- Concurrent `/moai` invocations from two terminals → both events recorded without line interleaving (file lock or O_APPEND atomicity).
- Disk-full condition → observer logs warning to stderr but does NOT crash the parent tool.

---

## AC-02: Tier 1 → Tier 2 → Tier 3 Promotion with Description Enrichment

**Covers**: REQ-HL-002, REQ-HL-003

**Given** `.moai/harness/usage-log.jsonl` contains 4 prior events for pattern `subcommand:/moai design`,
**When** a 5th matching event is recorded and the learner runs,
**Then**:
- `.moai/harness/learning-history/tier-promotions.jsonl` gains one line with `from_tier: "rule"` and `to_tier: "rule"` (or `from_tier: "heuristic"`, `to_tier: "rule"` if the prior tier was heuristic).
- At Tier 2 (3rd observation), only the `description` field of the matching `.claude/skills/my-harness-*/SKILL.md` frontmatter changed; body is byte-identical.
- The description change is a single appended line prefixed with `# heuristic:`.

### Edge Cases
- Pattern with confidence below 0.70 → remains classified as Observation regardless of count (verified by injecting a low-confidence pattern with 5+ events).
- Description already contains a `# heuristic:` line → updated in place, not duplicated.

---

## AC-03: Tier 3 Rule Injection Modifies Frontmatter Triggers and Creates Snapshot

**Covers**: REQ-HL-004

**Given** a pattern at 4 observations with the trigger keyword `oauth`,
**When** a 5th observation triggers Tier 3 promotion,
**Then**:
- The matching `.claude/skills/my-harness-*/SKILL.md` frontmatter `triggers` (or `keywords`) list contains `oauth` exactly once (deduplicated).
- A snapshot directory `.moai/harness/learning-history/snapshots/<ISO-DATE>/` is created containing the pre-change file copy.
- The snapshot directory contains a `manifest.json` listing every file modified in the snapshot operation.

### Edge Cases
- `triggers` list missing from frontmatter → list is created with the new keyword as its sole element.
- Two patterns promote in the same minute → snapshot directory uses microsecond precision in its name to avoid collision.

---

## AC-04: Tier 4 Auto-Update Gated by Human Approval and Rate Limit

**Covers**: REQ-HL-005, REQ-HL-008 (rate-limit dimension)

**Given** `learning.auto_apply: false` (default), a pattern reaches 10 observations, and zero auto-updates have occurred in the last 7 days,
**When** the learner attempts to apply the change to `.moai/harness/chaining-rules.yaml`,
**Then**:
- The change is **not** written to disk.
- A pending proposal is recorded at `.moai/harness/learning-history/pending-proposals.jsonl`.
- The coordinator skill `moai-harness-learner` surfaces the proposal payload to the orchestrator, which presents it via `AskUserQuestion`.
- Only after explicit user approval (Yes option selected) does the applier write the change.
- The rate-limit counter increments by 1.

**And When** the user attempts a 4th auto-update within the same 7-day window,
**Then**:
- The applier rejects the write with rationale `rate_limit_exceeded` logged to `learning-history/rate-limit-rejections.jsonl`.
- No file modification occurs.

### Edge Cases
- User declines proposal → proposal moved to `learning-history/declined-proposals.jsonl`, NOT auto-resurrected on next learner run for 30 days.
- `auto_apply: true` and rate limit not exceeded → applier writes without `AskUserQuestion`, but Frozen Guard (L1) and Canary (L2) still enforced.

---

## AC-05: Frozen Guard Blocks Writes to MOAI-Managed Area

**Covers**: REQ-HL-006

**Given** the learner has produced a synthetic proposal targeting `.claude/skills/moai-foundation-core/SKILL.md`,
**When** the applier processes the proposal,
**Then**:
- No write occurs to the target path.
- `.moai/harness/learning-history/frozen-guard-violations.jsonl` gains one entry containing the attempted path, timestamp, and the upstream caller (Phase 3/4 module).
- A non-blocking warning is emitted to stderr.
- The applier continues processing subsequent (legitimate) proposals.

### Edge Cases
- Proposal path uses symlink that resolves into MOAI area → resolved path checked against Frozen Guard, blocked.
- Configuration attempts to add a `frozen_guard.disable: true` key → key is ignored; Frozen Guard is hardcoded.
- Path traversal (`..`) attempt → normalized via `filepath.Clean`, then matched against Frozen Guard prefixes.

---

## AC-06: Canary Check Rejects Score-Regressing Changes; Contradiction Detector Surfaces Conflicts

**Covers**: REQ-HL-007 (Canary), REQ-HL-008 (Contradiction)

**Given** the most recent 3 sessions in `.moai/harness/usage-log.jsonl` produce a baseline effectiveness score of 0.85, and a proposed Tier 4 change shadow-evaluates to 0.70 (drop of 0.15, exceeding the 0.10 threshold),
**When** the safety pipeline processes the change,
**Then**:
- The change is rejected at Layer 2 (Canary).
- A rejection record is appended to `learning-history/canary-rejections.jsonl` with both scores and the threshold value.
- The change is NOT presented to the user via AskUserQuestion (rejected before reaching L5).

**And Given** a proposed change adds the trigger `database` to skill `my-harness-backend` while the user's existing customization already maps `database` to `my-harness-data`,
**When** the safety pipeline reaches Layer 3 (Contradiction Detector),
**Then**:
- Both the existing and proposed mappings are surfaced to the user via `AskUserQuestion` with three options: keep existing (recommended), accept proposed, manual merge.
- No silent override occurs.

### Edge Cases
- Canary baseline cannot be computed (fewer than 3 sessions) → change deferred until enough data, NOT auto-approved.
- Contradiction detector finds no conflict → change passes through L3 silently.

---

## AC-07: CLI Subcommand `/moai harness` Provides Status, Apply, Rollback, Disable

**Covers**: REQ-HL-009

**Given** an active harness with 12 promotions over 14 days and one pending Tier 4 proposal,
**When** the user runs `/moai harness status`,
**Then**:
- Output includes: tier distribution counts, last update timestamp, rate-limit window remaining, pending proposal count, observer enabled/disabled flag.
- Exit status is 0.

**And When** the user runs `/moai harness apply`,
**Then**:
- The next pending Tier 4 proposal is presented and (on user approval) applied via the safety pipeline.

**And Given** a snapshot exists at `.moai/harness/learning-history/snapshots/2026-04-25T10:00:00Z/`,
**When** the user runs `/moai harness rollback 2026-04-25T10:00:00Z`,
**Then**:
- All files in the snapshot's `manifest.json` are restored byte-identical to their snapshot state.
- A rollback event is recorded at `learning-history/rollbacks.jsonl`.
- Exit status is 0.

**And When** the user runs `/moai harness disable`,
**Then**:
- `.moai/config/sections/harness.yaml` `learning.enabled` is set to `false`.
- Subsequent `/moai` invocations do NOT write to `.moai/harness/usage-log.jsonl`.

### Edge Cases
- `/moai harness rollback <nonexistent-date>` → exits with status 1 and message listing available snapshot dates.
- `/moai harness apply` with no pending proposals → exits with status 0 and message `no pending proposals`.

---

## AC-08: Configuration Loaded from Template-First Mirror; Logs Pruned per Retention Policy

**Covers**: REQ-HL-010, REQ-HL-011

**Given** a fresh project initialized via `moai init test-project`,
**When** the user inspects `.moai/config/sections/harness.yaml`,
**Then**:
- The file exists.
- It contains the `learning:` key with sub-keys `enabled: true`, `auto_apply: false`, `tier_thresholds: [1, 3, 5, 10]`, `rate_limit: {max_per_week: 3, cooldown_hours: 24}`, `log_retention_days: 90`.
- The file content is byte-identical to `internal/template/templates/.moai/config/sections/harness.yaml` (Template-First mirror enforced).

**And Given** `.moai/harness/usage-log.jsonl` contains entries with `ts` older than `log_retention_days`,
**When** the observer writes a new event,
**Then**:
- Stale entries (older than `log_retention_days`) are removed from `usage-log.jsonl`.
- Removed entries are appended to `.moai/harness/learning-history/archive/<YYYY-MM>.jsonl.gz`.
- Archives older than 2× `log_retention_days` are deleted.

### Edge Cases
- `log_retention_days: 0` → entries deleted immediately on next write (no archive); explicit user choice respected.
- Archive write fails (disk full) → observer logs warning, retains stale entries until next successful archive.

---

## Definition of Done

- [ ] All 8 ACs pass on macOS, Linux, Windows CI runners.
- [ ] All 11 REQ-IDs covered (verified by AC traceback table in `spec.md` §6).
- [ ] Unit test coverage ≥ 85% for `internal/harness/` package.
- [ ] Frozen Guard violation count = 0 in a 1-week real-developer trial.
- [ ] `/moai harness status` returns valid output on a fresh project AND on a project with 100+ accumulated events.
- [ ] `internal/template/templates/.moai/config/sections/harness.yaml` mirrors the project default byte-identically (verified by `commands_audit_test.go`-style enforcer).
- [ ] No write recorded to `.claude/agents/moai/`, `.claude/skills/moai-*/`, or `.claude/rules/moai/` by this subsystem (verified by IT-04 + log audit).
- [ ] Documentation: `.moai/harness/README.md` documents all CLI verbs, config keys, and tier thresholds.
- [ ] TRUST 5 quality gates passed (Tested, Readable, Unified, Secured, Trackable).
- [ ] Plan-auditor PASS verdict on Phase 0.5 audit gate before `/moai run` proceeds to Phase 1.
