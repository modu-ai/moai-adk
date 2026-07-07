---
id: SPEC-INTERNAL-ARCH-001
title: "internal/ 아키텍처 정리 — DI seam + package boundary + config unification"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.x target"
module: "internal/cli, internal/core, internal/config"
lifecycle: spec-anchored
tags: "refactor, architecture, di-seam, package-boundary, config-unification, behavior-preservation"
tier: L
depends_on: [SPEC-INTERNAL-TEST-001]
related_specs: [SPEC-CLI-SUBPKG-SPLIT-001]
---

# SPEC-INTERNAL-ARCH-001: internal/ 아키텍처 정리 — DI seam + package boundary + config unification

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2026-07-08 | manager-spec | 최초 저작. **감사 origin**: 2026-07-08 internal/ 전수 아키텍처 감사(orchestrator-delegated audit)에서 검증된 findings 6건(P1×2 + P2×4)을 GEARS 요구사항으로 변환. 전 앵커 read-only 재검증 완료(research.md §A 참조). |

---

## §A Context (배경)

moai-adk-go의 `internal/` 트리는 세 갈래의 구조적 부채를 안고 있다:

1. **`internal/cli`의 package-global `deps` singleton** (`internal/cli/deps.go:76 var deps *Dependencies`)이 후속 subpackage 추출을 차단한다. SPEC-CLI-SUBPKG-SPLIT-001 M1(agentlint cluster)은 성공했으나, plan.md §F CHECKPOINT(REQ-CSS-010)에서 "STOP (ship M1 only)" — 나머지 cluster는 tri-axis coupling으로 추출 불가 판정되었다. import-cycle 회피 주석 4곳(`agentlint/agent_lint.go:21`, `specid/specid.go:6`, `preference/cmd.go:141`, `update_preserve_inventory.go:~176`)이 이 결합을 증거한다.
2. **`internal/core`는 bare namespace** — 직접 `.go` 파일 0개, 공통 도메인 없는 3개 subpackage(git/project/quality)를 묶고, 죽은 빈 스텁 2개(integration/, migration/ — `.gitkeep`만 존재)를 방치하고 있다.
3. **`internal/config`의 이중 resolution pipeline** — `resolver.go`(1,156줄, 8-tier SettingsResolver, production 소비처는 `doctor_config.go` 단일 파일)와 `manager.go`+`loader.go`(ConfigManager, 사실상 전역 사용)가 공존하며 **precedence 의미가 다르다**(resolver는 env-var override 미처리). `moai doctor`의 config 진단이 런타임 실해석과 불일치할 수 있다. 부수적으로 loader/manager의 per-section boilerplate(~13개 near-identical 메서드 + Save/get/set 병렬 switch)와 `internal/config/CLAUDE.md`의 env-var 문서 drift(미구현 `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` override 주장)가 확인되었다.

본 SPEC은 **행위 보존(behavior-preserving) 구조 변경**이다. 관찰 가능한 CLI 행위(플래그, help 출력, exit code, 출력 형식, config 해석 결과)는 변경 전후 동일해야 하며, 모든 acceptance criteria는 이 불변식을 중심으로 구성된다. 회귀 위험이 카탈로그 내 최고 수준이므로 run-phase는 DDD(characterization-first) cycle로 수행하고, plan-auditor 독립 감사 + 구현 후 sync-auditor 감사를 권고한다(plan.md §H).

### 용어

- **DI seam**: subpackage가 `internal/cli`를 import하지 않고도 공유 의존성(`Dependencies` 상당)을 주입받을 수 있게 하는 leaf package + constructor-injection 계약. 상세 설계는 design.md §A.
- **bare namespace**: 직접 컴파일 단위(.go)가 0개이면서 무관한 subpackage들을 묶기만 하는 디렉터리.
- **green-to-green**: 각 구조 변경 commit 직전/직후 모두 전체 테스트 suite가 PASS인 상태를 유지하는 마이크로 커밋 규율.

---

## §B Requirements (GEARS)

### REQ-ARCH-001 — CLI Dependencies DI seam (Finding 1, P1)

- **(Ubiquitous)** The CLI dependency seam package shall expose the shared dependency capabilities consumed by `internal/cli` command subpackages without importing `internal/cli` (directly or transitively).
- **(Event-driven)** **When** a command subpackage requires shared CLI dependencies, the subpackage shall receive them via constructor injection through the seam package, not via the `internal/cli` package-global `deps` singleton.
- **(Event-driven)** **When** the seam lands, at least one existing subpackage (pilot: `internal/cli/agentlint` or `internal/cli/specid`) shall be migrated to consume the seam, and its import-cycle-avoidance workaround comment shall be removed.
- **(Unwanted)** The `internal/cli` package shall not retain a package-global mutable `var deps *Dependencies` singleton after milestone M1 completes; wiring shall be confined to explicit constructor calls at command registration.

### REQ-ARCH-002 — CLI monolith 분할: update.go / hook.go (Finding 2, P2)

- **(Capability gate)** **Where** the DI seam (REQ-ARCH-001) is in place, the `internal/cli/update.go` (3,172줄) and `internal/cli/hook.go` (1,182줄) compilation units shall be split by concern so that each resulting file addresses a single responsibility (concern map: design.md §B).
- **(Ubiquitous)** The SEC-HARDEN-003/004/005 path-guard family (`restoreTargetContained` / `parentChainContained`) shall reside in a dedicated compilation unit, separated from the unrelated config 3-way merge / backup logic it currently shares a file with.
- **(Ubiquitous)** The hook dispatcher, CLI subcommand registration, harness-classify logic, and DB-sync path construction shall reside in separate compilation units.
- **(Unwanted)** The split shall not relocate any symbol across package boundaries in this milestone (file-level split within `internal/cli` only; package-level extraction of update/hook clusters is out of scope — 후속 SPEC).

### REQ-ARCH-003 — internal/core bare namespace 해체 (Finding 3, P1)

- **(Ubiquitous)** The `internal/` package tree shall not contain a bare namespace directory that groups unrelated subpackages while carrying zero direct `.go` files.
- **(Event-driven)** **When** `internal/core` is dissolved, the `git/`, `project/`, `quality/` subpackages shall relocate to top-level `internal/` packages (candidate names: design.md §C — run-phase 충돌 pre-check 필수), and all external import call sites (fan-in 5/3/2 files, disjoint) shall be updated in the same commit as each move.
- **(Unwanted)** The repository shall not retain the dead `.gitkeep`-only stubs `internal/core/integration/` and `internal/core/migration/` (removal is the primary direction; populating them would be new-feature scope and is excluded — Out of Scope 참조).

### REQ-ARCH-004 — config 단일 resolution pipeline (Finding 4, P2)

- **(Ubiquitous)** Exactly one config resolution pipeline shall be reachable from production code paths.
- **(Event-driven)** **When** `moai doctor` renders config diagnostics, the rendered values shall derive from the same resolution semantics — including the 4 implemented env-var overrides — as the runtime ConfigManager pipeline.
- **(Unwanted)** The production binary shall not contain a second resolution pipeline whose precedence semantics diverge from ConfigManager (`resolver.go`의 8-tier SettingsResolver는 env-var override를 처리하지 않아 현재 이 조항을 위반한다).
- **(Event-driven, fallback gate)** **When** the retirement path (resolver.go 제거)가 run-phase에서 차단되면(design.md §D decision gate), the reconciliation fallback shall (a) resolver에 env-var override 정합성을 추가하고 (b) 두 pipeline의 의도적 차이(diagnostic-view vs runtime-view)를 `internal/config/CLAUDE.md`에 명시 문서화한 뒤에만 이 요구를 충족한 것으로 본다.

### REQ-ARCH-005 — config loader/manager table-driven 정리 (Finding 5, P2)

- **(Ubiquitous)** The per-section load/save/get/set plumbing in `internal/config` shall be table-driven from a single `sectionSpec` registry (설계: design.md §E), replacing the ~13 near-identical `load*Section` methods, the 7 near-identical `saveSection` calls in `Save`, and the parallel `getSectionLocked`/`setSectionLocked` switch statements.
- **(Event-driven)** **When** a new config section is added after this refactor, registering one `sectionSpec` entry shall be sufficient — no new per-section boilerplate method shall be required.
- **(Ubiquitous)** The refactored loader shall produce a `Config` value and `loadedSections` map deep-equal to the pre-refactor implementation for the same fixture inputs (characterization 대상).

### REQ-ARCH-006 — config env-var 문서 drift 수정 (Finding 6, P2)

- **(Ubiquitous)** `internal/config/CLAUDE.md` shall document exactly the implemented env-var override set: `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR` (in `applyEnvOverrides`) plus `MOAI_CONFIG_DIR` (config directory location).
- **(Unwanted)** The doc shall not claim `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` override behavior, nor cite an `EnvUserName` constant, that `envkeys.go` / `manager.go` do not implement (미구현 사실은 CLAUDE.local.md §9가 독립 확인).
- **(Capability gate)** **Where** the two env vars are later intended for implementation, that work shall be a separate SPEC — 본 SPEC은 문서를 코드에 맞추는 방향만 취한다.

### REQ-ARCH-007 — 행위 보존 불변식 (cross-cutting, 전 milestone 구속)

- **(Ubiquitous)** Every structural milestone shall be behavior-preserving: 전체 기존 테스트 suite(`go test ./...`)가 각 milestone 착수 전과 완료 후에 동일하게 PASS해야 한다.
- **(State-driven)** **While** any milestone is in progress, each landed commit shall keep `go test ./...` green (green-to-green micro-commits; 중간 RED commit 금지).
- **(Unwanted)** The refactor shall not change any exported CLI flag, `moai --help`(및 서브커맨드 help) output, machine-readable output format, or process exit code.
- **(Event-driven)** **When** a structural step turns the suite red, the implementing agent shall revert that step before proceeding (revert-on-red), and shall record the failed approach in progress.md §E.2.
- **(Event-driven)** **When** a refactor target lacks test coverage for its critical paths (특히 update.go의 merge/path-guard, hook.go의 dispatcher), characterization tests shall be added and verified green BEFORE the structural change (DDD PRESERVE phase).

### REQ-ARCH-008 — green baseline 선행 gate (dependency)

- **(Capability gate)** **Where** the sibling SPEC-INTERNAL-TEST-001 has restored the green `go test ./...` baseline (2026-07-08 실측 기준 2건 FAIL 잔존: `internal/spec` TestCloseSubjectDoctrineAmendment, `internal/statusline` TestBuild_WritesContextUsageWithSessionID — research.md §C), run-phase milestone M1 may begin.
- **(Unwanted)** Run-phase shall not begin on a red baseline — red baseline 위에서는 "행위 보존" 판정 자체가 불가능하다(REQ-ARCH-007의 전제).

---

## §C Constraints (제약)

- **[HARD] 행위 보존**: 본 SPEC의 어떤 milestone도 사용자 관찰 가능 행위를 변경하지 않는다. 행위 변경이 필요하다고 판명되면 해당 항목을 본 SPEC에서 제외하고 별도 SPEC으로 이관한다.
- **[HARD] green-to-green**: RED 상태 commit 금지. 각 commit은 독립적으로 `go build ./...` + `go test ./...` PASS.
- **DDD cycle**: quality.yaml의 전역 `development_mode: tdd` pin과 무관하게, 본 SPEC의 run-phase는 리팩터링 특성상 `cycle_type=ddd`(ANALYZE-PRESERVE-IMPROVE, characterization-first)로 위임할 것을 명시 권고한다.
- **파일 이동 시 git 이력 보존**: 가능한 한 `git mv` 상당의 rename-detectable 이동으로 blame 연속성 유지.
- **Milestone 독립 착지**: M3/M4/M5/M6은 상호 독립적으로 landable. M2는 M1 완료 + CHECKPOINT 통과에 종속(plan.md §F).
- **병렬 세션 방어**: 본 리포는 다중 세션 레이스 이력이 있으므로, run-phase spawn 전 pre-spawn sync check(agent-common-protocol §Pre-Spawn Sync Check) + 대상 파일 겹침 확인을 필수 전제로 한다(plan.md §C).

---

## §D Acceptance Criteria (개요)

전체 AC matrix(AC-ARCH-001 ~ AC-ARCH-008c), Given-When-Then 시나리오, edge cases, quality gates, Definition of Done은 **acceptance.md**가 canonical이다. 요약:

| REQ | AC 그룹 | 핵심 기계 검증 |
|-----|---------|----------------|
| REQ-ARCH-008 | AC-ARCH-001 | `go test ./...` exit 0 (baseline gate) |
| REQ-ARCH-001 | AC-ARCH-002a/b/c/d | seam 패키지 deps 그래프에 `internal/cli` 부재 + `var deps` 전역 제거 grep + pilot 주석 제거 grep |
| REQ-ARCH-002 | AC-ARCH-003a/b/c/d/e | update.go/hook.go 파일 크기 상한 + path-guard 전용 파일 grep + 관심사 분리 파일 존재 |
| REQ-ARCH-003 | AC-ARCH-004a/b/c | integration/migration 스텁 부재 + `internal/core/` import 참조 0건 grep + `go build` PASS |
| REQ-ARCH-004 | AC-ARCH-005a/b/c | `config.NewResolver(` production call site 0건 grep + doctor/runtime 정합 characterization |
| REQ-ARCH-005 | AC-ARCH-006a/b/c | `load*Section` 메서드 ≤1 grep + LOC 감소 ≥200 + Config deep-equal characterization |
| REQ-ARCH-006 | AC-ARCH-007a | CLAUDE.md 미구현 env-var 주장 0건 grep |
| REQ-ARCH-007 | AC-ARCH-008a/b/c | milestone 경계 suite green + `moai --help` byte-identical + characterization 선행 증거 |

---

## Out of Scope (제외 범위)

본 섹션은 본 SPEC이 구축하지 **않는** 것을 명시한다 (out of scope).

### Out of Scope — 기능 추가 및 행위 변경

- `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` env-var override의 **구현** (REQ-ARCH-006은 문서를 코드에 맞추는 방향만 취한다; 구현 의도가 확정되면 별도 SPEC)
- CLI 플래그/출력/exit code의 어떤 변경도 포함하지 않음 — 순수 구조 리팩터링

### Out of Scope — internal/cli 나머지 cluster의 package-level 추출

- uikit / profile / migrate / doctor / update cluster의 subpackage **추출**은 SPEC-CLI-SUBPKG-SPLIT-001 post-M1 checkpoint가 별도 SPEC으로 이관한 범위이며 본 SPEC에 포함되지 않는다. 본 SPEC의 REQ-ARCH-002는 `internal/cli` 내 **파일 단위** 관심사 분리까지만 수행한다 (추출의 전제 조건인 DI seam은 REQ-ARCH-001이 제공)

### Out of Scope — 테스트 baseline 복구

- 현재 RED인 2건(`internal/spec` doc-test AC-DLC-011, `internal/statusline` env flaky)의 수복은 sibling **SPEC-INTERNAL-TEST-001** 소관 — 본 SPEC은 그 green baseline에 의존(depends_on)만 한다

### Out of Scope — 죽은 스텁의 신규 구현

- `internal/core/integration/`, `internal/core/migration/` 스텁을 **채워 넣는** 신규 기능 개발 — 본 SPEC은 제거만 수행한다

### Out of Scope — 인접 패키지 자체 리팩터링

- `internal/template`, `internal/hook`, `internal/mx` 등 인접 패키지의 내부 구조 개선 — import 경로 갱신(REQ-ARCH-003 call-site 수정) 이상의 변경은 하지 않는다
- `internal/mx.NewResolver`는 config resolver와 무관한 별개 심볼이며 본 SPEC의 대상이 아니다

---

## §H Cross-references

- **acceptance.md** — AC matrix canonical (본 디렉터리)
- **plan.md** — milestone 순서, CHECKPOINT, 리스크 (본 디렉터리)
- **design.md** — DI seam 설계 + concern map + 이행 순서 (본 디렉터리)
- **research.md** — 결합 증거(실측 커맨드+출력) + 선행 사례 (본 디렉터리)
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/spec.md` REQ-CSS-010 — post-M1 CHECKPOINT STOP 선례
- SPEC-INTERNAL-TEST-001 — green baseline 복구 (depends_on; 본 SPEC 저작 시점 미저작 sibling)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter SSOT
- CLAUDE.local.md §9 — env-var 미구현 독립 확인, §14 하드코딩 금지(envkeys.go 상수 규율)
