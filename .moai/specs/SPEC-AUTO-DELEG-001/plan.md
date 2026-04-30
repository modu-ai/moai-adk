# Plan: SPEC-AUTO-DELEG-001 (Auto-Delegation Rules Strengthening)

> Companion to `spec.md`. Implementation plan with milestones, technical approach, risks.

---

## Implementation Plan

### Milestone M0 — OQ Resolution + Baseline Capture (Priority: Critical)

**Goal**: Resolve open questions OQ-1 through OQ-7 AND capture pre-SPEC baseline measurement data BEFORE any changes.

**Tasks**:
1. **OQ-1**: Decide insertion form for the 5 triggers in CLAUDE.md §1. Recommend: grouped block with sub-heading "Delegation Triggers". Confirm via AskUserQuestion.
2. **OQ-2**: Confirm 30-percentage-point target as a target (not hard gate) per default.
3. **OQ-3**: Audit and confirm 24-agent count (23 local files + 1 Explore reference, OR 23 alone if Explore is excluded). Document in `.moai/reports/auto-deleg-count-{DATE}/inventory.md`.
4. **OQ-4**: Author 1-page heuristic for "independent work units" in `.moai/reports/auto-deleg-{DATE}/independence-heuristic.md`.
5. **OQ-5**: Decide multi-language section policy. Default: all four sections (EN/KO/JA/ZH) updated.
6. **OQ-6**: Use specific agent names in NOT-for clauses to enable graph (REQ-DEL-011).
7. **OQ-7**: Confirm 5 missing-keyword agents are in scope.
8. **Baseline measurement** (REQ-DEL-013):
   - Source: 50 user turns sampled from session transcripts in `~/.claude/projects/{hash}/` dated within prior 60 days. Random sample with random seed recorded.
   - For each turn: did the orchestrator invoke `Agent()` without the user explicitly saying "Use the X subagent"?
   - Compute baseline rate: count_auto_delegated / 50.
   - Artifact: `.moai/reports/auto-deleg-baseline-{DATE}/sample-50-turns.md`.

**Exit criteria**: All OQs resolved; baseline rate documented; user approves measurement methodology before pilot rollout.

---

### Milestone M1 — CLAUDE.md §1 Update (Priority: Critical)

**Goal**: REQ-DEL-001, REQ-DEL-002 — add 5 [HARD] delegation triggers.

**Tasks**:
1. Edit `internal/template/templates/CLAUDE.md` (Template-First). Insert the new "Delegation Triggers" sub-heading immediately after the "Multi-File Decomposition" rule:

   ```markdown
   #### Delegation Triggers (Auto-Delegation Mandates)

   - [HARD] File-Exploration Trigger: WHEN a task requires reading/exploring 10+ files, delegate to `Explore` or a domain manager subagent.
   - [HARD] Independence Trigger: WHEN a task contains 3+ independent work units (outputs don't depend on each other), delegate at least one unit to a specialist subagent.
   - [HARD] Freshness Trigger: WHEN a task requires fresh perspective on just-produced output (review, audit, evaluation), delegate to an isolated subagent (`evaluator-active`, `plan-auditor`, `manager-quality`).
   - [HARD] Pipeline Trigger: WHEN a task has clear pipeline structure (input → transform → output across phases), delegate per-stage specialists per `agent-patterns.md` Pattern 1.
   - [HARD] Multi-Write Trigger: WHEN a task creates 5+ files of the same kind (tests, agents, SPECs), delegate to the appropriate builder or expert subagent.

   Exceptions (no delegation required): typo fixes, single-line changes, single-file reads with explicit path, command invocations with all arguments provided.
   ```

2. Cross-reference §14 (Parallel Execution Safeguards): no contradiction; §14 governs file-write conflict prevention while §1 Delegation Triggers govern when to delegate.
3. Cross-reference §16 (CLAUDE.local.md self-check): the new §1 entries replace the local-only §16 quantitative triggers. Update CLAUDE.local.md §16 to defer to CLAUDE.md §1 with a "see §1 Delegation Triggers" pointer.
4. Run `make build` to regenerate embedded templates.
5. Verify local `CLAUDE.md` is in sync.

**Exit criteria**: CLAUDE.md §1 has 5 new [HARD] entries; template-mirrored; rebuilt; CLAUDE.local.md §16 updated to defer.

---

### Milestone M2 — Agent-Authoring Guide Update (Priority: High)

**Goal**: REQ-DEL-016, AC-6 — document the description trigger format.

**Tasks**:
1. Edit `.claude/rules/moai/development/agent-authoring.md` (template-first via `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`). Add new section "Description Trigger Format":

   ```markdown
   ## Description Trigger Format (REQ-DEL-009 / SPEC-AUTO-DELEG-001)

   The description field controls auto-delegation. Use trigger-centric phrasing.

   Required structure:
   ```
   <Capability summary one-liner>. Use PROACTIVELY when:
   - <Trigger condition 1>
   - <Trigger condition 2>
   - [optional] <Trigger condition 3+>
   NOT for: <explicit boundary> (use <other-agent> instead)
   ```

   Trigger conditions SHOULD be observable (file counts, keywords, task shapes), not subjective ("when complex"). NOT-for clauses SHOULD name the alternative agent to form a routing graph.
   ```

2. Reference the 5 [HARD] triggers in CLAUDE.md §1.
3. Mirror and rebuild.

**Exit criteria**: agent-authoring.md updated; cross-reference verified.

---

### Milestone M3 — Description Rewrite Pass 1 (Priority: Critical, High Volume)

**Goal**: REQ-DEL-009, REQ-DEL-010, REQ-DEL-012 — rewrite all 24 agent descriptions.

**Tasks**: For each of the 24 agents, perform an in-place rewrite:

   - Preserve the leading capability summary sentence (REQ-DEL-012).
   - Replace `Use PROACTIVELY for <area>` with `Use PROACTIVELY when:` followed by 2-3 trigger conditions (each observable).
   - Add or expand the `NOT for:` clause with specific alternative agent names (REQ-DEL-011).
   - Preserve the existing i18n keyword sections (EN/KO/JA/ZH per OQ-5).

Per-agent task breakdown (representative — exact list resolved in M0):

| Agent | Trigger conditions to add | NOT-for delegation |
|-------|---------------------------|---------------------|
| manager-spec | EARS requirements needed; SPEC document creation; user mentions "spec", "requirement", "EARS" | NOT for: implementation (use manager-ddd/tdd) |
| manager-strategy | Architecture decision pending; technology trade-off analysis; > 3 alternatives to compare | NOT for: implementation (use manager-ddd/tdd) |
| manager-ddd | DDD ANALYZE-PRESERVE-IMPROVE cycle; behavior-preserving refactor of existing code | NOT for: TDD-first work (use manager-tdd) |
| manager-tdd | New feature with test-first approach; coverage gap closure | NOT for: behavior-preserving refactor (use manager-ddd) |
| manager-quality | TRUST 5 validation; quality gate enforcement; lint/test toolchain run | NOT for: independent skeptical review (use evaluator-active) |
| manager-docs | Documentation generation post-implementation; API doc updates; CHANGELOG entry | NOT for: SPEC creation (use manager-spec) |
| manager-project | New project initialization; .moai/project/* document generation | NOT for: feature SPECs (use manager-spec) |
| manager-git | Git commits, branches, PR creation; release tag automation | NOT for: code implementation (use any expert/manager) |
| expert-backend | API contract design needed; auth flow implementation; > 200 LOC server code | NOT for: frontend (use expert-frontend); security audit (use expert-security) |
| expert-frontend | UI component design; accessibility audit; client-side state management | NOT for: API contract (use expert-backend) |
| expert-security | OWASP-relevant code review; auth code audit; secret handling review | NOT for: general backend (use expert-backend) |
| expert-devops | CI/CD pipeline change; Dockerfile / Kubernetes manifest; deployment automation | NOT for: code implementation (use expert-backend/frontend) |
| expert-performance | Profiling needed; benchmark regression investigation; latency optimization | NOT for: correctness bugs (use expert-debug) |
| expert-debug | Reproducer needed; error diagnosis with stack trace; flaky test investigation | NOT for: performance optimization (use expert-performance) |
| expert-testing | Test strategy design; E2E coverage planning; load testing | NOT for: unit-test writing within DDD/TDD cycle (use manager-ddd/tdd) |
| expert-refactoring | Codemod / AST migration; large-scale rename; > 10-file refactor | NOT for: small in-place edits (handle directly) |
| expert-mobile | iOS/Android native or Flutter/RN code | NOT for: backend APIs the mobile app calls (use expert-backend) |
| builder-agent | New agent definition needed | NOT for: skill creation (use builder-skill) |
| builder-skill | New skill creation | NOT for: agent creation (use builder-agent) |
| builder-plugin | Plugin scaffold or marketplace registration | NOT for: in-tree skill/agent (use builder-skill/agent) |
| evaluator-active | Independent skeptical evaluation post-implementation | NOT for: TRUST 5 enforcement (use manager-quality) |
| plan-auditor | Plan-phase document audit (SPEC schema, EARS compliance, bias check) | NOT for: post-implementation review (use evaluator-active) |
| researcher | Codebase exploration with > 10 files; cross-cutting investigation | NOT for: focused single-file reads (use Read tool directly) |

(The 24th agent depends on inventory resolution in OQ-3; if `Explore` is in scope, its description in CLAUDE.md §4 also receives a trigger update.)

For each agent, work in `internal/template/templates/.claude/agents/moai/<name>.md` first, then mirror.

**Exit criteria**: 24 agent files have updated descriptions; audit script (M4) green.

---

### Milestone M4 — Audit Script (Priority: High)

**Goal**: REQ-DEL-009, REQ-DEL-010, AC-4 — automated guardrail.

**Tasks**:
1. Add `internal/template/agent_description_audit_test.go`. Logic:
   - Glob `.claude/agents/moai/*.md` (template path).
   - For each file:
     - Parse frontmatter `description` field.
     - Assert presence of `Use PROACTIVELY when:` (positive match).
     - Assert absence of `Use PROACTIVELY for ` (negative match — legacy form).
     - Assert presence of `NOT for:` clause.
     - Assert at least 2 trigger conditions (count of `- ` bullets after PROACTIVELY when).
2. Run audit. Iterate M3 if any agent fails.
3. Document the audit's expected count (24 OR 23, per OQ-3).

**Exit criteria**: Audit test passes for all 24 (or confirmed 23) agents.

---

### Milestone M5 — NOT-for Graph Validation (Priority: Medium)

**Goal**: REQ-DEL-011, AC-8 — verify the routing graph is acyclic.

**Tasks**:
1. Author `internal/template/agent_routing_graph_test.go`. Logic:
   - Parse each agent's NOT-for clause.
   - Extract referenced agent names (e.g., "use expert-security" → edge to expert-security).
   - Build a directed graph: node = agent, edge = "NOT for X — use Y" yields A → Y.
   - Run cycle detection.
   - Assert no cycles AND every referenced target agent exists.
2. Document any unexpected edges (e.g., circular hints between two managers).

**Exit criteria**: Graph audit passes; cycle-free; all targets resolved.

---

### Milestone M6 — Pilot Measurement (Priority: Critical)

**Goal**: REQ-DEL-014, REQ-DEL-015, AC-3 — verify 30 pp improvement.

**Tasks**:
1. Wait at least 14 days post-rollout to accumulate post-rollout transcripts.
2. Sample 50 user turns from the post-rollout window using the same methodology as M0.
3. Compute pilot rate: count_auto_delegated / 50.
4. Compare to baseline. If `pilot_rate >= baseline_rate + 0.30`, AC-3 satisfied.
5. If not satisfied, REQ-DEL-015 triggers: file `.moai/reports/auto-deleg-tuning-{DATE}/regression.md` with:
   - The 50-turn sample with per-turn analysis.
   - Hypothesis on why triggers didn't fire (too high threshold? too narrow scope? trigger phrasing not picked up?).
   - Proposed tuning (e.g., "lower file-exploration threshold from 10 to 7").

**Exit criteria**: Pilot artifact recorded. AC-3 either satisfied or tuning hypothesis filed.

---

## Technical Approach

### Decision: Both layers in one PR (not split)

We could split into two SPECs (CLAUDE.md changes vs description changes). We don't because:
- Anthropic's prescription requires both (research.md §4.1).
- Splitting requires duplicate measurement runs.
- Risk of orphan state if one ships and the other doesn't.
- The atomic operation matches the audit test's expectations (M4 verifies post-state).

### Decision: Manual rewrite, not regex

Each agent's domain semantics are encoded in description text. Regex rewrite (`s/Use PROACTIVELY for /Use PROACTIVELY when:\n- /`) produces grammatically correct but semantically wrong triggers. Per-agent thoughtful rewrite is the right cost (one-time, ~10 minutes per agent).

### Decision: Description format, not body format

The body remains untouched (REQ-DEL-020). SPEC-AGENT-002 already minimized bodies. Trigger discoverability is a description concern; body content is a behavior concern.

### Decision: Sample-based measurement, not exhaustive telemetry

Building a real-time telemetry system to track auto-delegation rate is a multi-week engineering project. AC-3 requires only one before-after comparison. A 50-turn sample is statistically meaningful for a 30 pp effect size; methodology is documented for repeatability.

### Decision: Cross-reference, not duplicate

The 5 [HARD] triggers in §1 are the canonical text. agent-authoring.md links to them. CLAUDE.local.md §16 is updated to defer. No content is duplicated; only links.

### Code-Level Architecture Notes (informational, deferred to Run phase)

- The audit tests (M4, M5) will live in `internal/template/`. Existing audit tests in that directory follow patterns we can reuse (e.g., `commands_audit_test.go`).
- Frontmatter parsing in tests can use a lightweight YAML parser or regex (frontmatter starts/ends with `---`).
- Template-First discipline applies: template is canonical; mirror is generated by `make build`.

---

## Risks

| # | Risk | Severity | Likelihood | Mitigation | Tracked in |
|---|------|----------|------------|------------|-----------|
| R1 | Pilot rate doesn't reach +30 pp (target missed) | Medium | Medium | REQ-DEL-015 escalation; tuning artifact, not auto-rollback | M6 |
| R2 | Baseline sample is biased (too many trivial / too many complex turns) | Medium | High | Document sample selection in M0; use random seed for reproducibility | M0, M6 |
| R3 | Audit test (M4) breaks unrelated agent definitions during rollout window | Low | Medium | Audit is greenfield (new test); won't break existing tests | M4 |
| R4 | NOT-for graph cycles introduced inadvertently | Medium | Medium | M5 graph audit catches at PR time | M5 |
| R5 | Existing SPECs / docs reference legacy description text | Medium | Low | grep for legacy phrases; update affected references in same PR | M3 |
| R6 | i18n keyword sections become inconsistent (EN updated, KO not) | Medium | Medium | M3 per-agent checklist enforces all 4 languages | M3 |
| R7 | CLAUDE.md §1 becomes overlong (12 → 17 [HARD] rules) and harder to read | Low | High | Group the 5 new entries under "Delegation Triggers" sub-heading per OQ-1; visually distinct | M1 |
| R8 | "Use PROACTIVELY when:" phrase is not actually special to Claude (placebo effect) | Medium | Low | If M6 measurement shows 0% effect, escalate to Anthropic via /moai feedback; rollback is cheap (description-only) | M6 |
| R9 | Splitting work via [HARD] Multi-File Decomposition rule (existing) conflicts with new delegation triggers | Low | Low | M1 task 2 explicitly reconciles; no rule should require both "decompose into TodoWrite" AND "delegate to subagent" simultaneously without guidance on when to use which | M1 |

---

## Approval Points

- **AP-1**: After M0 — user approval on OQ resolutions and baseline measurement methodology.
- **AP-2**: After M1 — user review of CLAUDE.md §1 diff before M3 (description rewrites).
- **AP-3**: After M3 — user review of 24-description diff (`.moai/reports/auto-deleg-pilot-{DATE}/before-after.md`).
- **AP-4**: After M6 — user decision on tuning vs accept the pilot result.

---

## Handover to Run Phase

When this SPEC moves to /moai run:

- **Manager**: manager-strategy first (architecture), then manager-tdd (audit tests in M4/M5 are test-shaped work).
- **Expert delegation**: builder-agent for the 24 description rewrites (it's the agent expert); expert-testing for audit-test design.
- **Skills to load**: `moai-foundation-core`, `moai-foundation-cc`, `moai-workflow-spec`.
- **Quality gate**: Standard harness.
- **Critical path file order**:
  1. M0 baseline artifact (one-time before any rule changes)
  2. `internal/template/templates/CLAUDE.md` (M1)
  3. `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (M2)
  4. `internal/template/templates/.claude/agents/moai/*.md` (M3 — 24 files)
  5. `internal/template/agent_description_audit_test.go` (M4)
  6. `internal/template/agent_routing_graph_test.go` (M5)
  7. M6 pilot artifact (post-rollout, time-delayed)

---

**Status**: draft — review with plan-auditor and stakeholder before promotion.
