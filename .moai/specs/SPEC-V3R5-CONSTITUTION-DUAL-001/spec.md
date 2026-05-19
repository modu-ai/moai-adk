---
id: SPEC-V3R5-CONSTITUTION-DUAL-001
title: "Constitution Dual-Zone Formalization with Validate CLI"
version: "0.3.0"
status: completed
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: ".claude/rules/moai + internal/constitution + internal/cli"
lifecycle: spec-anchored
tags: "constitution, dual-zone, frozen, evolvable, zone-registry, mega-sprint, w1"
issue_number: 1014
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial draft — Mega-Sprint W1 — T2 Standard scope per AskUserQuestion |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision — addressed plan-auditor BLOCKING defects (zone-registry 75→72, HARD rules 102→111 empirical, AC methodology unified) + SHOULD defects (5 new ACs for traceability, V3R5-001 namespace, plan.md/acceptance.md Out of Scope sections, 3 sentinel REQs added, AC-CDL-005 split). |
| 0.2.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Status `draft → implemented`. D1 (commit b8e563020, ZONE markers 100% × 15 files × 111 [HARD]) + D2 (commit 4fb9c9ce8, zone-registry 72→111 entries + zone_class 4-enum) + D3 (commit 7b5f643fe, validate CLI verb + 9 sentinel keys + 13 tests, AC-CDL-003/004/006/008/009/010 + EC-CDL-002/005/007 PASS) 모두 main 머지 완료 (PR #1015 plan + #1016 run, admin squash override per chicken-and-egg W2 lint baseline). AC-CDL-005a (CI step) / AC-CDL-005b (branch protection 4→5) 는 plan R-CDL-04 mitigation 에 따라 follow-up SPEC 로 이관. Baseline drift 69 entries 는 validator 의 designed output (drift 정확 탐지) — SPEC-V3R5-CONSTITUTION-DRIFT-CLEAN-001 (가칭) 후속. |
| 0.3.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Status `implemented → completed`. Sync phase: spec lint clean (StatusGitConsistency 경고 해소), 13 ACs binary-PASS 유지, no regressions. W1 lifecycle 완료 — Mega-Sprint 다음 단계 (W2 CORE-SLIM-001 / W3 HARNESS-AUTONOMY-001 / W4 PROJECT-MEGA-001) 진입 가능 상태. |

---

## 1. 개요 (Overview)

Mega-Sprint W1의 헌법 이원화 작업. v3.5.0 Two-Zone Architecture (`.moai/research/harness-autonomy-vision-2026-05-18.md` §3.1) 의 헌법적 토대를 마련한다.

현재 상태 (empirical ground truth, verified against main HEAD `3bd2aa291`):

- **15개 헌법 source files** (.claude/rules/moai/{core,workflow,development} + CLAUDE.md + design/constitution.md) 에 분산된 **111개 empirical [HARD] 규칙** 이 존재 (단, `design/constitution.md`는 §2에서 부분적으로 FROZEN/EVOLVABLE을 정의함)
- `zone-registry.md`는 **72 entries** 보유 (CONST-V3R2-001..046 + 049 + 051..072 + 150..152; 3 internal gaps at 047/048/050 — 역사적 미할당 또는 제거됨)
- 따라서 현재 coverage 는 **72/111 = 65%** 이며, **39개 [HARD] 규칙이 unmapped** (F-006 후속 ground-truth 갱신)
- `moai constitution list/guard/amend` CLI 3 verbs 는 존재하나 **`validate` (drift 탐지)** verb 부재 → 헌법-소스 정합성 자동 검증 불가능

본 SPEC의 목표:

1. **D1**: 15개 헌법 source files 에 `[ZONE:Frozen]` / `[ZONE:Evolvable]` 마커를 100% 부착 (안전성-임계 vs 운영-튜닝 구분)
2. **D2**: `zone-registry.md`를 100% coverage (≥ 111 entries) 로 확장하고 `zone_class` 4-classification 추가
3. **D3**: `moai constitution validate` CLI verb 신설 — clause-anchor drift 탐지 + CI 통합 (기존 `guard` verb 와 직교: `guard` = pre-amendment safety check, `validate` = post-acceptance drift detection)

본 SPEC은 W2 CORE-SLIM-001 / W3 HARNESS-AUTONOMY-001의 헌법적 전제조건이다.

## 2. 배경 (Background)

### 2.1 v3.5.0 Two-Zone Architecture 비전

`harness-autonomy-vision-2026-05-18.md` (D1-D10 resolved v2)는 MoAI-ADK를 두 영역으로 분리한다:

- **Core MoAI (FROZEN)**: 17 canonical agents + 헌법 규칙. 학습자(learner)가 수정 불가
- **My Harness (EVOLVABLE)**: 프로젝트별 3-7개 customizable agents + 운영 튜닝 규칙. 점진적 학습 허용

이 분리는 헌법 수준의 마킹과 zone-registry의 완전한 매핑 없이는 자동 강제 불가능하다. W1은 이 토대를 제공한다.

### 2.2 F-006: zone-registry coverage gap (ground-truth 갱신)

`architecture-audit-2026-05-18.md` finding F-006 의 초기 추정 (102개 HARD, 44% coverage) 은 본 SPEC iteration 2 에서 empirical re-audit 완료:

| 항목 | F-006 초기 추정 | Iteration 2 empirical (main HEAD `3bd2aa291`) |
|------|----------------|-----------------------------------------------|
| Total [HARD] count | 102 (추정) | **111** (15 source files grep sum) |
| zone-registry entries | (미측정) | **72** (3 gaps: 047/048/050) |
| Coverage | 44% (추정) | **65% (72/111)** |
| Unmapped HARD rules | 57 (추정) | **39** |

15 source files breakdown:

| File | HARD count |
|------|-----------|
| `CLAUDE.md` | 14 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 11 |
| `.claude/rules/moai/core/askuser-protocol.md` | 3 |
| `.claude/rules/moai/core/moai-constitution.md` | 11 |
| `.claude/rules/moai/design/constitution.md` | 19 |
| `.claude/rules/moai/workflow/ci-autofix-protocol.md` | 10 |
| `.claude/rules/moai/workflow/ci-watch-protocol.md` | 8 |
| `.claude/rules/moai/workflow/context-window-management.md` | 5 |
| `.claude/rules/moai/workflow/session-handoff.md` | 5 |
| `.claude/rules/moai/workflow/spec-workflow.md` | 3 |
| `.claude/rules/moai/workflow/worktree-integration.md` | 12 |
| `.claude/rules/moai/workflow/worktree-state-guard.md` | 1 |
| `.claude/rules/moai/development/agent-authoring.md` | 1 |
| `.claude/rules/moai/development/branch-origin-protocol.md` | 7 |
| `.claude/rules/moai/development/skill-authoring.md` | 1 |
| **TOTAL** | **111** |

결과 (unchanged from F-006 initial framing):

- 향후 Frozen Guard hook (W3 scope) 이 unmapped HARD 규칙을 차단하지 못함
- 헌법 amendment 시 drift 탐지 불가 (clause가 변경되어도 registry는 stale)
- 학습자(learner)가 어떤 규칙을 EVOLVABLE로 가정해도 되는지 명시적 SSOT 부재

### 2.3 W2 CORE-SLIM-001의 정당성

`.moai/state/lint-w2-deferred.json` 12 residual은 expert-backend/expert-frontend agent retirement (W2 scope)을 통해 해소된다. 그러나 retirement의 헌법적 근거 (어떤 agent가 FROZEN canonical 17에 속하는가)는 W1의 zone classification에 의존한다.

### 2.4 design/constitution.md 선행 패턴

`.claude/rules/moai/design/constitution.md` v3.4.0 §2는 이미 FROZEN/EVOLVABLE zones를 명시한다:

- FROZEN: constitution file 자체, Section 3.1~3.3, Safety architecture, GAN Loop contract, Pass threshold floor 등
- EVOLVABLE: skill body content, pipeline adaptation weights, evaluation rubric criteria 등

본 SPEC은 이 패턴을 core constitution (CLAUDE.md, moai-constitution.md, agent-common-protocol.md, workflow rules) 으로 일반화한다.

### 2.5 기존 `moai constitution` CLI surface 와의 관계

`internal/cli/constitution.go` 는 이미 3 verbs 를 노출한다:

- `constitution` (parent command)
- `guard` — pre-amendment FROZEN zone violation check (CI integration in `moai constitution guard`)
- `list` — registry entries query (`--zone`, `--file`, `--format` filters)
- `amend` — constitutional amendment with 5-layer safety gate

본 SPEC 의 신규 `validate` verb 는 `guard` 와 **직교** 관계이다:

| Verb | 시점 | 대상 | 목적 |
|------|------|------|------|
| `guard` | Amendment **시도 시** (전) | 제안된 amendment | FROZEN zone 위반 차단 (proposed change vs constitution) |
| `validate` (NEW) | Amendment **수용 후** (지속적) | 기존 registry entries vs source files | clause-anchor drift 탐지 (registry vs reality) |

두 verbs 는 공존하며 함께 사용된다.

## 3. EARS Requirements

### 3.1 Ubiquitous (REQ-CDL-001..005) — 항상 참인 불변식

- **REQ-CDL-001**: The system shall ensure that every `[HARD]` rule in the 15 canonical constitution source files (CLAUDE.md + 14 files under `.claude/rules/moai/{core,workflow,development,design}/`) carries exactly one `[ZONE:Frozen]` or `[ZONE:Evolvable]` marker. The authoritative file list is enumerated in spec.md §2.2.

- **REQ-CDL-002**: The zone classification field `zone_class` in `zone-registry.md` entries shall be one of: `frozen-canonical` | `frozen-safety` | `evolvable-tuning` | `evolvable-experimental`.

- **REQ-CDL-003**: The `moai constitution validate` CLI verb shall be invokable with the `--format json` flag for CI integration, producing machine-parseable output compatible with `jq`.

- **REQ-CDL-004**: The `zone-registry.md` document shall achieve 100% [HARD] coverage, defined as the count of zone-registry entries being equal to or greater than the count of [HARD] rules across all 15 source files listed in §2.2 (empirical baseline M=111 at main HEAD `3bd2aa291`).

- **REQ-CDL-005**: All zone-registry entries newly introduced by this SPEC shall use the ID format `CONST-V3R5-NNN` (where NNN is a zero-padded 3-digit decimal) beginning at `CONST-V3R5-001`. This is a **parallel namespace** to the existing `CONST-V3R2-NNN` series; the V3R2 sequence is preserved unchanged (including its 3 internal gaps at 047/048/050).

### 3.2 Event-Driven (REQ-CDL-006..009) — 외부 트리거 기반

- **REQ-CDL-006**: When `moai constitution validate` is invoked and detects that a registered clause text no longer matches the source file at the recorded anchor, the CLI shall exit with code 1 and emit a JSON report listing each affected entry's `id`, `file`, `anchor`, and `status` field set to `DRIFT`. When a referenced source file is missing entirely, the CLI shall exit with code 2 and set `status: SOURCE_FILE_MISSING`.

- **REQ-CDL-007**: When CI workflow `.github/workflows/ci.yml` runs the `moai constitution validate --strict` step on a pull request, the workflow step shall fail (non-zero exit) if any drift is detected, blocking merge to `main`.

- **REQ-CDL-008**: When a [HARD] rule is added to any of the 15 canonical constitution source files (§2.2) without a corresponding `zone-registry.md` entry, `moai constitution validate --strict` shall fail with error key `ZONE_UNREGISTERED`, reporting the source file path and approximate line number of the unmapped rule.

- **REQ-CDL-009**: When `zone-registry.md` is updated and `moai constitution list --zone evolvable` is subsequently invoked, the output shall reflect the updated set of evolvable entries on the next invocation without requiring restart, daemon, or cache invalidation.

### 3.3 State-Driven (REQ-CDL-010..011) — 특정 모드 조건부

- **REQ-CDL-010**: While the process environment variable `CI` is set to `true` (any truthy value), `moai constitution validate` shall behave as if `--strict` was passed, failing on any drift or unregistered rule.

- **REQ-CDL-011**: While the process environment variable `MOAI_CONSTITUTION_SKIP_VALIDATE` is set to `1`, `moai constitution validate` shall emit a single warning line to stderr (e.g., `WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)`) and exit with code 0, regardless of drift state. This provides a development-only override for known transient drift.

### 3.4 Optional (REQ-CDL-012..013) — 조건부 기능

- **REQ-CDL-012**: Where `--format json` is provided to `moai constitution validate`, the output shall conform to JSON Schema v1.0 with the following top-level fields: `status` (string: `ok` | `drift` | `missing`), `drift_count` (integer), `missing_count` (integer), `unregistered_count` (integer), `entries` (array of objects with `id`, `file`, `anchor`, `status`, `detail`).

- **REQ-CDL-013**: Where a constitution `.md` file referenced by zone-registry entries has been deleted, `moai constitution validate` shall list every affected entry in the JSON output with `status: SOURCE_FILE_MISSING` and produce exit code 2 (distinct from drift exit code 1, to allow CI to distinguish file-deletion incidents from text drift).

### 3.5 Unwanted (REQ-CDL-014..016) — 금지 행동

- **REQ-CDL-014**: The `moai constitution validate` command shall not modify any source file, configuration, or state file. It is strictly read-only. Any future write operations require a separate command verb (e.g., `moai constitution sync`) which is out of scope for this SPEC.

- **REQ-CDL-015**: A zone-registry entry shall not declare `zone: Frozen` while having `canary_gate: false`. Frozen zone classification implies safety-criticality and requires `canary_gate: true` per existing SPEC-V3R2-CON-001 §7 OQ6 policy. `moai constitution validate --strict` shall reject such entries with error key `FROZEN_WITHOUT_CANARY`.

- **REQ-CDL-016**: A new zone-registry entry shall not reference an `anchor:` value that does not exist as a heading or section marker in the target `file:`. `moai constitution validate --strict` shall reject such entries with error key `ANCHOR_NOT_FOUND`.

### 3.6 Integrity (REQ-CDL-017..019) — 신규 sentinel error keys 거버닝

- **REQ-CDL-017**: A zone-registry entry's `id` field shall be globally unique across the registry. When two or more entries declare the same `id`, `moai constitution validate` shall fail with error key `DUPLICATE_ID` and exit code 1 regardless of `--strict`.

- **REQ-CDL-018**: A single `[HARD]` rule line in a constitution source file shall not be annotated with more than one `[ZONE:*]` marker. When `moai constitution validate` detects multiple zone markers on the same rule occurrence, it shall emit a warning with error key `DUPLICATE_ZONE_MARKER`. Warnings do not affect exit code unless `--strict` and `--fail-on-warning` are both set.

- **REQ-CDL-019**: A zone-registry entry's last-update timestamp older than 90 days (relative to invocation time) shall be reported as `STALE_ENTRY` warning. This is observation-only and does not affect exit code. Future SPECs may promote stale entries to a separate refresh workflow.

## 4. Acceptance Criteria

### AC-CDL-001 (D1 annotation completeness)

**Given**: The 15 canonical constitution source files (enumerated in §2.2) collectively contain N total `[HARD]` rule occurrences. The empirical baseline at main HEAD `3bd2aa291` is N=111.

**When**: The command sums `grep -hcE '\[HARD\]' <file>` across all 15 files (X), and the command sums `grep -hcE '\[ZONE:(Frozen|Evolvable)\]' <file>` across the same 15 files (Y).

**Then**: Y must be greater than or equal to X. Every [HARD] rule has at least one zone marker. (Multiple zone markers on the same rule are permitted but produce a `DUPLICATE_ZONE_MARKER` warning per REQ-CDL-018.)

### AC-CDL-002 (D2 100% coverage)

**Given**: The `zone-registry.md` parsed via `moai constitution list --format json` returns an array of N entries. The empirical baseline at main HEAD `3bd2aa291` is N=72 (pre-SPEC).

**When**: The total `[HARD]` rule count across all 15 canonical source files (§2.2) is M. Empirical baseline M=111 at main HEAD `3bd2aa291`.

**Then**: After Phase B completes, N must equal or exceed M (target N≥111). No `[HARD]` rule in any of the 15 source files remains unmapped to a registry entry. The `moai constitution validate --strict` command exits with code 0 on this condition. The Phase B deliverable count is **39 new entries** (111 − 72 = 39).

**Source file list (binding for AC-CDL-002 M computation)**: Identical to §2.2 table — 15 files. Any future addition of constitution files requires updating §2.2 and re-running the M computation.

### AC-CDL-003 (D3 validate CLI happy path)

**Given**: The `zone-registry.md` is fully in sync with all source files (no clause has been modified since registry entry was created, no source files are missing, no [HARD] rules are unregistered).

**When**: `moai constitution validate --strict --format json` is executed from the project root.

**Then**: The command exits with code 0 and the JSON output equals `{"status":"ok","drift_count":0,"missing_count":0,"unregistered_count":0,"entries":[]}` (or semantically equivalent with empty arrays). The command completes in under 5 seconds for the current corpus size (≤ 150 entries, ≤ 15 source files).

### AC-CDL-004 (D3 validate CLI drift detection)

**Given**: A constitution `.md` file is modified such that the text matching a registered entry's `clause:` field has been changed, deleted, or relocated to a different anchor.

**When**: `moai constitution validate --strict` is executed.

**Then**: The command exits with code 1. The JSON output (with `--format json`) contains an `entries` array with at least one element having `status: DRIFT` and `id` matching the affected registry entry. The textual output (without `--format json`) prints a human-readable diff summary identifying the affected entry.

### AC-CDL-005a (CI integration — automatable)

**Given**: A pull request is opened targeting the `main` branch of the moai-adk-go repository. The `.github/workflows/ci.yml` file has been updated with a `Constitution Validate` job per Phase D Task D-1.

**When**: The PR triggers the `ci.yml` workflow. The `constitution-validate` job executes `./moai constitution validate --strict --format json` with `CI=true` env.

**Then**: If the PR maintains registry-source sync, the job exits with code 0 and the step appears as ✓ in PR checks list. If the PR introduces drift, the job exits with code 1 and the step appears as ✗. This is **auto-verifiable** within a single CI run.

### AC-CDL-005b (Branch protection — manual maintainer verification)

**Given**: The Phase D Task D-2 branch protection update has been applied by a maintainer with admin permission after PR merge.

**When**: `gh api /repos/modu-ai/moai-adk/branches/main/protection` is invoked.

**Then**: The response's `required_status_checks.contexts` array contains `Constitution Validate` alongside the existing 4 entries (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`, `CodeQL`), totaling 5 required checks. This verification is **manual** because branch protection updates require admin permission that CI runners do not possess. CLAUDE.local.md §18.7 baseline updates from 4→5 in the same commit cycle as Phase D.

### AC-CDL-006 (REQ-CDL-002 zone_class enum compliance)

**Given**: All zone-registry entries (both pre-existing CONST-V3R2-NNN and newly added CONST-V3R5-NNN) declare a `zone_class:` field.

**When**: A YAML schema validation parses each entry's `zone_class` value against the allow-list `[frozen-canonical, frozen-safety, evolvable-tuning, evolvable-experimental]`.

**Then**: Every entry's `zone_class` matches one of the 4 allowed values. `moai constitution validate --strict` rejects entries with invalid `zone_class` values with error key `INVALID_ZONE_CLASS` and exit code 1.

### AC-CDL-007 (REQ-CDL-005 ID format compliance for new entries)

**Given**: The set of zone-registry entries newly introduced by this SPEC (all entries with id matching `^CONST-V3R5-`).

**When**: The command `grep -oE '^- id: CONST-V3R5-[0-9]{3}$' zone-registry.md | sort -u` returns the new entry IDs, and each is regex-matched against `^CONST-V3R5-[0-9]{3}$`.

**Then**: Every new entry ID matches the regex exactly. Sequencing begins at `CONST-V3R5-001` and is contiguous (no internal gaps in the V3R5 namespace introduced by this SPEC).

### AC-CDL-008 (REQ-CDL-009 live reload — no restart between writes)

**Given**: An in-progress shell session in which `./moai constitution list --zone evolvable` has been invoked once.

**When**: A test fixture modifies `zone-registry.md` (e.g., adds one new entry with `zone: Evolvable`), then immediately re-invokes `./moai constitution list --zone evolvable` within the same shell session without restart.

**Then**: The second invocation reflects the updated entry set. No daemon, cache, or restart is required. Validator test fixture `TestList_ReflectsUpdatesWithoutRestart` covers this scenario.

### AC-CDL-009 (REQ-CDL-011 MOAI_CONSTITUTION_SKIP_VALIDATE override)

**Given**: A controlled drift scenario exists (registry entry clause does not match source). Without override, `moai constitution validate --strict` returns exit code 1.

**When**: The same command is invoked with `MOAI_CONSTITUTION_SKIP_VALIDATE=1` in the environment.

**Then**: The command emits a warning line to stderr (`WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)`) and exits with code 0. The drift is not reported in stdout. In CI workflow YAML, this env is explicitly cleared (`env: { MOAI_CONSTITUTION_SKIP_VALIDATE: '' }`) to prevent CI bypass.

### AC-CDL-010 (REQ-CDL-014 read-only assertion)

**Given**: A controlled test environment where the project directory is made read-only (`chmod -R -w .` or equivalent) except for `/tmp` scratch area.

**When**: `./moai constitution validate --strict --format json` is executed. Then the read-only attribute is reverted.

**Then**: The command produces a complete validation report (success or drift detection) without any write attempt to source files, registry, configuration, or state. No `EACCES`/`EROFS` errors due to attempted writes. Validator test fixture `TestValidate_ReadOnlyAssertion` covers this scenario.

## 5. Scope

### 5.1 In Scope

- 15 헌법 source files (§2.2) 에 `[ZONE:Frozen]` / `[ZONE:Evolvable]` 마커 inline 부착 (D1)
- `zone-registry.md` 100% coverage 확장 (39 new entries, target N≥111) + `zone_class` 4-classification 도입 (D2)
- `moai constitution validate` CLI verb 신설 (D3) — drift 탐지 + missing source 탐지 + unregistered HARD rule 탐지
- `.github/workflows/ci.yml` validate step 추가 + main 브랜치 보호 required check 등록 (D3 CI integration)

### 5.2 Out of Scope (EXCL-001 through EXCL-006)

본 SPEC은 다음을 **명시적으로 제외**한다:

- **EXCL-001 (PreToolUse Frozen Guard hook 미구현)**: W1은 PreToolUse hook scaffold를 구현하지 않는다. 해당 작업은 W3 HARNESS-AUTONOMY-001의 scope. 단, W3 hook이 W1의 zone-registry를 참조 데이터로 사용할 수 있도록 SSOT을 제공한다.
- **EXCL-002 (agent/skill frontmatter `zone:` 필드 미추가)**: agent 또는 skill 파일의 frontmatter에 `zone: frozen` / `zone: evolvable` 필드를 추가하는 작업은 T3 Full scope이다. v3.5.0 이후 follow-up SPEC (가칭 SPEC-V3R5-AGENT-ZONE-001)으로 이관한다.
- **EXCL-003 (expert-backend / expert-frontend / expert-mobile retirement 미수행)**: 12 W2-deferred residual을 직접 해소하지 않는다. agent retirement는 W2 CORE-SLIM-001의 scope. W1은 retirement의 헌법적 근거 (FROZEN canonical 17 agent 정의)만 제공한다.
- **EXCL-004 (V3R3/V3R4 workflow rule retroactive classification 제한)**: §2.2 의 15 source files 외 추가 파일 (예: 미래 신규 workflow rule) 을 retroactive하게 zone-classification하지 않는다. 본 SPEC의 coverage 목표는 §2.2 enumerated EXISTING `[HARD]` 규칙 111개 매핑 완료까지로 한정한다.
- **EXCL-005 (design/constitution.md 구조 변경 금지)**: `.claude/rules/moai/design/constitution.md`는 이미 §2에서 FROZEN/EVOLVABLE 구조를 정의하므로 본 SPEC은 구조를 변경하지 않는다. inline `[ZONE:*]` 마커 추가만 수행하며, 기존 §2 zone list와 정합성을 검증한다.
- **EXCL-006 (precommit/generation-time check 미도입)**: 본 SPEC은 generation-time hook (예: precommit) 을 도입하지 않는다. CI-time 검증 (GitHub Actions `.github/workflows/ci.yml`) 만 강제한다. 로컬 개발자는 `make preflight` 또는 수동 `moai constitution validate` 실행 가능.

## 6. DELTA Markers (Brownfield)

본 SPEC은 brownfield 작업으로 다음과 같이 기존 파일을 [MODIFY]하거나 [NEW] 파일을 추가한다:

| Marker | 파일 경로 | 변경 유형 |
|--------|-----------|-----------|
| [EXISTING] | `.claude/rules/moai/core/zone-registry.md` schema + ID allocation policy | 기존 schema unchanged, V3R2 sequence 보존 (gaps 047/048/050 포함), V3R5 entry 추가만 수행 |
| [MODIFY] | `CLAUDE.md` (§1/§7/§8/§14/§19) | inline `[ZONE:*]` 마커 추가, 텍스트 의미 변경 없음 |
| [MODIFY] | `.claude/rules/moai/core/moai-constitution.md` | inline `[ZONE:*]` 마커 추가 |
| [MODIFY] | `.claude/rules/moai/core/agent-common-protocol.md` | inline `[ZONE:*]` 마커 추가 |
| [MODIFY] | `.claude/rules/moai/core/askuser-protocol.md` | inline `[ZONE:*]` 마커 추가 |
| [MODIFY] | `.claude/rules/moai/design/constitution.md` | inline `[ZONE:*]` 마커 추가 (기존 §2 zone list와 정합성 검증) |
| [MODIFY] | `.claude/rules/moai/workflow/*.md` (7 files per §2.2) | inline `[ZONE:*]` 마커 추가 |
| [MODIFY] | `.claude/rules/moai/development/*.md` (3 files per §2.2) | inline `[ZONE:*]` 마커 추가 |
| [MODIFY] | `.claude/rules/moai/core/zone-registry.md` | CONST-V3R5-001..NNN entry 추가 (목표 ≥ 39 신규 entries, 총 ≥ 111) + 기존 entries 에 `zone_class` 필드 retroactive 부여 |
| [NEW] | `internal/constitution/validator.go` | drift 탐지 로직 (~150 LOC 예상) |
| [NEW] | `internal/constitution/registry.go` | YAML registry parser (기존 코드 reuse 가능 시 minimal) |
| [NEW] | `internal/constitution/validator_test.go` | unit + fixture-based drift simulation test (~200 LOC 예상) |
| [NEW] | `internal/cli/constitution_validate.go` | CLI command wiring (~100 LOC 예상) |
| [MODIFY] | `internal/cli/constitution.go` | `validate` subcommand 등록 (~10 LOC 예상) — 기존 `guard`/`list`/`amend` verbs 와 직교 |
| [MODIFY] | `.github/workflows/ci.yml` | validate step 추가 (~20 LOC YAML) |

## 7. 제약 조건 (Constraints)

- **C-CDL-001 (Backward compatibility)**: 기존 `moai constitution list`, `guard`, `amend` 동작은 변경 불가. CONST-V3R2-NNN entry 형식 및 zone enum (`Frozen` | `Evolvable`)은 유지. `zone_class` 필드는 optional 로 도입하여 기존 entry 영향 없음 (단, AC-CDL-006 에 따라 모든 entries 에 retroactive 부여).

- **C-CDL-002 (16-language neutrality)**: validate CLI 출력은 영어. 단, error message는 향후 i18n 가능하도록 sentinel error key (`DRIFT`, `SOURCE_FILE_MISSING`, `ZONE_UNREGISTERED`, `FROZEN_WITHOUT_CANARY`, `ANCHOR_NOT_FOUND`, `DUPLICATE_ID`, `DUPLICATE_ZONE_MARKER`, `STALE_ENTRY`, `INVALID_ZONE_CLASS`) 분리.

- **C-CDL-003 (Performance)**: validate 명령은 현재 corpus size (≤ 150 entries, ≤ 15 source files) 에서 5초 이내 완료. 미래 corpus 확장에 대비하여 file read는 1회 캐시.

- **C-CDL-004 (No external dependencies)**: validator는 표준 Go 라이브러리 + 기존 `gopkg.in/yaml.v3` (이미 프로젝트에 존재) 만 사용. 신규 외부 dependency 도입 금지.

- **C-CDL-005 (TRUST 5 quality gate)**: `internal/constitution/` package coverage ≥ 85%. linter zero warnings. validator는 read-only (REQ-CDL-014).

- **C-CDL-006 (CLI verb orthogonality)**: 신규 `validate` verb 는 기존 `guard` verb 와 직교 (§2.5). `guard` = pre-amendment safety check (proposed change → constitution). `validate` = post-acceptance drift detection (registry → reality). 두 verbs 는 함께 사용되며 어느 한 쪽이 다른 쪽을 대체하지 않는다.

## 8. REQ ↔ AC Traceability Matrix

[HARD] 모든 REQ-CDL-* 는 ≥1 AC-CDL-* 에 매핑되어야 한다 (RQ-5 traceability rule).

| REQ ID | Primary AC | Verification Method |
|--------|------------|---------------------|
| REQ-CDL-001 (annotation completeness) | AC-CDL-001 | Grep-based zone marker count vs HARD count |
| REQ-CDL-002 (zone_class enum) | AC-CDL-006 | YAML schema validation against 4-enum allow-list |
| REQ-CDL-003 (--format json flag) | AC-CDL-003 | CLI invocation with `--format json` + jq parse |
| REQ-CDL-004 (100% coverage) | AC-CDL-002 | Registry entry count N ≥ HARD count M=111 |
| REQ-CDL-005 (ID format CONST-V3R5-NNN) | AC-CDL-007 | grep regex `^CONST-V3R5-[0-9]{3}$` per new entry |
| REQ-CDL-006 (drift detection + missing source exit codes) | AC-CDL-004 | Controlled drift fixture + exit-code assertion |
| REQ-CDL-007 (CI block on drift) | AC-CDL-005a | CI workflow step exit code on drift PR |
| REQ-CDL-008 (ZONE_UNREGISTERED) | AC-CDL-004 | Unregistered HARD rule fixture → error key |
| REQ-CDL-009 (live reload) | AC-CDL-008 | Edit-then-invoke test without restart |
| REQ-CDL-010 (CI=true → --strict) | AC-CDL-005a | CI env asserts strict mode |
| REQ-CDL-011 (MOAI_CONSTITUTION_SKIP_VALIDATE) | AC-CDL-009 | Env override fixture → exit 0 + stderr warn |
| REQ-CDL-012 (JSON Schema v1.0) | AC-CDL-003 | jq-parseable JSON with all 5 top-level fields |
| REQ-CDL-013 (SOURCE_FILE_MISSING exit 2) | AC-CDL-004 | File-deletion fixture → exit 2 + status |
| REQ-CDL-014 (read-only) | AC-CDL-010 | Read-only filesystem assertion test |
| REQ-CDL-015 (FROZEN_WITHOUT_CANARY) | AC-CDL-004 | Frozen + canary_gate=false fixture → reject |
| REQ-CDL-016 (ANCHOR_NOT_FOUND) | AC-CDL-004 | Missing-anchor fixture → reject |
| REQ-CDL-017 (DUPLICATE_ID) | AC-CDL-004 | Duplicate id fixture → exit 1 always |
| REQ-CDL-018 (DUPLICATE_ZONE_MARKER) | AC-CDL-001 | Multi-marker line emits warning |
| REQ-CDL-019 (STALE_ENTRY) | AC-CDL-004 | Entry timestamp > 90 days fixture → warning |

Coverage: 19 REQs ↔ 10 ACs (AC-CDL-001..010). Every REQ has ≥1 AC. Multiple REQs map to AC-CDL-004 because the validator dispatch path handles all error keys uniformly; individual REQs are differentiated by test fixtures (one fixture per error key).

## 9. 참조 자료 (References)

- `.moai/research/harness-autonomy-vision-2026-05-18.md` §3.1 — Two-Zone Architecture
- `.moai/research/architecture-audit-2026-05-18.md` F-006 — zone-registry coverage gap (iteration 2 empirical refresh)
- `.claude/rules/moai/core/zone-registry.md` — 현재 72 entries (CONST-V3R2-001..046 + 049 + 051..072 + 150..152)
- `.claude/rules/moai/design/constitution.md` v3.4.0 §2 — FROZEN/EVOLVABLE 선행 패턴
- `.moai/state/lint-w2-deferred.json` — W2 retirement target manifest
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema (본 SPEC 준수)
- `CLAUDE.local.md` §18.7 — Branch protection rule (validate step required check 등록 경로)
- `.moai/reports/plan-audit/SPEC-V3R5-CONSTITUTION-DUAL-001-review-1.md` — iteration 1 audit driving this revision
