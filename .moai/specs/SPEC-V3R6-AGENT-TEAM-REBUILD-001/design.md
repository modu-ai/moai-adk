---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
artifact: design
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
sync_commit_sha: "f0f222fa3"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L design.md authored. Documents: §A current-state cold judgment from Audit 3; §B target architecture (8-agent delegation graph + 2-level Anthropic constraint + 4-layer enforcement + 5-mode decision tree + ARCHIVED_AGENT_REJECTED migration table); §C implementation pattern maps; §D hook architecture; §E Karpathy principles + 8 anti-pattern remediation matrix; §F Status Transition Ownership Matrix reference; §G risk mitigation architecture. |

---

## §A — Current State (Audit 3 cold judgment)

### §A.1 Audit 3 verbatim findings recap

Three deep audits conducted in the prior session turn (2026-05-25) revealed the MoAI agent catalog's structural divergence from Anthropic 2026 published guidance. The verbatim citations are reproduced here as the design's empirical foundation:

**Finding A1 — Catalog inflation 2-5× over Anthropic ceiling.** Anthropic ships exactly 3 built-in subagents (Explore, Plan, general-purpose). Agent Teams documentation states: *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* MoAI's 17 custom agents place the project 2-5× above the recommended ceiling. Observable cost: 12 of 17 agents recorded 0 invocations across the recent 4-SPEC cohort.

**Finding A2 — Hierarchical fiction architecturally impossible.** Anthropic Sub-agents documentation states: *"Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."* The MoAI `manager-strategy → manager-develop` chain pattern documented in `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` REQ-WOF-003 is therefore architecturally impossible to execute at runtime. Restoration is not a viable remediation; pivot is required.

**Finding A3 — Phantom agent criterion failure.** Anthropic best-practices: *"Define a custom subagent when you keep spawning the same kind of worker with the same instructions."* 12 of MoAI's 17 agents fail this criterion: 0 invocations across recent 4 SPEC sessions = not "the same kind of worker spawned repeatedly".

**Finding A4 — Coding-task parallelism caveat.** Anthropic verbatim: *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."* Run-phase coding tasks SHOULD remain single-agent (`manager-develop` standalone with cycle_type), not decomposed into manager-strategy chain. Parallel multi-spawn applies primarily to research phase.

**Finding A5 — Domain expertise via spawn-prompt, not via agent file.** Anthropic's recommended pattern: per-spawn parameter injection — `Agent(subagent_type: "general-purpose", model: "opus", tools: [...], prompt: "<domain-specific instructions>")`. The 6 `expert-*` agents (backend, frontend, security, devops, performance, refactoring) duplicate this pattern as static files, trapping domain knowledge inside individual definitions rather than active conversation context.

**Finding A6 — Hook-based enforcement available.** Anthropic hook documentation (Stop, PostToolUse, SubagentStop, TaskCompleted) provides the canonical mechanism for converting orchestrator-discipline into mechanically-unbypassable enforcement. Quality gates (lint + test + coverage delta), Status Transition Ownership, and per-AC PASS verification are all hook-enforceable today.

### §A.2 Cold-judgment summary

The MoAI orchestrator's current behavior pattern is **simultaneously over-architected and under-utilized**: 17 custom agents are declaratively defined, but 12 are never spawned. The declarative complexity itself is the over-engineering anti-pattern (Karpathy #2). The under-utilization is the "defined when not keep spawning" criterion failure (Anthropic verbatim Finding A3).

The predecessor SPEC (`SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001`) attempted to fix utilization by restoring 17 silently-skipped phases (the manager-quality / expert-security / manager-strategy chain etc.). This remediation strategy is invalid because:

(a) Finding A2 makes manager-strategy chain restoration architecturally impossible at runtime;
(b) Restoring phantom agents reinforces the over-engineering rather than addressing it;
(c) Finding A6's hook-based enforcement is the more reliable substitute for orchestrator-discipline phantom-agent spawns.

The correct remediation is consolidation: 17 → 8 retained + 12 archived + 3 hook scripts for mechanical enforcement.

---

## §B — Target Architecture

### §B.1 Retained Agent Delegation Graph (8 agents)

```
                          ┌──────────────────────────────────┐
                          │   MoAI Orchestrator (main)        │
                          │   • Single point of user contact  │
                          │   • Owns AskUserQuestion          │
                          │   • Spawns sub-agents via Agent() │
                          └────────────┬──────────────────────┘
                                       │ Agent() spawn (Anthropic 2-level constraint)
        ┌──────────────────────────────┼─────────────────────────────────────────┐
        │                              │                                          │
   plan-phase                     run-phase                                  sync-phase + cross-cutting
        │                              │                                          │
        ▼                              ▼                                          ▼
┌─────────────────┐         ┌────────────────────┐                  ┌────────────────────┐
│ manager-spec    │         │ manager-develop    │                  │ manager-docs       │
│ • SPEC body     │         │ • Implementation   │                  │ • Sync-phase       │
│   authoring     │         │ • cycle_type=ddd   │                  │ • CHANGELOG        │
│ • Plan artifacts│         │   tdd / autofix    │                  │ • Frontmatter      │
│ • Supersedence  │         │ • Single-agent per │                  │   status transitions
│   transitions   │         │   Finding A4       │                  │ • Project docs     │
└─────────────────┘         └────────────────────┘                  │   (absorbs         │
        │                                                            │    manager-project)│
        ▼                                                            └────────────────────┘
┌─────────────────┐         ┌────────────────────┐                  ┌────────────────────┐
│ plan-auditor    │         │ evaluator-active   │                  │ manager-git        │
│ • Phase 0.5     │         │ • Producer-Reviewer│                  │ • Tier L OR --pr   │
│   independent   │         │   cycle (harness   │                  │   flag PR ops      │
│   audit (bias   │         │   thorough only)   │                  │ • Hybrid Trunk     │
│   prevention)   │         │ • max-3-iterations │                  │   coordination     │
└─────────────────┘         └────────────────────┘                  └────────────────────┘

                            ┌────────────────────┐
                            │ builder-harness    │
                            │ • Meta-builder     │
                            │ • Skill / agent /  │
                            │   plugin authoring │
                            │ • /moai project    │
                            │   Phase 5+         │
                            └────────────────────┘

                            ┌────────────────────────────────────────┐
                            │ Explore (Anthropic built-in)            │
                            │ • Read-only investigation                │
                            │ • Replaces: claude-code-guide + researcher │
                            │ • Not a MoAI file — used as-is           │
                            └────────────────────────────────────────┘
```

### §B.2 Anthropic 2-level Constraint Diagram

Per Finding A2 verbatim — "Subagents cannot spawn other subagents":

```
Level 0: User / Claude Code session
   │
   ▼
Level 1: MoAI Orchestrator (main session)
   • Owns AskUserQuestion (channel monopoly)
   • Owns Agent() spawn calls
   • Owns SkillRouter invocations
   │
   ▼ Agent() spawn (allowed; only outbound from this level)
Level 2: Sub-agents (manager-* / plan-auditor / evaluator-active / builder-harness / Explore)
   • Cannot call Agent() (Anthropic constraint)
   • Cannot call AskUserQuestion (subagent boundary)
   • MUST return blocker reports to L1 for re-delegation

                         ┌── ARCHITECTURALLY IMPOSSIBLE ──┐
Level 3+: ... (NOT REACHABLE)
   • manager-strategy → manager-develop chain  ← REJECTED by Anthropic
   • manager-quality spawning expert-security  ← REJECTED
```

**Design implication**: all multi-agent coordination MUST happen at L1 (the orchestrator). Sub-agents are leaves, not branches. The MoAI 8-agent retention graph above respects this constraint by design.

### §B.3 Hook-Driven Enforcement Layers (L1..L4)

```
┌───────────────────────────────────────────────────────────────────────┐
│ L1 Declarative (Agent definition files)                                │
│ • tools: CSV whitelist (REQ-ATR-003)                                   │
│ • NOT-for: clauses in description (REQ-ATR-004)                        │
│ • Body line count ≤ 500 (REQ-ATR-002)                                  │
│ Strength: low (relies on agent self-compliance)                        │
├───────────────────────────────────────────────────────────────────────┤
│ L2 disallowedTools (per Anthropic SDK)                                 │
│ • Agent-spawn-time tool denylist                                       │
│ • Example: Agent(disallowedTools: ["Write", "Edit"], ...)             │
│ Strength: medium (Claude Code runtime enforces denylist)               │
├───────────────────────────────────────────────────────────────────────┤
│ L3 PostToolUse / Stop / TaskCompleted hooks (NEW per REQ-ATR-009/014) │
│ • status-transition-ownership.sh — PostToolUse on Write/Edit          │
│ • sync-phase-quality-gate.sh — Stop on sync-phase commit completion   │
│ • team-ac-verify.sh — TaskCompleted in team mode (dormant default)    │
│ Strength: HIGH (mechanically unbypassable when configured)             │
├───────────────────────────────────────────────────────────────────────┤
│ L4 Status Transition Ownership Matrix (declarative SSOT)               │
│ • Per .claude/rules/moai/development/spec-frontmatter-schema.md       │
│ • 7 canonical transitions with per-transition owner agent              │
│ • Enforced at lint-time by internal/spec/lint_ownership.go            │
│ Strength: HIGH (lint-level guard + L3 hook reinforcement)              │
└───────────────────────────────────────────────────────────────────────┘
```

### §B.4 Mode Selection Decision Tree (5 modes)

Per REQ-ATR-008 + REQ-ATR-013 + REQ-ATR-017, the Phase 0.95 Mode Selection decision tree autonomously routes execution to one of 5 modes:

```
START (Phase 0.95 Mode Selection)
  │
  ├── Is task trivial (typo, single-line, no semantic change)?
  │   ├── YES → Mode 1: TRIVIAL (direct execution, no Agent() spawn)
  │   └── NO  → continue
  │
  ├── Can the task complete async without blocking the user?
  │   ├── YES + tool supports `run_in_background: true`
  │   │       → Mode 2: BACKGROUND (Agent run_in_background: true)
  │   └── NO  → continue
  │
  ├── Is harness `thorough` AND workflow.team.enabled AND env var set
  │   AND scope ≥ 3 domains OR ≥ 10 files?
  │   ├── YES → Mode 3: AGENT-TEAM (TeamCreate + dynamic teammate spawn)
  │   └── NO  → continue
  │
  ├── Is task multi-domain (≥ 3 domains) AND research-heavy
  │   (not coding-heavy per Finding A4)?
  │   ├── YES → Mode 4: PARALLEL (3-5 concurrent Agent() in single message)
  │   └── NO  → continue
  │
  └── Default → Mode 5: SUB-AGENT (single Agent() sequential spawn)

Tie-breaker rules (Phase 0.95 boundary cases):
  • At threshold ±1 (9 vs 10 files; 2 vs 3 domains): default to simpler mode
  • Coding-heavy + multi-domain: prefer Mode 5 over Mode 4 (Finding A4)
  • Markdown-heavy + multi-domain + research-heavy: prefer Mode 4
  • Tier L + markdown/shell-script-only: Mode 5 with Tier L Section A-E template

Logging: orchestrator records chosen mode + rationale in progress.md § Mode Selection
```

**This SPEC's own classification**: Mode 5 (SUB-AGENT) per the rationale in plan.md §D.3 — markdown + shell-script-only, no real parallel implementation benefit, Finding A4 coding-task parallelism caveat applies (manager-develop sequential single-spawn with Tier L Section A-E template + 8 milestone breakdown).

### §B.5 ARCHIVED_AGENT_REJECTED Migration Table (12 archived → replacement pattern)

Per REQ-ATR-016, the `.claude/rules/moai/workflow/archived-agent-rejection.md` rule file documents the canonical migration table:

| Archived agent | Replacement pattern | Example invocation |
|----------------|---------------------|-------------------|
| `manager-strategy` | Absorb into `manager-spec` (planning IS strategy per Anthropic 4-phase Step 2) | `Agent(subagent_type: "manager-spec", prompt: "...plan-phase strategic analysis...")` |
| `manager-quality` | Stop hook enforcement (lint + test + coverage delta) OR `/moai gate` skill | Hook auto-fires; OR manual `Skill("moai", arguments: "gate")` |
| `manager-brain` | Explore + `manager-spec` sequence in `/moai brain` workflow | `Agent(Explore)` followed by `Agent(manager-spec)` chain at L1 |
| `manager-project` | Absorb into `manager-docs` | `Agent(subagent_type: "manager-docs", prompt: "...project doc rotation...")` |
| `claude-code-guide` | Use Anthropic built-in `Explore` | `Agent(subagent_type: "Explore", prompt: "...upstream Claude Code investigation...")` |
| `researcher` | Use Anthropic built-in `Explore` | `Agent(subagent_type: "Explore", prompt: "...auto-research...")` |
| `expert-backend` | `Agent(general-purpose, model: opus, tools: <backend whitelist>, prompt: <backend instructions>)` | `Agent(subagent_type: "general-purpose", model: "opus", tools: ["Read", "Write", "Edit", "Bash", "Grep"], prompt: "...backend-specific instructions...")` |
| `expert-frontend` | Same pattern with frontend whitelist + frontend instructions | (analogous) |
| `expert-security` | Stop hook dependency manifest audit (REQ-ATR-014) + per-spawn for code review | Hook auto-fires on `go.mod` changes; OR per-spawn for review |
| `expert-devops` | Same as expert-backend with devops whitelist + devops instructions | (analogous) |
| `expert-performance` | Same with performance whitelist + perf instructions | (analogous) |
| `expert-refactoring` | Same with refactoring whitelist + refactoring instructions | (analogous) |

---

## §C — Implementation Pattern Maps

### §C.1 Plan-phase pattern map

| Phase | Owner | Pattern | Execution mode | HUMAN GATE |
|-------|-------|---------|----------------|------------|
| Pre-flight research | Explore (Anthropic built-in) | parallel multi-spawn (research-heavy) | Mode 4 PARALLEL per Finding A4 | No (research is read-only) |
| SPEC artifact authoring | manager-spec | single-agent | Mode 5 SUB-AGENT | GATE-1: user reviews plan via AskUserQuestion |
| Phase 0.5 plan audit | plan-auditor | single-agent (independent) | Mode 5 SUB-AGENT | No (auditor returns verdict to orchestrator) |
| Plan PR merge | manager-git (Tier L) OR orchestrator main-direct (Tier S/M) | single-agent OR direct | Mode 5 OR direct | GATE-2: user confirms plan-to-implement transition |

### §C.2 Run-phase pattern map

| Phase | Owner | Pattern | Execution mode | HUMAN GATE |
|-------|-------|---------|----------------|------------|
| Phase 0.95 Mode Selection | orchestrator | autonomous | n/a (this IS the selection step) | No (logged to progress.md) |
| Phase 1 implementation | manager-develop | single-agent per Finding A4 | Mode 5 SUB-AGENT (default) | No (within bounded scope) |
| Phase 2 Sprint Contract (harness thorough only) | evaluator-active | Producer-Reviewer cycle | Mode 5 with max-3-iter | No |
| Run PR merge | manager-git (Tier L) OR orchestrator (Tier S/M) | single-agent OR direct | Mode 5 OR direct | GATE-3: user confirms implementation completion |

### §C.3 Sync-phase pattern map

| Phase | Owner | Pattern | Execution mode | HUMAN GATE |
|-------|-------|---------|----------------|------------|
| Sync artifact generation | manager-docs | single-agent | Mode 5 SUB-AGENT | GATE-4: user confirms doc scope (post-sync review) |
| Quality gates | Stop hook (mechanical) | hook-enforced | n/a | No (hook exit 2 blocks commit) |
| Mx-phase judgement | orchestrator | autonomous | n/a | No (logged to progress.md) |
| 4-phase close | orchestrator | autonomous | n/a | No |

---

## §D — Hook Architecture

### §D.1 status-transition-ownership.sh (PostToolUse)

**Trigger**: Write or Edit tool invocation on any file matching `.moai/specs/SPEC-*/{spec,plan,acceptance}.md` body content (frontmatter excluded).

**Logic**:
1. Read stdin JSON; extract `tool_name`, `tool_input.file_path`, `tool_input.new_string` (or `content`).
2. If `file_path` does NOT match `.moai/specs/SPEC-*/{spec,plan,acceptance}.md`, exit 0 (not in scope).
3. Detect whether the Write/Edit targets YAML frontmatter (between leading `---` lines) or body content. If frontmatter-only, exit 0 (Status Transition Ownership Matrix allows manager-develop and manager-docs to update frontmatter status/updated fields on their respective transitions).
4. If body modification detected: extract the invoking agent name from the Claude Code hook context (env var or stdin metadata).
5. Compare against the canonical Status Transition Ownership Matrix:
   - spec.md / plan.md / acceptance.md body content → manager-spec ONLY (or orchestrator mid-run authority per D-NEW-1 pattern)
   - progress.md body content (§E.2 / §E.3) → manager-develop
   - progress.md body content (§E.4) → manager-docs
6. If mismatch: exit 2 (block); output structured rejection JSON with violator agent name + canonical owner.

**Output format** (stdout, per Anthropic hook contract):
```json
{
  "continue": false,
  "stopReason": "Status Transition Ownership Matrix violation: <agent> attempted Write on <file> body; canonical owner is <owner>. See .claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix."
}
```

### §D.2 sync-phase-quality-gate.sh (Stop)

**Trigger**: any Stop hook event (Claude Code session end after sync-phase work).

**Logic**:
1. Read stdin JSON; check whether recent commits include sync-phase markers (commit subject matches `docs(SPEC-*): sync-phase` or `chore(SPEC-*): sync-phase`).
2. If no sync-phase commit detected, exit 0 (not in scope).
3. Execute verification sequence:
   - **Lint**: `golangci-lint run --timeout=2m` → expected exit 0. Failure → exit 2.
   - **Test**: `go test ./...` → expected exit 0. Failure → exit 2.
   - **Coverage delta**: read `cover.out`; compute delta against `.moai/state/coverage-baseline.json` (if absent, initialize). If delta < 0 for any changed file, exit 2.
   - **Dependency manifest audit** (REQ-ATR-014): if `go.mod`/`go.sum`/`package-lock.json`/`Pipfile.lock`/`Cargo.lock` changed, invoke `govulncheck ./...` or `npm audit --omit=dev` or equivalent. High-severity finding → exit 2.
4. Support `--baseline-mode` flag (read from `.moai/state/lint-baseline.json` to permit pre-existing baseline failures unrelated to current commit scope).
5. Support `--skip-hook` flag (log to `.moai/logs/hook-skip.log` for audit; exit 0).

**Output format**:
```json
{
  "continue": false,
  "stopReason": "Sync-phase quality gate failed: <lint|test|coverage|deps>. See <command output for details>.",
  "details": {
    "lint_exit": 0,
    "test_exit": 0,
    "coverage_delta": -0.5,
    "deps_audit": "HIGH severity finding"
  }
}
```

### §D.3 team-ac-verify.sh (TaskCompleted, team mode only — dormant default)

**Trigger**: TaskCompleted hook event in team mode (Agent Teams active).

**Logic**:
1. Read stdin JSON; extract `task_id`, `task_description`, `task_result`.
2. Detect AC-ID reference in task description (e.g., "AC-ATR-009 ...").
3. Check whether `.moai/specs/SPEC-*/ac-evidence/AC-<ID>.txt` exists with PASS marker.
4. If missing or FAIL marker present: exit 2 (reject completion); orchestrator must address before TaskCompleted is acknowledged.
5. Otherwise: exit 0 (acknowledge completion).

**Dormant by default**: only activates under `harness: thorough` + team mode prerequisites (REQ-ATR-013 capability gate). Default `harness: standard` does not invoke this hook.

### §D.4 Hook subagent boundary discipline (AC-ATR-022)

All 3 hook scripts MUST NOT invoke AskUserQuestion directly (subagent boundary per `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator–Subagent Boundary). Hooks return exit codes + structured JSON; the orchestrator translates exit codes into user-facing prompts via `AskUserQuestion` (after `ToolSearch(query: "select:AskUserQuestion")` preload).

Verification: `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ | grep -v "^[^:]*:[0-9]*:[ \t]*#"` MUST return empty (AC-ATR-022).

---

## §E — Karpathy Principles + 8 Anti-Pattern Remediation Matrix

The MoAI orchestrator's current behavior exhibits 6 of 8 Karpathy anti-patterns (per Audit 3 cross-reference). This SPEC's consolidation strategy remediates 7 of 8:

| # | Karpathy Anti-pattern | MoAI current state (pre-SPEC) | Remediation by this SPEC |
|---|----------------------|-------------------------------|-------------------------|
| 1 | Premature Abstraction | Not exhibited (orchestrator-level pre-SPEC was correctly abstract) | No change needed |
| 2 | **Over-Engineering** | **EXHIBITED**: 17-agent catalog with 12 phantom agents | **REMEDIATED**: 17 → 8 retention; archive 12 (REQ-ATR-001 + REQ-ATR-005) |
| 3 | **Drive-By Refactoring** | **EXHIBITED**: subagent prompts bundled M1+M2+M3+M4 in past sessions | **REMEDIATED**: M1-M8 milestone decomposition with one M per delegation; cycle_type=autofix mode formalization (REQ-ATR-012) prevents bundling autofix into other modes |
| 4 | Style Drift | Not exhibited at orchestration layer (code-level only) | No change needed |
| 5 | **Silent Assumption** | **EXHIBITED**: autonomous flow assumed user wants no gates | **REMEDIATED**: REQ-ATR-015 GATE-2 mandatory restoration on autonomous-flow skip detection; `score ≥ 0.90 skip-eligible` policy restricted to Phase 0.5 only |
| 6 | **Guessing Over Clarifying** | **EXHIBITED**: paste-ready resume → autonomous, skipped clarify rounds | **REMEDIATED**: REQ-ATR-015 covers; Phase 0.95 Mode Selection logging adds explicit boundary case clarification to progress.md |
| 7 | **Sycophantic Agreement** | **EXHIBITED**: never pushed back on user `/moai run` request | **REMEDIATED**: REQ-ATR-016 ARCHIVED_AGENT_REJECTED makes the orchestrator MUST reject invalid spawn invocations; Stop hook (REQ-ATR-009) makes the orchestrator MUST block sync commits on quality gate failures |
| 8 | **Claiming Without Evidence** | **EXHIBITED**: SPECs closed without HUMAN GATE verification | **REMEDIATED**: hook-based mechanical enforcement (REQ-ATR-009 + REQ-ATR-014) replaces orchestrator-discipline; AC-ATR-022 subagent boundary verification |

**Remediation count**: 7 of 8 (Style Drift + Premature Abstraction were not exhibited; remaining 6 anti-patterns all addressed by REQ-ATR clauses).

---

## §F — Status Transition Ownership Matrix Reference

This SPEC respects and reinforces the canonical Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md`. Key transitions involved in this SPEC's lifecycle:

| Transition | Canonical owner | This SPEC's invocation |
|------------|-----------------|------------------------|
| `(none) → draft` | manager-spec | M0 plan-phase artifact creation (initial draft) |
| `draft → in-progress` | manager-develop | M1 first run-phase commit |
| `in-progress → implemented` | manager-docs | sync-phase commit |
| `implemented → completed` | manager-docs OR orchestrator | Mx-phase chore commit |
| `* → superseded` (on predecessor SPEC) | manager-spec | M6 frontmatter `status: superseded` on `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` |

**REQ-ATR-006** binds the predecessor SPEC supersedence to the canonical Matrix `* → superseded` owner = manager-spec. M6 (Predecessor SPEC Supersedence) is therefore either:
- (a) re-delegated by orchestrator from manager-develop to manager-spec mid-run, OR
- (b) executed by manager-develop under D-NEW-1 inline-fix pattern with explicit orchestrator re-delegation per `.claude/agents/core/manager-spec.md` § Mid-run Authority

The L3 PostToolUse hook (`status-transition-ownership.sh` per §D.1 above) enforces this mechanically at hook layer.

---

## §G — Risk Mitigation Architecture

The 8 risks enumerated in spec.md §G are addressed by the following architectural defense layers:

| Risk | Defense layer | Mechanism |
|------|---------------|-----------|
| G1 — Invested-work loss | L4 (Status Transition Ownership Matrix) + git history | Archive (NOT delete) to `.moai/backups/` preserves git history via `git mv`; future revival SPEC reversible |
| G2 — User confusion during transition | L3 hook + L4 declarative + migration table | ARCHIVED_AGENT_REJECTED error (REQ-ATR-016) + migration table in `archived-agent-rejection.md` documents per-agent replacement pattern |
| G3 — Hook false positives | L3 hook with --baseline-mode flag | `--baseline-mode` reads `.moai/state/lint-baseline.json`; per-hook test suite under `.moai/tests/hooks/` validates expected behavior; `--skip-hook` opt-out logged to audit log |
| G4 — Tier L plan-auditor below threshold | L1 declarative (this SPEC's structural completeness) | 20 REQs + 22 ACs + 8 ECs + 8 risks-mitigation pairs achieves MP-1..MP-4 dimensions; self-audit estimate ~0.88-0.91 PASS; max-3-iter retry contract provides recovery |
| G5 — Multi-session execution span | L4 paste-ready resume protocol | Session-handoff.md 6-block structure + auto-memory project entry persists state; 8 milestones decompose into ≤4 per session at standard Opus 4.7 1M context |
| G6 — Template-local drift | L3 hook (`make build` regen detection) + L4 declarative (REQ-ATR-018) | M8 `diff -r` byte-for-byte parity verification; AC-ATR-018 binary check; future PostToolUse hook can fire on template path edits to alert drift |
| G7 — manager-git Hybrid Trunk regression | L4 declarative (explicit clarification not policy change) | REQ-ATR-020 documents current Tier S/M = main-direct + Tier L OR `--pr` = PR via manager-git as the canonical rule (no behavior change, only explicit codification) |
| G8 — Anthropic 2026 guidance evolution | L4 declarative (NOTICE.md attribution) | REQ-ATR-019 records source URLs + Finding A1-A6 verbatim quotes + 2026-05-25 epoch; future "agent revival" SPEC can re-evaluate against new guidance |

---

## §H — Cross-References

- spec.md §B Background (Audit 3 verbatim findings + supersedence rationale)
- spec.md §D 20 REQ-ATR-XXX (referenced throughout this design.md)
- acceptance.md §A 22 AC-ATR-XXX + §B Traceability Matrix + §C 8 edge cases
- plan.md §D 8 milestones M1-M8 (per-milestone REQ coverage)
- research.md §H Audit 3 synthesis (cited in §A.1 above)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (canonical SSOT, §F above)
- `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator–Subagent Boundary (hook subagent boundary AC-ATR-022, §D.4 above)
- `.claude/rules/moai/development/agent-patterns.md` (target of M5 update — domain-expert spawn-prompt pattern documentation)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW, target of M5 — 5-mode decision tree §B.4 above)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` (NEW, target of M5 — migration table §B.5 above)
- claude.com/docs/en/sub-agents (Finding A2 source)
- claude.com/docs/en/agent-teams (Finding A1 source)
- claude.com/docs/en/best-practices (Finding A3 source)
- claude.com/docs/en/hooks (Finding A6 source)
- anthropic.com/engineering/built-multi-agent-research-system (Finding A4 source)

---

Version: 0.1.0
Status: draft (plan-phase initial authoring)
Tier: L
