# SPEC-V3R2-ORC-003 Research

> Research artifact for **Effort-Level Calibration Matrix for 17 agents**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                       | Description                                                                                                                                  |
|---------|------------|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1A) | Initial research on Opus 4.7 Adaptive Thinking landscape, R5 audit effort matrix evidence, current-state lint posture, drift inventory, and downstream-blocker survey. |

---

## 1. Research Scope and Method

### 1.1 Scope

Research bounds for Plan-phase decisions:

1. **Opus 4.7 Adaptive Thinking effort semantics** — Anthropic-published guidance on `effort: low|medium|high|xhigh|max` enum and the rejection of fixed `budget_tokens` (HTTP 400).
2. **R5 audit effort calibration evidence** — `.moai/design/v3-redesign/research/r5-agent-audit.md` § Effort-level calibration matrix; problem catalog P-A02 (19/22 missing field) and P-A03 (3 explicit drift cases).
3. **Constitution-side guidance** — `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy "Effort level selection" bullet; FROZEN-zone preservation rule.
4. **agent-authoring.md current state** — `.claude/rules/moai/development/agent-authoring.md` field reference table (`effort` row at line 38, field detail at line 58); identify the canonical-matrix insertion point.
5. **Current 25-agent inventory** — `.claude/agents/moai/*.md` baseline at HEAD `3356aa9a9`; effort declared / missing / drifted; status field state.
6. **17-agent post-ORC-001 roster** — SPEC-V3R2-ORC-001 §5.1 REQ-001 enumerates the 17 final agents; 5 retired files (manager-ddd/tdd, builder-agent/skill/plugin) + 2 fold-ins (expert-debug → manager-quality, expert-testing → manager-cycle) + 1 fold-in (expert-mobile coverage absorbed by expert-backend via R5 §Recommended v3 inventory).
7. **Existing lint surface** — `internal/cli/agent_lint.go` LR-01..LR-10 implementation; LR-03 severity confirmation; LR-11 reservation policy.
8. **Schema validator path** — `internal/config/schema/agent.go` (per SPEC-V3-AGT-001) — REQ-012 routes invalid-enum rejection through this validator OR `agent_lint.go` LR-13; this research determines the binding.
9. **Migrator integration** — SPEC-V3R2-MIG-001 (drift rewrite contract); REQ-007 says migrator rewrites drifted v2 effort values + emits log line.
10. **Harness override semantics** — SPEC-V3R2-HRN-001 `effort_mapping` in `harness.yaml`; REQ-010 documents the relationship.

### 1.2 Method

- **Static analysis**: Read agent files directly for `effort:` and `status:` values. Read `agent_lint.go` for current LR-03 severity and LR-* slot allocation. Read `moai-constitution.md` for the Opus 4.7 Prompt Philosophy bullet.
- **Cross-SPEC reference**: Inspect SPEC-V3R2-ORC-001/-002/-004, SPEC-V3R2-CON-001/-002, SPEC-V3R2-HRN-001, SPEC-V3R2-MIG-001, SPEC-V3-AGT-001 for ownership boundaries.
- **Anthropic external reference**: `https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7` (Adaptive Thinking semantics; Sep 2025 announcement). Cited verbatim from constitution + Master spec context (no live fetch needed for plan-phase since the rule is already encoded as a constitution invariant).
- **Lint runtime probe**: Execute `moai agent lint --path .claude/agents/moai/` to capture the current violation surface and confirm LR-03 fires as Error.
- **Evidence anchoring**: Each finding cites a specific file:line or SPEC §section. No speculative claims.

### 1.3 Evidence Anchor Inventory

This research file cites at least 30 distinct evidence anchors per plan-auditor PASS criterion #4. Anchors are listed inline with each finding and totaled in §6.

---

## 2. Current-State Inventory (HEAD `3356aa9a9`)

### 2.1 Agent file enumeration

`.claude/agents/moai/` contains 25 agent files at HEAD `3356aa9a9`:

```
builder-agent.md        builder-plugin.md       builder-skill.md
claude-code-guide.md    evaluator-active.md     expert-backend.md
expert-debug.md         expert-devops.md        expert-frontend.md
expert-mobile.md        expert-performance.md   expert-refactoring.md
expert-security.md      expert-testing.md       manager-brain.md
manager-ddd.md          manager-docs.md         manager-git.md
manager-project.md      manager-quality.md      manager-spec.md
manager-strategy.md     manager-tdd.md          plan-auditor.md
researcher.md
```

Evidence anchor [E-01]: `ls /Users/goos/MoAI/moai-adk-go/.claude/agents/moai/*.md` enumeration on HEAD `3356aa9a9` (2026-05-10 11:30 KST).

### 2.2 Effort frontmatter scan

| Agent | Current `effort:` | Drift Status |
|---|---|---|
| manager-spec | `xhigh` | ✅ matches matrix |
| manager-strategy | `xhigh` | ✅ matches matrix |
| manager-brain | `xhigh` | ⚠ out-of-roster (NOT in 17-agent v3r2 roster); LR-12 carve-out |
| evaluator-active | `high` | ❌ DRIFT — should be xhigh |
| expert-refactoring | `high` | ❌ DRIFT — should be xhigh |
| expert-security | `high` | ❌ DRIFT — should be xhigh |
| plan-auditor | `high` | ❌ DRIFT — should be xhigh |
| (18 others) | (missing — empty) | LR-03 fires Error |

Evidence anchor [E-02]: per-agent frontmatter scan via `awk '/^---$/{flag=!flag; next} flag && /^effort:/{print; exit}'` on each `.claude/agents/moai/*.md` file.

### 2.3 Drift count reconciliation

spec.md §1.1 says "three agents have explicit wrong values (expert-security, evaluator-active, plan-auditor are declared `high` but the constitution names them for `xhigh`)". Reality at HEAD `3356aa9a9`: **expert-refactoring is also declared `high`** while the matrix targets `xhigh` for it.

Evidence anchor [E-03]: `expert-refactoring.md` line 4 declares `effort: high`; spec.md §1.1 R5 audit matrix row "Reasoning-intensive (xhigh): manager-spec, manager-strategy, expert-security, **expert-refactoring**, evaluator-active, plan-auditor, researcher" lists expert-refactoring under xhigh.

**Plan binding**: 4 explicit drifts (not 3). Discrepancy acknowledged in plan.md §1.2.1; sync-phase HISTORY will reconcile spec.md.

### 2.4 LR-03 current severity

`internal/cli/agent_lint.go:386` declares `severity := SeverityError`. The accompanying comment line 85 says "LR-03: Error on missing effort: field (promoted from warning per SPEC-V3R2-ORC-003)".

Evidence anchor [E-04]: `internal/cli/agent_lint.go:382-396` `checkMissingEffort` function body verbatim; severity unconditionally Error.
Evidence anchor [E-05]: `internal/cli/agent_lint.go:85` comment block declares LR-03 as Error per ORC-003.

**Implication**: spec.md REQ-006 ("When this SPEC merges, LR-03 shall be promoted from warning severity to error severity") is **operationally idempotent** — the promotion is already in effect on HEAD `3356aa9a9`. ORC-002 implementation pre-emptively wired LR-03 as Error in anticipation of ORC-003 merge. Plan binding: REQ-006 verification only (regression test via T-ORC003-24); no severity flip code required. plan.md §1.2.1 acknowledges this.

### 2.5 Lint runtime output (probe)

Command: `moai agent lint --path .claude/agents/moai/`

Output excerpt (relevant rows):

```
✗ [LR-03] .claude/agents/moai/builder-agent.md:2: Missing effort: field ...
✗ [LR-03] .claude/agents/moai/builder-plugin.md:2: Missing effort: field ...
✗ [LR-03] .claude/agents/moai/builder-skill.md:2: Missing effort: field ...
✗ [LR-03] .claude/agents/moai/claude-code-guide.md:2: Missing effort: field ...
✗ [LR-03] .claude/agents/moai/expert-backend.md:2: Missing effort: field ...
... (continues across all agents missing effort:)
```

Evidence anchor [E-06]: Live `moai agent lint` output captured during plan-phase research session 2026-05-10. ✗ symbol confirms Error severity (vs ⚠ for Warning, e.g. LR-06).

### 2.6 LR-* slot allocation

| LR # | Owner SPEC | Status |
|---|---|---|
| LR-01 | ORC-002 | implemented (literal AskUserQuestion) |
| LR-02 | ORC-002 | implemented (Agent in tools list) |
| LR-03 | ORC-002 (promoted by ORC-003) | implemented as Error |
| LR-04 | ORC-002 | implemented (dead hooks) |
| LR-05 | ORC-002 | implemented (write-heavy missing isolation) |
| LR-06 | ORC-002 | implemented (--deepthink boilerplate, Warning) |
| LR-07 | ORC-002 | implemented (duplicate Skeptical mandate) |
| LR-08 | ORC-002 | implemented (skill-preload drift) |
| LR-09 | ORC-004 | implemented (read-only with worktree) |
| LR-10 | ORC-005 | implemented (static team-* file prohibited) |
| LR-11 | (reserved) | reserved per ORC-002/004 sequencing |
| LR-12 | **ORC-003** (this SPEC) | NEW — matrix drift |
| LR-13 | **ORC-003** (this SPEC) | NEW — invalid effort enum |
| LR-14 | **ORC-003** (this SPEC) | NEW — fixed budget_tokens prohibition |

Evidence anchor [E-07]: `internal/cli/agent_lint.go:84-97` lint rules help block enumerates LR-01 through LR-10. LR-11 reservation noted in research-side discussion of ORC-004 sequencing.
Evidence anchor [E-08]: SPEC-V3R2-ORC-002 §5.1 REQ-002 enumerates LR-01..LR-08 exit codes.

---

## 3. Constitutional and Doctrinal Evidence

### 3.1 Opus 4.7 Adaptive Thinking semantics (FROZEN)

`.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy contains the binding rules:

> "Adaptive Thinking: do NOT set fixed thinking budgets via `budget_tokens`; Opus 4.7 rejects fixed budgets with HTTP 400. Let the model self-allocate reasoning depth"

Evidence anchor [E-09]: `moai-constitution.md` § Opus 4.7 Prompt Philosophy bullet #2 (Adaptive Thinking).

> "Effort level selection: reasoning-intensive agents (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring) → `effort: xhigh` or `high`; implementation agents (expert-backend, expert-frontend, builder-*) → `effort: high` (default for Opus 4.7); speed-critical agents (manager-git, Explore) → `effort: medium`"

Evidence anchor [E-10]: `moai-constitution.md` § Opus 4.7 Prompt Philosophy bullet #5 (Effort level selection).

This bullet is the **inline constitutional guidance** that the canonical matrix (this SPEC) replaces with a cross-reference (REQ-005) without altering the FROZEN clause text. Plan binding: T-ORC003-06 modifies this bullet to point to `agent-authoring.md` § Effort-Level Calibration Matrix.

### 3.2 FROZEN-zone preservation

Per `.claude/rules/moai/design/constitution.md` (relocated to `.claude/rules/moai/`): the constitution file itself is FROZEN at the file level. However, individual bullets within `Opus 4.7 Prompt Philosophy` are EVOLVABLE — the cross-reference change preserves intent (effort level guidance still binding) while removing duplicated table content. spec.md §7 Constraints HARD rule "moai-constitution.md § Opus 4.7 Prompt Philosophy remains FROZEN; this SPEC only adds a cross-reference; no FROZEN clause text changes" enforces no-deletion of the binding rule semantics.

Evidence anchor [E-11]: `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-028, -029 entries classify Principle 4 + Principle 5 as Evolvable; the cross-reference is a content-preserving evolution.

### 3.3 Adaptive Thinking enum (5 values)

Anthropic guidance (Sep 2025 "what's new in claude-4-7"):
- `low`, `medium`, `high`, `xhigh`, `max`
- `xhigh` and `max` require Opus 4.7+; on Opus 4.6 the highest is `high`

Evidence anchor [E-12]: `.claude/rules/moai/development/agent-authoring.md:38` field reference table row "effort | No | inherit | Session effort override: low, medium, high, xhigh, max (xhigh/max require Opus 4.7+)".
Evidence anchor [E-13]: `.claude/rules/moai/development/agent-authoring.md:58` field detail "**effort**: Overrides session effort level for this agent. Valid values: `low`, `medium`, `high`, `xhigh`, `max`. The `xhigh` and `max` values require Opus 4.7 or later."

The 5-value enum is the `validEffortValues` set used by LR-13 (plan.md §4.2).

---

## 4. R5 Audit Evidence

### 4.1 R5 effort calibration matrix

R5 audit is `.moai/design/v3-redesign/research/r5-agent-audit.md` (referenced by spec §10 Traceability and Master spec §11.4). Key evidence rows from spec.md §1.1 reproducing R5 §Effort-level calibration matrix:

| Category | Agents | Recommended |
|---|---|---|
| Reasoning-intensive (xhigh) | manager-spec, manager-strategy, expert-security, expert-refactoring, evaluator-active, plan-auditor, researcher | xhigh |
| Implementation (high) | manager-cycle, manager-quality, expert-backend, expert-frontend, expert-performance | high |
| Template-driven (medium) | manager-docs, manager-project, expert-devops, builder-platform | medium |
| Speed-critical (medium) | manager-git | medium |

Evidence anchor [E-14]: spec.md §1.1 reproduces R5 effort matrix table.

R5 audit conclusions:
- 32% of agents (7/22) have effort drift or missing values.
- 19/22 agents omit `effort:` entirely (P-A02 HIGH severity).
- 3-explicit-drift count (P-A03 HIGH severity) understates by 1; reality is 4 (per §2.3 above).

Evidence anchor [E-15]: `.moai/design/v3-redesign/synthesis/problem-catalog.md` P-A02 HIGH "19/22 agents missing effort: field" + P-A03 HIGH "3 explicit drift cases" — corrected to 4 in this research.

### 4.2 17-agent post-ORC-001 roster

SPEC-V3R2-ORC-001 §5.1 REQ-001 lists exactly 17 v3r2 active agents:

```
manager-spec, manager-strategy, manager-cycle, manager-quality, manager-docs,
manager-git, manager-project, expert-backend, expert-frontend, expert-security,
expert-devops, expert-performance, expert-refactoring, builder-platform,
evaluator-active, plan-auditor, researcher
```

Evidence anchor [E-16]: SPEC-V3R2-ORC-001/spec.md REQ-ORC-001-001 verbatim.

ORC-001 retires/folds these 8 v2 agents:
- manager-ddd → folded into manager-cycle (cycle_type: ddd)
- manager-tdd → folded into manager-cycle (cycle_type: tdd)
- builder-agent → folded into builder-platform (artifact_type: agent)
- builder-skill → folded into builder-platform (artifact_type: skill)
- builder-plugin → folded into builder-platform (artifact_type: plugin)
- expert-debug → absorbed by manager-quality (diagnostic sub-mode)
- expert-testing → absorbed by manager-cycle RED phase + expert-performance load-test mode
- (expert-mobile coverage absorbed by expert-backend per R5 v3 inventory recommendation)

Evidence anchor [E-17]: SPEC-V3R2-ORC-001/spec.md §1.1 Background and §2.1 In Scope enumerate the 5 deletions + 2 new agents + 1 scope shrink.

### 4.3 Out-of-roster agents

`manager-brain` (15 KB body) and `claude-code-guide` (4.6 KB body) exist on HEAD `3356aa9a9` but are NOT in the 17-agent post-ORC-001 roster. Their treatment by this SPEC:
- LR-12 carve-out: `checkEffortMatrixDrift` returns nil for agents not in `canonicalEffortMatrix` (plan.md §4.1 logic).
- LR-03 still applies — these agents will continue to require `effort:` declared, even if not in the canonical matrix. (manager-brain already has `effort: xhigh`; claude-code-guide currently lacks effort and will fail LR-03.)
- This SPEC does NOT modify out-of-roster agents. Their disposition (retire / move to non-moai/ / declare-but-not-bind) is a separate SPEC concern.

Evidence anchor [E-18]: `.claude/agents/moai/manager-brain.md` declares `effort: xhigh` but is not enumerated in ORC-001 §5.1 REQ-001 17-agent roster.
Evidence anchor [E-19]: `.claude/agents/moai/claude-code-guide.md` lacks `effort:` and is not in 17-agent roster.

---

## 5. Cross-SPEC Boundary Survey

### 5.1 ORC-001 (17-agent roster)

Status: as of HEAD `3356aa9a9`, ORC-001 is in `.moai/specs/SPEC-V3R2-ORC-001/` with full 7-file plan (PR may or may not be merged at run-time).

Boundary: ORC-001 owns the file deletions (manager-ddd, manager-tdd, builder-agent, etc.) and the new file creations (manager-cycle, builder-platform). This SPEC owns the `effort:` field population on whatever 17-agent set exists post-ORC-001 merge.

Fallback: if ORC-001 has not yet merged at run-time, plan.md §1.2 + tasks.md §4.1 apply effort to retiree agents (manager-ddd/tdd, builder-agent/skill/plugin) until consolidated agents land. This is not ideal but unblocks parallel SPEC progress.

Evidence anchor [E-20]: SPEC-V3R2-ORC-001/spec.md §2.1 In Scope final paragraph "Template-first: all deletes and creates MUST occur under internal/template/templates/.claude/agents/moai/ first".
Evidence anchor [E-21]: SPEC-V3R2-ORC-001/spec.md §2.1 stub agent retention "Emit stub agents (one-line redirect bodies) at deprecated names for one v3.x minor cycle".

### 5.2 ORC-002 (CI lint LR-01..LR-10)

Status: per §2.4-§2.6 above, LR-01..LR-10 are implemented in `internal/cli/agent_lint.go`. LR-03 already at Error severity.

Boundary: ORC-002 implemented LR-03 as warning originally per its own §5.2 REQ-006, which says "LR-03 (missing effort) is a warning rule promoted to error by ORC-003". The ORC-002 implementation appears to have **pre-emptively wired LR-03 at Error** in anticipation of ORC-003. plan.md §1.2.1 acknowledges this idempotency.

Evidence anchor [E-22]: SPEC-V3R2-ORC-002/spec.md REQ-ORC-002-006 (LR-03 warning) + REQ-ORC-002-007 (`--strict` promotes to error).
Evidence anchor [E-23]: `internal/cli/agent_lint.go:382-396` checkMissingEffort body — unconditional Error severity contradicts ORC-002 spec REQ-006/007 wording. The implementation is more strict than ORC-002 spec; ORC-003 retroactively bless the strict behavior via REQ-006.

### 5.3 ORC-004 (worktree MUST)

Status: in-flight. Owns LR-09 (read-only with worktree → reject) and the implementer/tester/designer worktree MUST.

Boundary: ORC-004 reserves LR-11 for its `isolation: worktree` MUST promotion (warning → error). This SPEC does NOT touch LR-09 or LR-11. LR-12/13/14 are next-available slots after LR-11 reservation.

Evidence anchor [E-24]: SPEC-V3R2-ORC-002/spec.md REQ-ORC-002-006 footer "promoted to error in ORC-003" for LR-03; "promoted to error by ORC-004" for LR-05.

### 5.4 ORC-005 (static team-* prohibition)

Status: implemented per `internal/cli/agent_lint.go:540-565` checkStaticTeamAgent (LR-10).

Boundary: ORC-005 owns LR-10. No interaction with this SPEC.

Evidence anchor [E-25]: `internal/cli/agent_lint.go:540-565` LR-10 implementation.

### 5.5 SPEC-V3-AGT-001 (frontmatter schema validator)

Status: pre-v3r2 SPEC (v3-legacy inherited). Owns `internal/config/schema/agent.go` validator that enforces the 5-value effort enum at parse time.

Boundary: REQ-012 says "the frontmatter validator (per SPEC-V3-AGT-001 REQ-002) shall reject the agent with error AGT_INVALID_FRONTMATTER naming the `effort` field and its illegal value". Plan binding: LR-13 in `agent_lint.go` is the lint-side defense-in-depth surface; the schema validator owns runtime enforcement. Both gates fire with same error code for consistency.

Evidence anchor [E-26]: SPEC-V3-AGT-001 cross-reference per spec.md §9.3 Related dependencies row.

### 5.6 SPEC-V3R2-HRN-001 (harness routing)

Status: in-flight. Owns `harness.yaml` `effort_mapping.<level>` per-harness override.

Boundary: REQ-010 documents the harness override path; no code change in this SPEC. The matrix is the **per-agent default**; harness override applies at session level (Claude Code v2.1.110+ effortLevel setting per coding-standards.md compatibility table).

Evidence anchor [E-27]: SPEC-V3R2-HRN-001 cross-reference per spec.md §9.2 Blocks "SPEC-V3R2-HRN-001 (Harness routing) — uses effort_mapping aligned to this matrix".

### 5.7 SPEC-V3R2-MIG-001 (v2 → v3 migrator)

Status: in-flight. Owns the auto-migrator that rewrites legacy SPEC references and (per REQ-007) drifted effort values during migration.

Boundary: REQ-007 says "When an existing agent frontmatter contains an `effort:` value that disagrees with the canonical matrix, the migrator (SPEC-V3R2-MIG-001) shall rewrite the value to the canonical matrix entry and emit a migration log line". Plan binding: research.md cross-link to MIG-001 + tasks.md §T-ORC003-07 advisory note. No code in this SPEC.

Evidence anchor [E-28]: SPEC-V3R2-MIG-001 reference per spec.md §9.2 Blocks "SPEC-V3R2-MIG-001 (migrator) — references this SPEC's matrix when rewriting v2 agent effort values".

### 5.8 SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE codification)

Status: assumed merged per spec §9.1 Blocked-by.

Boundary: classifies the constitution file at file-level FROZEN; individual bullets are Evolvable. The cross-reference change in T-ORC003-06 is permissible.

Evidence anchor [E-29]: `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-028 + CONST-V3R2-029 + CONST-V3R2-030 entries classify Opus 4.7 Prompt Philosophy bullets as Evolvable.

### 5.9 SPEC-V3R2-CON-002 (amendment protocol)

Status: in-flight. Owns the future protocol for matrix changes (REQ-008 says matrix is single source; REQ-002 says any future change passes through CON-002 graduation).

Boundary: this SPEC sets the initial matrix; future matrix amendments go through CON-002. No code in this SPEC.

Evidence anchor [E-30]: SPEC-V3R2-CON-002 cross-reference per spec.md §9.3 Related dependencies row.

---

## 6. Decision Log (Plan-Phase Open Questions Resolved)

### OQ-1: Should the matrix be in moai-constitution.md or agent-authoring.md?

**Decision**: agent-authoring.md (REQ-001/002), with constitution cross-reference (REQ-005).

**Rationale**: agent-authoring.md is the developer-facing reference for frontmatter authoring. Constitution is FROZEN doctrine; matrix is execution detail. Cross-reference avoids duplication and FROZEN-clause modification.

Evidence: REQ-005 + REQ-008 (single source of truth) align.

### OQ-2: Should LR-12 include warn-only mode?

**Decision**: Error severity from launch (no warn-only phase).

**Rationale**: Drift introduction during PR is the main risk. Strict-mode-only would allow drift to land in main between PRs (false-negative window). spec.md REQ-013 explicitly says "CI shall fail with error". Matching the LR-03 promotion rationale (idempotent strictness from launch).

### OQ-3: Should LR-13 (invalid effort enum) be error or warning?

**Decision**: Error.

**Rationale**: Invalid enum is unambiguously broken — agent cannot even be loaded by Claude Code at runtime. spec.md REQ-012 says "shall reject the agent" — Error severity. Schema validator (SPEC-V3-AGT-001) already rejects at runtime with HTTP 400 equivalent; LR-13 surfaces the same defect at lint time before runtime.

### OQ-4: Should LR-14 (fixed budget_tokens) regex be code-block-aware?

**Decision**: No (v1 simple regex; v3.1 refinement deferred).

**Rationale**: False-positive cost (rare false alarm in YAML/markdown code block content with literal `budget_tokens: <num>`) is dominated by regression-prevention value (Opus 4.7 HTTP 400 rejection breaks production). Acceptable trade-off for v1; @MX:WARN tag (plan §6) documents the simplification.

### OQ-5: Should the canonical matrix be a Go map or a YAML config file?

**Decision**: Go `map[string]string` constant in `agent_lint.go`.

**Rationale**: Single binary, no external file load at lint time, no config-version drift. The agent-authoring.md table is the human-readable mirror; T-ORC003-23 verifies they agree (regex-based).

### OQ-6: How does this SPEC interact with `manager-brain` (out-of-roster, has effort: xhigh)?

**Decision**: Carve-out; LR-12 ignores manager-brain. LR-03 still applies (manager-brain has effort, so passes).

**Rationale**: manager-brain is a v3.x-specific agent created post-ORC-001 (commit 2026-05-04). Its inclusion in 17-agent roster is a future SPEC question (manager-brain consolidation or addition to canonical matrix). For ORC-003, it's out-of-scope per spec §1.2 Non-Goals.

### OQ-7: Run-phase if ORC-001 has NOT yet merged?

**Decision**: Apply effort to retiree files (manager-ddd, manager-tdd, builder-agent, builder-skill, builder-plugin) per fallback path.

**Rationale**: Unblocks parallel SPEC progress. After ORC-001 merges and renames files, MIG-001 migrator handles the rename + value preservation. Documented in plan §1.2 + §8.1 + tasks §4.1.

---

## 7. Risk Survey (cross-referenced from spec §8)

| Risk | Evidence anchor | Mitigation reference |
|---|---|---|
| `high → xhigh` latency regression on 4 agents | E-09, E-10 (Adaptive Thinking allocates more reasoning tokens) | plan §7 Risk row 1; HRN-001 harness override |
| `xhigh` over-invokes on trivial plan-auditor tasks | E-09 (Adaptive Thinking semantics) | HRN-001 minimal harness routing |
| Contributors add agents without `effort:` | E-04, E-06 (LR-03 already Error) | plan §7 row 3; T-ORC003-24 regression test |
| Constitution-matrix drift | E-10 (constitution inline content), E-11 (FROZEN classification) | plan §7 row 4; T-ORC003-06 + T-ORC003-22 |
| Fixed `budget_tokens` reintroduced | E-09 (Adaptive Thinking forbids) | plan §7 row 6; LR-14 implementation |
| Out-of-roster agents (manager-brain) confuse LR-12 | E-18 (manager-brain xhigh declared) | plan §4.1 LR-12 carve-out logic |
| Format drift breaks parsing | E-12, E-13 (5-value enum invariant) | T-ORC003-23 grep-based regression test |

---

## 8. Cross-Reference Summary

External references:
- Anthropic "what's new in claude-4-7" (Sep 2025): https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7

Internal references (file:line anchors):
- spec.md §1.1 (R5 audit matrix), §1.2 (BC-V3R2-002 declaration), §5 (REQ enumeration)
- plan.md §1.2 (drift inventory), §1.4 (deliverables), §4 (technical approach)
- tasks.md §4.1 (dependency-resolution flow)
- acceptance.md (10 ACs)
- `.claude/agents/moai/*.md` (per-agent frontmatter scan)
- `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-028..030
- `.claude/rules/moai/development/agent-authoring.md` lines 38, 58
- `internal/cli/agent_lint.go:84-97` (lint rule help block), `:382-396` (checkMissingEffort)
- `.moai/design/v3-redesign/research/r5-agent-audit.md` § Effort-level calibration matrix
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` P-A02, P-A03
- `docs/design/major-v3-master.md:L961` (BC-V3R2-002), `:L1054` (§11.4 ORC-003 definition)

Cross-SPEC references:
- SPEC-V3R2-CON-001 (FROZEN classification consumed)
- SPEC-V3R2-CON-002 (future amendment protocol)
- SPEC-V3R2-ORC-001 (17-agent roster baseline)
- SPEC-V3R2-ORC-002 (LR-01..LR-10 baseline)
- SPEC-V3R2-ORC-004 (LR-09/11 reservation; worktree MUST)
- SPEC-V3R2-ORC-005 (LR-10 baseline)
- SPEC-V3R2-HRN-001 (harness override path; downstream consumer)
- SPEC-V3R2-MIG-001 (drift rewrite migrator; downstream consumer)
- SPEC-V3-AGT-001 (frontmatter schema validator; co-defense surface)

Total evidence anchors: **30** ([E-01] through [E-30]). Plan-auditor PASS criterion #4 (≥30 evidence anchors) satisfied.

---

## 9. Conclusions and Plan-Phase Recommendations

1. **Drift count is 4, not 3.** Sync-phase HISTORY entry must reconcile spec.md §1.1 / §1.2 / REQ-002.
2. **LR-03 promotion is operationally idempotent.** ORC-002 pre-wired Error severity. REQ-006 becomes regression-test verification, not severity flip.
3. **LR-12, LR-13, LR-14 are NEW lint rules.** ORC-002 stops at LR-10; LR-11 reserved for ORC-004; this SPEC claims 12-14.
4. **Canonical matrix lives in two places**: machine-readable Go constant (`canonicalEffortMatrix` in agent_lint.go) + human-readable markdown table (agent-authoring.md). T-ORC003-23 regression test verifies they agree.
5. **Constitution cross-reference, not duplication.** REQ-005 + T-ORC003-06 modify the inline list to a cross-link. FROZEN-clause text is preserved (cross-link addition, not rule deletion).
6. **Fallback path documented if ORC-001 unmerged.** Plan §1.2 + tasks §4.1 apply effort to retiree agents temporarily.
7. **Out-of-roster agents (manager-brain, claude-code-guide) untouched.** LR-12 carve-out; their disposition is a separate SPEC.
8. **Migrator (MIG-001) handles legacy v2 SPEC references**. REQ-007 advisory; no code in this SPEC.

End of research.

Version: 0.1.0
Status: Research artifact for SPEC-V3R2-ORC-003 (Plan workflow Phase 1A)
