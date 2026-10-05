---
id: SPEC-V3R5-HARNESS-AUTONOMY-001
title: "Harness Autonomy — Implementation Plan"
version: "0.1.0"
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.0.0 — Round 5"
module: "internal/harness + .moai/harness + .claude/skills/moai-harness-learner + internal/hook"
lifecycle: spec-anchored
tags: "harness, autonomy, self-evolution, 4-tier, 5-layer-safety, plan, mega-sprint, w3"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial implementation plan — 6 milestones (M1→M6) sequential per W1 single-run-phase pattern. Scope T2 Standard per user AskUserQuestion. |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision per plan-auditor iter 1 BLOCKING + SHOULD defects. B1: §3.2/§11.5/§12 W1 deliverable corrected (W1 ships zone-registry data SSOT only; W3 IMPLEMENTS PreToolUse hook code). B2: §12 brownfield reconciliation — `internal/harness/safety/` extended; `internal/harness/layer*.go` PRESERVED (different concern, package `harness` trigger verification per spec.md §1.5). Consolidation strategy (b) — extend safety/ subdirectory, do NOT consolidate layer*.go. B3: §4.2 seed location dual-path SSOT/cache resolved per spec.md §1.6. B4: §3.3 Canary Veto cooldown rejection clarified. B5: §3.3 L3+L5 unified blocker-report pattern documented. S4: §9.3 R11 mitigation rewritten — timeout=FAIL (auto-rollback) preserves Canary final-gate. S10: §4.2 M5 stub-only project-type detection contract documented. Milestone M5 estimate updated (interface-only). |

---

## 1. Architecture Overview

W3 implements 5 cooperating subsystems. W1 ships the **zone-registry data SSOT only** (per W1 EXCL-001); W3 implements the PreToolUse Frozen Guard hook code that consumes it:

```
┌──────────────────────────────────────────────────────────────────────┐
│  Workflow Event (any /moai * subcommand completes)                    │
└────────────────────┬─────────────────────────────────────────────────┘
                     │
                     ▼  SubagentStop hook trigger (REQ-HRA-001)
┌──────────────────────────────────────────────────────────────────────┐
│  M1 — Lesson Capture Pipeline                                         │
│  - harness-learner subagent (background, read-only initial scan)     │
│  - heuristic match (no LLM call, <500ms p95) — REQ-HRA-002           │
│  - emit observation candidate                                         │
└────────────────────┬─────────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────────┐
│  M2 — Tier Engine                                                     │
│  - write to .moai/harness/observations.yaml                          │
│  - count increment + status transition (1x/3x/5x/10x)                │
│  - anti-pattern auto-flag on critical failure (REQ-HRA-006)          │
│  - tier-progression event → evolution-log.md                         │
└────────────────────┬─────────────────────────────────────────────────┘
                     │ Tier 4 reached (count ≥ 10)
                     ▼
┌──────────────────────────────────────────────────────────────────────┐
│  M3 — 5-Layer Safety (sequential, with Canary veto)                  │
│  L1 Frozen Guard  (sync, < 10ms p99, W3 hook + Vision §3.4 sentinels)│
│                   (reads W1 zone-registry data; W3 implements hook)  │
│         ↓                                                              │
│  L2 Canary        (async, ~30s, shadow eval last 3 SPECs)            │
│         ↓                                                              │
│  L3 Contradiction (sync, < 1s, semantic conflict scan)               │
│         ↓                                                              │
│  L4 Rate Limiter  (sync, < 100ms, 3/week + 24h cd + 50 active max)   │
│         ↓                                                              │
│  L5 Human Oversight (orchestrator AskUserQuestion via blocker report)│
│         │                                                              │
│         ├─ approve → provisional apply (if L2 pending) or apply       │
│         ├─ reject  → status=anti-pattern (permanent block)            │
│         └─ defer   → cooldown (24h) → re-queue                        │
└────────────────────┬─────────────────────────────────────────────────┘
                     │
                     ▼  M4 — Proposal Throttling intercepts before L5
┌──────────────────────────────────────────────────────────────────────┐
│  M4 — Proposal Throttling (4 modes)                                  │
│  immediate (default): direct AskUserQuestion at Tier 4               │
│  batch:    queue + emit at window boundary (weekly/sprint_end)       │
│  quiet:    defer during quiet.hours window                            │
│  mute:     per-category silence (log to evolution-log status=muted)  │
└──────────────────────────────────────────────────────────────────────┘

                     ▲ Cold-start (observations.yaml empty)
                     │
┌──────────────────────────────────────────────────────────────────────┐
│  M5 — Cold-Start Seed Load Hook                                       │
│  - schema defined (REQ-HRA-022)                                       │
│  - load from .claude/skills/moai-meta-harness/seeds/* (REQ-HRA-024)  │
│  - inject as Tier 3 (status=rule) into observations.yaml (REQ-HRA-023)│
│  - NOTE: seed library CONTENT is W4 scope; W3 ships empty seed dir   │
└──────────────────────────────────────────────────────────────────────┘

                     ┌──────────────────────────────────────────────┐
                     │  M6 — End-to-end Test                         │
                     │  - Tier 1→4 progression                       │
                     │  - 5-Layer simulation per-layer + integration │
                     │  - Frozen Guard violation block               │
                     │  - Cold-start regression                       │
                     │  - Canary Veto provisional + rollback         │
                     │  - Throttling 4-mode comparison               │
                     └──────────────────────────────────────────────┘
```

Data layout (W3 creates `.moai/harness/` directory):

```
.moai/harness/
├── observations.yaml      # Tier 1-4 entries (lesson capture log)
├── evolution-log.md       # append-only evolution history
├── anti-patterns.yaml     # FROZEN learned anti-patterns
└── proposal-queue.yaml    # batch mode pending proposals (M4)
```

---

## 2. Tier Engine (§6.3 + §6.1)

### 2.1 Count threshold table (preserved from harness.yaml)

| Observations | Classification | Status field value | Action |
|---|---|---|---|
| 1x | Observation | `observation` | logged only |
| 3x | Heuristic | `heuristic` | suggestion (manager-develop hint) |
| 5x | Rule | `rule` | Sprint Contract auto-add candidate |
| 10x | High-confidence | `high-confidence` | AskUserQuestion auto-propose (subject to M4 throttling) |
| 1x critical | Anti-Pattern | `anti-pattern` | FROZEN immediate flag |
| Seed start | Pre-loaded | `rule` (status=rule from §4.4) | Tier 3 starting point |

### 2.2 Tier transition rules

State machine (REQ-HRA-004):

```
[seed inject]    →  rule         (count=5 synthetic)
observation      →  heuristic    (count crosses 3)
heuristic        →  rule         (count crosses 5)
rule             →  high-confidence  (count crosses 10)
high-confidence  →  graduated    (after L5 approve + Canary PASS)
high-confidence  →  anti-pattern (after L5 reject_permanent)
*                →  archived     (when active count > 50, oldest observation status)
```

Transitions are atomic write operations on `.moai/harness/observations.yaml` (file lock via `flock(2)` to avoid concurrent SubagentStop race per EC-HRA-002).

### 2.3 Anti-pattern auto-flag mechanism (§6.3 — REQ-HRA-006)

Trigger conditions (any one activates):

1. SPEC quality score drops > 0.20 between consecutive iterations of the same SPEC
2. evaluator-active reports must-pass criterion FAIL on a previously-passing dimension
3. characterization test regression on previously-green test

Flag operation:

- Write entry to `.moai/harness/anti-patterns.yaml` with full evidence (commit, before/after, context)
- FROZEN status (no further evolution allowed; only human can reclassify)
- evaluator-active future invocations consult anti-patterns.yaml as score cap (per Vision §12 Mechanism 5)

---

## 3. 5-Layer Safety Sequencing (§6.5)

### 3.1 Sequential execution

Each layer must pass before the next is invoked. Synchronous user-blocking: L1 + L3 + L4 + L5. L2 is asynchronous (background) but holds **veto power** per §3.3.

| Layer | Sync/Async | Budget | Failure action |
|---|---|---|---|
| L1 Frozen Guard | sync | < 10ms p99 | reject with HARNESS_LEARNING_FROZEN_BLOCKED |
| L2 Canary | async (~30s) | 30s soft / 60s hard | veto + rollback (§3.3) |
| L3 Contradiction | sync | < 1s | blocker report → orchestrator AskUserQuestion (resolve-by-replacing / amending / reject) per §3.3a |
| L4 Rate Limiter | sync | < 100ms | defer to next eligible window (HARNESS_LEARNING_RATELIMIT_EXCEEDED) |
| L5 Human Oversight | sync (user-paced) | unbounded | per user choice (Apply / Apply-modify / Defer / Reject) |

### 3.2 Layer interaction with W1 Frozen Guard (B1 corrected)

W1 explicitly disclaimed PreToolUse hook implementation per W1 spec.md §5.2 EXCL-001. **W3 is the first runtime implementer of the L1 Frozen Guard hook.** W3 extends `internal/hook/pre_tool.go` (existing 20,548 bytes since 2026-05-18, see spec.md §1.5 brownfield inventory) with:

1. **zone-registry consumer**: reads `.claude/rules/moai/core/zone-registry.md` (W1 deliverable, 111 entries) at hook init time
2. **8 HARNESS_FROZEN_* sentinel emission**: per Vision §3.4 catalog (catalog itself defined in Vision, NOT W1)
3. **harness-learner agent identity gate**: only blocks writes when the invoking subagent matches the harness-learner pattern
4. **W3 wrapper sentinel** `HARNESS_LEARNING_FROZEN_BLOCKED` carrying the matched W1 catalog sentinel as `cause` field

Reference path: zone-registry data → W3 L1 hook code (NEW) → 8 sentinel catalog (Vision-defined) → wrapper sentinel.

The data flow is **uni-directional**: W3 reads W1 data; W3 does NOT call any W1 function or hook (none exists).

### 3.3 Canary Veto Policy (E5 verbatim from Vision §6.5)

> Layer 2 Canary는 asynchronous로 ~30s 소요되어 Layer 5 user approval보다 늦게 완료될 수 있다. 다음 정책으로 race 해소:
>
> 1. Layer 5 (user approval)이 Canary 완료 전 도착 → **provisional apply** (my-harness-* file write 수행, 단 evolution status = `provisional`)
> 2. Canary 완료:
>    - PASS → evolution status `provisional → applied`, evolution-log.md에 기록
>    - FAIL → **automatic rollback** (provisional file revert) + AskUserQuestion notification ("Canary가 regression 감지하여 자동 롤백됨. Override 또는 deeper review?")
> 3. **Canary는 Layer 5 approval에 대한 veto power 보유** — user approval이 final이 아님, Canary가 final gate
> 4. Veto 발생 시 해당 proposal은 48h cooldown 후 재제안 가능 (rate limiter에 별도 entry)

Implementation in W3 (`internal/harness/safety/canary.go` EXTEND per §1.5 brownfield inventory):

- Async eval dispatch via goroutine; result delivered via channel
- `evolution_status` field on every applied evolution: `provisional` | `applied` | `vetoed_by_canary`
- File revert preserves original content via copy-before-write pattern (snapshot to `.moai/harness/revert/<evolution-id>/` before applying)
- 48h cooldown tracked in rate-limiter dedicated entry

**B4 cooldown finality (per AC-HRA-008b)**: After auto-rollback, the proposal enters a 48h cooldown tracked in L4 rate-limiter as a dedicated entry. Re-application via `moai harness apply <evolution-id>` within cooldown MUST be REJECTED with sentinel `HARNESS_LEARNING_RATELIMIT_EXCEEDED` and exit code 1. The post-rollback orchestrator notification text is rephrased per B4 — see Output Surface in §3.3b below.

**B4 notification text (§3.3b)**: Replace the Vision §6.5 step 2(b) wording "Override 또는 deeper review?" with cooldown-preserving phrasing: "Canary가 regression을 감지하여 provisional 변경이 자동 롤백되었습니다. 다음 옵션: (a) deeper review (새 proposal 생성 — fresh tier 요구사항 적용, 48h cooldown 후), (b) 거부 (영구 anti-pattern으로 분류)". The "Override" path (즉시 적용 강제) is **removed** to preserve Canary final-gate semantic — there is no user-side override of a Canary veto within the cooldown window.

### 3.3a L3+L5 Unified Blocker-Report Pattern (B5 + S5 resolution)

REQ-HRA-014 (L3) and REQ-HRA-018 (L5) **both** use the blocker-report pattern — harness-learner subagent NEVER invokes AskUserQuestion directly. The orchestrator owns ALL AskUserQuestion invocations.

Pattern (applies to L3 + L5):

```
harness-learner (subagent)                    Orchestrator
       │                                          │
       │ emit blocker report (markdown)           │
       │─────────────────────────────────────────>│
       │                                          │ parse blocker report
       │                                          │ ToolSearch(select:AskUserQuestion)
       │                                          │ AskUserQuestion(...)
       │                                          │ collect user response
       │ orchestrator re-delegates with response  │
       │<─────────────────────────────────────────│
       │ continue pipeline                        │
       ▼                                          ▼
```

Layer-by-layer interaction model:

| Layer | User interaction? | Mechanism |
|-------|-------------------|-----------|
| L1 Frozen Guard | No | Pure data-driven block; no user channel |
| L2 Canary | No | Async background; result published to evolution-log.md + (on veto) triggers notification report |
| L3 Contradiction | **Yes via orchestrator** | Blocker report → orchestrator AskUserQuestion (resolve-by-replacing/amending/reject) |
| L4 Rate Limiter | No | Purely-internal; the only layer with NO user interaction path |
| L5 Human Oversight | **Yes via orchestrator** | Blocker report → orchestrator AskUserQuestion (Apply / Apply-with-modification / Defer / Reject permanently) |

Note: L4 is "the only purely-internal layer" per spec.md §C-HRA-008 binary verification. Static grep MUST find zero `AskUserQuestion` references in `internal/harness/` and `internal/hook/` (C-HRA-008 + S5).

### 3.4 Sequencing edge case (L2 returns before L5)

When Canary completes before L5 (fast canary, slow user):

- `canary_status: PASS` → proceed to L5 normally (no provisional path)
- `canary_status: FAIL` → reject at L2 (L5 never invoked), emit `HARNESS_LEARNING_CANARY_REJECTED` blocker report

---

## 4. Cold-Start Seed Library (§4.4)

### 4.1 Seed file format (REQ-HRA-022)

YAML schema at `.claude/skills/moai-meta-harness/seeds/<framework>/<category>.yaml`:

```yaml
seeds:
  - id: SEED-{LANG}-{NNN}      # e.g., SEED-GO-001
    pattern: <short description>
    tier: 3                    # always 3 (Tier 3 starting point)
    confidence: 0.85           # initial confidence
    category: <enum>           # error-handling | naming | testing | architecture |
                               # security | performance | hardcoding | workflow
    body: |
      <multi-line lesson body>
    references:
      - <URL>
```

### 4.2 Seed lifecycle (REQ-HRA-023) — B3 + S10 resolution

**Dual-path SSOT/cache model (per spec.md §1.6 D11 resolution)**:

| Layer | Path | Lifecycle | Ownership |
|-------|------|-----------|-----------|
| Canonical SSOT (shipped) | `.claude/skills/moai-meta-harness/seeds/` | Core repo 일부, `moai update`로 갱신 | Core MoAI maintainer |
| Project-local cache (runtime) | `.moai/harness/seeds/` | `/moai project` (W4)이 populate, W3 ship 시 empty | User project (per-project) |

**Precedence rule (B3)**: 같은 seed ID에 대해 project-local (`.moai/harness/seeds/`) 우선, 그 다음 SSOT (`.claude/skills/moai-meta-harness/seeds/`).

**Seed loader contract (S10 stub-only)**:

W3 ships interface only:

```go
// seeds/loader.go (W3 NEW)
package seeds

type Loader interface {
    LoadForProject(projectType string) ([]Seed, error)
}

// DetectProjectType is a W3 STUB returning literal "unknown".
// Full marker-based detection (go.mod / package.json / Cargo.toml / etc.)
// is W4 PROJECT-MEGA-001 scope. The stub is unit-tested via
// TestLoadForProject_UnknownProject which verifies empty seed list returned
// with no error when projectType="unknown".
func DetectProjectType() string {
    return "unknown"  // W4 will replace with marker-based detection
}
```

**Lifecycle steps**:

1. (W3) `harness-learner` first invocation in a project checks `.moai/harness/observations.yaml` — if absent or empty, attempts seed load.
2. (W3) Calls `seeds.LoadForProject(seeds.DetectProjectType())` — STUB returns `"unknown"` → empty seed list → no seed inject (graceful no-op).
3. (W4) `/moai project --refresh` will replace DetectProjectType stub with marker-based detection (go.mod / package.json / etc.) and populate `.moai/harness/seeds/` from SSOT.
4. (W4 future) Matching seed file(s) loaded by harness-learner; seeds injected into `observations.yaml` with `status: rule` (Tier 3 starting point), `count: 5` synthetic, `confidence` from seed file.
5. (Post-W4) Workflow execution accumulates observations; seed may graduate to Tier 4 (`high-confidence`) via the standard tier engine path.

### 4.3 Meta-harness integration point

W3 invokes the existing `moai-meta-harness` skill body for seed loading **only at the lookup level** (consulting the SSOT directory). The full meta-harness 7-Phase workflow is W4 scope (EXCL-HRA-004). W3 uses only the `seeds.LoadForProject(projectType string) ([]Seed, error)` interface, defined in W3 run-phase as a Go function in `internal/harness/seeds/loader.go`.

### 4.4 W4 boundary

W3 ships `.claude/skills/moai-meta-harness/seeds/` directory + `.moai/harness/seeds/` directory **both with `.gitkeep` placeholder**. Actual seed content (8 baseline files per Vision §5 W4) is W4 deliverable. W3 acceptance test (AC-HRA-007) uses synthetic seed fixtures in `internal/harness/seeds/testdata/`, NOT production seeds.

---

## 5. Proposal Throttling (§6.6)

### 5.1 4 modes (REQ-HRA-025..028)

`.moai/config/sections/workflow.yaml` extension (verbatim from Vision §6.6):

```yaml
harness:
  proposal:
    mode: immediate | batch | quiet  # default: immediate
    batch:
      window: weekly                  # weekly | sprint_end | manual
      max_per_window: 5
    quiet:
      hours: [18, 9]                   # 18:00 ~ next 09:00 quiet
      timezone: Asia/Seoul
    mute:
      categories: [error-handling]    # mute specific categories
    cooldown_hours: 24                # per-proposal cooldown
```

### 5.2 Mode behavior

| Mode | Trigger condition | AskUserQuestion timing |
|------|-------------------|------------------------|
| `immediate` | Tier 4 reached AND L1-L4 pass | Immediately (via blocker report) |
| `batch` | Tier 4 reached AND L1-L4 pass | Queued; emitted at window boundary |
| `quiet` | Tier 4 reached + current time outside quiet.hours | Immediately if outside window; deferred if inside |
| `mute` | Tier 4 reached AND category in mute.categories[] | NEVER emitted (logged to evolution-log.md status=muted) |

### 5.3 Quiet hours timezone semantics (EC-HRA-004)

Use `time.LoadLocation(workflow.harness.proposal.quiet.timezone)`. Asia/Seoul has no DST so the window is deterministic. For locales with DST, fall back to absolute UTC offset at session start (documented limitation; explicit DST handling deferred).

### 5.4 Multi-round split for batch mode (Q5 resolution)

When batch mode accumulates 5 proposals but AskUserQuestion limits to 4 questions per round:

- Round 1: First 4 proposals (each as one question)
- Round 2: Remaining 1 proposal + summary
- Automatic split via orchestrator (W3 returns blocker reports in batches of ≤4)

### 5.5 CLI surface for mute management

- `moai harness mute <category>` — append category to mute.categories[]
- `moai harness mute-list` — print current muted categories
- `moai harness unmute <category>` — remove from list

All mutations write through `internal/config/loader.go` with atomic file write (write-tmp + rename).

---

## 6. CLI Surface (§5 W3 + §6.6)

### 6.1 6 verbs (REQ-HRA-029..033, REQ-HRA-036)

| Verb | Purpose | Exit codes |
|------|---------|------------|
| `moai harness status` | Show observation/tier/evolution summary | 0 normal, 1 read failure |
| `moai harness apply <proposal-id>` | Manually trigger 5-Layer for queued proposal | 0 applied, 1 rejected, 2 deferred |
| `moai harness rollback <evolution-id>` | Revert applied evolution | 0 rolled back, 1 not found, 2 read-only filesystem |
| `moai harness disable` | Set `learning.enabled: false` | 0 disabled (with confirmation), 1 user cancel |
| `moai harness mute <category>` / `mute-list` / `unmute <category>` | Manage mute list | 0 success, 1 invalid category |
| `moai harness verify --determinism` | (Placeholder for W4) | 0 with deferred message |

### 6.2 Flag conventions

- `--format json|text` (default text)
- `--strict` (where applicable): treat warnings as errors
- `--dry-run`: show what would happen without writing

### 6.3 JSON output schema (`--format json`)

```json
{
  "status": "ok|error|deferred",
  "verb": "status|apply|rollback|disable|mute|unmute|mute-list|verify",
  "data": {
    // verb-specific payload
  },
  "warnings": []
}
```

For `status` verb data fields:

```json
{
  "observations_total": 0,
  "tier_distribution": { "observation": 0, "heuristic": 0, "rule": 0, "high-confidence": 0, "graduated": 0, "anti-pattern": 0 },
  "recent_evolutions": [],
  "active_learning_total": 0,
  "active_at_limit": false
}
```

### 6.4 Cobra command tree

```
moai harness
├── status
├── apply <proposal-id>
├── rollback <evolution-id>
├── disable
├── mute <category>
├── mute-list
├── unmute <category>
└── verify --determinism
```

Wired via `internal/cli/harness.go` (extend if exists, create if new) + `internal/cli/harness_<verb>.go` per verb (matches W1 `internal/cli/constitution_validate.go` pattern).

---

## 7. Sentinel Catalog Extension

W3 introduces NEW sentinel error codes for the learning subsystem. These ARE ADDITIONS to the 8 HARNESS_FROZEN_* catalog defined in **Vision §3.4** (B1 corrected: catalog is Vision-defined, NOT W1; W3 is the first runtime implementer). The Vision catalog is NOT modified by W3.

| Sentinel key | Source layer | Exit code | Meaning |
|---|---|---|---|
| `HARNESS_LEARNING_FROZEN_BLOCKED` | L1 | 1 | W3 L1 hook matched a HARNESS_FROZEN_* catalog entry from Vision §3.4; W3 wrapper sentinel carries the matched catalog sentinel as `cause` field |
| `HARNESS_LEARNING_CANARY_FAILED` | L2 (synchronous return) | 1 | Canary completed before L5 with FAIL |
| `HARNESS_LEARNING_CANARY_VETO` | L2 (post-L5 async) | 1 (rollback path) | Canary veto after provisional apply |
| `HARNESS_LEARNING_CONTRADICTION` | L3 | 1 (resolved via L3 user choice) | Existing rule conflict |
| `HARNESS_LEARNING_RATELIMIT_EXCEEDED` | L4 | 2 | Weekly/cooldown/active-count exceeded |
| `HARNESS_LEARNING_USER_REJECTED` | L5 | 0 (logged as anti-pattern) | User chose Reject permanently |
| `HARNESS_LEARNING_TIER_VIOLATION` | Tier engine | 1 | Status transition not allowed (e.g., anti-pattern → graduated attempt) |
| `HARNESS_LEARNING_SCHEMA_DRIFT` | observations.yaml parse | 1 | Schema field name mismatch (e.g., `created_at` instead of `created`) |
| `HARNESS_LEARNING_SEED_INVALID` | Cold-start seed load | 1 | Seed YAML malformed or missing required field |
| `HARNESS_LEARNING_MUTE_INVALID_CATEGORY` | CLI mute | 1 | Category not in enum [error-handling, naming, testing, architecture, security, performance, hardcoding, workflow] |

Defense in depth: W3 introduces a CI guard (`internal/harness/sentinel_catalog_test.go`) that verifies the catalog list matches the documented entries in this plan.md §7 — prevents silent drift.

---

## 8. Milestones (Priority-based, no time estimates)

Per W1 single-run-phase pattern, all 6 milestones execute in one run-phase delegation to manager-develop (cycle_type=tdd per quality.yaml).

| Milestone | Priority | Dependencies | Deliverables |
|-----------|----------|--------------|--------------|
| M1 — Lesson Capture Pipeline | P0 | W1 (zone-registry) | `internal/harness/capture/` package + SubagentStop hook integration + observations.yaml write |
| M2 — Tier Engine | P0 | M1 | `internal/harness/tier/` package + state machine + anti-pattern flag |
| M3 — 5-Layer Safety | P0 | M2 + W1 data (zone-registry) | `internal/harness/safety/` package EXTENDED (per §1.5 brownfield inventory; existing canary/contradiction/frozen_guard/oversight/pipeline/rate_limit files extended) + `internal/hook/pre_tool.go` extension (NEW W3 PreToolUse hook implementation per B1) + Canary veto + provisional apply |
| M4 — Throttling + CLI | P0 | M3 | `internal/harness/throttle/` + `internal/cli/harness_*.go` (6 verbs) + workflow.yaml extension |
| M5 — Cold-Start Seed (stub-only) | P1 | M2 | `internal/harness/seeds/` loader interface + STUB `DetectProjectType() = "unknown"` (S10 — marker-based detection deferred to W4) + harness.yaml extension + two empty seeds/ dirs with .gitkeep (per §4.2 dual-path) |
| M6 — End-to-end Test | P0 | M1-M5 | Integration test suite per AC-HRA-001..012, EC-HRA-001..006, R-HRA-001..005 |

Sequential within run-phase (M1 → M2 → M3 → M4 → M5 → M6). Each milestone independently testable (matches incremental layer activation per R7 risk mitigation).

---

## 9. Dependencies & Risk Matrix (Vision §7 + §8)

### 9.1 Dependencies (B1 corrected)

- **Hard dependency**: W1 complete — **DATA SSOT only** (`.claude/rules/moai/core/zone-registry.md` with 111 entries + `internal/constitution/validator.go`). The 8 HARNESS_FROZEN_* catalog is **Vision-defined**, NOT W1; W3 is the first runtime implementer. PreToolUse hook does NOT exist before W3 per W1 EXCL-001 disclaimer.
- **Soft dependency**: W2 complete (moai-foundation-quality preload — affects expert agent invocation reliability)
- **No dependency on W4**: W4 (PROJECT-MEGA) is downstream; W3 produces the autonomy mechanism, W4 produces seed library content + meta-harness 7-Phase + marker-based DetectProjectType (replacing W3 stub)

### 9.2 Risk Matrix (Vision §8 mapped to W3 mitigations)

| ID | Risk | Probability | Impact | Mitigation in W3 |
|---|---|---|---|---|
| R2 | Harness over-aggressive evolution | Medium | High | 5-layer safety (L1 runtime guard + L5 AskUserQuestion + L2 Canary veto). AC-HRA-003 + R-HRA-001 verify gate path. |
| R3 | Frozen Guard false positive | Medium | Medium | W1 already mitigates via `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env. W3 has no own bypass — depends on W1 correctness. |
| R4 | Lesson capture overhead (workflow latency) | Medium | Low | Heuristic match only (REQ-HRA-002 <500ms p95) + background goroutine + no LLM call. R-HRA-002 benchmark verifies. |
| R7 | W3 mechanism complexity | High | Medium | Incremental layer activation — each layer independently testable (M3 task decomposition). R-HRA-003 verifies. |
| R8 | Cold-start regression | High | High | Vision §4.4 seed inject — M5 schema + load hook. W3 ships empty seed dir; W4 fills content. R-HRA-004 fixture verifies seed inject path works even with synthetic seed. |
| R9 | Tier 4 proposal fatigue | Medium | Medium | Vision §6.6 4-mode throttling (M4). R-HRA-005 compares immediate vs batch AskUserQuestion event count. |

### 9.3 W3-specific risks (introduced by this SPEC, NOT in Vision §8)

| ID | Risk | Probability | Impact | Mitigation |
|---|---|---|---|---|
| R11 (S4 corrected) | Canary async race causes false rollback (network/disk latency) | Medium | Medium | 30s soft / 60s hard timeout. **On Canary timeout (>60s), treat as FAIL (auto-rollback)** — this preserves Canary final-gate semantic (per §3.3 B4 corrected). Emit blocker report `HARNESS_LEARNING_CANARY_TIMEOUT` to orchestrator. Orchestrator presents AskUserQuestion: (a) Accept FAIL + remove from cooldown (let user re-propose after diagnostics), (b) Treat as PASS (explicit user override + log warning), (c) Extend timeout once to 180s. Default option (a) preserves Canary final-gate. Note: option (b) is the **only** user-side override of a Canary verdict, and is explicit + logged (NOT silent timeout-as-PASS). |
| R12 | observations.yaml file corruption (concurrent SubagentStop) | Medium | High | `flock(2)` advisory lock per write. EC-HRA-002 parallel-subagent fixture verifies. |
| R13 | Tier 4 backlog growth in batch mode (>50 active) | Low | Medium | Active count cap REQ-HRA-017 (50). Oldest observations archived. EC-HRA-005 fixture. |
| R14 | Seed schema breaking change between W3 and W4 | Low | High | W3 publishes schema (REQ-HRA-022) with `version: 1` field. W4 maintains backward compat or releases as v2. |

---

## 10. Open Questions Resolved Inline

| Q | Resolution |
|---|------------|
| Q1 (PreToolUse perf 10ms cumulative) | NFR: L1 < 10ms p99 per call (REQ-HRA-002 budget). For a typical Tier 4 proposal, L1 fires once per file-write attempt; cumulative impact on workflow latency < 50ms (assuming <5 file writes per evolution). Benchmark via `BenchmarkL1FrozenGuard` in `internal/harness/safety/canary_test.go`. |
| Q2 (Seed library maintenance) | Core repo ships seeds (W4 deliverable). Community contribution mechanism deferred. W3 only defines schema + load hook + library path config (REQ-HRA-022..024). |
| Q4 (Tier threshold for project size) | v3.5.0 uses fixed `[1, 3, 5, 10]` (REQ-HRA-007). Project-size-adaptive threshold deferred to follow-up SPEC after observation data accumulates. |
| Q5 (AskUserQuestion fatigue batch multi-round split) | Multi-round auto-split implemented in §5.4. Orchestrator receives blocker reports in batches of ≤4 (Claude Code AskUserQuestion limit). 5 proposals → 2 rounds (4 + 1 + summary). |
| Q7 (Lesson Capture Trigger Breadth) | SubagentStop only for v3.5.0 (REQ-HRA-001). Manual `moai harness capture <pattern>` CLI deferred. |

---

## 11. Testing Strategy

### 11.1 TDD Cycle (per W1 single-run-phase + quality.yaml development_mode=tdd)

manager-develop cycle_type=tdd. Each milestone follows RED→GREEN→REFACTOR:

- M1 RED: `capture_test.go` — TestCapture_SubagentStopTrigger fixture
- M1 GREEN: minimal capture goroutine + observations.yaml emit
- M1 REFACTOR: error wrapping (`fmt.Errorf("operation: %w", err)`) + lock acquire pattern

Repeat for M2-M5. M6 is integration-test-first (RED for entire pipeline, GREEN incremental).

### 11.2 Unit tests (per package)

- `internal/harness/capture/capture_test.go` — observation emit, SubagentStop dispatch
- `internal/harness/tier/tier_test.go` — state machine table-driven (12 transition cases)
- `internal/harness/safety/{l1,l2,l3,l4,l5}_test.go` — per-layer PASS/FAIL fixture
- `internal/harness/safety/canary_veto_test.go` — provisional apply + auto-rollback
- `internal/harness/throttle/throttle_test.go` — 4 mode table-driven
- `internal/harness/seeds/loader_test.go` — schema decode + inject
- `internal/cli/harness_status_test.go` (and one per verb) — Cobra command surface

### 11.3 Integration tests

- `internal/harness/integration_test.go`:
  - TestTier1ToTier4Progression (AC-HRA-002)
  - TestFrozenGuardViolation (AC-HRA-004)
  - TestUserRejectPermanent (AC-HRA-005)
  - TestThrottlingFourModes (AC-HRA-006)
  - TestColdStartSeedInject (AC-HRA-007)
  - TestCanaryVetoProvisionalRollback (AC-HRA-008)
  - TestEvolutionLogAppendOnly (AC-HRA-010)
  - TestAntiPatternAutoFlag (AC-HRA-011)
  - TestObservationsSchemaCanonical (AC-HRA-012)

### 11.4 Benchmarks

- `BenchmarkLessonCapture` (10MB synthetic diff) — verify <500ms p95 (REQ-HRA-002)
- `BenchmarkL1FrozenGuard` (single proposal) — verify <10ms p99 (Q1)
- `BenchmarkL3Contradiction` (50-entry observations + 5-skill body) — verify <1s
- `BenchmarkL4RateLimit` (single check) — verify <100ms

### 11.5 Coverage target

`internal/harness/` package coverage ≥ 85% per quality.yaml. Verified via `go test -cover ./internal/harness/...`.

### 11.6 plan-auditor checkpoints

This plan.md will be reviewed by plan-auditor at Phase 2.3 of `/moai run`. Anticipated audit dimensions (per W1 iter 1 findings):

- D1 (Brief Quality): 38 REQs / 14 ACs (12 + 008b + 013 + 014) / 6 EC / 5 R / 1 C-binary / EXCL-HRA-001..010 explicit
- D2 (Phase Decomposition): 6 milestones with explicit dependencies + brownfield consolidation strategy (b) per §1.5
- D3 (Risk Management): 6 Vision §8 risks + 4 W3-specific risks mapped to mitigations + R11 timeout corrected (S4)
- D4 (Frontmatter Compliance): 12-field canonical schema, `created:`/`updated:`/`tags:` strict; Field Naming Policy explicit per §1.7
- D5 (Exclusion Discipline): 10 EXCL-HRA-* explicit + B1/B3 boundary corrections
- D6 (Lint Baseline): expects baseline parity (no new lint findings on this SPEC text)

---

## 12. Implementation Hints (Go package layout) — B2 brownfield-aware

**Consolidation Strategy (b) per spec.md §1.5**: Extend `internal/harness/safety/` subdirectory (package `safety`) for W3 5-Layer concerns. `internal/harness/layer{1,2,3,5}.go` (package `harness`, TRIGGER VERIFIER for my-harness-* skill frontmatter) is PRESERVED unchanged — different concern, different package namespace.

```
internal/harness/                              # (existing — mixed root + safety subdir)
│
│   # ── ROOT NAMESPACE (package `harness`) — TRIGGER VERIFIERS, PRESERVED ──
├── layer1.go                # PRESERVE — package `harness` — VerifyTriggers for my-harness-* SKILL.md
├── layer1_test.go           # PRESERVE
├── layer2.go                # PRESERVE — same concern (trigger verification)
├── layer2_test.go           # PRESERVE
├── layer3.go                # PRESERVE
├── layer3_test.go           # PRESERVE
├── layer5.go                # PRESERVE
├── layer5_test.go           # PRESERVE
│   # NOTE: no layer4.go in root (intentional — see §1.5)
│
│   # ── ROOT NAMESPACE (package `harness`) — EXTEND for tier engine + capture ──
├── applier.go               # EXTEND — add provisional/applied/vetoed_by_canary status
├── learner.go               # EXTEND — wire to capture pipeline
├── observer.go              # EXTEND — wire to SubagentStop hook
│   ... (~20 other PRESERVE files, out of W3 scope)
│
│   # ── NEW W3 SUBPACKAGES ──
├── capture/                 # NEW
│   ├── capture.go           # SubagentStop dispatch + observation emit
│   └── capture_test.go
├── tier/                    # NEW
│   ├── tier.go              # state machine + tier transitions
│   ├── observations.go      # YAML schema + flock(2) lock
│   └── tier_test.go
│
│   # ── EXTEND EXISTING SUBPACKAGE (package `safety`) — 5-Layer Safety ──
├── safety/                  # EXTEND existing subdirectory
│   ├── pipeline.go          # EXTEND — existing L1..L5 orchestrator; add Canary veto + provisional apply
│   ├── pipeline_test.go     # EXTEND
│   ├── frozen_guard.go      # EXTEND — wire zone-registry consumer (B1)
│   ├── frozen_guard_test.go # EXTEND
│   ├── canary.go            # EXTEND — add veto power + auto-rollback
│   ├── canary_test.go       # EXTEND
│   ├── contradiction.go     # EXTEND — emit blocker report (B5)
│   ├── contradiction_test.go # EXTEND
│   ├── oversight.go         # EXTEND — 4-option blocker report format
│   ├── oversight_test.go    # EXTEND
│   ├── rate_limit.go        # EXTEND — add 48h cooldown post-Canary-veto
│   ├── rate_limit_test.go   # EXTEND
│   └── canary_veto.go       # NEW — provisional apply + auto-rollback resolver
│
├── throttle/                # NEW
│   ├── throttle.go          # 4-mode dispatcher
│   └── throttle_test.go
├── seeds/                   # NEW
│   ├── schema.go            # seed YAML struct
│   ├── loader.go            # SSOT/cache dual-path lookup + STUB DetectProjectType (S10)
│   ├── loader_test.go
│   └── testdata/            # synthetic seed fixtures
├── sentinel_catalog_test.go # NEW — CI guard for sentinel list ↔ plan.md §7 alignment
├── subagent_boundary_test.go # NEW — C-HRA-008 binary verification (S5 grep negative test)
└── integration_test.go      # NEW — end-to-end M6

internal/cli/
├── harness.go               # NEW or EXTEND existing — parent command
├── harness_status.go        # NEW
├── harness_apply.go         # NEW
├── harness_rollback.go      # NEW
├── harness_disable.go       # NEW
├── harness_mute.go          # NEW
├── harness_verify.go        # NEW — placeholder for W4 determinism
└── harness_*_test.go        # NEW

internal/hook/
├── pre_tool.go              # EXTEND existing (20,548 bytes since 2026-05-18) — add HARNESS_FROZEN_* catalog + harness-learner identity gate (B1: W3 FIRST implementer per W1 EXCL-001)
└── subagent_stop.go         # EXTEND existing (5,765 bytes) — add harness-learner capture pipeline dispatch

.moai/config/sections/
├── harness.yaml             # EXTENDED: seeds.library_path
└── workflow.yaml          # EXTENDED: harness.proposal.*

.claude/skills/moai-meta-harness/seeds/
└── .gitkeep               # placeholder (content = W4)

.moai/harness/
├── observations.yaml      # created on first capture
├── evolution-log.md       # created on first evolution
├── anti-patterns.yaml     # created on first anti-pattern
├── proposal-queue.yaml    # created on first batch-mode proposal
└── revert/                # snapshot dir for rollback
```

---

## 13. Scope Boundaries

### 13.1 Out of Scope (cross-reference with spec.md §4)

See `spec.md` §4 for the canonical exclusion list (EXCL-HRA-001 through EXCL-HRA-010). Plan-specific reminder:

- Determinism (Vision §3.5) — W4 scope. W3 implements CLI verb skeleton only (`verify --determinism` deferred message).
- Seed library content (8 baseline files) — W4 scope. W3 ships empty seed dir.
- meta-harness 7-Phase workflow — W4 scope. W3 invokes existing skill body only.
- LLM-based lesson capture — deferred (heuristic only in W3).

### 13.2 FROZEN constraints (do NOT modify)

- `harness.yaml` `learning.tier_thresholds: [1, 3, 5, 10]` — strictly preserved per REQ-HRA-007.
- The 8 HARNESS_FROZEN_* sentinel catalog (Vision §3.4-defined) — W3 extends with 10 HARNESS_LEARNING_* additive sentinels; the Vision catalog itself is not modified.
- spec-frontmatter-schema.md 12-field canonical names — `created:`/`updated:`/`tags:` only.

---

## 14. 산출물 요약 (Summary of Deliverables)

| Milestone | Files | LOC estimate |
|-----------|-------|--------------|
| M1 — Lesson Capture | 2 new (capture.go + test) + 1 modified (subagent_stop.go) | ~250 LOC |
| M2 — Tier Engine | 3 new (tier.go + observations.go + test) | ~400 LOC |
| M3 — 5-Layer Safety | 7 new (per-layer files + canary_veto + test) | ~900 LOC |
| M4 — Throttling + CLI | 8 new (throttle + 7 CLI files) + 2 modified (workflow.yaml + harness.yaml schema) | ~600 LOC |
| M5 — Cold-Start Seed | 3 new (schema + loader + test) + 1 new dir (.gitkeep) | ~200 LOC |
| M6 — Integration Test | 1 new (integration_test.go) | ~500 LOC |
| **Total** | **~25 files** | **~2850 LOC (code + tests)** |

---

## 15. 후속 SPEC 연결 (Dependencies)

- **Unblocks**:
  - SPEC-V3R5-PROJECT-MEGA-001 (W4): seed library content + meta-harness 7-Phase + project-specific my-harness generation depend on W3's tier engine + 5-Layer safety being operational

- **Depends on**:
  - SPEC-V3R5-CONSTITUTION-DUAL-001 (W1, COMPLETE): zone-registry 111 entries + `internal/constitution/validator.go` — DATA SSOT only per W1 EXCL-001 (PreToolUse hook implementation is W3 scope; sentinel catalog is Vision §3.4-defined)

- **Parallel** (W2 was completed before W3 plan-phase; no actual race):
  - SPEC-V3R5-CORE-SLIM-001 (W2, COMPLETE): expert agent preload — does not block W3

- **Follow-up SPECs (v3.5.0 후)**:
  - SPEC-V3R5-PROJECT-MEGA-001 (W4, planned): determinism + seed library content
  - Hypothetical SPEC-V3R5-HARNESS-LLM-CAPTURE-001 (deferred): LLM-based lesson capture upgrade
  - Hypothetical SPEC-V3R5-HARNESS-COMMUNITY-001 (deferred): cross-project harness sharing
