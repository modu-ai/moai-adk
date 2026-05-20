---
id: SPEC-V3R5-HARNESS-AUTONOMY-001
title: "Harness Autonomy — 4-Tier Self-Evolution + 5-Layer Safety + Cold-Start Seeds"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.5.0"
module: "internal/harness + internal/harness/safety + internal/hook + internal/cli"
lifecycle: spec-anchored
tags: "harness, autonomy, learner, safety, frozen-guard, canary, mega-sprint, w3"
issue_number: 1022
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via Claude Code orchestrator on issue #1022) | Initial draft — Mega-Sprint W3 — T2 Standard scope per orchestrator AskUserQuestion. Derived from `.moai/research/harness-autonomy-vision-2026-05-18.md` §3.4 / §6 (iteration 3, plan-auditor reviewed). |

---

## 1. 개요 (Overview)

Mega-Sprint W3의 자율 진화 메커니즘 구현. v3.5.0 Two-Zone Architecture (`harness-autonomy-vision-2026-05-18.md` §3.1)의 EVOLVABLE 영역(My Harness)이 사용자 워크플로우 관찰을 통해 점진적으로 자기 개선할 수 있도록, **4-Tier Evolution Pipeline** + **5-Layer Safety** + **Cold-Start Seeds** 인프라를 본격 구축한다.

W1 (CONSTITUTION-DUAL-001, COMPLETE)이 정한 zone classification(111 entries)을 데이터 SSOT로 소비하여 L1 Frozen Guard runtime hook이 my-harness-* 학습자의 Core 경로 writes를 차단한다. W2 (CORE-SLIM-001)와 병렬 실행 가능하며, W4 (PROJECT-MEGA-001)의 seed library content / meta-harness 7-Phase / deterministic generation 보장은 본 SPEC scope에서 제외한다.

본 SPEC의 목표:

1. **M1 — Lesson Auto-Capture**: SubagentStop hook trigger + `harness-learner` background scan + heuristic pattern extraction → `.moai/harness/observations.yaml` 누적
2. **M2 — 4-Tier Pipeline**: 1x → 3x → 5x → 10x observation 진급 + anti-pattern detection + Tier 4 도달 시 proposal 생성
3. **M3 — 5-Layer Safety** (브라운필드 EXTEND): L1 Frozen Guard (W3 PreToolUse hook, sync <10ms p99) + L2 Canary (async ~30s shadow eval, **veto power** post-L5) + L3 Contradiction Detector + L4 Rate Limiter + L5 Human Oversight (blocker report → orchestrator AskUserQuestion)
4. **M4 — Proposal Throttling**: immediate / batch / quiet / mute 4 modes + `.moai/config/sections/workflow.yaml` harness.proposal section
5. **M5 — Cold-Start Mitigation**: seed schema + load hook (W3) — seed library content(8 baseline seeds)는 W4 scope
6. **M6 — CLI**: `moai harness {status, apply, rollback, disable, mute, verify}` 6 verbs (+ `mute-list`, `unmute`)

본 SPEC은 W4 PROJECT-MEGA-001의 전제이며 (seed inject point + harness-learner mechanism), v3.5.0 vision의 **vision core**이다.

## 1.5 Brownfield Inventory (B2 resolution)

empirical re-audit on main HEAD `7bd23bb69` (post-W2 sync) — `internal/harness/` 30+ existing files, `internal/hook/pre_tool.go` 653 LOC existing. W3 strategy = **(b) EXTEND**, not (a) GREENFIELD.

### 1.5.1 Existing `internal/harness/` files (preserve unless noted)

| File | LOC | Package | W3 Action |
|------|-----|---------|-----------|
| `applier.go` + `_test.go` | ~ | `harness` | **PRESERVE** (my-harness-* trigger applier — separate concern from W3 safety pipeline) |
| `chaining_rules.go` + `_test.go` | ~ | `harness` | **PRESERVE** |
| `cleanup.go` | ~ | `harness` | **PRESERVE** |
| `e2e_ios_test.go` | ~ | `harness` | **PRESERVE** |
| `evaluator_leak.go` + `_test.go` | ~ | `harness` | **PRESERVE** |
| `frozen_guard.go` | ~ | `harness` | **PRESERVE** (existing package-level guard — W3 hook layer is the new entry point, this file is the CI defense-in-depth) |
| `integration_test.go` | ~ | `harness` | **PRESERVE** |
| `interview.go` + `interview_test.go` + `interview_writer.go` | ~ | `harness` | **PRESERVE** (meta-harness interview — W4 scope) |
| `layer1.go` / `layer2.go` / `layer3.go` / `layer5.go` (+ `_test.go`) | ~ | `harness` | **PRESERVE** (my-harness-* trigger verifier layers — different abstraction; W3 5-Layer pipeline lives in `safety/`) |
| `learner.go` + `learner_test.go` | ~ | `harness` | **EXTEND** — add Tier 1-4 progression machinery |
| `meta_invocation_test.go` | ~ | `harness` | **PRESERVE** |
| `observer.go` + `observer_test.go` | ~ | `harness` | **EXTEND** — add lesson-auto-capture pipeline (heuristic pattern extraction, no LLM call) |
| `prefix_conflict.go` + `_test.go` | ~ | `harness` | **PRESERVE** |
| `profile_loader.go` + `_test.go` | ~ | `harness` | **PRESERVE** |
| `retention.go` | ~ | `harness` | **PRESERVE** |
| `router/` | ~ | `router` | **PRESERVE** |
| `rubric.go` + `_test.go` | ~ | `harness` | **PRESERVE** |
| `safety_preservation_test.go` | ~ | `harness` | **EXTEND** — add invariants for 5-Layer pipeline preservation |
| `scorer.go` / `scorer_engine.go` (+ tests) | ~ | `harness` | **PRESERVE** |
| `session_replay_test.go` | ~ | `harness` | **PRESERVE** |
| `testdata/` | ~ | — | **EXTEND** — add W3 fixtures (frozen-path matrix, canary regression fixtures, contradiction conflict fixtures) |
| `types.go` + `types_extension_test.go` | ~ | `harness` | **EXTEND** — add `Observation`, `Proposal`, `EvolutionRecord`, `SafetyVerdict` types |

### 1.5.2 Existing `internal/harness/safety/` files (EXTEND — W3 core)

| File | Package | W3 Action |
|------|---------|-----------|
| `canary.go` + `_test.go` | `safety` | **EXTEND** — add async shadow-eval driver, Canary Veto Policy E5 |
| `contradiction.go` + `_test.go` | `safety` | **EXTEND** — add blocker report pattern (no AskUserQuestion inside subagent) |
| `frozen_guard.go` + `_test.go` | `safety` | **EXTEND** — wire to W1 zone-registry data, 8 sentinel catalog (`HARNESS_FROZEN_*_VIOLATION`) |
| `oversight.go` + `_test.go` | `safety` | **EXTEND** — return blocker reports; orchestrator-side AskUserQuestion is out of subagent scope |
| `pipeline.go` + `_test.go` | `safety` | **EXTEND** — wire L1→L2→L3→L4→L5 sequential, async L2 with veto post-L5 |
| `rate_limit.go` + `_test.go` | `safety` | **EXTEND** — 3/week + 24h cooldown + 50 active max, persistent state in `.moai/harness/rate-limit-state.json` |

### 1.5.3 Existing `internal/hook/pre_tool.go` (653 LOC) — EXTEND

W3가 추가하는 변경 영역:
- 8 sentinel catalog (`HARNESS_FROZEN_*_VIOLATION`) 등록
- `harness-learner` / `my-harness-*` agent identity gate (PreToolUse hook payload의 `agent_name` 추출 + fallback `MOAI_INVOKING_AGENT` env)
- path-glob deny matcher (`.claude/agents/moai/**`, `.claude/skills/moai-*/**` (단 `my-harness-*` ALLOW), `.claude/rules/moai/**`, etc.)
- sync <10ms p99 NFR (R-HRA-S1)
- `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env 우회 (CLI 내부 only)

### 1.5.4 New packages (W3 greenfield)

| Path | Purpose |
|------|---------|
| `internal/harness/capture/` | Lesson auto-capture pipeline (heuristic pattern extraction) |
| `internal/harness/tier/` | 4-Tier progression machinery (1x → 3x → 5x → 10x) + anti-pattern detector |
| `internal/harness/throttle/` | Proposal throttling (immediate / batch / quiet / mute) |
| `internal/harness/seeds/` | Seed schema + load hook (W3) — seed library content is W4 scope |

## 1.6 D11 Seed Decision

본 SPEC은 **seed schema + load hook (W3)** 만 구현하고, **seed library content (8 baseline seeds: Go/Node/Python/Rust/React/Vue/Flutter/iOS)는 W4 PROJECT-MEGA-001 scope로 이관**한다. Rationale: W3는 자율 진화 mechanism의 substrate; seed content는 meta-harness 7-Phase workflow(W4)와 결합되어야 의미가 있다. W3에서는 dummy seed 1개로 contract test 통과만 보장한다.

## 1.7 Field Naming Policy

본 SPEC의 모든 YAML data files (`.moai/harness/observations.yaml`, `.moai/harness/evolution-log.md`, `.moai/harness/rate-limit-state.json`)는 canonical `created:` / `updated:` field names를 사용한다 (`created_at:` / `updated_at:` 금지). Rationale: vision §6.7 D3 fix per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT alignment. `harness_schema_test.go`가 이 invariant를 검증한다.

## 2. 배경 (Background)

### 2.1 v3.5.0 Two-Zone Architecture 비전

`harness-autonomy-vision-2026-05-18.md` (D1-D10 v2 + E1-E5 iter3)는 MoAI-ADK를 두 영역으로 분리한다:

- **Core MoAI (FROZEN)**: 17 canonical agents + 29 skills + 헌법 규칙. 학습자(learner)가 수정 불가
- **My Harness (EVOLVABLE)**: 프로젝트별 3-7개 customizable agents + 운영 튜닝 규칙. 점진적 학습 허용

이 분리는 **runtime enforcement** 없이는 환각 harness-learner를 차단할 수 없다. W1이 헌법 마킹/registry를 제공했고, W3는 이 데이터를 소비하는 PreToolUse hook을 구현한다.

### 2.2 W1 의존성 (Hard, COMPLETE)

W1 (CONSTITUTION-DUAL-001, COMPLETE on main `7bd23bb69`)이 제공하는 자원:

- `.claude/rules/moai/core/zone-registry.md` 111 entries (CONST-V3R2-001..152 + CONST-V3R5-001..039) — `zone_class` 4-enum (`frozen-canonical`, `frozen-safety`, `evolvable-tuning`, `evolvable-experimental`)
- `moai constitution validate` CLI verb (sentinel: `DRIFT`, `ZONE_UNREGISTERED`, etc.)
- 15 source files × 111 [HARD] rules × 100% `[ZONE:Frozen]`/`[ZONE:Evolvable]` marker coverage

**W1 EXCL-001 per W1 spec**: W1은 데이터 SSOT만 제공; runtime enforcement는 W3 scope. 따라서 W3 L1 Frozen Guard hook은 W1의 zone-registry를 읽기만 한다 (수정 금지).

### 2.3 W2 병렬성

W2 CORE-SLIM-001은 W3와 race 없이 병렬 가능 (D5 resolution). W2는 expert-* agent retirement + lint baseline; W3는 hook layer + harness-learner machinery. Touch points 없음.

### 2.4 5-Layer Safety의 정당성 (vision §6.5)

Frontmatter flag만으로 (e.g., agent definition의 `zone: harness`) 환각 학습자를 차단할 수 없다. Runtime layer가 필요한 이유:

1. **L1 (Frozen Guard, sync <10ms)**: 환각 harness-learner가 Core path write 시도 시 즉시 차단
2. **L2 (Canary, async ~30s + VETO)**: 학습 변경이 마지막 3 프로젝트에 regression 야기하면 자동 롤백
3. **L3 (Contradiction, sync <1s)**: 새 학습이 기존 rule과 충돌하면 blocker report로 orchestrator에 위임
4. **L4 (Rate Limiter, sync <100ms)**: 3/week + 24h cooldown으로 evolution velocity 제한
5. **L5 (Human Oversight, user-paced)**: 모든 학습 적용은 사용자 승인 (blocker report 패턴)

**Canary Veto Policy (E5 resolution)**: L2는 async이므로 L5 user approval보다 늦게 완료될 수 있다. 정책:
1. L5 approval이 Canary 완료 전 도착 → **provisional apply** (file write, status=`provisional`)
2. Canary 완료 시 PASS → `provisional → applied`; FAIL → **automatic rollback** + AskUserQuestion notification
3. Canary는 L5 approval에 대한 **veto power 보유** — user approval이 final이 아님, Canary가 final gate

### 2.5 Cold-Start 문제와 Seed Pattern (vision §4.4)

신규 프로젝트는 observation 0개로 시작하여 Tier 4(10x) 도달까지 수십 SPEC 사이클이 필요하다. Seed는 meta-harness가 generation 시점에 inject하는 **Tier 3 starting point**로 이 cold-start gap을 메운다. 본 SPEC은 seed schema + load hook만 구현; 8 baseline seed content는 W4.

## 3. 기능 요구사항 (Functional Requirements — EARS)

### 3.1 Ubiquitous (REQ-HRA-001..005)

- **REQ-HRA-001**: The system **shall** maintain `.moai/harness/observations.yaml` as the SSOT for all captured observations, conforming to the schema in §1.7 (canonical `created:` / `updated:` field names).
- **REQ-HRA-002**: The system **shall** support 4-Tier classification: `observation` (1x), `heuristic` (3x), `rule` (5x), `high-confidence` (10x) — plus 1x-critical `anti-pattern` and configurable `seed` Tier 3 start.
- **REQ-HRA-003**: The system **shall** execute the 5-Layer Safety pipeline (L1 → L2 → L3 → L4 → L5) sequentially before applying any evolution, except that L2 (Canary) runs async with veto power post-L5.
- **REQ-HRA-004**: The system **shall** persist all evolution actions to `.moai/harness/evolution-log.md` (apply, rollback, defer, reject, canary-veto) with timestamp + sentinel.
- **REQ-HRA-005**: The system **shall** expose 6 CLI verbs under `moai harness {status, apply, rollback, disable, mute, verify}` (+ `mute-list`, `unmute`) with structured JSON output for scripting.

### 3.2 Event-driven (REQ-HRA-006..014)

- **REQ-HRA-006**: When a `SubagentStop` hook fires, the system **shall** invoke `harness-learner` in background (read-only initial scan) to extract heuristic patterns from the completed subagent's diff.
- **REQ-HRA-007**: When a heuristic pattern matches an existing observation, the system **shall** increment `count` and update `updated`; when no match, the system **shall** create a new Tier-1 entry.
- **REQ-HRA-008**: When observation `count` reaches 3, the system **shall** promote status `observation → heuristic`; at 5 → `heuristic → rule`; at 10 → `rule → high-confidence`.
- **REQ-HRA-009**: When a high-confidence observation is reached, the system **shall** generate a proposal in the format of vision §6.4 and dispatch via the throttling mode.
- **REQ-HRA-010**: When a PreToolUse hook fires with `tool ∈ {Write, Edit, MultiEdit}` AND the invoking agent matches `harness-learner` OR `my-harness-*`, the system **shall** run L1 Frozen Guard within 10ms p99.
- **REQ-HRA-011**: When L1 detects a deny-pattern match, the system **shall** reject the tool call and emit JSON `{"status": "denied", "sentinel": "HARNESS_FROZEN_*_VIOLATION", "agent": "<name>", "path": "<path>", "reason": "..."}` with the appropriate sentinel from §5 catalog.
- **REQ-HRA-012**: When L2 Canary shadow-eval completes after L5 provisional apply, the system **shall** transition status `provisional → applied` on PASS OR trigger automatic rollback + AskUserQuestion notification on FAIL (Canary Veto Policy E5).
- **REQ-HRA-013**: When L3 detects rule conflict, the system **shall** return a structured blocker report to the orchestrator (no AskUserQuestion invocation from inside subagent) per `.claude/rules/moai/core/askuser-protocol.md` §Orchestrator–Subagent Boundary.
- **REQ-HRA-014**: When a single critical failure (score drop > 0.20 or must-pass criterion failure) is observed, the system **shall** immediately classify the pattern as `anti-pattern` and FROZEN (only human intervention can reclassify).

### 3.3 State-driven (REQ-HRA-015..020)

- **REQ-HRA-015**: While L4 Rate Limiter shows ≥3 evolutions in the current 7-day window OR <24h since last evolution OR ≥50 active learnings, the system **shall** defer new proposals to the next eligible window.
- **REQ-HRA-016**: While `.moai/config/sections/workflow.yaml` `harness.proposal.mode` is `batch`, the system **shall** queue Tier-4 proposals and dispatch ≤`batch.max_per_window` at `batch.window` boundary.
- **REQ-HRA-017**: While the current local time is within `harness.proposal.quiet.hours` range, the system **shall** suppress Tier-4 proposal dispatch (queue for next non-quiet boundary).
- **REQ-HRA-018**: While a category appears in `harness.proposal.mute.categories`, the system **shall** suppress Tier-4 proposals for that category (silent count++).
- **REQ-HRA-019**: While `.moai/harness/disabled` sentinel file is present, the system **shall** halt all auto-capture, tier-progression, and proposal dispatch (rollback CLI still works).
- **REQ-HRA-020**: While a seed marked `tier: 3` is loaded for a new project, the system **shall** treat the seed as starting point for tier progression (subsequent observations increment from 5x baseline).

### 3.4 Unwanted-behavior (REQ-HRA-021..028)

- **REQ-HRA-021**: If a `harness-learner`-initiated tool call attempts to write a path matching deny-glob, **then** L1 **shall** reject and emit `HARNESS_FROZEN_*_VIOLATION` sentinel without modifying the filesystem.
- **REQ-HRA-022**: If L1 hook latency exceeds 10ms p99 over a rolling 100-sample window, **then** the system **shall** emit `HARNESS_LEARNING_LATENCY_BUDGET_BREACH` sentinel to `.moai/logs/harness/latency.jsonl`.
- **REQ-HRA-023**: If L2 Canary detects any project score drop >0.10, **then** the system **shall** VETO the evolution (rollback if provisional, block if pre-apply) and emit `HARNESS_LEARNING_CANARY_VETO` sentinel.
- **REQ-HRA-024**: If L3 detects contradiction between proposed change and existing rule, **then** the system **shall** return blocker report `{ contradiction: { existing_rule_id, proposed_change, recommendation }}` and emit `HARNESS_LEARNING_CONTRADICTION_DETECTED` sentinel.
- **REQ-HRA-025**: If `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env is set AND invocation chain originates from `moai update` CLI, **then** L1 **shall** allow Core path writes (exception for sanctioned updates only).
- **REQ-HRA-026**: If observation YAML file violates schema (missing canonical `created:`, malformed `category`, invalid enum), **then** the loader **shall** emit `HARNESS_LEARNING_SCHEMA_DRIFT` sentinel and skip the malformed entry (degraded operation, no halt).
- **REQ-HRA-027**: If a rollback CLI invocation references unknown `evolution-id`, **then** the CLI **shall** exit with code `2` and emit `HARNESS_LEARNING_UNKNOWN_EVOLUTION` sentinel.
- **REQ-HRA-028**: If R11 timeout: a Tier-4 proposal AskUserQuestion times out (no user response within 7 days), **then** the system **shall** auto-defer to `quiet` mode + emit `HARNESS_LEARNING_PROPOSAL_TIMEOUT` sentinel.

### 3.5 Optional-feature / Subagent boundary (REQ-HRA-029..033)

- **REQ-HRA-029**: The harness-learner subagent **shall not** invoke `AskUserQuestion` (HARD per `.claude/rules/moai/core/askuser-protocol.md` §Orchestrator–Subagent Boundary).
- **REQ-HRA-030**: The harness-learner subagent **shall** return structured blocker reports for any user-input requirement (per agent-common-protocol §Blocker Report Format).
- **REQ-HRA-031**: The system **may** support per-category mute via `moai harness mute <category>` with persistent state in `.moai/harness/mute-state.json`.
- **REQ-HRA-032**: The system **may** support `harness verify --determinism` as a stub (full impl is W4 scope) — returns "not yet implemented" but does not fail.
- **REQ-HRA-033**: The system **may** load seed library files from `.claude/skills/moai-meta-harness/seeds/` if present (W4 will populate; W3 ships with 1 dummy seed for contract test).

### 3.6 Integrity / Hybrid (REQ-HRA-034..038)

- **REQ-HRA-034**: The L1 Frozen Guard PreToolUse hook **shall** consume W1 `zone-registry.md` data only (no modification, no other source — single SSoT per W1 EXCL-001).
- **REQ-HRA-035**: The 10-sentinel `HARNESS_LEARNING_*` catalog (extension to W3-vision §3.4's 8 `HARNESS_FROZEN_*` set) **shall** be enumerated in plan.md §Sentinel Catalog and CI-tested per sentinel.
- **REQ-HRA-036**: The 5-Layer pipeline ordering **shall** be invariant: L1 → L3 → L4 → L5 sync, L2 async with veto post-L5; reordering requires plan-auditor approval.
- **REQ-HRA-037**: The system **shall** preserve all `internal/harness/*.go` files in §1.5.1 marked PRESERVE (no deletion or breaking signature change).
- **REQ-HRA-038**: The harness-learner agent definition (`.claude/agents/moai/harness-learner.md`, W3 NEW) **shall** declare `tools: Read, Grep, Glob, Bash` only (no Write/Edit — all evolution writes go through 5-Layer pipeline orchestrated by main session).

## 4. Non-Functional Requirements

| ID | Category | Requirement | Validation |
|----|----------|-------------|------------|
| **R-HRA-S1** | Performance (L1) | L1 Frozen Guard hook MUST complete in <10ms p99 over rolling 100-sample window | `internal/hook/pre_tool_bench_test.go` — `BenchmarkFrozenGuardL1` |
| **R-HRA-S2** | Performance (L3) | L3 Contradiction Detector MUST complete in <1s sync | `internal/harness/safety/contradiction_bench_test.go` |
| **R-HRA-S3** | Performance (L4) | L4 Rate Limiter MUST complete in <100ms sync | `internal/harness/safety/rate_limit_bench_test.go` |
| **R-HRA-S4** | Performance (L2) | L2 Canary async shadow-eval target ≤30s on last 3 projects | `internal/harness/safety/canary_bench_test.go` |
| **R-HRA-Q1** | Quality | All 10 sentinels MUST have ≥1 unit test (`HARNESS_FROZEN_*` 8 + `HARNESS_LEARNING_*` 2) | sentinel grep enforcement test |
| **R-HRA-Q2** | Quality | Subagent boundary: `harness-learner` agent definition MUST NOT contain `AskUserQuestion` reference | `grep -r AskUserQuestion .claude/agents/moai/harness-learner.md` returns 0 hits |
| **R-HRA-A1** | Auditability | Every evolution (apply/rollback/veto/defer/reject) MUST be logged to `.moai/harness/evolution-log.md` | `evolution_log_test.go` per outcome |
| **R-HRA-O1** | Operations | `moai harness disable` MUST halt auto-capture within next hook cycle (no in-flight observation persisted) | `cli_disable_test.go` |
| **R-HRA-I1** | Idempotency | `moai harness rollback <evolution-id>` MUST be idempotent (second call returns "already rolled back", exit 0) | `cli_rollback_test.go` |
| **R-HRA-T1** | Timeout (R11) | AskUserQuestion proposal timeout = 7 days, then auto-defer to `quiet` | `proposal_timeout_test.go` |

## 5. Exclusions (EXCL-HRA-*)

| ID | Exclusion | Rationale | Downstream SPEC |
|----|-----------|-----------|-----------------|
| EXCL-HRA-001 | Deterministic harness generation guarantees (vision §3.5) | meta-harness output determinism is a W4 concern (template-bit-exact + LLM semantic-equiv ≥0.95) | W4 PROJECT-MEGA-001 |
| EXCL-HRA-002 | 8 baseline seed library content (Go/Node/Python/Rust/React/Vue/Flutter/iOS) | seed content is meaningless without W4 meta-harness 7-Phase; W3 ships dummy seed only | W4 PROJECT-MEGA-001 |
| EXCL-HRA-003 | `/moai project --refresh` command | refresh re-runs meta-harness 7-Phase; W3 is substrate only | W4 PROJECT-MEGA-001 |
| EXCL-HRA-004 | meta-harness 7-Phase workflow itself | revfactory/harness adoption is W4 polish | W4 PROJECT-MEGA-001 |
| EXCL-HRA-005 | Project-specific my-harness-* skill/agent generation | meta-harness output is W4; W3 just enforces FROZEN boundary | W4 PROJECT-MEGA-001 |
| EXCL-HRA-006 | Migration / backward-migration tooling (legacy harness → W3 schema) | greenfield substrate; no prior W3 install to migrate from | n/a |
| EXCL-HRA-007 | LLM-based lesson capture (semantic similarity, embedding match) | W3 uses heuristic matching only (no LLM call in capture pipeline per vision §6.1) | future SPEC if telemetry justifies |
| EXCL-HRA-008 | Multi-project harness sharing / sync across repos | per-repo isolated `.moai/harness/` only; cross-repo learning is out of scope | future SPEC |
| EXCL-HRA-009 | Harness-learner self-introspection (learner observes its own behavior) | recursive observation forbidden to prevent feedback loops | n/a (FROZEN exclusion) |
| EXCL-HRA-010 | Programmatic evolution rollback past 30-day window | `staleness_window_days=30` (design.yaml); W3 enforces but does not extend | n/a |

## 6. Dependencies

### 6.1 Hard Dependencies (BLOCKING)

- **W1 SPEC-V3R5-CONSTITUTION-DUAL-001** (COMPLETE) — zone-registry 111 entries data SSOT; L1 reads only per W1 EXCL-001.

### 6.2 Parallel Dependencies (no race)

- **W2 SPEC-V3R5-CORE-SLIM-001** (COMPLETE in main `7bd23bb69`) — no touch point with W3; expert-* retirement vs harness-learner machinery are orthogonal.

### 6.3 Downstream Unblocks

- **W4 SPEC-V3R5-PROJECT-MEGA-001** — needs W3 substrate (5-Layer pipeline, harness-learner agent, seed load hook) before adding meta-harness 7-Phase + 8 seed content + deterministic generation.

## 7. Sentinel Catalog (forward reference)

8 `HARNESS_FROZEN_*_VIOLATION` (vision §3.4) + 2 `HARNESS_LEARNING_*` (W3 extension) = **10 sentinels total**. Full enumeration with grep-stable matcher patterns lives in `plan.md` §Sentinel Catalog. Each sentinel has ≥1 unit test (R-HRA-Q1).

## 8. References

- `.moai/research/harness-autonomy-vision-2026-05-18.md` §3.4 (Frozen Guard runtime), §6.1 (Lesson capture), §6.2 (Observation schema D3), §6.3 (Tier table), §6.4 (Proposal format), §6.5 (5-Layer sequencing + Canary Veto E5), §6.6 (Throttling D7)
- `.claude/rules/moai/core/zone-registry.md` (W1 output, 111 entries) — data SSOT consumed by L1
- `.claude/rules/moai/core/askuser-protocol.md` §Orchestrator–Subagent Boundary (REQ-HRA-029/030 derivation)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical `created:` / `updated:` field naming — §1.7 alignment)
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/spec.md` (W1 precedent, template source)

---

End of spec.md (draft v0.1.0, awaiting plan-auditor iter1/iter2).
