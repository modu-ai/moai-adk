---
id: SPEC-V3R5-HARNESS-AUTONOMY-001
title: "Harness Autonomy — 4-Tier Self-Evolution + 5-Layer Safety + Cold-Start Seeds"
version: "0.3.0"
status: completed
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.5.0"
module: "internal/harness + .moai/harness + .claude/skills/moai-harness-learner + internal/hook"
lifecycle: spec-anchored
tags: "harness, autonomy, self-evolution, 4-tier, 5-layer-safety, cold-start, mega-sprint, w3"
issue_number: 1022
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.3.0 | 2026-05-20 | GOOS Kim (via MoAI) | sync-phase complete — lifecycle close. PR #1023 plan + PR #1024 run merged to main HEAD `24f96e266` (post PR #1025/#1026 WORKFLOW-OPT merge). W3 산출물 main verified: 7 packages (harness/capture/router/safety/seeds/throttle/tier) ≥ 85% coverage (harness 87.9% / capture 94.9% / router 89.2% / safety 86.5% / seeds 100% / throttle 88.2% / tier 90.0%), 18 sentinels catalog (8 HARNESS_FROZEN_* + 10 HARNESS_LEARNING_*), 10 CLI verbs (route/validate + status/apply/rollback/disable/mute/mute-list/unmute/verify per AC-HRA-009 V3R4 reversal), C-HRA-008 binary 0 matches, Cross-platform builds PASS (Windows flock split applied). 메타-분석 결과는 SPEC-V3R5-WORKFLOW-OPT-001 (PR #1025/#1026)로 formalization 완료 — 본 W3 결과가 그 SPEC dogfooding의 기준 (-73% wall-time 검증). Status: implemented → completed. |
| 0.2.0 | 2026-05-20 | GOOS Kim (via MoAI) | run-phase complete — M1 (capture) + M2 (tier engine) + M3 (Canary Veto + HARNESS_FROZEN_* sentinels) + M4 (throttle + mute/unmute/mute-list/verify CLI) + M5 (seed stub) + M6 (integration + sentinel catalog + C-HRA-008 boundary tests). 6 commits on feat/SPEC-V3R5-HARNESS-AUTONOMY-001. Coverage: harness 87.9% / safety 86.5% / seeds 100% / throttle 88.2% / tier 89.2%. Benchmarks: L1 46ns << 10ms p99, L4 1.54µs << 100ms p99. C-HRA-008 PASS (zero non-comment AskUserQuestion references). Status: planned → implemented. |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial draft — Mega-Sprint W3 — T2 Standard scope per user AskUserQuestion. Vision source: `.moai/research/harness-autonomy-vision-2026-05-18.md` (iter3, pending audit). W0/W1/W2 lifecycles complete (main HEAD `7bd23bb69`). Scope: 4-Tier evolution pipeline, 5-layer safety with Canary veto, cold-start seed schema, proposal throttling, 6 CLI verbs. |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Plan-PR opening commit (post-audit cleanup) — plan-auditor iter 2 verdict **PASS aggregate 0.902** (D1=0.86 / D2=0.92 / D3=0.92 / D4=0.96 / D5=0.90 / D6=0.82, all must-pass ≥0.86; W1 precedent recovery delta +0.061). 10 residual B1 narrative carryovers (R1-R10, R12 from iter 2 audit findings) mechanically cleaned across all 4 files in single plan-PR opening commit per plan-auditor recommendation (vs separate iter 3 PR). All catalog references now correctly attribute Vision §3.4 as 8 HARNESS_FROZEN_* sentinel source per W1 EXCL-001. R7 critical (acceptance.md §5.4 broken verification step `parsing W1 spec.md §3.4`) replaced with executable `grep -c HARNESS_FROZEN .moai/research/harness-autonomy-vision-2026-05-18.md ≥ 8` (verified returns 11). Zero design impact. issue_number 1022 added (#1022). |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision — plan-auditor iter 1 BLOCKING (B1-B5) + SHOULD (S1/S3/S4/S5/S6/S10) defects resolved. B1: W1 deliverable corrected — W1 ships zone-registry DATA SSOT only; W3 IMPLEMENTS PreToolUse hook code + 8 HARNESS_FROZEN_* runtime sentinels (rewrote §1.3, §2.3, §3.3 REQ-HRA-008/009, §6 References; AC-HRA-004 precondition corrected). B2: §1.5 Brownfield State Inventory added — 30+ existing files in `internal/harness/` (layer1-5 root namespace are TRIGGER VERIFIERS for my-harness-* skill frontmatter, package `harness`) and `internal/harness/safety/` (canary/contradiction/frozen_guard/oversight/pipeline/rate_limit, package `safety`) coexist legitimately; consolidation strategy (b) "extend safety/ subdirectory; root layer*.go untouched (different concern)". REQ-HRA-038 added. B3: §1.6 D11 Seed Location Resolution — SSOT `.claude/skills/moai-meta-harness/seeds/` (Core repo) + cache `.moai/harness/seeds/` (W4 project-local); REQ-HRA-024 updated with precedence. B4: AC-HRA-008b binary cooldown rejection assertion added; acceptance.md notification rephrased to remove "Override" wording per Canary final-gate semantic. B5: REQ-HRA-014 rewritten — L3 emits blocker report (NOT AskUserQuestion); plan.md §3.3 documents L3+L5 unified blocker pattern; AC C-HRA-008 binary verification added (S5). S1: §1.7 Field Naming Policy explicit. S3: REQ-HRA-007 enforcement path + AC-HRA-013. S4: R11 mitigation rewritten — timeout=FAIL (auto-rollback) preserves Canary final-gate. S6: REQ-HRA-037 L1 perf NFR + AC-HRA-014. S10: M5 stub-only project-type detection (plan.md §4.2). Total: 4 new REQs (037/038) + 3 new ACs (008b/013/014) + 1 new C (C-HRA-008) + 2 new sub-sections (§1.5, §1.6, §1.7). |

---

## 1. 개요 (Overview)

Mega-Sprint W3의 **하네스 자율 진화 메커니즘**. v3.5.0 Two-Zone Architecture (`.moai/research/harness-autonomy-vision-2026-05-18.md` §3.1) 의 두 영역 중 **My Harness (EVOLVABLE)** layer가 워크플로우 실행 중 지속적으로 학습하고 자가 진화하는 메커니즘을 구현한다.

### 1.1 사용자 directive (verbatim, 2026-05-18 세션)

> "harness는 프로젝트 맞게 생성하고 나서 사용자와 프로젝트 개발을 하면서 스스로 학습해서 자율적으로 하네스가 진화 업데이트가 되도록 설계가 되어야 한다. 이걸 꼭 제대로 설계해서 워크플로우 진행시 지속적으로 학습과 업데이트가 되도록 하자. 하네스 스킬/에이전트 등 하네스만 자율 진화 하도록 한다. 그래서 기본 moai-* 스킬/에이전트/커먼드/rules등은 범용적으로 사용이 가능하고 최소화 하고 나머지는 프로젝트에 맞게 하네스 설치가 되어서 프로젝트에 맞는 개발 환경을 제공하는것이 moai-adk 이번 대규모 업데이트의 목적이다."

### 1.2 W3 mission statement

워크플로우 (`/moai plan`, `/moai run`, `/moai sync`, `/moai fix`, `/moai loop`) 실행 중 발생하는 lesson을 자동 포착하여 4-Tier 진화 파이프라인에 누적하고, Tier 4 도달 시 5-Layer Safety를 통과한 proposal을 사용자 승인 후 my-harness-* skill/agent body에 자율 적용한다. **Core MoAI는 변경하지 않는다** (W3 §3.3 PreToolUse Frozen Guard hook + Vision §3.4 8 HARNESS_FROZEN_* sentinel catalog가 침범 차단; W1은 zone-registry data SSOT 제공만 per W1 EXCL-001).

### 1.3 현재 상태 (empirical, main HEAD `7bd23bb69`)

- `.claude/skills/moai-harness-learner/SKILL.md` (6,454 bytes) — skill body 존재, but **autonomous evolution mechanism 미구현**
- `.claude/skills/moai-meta-harness/SKILL.md` (16,825 bytes) — meta-harness skill 존재, but **cold-start seed library 미존재**
- `.moai/config/sections/harness.yaml` `learning` section 존재 (`tier_thresholds: [1, 3, 5, 10]` + `rate_limit: max_per_week=3, cooldown_hours=24` + `auto_apply: false`)
- **W1 deliverables (DATA SSOT only)**: `internal/constitution/validator.go` + `.claude/rules/moai/core/zone-registry.md` (111 entries). W1 explicitly disclaimed PreToolUse hook implementation per W1 spec.md §5.2 EXCL-001: "W1은 PreToolUse hook scaffold를 구현하지 않는다. 해당 작업은 W3 HARNESS-AUTONOMY-001의 scope. 단, W3 hook이 W1의 zone-registry를 참조 데이터로 사용할 수 있도록 SSOT을 제공한다."
- **W3 implements (NEW in this SPEC)**: PreToolUse Frozen Guard hook code (`internal/hook/pre_tool.go` extension), 8 HARNESS_FROZEN_* runtime sentinels (Vision §3.4 catalog), 10 HARNESS_LEARNING_* learning sentinels (additive). L1 layer **consumes zone-registry as input data**, NOT "consults W1's hook".
- `.moai/harness/` 디렉토리: **부재** (W3에서 생성)

### 1.4 본 SPEC의 목표 (T2 Standard scope)

1. **M1**: Lesson auto-capture pipeline (SubagentStop hook trigger + heuristic matching, no LLM call)
2. **M2**: Tier engine (1x→3x→5x→10x progression + anti-pattern auto-flag)
3. **M3**: 5-Layer Safety sequencing (L1 Frozen Guard sync + L2 Canary async with **veto power** per E5 + L3 Contradiction + L4 Rate Limiter + L5 AskUserQuestion)
4. **M4**: Proposal Throttling (immediate/batch/quiet/mute 4 modes) + 6 CLI verbs
5. **M5**: Cold-Start Seed schema + load hook (W3 defines schema + load mechanism; seed library content is W4 scope per Out of Scope §4)
6. **M6**: End-to-end test (Tier 1→4 progression + 5-Layer safety simulation + Frozen Guard violation block + cold-start regression)

본 SPEC은 W1 (FROZEN/EVOLVABLE 경계) 의존, W2 (CORE-SLIM) 와 병렬 가능 (vision §5 D5 resolution).

### 1.5 Brownfield State Inventory (B2 resolution)

본 SPEC은 brownfield 작업이다. `internal/harness/` 디렉토리에 30+ 기존 파일이 존재하며 두 개의 별개 패키지 namespace로 분리되어 있다. iteration 2에서 empirical verification 완료 (main HEAD `7bd23bb69`).

**Critical Discovery**: `internal/harness/layer{1,2,3,5}.go` (root) 와 `internal/harness/safety/{canary,contradiction,frozen_guard,oversight,pipeline,rate_limit}.go` (subdirectory) 는 **다른 concern**이다 — 같은 이름이지만 충돌하지 않는다:

- **`internal/harness/layer1.go` (package `harness`)**: my-harness-* skill frontmatter의 **trigger verification** (REQ-PH-008 관련). `VerifyTriggers(skillPath string) error` 함수 등. 5-Layer Safety와 무관한 다른 SPEC의 산출물.
- **`internal/harness/safety/*.go` (package `safety`)**: 본 SPEC W3가 확장할 **5-Layer Safety implementation**. `safety.Pipeline.Evaluate()` 가 L1→L2→L3→L4→L5 진입점.

#### 기존 파일 분류 (DELTA marker per W1 §6 pattern)

| Marker | 파일 / 디렉토리 | 분류 | 근거 |
|--------|-----------------|------|------|
| [EXTEND] | `internal/harness/safety/pipeline.go` | EXTEND | 기존 L1-L5 entry point 존재. W3 Canary veto + provisional apply 로직 추가. |
| [EXTEND] | `internal/harness/safety/canary.go` (3,815 bytes) | EXTEND | 기존 Canary L2 존재. W3 veto power + auto-rollback 로직 추가. |
| [EXTEND] | `internal/harness/safety/contradiction.go` (4,926 bytes) | EXTEND | 기존 L3 존재. W3 blocker report emission 패턴 추가 (B5 resolution). |
| [EXTEND] | `internal/harness/safety/frozen_guard.go` (3,967 bytes) | EXTEND | 기존 frozen guard. W3가 zone-registry consumer로 wire (B1 resolution). |
| [EXTEND] | `internal/harness/safety/oversight.go` (3,189 bytes) | EXTEND | 기존 L5 blocker report emitter. W3 4-option choice 포맷 보강. |
| [EXTEND] | `internal/harness/safety/rate_limit.go` (4,846 bytes) | EXTEND | 기존 L4. W3 48h cooldown post-Canary-veto entry 추가. |
| [PRESERVE] | `internal/harness/layer1.go` (2,568 bytes) | PRESERVE | TRIGGER VERIFIER for my-harness-* skill frontmatter — W3 scope 외, 변경 없음. |
| [PRESERVE] | `internal/harness/layer2.go` / `layer3.go` / `layer5.go` | PRESERVE | 동일 — trigger verification 영역, 변경 없음. |
| [EXTEND] | `internal/harness/applier.go` (16,650 bytes) | EXTEND | 기존 evolution applier. W3 provisional/applied/vetoed_by_canary status field 추가. |
| [EXTEND] | `internal/harness/learner.go` (5,537 bytes) | EXTEND | 기존 learner. W3 4-Tier engine + observations.yaml writer 추가. |
| [EXTEND] | `internal/harness/observer.go` (6,503 bytes) | EXTEND | 기존 observer. W3 SubagentStop hook 연동 추가. |
| [EXTEND] | `internal/hook/pre_tool.go` (20,548 bytes, since 2026-05-18) | EXTEND | 기존 PreToolUse hook 존재. W3 8 HARNESS_FROZEN_* sentinel catalog + harness-learner agent identity 차단 로직 추가. |
| [EXTEND] | `internal/hook/subagent_stop.go` (5,765 bytes) | EXTEND | 기존 SubagentStop hook. W3 harness-learner capture pipeline 연동 추가. |
| [NEW] | `internal/harness/tier/` package | NEW | 4-Tier state machine + observations.yaml schema 신설. |
| [NEW] | `internal/harness/capture/` package | NEW | Lesson capture heuristic 신설. |
| [NEW] | `internal/harness/throttle/` package | NEW | Proposal Throttling 4-mode 신설. |
| [NEW] | `internal/harness/seeds/` package | NEW | Cold-start seed loader 신설. |
| [NEW] | `internal/cli/harness_*.go` (6 verb files) | NEW | CLI verbs (status/apply/rollback/disable/mute/verify) 신설. |
| [PRESERVE] | `internal/harness/{interview, prefix_conflict, chaining_rules, evaluator_leak, profile_loader, retention, rubric, scorer, scorer_engine, types}.go` | PRESERVE | Out-of-scope to W3. 변경 없음. |

#### Consolidation Strategy Decision (B2 resolution)

플랜-감사 iter 1에서 제시된 3가지 옵션 중 **(b) 채택**:

- ~~(a) Move existing layer*.go into safety/ subdirectory~~: REJECTED. layer*.go는 TRIGGER VERIFIER (다른 concern). 이동 시 import path break + 기존 테스트 영향.
- **(b) Keep existing layer*.go in package root; extend safety/ subdirectory for new W3 concerns**: ACCEPTED. safety/ subdirectory가 이미 5-Layer를 담고 있으므로 W3는 safety/에 file 추가/확장. layer*.go는 무관 영역으로 보존.
- ~~(c) Hybrid via shim+deprecation~~: REJECTED. Deprecation은 W3 scope가 아니며 layer*.go는 deprecate 대상도 아님.

**Resulting package layout**: `internal/harness/safety/` 가 W3 5-Layer 확장의 home. plan.md §12 layout은 safety/ subdirectory 기준으로 작성됨.

### 1.6 D11 Seed Location Resolution (B3 resolution)

Vision §4.4 에 seed location이 두 곳으로 명시되어 있음을 plan-auditor iter 1에서 지적. iteration 2에서 다음과 같이 명시적 dual-path SSOT/cache 모델로 해소:

| 위치 | 역할 | Lifecycle | Ownership |
|------|------|-----------|-----------|
| `.claude/skills/moai-meta-harness/seeds/` | **Canonical SSOT (shipped)** | Core repo 일부, `moai update`로 갱신 | Core MoAI maintainer |
| `.moai/harness/seeds/` | **Project-local cache (runtime)** | `/moai project` (W4) 가 populate, 초기 W3 ship 시 empty | User project (per-project) |

**Precedence rule**: 같은 seed ID에 대해 project-local (`.moai/harness/seeds/`) > SSOT (`.claude/skills/moai-meta-harness/seeds/`).

**W3 scope**: SSOT lookup만 구현 (`/moai project` 미존재 가정). project-local cache는 W4 PROJECT-MEGA-001 deliverable (`/moai project --refresh` 시 SSOT → project-local 복사).

**W3 ship 상태**: 두 디렉토리 모두 `.gitkeep` placeholder. Vision §4.4 line 362-365의 path는 **target state W4**로 해석. REQ-HRA-024는 두 경로를 모두 인식하며 precedence rule을 따른다.

### 1.7 Field Naming Policy (S1 resolution)

본 SPEC + W3 산출물의 field naming은 다음 정책을 따른다:

| 영역 | 정책 | 예시 |
|------|------|------|
| **Timestamp fields in SPEC frontmatter** | Canonical per `spec-frontmatter-schema.md` SSOT — snake_case alias **금지** | `created:` / `updated:` (NOT `created_at:` / `updated_at:`). `tags:` (NOT `labels:`). Lint enforced (`FrontmatterInvalid`). |
| **Domain field names in YAML data files** | snake_case per Go YAML convention | `.moai/harness/observations.yaml`: `created:` (timestamp, SSOT 동일), `category:`, `pattern:`. evolution-log.md 구조 entry: `evolution_id`, `proposal_id`, `layer_results`, `affected_files`, `final_status`. |
| **Go struct tags** | snake_case via `yaml:"..."` tags | `type EvolutionEntry struct { EvolutionID string \`yaml:"evolution_id"\`; ProposalID string \`yaml:"proposal_id"\`; ... }` |
| **Sentinel error keys** | UPPER_SNAKE_CASE, language-agnostic | `HARNESS_LEARNING_CANARY_VETO`, `HARNESS_FROZEN_RULE_VIOLATION` |
| **CLI flag names** | kebab-case per Cobra convention | `--format`, `--strict`, `--batch-flush` |

**Distinction rationale**: Timestamp fields (`created`/`updated`) carry frontmatter-lint constraint (FrontmatterInvalid blocks merge). Domain YAML data fields are not subject to lint and follow community convention (snake_case for multi-word YAML keys). Both share the timestamp canonical for consistency where they overlap (observations.yaml `created:` field uses non-underscore form to match SPEC frontmatter schema).

---

## 2. 배경 (Background)

### 2.1 v3.5.0 Two-Zone Architecture (Vision §3.1)

```
Core MoAI (FROZEN)  ──→ moai update only, no autonomous evolution
        ↑
   Frozen Guard (W1 runtime hook + zone-registry SSOT)
        ↓
My Harness (EVOLVABLE) ──→ harness-learner auto-evolves via 4-Tier pipeline
```

W3은 My Harness layer의 **자율 진화 메커니즘**을 구현한다. Core MoAI 침범 시도는 W3 §3.3 PreToolUse hook 구현 + Vision §3.4 8 HARNESS_FROZEN_* sentinel catalog로 차단된다 (W1은 zone-registry data SSOT 제공만; PreToolUse hook 구현은 W1 EXCL-001에 따라 W3 scope).

### 2.2 4-Tier 진화 파이프라인 (Vision §6.3)

| Observations | Classification | Action | Storage |
|---|---|---|---|
| 1x | Observation | logged only | `.moai/harness/observations.yaml` |
| 3x | Heuristic | manager-develop hint로 참고 | observations.yaml status=`heuristic` |
| 5x | Rule | Sprint Contract 자동 추가 후보 | observations.yaml status=`rule` |
| 10x | High-confidence | AskUserQuestion auto-propose (throttling §6.6 적용) | observations.yaml status=`graduated` after approve |
| 1x (critical failure) | Anti-Pattern | 즉시 flag + FROZEN | `.moai/harness/anti-patterns.yaml` |
| Seed start | (configurable) | meta-harness seed = Tier 3 starting point (D4) | seeds loaded from `.claude/skills/moai-meta-harness/seeds/` |

### 2.3 5-Layer Safety (Vision §6.5) + Canary Veto Policy (E5)

Sequential execution (각 layer 통과 후 다음 layer):

| Layer | Type | Latency budget | Veto power |
|---|---|---|---|
| L1 Frozen Guard | sync (W3 implements PreToolUse hook code; reads W1 zone-registry as data) | < 10ms p99 | YES (catalog match → block) |
| L2 Canary | async (shadow eval on last 3 projects) | ~30s | **YES** (E5 — applies post-L5, may auto-rollback) |
| L3 Contradiction Detector | sync (emits blocker report to orchestrator, NOT direct AskUserQuestion — B5 resolution) | < 1s | YES (existing rule conflict) |
| L4 Rate Limiter | sync (only purely-internal layer) | < 100ms | YES (3/week + 24h cooldown + 50 active max) |
| L5 Human Oversight | sync (orchestrator AskUserQuestion via blocker report) | user-paced | YES (final approval) |

### 2.4 Canary Veto Policy (Vision §6.5 E5, verbatim)

> Layer 2 Canary는 asynchronous로 ~30s 소요되어 Layer 5 user approval보다 늦게 완료될 수 있다. 다음 정책으로 race 해소:
>
> 1. Layer 5 (user approval)이 Canary 완료 전 도착 → **provisional apply** (my-harness-* file write 수행, 단 evolution status = `provisional`)
> 2. Canary 완료:
>    - PASS → evolution status `provisional → applied`, evolution-log.md에 기록
>    - FAIL → **automatic rollback** (provisional file revert) + AskUserQuestion notification ("Canary가 regression 감지하여 자동 롤백됨. Override 또는 deeper review?")
> 3. **Canary는 Layer 5 approval에 대한 veto power 보유** — user approval이 final이 아님, Canary가 final gate
> 4. Veto 발생 시 해당 proposal은 48h cooldown 후 재제안 가능 (rate limiter에 별도 entry)

### 2.5 Cold-Start Mitigation (Vision §4.4)

Problem: 신규 프로젝트 observations.yaml = empty. Tier 4 도달까지 ~5-10 SPEC 동안 my-harness 깊이 부족 → quality regression risk.

Solution: `meta-harness`가 generation time에 curated baseline pattern을 seed (W3 정의 schema + load hook; W4가 seed library 실제 채움).

### 2.6 Proposal Throttling (Vision §6.6)

Tier 4 proposal fatigue 방지. 4 modes: `immediate` (default), `batch` (weekly review), `quiet` (quiet hours window), `mute` (per-category silence). `.moai/config/sections/workflow.yaml` `harness.proposal.*` 트리 신설.

### 2.7 W1 deliverables 의존성 (B1 corrected)

W1 ships DATA SSOT only (per W1 EXCL-001). W3 implements the runtime mechanism that consumes that data.

- `internal/constitution/validator.go` — W1 deliverable (DATA validator), W3 doesn't modify
- **`.claude/rules/moai/core/zone-registry.md` (111 entries)** — W3 L1 layer **reads this as input data**, NOT "consults W1's hook"
- **8 HARNESS_FROZEN_* sentinel catalog** — defined in Vision §3.4; **W3 is the FIRST implementer** of this catalog at runtime (no prior W1 hook exists per W1 EXCL-001)
- W3 extends sentinel namespace with **10 HARNESS_LEARNING_*** sentinels (plan.md §7)

---

## 3. EARS Requirements

### 3.1 LESSON-CAPTURE (REQ-HRA-001..003)

- **REQ-HRA-001**: The harness-learner subagent shall be invoked by the `SubagentStop` hook after every workflow agent (manager-develop, manager-quality, expert-*) completes, performing a background read-only scan for change patterns. (Ubiquitous)

- **REQ-HRA-002**: The lesson capture pipeline shall perform heuristic pattern matching (no LLM call) against changed files (diff input) and produce zero or more candidate observations per workflow event, completing within 500ms p95. (Ubiquitous)

- **REQ-HRA-003**: When a new observation is detected, the system shall write to `.moai/harness/observations.yaml` with canonical schema fields (`id`, `category`, `pattern`, `evidence[]`, `count`, `confidence`, `status`, `created`, `updated`) per Vision §6.2 (D3 resolution: `created:`/`updated:` field names, NOT `created_at`/`updated_at`). (Event-Driven)

### 3.2 TIER-ENGINE (REQ-HRA-004..007)

- **REQ-HRA-004**: When an observation's count reaches a tier threshold (1/3/5/10 per `harness.yaml` `learning.tier_thresholds`), the system shall transition its `status` field through the canonical sequence: `observation` → `heuristic` → `rule` → `high-confidence` and emit a tier-progression event to `.moai/harness/evolution-log.md`. (Event-Driven)

- **REQ-HRA-005**: When an observation reaches the `high-confidence` tier (count ≥ 10), the system shall trigger the 5-Layer Safety pipeline (L1 → L2 → L3 → L4 → L5) before any my-harness-* file write occurs. (Event-Driven)

- **REQ-HRA-006**: When an observation is detected as a critical-failure pattern (score drop > 0.20 or must-pass criterion failure), the system shall classify it as `anti-pattern` immediately (single occurrence) and write to `.moai/harness/anti-patterns.yaml` with FROZEN status (no further evolution allowed). (Event-Driven)

- **REQ-HRA-007**: While `harness.yaml` `learning.tier_thresholds` is set, the system shall NOT use any threshold other than the configured `[1, 3, 5, 10]` for the v3.5.0 cycle (Vision §9 Q4 resolved: fixed thresholds for v3.5.0). While `harness.yaml` `learning.tier_thresholds` differs from `[1, 3, 5, 10]`, the system shall emit `HARNESS_LEARNING_SCHEMA_DRIFT` at config load time and refuse to start the learning loop. The check executes in `LoadHarnessConfig()` (`internal/config`). (State-Driven)

### 3.3 LAYER-1-FROZEN (REQ-HRA-008..009) — B1 corrected

- **REQ-HRA-008**: While processing a Tier 4 proposal, the L1 Frozen Guard layer (implemented by W3 in `internal/hook/pre_tool.go` extension) shall **consume the W1 zone-registry (`.claude/rules/moai/core/zone-registry.md`, 111 entries) as input data** and verify that the proposed my-harness-* file write does not intersect any path matched by the 8 HARNESS_FROZEN_* sentinel catalog (Vision §3.4 catalog: `HARNESS_FROZEN_AGENT_VIOLATION`, `HARNESS_FROZEN_SKILL_VIOLATION`, `HARNESS_FROZEN_RULE_VIOLATION`, `HARNESS_FROZEN_COMMAND_VIOLATION`, `HARNESS_FROZEN_HOOK_VIOLATION`, `HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION`, `HARNESS_FROZEN_INSTRUCTION_VIOLATION`, `HARNESS_FROZEN_CONFIG_VIOLATION`). The L1 layer is **first-time implemented** in this SPEC; no prior hook exists per W1 EXCL-001 disclaimer. (State-Driven)

- **REQ-HRA-009**: When L1 Frozen Guard detects a path intersecting any of the 8 HARNESS_FROZEN_* catalog patterns, the system shall reject the proposal with W3 wrapper sentinel `HARNESS_LEARNING_FROZEN_BLOCKED` carrying the matched W1 catalog sentinel (e.g., `HARNESS_FROZEN_AGENT_VIOLATION`) as the `cause` field, and exit the safety pipeline (L2-L5 not invoked). (Event-Driven)

### 3.4 LAYER-2-CANARY (REQ-HRA-010..012)

- **REQ-HRA-010**: When L1 passes, the system shall asynchronously dispatch a shadow evaluation of the proposal against the last 3 project SPECs stored under `.moai/specs/` (most recent by `updated:` field), comparing pre/post quality scores within a 30-second budget. (Event-Driven)

- **REQ-HRA-011**: When L2 Canary detects a score drop greater than 0.10 on any of the 3 evaluated SPECs, the system shall mark the proposal `canary_status: FAIL` and exercise its veto power per Canary Veto Policy (§2.4) — automatic rollback of any provisional file write. (Event-Driven)

- **REQ-HRA-012**: When L5 user approval completes before L2 Canary returns, the system shall write the my-harness-* file with `evolution_status: provisional` AND emit a notification to the user that L2 verdict is pending. (Event-Driven)

### 3.5 LAYER-3-CONTRADICTION (REQ-HRA-013..014)

- **REQ-HRA-013**: When L2 passes (or is provisional pending), the system shall scan `.moai/harness/observations.yaml` (status=`rule` or `graduated`) and existing my-harness-* skill bodies for rule statements that semantically contradict the proposal, completing within 1 second. (Event-Driven)

- **REQ-HRA-014** (B5 corrected): When Layer 3 Contradiction Detector identifies a conflict between the proposed change and an existing rule (status `rule` or `graduated` in observations.yaml, or any rule statement embedded in my-harness-* skill bodies), the system shall **emit a structured L3-conflict blocker report** to the orchestrator containing (a) old rule text + location, (b) new rule text + source proposal, (c) suggested resolution options `[resolve-by-replacing, resolve-by-amending, reject]`. **The harness-learner subagent MUST NOT invoke AskUserQuestion directly** per agent-common-protocol §User Interaction Boundary HARD. The orchestrator translates the blocker report into an AskUserQuestion round. (Event-Driven)

### 3.6 LAYER-4-RATELIMIT (REQ-HRA-015..017)

- **REQ-HRA-015**: While the L4 Rate Limiter counts the number of evolutions applied in the trailing 7-day window, it shall reject any new proposal that would exceed `harness.yaml` `learning.rate_limit.max_per_week` (default 3). (State-Driven)

- **REQ-HRA-016**: While the L4 Rate Limiter checks the cooldown timer, it shall reject any new proposal within `harness.yaml` `learning.rate_limit.cooldown_hours` (default 24) of the previous applied evolution, regardless of category. (State-Driven)

- **REQ-HRA-017**: While the active learning count (entries with status in `[observation, heuristic, rule, high-confidence]`) exceeds 50 in `observations.yaml`, the system shall archive the oldest entries (by `created:` field, status=`observation`) until the active count drops to ≤ 50. (State-Driven)

### 3.7 LAYER-5-OVERSIGHT (REQ-HRA-018..019)

- **REQ-HRA-018**: When L1-L4 all pass (or L3 user-resolved), the system shall return a structured blocker report to the orchestrator (subagents cannot invoke AskUserQuestion per agent-common-protocol §User Interaction Boundary) containing the proposal payload, canary verdict (or pending status), and the 4-option choice (`Apply / Apply with modification / Defer / Reject permanently` per Vision §6.4). (Event-Driven)

- **REQ-HRA-019**: When the user selects `Reject permanently` via the orchestrator's AskUserQuestion, the system shall mark the observation `status: anti-pattern` (FROZEN, no further re-proposal allowed per §6.6 of Vision). (Event-Driven)

### 3.8 CANARY-VETO (REQ-HRA-020..021)

- **REQ-HRA-020**: When L5 user approval has been granted and L2 Canary subsequently returns FAIL, the system shall automatically rollback the provisional file write (revert the my-harness-* file to its pre-evolution state) AND emit a structured blocker report to the orchestrator for user notification per §2.4 step 2(b). (Event-Driven)

- **REQ-HRA-021**: When a Canary veto has occurred, the system shall record the evolution in `.moai/harness/evolution-log.md` with `result: vetoed_by_canary` and apply a 48-hour cooldown before the same proposal can be re-submitted (Vision §6.5 step 4). (Event-Driven)

### 3.9 COLD-START-SEEDS (REQ-HRA-022..024)

- **REQ-HRA-022**: The cold-start seed schema shall define seed YAML files with fields (`id`, `pattern`, `tier`, `confidence`, `category`, `body`, `references`) per Vision §4.4 Seed Structure. (Ubiquitous)

- **REQ-HRA-023**: When the harness-learner loads a seed file on first invocation in a project (observations.yaml empty or absent), the system shall inject seed entries into `observations.yaml` with `status: rule` (Tier 3 starting point per Vision §4.4 step 3). (Event-Driven)

- **REQ-HRA-024**: The seed library location (`.claude/skills/moai-meta-harness/seeds/`) shall be defined in `harness.yaml` under a new `learning.seeds.library_path` field. The actual seed file content is OUT OF SCOPE for W3 (Vision §5 W4 — seed library 8 baseline files). (Ubiquitous)

### 3.10 THROTTLING (REQ-HRA-025..028)

- **REQ-HRA-025**: While `workflow.yaml` `harness.proposal.mode` is set to `immediate` (default), the system shall trigger AskUserQuestion at the moment of Tier 4 attainment after passing L1-L4. (State-Driven)

- **REQ-HRA-026**: While `workflow.yaml` `harness.proposal.mode` is set to `batch`, the system shall accumulate Tier 4 proposals into a pending queue and emit a single multi-proposal blocker report at the configured window boundary (`weekly` or `sprint_end`), capped at `max_per_window` (default 5). (State-Driven)

- **REQ-HRA-027**: While `workflow.yaml` `harness.proposal.mode` is set to `quiet` AND the current local time falls within the configured `quiet.hours` window (default `[18, 9]` Asia/Seoul), the system shall defer Tier 4 proposals until the window closes. (State-Driven)

- **REQ-HRA-028**: While a category appears in `workflow.yaml` `harness.proposal.mute.categories[]`, the system shall NOT trigger AskUserQuestion for Tier 4 proposals in that category (proposal logged to evolution-log.md with `status: muted`). (State-Driven)

### 3.11 CLI (REQ-HRA-029..033)

- **REQ-HRA-029**: The `moai harness status` CLI shall output the current observation count, tier distribution (counts per status), recent evolutions (last 5), and active learning total, completing within 2 seconds. (Ubiquitous)

- **REQ-HRA-030**: The `moai harness apply <proposal-id>` CLI shall trigger the 5-Layer Safety pipeline for a queued Tier 4 proposal (used in `batch` mode for manual processing). (Event-Driven)

- **REQ-HRA-031**: The `moai harness rollback <evolution-id>` CLI shall revert a previously applied evolution by restoring the pre-evolution my-harness-* file content from `.moai/harness/evolution-log.md` provenance records. (Event-Driven)

- **REQ-HRA-032**: The `moai harness disable` CLI shall set `harness.yaml` `learning.enabled: false` (with confirmation prompt via orchestrator AskUserQuestion), pausing all Tier 4 auto-proposals until re-enabled. (Event-Driven)

- **REQ-HRA-033**: The `moai harness mute <category>` and `moai harness mute-list` and `moai harness unmute <category>` CLI verbs shall manage the `workflow.yaml` `harness.proposal.mute.categories[]` list. (Event-Driven)

### 3.12 EVOLUTION-LOG (REQ-HRA-034..035)

- **REQ-HRA-034**: The `.moai/harness/evolution-log.md` file shall append a structured entry for every evolution attempt (regardless of outcome) containing fields `evolution_id`, `timestamp`, `proposal_id`, `layer_results` (L1/L2/L3/L4/L5 per-layer verdict), `final_status` (applied | provisional | rejected | vetoed_by_canary | rolled_back), `affected_files`. (Ubiquitous)

- **REQ-HRA-035**: The system shall NOT modify or delete existing evolution-log.md entries; it is append-only. The `moai harness rollback` command writes a NEW reverse-evolution entry, leaving the original record intact. (Unwanted)

### 3.13 VERIFY (REQ-HRA-036)

- **REQ-HRA-036**: The `moai harness verify --determinism` CLI shall (placeholder for W4 — surface acknowledgment that determinism is W4 scope; W3 implements CLI verb skeleton only). The verb shall print "Determinism verification deferred to W4 (PROJECT-MEGA-001)" and exit 0. (Optional)

### 3.14 PERFORMANCE-NFR (REQ-HRA-037) — S6 resolution

- **REQ-HRA-037**: The L1 Frozen Guard layer shall return its allow/deny verdict within **10ms p99** per file-write attempt, measured via `BenchmarkL1FrozenGuard` over 10,000 randomized path inputs against zone-registry of 111 entries (W1 deliverable size). Failure threshold: any measured p99 ≥ 15ms is BLOCKING and fails CI. (Ubiquitous, Performance NFR)

### 3.15 BROWNFIELD-CONSOLIDATION (REQ-HRA-038) — B2 resolution

- **REQ-HRA-038**: Existing `internal/harness/layer{1,2,3,5}.go` files (package `harness`, trigger verification concern per §1.5) shall be **preserved unchanged**. W3 5-Layer Safety extensions shall be applied **only** to `internal/harness/safety/` subdirectory (package `safety`), per consolidation strategy (b) in §1.5. Characterization test for the preserved files: run `go test ./internal/harness/` (root level test files `layer{1,2,3,5}_test.go`) before and after W3 changes — pass rate must be identical (DDD analyze-preserve-improve cycle). (Ubiquitous)

---

## 4. Out of Scope (Exclusions — What NOT to Build)

W3 explicitly excludes the following items. Each EXCL-* maps to a deferred SPEC or W phase.

### 4.1 Out of Scope — Exclusion List

- **EXCL-HRA-001 (Determinism guarantee)**: Deterministic harness generation per Vision §3.5 is W4 PROJECT-MEGA-001 scope. W3 implements only the CLI verb skeleton (REQ-HRA-036 placeholder) with deferred message.

- **EXCL-HRA-002 (Seed library content)**: The actual 8 baseline seed files (Go/Node/Python/Rust/React/Vue/Flutter/iOS per Vision §5 W4) are W4 scope. W3 defines schema (REQ-HRA-022) + load hook (REQ-HRA-023) + library path config field (REQ-HRA-024) only.

- **EXCL-HRA-003 (`/moai project --refresh` command)**: The refresh command surface is W4 scope. W3 does NOT add `--refresh` flag handling to `/moai project`.

- **EXCL-HRA-004 (meta-harness 7-Phase workflow)**: The full meta-harness orchestration (Phase 1-7) is W4 scope. W3 only invokes the existing meta-harness skill body for seed loading.

- **EXCL-HRA-005 (Project-specific my-harness generation)**: The actual generation of project-detected my-harness-* skills/agents from Socratic interview is W4 scope. W3 assumes my-harness-* files exist (created externally) and evolves them.

- **EXCL-HRA-006 (Migration tooling for existing projects)**: An upgrade path for existing v3.4.x projects to enable W3 harness autonomy is deferred to a separate follow-up SPEC (gated on adoption signal).

- **EXCL-HRA-007 (Backward migration tool)**: A rollback from v3.5.0 harness autonomy to pre-v3.5.0 architecture is an explicit non-goal. Users who want to disable autonomy use `moai harness disable` (REQ-HRA-032).

- **EXCL-HRA-008 (LLM-based pattern matching)**: Lesson capture (REQ-HRA-002) uses heuristic matching only. LLM-based semantic capture is deferred (latency budget + cost considerations).

- **EXCL-HRA-009 (Cross-project harness sharing)**: Community harness library — sharing my-harness-* skills between projects — is deferred per Vision §11.2 Out of Scope.

- **EXCL-HRA-010 (Real-time canary evaluation streaming)**: L2 Canary is async with ~30s budget. Real-time streaming canary results (live progress UI) is out of scope.

---

## 5. Acceptance Criteria summary

Detailed Given/When/Then scenarios are in `acceptance.md`. Summary mapping (12 binary ACs + 6 edge cases + 5 risk mitigations):

| AC | Description | Verification Method |
|----|-------------|---------------------|
| AC-HRA-001 | Lesson auto-capture via SubagentStop hook trigger | Hook fixture + observations.yaml file emit check |
| AC-HRA-002 | Tier 1 → Tier 4 end-to-end progression | Synthetic 10x observation injection → status transitions |
| AC-HRA-003 | 5-Layer safety unit tests (each layer PASS/FAIL) | `safety_pipeline_test.go` table-driven |
| AC-HRA-004 (B1 corrected) | Frozen Guard violation block (8 sentinel simulation) using zone-registry as data | **Precondition: zone-registry exists at `.claude/rules/moai/core/zone-registry.md` with 111 entries (W1 deliverable, main HEAD `7bd23bb69`).** W3 L1 hook implementation reads registry → 8 path-pattern fixtures → reject + correct W1 catalog sentinel + W3 wrapper sentinel |
| AC-HRA-005 | User rejection → permanent block (anti-pattern) | Orchestrator AskUserQuestion mock + status=anti-pattern verification |
| AC-HRA-006 | Proposal Throttling 4-mode (immediate/batch/quiet/mute) | Mode-by-mode timer/state fixture |
| AC-HRA-007 | Cold-start regression: seed inject → first SPEC quality | Synthetic empty project + 1 seed load → harness body content; SSOT path tested, project-local cache deferred to W4 |
| AC-HRA-008 | Canary Veto Policy (E5) provisional apply + auto-rollback | L5 approve + L2 FAIL fixture → file revert + log entry |
| **AC-HRA-008b** (B4 added) | After Canary-veto auto-rollback, re-application within 48h cooldown REJECTED | `moai harness apply <evolution_id>` within cooldown → exit code 1 + `HARNESS_LEARNING_RATELIMIT_EXCEEDED` sentinel emitted |
| AC-HRA-009 | 6 CLI verb surface (status/apply/rollback/disable/mute/verify) | Cobra command registration + help output |
| AC-HRA-010 | evolution-log.md append-only with structured entries | Multiple evolution events → 5+ entries, no deletion |
| AC-HRA-011 | Anti-pattern auto-flag on critical failure (REQ-HRA-006) | Critical failure fixture → anti-patterns.yaml entry |
| AC-HRA-012 | observations.yaml schema canonical field names | YAML decode test against canonical `created:`/`updated:` |
| **AC-HRA-013** (S3 added) | REQ-HRA-007 enforcement — non-canonical tier_thresholds rejected at config load | Set `harness.yaml` `tier_thresholds: [1, 3, 5, 11]` → invoke `moai harness status` → exit code 1 + `HARNESS_LEARNING_SCHEMA_DRIFT` in stderr |
| **AC-HRA-014** (S6 added) | REQ-HRA-037 L1 latency NFR enforcement | `go test -bench=BenchmarkL1FrozenGuard ./internal/harness/safety/` → p99 latency ≤ 10ms over 10,000 inputs |
| EC-HRA-001 | SubagentStop hook fires on every subagent including read-only | Read-only agent fixture → hook invoked |
| EC-HRA-002 | Concurrent lesson capture race (parallel subagents) | 3 parallel SubagentStop events → all captured |
| EC-HRA-003 | Rate limiter cross-week boundary (Sunday→Monday) | 3 events spanning week boundary → 4th rejected |
| EC-HRA-004 | Quiet hours timezone DST transition | Asia/Seoul (no DST) fixture |
| EC-HRA-005 | Active learning count exceeds 50 (archival) | 51 entries fixture → oldest observation archived |
| EC-HRA-006 | Canary veto AFTER 48h cooldown (re-submission path) | 48h+1s elapsed → re-proposal accepted |
| R-HRA-001 | R2: Over-aggressive evolution — L1+L5 sufficient gate | L1 + L5 unit test verifies block path |
| R-HRA-002 | R4: Lesson capture latency budget (<500ms p95) | Benchmark test on synthetic 10MB diff |
| R-HRA-003 | R7: W3 mechanism complexity — incremental layer activation | Each layer independently testable |
| R-HRA-004 | R8: Cold-start regression mitigated by seed | Empty observations + seed → Tier 3 entry present |
| R-HRA-005 | R9: Tier 4 fatigue — batch + mute reduce AskUserQuestion calls | Mode comparison: immediate vs batch event count |

---

## 6. References

**Primary vision (binding source)**:
- `.moai/research/harness-autonomy-vision-2026-05-18.md` (iter3) — §3.4 Frozen Guard, §3.5 Determinism (W4 scope), §4.4 Cold-Start Seeds, §5 W3 scope, §6.1-6.7 mechanism deep dive, §8 risk assessment, §9 open questions

**W1 deliverables (B1 corrected — DATA SSOT only, NO hook)**:
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/` — W1 SPEC (completed) — see W1 spec.md §5.2 EXCL-001 for explicit PreToolUse hook disclaimer (deferred to W3)
- `.claude/rules/moai/core/zone-registry.md` — 111 entries (W1 product, **W3 reads as input data**, NOT "consults W1's hook")
- `internal/constitution/validator.go` — W1 DATA validator, W3 doesn't modify
- 8 HARNESS_FROZEN_* sentinel catalog (Vision §3.4) — **W3 is the first runtime implementer**; the catalog itself is defined in Vision, not W1
- W3 introduces additive HARNESS_LEARNING_* sentinels (plan.md §7)

**W2 deliverables (parallel, non-blocking)**:
- `.moai/specs/SPEC-V3R5-CORE-SLIM-001/` — W2 SPEC (completed) — moai-foundation-quality preload + 4 expert agents

**Existing infrastructure (W3 modifies/extends)**:
- `.moai/config/sections/harness.yaml` — `learning.tier_thresholds`, `rate_limit`, `auto_apply: false` (existing; W3 extends with `proposal.*`, `seeds.*`)
- `.claude/skills/moai-harness-learner/SKILL.md` — 6,454 bytes (W3 extends in run-phase; this SPEC documents proposed extension only)
- `.claude/skills/moai-meta-harness/SKILL.md` — 16,825 bytes (W3 references for seed load hook)
- `internal/hook/` — existing hook handlers (W3 adds harness-learner SubagentStop integration)

**Frontmatter/schema SSOT**:
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema (this SPEC complies)

**W1 SPEC reference patterns**:
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/spec.md` — frontmatter pattern, HISTORY narrative, EARS hierarchical structure
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/plan.md` — phase decomposition pattern (4 phases sequential)
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/acceptance.md` — binary AC pattern (AC-* + EC-* + R-*)

---

## 7. Open Questions (Vision §9 Status)

Vision §9 lists 9 open questions. Status for W3:

| Q | Topic | Status in W3 | Resolution |
|---|-------|--------------|------------|
| Q1 | PreToolUse hook performance (10ms/call cumulative latency) | RESOLVED in plan.md §10 | NFR: < 10ms p99 per call, total latency budget enforced via timing test (AC-HRA-003) |
| Q2 | Seed library maintenance ownership | RESOLVED in plan.md §10 | Core repo ships seeds (W4 W3 only defines schema). Community contribution mechanism deferred. |
| Q3 | 5-Layer Safety sequencing — Canary async vs L3+ entry | RESOLVED earlier in Vision §6.5 + E5 Canary Veto Policy | Provisional apply pattern (REQ-HRA-012) |
| Q4 | Tier threshold adequacy (1/3/5/10 for small/large projects) | RESOLVED in plan.md §10 | v3.5.0 uses fixed 1/3/5/10 (REQ-HRA-007). Project-size adaptive threshold deferred. |
| Q5 | AskUserQuestion fatigue — batch mode 5-proposal multi-round split | RESOLVED in plan.md §10 | Multi-round automatic split (Claude Code max 4 questions/round, batch can carry up to 5 — split into rounds of 4+1) |
| Q6 | W2 Stub Agent Lifetime (v3.6.0 vs v4.0.0 retire) | N/A for W3 (W2 concern) | Deferred to W2 sync narrative |
| Q7 | W3 Lesson Capture Trigger Breadth — SubagentStop only or also manual CLI | RESOLVED in plan.md §10 | SubagentStop only for v3.5.0; manual `moai harness capture <pattern>` deferred |
| Q8 | Migration cost — 5 SPEC × 6 PR = 30 PR incremental cadence | N/A for W3 (sprint planning concern) | Deferred to v3.5.0 release narrative |
| Q9 | Vision-Implementation Drift Detection — vision evolves during W execution | RESOLVED earlier | plan-auditor re-runs at each W; zone-registry 100% coverage from W1 acts as drift signal |

---

## 8. REQ ↔ AC Traceability Matrix

[HARD] All REQ-HRA-* must map to ≥1 AC-HRA-* (RQ-5 traceability rule).

| REQ ID | Primary AC | EC/R secondary | Verification Method |
|--------|------------|----------------|---------------------|
| REQ-HRA-001 (SubagentStop trigger) | AC-HRA-001 | EC-HRA-001 | Hook fixture (read-only and write-mode agents) |
| REQ-HRA-002 (heuristic match <500ms) | AC-HRA-001 | R-HRA-002 | Benchmark on 10MB diff |
| REQ-HRA-003 (observations.yaml schema) | AC-HRA-012 | — | YAML decode against canonical fields |
| REQ-HRA-004 (Tier transitions) | AC-HRA-002 | — | 10x observation synthetic injection |
| REQ-HRA-005 (Tier 4 → 5-Layer trigger) | AC-HRA-003 | — | Pipeline entry assertion |
| REQ-HRA-006 (anti-pattern auto-flag) | AC-HRA-011 | — | Critical failure fixture |
| REQ-HRA-007 (fixed thresholds + SCHEMA_DRIFT enforcement) | AC-HRA-002 + AC-HRA-013 | — | Config override test + non-canonical threshold rejection |
| REQ-HRA-008 (L1 reads zone-registry; B1 corrected) | AC-HRA-004 | R-HRA-001 | 8 sentinel simulation with zone-registry as data input |
| REQ-HRA-009 (L1 violation reject + W3 wrapper sentinel) | AC-HRA-004 | — | Wrapper sentinel `cause` field assertion |
| REQ-HRA-010 (L2 Canary async dispatch) | AC-HRA-003 | — | Async eval timing test |
| REQ-HRA-011 (L2 score-drop veto) | AC-HRA-008 | — | Score fixture + veto path |
| REQ-HRA-012 (L5-before-L2 provisional) | AC-HRA-008 | — | Race condition fixture |
| REQ-HRA-013 (L3 scan ≤1s) | AC-HRA-003 | — | Latency benchmark |
| REQ-HRA-014 (L3 emits blocker report; B5 corrected — no direct AskUserQuestion) | AC-HRA-003 + C-HRA-008 | — | Conflict fixture + blocker report emission + static grep negative test |
| REQ-HRA-015 (L4 weekly limit) | AC-HRA-003 | EC-HRA-003 | Week boundary fixture |
| REQ-HRA-016 (L4 cooldown 24h) | AC-HRA-003 | — | Timer fixture |
| REQ-HRA-017 (L4 active count ≤50) | AC-HRA-003 | EC-HRA-005 | 51-entry archival fixture |
| REQ-HRA-018 (L5 blocker report) | AC-HRA-003 | — | Report structure validation |
| REQ-HRA-019 (L5 reject permanent → anti-pattern) | AC-HRA-005 | — | AskUserQuestion mock + status check |
| REQ-HRA-020 (Canary veto auto-rollback) | AC-HRA-008 | — | File revert + log entry assertion |
| REQ-HRA-021 (48h cooldown after veto) | AC-HRA-008 + AC-HRA-008b | EC-HRA-006 | Time-elapsed fixture + cooldown rejection assertion |
| REQ-HRA-022 (seed schema definition) | AC-HRA-007 | — | YAML decode against seed schema |
| REQ-HRA-023 (seed load on empty observations) | AC-HRA-007 | R-HRA-004 | Empty observations fixture |
| REQ-HRA-024 (seeds.library_path config) | AC-HRA-007 | — | harness.yaml field presence |
| REQ-HRA-025 (mode=immediate) | AC-HRA-006 | — | Default mode timing |
| REQ-HRA-026 (mode=batch + max_per_window) | AC-HRA-006 | R-HRA-005 | Batch window fixture |
| REQ-HRA-027 (mode=quiet + timezone) | AC-HRA-006 | EC-HRA-004 | Quiet hours fixture |
| REQ-HRA-028 (mode=mute per-category) | AC-HRA-006 | — | Category mute fixture |
| REQ-HRA-029 (CLI status) | AC-HRA-009 | — | Cobra help + output structure |
| REQ-HRA-030 (CLI apply) | AC-HRA-009 | — | Queued proposal apply test |
| REQ-HRA-031 (CLI rollback) | AC-HRA-009 | — | Reverse-evolution entry |
| REQ-HRA-032 (CLI disable) | AC-HRA-009 | — | Config field mutation |
| REQ-HRA-033 (CLI mute/unmute/list) | AC-HRA-009 | — | Workflow.yaml mutation |
| REQ-HRA-034 (evolution-log structured entries) | AC-HRA-010 | — | Entry field validation |
| REQ-HRA-035 (evolution-log append-only) | AC-HRA-010 | — | No-modify assertion |
| REQ-HRA-036 (verify --determinism placeholder) | AC-HRA-009 | — | Exit 0 + deferred message |
| REQ-HRA-037 (L1 latency NFR) | AC-HRA-014 | R-HRA-002 | BenchmarkL1FrozenGuard p99 ≤ 10ms |
| REQ-HRA-038 (brownfield preserve layer*.go) | AC-HRA-003 + R-HRA-003 | — | `go test ./internal/harness/` characterization (pre/post identical) |

Coverage: 38 REQs ↔ 14 ACs (12 + AC-HRA-008b + AC-HRA-013 + AC-HRA-014) + 6 EC + 5 R + 1 C (C-HRA-008) = 26 total verification surfaces. Every REQ has ≥1 AC. EC/R provide secondary coverage for edge cases and risks.

---

## 9. Scope Boundaries

In Scope (T2 Standard):
- 4-Tier evolution pipeline (REQ-HRA-001..007)
- 5-Layer Safety with Canary Veto (REQ-HRA-008..021)
- Cold-Start Seed schema + load hook (REQ-HRA-022..024)
- Proposal Throttling 4 modes (REQ-HRA-025..028)
- 6 CLI verbs (REQ-HRA-029..033, REQ-HRA-036)
- evolution-log.md append-only (REQ-HRA-034..035)

Out of Scope (deferred):
- W4 PROJECT-MEGA scope items: Determinism, seed library content, `/moai project --refresh`, meta-harness 7-Phase, project-specific my-harness generation (EXCL-HRA-001..005)
- Backward migration tool (EXCL-HRA-007, explicit non-goal)
- LLM-based capture, cross-project sharing, real-time canary streaming (EXCL-HRA-008..010)

---

## 10. Constraints (forwarded from plan.md §10 for completeness)

- **C-HRA-001 (Frontmatter SSOT)**: This SPEC complies with 12-field canonical schema. `created:`/`updated:`/`tags:` (NOT `created_at:`/`updated_at:`/`labels:`).
- **C-HRA-002 (16-language neutrality)**: Sentinel error keys language-agnostic (`HARNESS_LEARNING_*`).
- **C-HRA-003 (Performance)**: Lesson capture <500ms p95 (REQ-HRA-002). L1 < 10ms p99. L3 < 1s. L4 < 100ms.
- **C-HRA-004 (No external dependencies)**: Go stdlib + existing `gopkg.in/yaml.v3` only.
- **C-HRA-005 (TRUST 5)**: `internal/harness/` coverage ≥ 85%. Read-only data flows where possible (file writes only in evolution apply path).
- **C-HRA-006 (FROZEN constraints)**: `harness.yaml` `tier_thresholds: [1, 3, 5, 10]` MUST be preserved as-is. Extend only with new sub-trees (`proposal`, `seeds`).
- **C-HRA-007 (Sentinel catalog provenance)**: The 8 HARNESS_FROZEN_* sentinels are defined in Vision §3.4, NOT in W1. W3 is the first runtime implementer; subsequent SPECs treating these sentinels as published API must respect W3 as the source of truth. W3 additionally introduces 10 HARNESS_LEARNING_* sentinels (plan.md §7).
- **C-HRA-008 (Subagent boundary, S5 binary AC)**: harness-learner subagent NEVER invokes AskUserQuestion directly. Returns blocker report to orchestrator (per agent-common-protocol §User Interaction Boundary HARD). **Binary verification**: static analysis via `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/` MUST return zero matches. CI guard: `internal/harness/subagent_boundary_test.go` runs the grep assertion as a Go test. Applies to L3 (REQ-HRA-014) AND L5 (REQ-HRA-018) — both layers emit blocker reports; only L4 is purely-internal (no user interaction).
