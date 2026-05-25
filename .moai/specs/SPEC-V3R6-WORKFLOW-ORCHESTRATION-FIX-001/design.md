---
id: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001
artifact: design
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
plan_commit_sha: "<pending>"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L design artifact — architecture rationale + delegation graph + state machine + 5-mode autonomous decision tree |

---

## §A — Current State Analysis

### §A.1 Verdict (cite research.md §A)

The MoAI orchestrator's current behavior violates the spirit of every documented orchestration framework it inherits. Across the recent 4-SPEC cohort (`SKILL-GEARS-ALIGN-001`, `PLAN-AUDITOR-GEARS-ALIGN-001`, `FOUNDATION-CORE-GEARS-ALIGN-001`, `WORKFLOW-PLAN-GEARS-ALIGN-001`), 17 of the 25 documented workflow phases were silently skipped — a 70-93% phase-skip rate depending on the SPEC.

### §A.2 Six measured findings

| # | Finding | Quantitative evidence |
|---|---------|----------------------|
| 1 | sync-phase manager-quality / expert-security / manager-develop coverage never invoked | 0 invocations across 4 SPECs (4-SPEC sample) |
| 2 | run-phase manager-strategy → manager-develop chain collapsed to flat single-spawn | 4-of-4 SPECs used single manager-develop spawn |
| 3 | Sub-skill on-demand loading bypassed — orchestrator Read()s SKILL.md body directly | 4-of-4 SPECs (router not invoked) |
| 4 | 5 HUMAN GATE decision points silently skipped | 5×4=20 skipped gate invocations |
| 5 | plan-phase Explore + research.md + GitHub Issue + BODP audit all missing | research.md present only when manually authored (1-of-4 SPECs with this current SPEC's research.md being the exception) |
| 6 | Autonomous mode selection (5 modes) never implemented; sequential single-spawn always | 4-of-4 SPECs default to sequential |

### §A.3 Anti-pattern mapping (Karpathy 8 anti-patterns)

The MoAI orchestrator currently exhibits 6 of 8 Karpathy anti-patterns at the orchestration layer (research.md §B.3 verbatim):

- **#2 Over-Engineering**: 17-agent catalog mostly unused; complex workflow architecture not executed
- **#3 Drive-By Refactoring**: Subagent prompts bundle unrelated milestones M1+M2+M3+M4
- **#5 Silent Assumption**: Autonomous flow assumes user wants no gates
- **#6 Guessing Over Clarifying**: Paste-ready resume → autonomous, skips clarify rounds
- **#7 Sycophantic Agreement**: Never pushes back on user `/moai run` request
- **#8 Claiming Without Evidence**: Claims SPEC closed without HUMAN GATE verification

This SPEC remediates anti-patterns #2 / #3 / #5 / #6 / #8 via REQ-WOF-001 (HUMAN GATEs) + REQ-WOF-003 (chain restoration) + REQ-WOF-005 (router discipline) + REQ-WOF-011 (AskUserQuestion HARD).

---

## §B — Target State Architecture

### §B.1 Delegation Graph (Anthropic-aligned hierarchical pattern)

The restored delegation graph follows the harness `Hybrid-orchestrator` template per research.md §B.3 (orchestrator-templates.md). The graph below renders the canonical flow for `/moai plan → /moai run → /moai sync`; each arrow represents an orchestrator-mediated `Agent()` spawn with context passing.

```
                          ┌──────────────┐
                          │  User input  │
                          └──────┬───────┘
                                 │ /moai plan "feature"
                                 ▼
        ┌────────────────────────────────────────────────────────┐
        │                  Plan Phase                            │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 0 — Parallel Research (R7)               │   │
        │  │  ┌───────────────┐  ┌───────────────┐           │   │
        │  │  │  Explore #1   │  │  Explore #2   │  ... ×N   │   │
        │  │  │  (read-only)  │  │  (read-only)  │           │   │
        │  │  └───────┬───────┘  └───────┬───────┘           │   │
        │  │          └────────┬─────────┘                   │   │
        │  │                   ▼                              │   │
        │  │      orchestrator synthesizes research.md       │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 1 — Socratic Interview (clarify)         │   │
        │  │  orchestrator → AskUserQuestion (≤4 rounds)     │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 2 — manager-spec spawn                   │   │
        │  │  → authors spec.md + plan.md + acceptance.md    │   │
        │  │    + design.md (Tier L) + tier judgment         │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Decision Point 1 (HUMAN GATE) — AskUserQuestion│   │
        │  │  → user confirms tier + scope before commit     │   │
        │  └─────────────────────────────────────────────────┘   │
        └────────────────────────┬───────────────────────────────┘
                                 │ commit + push (Hybrid Trunk default; PR for Tier L)
                                 ▼
        ┌────────────────────────────────────────────────────────┐
        │                  Run Phase                             │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 0.5 — Plan Audit Gate (plan-auditor)     │   │
        │  │  → verdict ≥ Tier threshold OR retry max-3      │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 0.95 — Mode Selection (NEW per REQ-004)  │   │
        │  │  → autonomous 5-mode decision (see §B.3 tree)   │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Plan Approval (HUMAN GATE) — AskUserQuestion   │   │
        │  │  → user confirms mode + run-phase entry         │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 1 — manager-strategy (analysis-only HARD)│   │
        │  │  → emits tasks.md (M1-M6+ decomposition)        │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 2.0 — evaluator-active Sprint Contract   │   │
        │  │  (gated by harness: thorough AND tier M/L)      │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 2 — manager-develop                      │   │
        │  │  → per-milestone spawn (separate Agent() calls) │   │
        │  │  → DDD/TDD per quality.yaml development_mode    │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 2.5 — GitHub Issue auto-creation         │   │
        │  │  Phase 2.8 — BODP audit trail                   │   │
        │  └─────────────────────────────────────────────────┘   │
        └────────────────────────┬───────────────────────────────┘
                                 │ commit + push
                                 ▼
        ┌────────────────────────────────────────────────────────┐
        │                  Sync Phase                            │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 0 — Working-tree pre-flight              │   │
        │  │  Phase 0.1 — Test suite full run                │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  GATE 1 (HUMAN GATE) — AskUserQuestion          │   │
        │  │  → working-tree-and-tests confirmation          │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 0.5.4 — manager-quality (TRUST 5)        │   │
        │  │  Phase 0.55 — expert-security manifest audit    │   │
        │  │              (always-runs HARD)                  │   │
        │  │  Phase 0.7 — manager-develop coverage           │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 1-3 — manager-docs sync artifacts        │   │
        │  │  → CHANGELOG + frontmatter + progress.md §E.4   │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  GATE 2 (HUMAN GATE) — AskUserQuestion          │   │
        │  │  → doc scope confirmation before commit         │   │
        │  └─────────────────────────────────────────────────┘   │
        │                                                        │
        │  ┌─────────────────────────────────────────────────┐   │
        │  │  Phase 4 — manager-git (Tier L) OR Hybrid Trunk │   │
        │  │  → PR creation OR main-direct push              │   │
        │  └─────────────────────────────────────────────────┘   │
        └────────────────────────┬───────────────────────────────┘
                                 │ Mx + 4-phase close
                                 ▼
        ┌────────────────────────────────────────────────────────┐
        │                  Mx + Close                            │
        │  Step C SKIP-judge → §E.5 audit-ready signal           │
        │  4-phase close → status: implemented → completed       │
        │  Decision Point 2 (HUMAN GATE optional)                │
        └────────────────────────────────────────────────────────┘
```

### §B.2 State Machine

The orchestrator transitions through 16 distinct workflow states. State transitions are gated by either deterministic checks (audit verdicts, scope thresholds) or HUMAN GATE confirmations (AskUserQuestion responses).

```
Initial → Plan Phase 0 (Research)
       → Plan Phase 1 (Clarify)
       → Plan Phase 2 (Author SPEC artifacts)
       → Decision Point 1 (HUMAN GATE)        ← AskUserQuestion
       → Plan-PR Merged
       ↓
       Run Phase 0.5 (Plan Audit Gate)
       → Phase 0.6 (Environment assessment)   [optional, harness-conditional]
       → Phase 0.9 (Scale-based mode signal)
       → Phase 0.95 (Mode Selection)          ← orchestrator autonomous decision
       → Plan Approval (HUMAN GATE)           ← AskUserQuestion
       → Phase 1 (manager-strategy chain)     ← code-prohibited HARD
       → Phase 1.5/1.6/1.7/1.8 (sub-phases — see run.md)
       → Phase 2.0 (Sprint Contract)          [thorough harness + Tier M/L only]
       → Phase 2 (manager-develop per M-N)
       → Phase 2.5 (GitHub Issue)
       → Phase 2.8 (BODP audit)
       → Phase 3 (Tests + drift guard)
       → Run-PR Merged
       ↓
       Sync Phase 0 (Working tree pre-flight)
       → Phase 0.08/0.1 (Tests)
       → GATE 1 (HUMAN GATE)                  ← AskUserQuestion
       → Phase 0.5/0.5.4 (manager-quality TRUST 5)
       → Phase 0.55 (expert-security)         ← always-runs HARD
       → Phase 0.6/0.7 (coverage)
       → Phase 1-3 (manager-docs sync)
       → GATE 2 (HUMAN GATE)                  ← AskUserQuestion
       → Phase 4 (manager-git PR OR Hybrid Trunk)
       → Sync-PR Merged
       ↓
       Mx Step C SKIP-judge
       → 4-phase close
       → Decision Point 2 (HUMAN GATE optional)
       → Terminal (status: completed)
```

State machine validation: the transition `Phase 1 → Phase 2` MUST NOT be direct — it requires manager-strategy completion + tasks.md artifact emission first. The transition `Phase 0.5 → Phase 1` MUST NOT skip Phase 0.95 Mode Selection. The transition `Run-PR Merged → Sync Phase 0.5.4` MUST NOT skip GATE 1.

### §B.3 Mode Selection Decision Tree (5 modes per REQ-WOF-004)

Phase 0.95 Mode Selection is the central NEW decision logic introduced by this SPEC. The orchestrator autonomously classifies the current task into exactly one of 5 modes based on observable signals.

```
                  ┌────────────────────────────────────────┐
                  │  Inputs:                               │
                  │  - scope_files (from spec.md §C.1)     │
                  │  - domain_count (file path categories) │
                  │  - harness_level (minimal/std/thorough)│
                  │  - team_enabled (workflow.yaml)        │
                  │  - team_env (CLAUDE_CODE_EXPERIMENTAL  │
                  │              _AGENT_TEAMS=1)            │
                  │  - tier (S | M | L from frontmatter)   │
                  │  - prior_audit_score (plan-auditor)    │
                  └────────────────┬───────────────────────┘
                                   │
                                   ▼
                  ┌────────────────────────────────────────┐
                  │  Q1: Is task trivial?                  │
                  │      (typo, single-line, <50 LOC)      │
                  │      → tier=S AND scope_files=1        │
                  └────────────────┬───────────────────────┘
                              YES  │  NO
                       ┌───────────┘   └───────────┐
                       ▼                           ▼
              ┌────────────────┐         ┌──────────────────────┐
              │ Mode: direct   │         │ Q2: Long-running     │
              │  (no Agent)    │         │ task that can run    │
              │  → orchestrator│         │ async non-blocking?  │
              │    inline exec │         │ (e.g., 10+ min       │
              │    only typo/  │         │  research, batch     │
              │    1-line edit │         │  test execution)     │
              └────────────────┘         └──────────┬───────────┘
                                                YES │  NO
                                       ┌────────────┘   └────────────┐
                                       ▼                             ▼
                              ┌──────────────────┐         ┌─────────────────────┐
                              │ Mode: background │         │ Q3: Real-time inter-│
                              │ run_in_background│         │ teammate coordina-  │
                              │ = true           │         │ tion needed AND     │
                              │ (read-only only  │         │ team prerequisites  │
                              │  per CONST-V3R2- │         │ met (team_enabled + │
                              │  020)            │         │ team_env + harness  │
                              └──────────────────┘         │ thorough + tier M/L)│
                                                           └──────────┬──────────┘
                                                                  YES │  NO
                                                       ┌──────────────┘  └──────────────┐
                                                       ▼                                 ▼
                                              ┌──────────────────┐            ┌─────────────────────┐
                                              │ Mode: agent-team │            │ Q4: Multi-domain    │
                                              │ TeamCreate + Send│            │ (≥3 domain catego-  │
                                              │ Message; dynamic │            │ ries) OR ≥10 files  │
                                              │ teammates per    │            │ scope?              │
                                              │ workflow.yaml    │            └──────────┬──────────┘
                                              │ role_profiles    │                   YES │  NO
                                              └──────────────────┘             ┌────────┘   └────────┐
                                                                               ▼                     ▼
                                                                       ┌─────────────────┐  ┌──────────────────┐
                                                                       │ Mode: parallel  │  │ Mode: sub-agent  │
                                                                       │ 3-5 Agent()     │  │ Single Agent()   │
                                                                       │ calls in single │  │ spawn — manager- │
                                                                       │ response message│  │ develop default  │
                                                                       │ ("3-5 parallel  │  │ for typical Tier │
                                                                       │ cut 90%")       │  │ M/L SPECs        │
                                                                       └─────────────────┘  └──────────────────┘
                                                                                                  │
                                                                                                  ▼
                                                                                          (= Mode: sequential
                                                                                           in 5-mode lexicon —
                                                                                           "sequential" is the
                                                                                           degenerate case of
                                                                                           sub-agent with N=1)
```

### §B.3.1 Tie-breaker rules (REQ-WOF-009 boundary)

When scope falls at a threshold ±1:

- 9 files vs 10 files (parallel/sub-agent boundary): default to sub-agent + log `boundary_case: true`
- 2 domains vs 3 domains (sub-agent/parallel boundary): default to sub-agent
- harness `standard` vs `thorough` (sprint-contract gate): default to no Sprint Contract unless explicit `harness: thorough`
- Agent Teams prerequisite check fails: AUTO-DOWNGRADE to autopilot mode + `[mode-auto-downgrade]` info log per REQ-WF003-012

### §B.3.2 Per-iteration stability (edge case 6)

Mode Selection is re-evaluated on each `/moai run` invocation (including `/moai loop` iterations). Per-iteration drift is allowed and recorded in `progress.md § Mode Selection` as an append (not overwrite). This enables retrospective analysis of why mode shifted across iterations (e.g., scope reducing as milestones complete).

### §B.3.3 Mode taxonomy reconciliation

The 5-mode taxonomy described in research.md §C and REQ-WOF-004 maps to the existing `--mode` flag dispatch in spec-workflow.md § Mode Dispatch Cross-Reference:

| 5-mode taxonomy (research) | `--mode` flag (existing) | Mechanism |
|----------------------------|--------------------------|-----------|
| sequential single-spawn | `autopilot` (default) | Single Agent() per phase |
| parallel multi-spawn | `autopilot` w/ multi-Agent() | 3-5 Agent() in single response |
| background non-blocking | (not yet a `--mode` value) | `Agent(run_in_background: true)` |
| sub-agent | `autopilot` (default) | Same as sequential (single Agent()) |
| agent-team | `team` | TeamCreate + dynamic teammates |

This mapping clarifies that "sequential" and "sub-agent" are the same in practice (single Agent() spawn); "parallel" is an enhancement to autopilot mode (multi-Agent() in single message); "background" is a new dimension orthogonal to `--mode`; "agent-team" maps directly to `--mode team`.

---

## §C — Architectural Patterns

Per research.md §B.3 (Research Agent 3 — revfactory/harness), the 6 canonical orchestration patterns map to MoAI workflow phases as follows:

### §C.1 Pattern #1 Pipeline — Sequential A → B → C → D

**MoAI mapping**: the canonical Plan → Run → Sync workflow IS a Pipeline. Each subcommand's output feeds the next.
**Used at**: `/moai plan` → `/moai run` → `/moai sync` overall pipeline.
**Anti-pattern avoided**: serial execution within a single phase that should parallelize.

### §C.2 Pattern #2 Fan-out/Fan-in — Parallel dispatch + synthesize

**MoAI mapping**: plan-phase Phase 0 parallel Explore subagents (3-5 read-only); orchestrator synthesizes research.md.
**Used at**: REQ-WOF-015 compound clause (parallel multi-spawn for multi-domain SPECs).
**Anti-pattern avoided**: monolithic single Explore that covers all domains sequentially.

### §C.3 Pattern #3 Expert Pool — Router → {Expert A | Expert B | Expert C}

**MoAI mapping**: sync-phase Phase 0.5.4/0.55/0.7 routing — manager-quality (TRUST 5) | expert-security (OWASP) | manager-develop (coverage).
**Used at**: REQ-WOF-002 sync-phase 3 specialists restoration.
**Anti-pattern avoided**: orchestrator inlining quality checks without expert delegation.

### §C.4 Pattern #4 Producer-Reviewer — Iterative loop with bounded iteration count

**MoAI mapping**: plan-auditor producer-reviewer cycle (max-3 retry per spec-workflow.md § plan-auditor escalation policy); evaluator-active Sprint Contract for thorough harness.
**Used at**: REQ-WOF-010 evaluator-active invocation (max_iterations: 3); AC-WOF-013 plan-auditor max-3 retry preservation.
**Anti-pattern avoided**: unbounded retry loop.

### §C.5 Pattern #5 Supervisor — Central coordinator → workers

**MoAI mapping**: MoAI orchestrator IS the supervisor. Workers are manager-spec, manager-strategy, manager-develop, manager-quality, manager-docs, manager-git, expert-security, expert-backend, expert-frontend.
**Anti-pattern flagged**: "Supervisor becomes a bottleneck" — addressed by Pattern #2 Fan-out + Pattern #6 Hierarchical Delegation usage.

### §C.6 Pattern #6 Hierarchical Delegation — Multi-level (max 2-3)

**MoAI mapping**: run-phase Phase 1 manager-strategy → Phase 2 manager-develop chain; the chain is 2 levels (strategy + develop), well within the "max 2-3" guidance.
**Used at**: REQ-WOF-003 chain restoration; REQ-WOF-006 manager-strategy spawn before manager-develop.
**Anti-pattern avoided**: ">2-3 levels causes coordination overhead and context loss" — by design we cap at 2 levels (orchestrator → manager-strategy AND orchestrator → manager-develop, not orchestrator → manager-strategy → manager-develop where manager-strategy spawns manager-develop directly — Anthropic constraint "Subagents cannot spawn other subagents" enforces this).

---

## §D — Karpathy 4 Principles + 8 Anti-patterns Application

Per research.md §B.3, the Karpathy 4 Coding Principles + 8 anti-patterns are MoAI-derivative (the upstream README has 4 principles + 3 problems; the 8-pattern catalog is MoAI's expansion). The applicable principles for this SPEC:

### §D.1 Principle 1 — Think Before Coding

**Application**: research.md authored BEFORE any spec.md drafting — surfaces 6 findings + 12 recommendations explicitly. This SPEC is itself an instance of Think-Before-Coding.

### §D.2 Principle 2 — Simplicity First

**Application**: spec.md §F Non-Goals explicitly enumerates 8 out-of-scope items to resist scope expansion. plan.md §C.2 PRESERVE list is the structural enforcement of this principle.

### §D.3 Principle 3 — Surgical Changes

**Application**: §C.1 22-file scope inventory + L46 path-specific staging discipline + manager-develop-prompt-template.md §1.B B10 PRESERVE list invariant.

### §D.4 Principle 4 — Goal-Driven Execution

**Application**: acceptance.md §A 17 AC-WOF-XXX entries each define a testable assertion with explicit Given/When/Then + evidence command + pass criterion. The goal IS the test.

### §D.5 Anti-pattern remediation matrix

| Anti-pattern | Remediation in this SPEC |
|--------------|--------------------------|
| #2 Over-Engineering | spec.md §F Non-Goals + plan.md §C.3 Out of Scope; resist 17-agent catalog expansion |
| #3 Drive-By Refactoring | §C.2 PRESERVE list + L46 path-specific staging; manager-develop prompt enforces scope discipline |
| #5 Silent Assumption | REQ-WOF-011 AskUserQuestion HARD + REQ-WOF-001 HUMAN GATEs surface assumptions |
| #6 Guessing Over Clarifying | §B.1 Clarify phase + Decision Point 1 HUMAN GATE before commit |
| #7 Sycophantic Agreement | manager-strategy code-prohibited HARD assertion forces independent analysis; plan-auditor max-3 contract enforces skeptical evaluation |
| #8 Claiming Without Evidence | AC-WOF-001..017 evidence commands + manager-develop §1.E E1-E7 self-verification + orchestrator 7-item Trust-but-verify batch |

---

## §E — Hook-based Enforcement Strategy (OPTIONAL future enhancement)

Per research.md §B.2 (Research Agent 2 — Context7 verbatim from Anthropic plugin-dev/hook-development) + research.md §D.4 R10:

### §E.1 Vision

Convert orchestrator-discipline into mechanically-unbypassable contracts via PostToolUse / Stop hooks. Example Stop hook from Anthropic verbatim:

> "Review transcript. If code was modified (Write/Edit tools used), verify tests were executed. If no tests were run, block with reason 'Tests must be run after code changes'."

### §E.2 Hook design for this SPEC (future)

The hook implementation is **out of scope** for this SPEC (deferred to follow-up `SPEC-V3R6-WORKFLOW-HOOK-ENFORCEMENT-001`). The vision below documents what future hooks would enforce:

| Hook | Trigger | Enforcement |
|------|---------|-------------|
| Stop hook | end of `/moai sync` | Verify Phase 0.55 expert-security was invoked (`grep .moai/specs/<ID>/progress.md "Phase 0.55"` returns ≥1 match) — block if missing |
| Stop hook | end of `/moai run` | Verify Phase 1 manager-strategy + tasks.md artifact exist — block if missing |
| PostToolUse hook | after Agent() spawn | Verify manager-develop prompt does NOT contain `AskUserQuestion` (B11 subagent boundary discipline) — block if found |
| PostToolUse hook | after manager-docs Write/Edit on `.moai/specs/**/spec.md` body | Block per Status Transition Ownership Matrix (manager-docs forbidden from spec.md body modification) |

### §E.3 Why deferred

This SPEC delivers declarative orchestrator-discipline at the markdown rule layer (declarative HARD assertions in rule files). Hook implementation requires Go code changes (`internal/hook/`) which are explicitly out of scope per spec.md §F.1 and plan.md §C.3.1.

---

## §F — Status Transition Ownership Matrix Reference

Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (canonical SSOT), this SPEC's transitions are owned as follows:

| Transition | Agent | Commit subject pattern |
|------------|-------|------------------------|
| `(none) → draft` | manager-spec (this turn) | `feat(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): plan-phase artifacts (Tier L Section A-G, 4 artifacts)` |
| `draft → in-progress` | manager-develop (M1 commit) | `feat(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): M1 HUMAN GATE 5종 + Mode Selection logic` (or similar M1 subject) |
| `in-progress → implemented` | manager-docs (sync commit) | `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): sync-phase artifacts` |
| `implemented → completed` | orchestrator (Mx + 4-phase close) | `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): Mx-phase audit-ready signal + 4-phase close` |

### §F.1 Forbidden ownership crossings (canonical SSOT)

- manager-docs MUST NOT modify spec.md / plan.md / acceptance.md / design.md body content during sync-phase (frontmatter `status:` + `updated:` + `sync_commit_sha:` are permitted; body is forbidden)
- manager-develop MUST NOT modify spec.md / plan.md / acceptance.md / design.md body content during run-phase (frontmatter `status:` + `updated:` transitions permitted; body is forbidden — except via D-NEW-1 inline-fix pattern with orchestrator re-delegation to manager-spec per REQ-WOF-008)

---

## §G — Risk Mitigation Architecture

Per spec.md §G Risks (6 identified risks R1-R6), the design-level mitigation strategies are:

### §G.1 R1 — HUMAN GATE 5종 user-experience regression

**Design mitigation**: Plan Audit Gate skip-eligibility policy (PASS ≥ 0.90) is preserved — gates SHOULD be present but MAY be auto-confirmed when audit verdict is sufficiently strong. New `human-gates.md` documents this skip-eligibility explicitly + cross-references plan-auditor.md skip-eligibility policy.

### §G.2 R2 — Hierarchical chain wall-time cost

**Design mitigation**: REQ-WOF-015 compound clause + Pattern #2 Fan-out — parallel multi-spawn at Phase 2 manager-develop per-milestone compensates for the chain depth cost. Anthropic verbatim "3-5 parallel cut 90%" benchmark validates the approach.

### §G.3 R3 — Skill router invocation discipline breaking paste-ready resume

**Design mitigation**: session-handoff.md Block 5 already documents `/moai <subcommand>` syntax; the fix is clarification (NOT new format) — clarify orchestrator MUST invoke `Skill()` from the prefix vs Read SKILL.md body. Existing resume messages remain compatible.

### §G.4 R4 — Mode Selection boundary ambiguity

**Design mitigation**: §B.3.1 Tie-breaker rules — default to simpler mode + log `boundary_case: true`. Per-iteration re-classification (§B.3.2) allows retrospective fine-tuning.

### §G.5 R5 — Scope inventory drift during run-phase

**Design mitigation**: L46 path-specific staging discipline + manager-develop §1.B B10 PRESERVE list + plan-auditor MP-3 Traceability check detects divergence. Genuine scope expansion requires blocker report + manager-spec re-delegation (no autonomous expansion).

### §G.6 R6 — New rule file token cost

**Design mitigation**: `paths:` frontmatter conditional loading — `orchestration-mode-selection.md` loads only at `paths: "**/.moai/specs/**"`; `human-gates.md` is intentionally always-loaded (Trigger #3 cross-cutting applicability per session-handoff.md loading scope note) but minimized to ~1200 tokens.

---

## §H — Cross-References

### §H.1 Plan-phase canonical SSOTs

- `.claude/rules/moai/workflow/spec-workflow.md` — Plan/Run/Sync 4-step lifecycle + Tier S/M/L + Phase 0.5 Plan Audit Gate
- `.claude/skills/moai/workflows/plan.md` — current plan router (M1 + M5 update target)

### §H.2 Run-phase canonical SSOTs

- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E 5-section template (Tier L mandatory)
- `.claude/rules/moai/development/orchestrator-templates.md` — 3 templates (Team / Sub / Hybrid)
- `.claude/rules/moai/development/agent-patterns.md` — 6 harness orchestration patterns

### §H.3 Architectural references

- `.claude/skills/moai/references/anti-patterns.md` — Karpathy 8 anti-patterns catalog
- `.claude/rules/moai/development/branch-origin-protocol.md` — BODP audit trail (CONST-V3R5-030..036)
- `.claude/rules/moai/quality/boundary-verification.md` — boundary-verification methodology

### §H.4 External authoritative sources

- claude.com/docs/en/best-practices — 4-phase canonical pattern
- claude.com/docs/en/sub-agents — "Subagents cannot spawn other subagents"
- anthropic.com/engineering/built-multi-agent-research-system — "3-5 parallel cut research time 90%"

### §H.5 Sibling SPEC artifacts

- spec.md — REQ-WOF-001..015 + scope + risks
- plan.md — Lifecycle + milestones M1-M6 + verification strategy
- acceptance.md — 17 AC-WOF-XXX + traceability matrix + edge cases
- research.md — 3-agent parallel research synthesis

---

Version: 0.1.0
Status: plan-phase (M0) — Tier L exclusive design artifact
Scope: architecture rationale + delegation graph + state machine + 5-mode decision tree
