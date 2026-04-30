# Plan: SPEC-ADVISOR-001 (Advisor Strategy Adoption)

> Companion to `spec.md`. Implementation plan with milestones, technical approach, and risks.

---

## Implementation Plan

### Milestone M0 — Verification Probe (Priority: Critical)

**Goal**: Resolve OQ-1 (advisor_20260301 availability in Claude Code sub-agent context).

**Tasks**:
1. Create probe agent at `.claude/agents/moai/.scratch/advisor-probe.md` with `tools: ..., advisor_20260301`. Run via Agent() and observe whether Claude Code accepts the tool declaration.
2. Document probe outcome in `.moai/reports/advisor-probe-{DATE}/result.md`.
3. Branch implementation: native path if accepted, fallback path if rejected.

**Exit criteria**: Probe artifact exists; native-vs-fallback decision recorded.

---

### Milestone M1 — Protocol Documentation (Priority: High)

**Goal**: REQ-ADV-001, REQ-ADV-002, REQ-ADV-006 — capture the advisor contract in `.claude/rules/moai/development/model-policy.md`.

**Tasks**:
1. Add new section "Advisor Protocol" to model-policy.md with:
   - Advisor-vs-executor role contract (from research.md §1.1)
   - Frontmatter schema for `advisor:` block (model, max_uses, trigger_class)
   - Telemetry schema (REQ-ADV-014 fields)
2. Update `.claude/rules/moai/development/agent-authoring.md` with frontmatter schema reference.
3. Update Constitution registry (`.claude/rules/moai/core/zone-registry.md`) with advisor-related HARD rules (REQ-ADV-002, REQ-ADV-015, REQ-ADV-016).

**Exit criteria**: Protocol doc reviewed; agent-authoring guide cross-references it; zone-registry has new entries.

---

### Milestone M2 — Telemetry Schema Extension (Priority: High)

**Goal**: REQ-ADV-014 — extend `internal/hook/post_tool.go` schema with advisor fields.

**Tasks**:
1. Extend `internal/hook/types.go` (already in working tree) with `AdvisorTelemetry` struct: `Model`, `Invocations`, `Tokens`, `LatencyMsTotal`, `CapReached`.
2. Update `internal/hook/post_tool.go` to populate advisor fields when present in agent completion data.
3. Add unit test `internal/hook/post_tool_advisor_test.go`.
4. Update `.moai/config/sections/observability.yaml` with advisor field documentation.
5. Validate REQ-ADV-015 (no raw response logging) via test.

**Exit criteria**: Unit test passes; sample log line contains advisor fields; raw-response leak test fails when intentional violation introduced.

---

### Milestone M3 — Pilot #1: manager-spec migration (Priority: High)

**Goal**: REQ-ADV-003, REQ-ADV-007 / REQ-ADV-013, REQ-ADV-006 — migrate manager-spec to Sonnet executor + Opus advisor.

**Tasks (native path, M0 succeeded)**:
1. Update `.claude/agents/moai/manager-spec.md` frontmatter:
   - `model: sonnet` (was opus)
   - `effort: high` (was xhigh)
   - Add `advisor:` block with `model: opus`, `max_uses: 3`, `trigger_class: spec_decisions`
2. Add to body a "Advisor Consultation Protocol" section listing the three trigger classes (EARS pattern selection, SPEC vs Report classification, Exclusions judgment).
3. Mirror to `internal/template/templates/.claude/agents/moai/manager-spec.md` (Template-First).
4. Run `make build`.

**Tasks (fallback path, M0 indicated rejection)**:
1. Update manager-spec frontmatter: `model: sonnet`, `effort: high`.
2. Add to body a "Advisor Consultation Protocol" section that uses `Agent(subagent_type: "researcher", model: "opus", mode: "plan", description: "advisor consultation: <decision class>")`.
3. Document the call signature and parsing of the response.
4. Mirror and rebuild.

**Exit criteria**: 5 manager-spec invocations succeed and produce schema-valid SPECs (AC-2). Telemetry logs show `advisor_invocations >= 1` on at least 3 of 5.

---

### Milestone M4 — Pilot #2: evaluator-active augmentation (Priority: High)

**Goal**: REQ-ADV-003 — add advisor budget to evaluator-active.

**Tasks**:
1. Update `.claude/agents/moai/evaluator-active.md` frontmatter: add `advisor:` block with `model: opus`, `max_uses: 2`, `trigger_class: dimension_scoring`.
2. Add to body a "Dimension Scoring Advisor Protocol" section listing trigger boundaries (Functionality 0.75/0.50, Security Critical/High borderline, Consistency severity).
3. Mirror and rebuild.
4. Run baseline collection: 5 evaluator-active runs on existing SPECs **without advisor** (control group).
5. Run pilot collection: 5 evaluator-active runs on the same SPECs **with advisor** (treatment group).
6. Compute dimension-score variance (REQ-ADV-017).

**Exit criteria**: Variance < 0.10 on 4 of 4 dimensions across the 5-pair sample (AC-3). If variance >= 0.10, escalate per REQ-ADV-017.

---

### Milestone M5 — Cap and Fallback Validation (Priority: Critical)

**Goal**: REQ-ADV-005, REQ-ADV-008, REQ-ADV-013 — verify cap enforcement and fallback resilience.

**Tasks**:
1. Add fault-injection test: temporarily set `advisor.max_uses: 0` and verify executor produces valid output without consultation (AC-5 baseline behavior under cap).
2. Add fault-injection test: stub the advisor tool to return error; verify executor falls through to primary reasoning and logs `advisor_fallback: failed` (AC-6).
3. Add cap-exceeded test: stub the executor to attempt 4 advisor calls under `max_uses: 3`; verify the 4th is rejected and `advisor_cap_reached: true` is logged.

**Exit criteria**: Three test cases pass; logs contain expected fallback markers.

---

### Milestone M6 — Cost Telemetry Validation (Priority: High)

**Goal**: REQ-ADV-014 — validate AC-4 cost-reduction target.

**Tasks**:
1. Run baseline window: 10 manager-spec invocations on representative SPEC requests **before** M3 (capture from existing telemetry or replay).
2. Run pilot window: 10 manager-spec invocations **after** M3.
3. Compute aggregate Opus token usage delta. AC-4 target: >= 30% reduction.
4. Document result in `.moai/reports/advisor-pilot-results-{DATE}/`.

**Exit criteria**: Reduction artifact stored. If < 30%, escalate via plan-auditor (rather than auto-rollback) since the floor is conservatism.

---

## Technical Approach

### Decision: Single SPEC, two pilots

We pilot two agents in one SPEC instead of one-pilot-per-SPEC because:
- The protocol (REQ-ADV-001 / REQ-ADV-002 / REQ-ADV-014) is shared.
- Telemetry schema must land before either pilot.
- Roll-back semantics are uniform.

If the pilots diverge in success outcomes (e.g., manager-spec succeeds but evaluator-active fails REQ-ADV-017), the rollback is per-agent — manager-spec stays migrated; evaluator-active reverts.

### Decision: Native path preferred, fallback path validated

We do not skip the verification probe (M0) on the assumption that Claude Code supports the native tool. Costs of getting this wrong: a half-completed migration with broken agents. M0 cost: one probe agent run (~minutes).

### Decision: max_uses caps are starting points

`max_uses: 3` for manager-spec and `max_uses: 2` for evaluator-active are calibrated against research.md §2.2 / §2.3 observed decision counts. After M6 pilot data, they may be tuned. No code-level minimum is enforced; the cap is purely runtime.

### Code-Level Architecture Notes (informational, deferred to Run phase)

The Run phase will resolve specifics not bound by this SPEC:
- Whether `advisor:` block is parsed in `internal/template/agents/parser.go` (if such a parser exists) or directly in the Claude Code runtime.
- Whether telemetry uses the existing `hook-metrics.jsonl` or a new `advisor-metrics.jsonl` file.
- The fallback Agent() spawn signature exactness (the spec.md REQ-ADV-013 wording is canonical; the Run phase can refine the prompt template).

These are HOW-level decisions and remain out of scope for spec.md per [HARD] SPEC scope rule.

---

## Risks

| # | Risk | Severity | Likelihood | Mitigation | Tracked in |
|---|------|----------|------------|------------|-----------|
| R1 | M0 probe ambiguous (tool accepted but no-op) | High | Medium | Define probe success as "advisor token returns measurable response", not just "tool declaration accepted" | M0 task 1 |
| R2 | M3/M4 telemetry undercount (KV cache hits not counted in advisor_tokens) | Medium | High | Anthropic API returns input/output tokens; both fields recorded | M2, M6 |
| R3 | Sonnet executor regression on edge SPECs (e.g., complex Multi-Phase frontmatter) | High | Medium | Compare AC-2 outputs to pre-pilot Opus baselines via diff; reject pilot if frontmatter schema violations appear | M3 exit |
| R4 | Pilot data window too small (5 SPECs, 5 evaluations) for statistical confidence | Medium | High | Acknowledge in M6 report; treat as preliminary; commit to extending window in Wave 2 if needed | M6 |
| R5 | manager-spec body bloat from added Advisor Protocol section | Low | Medium | Keep new section under 30 lines; cross-reference model-policy.md for full protocol | M3 task 2 |
| R6 | Roll-back during a live SPEC creation produces inconsistent artifact | High | Low | Roll-back applies only at session start (frontmatter swap); in-flight SPECs complete on current configuration | M5 task design |
| R7 | Telemetry log file contains accidental PII via guidance text | Critical | Low | REQ-ADV-015 forbids raw response logging; M2 includes negative test ("forbidden field" scanner) | M2 task 5 |

---

## Approval Points

- **AP-1**: After M0 — user approval to proceed with native-vs-fallback decision (AskUserQuestion in plan phase)
- **AP-2**: After M2 — user review of telemetry schema before pilots run
- **AP-3**: After M3 — user review of first manager-spec output via the new protocol
- **AP-4**: After M6 — user decision: keep pilot, expand to expert-security in Wave 2, or roll back

---

## Handover to Run Phase

When this SPEC moves to /moai run:

- **Manager**: manager-strategy (architecture finalization), then manager-tdd (test-first for telemetry path)
- **Expert delegation**: expert-backend for `internal/hook/` Go changes; no frontend impact.
- **Skills to load**: `moai-foundation-core`, `moai-foundation-thinking`, `moai-workflow-spec`.
- **Quality gate**: Standard harness (this is observability + frontmatter changes, not deeply complex).
- **Critical path file order**:
  1. `.claude/rules/moai/development/model-policy.md` (M1)
  2. `internal/hook/types.go`, `internal/hook/post_tool.go` (M2)
  3. `internal/template/templates/.claude/agents/moai/manager-spec.md` (M3)
  4. `internal/template/templates/.claude/agents/moai/evaluator-active.md` (M4)
  5. `.claude/rules/moai/core/zone-registry.md` (M1 trailing)

---

**Status**: draft — review with plan-auditor and stakeholder before promotion to approved.
