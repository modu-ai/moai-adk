# SPEC-V3R5-HARNESS-AUTONOMY-001 — Plan

> Derived from `harness-autonomy-vision-2026-05-18.md` §3.4 + §6.5 (iteration 3, plan-auditor reviewed). Scope tier: **T2 Standard** (per orchestrator AskUserQuestion). Brownfield strategy: **EXTEND** (B2 resolution in spec.md §1.5).

## 0. Plan-Auditor Iteration Targets

| Iteration | Target verdict | Target score | Must-pass dimensions |
|-----------|---------------|--------------|----------------------|
| iter1 | REVISE expected | ≥0.75 | identify BLOCKING/SHOULD defects |
| iter2 | **PASS** | **≥0.85** | D1 (correctness) ≥0.85, D2 (completeness) ≥0.85, D3 (clarity) ≥0.85, D4 (testability) ≥0.85, D5 (traceability) ≥0.85, D6 (risk) ≥0.80 |

W1 precedent (`SPEC-V3R5-CONSTITUTION-DUAL-001`): iter1 0.71 → iter2 0.96 (recovery delta +0.25). W3 expected delta: **+0.06 (0.841 → 0.902)** per issue #1022 body.

## 1. Architecture Overview

### 1.1 Component diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│  User Workflow (run/sync/fix/loop)                                       │
└─────────────────────────────┬────────────────────────────────────────────┘
                              │ SubagentStop hook
                              ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  internal/harness/capture/  ── heuristic pattern extraction (no LLM)     │
│   └─ writes ──▶ .moai/harness/observations.yaml                          │
└─────────────────────────────┬────────────────────────────────────────────┘
                              │ count++ on match, NEW on miss
                              ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  internal/harness/tier/  ── 1x→3x→5x→10x progression                     │
│   └─ on Tier 4 reach ──▶  generate Proposal (vision §6.4 format)         │
└─────────────────────────────┬────────────────────────────────────────────┘
                              │ Proposal
                              ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  internal/harness/throttle/  ── immediate / batch / quiet / mute         │
│   └─ on dispatch ──▶  internal/harness/safety/pipeline.go                │
└─────────────────────────────┬────────────────────────────────────────────┘
                              │ run5Layer(Proposal)
                              ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  5-LAYER SAFETY PIPELINE (internal/harness/safety/)                      │
│                                                                          │
│   L1 Frozen Guard      sync <10ms p99  ──▶ pass/HARNESS_FROZEN_*         │
│      │   (also enforced at runtime via PreToolUse hook —                 │
│      │    internal/hook/pre_tool.go EXTEND, sees every Write/Edit)       │
│      ▼                                                                   │
│   L3 Contradiction     sync <1s        ──▶ pass/blocker-report           │
│      ▼                                                                   │
│   L4 Rate Limiter      sync <100ms     ──▶ pass/defer                    │
│      ▼                                                                   │
│   L5 Human Oversight   user-paced      ──▶ approve → provisional apply   │
│      │                                                                   │
│      │  ┌─────────────────────────────────────────────────┐              │
│      └─▶│ L2 Canary  async ~30s  shadow eval last 3 SPECs │              │
│         │   PASS → status: provisional → applied          │              │
│         │   FAIL → automatic rollback + AskUserQuestion   │              │
│         │          notification (Canary VETO POWER)       │              │
│         └─────────────────────────────────────────────────┘              │
└──────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                .moai/harness/evolution-log.md (audit trail)
```

### 1.2 Why 5-Layer ordering is INVARIANT (REQ-HRA-036)

- **L1 first**: cheapest + hardest fail (frozen path matchers are deterministic). Filter out 100% noise.
- **L3 second** (not L2): contradiction detection is sync + fast (<1s). If a new rule conflicts with existing, we want to know before spending 30s on Canary.
- **L4 third**: rate-limit is sync gate; cheap to evaluate.
- **L5 fourth**: user is the most expensive resource — only ask after L1/L3/L4 pass.
- **L2 fifth (async, VETO power)**: Canary is slow but authoritative on regression. We don't block L5 on L2 (would feel sluggish), but L2 retains the right to undo L5.

Reordering requires plan-auditor approval per REQ-HRA-036.

## 2. Milestones (M1–M6)

| # | Milestone | Files (new/EXTEND) | Est LOC | Acceptance |
|---|-----------|----------------------|---------|-----------|
| M1 | Lesson Auto-Capture | `internal/harness/capture/` (NEW), `internal/harness/observer.go` EXTEND | ~450 | AC-HRA-001, AC-HRA-002 |
| M2 | 4-Tier Pipeline + Anti-Pattern | `internal/harness/tier/` (NEW), `internal/harness/types.go` EXTEND | ~350 | AC-HRA-003, AC-HRA-004, AC-HRA-005 |
| M3 | 5-Layer Safety (EXTEND brownfield) | `internal/harness/safety/{pipeline,frozen_guard,canary,contradiction,rate_limit,oversight}.go` EXTEND + `internal/hook/pre_tool.go` EXTEND | ~900 | AC-HRA-006..010 + R-HRA-S1 |
| M4 | Proposal Throttling | `internal/harness/throttle/` (NEW), `.moai/config/sections/workflow.yaml` schema extension | ~250 | AC-HRA-011, AC-HRA-012 |
| M5 | Cold-Start Seeds (stub only) | `internal/harness/seeds/` (NEW), 1 dummy seed fixture | ~150 | AC-HRA-013 (contract test only; AC-HRA-013-full = W4) |
| M6 | CLI 6 verbs (+2 helpers) | `internal/cli/harness.go` (NEW or EXTEND existing) | ~400 | AC-HRA-014 |
| **Total** | | **~2500 + ~350 tests = ~2850 LOC** | | **14 ACs, 6 EC, 5 Risk Mitigations** |

### 2.1 M1 — Lesson Auto-Capture

**Files**:
- NEW `internal/harness/capture/extractor.go` — heuristic pattern extractor (no LLM call)
- NEW `internal/harness/capture/extractor_test.go`
- EXTEND `internal/harness/observer.go` — wire SubagentStop hook → `capture.Extract(diff)`

**Pattern categories** (vision §6.2 enum):
`[error-handling, naming, testing, architecture, security, performance, hardcoding, workflow]`

**Heuristic examples**:
- error-handling: regex `fmt\.Errorf\([^)]*%w` → "uses %w wrapping"
- naming: regex `func [A-Z][a-zA-Z0-9]+` → exported function naming
- testing: regex `t\.TempDir\(\)` → uses test-isolation pattern

**Output**: append to `.moai/harness/observations.yaml` with canonical `created:` / `updated:` field names.

### 2.2 M2 — 4-Tier Pipeline

**Files**:
- NEW `internal/harness/tier/progression.go` — count threshold matcher
- NEW `internal/harness/tier/anti_pattern.go` — critical failure detector
- NEW `internal/harness/tier/proposal_generator.go` — emits vision §6.4 YAML
- EXTEND `internal/harness/types.go` — add `Observation`, `Proposal`, `TierStatus`, `EvolutionRecord` types

**Tier transitions** (REQ-HRA-008):
- 1x → `observation` (logged, no action)
- 3x → `heuristic` (suggestion only, no AskUserQuestion)
- 5x → `rule` (graduation candidate, queued)
- 10x → `high-confidence` (Tier 4: generate Proposal, route via throttle/)

**Anti-pattern (REQ-HRA-014)**: single observation with `severity: critical` (score drop >0.20 OR must-pass fail) → status `anti-pattern`, FROZEN immediately.

### 2.3 M3 — 5-Layer Safety (CORE OF SPEC — brownfield EXTEND)

This is the largest milestone. Per §1.5.2, the `internal/harness/safety/` package already exists with stub-level pipeline / layers; W3 makes them production.

#### M3.1 — L1 Frozen Guard (sync <10ms p99)

**Files** (EXTEND):
- `internal/harness/safety/frozen_guard.go` — algorithm + W1 zone-registry consumer
- `internal/harness/safety/frozen_guard_test.go` — 8 sentinel coverage
- `internal/hook/pre_tool.go` — wire as PreToolUse runtime hook (path of enforcement)

**Algorithm** (R-HRA-S1):
1. Hook fires on every Write/Edit/MultiEdit tool call (already existing routing in `pre_tool.go`)
2. Extract `agent_name` from hook payload; fallback to `MOAI_INVOKING_AGENT` env (SubagentStart-injected)
3. If `agent_name` does NOT match `harness-learner|my-harness-*` → ALLOW (early exit, ~50ns)
4. Compiled glob matcher (init-time, once) checks `file_path` against 8 deny patterns
5. On match: return JSON `{status: denied, sentinel, agent, path, reason}` + exit non-zero
6. On no match: ALLOW

**Performance budget**:
- p50 < 1ms (early-exit path for non-harness agents)
- p99 < 10ms (harness agent + full glob match)
- Bench: `BenchmarkFrozenGuardL1` with mixed-agent simulated payload, 100-sample window

**8-sentinel matcher catalog** (vision §3.4):

| Sentinel | Glob pattern (deny) |
|----------|---------------------|
| `HARNESS_FROZEN_AGENT_VIOLATION` | `.claude/agents/moai/**` |
| `HARNESS_FROZEN_SKILL_VIOLATION` | `.claude/skills/moai-*/**` (EXCEPT `.claude/skills/my-harness-*/**`) |
| `HARNESS_FROZEN_RULE_VIOLATION` | `.claude/rules/moai/**` |
| `HARNESS_FROZEN_COMMAND_VIOLATION` | `.claude/commands/moai/**` |
| `HARNESS_FROZEN_HOOK_VIOLATION` | `.claude/hooks/moai/**` |
| `HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION` | `.claude/output-styles/moai/**` |
| `HARNESS_FROZEN_INSTRUCTION_VIOLATION` | `CLAUDE.md` (project root) |
| `HARNESS_FROZEN_CONFIG_VIOLATION` | `.moai/config/sections/*.yaml` |

**Bypass** (REQ-HRA-025): `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env, set only by `moai update` CLI internal flow. Tested in `cli/update_test.go`.

#### M3.2 — L2 Canary (async ~30s shadow eval, VETO power)

**Files** (EXTEND):
- `internal/harness/safety/canary.go` — shadow-eval driver, async-veto orchestration
- `internal/harness/safety/canary_test.go` — fixture: 3-project rolling baseline

**Algorithm** (Canary Veto Policy E5):
1. On L5 user-approve, immediately do `provisional apply` (write my-harness-* file with status=`provisional`)
2. Spawn goroutine: re-evaluate last 3 projects in `.moai/harness/canary-baseline/` against the new rule
3. Compare score: if ANY project drops >0.10 → VETO
4. On VETO:
   - Revert the provisional file (atomic rename of `.bak` copy)
   - Append `HARNESS_LEARNING_CANARY_VETO` to evolution-log.md
   - Enqueue AskUserQuestion notification (orchestrator-side; subagent just emits the blocker report)
   - Set 48h cooldown on this proposal (rate_limit state)
5. On PASS:
   - Atomic transition `provisional → applied`
   - Append `apply` to evolution-log.md

**Why VETO post-L5**: 30s blocking on L5 user wait feels sluggish; provisional apply gives instant feedback, and Canary acts as backstop for the rare false-positive user approval.

#### M3.3 — L3 Contradiction Detector (sync <1s)

**Files** (EXTEND):
- `internal/harness/safety/contradiction.go`
- `internal/harness/safety/contradiction_test.go`

**Algorithm**:
1. Load existing harness rules from `.moai/harness/rules/*.yaml`
2. Compute conflict matrix: proposed rule vs each existing (text-similarity + semantic anti-pattern check)
3. On contradiction (>0.85 similarity but opposite directive): emit blocker report `{ contradiction: { existing_rule_id, proposed_change, recommendation }}`
4. Per REQ-HRA-013: subagent does NOT invoke AskUserQuestion; returns blocker → orchestrator runs AskUserQuestion

#### M3.4 — L4 Rate Limiter (sync <100ms)

**Files** (EXTEND):
- `internal/harness/safety/rate_limit.go`
- `internal/harness/safety/rate_limit_test.go`

**State**: `.moai/harness/rate-limit-state.json`
```json
{ "evolutions": [{"id": "EVO-2026-001", "applied_at": "..."}, ...],
  "active_learnings": 42 }
```

**Limits** (vision §6.5 + design.yaml):
- ≤3 evolutions per 7-day window
- ≥24h cooldown between consecutive evolutions
- ≤50 active learnings total

**On limit hit**: emit `defer` verdict; proposal re-queued for next eligible window.

#### M3.5 — L5 Human Oversight (blocker report pattern)

**Files** (EXTEND):
- `internal/harness/safety/oversight.go`
- `internal/harness/safety/oversight_test.go`

**Per REQ-HRA-029/030**: subagent NEVER calls AskUserQuestion. Layer emits:

```yaml
blocker_report:
  type: human-oversight-required
  proposal_id: PROP-2026-042
  question: "다음 패턴을 my-harness-go-specialist에 자동 적용할까요?"
  options:
    - { label: "Apply (권장)", description: "..." }
    - { label: "Apply with modification", description: "..." }
    - { label: "Defer", description: "..." }
    - { label: "Reject permanently", description: "..." }
```

Orchestrator (main session) receives blocker, runs `ToolSearch(query: "select:AskUserQuestion")` + `AskUserQuestion`, injects user's selection into a fresh subagent prompt to continue the pipeline.

**Constraint AC C-HRA-008** verifies no AskUserQuestion in subagent definition.

### 2.4 M4 — Proposal Throttling

**Files**:
- NEW `internal/harness/throttle/throttle.go` — 4 modes (immediate / batch / quiet / mute)
- NEW `internal/harness/throttle/throttle_test.go`
- Schema extension in `.moai/config/sections/workflow.yaml` (per vision §6.6)

**4 modes**:
| Mode | Behavior |
|------|----------|
| `immediate` (default) | Dispatch instantly on Tier 4 reach + 24h per-proposal cooldown |
| `batch` | Queue; dispatch ≤`max_per_window` at `window` boundary (weekly / sprint_end / manual) |
| `quiet` | Suppress during `quiet.hours` range; dispatch at next non-quiet boundary |
| `mute` | Per-category silence; observation count still increments but no dispatch |

R11 timeout (REQ-HRA-028): if user doesn't respond to AskUserQuestion within 7 days, auto-defer to `quiet` mode + emit `HARNESS_LEARNING_PROPOSAL_TIMEOUT`.

### 2.5 M5 — Cold-Start Seeds (stub only)

**Files**:
- NEW `internal/harness/seeds/loader.go` — schema + load hook
- NEW `internal/harness/seeds/loader_test.go`
- NEW `.claude/skills/moai-meta-harness/seeds/_dummy.yaml` — 1 dummy seed for contract test

**Seed schema** (REQ-HRA-020 / REQ-HRA-033):
```yaml
seed:
  id: SEED-GO-ERROR-WRAP
  category: error-handling
  tier: 3                           # starting point
  pattern: "fmt.Errorf with %w"
  description: "Standard Go error wrapping"
  applies_to:
    project_markers: [go.mod]
```

**M5 scope limit** (D11): no seed library content; W4 will populate 8 baseline seeds (Go/Node/Python/Rust/React/Vue/Flutter/iOS).

### 2.6 M6 — CLI

**Files**:
- NEW or EXTEND `internal/cli/harness.go` — 6 verbs + 2 helpers

| Verb | Purpose | Exit codes |
|------|---------|-----------|
| `moai harness status` | Show observations, tiers, active learnings, queue depth, mode | 0 ok, 1 disabled |
| `moai harness apply <proposal-id>` | Manual apply (bypass throttle wait, still goes through 5-Layer) | 0, 1 not-found, 2 safety-rejected |
| `moai harness rollback <evolution-id>` | Revert an applied evolution (idempotent per R-HRA-I1) | 0, 2 unknown |
| `moai harness disable` | Toggle disabled sentinel file | 0 |
| `moai harness mute <category>` | Add category to mute list | 0 |
| `moai harness verify --determinism` | W3 stub returning "not yet implemented" (full impl = W4) | 0 with stub message |
| `moai harness mute-list` (helper) | Print muted categories | 0 |
| `moai harness unmute <category>` (helper) | Remove category from mute list | 0 |

All verbs support `--json` flag for structured output (REQ-HRA-005).

## 3. Sentinel Catalog (canonical, 10 total)

8 `HARNESS_FROZEN_*_VIOLATION` from M3.1 +

| Sentinel | Source | Severity |
|----------|--------|----------|
| `HARNESS_LEARNING_LATENCY_BUDGET_BREACH` | M3.1 R-HRA-S1 monitor | warn |
| `HARNESS_LEARNING_CANARY_VETO` | M3.2 L2 VETO | warn |
| `HARNESS_LEARNING_CONTRADICTION_DETECTED` | M3.3 L3 blocker | info |
| `HARNESS_LEARNING_SCHEMA_DRIFT` | M1 loader | warn |
| `HARNESS_LEARNING_UNKNOWN_EVOLUTION` | M6 rollback CLI | error |
| `HARNESS_LEARNING_PROPOSAL_TIMEOUT` | M4 R11 R-HRA-T1 | info |

Total **8 + 6 = 14** sentinels — initial spec.md §7 stated 10 (8 FROZEN + 2 LEARNING); the **expanded catalog adopted here is 14** as M3/M4 review surfaced 6 distinct LEARNING-* failure modes. Each has ≥1 unit test (R-HRA-Q1, expanded).

## 4. File-Touch Manifest

| Layer | NEW | EXTEND | PRESERVE-only |
|-------|-----|--------|---------------|
| `internal/harness/capture/` | ✅ 2 files | — | — |
| `internal/harness/tier/` | ✅ 3 files | — | — |
| `internal/harness/throttle/` | ✅ 2 files | — | — |
| `internal/harness/seeds/` | ✅ 2 files | — | — |
| `internal/harness/safety/` | — | ✅ 6 files | — |
| `internal/harness/{observer,learner,types,safety_preservation_test,testdata}` | — | ✅ 5 files | — |
| `internal/harness/{applier,chaining_rules,cleanup,frozen_guard,layer{1,2,3,5},interview,...}` | — | — | ✅ 20+ files |
| `internal/hook/pre_tool.go` | — | ✅ 1 file | — |
| `internal/cli/harness.go` | ✅ (or EXTEND if exists) | — | — |
| `.claude/agents/moai/harness-learner.md` | ✅ 1 file | — | — |
| `.claude/skills/moai-meta-harness/seeds/_dummy.yaml` | ✅ 1 file | — | — |
| `.moai/config/sections/workflow.yaml` | — | ✅ 1 file (harness.proposal section) | — |
| **Total estimated** | **~12 NEW** | **~13 EXTEND** | **~20 PRESERVE** |

≈25 touched files, ~2850 LOC (code + tests) — matches issue #1022 estimate.

## 5. Out of Scope (explicit cross-ref to spec §5)

- Deterministic harness generation (EXCL-HRA-001, W4)
- 8 baseline seed content (EXCL-HRA-002, W4)
- `/moai project --refresh` (EXCL-HRA-003, W4)
- meta-harness 7-Phase workflow (EXCL-HRA-004, W4)
- Project-specific my-harness-* generation (EXCL-HRA-005, W4)
- Migration tooling (EXCL-HRA-006)
- LLM-based capture (EXCL-HRA-007)
- Cross-repo sync (EXCL-HRA-008)
- Recursive learner self-introspection (EXCL-HRA-009)
- Past-30-day rollback extension (EXCL-HRA-010)

## 6. Risks (5 mitigations — cross-ref to acceptance.md R-HRA-*)

| # | Risk | Probability | Impact | Mitigation |
|---|------|-------------|--------|------------|
| R-HRA-01 | L1 hook latency creeps past 10ms p99 as deny-pattern list grows | Medium | High (every Write/Edit slows) | Pre-compiled glob matcher at init; benchmark gate in CI (R-HRA-S1); pattern count cap = 16 |
| R-HRA-02 | Canary VETO + provisional apply creates inconsistent state if process crashes mid-revert | Low | High (Core path corruption) | Atomic file rename via `.bak`; recovery routine in `cleanup.go` on startup |
| R-HRA-03 | Subagent boundary leakage (someone adds AskUserQuestion to harness-learner.md) | Medium | Critical (HARD rule violation) | C-HRA-008 grep gate in CI; R-HRA-Q2 enforcement |
| R-HRA-04 | Brownfield `internal/harness/safety/*.go` existing tests fail after EXTEND | Medium | Medium (CI red, rework) | Pre-flight: read each file + run `go test ./internal/harness/safety/...` before EXTEND; preserve all existing test names |
| R-HRA-05 | W4 inheriting W3 substrate finds API mismatch (seed loader doesn't fit meta-harness 7-Phase consumer) | Low | Medium (W4 rework) | M5 stub keeps signature minimal (`Load() []Seed`); document API contract in `internal/harness/seeds/CONTRACT.md` |

## 7. Open Questions (for plan-auditor)

1. **OQ1 — Sentinel count**: spec §7 says 10, plan §3 says 14 after expansion. Authoritative is plan §3 (acceptance.md C-HRA-008 must align).
2. **OQ2 — Canary baseline source**: vision §6.5 "last 3 projects" — for a single-project repo, what fills the baseline? Proposal: synthetic snapshots from `.moai/specs/SPEC-*/` last 3 SPECs' acceptance.md PASS state.
3. **OQ3 — Disabled file persistence**: `.moai/harness/disabled` sentinel — survive `git clean`? Proposal: yes (commit if user wants), but `.gitignore` it by default per per-repo isolation (EXCL-HRA-008).

---

End of plan.md (draft v0.1.0, awaiting plan-auditor iter1).
