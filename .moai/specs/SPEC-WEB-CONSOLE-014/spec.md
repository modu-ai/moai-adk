---
id: SPEC-WEB-CONSOLE-014
title: "Web Console — Existing-Section Key Augmentation + Dormant-Config Exposure Guards (Track 3)"
version: "0.2.0"
status: in-progress
created: 2026-07-10
updated: 2026-07-11
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/settings"
lifecycle: spec-anchored
tags: "web, console, settings, config-exposure, dormant-guards, raw-view, seam, i18n"
tier: M
era: V3R6
depends_on: [SPEC-WEB-CONSOLE-013]
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-MERGE-METHOD-CONFIG-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.2.0 | 2026-07-10 | manager-spec | plan-audit iter-1 (PASS-WITH-DEBT 0.86) D1-D6 정정. D1: F6 sandbox 리스트 키를 **config-미배선 scaffold**로 재분류(config→Options 브리지 부재 — Options 유일 populator는 내장 Default, `EnvScrubExtra`는 Options 필드 자체 부재, `ScrubEnv(parent, passthrough)` 파라미터 없음) → REQ-020에 F2-style 정직 라벨 + 배선 Out of Scope + pre-flight 반증 grep 추가. D2: F9 분할 — `output_path`는 dead config(reader 0건, 기록 경로는 `hookMetricsRelPath` 상수) → REQ-040 분할(threshold editable 핀 / output_path read-only 강등 + 쓰기 거부), AC-040a/b. D5: merge_method absent-key 초기 표시 처리(empty NOT a member) REQ-010 보강 + AC-010c. D3(013 실존 갱신)/D4(raw view 카운트 6건)/D6(AC-010a 테스트명 바인딩)은 plan/acceptance 측 정정. |
| 0.1.0 | 2026-07-10 | manager-spec | 최초 draft. 2026-07-10 CLI/web 감사 Track 3 위임 브리프 기반 + 전 대상 키 런타임 reader 실측 (acceptance.md §B 증거). 브리프 대비 실측 정정 2건: (1) `learning.rate_limit.*`는 enforcement가 컴파일 상수(config 미배선)라 편집 노출 대신 read-only raw view, (2) `security.permission.{pre_allowlist,session_rules}`는 Go struct 바인딩 부재(yaml scaffold)라 노출 대상이 아니라 금지 가드 대상. |

---

## §A Context & Motivation

`moai web` 콘솔의 필드 집합 SSOT(`internal/settings` — schema.go 34필드 + schema_sections.go 확장, SPEC-WEB-CONSOLE-011이 확립)는 v2.14.0 이후 **이미 노출된 섹션 내부에 추가된 신규 키**를 반영하지 못했고, 동시에 **절대 편집 표면에 나타나면 안 되는 dormant/governance config**에 대한 회귀 가드가 없다. 본 SPEC은 두 갭을 함께 닫는다:

- **P2 증강**: harness `learning` 블록, git-strategy `merge_method`, security sandbox 리스트 키, mx 정책 맵/리스트, observability hook_metrics(iter-2 분할 — `slow_hook_threshold_ms` editable 핀 / `output_path` read-only 강등)를 기존 seam/typed 경로 재사용으로 콘솔 표면에 정합화.
- **P3 노출 금지 가드**: sunset.*, harness.model_upgrade_review.*, tool-policy, workflow.model_routing 및 실측으로 추가된 금지 대상이 편집 가능 집합에 유입되는 것을 기계적으로 차단하는 회귀 테스트.

### §A.1 실측 Findings (2026-07-10 — 전체 명령+출력은 acceptance.md §B; 라인 앵커는 실측 시점, run-phase에서 content-token 기준 재검증)

| # | 키 | 런타임 reader 증거 | 처분 |
|---|-----|--------------------|------|
| F1 | `learning.tier_thresholds` | **행동적**: lesson capture hook의 tier 분류가 이 값을 소비 (`internal/cli/hook.go` `readTierThresholds` → `ClassifyTier`) | read-only raw view (리스트 값 — 리스트 편집 위젯 부재) |
| F2 | `learning.rate_limit.{max_per_week,cooldown_hours}` | **표시 전용**: `moai harness` 상태 출력만 소비 (`internal/cli/harness.go`). enforcement rate limiter는 컴파일 상수 사용 (`internal/harness/safety/rate_limit.go` `rateLimitMaxUpdates = 3` / `rateLimitCooldown = 24h`; `NewRateLimiter(statePath)`는 config 값을 받지 않음) | read-only raw view — 편집 노출은 config theater (**브리프 정정 1**) |
| F3 | `learning.auto_apply` | **표시 전용 + governance FROZEN false**: `internal/cli/harness/execute.go` @MX:NOTE — 파이프라인 in-memory `AutoApply: true`는 디스크 값(`auto_apply: false`)을 절대 변경하지 않는 FROZEN 불변식 | 현행 편집 필드를 **철거**하고 read-only 표시 + 거버넌스 설명으로 강등 |
| F4 | `git_strategy.<mode>.merge_method` | **agent-prose consumer** (SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007: sync delivery/manager-git가 active mode의 값으로 `gh pr merge` 플래그 선택); enum SSOT = `internal/config/validation.go` `validMergeMethods` (squash\|merge\|rebase) | 편집 select ×3 mode (typed dirty-flag Save 경로) |
| F5 | `git_strategy.team.branch_creation.{prompt_always,auto_enabled}` + `automation.{auto_branch,auto_pr}` | **Go 행동적 reader 0건** (internal/config types/defaults 밖 grep 공집합; `PromptAlways`는 소스 주석상 "forward-compat scaffold") | 비노출 + 부재 가드 |
| F6 | `security.sandbox.{network_allowlist,env_scrub_extra}` | **config-미배선 scaffold (iter-2 재분류 — F2와 동일 클래스)**: sandbox 패키지는 Options 필드를 소비하나 **config→Options 브리지가 미배선** — Options.NetworkAllowlist의 유일 populator는 내장 Default(`internal/cli/doctor_sandbox.go`의 `sandbox.DefaultNetworkAllowlist`)이고, `EnvScrubExtra`는 sandbox Options 필드 자체가 부재하며 `ScrubEnv(parent, passthrough)`에 해당 파라미터가 없다(`env.go` 주석은 미구현 약속). 재실측 증거: acceptance.md §B E-13 | read-only raw view ×2 + **F2-style 정직 라벨**(값은 로드되지만 sandbox 실행 계층에 미배선; 배선은 별도 SPEC) |
| F7 | `security.permission.{pre_allowlist,session_rules}` | **Go struct 바인딩 부재** (internal/config grep 공집합); `internal/permission/stack.go` `PreAllowlist()`는 SrcBuiltin 내장 정의로 config 미소비; security.yaml 자체 주석이 "실제 패턴은 stack.go에 정의"라 명시 | 비노출(편집·raw 모두) + 금지 가드 (**브리프 정정 2**) |
| F8 | `mx.danger_categories` + `mx.test_paths` | **행동적**: `internal/mx/danger_category.go` `LoadDangerConfig`가 `.moai/config/sections/mx.yaml`을 읽고 `mx query` fan-in/danger 분류가 소비 (`internal/cli/mx_query.go`) | read-only raw view (쓰기는 계속 RouteExcluded) |
| F9 | `observability.hook_metrics.{output_path,slow_hook_threshold_ms}` | **분할 (iter-2 정정)**: `slow_hook_threshold_ms`는 reader 실존(`internal/hook/post_tool_duration.go` 로컬 struct가 이 키만 디코드); `output_path`는 **dead config** — reader 0건, 실제 기록 경로는 `hookMetricsRelPath` 상수(post_tool_duration.go)로 고정. 재실측 증거: acceptance.md §B E-14 | `slow_hook_threshold_ms`: editable 잔류 회귀 핀; `output_path`: 편집 필드 **철거** → read-only 표시 + dead-config 라벨 |
| F10 | `sunset.*` / `harness.model_upgrade_review.*` | **DORMANT**: sunset.yaml 헤더 주석 verbatim "typed and loaded, but no runtime hot path enforces"; model_upgrade_review는 `harness_validate.go`의 리마인더 출력만 소비(비강제) | never-editable 가드 |
| F11 | `workflow.model_routing` | legacy flat DEPRECATED 블록; Model Policy 페이지는 SPEC-WEB-CONSOLE-013 소관 | 본 SPEC은 **가드만** 소유 |
| F12 | `tool-policy` | dev-only — 템플릿 배포 의도적 제거 | RouteExcluded 잔류 핀 가드 |

### §A.2 노출 3-계층 원칙 (본 SPEC의 판정 기준)

1. **편집 가능(editable)**: 행동적 소비자(Go 런타임 reader 또는 SPEC으로 문서화된 agent-prose consumer 계약)가 존재하는 스칼라/enum 키만.
2. **read-only raw view**: 행동적 reader가 존재하지만 리스트/맵 값이라 form UI에 부적합한 키 (기존 RawViewBlocks 선례), 또는 값 표시가 유용하나 편집이 오해를 유발하는 키.
3. **비노출 + 가드**: reader 부재(scaffold), governance frozen, dormant, 타 SPEC 소관 키 — 기계적 회귀 테스트로 유입 차단.

이 원칙은 SPEC-WEB-CONSOLE-011 M4 다이어트("런타임 reader 없는 키 제거")와 SPEC-WEB-CONSOLE-008 "honest hybrid"(config theater 제거) 선례의 연속이다.

---

## §B Requirements (GEARS)

### §B.1 harness learning 정직화 (F1-F3)

- **REQ-WC14-001**: 설정 스키마는 `learning.auto_apply`를 읽기 전용 표시 키(거버넌스 설명 라벨 포함)로 제공해야 하며(shall), 이에 대한 편집 가능 필드를 제공하지 **않아야 한다(shall not)**. **When** 쓰기 요청이 `learning.auto_apply`를 대상으로 감지되면, 설정 적용 계층은 해당 쓰기를 거부해야 한다.
- **REQ-WC14-002**: 콘솔은 `learning.tier_thresholds`를 읽기 전용 raw view로 렌더해야 한다(shall).
- **REQ-WC14-003**: 콘솔은 `learning.rate_limit` 블록을 읽기 전용 raw view로 렌더해야 하며(shall), 라벨/설명은 이 값이 정보성 표시(현행 enforcement는 컴파일 상수)임을 명시해야 한다.

### §B.2 git-strategy merge_method (F4-F5)

- **REQ-WC14-010**: 콘솔은 3개 mode profile(manual/personal/team) 각각의 `git_strategy.<mode>.merge_method`를 닫힌 enum select 편집 필드로 노출하고(shall), 기존 git-strategy typed 저장 경로(dirty-flag Save)로 영속화해야 한다. **While** live git-strategy.yaml에 `merge_method` 키가 부재한 동안(컴파일 기본 `squash` — REQ-MMC-003 partial-override 계약; 실측: live 0건 vs 템플릿 .tmpl 3건), 콘솔은 검증 실패 없이 select를 렌더하고 유효 컴파일 기본값을 표시해야 한다(shall). 빈 문자열은 enum 멤버가 **아니므로**(empty NOT a member), 저장은 항상 명시적 enum 값을 기록해야 한다.
- **REQ-WC14-011**: merge_method의 옵션 집합과 멤버십 검증은 기존 config enum SSOT(squash|merge|rebase)에서 파생해야 하며(shall), 리터럴 집합을 재선언하지 않아야 한다. **When** 저장 요청이 enum 밖의 merge_method 값을 담은 것이 감지되면, 설정 적용 계층은 해당 쓰기를 거부해야 한다.
- **REQ-WC14-012**: 콘솔은 `git_strategy.<mode>.branch_creation.{prompt_always,auto_enabled}` 및 `git_strategy.<mode>.automation.{auto_branch,auto_pr}`에 대한 편집 필드를 제공하지 **않아야 한다(shall not)** (F5 — Go 행동적 reader 부재).

### §B.3 security (F6-F7)

- **REQ-WC14-020**: 콘솔은 `security.sandbox.network_allowlist`와 `security.sandbox.env_scrub_extra`를 읽기 전용 raw view로 렌더해야 하며(shall), 라벨/설명은 이 값이 정보성 표시(config는 로드되지만 sandbox 실행 계층으로의 브리지가 미배선 — 배선은 별도 SPEC)임을 명시해야 한다 (F6, F2-style 정직 라벨).
- **REQ-WC14-021**: `security.permission.pre_allowlist`와 `security.permission.session_rules`는 편집 필드로도 raw view로도 노출되지 **않아야 하며(shall not)**, 회귀 가드 테스트가 편집 가능 집합에서의 부재를 고정해야 한다(shall) (F7 — Go-unbound scaffold).

### §B.4 mx read-only (F8)

- **REQ-WC14-030**: 콘솔은 `mx.danger_categories`와 `mx.test_paths`를 읽기 전용 raw view로 렌더해야 한다(shall).
- **REQ-WC14-031**: **While** mx 섹션이 쓰기 제외군(RouteExcluded)에 속하는 동안, 쓰기 라우터는 mx 섹션에 대한 모든 쓰기를 계속 거부해야 하며(shall), `mx.` prefix의 편집 가능 필드는 존재하지 않아야 한다(shall not).

### §B.5 observability 분할 처분 (F9 — iter-2 정정)

- **REQ-WC14-040**: `observability.hook_metrics.slow_hook_threshold_ms`는 편집 가능 집합에 잔류해야 한다(shall) — reader 실존(F9); 회귀 핀. 반면 `observability.hook_metrics.output_path`는 dead config(reader 0건; 기록 경로는 컴파일 상수)이므로 편집 필드를 **철거**하고 read-only 표시 + dead-config 설명 라벨로 강등해야 한다(shall). **When** 쓰기 요청이 `observability.hook_metrics.output_path`를 대상으로 감지되면, 설정 적용 계층은 해당 쓰기를 거부해야 한다.

### §B.6 노출 금지 가드 (F10-F12 + 실측 추가분)

- **REQ-WC14-050**: 회귀 가드 테스트는 편집 가능 필드 집합의 어떤 이름도 다음 금지 목록과 매칭되지 않음을 단언해야 한다(shall): prefix `sunset.` · `model_upgrade_review`(harness 하위 포함) · `workflow.model_routing` · `mx.` · `tool_policy`/`tool-policy`, 정확 이름 `security.permission.pre_allowlist` · `security.permission.session_rules` · `learning.auto_apply` · `observability.hook_metrics.output_path`(iter-2 D2) · F5의 4개 late-branch 키. 목록은 테이블 주도로 확장 가능해야 한다. (탑재 순서: 현행-clean 항목은 M1, 강등 동반 항목 2건은 M2 — plan.md B14.)
- **REQ-WC14-051**: `sunset` · `tool-policy` · `mx` 섹션은 쓰기 라우팅에서 계속 RouteExcluded로 남아야 한다(shall) — 회귀 핀.

### §B.7 교차 관심사

- **REQ-WC14-060**: 본 SPEC이 도입하는 모든 신규 사용자 표시 라벨(필드 라벨, raw view 제목, read-only 설명, 신규 렌더 그룹 제목)은 콘솔 i18n 사전의 4개 locale(en/ko/ja/zh) 전부에 키를 가져야 한다(shall).
- **REQ-WC14-061**: **When** 스키마 필드 집합이 변경되면, 웹 측 parity 테스트와 TUI 측 schema-bridge parity 테스트(`internal/cli/schema_bridge_test.go`)가 함께 갱신되어 green이어야 한다(shall) — SPEC-WEB-CONSOLE-011 M2b 회귀 교훈.
- **REQ-WC14-062**: 본 SPEC은 `internal/statusline/` 하위의 어떤 파일도 수정하지 **않아야 한다(shall not)** (병렬 세션 무접촉 계약 승계).

---

## §C Exclusions — Out of Scope

### Out of Scope — statusline 표면
- `internal/statusline/*` 및 statusline 전용 동기화 경로(RouteStatusline)는 무접촉 (REQ-WC14-062).

### Out of Scope — rate-limit config→enforcement 배선
- `learning.rate_limit.*` 값을 `internal/harness/safety/rate_limit.go` enforcement 상수에 배선하는 작업은 별도 wiring SPEC 소관. 본 SPEC은 정직한 read-only 표시까지만.

### Out of Scope — sandbox/observability config→실행 계층 배선
- `security.sandbox.{network_allowlist,env_scrub_extra}`의 config→sandbox Options 브리지 배선(Options populator 신설, `EnvScrubExtra` Options 필드·`ScrubEnv` 파라미터 신설)은 별도 wiring SPEC 소관 (F6).
- `observability.hook_metrics.output_path`를 실제 기록 경로(`hookMetricsRelPath` 상수)에 배선하는 작업도 동일하게 별도 SPEC 소관 (F9). 본 SPEC은 정직한 read-only 표시까지만.

### Out of Scope — permission scaffold 키 활성화
- `security.permission.{pre_allowlist,session_rules}`의 Go struct 바인딩·loader 배선·permission stack 소비는 활성화 SPEC 소관. 본 SPEC은 비노출 가드까지만.

### Out of Scope — Model Policy 페이지
- `workflow.model_routing` / model routing profiles의 웹 노출 설계는 SPEC-WEB-CONSOLE-013 소관. 본 SPEC은 "flat legacy 블록이 편집 표면에 유입되지 않는다"는 가드만 소유.

### Out of Scope — 리스트 편집 위젯 신설
- 리스트/맵 값용 편집 위젯(FieldType 신설)은 도입하지 않는다. 리스트 값 노출은 전부 기존 raw read-only view 패턴.

### Out of Scope — mx/tool-policy/sunset 쓰기 경로 개설
- ExcludedSections 멤버십 변경 및 신규 쓰기 라우트 등재 없음. mx raw view는 표시 전용이며 RouteForSection 결과를 변경하지 않는다.

### Out of Scope — dormant config 활성화
- sunset 조건 enforcement, model_upgrade_review 실행 로직 등 dormant 블록의 행동 활성화는 각 활성화 SPEC 소관.

---

## §D Constraints (비기능)

1. **seam 재사용**: 모든 신규 편집/표시 키는 기존 영속화·표시 메커니즘(yamlpatch seam, typed dirty-flag Save, ReadOnlyDisplayFields, RawViewBlocks)을 재사용한다. 신규 쓰기 경로·신규 위젯 금지.
2. **닫힌 집합 재선언 금지**: enum 옵션은 소유 패키지의 기존 SSOT 접근자에서 파생 (merge_method → `internal/config`의 enum SSOT).
3. **증거 기반 노출**: 편집 가능 판정은 §A.2 3-계층 원칙을 따르며, 각 판정의 grep 증거는 acceptance.md §B에 명령+출력으로 기록된다 (verification-claim-integrity §1.1 surface 3).
4. **의존성 직렬화**: SPEC-WEB-CONSOLE-013과 공유 파일(`internal/settings/schema_sections.go`, `sectionroute.go` 및 그 테스트)이 겹치므로 `depends_on: SPEC-WEB-CONSOLE-013` — 013 완결 후 run 진입 (plan.md §A.1 gate).
5. **쓰기 안전 모델 보존**: loopback bind + Host-check 미들웨어 모델 무변경, CSRF/token 인프라 도입 금지 (SPEC-WEB-CONSOLE-011 계약 승계).

---

## §E Traceability

REQ ↔ AC 매핑은 acceptance.md §C AC Matrix의 매핑 컬럼이 SSOT다. 총 16 REQ / 23 AC.
