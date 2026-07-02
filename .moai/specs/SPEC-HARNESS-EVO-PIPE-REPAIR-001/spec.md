---
id: SPEC-HARNESS-EVO-PIPE-REPAIR-001
title: "하네스 학습 파이프라인 수리 + v4 스모크 게이트 (Epic Harness-Evolution 1/4)"
version: "0.1.0"
status: draft
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, learning-loop, smoke-gate, hooks, proposalgen, template-first"
related_specs: [SPEC-V3R6-HARNESS-V4-001, SPEC-V3R6-HARNESS-PROPOSAL-GEN-001, SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001, SPEC-HARNESS-LOOP-CLOSURE-001]
---

# SPEC-HARNESS-EVO-PIPE-REPAIR-001: 하네스 학습 파이프라인 수리 + v4 스모크 게이트

## HISTORY

| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|-----------|--------|
| 0.1.0 | 2026-07-02 | 최초 작성 — Epic Harness-Evolution entry SPEC (파이프 수리 + 스모크 게이트) | manager-spec |

---

## §A 배경

사용자 요청: `/moai:harness` 재설계 — R1) 요구사항 유도(Context-First Discovery), R2) 커스텀 슬래시 커맨드 전달, R3) 커맨드 호출 시 워크플로우 end-to-end 실행, R4) 실행 중 발견된 개선 필요가 재귀적 자기개선(진화/학습) 루프로 하네스 자체를 갱신.

8-표면 분석 결과: R1/R2는 v4 Builder가 대부분 구현, R3는 관례 의존 + live E2E 검증 0건, **R4는 6개 끊어진 링크로 구조적 0 커버리지**. 본 SPEC은 Epic Harness-Evolution(4 SPECs)의 1번으로, R4 복원의 전제인 **파이프 수리**(분류→제안 어휘/스키마/자동화)와 **v4 스모크 게이트**(R3의 참조 무결성 기계 검증)를 담당한다. Epic 전체 구조는 `plan.md §A.1` 참조.

관측 베이스라인 (2026-07-02 plan-phase 실측):

- `.moai/harness/usage-log.jsonl`: **608 events** (`wc -l` 실측)
- `.moai/harness/learning-history/tier-promotions.jsonl`: 16 promotion 레코드 존재 (전량 Read 실측)
- `.moai/harness/learning-history/applied/`: **0건** (디렉터리 부재 — 적용 이력 없음)
- `.moai/harness/proposals/`: 부재 (proposal 생성 이력 없음)

즉 관측→분류는 살아 있으나(608 events, promotions 생성됨) **제안 생성이 구조적으로 0**이다.

---

## §B 문제 정의 (검증 앵커 — 2026-07-02 전량 재검증)

> 아래 앵커는 본 plan-phase에서 파일을 직접 Read/Grep하여 재검증했다 (verification-claim-integrity §2 baseline 귀속). 라인 번호는 참고용이며 run-phase에서는 content-token 앵커로 재고정할 것.

### B1. classify→propose 어휘 불일치 (제안 생성 구조적 0)

- `internal/harness/types.go:217-245` — `Tier.String()`이 방출하는 어휘: `{observation, heuristic, rule, auto_update}`
- `internal/harness/proposalgen/mapper.go:41-44` — `actionableTiers = {recommendation, approval_required}` — **교집합 없음**
- 실데이터 증거 (tier-promotions.jsonl 실측): `"to_tier":"auto_update"` (observation_count 206), `"to_tier":"observation"` — `recommendation`/`approval_required`는 실데이터에 존재하지 않음
- mapper.go:38-40 주석 자체가 혼동 상태: "pre-actionable tiers (observation, auto_update)"라며 4-tier 어휘 중 2개만 언급하면서 actionable 집합은 별개 어휘를 사용

### B2. pattern_key 스키마 불일치 (regex 전량 거부)

- `internal/harness/proposalgen/mapper.go:36` — `actionablePatternRE = ^(code_change|error_pattern|tool_failure|repeated_edit):[a-z_]+:[^:]+$`
- `internal/harness/learner.go:98-99` — `buildPatternKey`는 `<event_type>:<subject>:<context_hash>` 형식으로 생성; `internal/harness/types.go:33-63` EventType enum = `{moai_subcommand, agent_invocation, spec_reference, feedback, session_stop, subagent_stop, user_prompt, apply_outcome}` — regex의 4개 prefix를 방출하는 observer 경로는 **존재하지 않음**
- 실데이터 증거: `"pattern_key":"user_prompt::"`, `"agent_invocation:Bash:"`, `"session_stop::"` — (a) prefix 불일치, (b) subject가 빈 문자열인 경우 존재(`session_stop::`), (c) 말단 context_hash가 빈 경우 존재(`agent_invocation:Bash:`) → regex의 `[a-z_]+:[^:]+$`는 세 사유 모두로 거부

### B3. classify 수동 실행 + generic observer 미등록

- `internal/cli/hook.go:144-162` — `moai hook harness-classify` subcommand 존재
- 유일 호출 표면: `.claude/skills/moai/workflows/harness.md:135-136` — `/moai:harness status` step 0에서만 수동 실행. Stop hook 경로에 미배선
- `internal/cli/hook.go:96-122` — `harness-observe`(generic) + observe-stop/subagent-stop/user-prompt-submit subcommand 존재; wrapper `handle-harness-observe.sh` 디스크 존재
- `.claude/settings.json` (grep 실측): `handle-harness-observe-{stop,subagent-stop,user-prompt-submit}.sh` 3종만 등록 — **generic PostToolUse observer 미등록**. `internal/template/templates/.claude/settings.json.tmpl` 동일

### B4. v4 스모크 게이트 부재

- `internal/harness/v4manifest/validate.go` — manifest 단일 파일 검증만 존재
- `internal/cli/harness/v4lifecycle.go` — list/edit/remove만 존재 (command↔manifest 조인 스캔 로직은 보유)
- `internal/cli/doctor.go:194` + `doctor_harness.go` — 레거시 `moai doctor`의 "Harness 5-Layer" 진단은 learning-loop 대상이며 v4 command↔manifest↔Runner↔agent 참조 무결성 검사는 **어느 경로에도 없음**

### B5. 예시(exemplar) Runner 부패 — 스모크 게이트 1호 검증 대상

- `.claude/workflows/harness-release-update-run.js:31` — `MANIFEST_PATH = ".claude/commands/harness/manifest.json"` — 실제 파일은 `.claude/commands/harness/release-update/manifest.json` (find 실측): **경로 오류**
- 동 파일 :56 — `harness-devkit-release-update-specialist` 참조 — 실존 agent는 `.claude/agents/harness/harness-release-update-specialist.md` (ls 실측): **비실존 agent 이름**
- (부수) 동 파일 :1 파일 헤더가 `harness-devkit-run.js`로 구 명칭 잔존

### B6. 디스패처 드리프트 (reserved-verb 가드 / help / manifest 위치 3-way 분기)

- `.claude/skills/moai/SKILL.md:195-216` — 디스패처 예약 동사 = Branch A `{status, apply, rollback, disable}` + Branch A.1 `{list, edit, remove}` (7개)
- `.claude/skills/moai/workflows/harness-build-entry.md:53-55` — Phase 0 가드는 4개(`status/apply/rollback/disable`)만 재검사 → `list/edit/remove`가 직접 호출 시 자연어 빌드 파이프라인으로 누수
- `.claude/commands/moai/harness.md:3` — argument-hint가 4-verb 시대에 고착; `.claude/skills/moai/SKILL.md:70` 라우터 테이블 행도 동일
- manifest 위치 3-way 분기: (i) `.claude/skills/moai/workflows/harness-builder.md:257` — "`.claude/commands/harness/manifest.json` OR `.claude/harness/<name>/manifest.json`" 이중 서술, (ii) `internal/cli/harness/v4lifecycle.go:33-34` — `.claude/commands/harness/<name>/manifest.json` (Go norm), (iii) B5 Runner 상수 — `<name>` 서브디렉터리 없는 경로

### B7. 문서-코드 모순 (FROZEN / rate-limit SSOT)

- FROZEN: `.claude/skills/moai/workflows/harness.md` §2.2 apply Layer 1 + `.claude/skills/moai-harness-learner/SKILL.md` L1 표(~:131-140)가 `.claude/agents/{moai,harness}/`를 FROZEN으로 선언 — 그러나 `internal/harness/frozen_guard.go:21-27` `allowedPrefixes`는 `.claude/agents/harness/`를 **허용**, `:32-37` `frozenPrefixes`는 `.claude/agents/moai/`만 FROZEN
- rate-limit: `.claude/skills/moai/workflows/harness.md:91` — "7일당 1회 floor는 REQ-HRN-FND-018 invariant" + `:173` Layer 4 pre-screen "last 7 days count ≥ 1 defer" — 그러나 `.moai/config/sections/harness.yaml:119-121` = `max_per_week: 3` + `cooldown_hours: 24`, `internal/cli/harness.go:486` Go 기본값 동일(`MaxPerWeek: 3, CooldownHours: 24`), learner SKILL.md L4 행도 "max 3 per week, 24h cooldown" — **harness.md만 고립된 구 규범**

---

## §C GEARS 요구사항

### 파이프 수리 (Go 엔진)

**REQ-HEP-001 (어휘 정렬)** — The proposalgen mapper의 actionable tier 판정 어휘는 classifier가 방출하는 `Tier.String()` 어휘(SSOT)에서 파생되어야 한다(shall). **When** `ToTier ∈ {rule, auto_update}`이고 `Confidence ≥ ConfidenceThreshold(0.70)`이며 pattern_key가 스키마 유효한 promotion이 입력되면, the mapper shall 해당 promotion을 proposal 후보로 채택해야 한다. The mapper shall not `{observation, heuristic}` tier의 promotion을 후보로 채택해서는 안 된다. (채택 방향 결정 근거: plan.md §D-D1)

**REQ-HEP-002 (pattern_key 스키마 정렬)** — The mapper의 pattern_key 수용 스키마는 learner `buildPatternKey`가 실제로 방출하는 스키마(`<event_type>:<subject>:<context_hash>`, event_type ∈ EventType enum, subject/context_hash는 빈 문자열 허용)와 단일 SSOT로 정렬되어야 한다(shall). The mapper shall not 수기 유지되는 병렬 prefix 목록(`{code_change, error_pattern, tool_failure, repeated_edit}`)을 근거로 실데이터 promotion을 거부해서는 안 된다. **When** 실측 형태의 promotion(예: `user_prompt::` + `to_tier: auto_update` + confidence 1.0)이 입력되면, the pipeline shall 스키마 사유만으로 이를 걸러내지 않아야 한다.

### 자동화 배선 (hooks / settings)

**REQ-HEP-003 (classify 자동 실행)** — **When** 세션 Stop hook 경로가 실행되면, the harness pipeline shall `harness-classify` 상당의 분류를 자동 수행해야 한다 (수동 `/moai:harness status` step 0 의존 제거). **While** 분류가 실패하거나 훅 시간 예산(5s 정책)을 초과하면, the hook shall fail-open으로 동작하고 세션 종료를 차단하지 않아야 한다(shall not block). (배선 방식 후보와 권고는 plan.md §D-D3)

**REQ-HEP-004 (PostToolUse observer 등록)** — **Where** generic PostToolUse observer(`handle-harness-observe.sh` + `moai hook harness-observe`)가 존재하면, the settings deployment shall 해당 observer를 settings 훅 설정에 등록해야 한다 — template(`internal/template/templates/.claude/settings.json.tmpl`)과 로컬 `.claude/settings.json` 양쪽 (Template-First).

### 스모크 게이트 (v4 참조 무결성)

**REQ-HEP-005 (moai harness doctor)** — The moai CLI shall v4 하네스에 대한 참조 무결성 검사(`moai harness doctor` 또는 동등 verb)를 제공해야 한다. **When** 검사가 실행되면, the checker shall 각 v4 하네스에 대해 최소 다음 4축 cross-reference를 검증해야 한다:
1. entry command(`.claude/commands/harness/<name>.md`) 존재 + Runner/manifest 참조 해석 가능
2. manifest(`.claude/commands/harness/<name>/manifest.json`) 존재 + 스키마 유효(`v4manifest` validate 재사용)
3. Runner(`.claude/workflows/harness-<name>-run.js`) 존재 + Runner 내 manifest 경로 상수가 실존 파일로 해석
4. manifest/Runner가 참조하는 specialist agent 이름이 `.claude/agents/harness/*.md` 실존 파일로 해석

**When** 하네스가 0개인 프로젝트에서 실행되면, the checker shall 정상 종료(exit 0)해야 한다.

**REQ-HEP-006 (Builder ACTIVATE 계약)** — **When** Builder ACTIVATE 단계가 산출물 생성을 완료하면, the Builder workflow contract(harness-builder.md ACTIVATE 절) shall 스모크 게이트 실행을 요구해야 한다. **When** 게이트가 결함을 보고하면, the workflow shall 해당 하네스를 활성(active)으로 선언하지 않아야 한다(shall not).

**REQ-HEP-007 (exemplar 부패 수리 + 게이트 회귀 증명)** — The release-update Runner(`.claude/workflows/harness-release-update-run.js`, dev-only 로컬 전용)는 실존 manifest 경로(`.claude/commands/harness/release-update/manifest.json`)와 실존 specialist agent 이름(`harness-release-update-specialist`)을 참조해야 한다(shall). The smoke gate shall 수리 전 상태의 두 결함(잘못된 MANIFEST_PATH + 비실존 agent 이름)을 탐지할 수 있어야 한다 — 이 결함 클래스가 게이트의 1호 검증 대상이다.

### 디스패처 / 문서 정합

**REQ-HEP-008 (reserved-verb 가드 완전화)** — **When** `harness-build-entry.md` Phase 0 가드가 `$ARGUMENTS` 첫 토큰을 검사하면, the guard shall 디스패처 SSOT의 전체 예약 동사 집합(`status/apply/rollback/disable` + `list/edit/remove` + 본 SPEC이 추가하는 `doctor`)을 검사해야 한다.

**REQ-HEP-009 (help 표면 현행화)** — The `/moai:harness` help 표면(`.claude/commands/moai/harness.md` argument-hint, `.claude/skills/moai/SKILL.md` 라우터 테이블 행 및 §harness 디스패처 절) shall 현행 동사 집합 전체(`doctor` 포함)를 열거해야 한다.

**REQ-HEP-010 (manifest 위치 단일 정본)** — The 하네스 문서 표면 shall manifest 위치를 단일 정본 `.claude/commands/harness/<name>/manifest.json`(Go `v4lifecycle` norm)으로 선언해야 한다. The `harness-builder.md`의 대안 위치 서술(`.claude/harness/<name>/manifest.json` OR-분기) shall not 잔존해서는 안 된다.

**REQ-HEP-011 (FROZEN 문서-코드 정렬)** — The 스킬/워크플로우 문서(`harness.md` Layer 1 목록, learner `SKILL.md` L1 목록) shall `frozen_guard.go` norm과 정렬되어야 한다: `.claude/agents/harness/`는 허용 쓰기 대상(`allowedPrefixes`), `.claude/agents/moai/`는 FROZEN(`frozenPrefixes`). 문서 shall not `.claude/agents/harness/`를 FROZEN으로 선언해서는 안 된다.

**REQ-HEP-012 (rate-limit SSOT 정정)** — The `harness.md` 워크플로우의 rate-limit 서술 shall 단일 SSOT(`harness.yaml` + Go 기본값: `max_per_week: 3`, `cooldown_hours: 24`)를 따라야 한다. "7일당 1회 invariant" 문구(:91, :173)는 정정하고, REQ-HRN-FND-018 provenance note(구 floor 주장이 본 SPEC에 의해 supersede됨)를 남겨야 한다(shall).

### 횡단 제약

**REQ-HEP-013 (Template-First + 중립성)** — **Where** 본 SPEC이 수정하는 표면이 template-managed이면(mirror 실측 확인: `harness.md`, `harness-build-entry.md`, `harness-builder.md`, learner `SKILL.md`, moai `SKILL.md`, `commands/moai/harness.md`, `harness.yaml`, `settings.json.tmpl`), the run-phase shall template mirror를 함께 갱신하고 `make build`를 실행해야 한다. The template content shall not 본 SPEC의 ID / REQ-HEP 토큰 / 감사 인용을 포함해서는 안 된다(§25 neutrality). The dev-only Runner(`harness-release-update-run.js`)와 생성물 `harness-*` artifacts shall not template로 유출되어서는 안 된다 (CI guard: `TestSplitHarnessNamespaceNoLeak`, `template-neutrality-check`).

---

## §D 수용 기준

AC 매트릭스는 `acceptance.md` §D 참조 (AC-HEP-001 ~ AC-HEP-015, REQ↔AC 추적표 포함).

---

## §E 제외 범위 (Exclusions)

본 SPEC은 Epic Harness-Evolution의 1/4이며, 아래 항목은 명시적으로 범위 밖이다.

### Out of Scope — frozen_guard allowlist 확장 (SPEC-3)

- `frozen_guard.go` `allowedPrefixes`에 `.claude/commands/harness/`, `.claude/workflows/harness-*.js`, specialists를 추가하는 단계적 write-surface 개방 (M1 manifest-first → M2 full)
- snapshot + rollback + regression-gate 의무화
- → `SPEC-HARNESS-EVO-WRITE-SURFACE-001`

### Out of Scope — 실행→학습 배선 (SPEC-2)

- manifest `learning` 블록, Runner return-schema `findings` 필드, specialist 필수 "improvement findings" 최종 단계
- 오케스트레이터 post-run 단계 표준화 (findings 수집 → 즉시 AskUserQuestion push; 현행 pull-only apply 대체)
- `learner.go` confidence 하드코딩 1.0의 outcome 기반 실측화 (실데이터 전 레코드 `confidence: 1` 확인됨 — 본 SPEC은 수리하지 않음)
- → `SPEC-HARNESS-EVO-RUN-REPORT-001`

### Out of Scope — 헌법적 제약 개정 (SPEC-3)

- LOOP-CLOSURE C1 제약(`auto_apply: false` per-item 승인)의 티어별 표면 자율(tiered per-surface autonomy)로의 supersede/amendment
- `.moai/docs/harness-namespace-doctrine.md`의 "legitimate learning-loop writes" 조항 추가
- → `SPEC-HARNESS-EVO-WRITE-SURFACE-001`

### Out of Scope — 요구사항 아티팩트 스키마 + 레거시 retire (SPEC-4)

- manifest `source_request`(단일 raw string)의 구조화 요구사항 스키마(domain/goal/constraints/scope + AC + Discovery 응답 기록) 승격 + drift 감지 시 재-Discovery
- 레거시 5-layer marker 경로(`/moai project` meta-harness route, archived agent 이름을 방출하는 layer5 scaffold) retire
- `harness-delivery-strategy.md`의 Model B rejection에 대한 supersede 선언
- → `SPEC-HARNESS-EVO-REQ-ARTIFACT-001`

### Out of Scope — 학습 루프 알고리즘 개선 일반

- tier 승급 임계값(`{1,3,5,10}`) 조정, 새 EventType 추가, 관측 커버리지 확대 등 파이프 "수리"를 넘는 개선은 본 SPEC 범위 밖 (Epic 후속 또는 별도 SPEC)

---

## §F 가정 및 미검증 항목 (verification-claim-integrity)

- **감사 보고 수치 "258 observations / 536 events / 0 applies"**: 오케스트레이터 제공 감사 수치이며 본 plan-phase에서 세부 분해를 재검증하지 않았다. 본 plan-phase 실측(2026-07-02): usage-log 608 lines / applied 0건 / proposals 부재 — 방향성 일치. run-phase M1에서 baseline 재측정 의무.
- **"legacy path had `moai doctor harness`"**: 정확히는 `moai doctor`의 진단 항목 "Harness 5-Layer"(`doctor.go:194`, `doctor_harness.go`)로 실재 확인 — 별도 `moai doctor harness` subcommand 형태는 아니며, 대상도 learning-loop(5-layer)이지 v4 참조 무결성이 아니다.
- **Promotion.ToTier의 기록 주체**: `Tier.String()` 어휘가 실데이터(tier-promotions.jsonl)에 그대로 나타남을 확인했으나, classify 경로의 writer 함수 자체는 라인 단위로 재확인하지 않았다. run-phase M1에서 writer 앵커 재고정.
- **라인 번호 드리프트**: 본 문서의 모든 `파일:라인` 앵커는 2026-07-02 시점 실측이며, run-phase에서는 content-token(식별자/문자열 리터럴) 기준으로 재고정한다.
