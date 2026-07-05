---
id: SPEC-WEB-CONSOLE-CONFIG-DIET-001
title: "Web Console Config 과다 노출 정리 — Dead-Config 다이어트"
version: "0.1.0"
status: draft
created: 2026-07-05
updated: 2026-07-05
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/settings"
lifecycle: spec-anchored
tags: "web, console, config, dead-config, diet, schema, seam, yaml-node, correctness-bug"
tier: M
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-WEB-CONSOLE-010]
---

## HISTORY

- 2026-07-05: draft 최초 작성 (manager-spec). SPEC-WEB-CONSOLE-011 M2 재설계(약 163 필드로 확장)의
  후속. 5-에이전트 감사가 발견한 dead-config 과다 노출을 정리하는 "다이어트" SPEC.
  감사 findings는 clean worktree(`preview-wc011`, origin/main `97723664c`)에서 grep 재검증됨
  (research.md의 file:line 인용 참조). 본 SPEC은 **활성화(wiring)가 아닌 노출 축소**가 범위다.

---

## §A. 개요 (Overview)

### A.1 배경

SPEC-WEB-CONSOLE-011 M2 재설계는 moai web 콘솔을 약 163개 config 필드로 확장했다. `seam` 섹션들은
`internal/settings/schema_sections.go`의 `seamSectionFields()`를 통해 각 yaml 섹션의 **모든 키를
yaml.Node 경유로 편집 가능하게 렌더**한다. 그러나 Go 런타임은 이들 대부분에 대해 하드코딩 기본값을
쓰거나, 게이트가 CLI에 미배선(unwired)이거나, 소비 패키지가 이미 삭제된 상태다.

결과적으로 **편집 가능한 필드의 다수가 dead config**다 — 사용자가 콘솔에서 값을 바꿔도 런타임 동작에
**아무 효과가 없다**. 이는 특히 Tier 1(아래) 그룹에서 "설정을 바꿨는데 왜 안 먹지?"라는 사용자 신뢰
훼손으로 이어진다.

### A.2 핵심 원칙 — 다이어트, 활성화 아님

[HARD] 본 SPEC은 **노출 축소(diet)**다:
- IN: 콘솔이 dead / no-effect config를 편집 가능하게 제시하는 것을 중단한다. 주 레버는 렌더 SSOT
  (`internal/settings/schema.go` + `internal/settings/schema_sections.go`)에서 dead 키를 필터/강등/제거하는 것.
- OUT: 미배선 소비자(TrustGate, escalation, workflow_agents 런타임, role_profiles dispatch)를
  **배선(wiring)하는 것은 기능 작업**이며 다이어트가 아니다 — 명시적으로 제외하고 후속 SPEC으로 남긴다(§D).

### A.3 렌더 SSOT (수정 대상)

- `internal/settings/schema.go` — `AllSections()`, `allFields()`, 6개 수제 섹션.
- `internal/settings/schema_sections.go` — `sectionExtraFields()`, `seamSectionFields()`,
  `ReadOnlyDisplayFields()`, `RawViewBlocks()`, `SchemaSectionIDs()`.

두 표면(moai web templ 콘솔, moai profile setup TUI)이 모두 이 패키지를 import 하므로, 여기서
필드를 제거/강등하면 두 표면 모두에 일관되게 반영된다.

### A.4 "dead"의 3가지 결(nuance) — 중요

감사가 "dead"로 묶은 그룹들은 실제로 **효과 없음의 강도가 다르다**. 다이어트 결정은 이 결을 구분해야 한다:

| 결 | 정의 | 예시 | 다이어트 함의 |
|----|------|------|--------------|
| N1 진짜 사강(no-effect) | 소비 코드/패키지 부재, 어떤 경로로도 읽히지 않음 | research.yaml (패키지 삭제), llm.claude_models | 제거/은닉 안전 |
| N2 미배선 게이트(unwired) | 소비 struct/게이트는 존재하나 CLI/런타임에 배선 안 됨. 향후 배선 시 부활 | quality.TrustGate, harness.escalation, role_profiles.model | 은닉 또는 read-only (yaml 키는 향후 배선용으로 보존) |
| N3 프롬프트-소비(skill-body) | Go는 안 읽지만 skill body가 yaml-direct read로 소비 → **효과 있음** | git-strategy 다수 키 (types.go:53 "consumed by skill bodies via direct yaml reads, not by Go code") | **제거 금지**. 다이어트 각도는 중복 3-mode 노출 축소지 삭제 아님 |

이 구분은 §B 요구사항과 plan.md의 그룹별 결정 옵션에 반영된다.

---

## §B. GEARS 요구사항

> 표기: GEARS(현행). `<subject>`는 일반화된 명사(콘솔, 스키마, 함수 등). 각 REQ는 acceptance.md의
> grep/go-test 검증 AC와 1:1 대응한다.

### B.0 상위(Ubiquitous)

- **REQ-CD-001** (Ubiquitous): The moai web 콘솔은, 런타임 소비자가 존재하지 않아 편집이 효과 없음(N1)으로
  판정된 config 키를 **편집 가능 필드로 제시하지 않아야 한다(shall not)**.
- **REQ-CD-002** (Ubiquitous): The 정리 작업은 `internal/settings/schema.go`/`schema_sections.go`의
  렌더 SSOT를 유일한 변경 레버로 사용해야 한다(shall) — templ/TUI 표면에 개별 하드코딩 필터를 추가하지 않는다.
- **REQ-CD-003** (Ubiquitous): The 정리 작업은 skill-body가 yaml-direct read로 소비하는 키(N3)를
  삭제하지 않아야 한다(shall not).

### B.1 Tier 1 — 편집 가능하나 100% 무효과 (최우선, 신뢰 훼손)

- **REQ-CD-010** (Event-driven): **When** `settings.AllSections()`가 렌더 섹션을 열거할 때, the 스키마는
  research 섹션(`SectionResearch`)의 12개 seam 키를 편집 필드 집합에서 제외해야 한다(shall).
  근거: `internal/research` 패키지 삭제(SPEC-DEADPKG-INVESTIGATE-001), `ResearchConfig`는 round-trip
  전용(`.Research.` 필드 read 0건).
- **REQ-CD-011** (Event-driven): **When** db 섹션이 렌더될 때, the 스키마는 인터뷰 3키
  (`db.orm`/`db.multi_tenant`/`db.migration_tool`)를 편집 필드에서 제외하거나 read-only로 강등해야 한다(shall).
  근거: 채움 소스 `/moai db init` 2026-05-16 폐지, 소비자 부재.
- **REQ-CD-012** (Event-driven): **When** security 섹션이 렌더될 때, the 스키마는 편집 스칼라 3키
  (`security.permission.strict_mode`/`security.sandbox.required`/`security.sandbox.docker_image`)를
  편집 필드에서 제외하거나 read-only로 강등해야 한다(shall). 근거: config struct 필드가 동작 경로에서
  읽히지 않음(permission resolver의 `StrictMode`는 `doctor_permission.go:73`에서 하드코딩 `false`;
  sandbox opts는 config가 아닌 `SandboxOptions`에서 파생).

### B.2 Tier 2 — 대량 사강 (게이트/엔진 미배선)

- **REQ-CD-020** (Where): **Where** quality.yaml의 TrustGate 소비 경로가 CLI에 배선되지 않은 동안, the
  스키마는 `SectionQualityExtras`의 무효과 키를 은닉 또는 read-only로 강등해야 한다(shall). 라이브 유지
  대상은 `development_mode`, `test_coverage_target`(프롬프트 소비)뿐이다. 근거: `NewTrustGate` production
  호출 0건, `cfg.Quality`→TrustGate 매핑 부재.
- **REQ-CD-021** (Where): **Where** git-strategy 키가 skill-body yaml-direct read로 소비되는(N3) 동안, the
  스키마는 이들을 **삭제하지 않되(shall not delete)**, 활성 `mode`(manual/personal/team) 외 **비선택 mode
  프로파일의 중복 편집 노출을 축소해야 한다(shall reduce)**. 근거: 3-mode 프로파일이 전부 렌더되나 실제로는
  선택된 mode만 소비됨.
- **REQ-CD-022** (Event-driven): **When** agent-settings 섹션이 렌더될 때, the 스키마는 `workflow_agents`
  7-purpose × {model, effort} 14 필드를 편집 필드에서 제외하거나 read-only로 강등해야 한다(shall). 근거:
  `.WorkflowAgents` 맵의 production read 0건, `.claude/workflows/*.js`가 이 블록을 읽지 않음.
- **REQ-CD-023** (Event-driven): **When** ralph 섹션이 렌더될 때, the 스키마는 컴파일된 loop 엔진이 읽지
  않는 중첩 키(`ralph.ast_grep.*`/`ralph.loop.*`/`ralph.lsp.*`/`ralph.loop.completion.*`)를 은닉하거나
  프롬프트-doc 전용임을 명시하는 read-only로 강등해야 한다(shall). 라이브 유지 대상은 Go 소비되는
  `lint_as_instruction`/`warn_as_instruction`이다. 근거: `internal/ralph/engine.go`는 flat 플래그만 읽음.

### B.3 Tier 3 — 부분 사강

- **REQ-CD-030** (Event-driven): **When** workflow 섹션이 렌더될 때, the 스키마는 접근자에 호출자가 없는
  (dead-code) 키 `workflow.auto_clear.*`(4)/`workflow.token_budget.*`(3)/`workflow.loop_prevention.*`(2)/
  `workflow.worktree.auto_merge`(1)를 은닉 또는 read-only로 강등해야 한다(shall).
- **REQ-CD-031** (Event-driven): **When** harness 섹션이 렌더될 때, the 스키마는 읽히지 않는 키
  `harness.escalation.*`/`harness.mode_defaults.*`/`harness.auto_detection.enabled`를 은닉 또는
  read-only로 강등해야 한다(shall).
- **REQ-CD-032** (Event-driven): **When** observability 섹션이 렌더될 때, the 스키마는 Go 하드코딩 상수의
  미소비 미러 키 `observability.trace_dir`/`report_dir`/`max_file_size_mb`/`retention_days`/
  `hook_metrics.output_path`를 은닉 또는 read-only로 강등해야 한다(shall). 라이브 유지 대상은
  `enabled`/`hook_metrics.slow_hook_threshold_ms`이다.
- **REQ-CD-033** (Event-driven): **When** llm 섹션이 렌더될 때, the 스키마는 모델 dispatch가 읽지 않는
  편집 키 `llm.performance_tier`/`llm.claude_models.{high,medium,low}`를 은닉 또는 read-only로 강등해야
  한다(shall). 근거: Claude 모델은 launch/settings.json에서 해석됨; `PerformanceTier`는 oneof 검증에만 등장.
- **REQ-CD-034** (Event-driven): **When** agent-settings 섹션이 렌더될 때, the 스키마는 dispatch에 미배선된
  `role_profiles.*.model`/`role_profiles.*.mode`/`role_profiles.*.effort`(7 profiles × 유효 필드)를 은닉
  또는 read-only로 강등해야 한다(shall). **라이브인 `role_profiles.*.isolation`은 유지한다(`moai workflow
  lint`가 소비)**. 근거: `LoadRoleProfiles`(team_spawn.go)는 production 호출 0건.

### B.4 정합성 결함 수정 (Correctness Bugs)

- **REQ-CD-040** (Event-driven): **When** harness learning 보존 기간이 표시/집행될 때, the 시스템은 표시
  값과 실제 prune 값을 일치시켜야 한다(shall). 현재 `learning.log_retention_days`는 90(config 기본값)으로
  표시되나 실제 prune은 `internal/harness/observer.go:147`의 하드코딩 `defaultRetentionDays = 30`을 사용한다.
  수정 방향: (a) prune이 config 값을 읽도록 배선하거나, (b) 표시/문서를 30으로 정정 (plan.md 결정).
- **REQ-CD-041** (Ubiquitous): The `SecuritySandbox`의 `@MX:ANCHOR`/`@MX:REASON` 주석
  (`internal/config/types.go:429`, "fan_in >= 3: loaded by config/loader.go, consumed by
  sandbox/launcher.go, displayed by doctor_sandbox.go")은 실제 참조와 일치해야 한다(shall). 현재 config
  struct 필드(`Required`/`DockerImage`)의 동작 read가 0건이므로 주석은 stale다 — 주석을 실측에 맞게
  정정하거나 앵커를 강등한다.
- **REQ-CD-042** (Ubiquitous): The db `auto_sync` 스키마 서술은 실제 yaml 구조와 일치해야 한다(shall).
  현재 산문은 `db.auto_sync: true`(스칼라)로 서술하나 실제는 nested 블록(`schema_sections.go:413`의
  `RawViewBlocks`가 `db.auto_sync`를 nested block으로 렌더). 서술/렌더 형태를 단일화한다. (신뢰도 낮음 —
  research.md §R.4 참조; doc-level 정정.)
- **REQ-CD-043** (Ubiquitous): The `workflow.worktree.session_name_pattern` config 키는 실제 tmux 세션
  명명에 반영되거나(shall), 반영되지 않으면 콘솔에서 편집 불가로 강등되어야 한다(shall). 현재 config 필드
  (`types.go:411`, 기본값 `"moai-{ProjectName}-{SPEC-ID}"`)는 round-trip 전용이며 read 소비자가 없다 —
  세션 명명이 하드코딩됨.

### B.5 선택(SHOULD) — 과소 노출 라이브 키 표면화

- **REQ-CD-050** (Where, SHOULD): **Where** 라이브 소비가 **재검증으로 확인된** 경우, the 스키마는 현재
  은닉된 라이브 키(`workflow.team.auto_selection.min_*`, `llm.glm.base_url`)를 노출하는 것을 고려해야 한다
  (should). **주의**: 감사는 `team.auto_selection.min_*`를 "consumed"로 주장했으나, 본 SPEC의 재-grep은
  `internal/config/defaults.go`에서만 기본값 설정을 발견하고 **production read 소비자를 확인하지 못했다**
  (research.md §R.5). 따라서 본 REQ의 각 키 노출은 **소비자 재확인을 전제조건으로** 한다 — 확인 실패 시
  해당 키는 노출하지 않는다. 이 REQ는 부차적(secondary)이며 M1~M4의 핵심 다이어트를 차단하지 않는다.

---

## §C. Tier 분류

**Tier: M (standard)**

근거:
- 범위: 렌더 SSOT 2파일(`schema.go`/`schema_sections.go`) 중심 + 정합성 결함 수정 시 `internal/harness/`,
  `internal/config/` 소폭 접촉. 약 10개 config 그룹 강등/제거 + 4개 결함.
- 위험: 두 표면(templ/TUI)이 공유하는 스키마 데이터 변경이므로 회귀 표면이 존재하나, 동작 변경이 아닌
  **필드 노출 축소**라 런타임 로직 위험은 낮음. N3(skill-body 소비) 오삭제만 피하면 안전.
- 복잡도: 결정 옵션(은닉/read-only/제거)이 그룹마다 다르므로 plan-phase 결정 밀도는 있으나 구현 자체는
  데이터 편집 위주. S(minimal)로 보기엔 그룹 수와 결함 수가 많고, L(thorough)로 보기엔 신규 로직/아키텍처
  변경이 없다 → **M** 적정.

LSP/harness 게이트: standard(기본 checks). run-phase에서 zero errors/type-errors/lint-errors 필요.

---

## §D. 범위 밖 (Out of Scope / Exclusions)

본 SPEC은 노출 축소(diet)에 한정된다. 아래는 **out of scope**이며 별도 후속 SPEC 소관이다.

### Out of Scope — 미배선 소비자 배선(wiring)

- `core/quality.TrustGate`를 CLI/게이트 파이프라인에 배선하는 것 (quality.yaml 키 부활). 이는 기능 작업이다.
- `harness.escalation.*` → `EscalationManager`의 production 배선.
- `workflow_agents` 블록을 `.claude/workflows/*.js` 런타임이 읽도록 하는 config-default+override 병합점 추가.
- `role_profiles.*.model`/`.mode`/`.effort`를 팀 spawn dispatch에 배선하는 것.
- ralph 중첩 키(`ast_grep.*`/`loop.*`/`lsp.*`)를 컴파일 loop 엔진이 읽도록 배선하는 것.

### Out of Scope — research.yaml/ResearchConfig 완전 삭제 여부의 선(先)결정

- `internal/research` 패키지는 이미 삭제되었고 `ResearchConfig`는 round-trip 전용이다. 콘솔 노출 제거는
  IN scope이나, **`ResearchConfig` struct + yaml 키 자체를 config 스키마에서 완전 삭제**하는 것은 별도
  판단이 필요한 큰 결정이며 plan.md §F에서 옵션으로 제시하되, 본 SPEC 구현 필수 범위에는 넣지 않는다
  (out of scope for the mandatory milestone set; SPEC-DEADPKG 선례 참조).

### Out of Scope — 콘솔 UI/렌더링 레이어 재설계

- templ 템플릿(`internal/web/*_templ.go`)·TUI 위저드(`internal/cli/profile_setup*.go`)의 레이아웃/UX
  재설계는 범위 밖. 본 SPEC은 스키마 데이터(FieldDef 집합)만 변경하고 렌더 레이어는 그 데이터를 그대로 소비한다.

### Out of Scope — i18n 키 대량 정리

- dead 필드 제거로 고아가 되는 i18n 키(`f.<name>` prefix)의 대량 정리는 부수 효과로 발생할 수 있으나,
  본 SPEC의 검증은 "필드가 렌더 집합에서 빠졌는가"이지 "고아 i18n 키가 0인가"가 아니다. i18n 파리티 전수
  정리는 후속 위생 작업으로 남긴다.

---

## §E. 검증 요약 (Self-Verification pointer)

전체 AC와 grep/go-test 검증 술어는 acceptance.md §D를 SSOT로 한다. 각 REQ-CD-0xx는 acceptance.md의
AC-CD-0xx와 1:1 대응하며, 대부분은 `grep -c`(필드 부재) 또는 `go test`(스키마 카운트) assertion으로
기계 검증 가능하다.

---

## §F. 교차 참조 (Cross-References)

- research.md — 감사 증거 file:line 인용 (본 SPEC의 사실 근거).
- plan.md — 마일스톤(Tier 1/2/3 + 결함) + 그룹별 결정 옵션(은닉/read-only/제거).
- acceptance.md — grep/go-test 검증 AC.
- SPEC-WEB-CONSOLE-011 — 노출을 만든 M2 재설계(선행 SPEC).
- SPEC-DEADPKG-INVESTIGATE-001 — research 패키지 삭제 선례.
- `internal/settings/schema.go`, `internal/settings/schema_sections.go` — 렌더 SSOT.
