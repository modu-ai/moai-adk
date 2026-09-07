---
id: SPEC-WEB-WRITE-SAFETY-001
title: "moai web 쓰기 안전성 — 무저장 재기록 차단, 쓰기 범위 한정, 포맷 충실도 보존"
version: "0.1.1"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/web, internal/config"
lifecycle: spec-anchored
tags: "web, config, write-safety, dirty-gate, yaml-patch, form-parser, reproduction-first, defect"
tier: M
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-WEB-CONSOLE-010, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-FEEDBACK-AUTO-SUBMIT-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-09-07 | manager-spec | plan-audit iter-1 (FAIL 0.88) 정정. **D1(차단)** — REQ-WWS-008 신설로 측정-선결 의무를 요구 계층에 앵커(§5→acceptance→§4 순환 참조 해소), §D/§D.3 매핑 갱신. **D2** — plan.md M4 무조건 재기록 exemplar를 5섹션으로 정정(manager.go:187/192/197/202/224) + M2에 6종 판별 증거 항목 추가 + spec.md C3를 5종 일반화로 정정. **D3** — AC-WWS-002 GET 열거에 `/static/` 추가 + §D.2에 glmkey reveal·static edge 추가. **D4** — REQ-WWS-001 허용 예외를 닫힌 4라우트 목록 + 사용자 개시 판별 기준으로 교체. |
| 0.1.0 | 2026-09-07 | manager-spec | 최초 draft. 카드 t517 (Class B — 결함, 원인 미특정, 조사-선결 구조). 리드 실측 관측 2건 + 미귀속 관측 1건을 전제로, 수리(M4)를 첫 측정 2건(M-a 쓰기 경로 귀속, M-b gate 우회·seam 충실도 원인) 뒤에 배치. 코드 근거 9곳 plan-phase 직접 확인(트리 `0b1e27877`). |

---

## §1 Context & Motivation

### §1.1 관측된 결함 (리드 실측 — ground truth로 취급, 본 세션 재측정은 아님)

2026-09-07, 관측 워크트리에서 `moai web`을 기동하고 Save를 누르지 않았는데도 git 추적 중인 설정 파일 2개가 재기록됐다:

| 파일 | 관측된 변경 | 성격 |
|------|------------|------|
| `.moai/config/sections/feedback.yaml` | 빈 줄 1행 삭제 | seam 섹션(`RouteSeam`, `internal/settings/sectionroute.go:99`) — yamlpatch 경로가 보존해야 할 표현 요소 |
| `.moai/config/sections/git-strategy.yaml` | 키 순서 변경 + 부재하던 `develop_branch: ""` 키 신규 추가 | dirty-gate가 적용된 typed 섹션 (`internal/config/manager.go:206-221`) |

귀속: 두 파일 모두 레인 워크트리에서 기록됐다(mtime 2026-09-07 13:45:57 동시), 이후 explicit-path `git restore`로 복구, `git status` clean 확인. primary checkout은 무접촉(두 파일 mtime 2026-09-03 23:06:02, 세션 이전부터 dirty — 출처 불명 변경 보유).

형제 관측(미귀속): primary checkout의 `.moai/config/sections/llm.yaml`이 오늘 기록됐다(`find -newermt` 유일 적중). 기록 주체 미특정. 코드상 llm.yaml은 typed 경로로 `Save()` 때마다 무조건 재기록된다(`internal/config/manager.go:223-226`) — 그럴듯한 기제이나 측정된 귀속이 아니다. 본 SPEC은 이 귀속을 1급 조사 항목(M-a)으로 격상한다.

### §1.2 핵심 모순 — 이 카드의 중심

`git-strategy.yaml`은 SPEC-GITSTRATEGY-SAVE-ISOLATION-001이 확립한 dirty-gate로 보호된다: 세션 내 `SetSection("git_strategy", ...)` 수정(`gitStrategyDirty` 플래그) 또는 파일 부재(greenfield)가 없으면 byte-단위로 보존된다(`internal/config/manager.go:206-221`, REQ-GSI-001/002). 성공적인 `Save()` 이후 플래그는 리셋된다(`manager.go:230`). 그런데 이 파일이 변경됐다.

**보호된다고 믿었던 파일이 변경된 것**이 이 카드의 핵심이며, 원인은 미특정이다 — 저장 흐름 어딘가가 `SetSection(git_strategy)`을 조용히 불러 dirty 플래그를 세팅했을 수도, gate 밖의 별도 쓰기 경로가 있을 수도 있다. 원인 규명이 수리에 선행한다(Class B). 이것이 M-b 게이트 측정이다.

### §1.3 코드 근거 (plan-phase 직접 확인, 트리 `0b1e27877`, 2026-09-07)

| # | 근거 | 위치 | 확인 방법 |
|---|------|------|----------|
| C1 | `ConfigManager.Save()`는 6종(user/language/quality/git-convention/git-strategy/llm)을 기록하며 **feedback.yaml은 저장 목록에 없다** — feedback.yaml의 기록 주체는 `Save()` 밖의 별도 경로(yamlpatch seam 경로 추정)다 | `internal/config/manager.go:171-233` | 직접 판독 |
| C2 | git-strategy.yaml dirty-gate: `gitStrategyDirty \|\| absent`일 때만 재기록 | `internal/config/manager.go:206-221, 230` | 직접 판독 |
| C3 | `Save()`의 무조건 재기록 대상은 5종 — user.yaml(`:187`), language.yaml(`:192`), quality.yaml(`:197`), git-convention.yaml(`:202`), llm.yaml(`:224`); gated는 git-strategy(C2)뿐이다 | `internal/config/manager.go:171-233` | 직접 판독 |
| C4 | `handleSave`(POST /save)는 쓰기를 9개 seam에서 순차 수행(writePreferences → recordLastProfile(advisory) → syncToProject → writeProjectConfig → writeProjectNestedConfig → applySchemaEdits → applyPerfTierEdits → patchAgentFM → glmcred.Save)하며, 후행 단계 실패 시 선행 쓰기를 롤백하지 않는다(에러 배너로 문서화된 동작). 단일 파일 쓰기는 temp+rename 원자적 | `internal/web/handlers.go:350-558` | 직접 판독 |
| C5 | 편집 가능 표면은 typed 2종(git-strategy, llm) + seam 전용 6종; feedback은 `RouteSeam`으로 seam-writable 재개됨 | `internal/web/projectconfig.go:160-164`, `internal/settings/sectionroute.go:92-99` | 직접 판독 |
| C6 | `r.PostFormValue`/`r.PostForm[f.Name]`은 동일 name 다중 제출 시 첫 값만 반환 — 파서는 중복을 인지하지 못한다 | `internal/web/schemaform.go:320-352` | 직접 판독 |
| C7 | 서버 기동 경로(`Run`/`NewServer`)와 CLI 커맨드(`internal/cli/web.go`)에는 config 쓰기 호출이 관측되지 않는다 | `internal/web/server.go:94-141` + `internal/cli/web.go` grep 무매치 | 직접 판독 |
| C8 | 라우트 테이블: GET `/`·`/kanban`·`/monitor`·`/todo`·`/settings`·`/specs`·`/events`, POST `/save`·`/__shutdown__`, `/profile/{create,delete,rename}`, glmkey reveal, `/static/` | `internal/web/app.go:157-197` | 직접 판독 |
| O1 | §1.1의 관측 사실 자체(파일 2건 변경 + 귀속 + 복구 경과) | — | 리드 실측(전달 근거) |
| O2 | primary checkout llm.yaml 오늘 기록(귀속 미완) | — | 리드 실측(전달 근거) |

모든 file:line 앵커는 트리 `0b1e27877`(2026-09-07) 실측값이다. run-phase 착수 시 content-token 기준 재검증 의무가 있다(plan.md §C).

### §1.4 왜 Save 없이 썼는가 — 미해결 질문 (M-a)

C7(기동 경로 무쓰기)과 C4(POST /save 게이트)를 고려하면, 무저장 재기록의 후보 기제는 최소한 다음과 같고 어느 것도 아직 배제되지 않았다:

1. 페이지 렌더·htmx 트리거가 개시하는 POST(예: `hx-trigger="load"`류 자동 제출)가 쓰기 라우트에 도달
2. GET/탐색 라우트 핸들러 내부의 lazy 쓰기(정규화·마이그레이션·센티널 갱신 등)
3. `moai web` CLI 래퍼 또는 그것이 호출하는 헬퍼의 기동 시 config 접촉
4. 관측 세션 외부의 동시 작성자

M-a가 이 중 실제 경로를 특정한다. 본 SPEC은 어느 가설도 단정하지 않는다. 가설 4로 판명되면 수리 대상이 이 카드의 코드가 아니므로 blocker report로 회신한다(plan.md §F M2).

---

## §2 Requirements (GEARS)

**사용자 스토리**: "나(moai-adk 개발자)가 `moai web`을 열어 설정을 구경만 할 때는 내 프로젝트의 추적 파일이 조용히 바뀌지 않는다. Save를 눌렀을 때는 내가 편집한 것만 정확히 바뀌고, 편집하지 않은 파일의 포맷(키 순서·빈 줄·주석)은 그대로다."

### §2.1 쓰기 시점 (write timing)

**REQ-WWS-001 (Ubiquitous):** The `moai web` console shall not write any file under `.moai/config/` to disk in the absence of a user-initiated explicit write action. 허용 예외는 닫힌 목록으로 한정한다 — POST `/save` 제출, `/profile/create` 제출, `/profile/delete` 제출, `/profile/rename` 제출(라우트 테이블 C8의 변경 라우트 중 쓰기 동작과 결부된 4개; `/__shutdown__`·glmkey reveal은 config를 기록하지 않으므로 예외가 아니다). **판별 기준**: 콘솔 자신의 렌더·폴링·htmx 자동 트리거(`hx-trigger="load"` 등)가 사용자 개입 없이 발화하는 요청은, 대상이 변경 라우트여도 사용자 개시로 인정하지 않는다.

**REQ-WWS-002 (When):** When 사용자가 `moai web`을 기동하고 페이지를 탐색만 하면(Save 제출 없음, 종료 포함), the web console shall 기동·렌더·폴링·종료 전 과정에서 `.moai/config/` 하위의 git 추적 파일을 내용상 byte-동일하게 유지한다.

### §2.2 쓰기 범위 (write scope)

**REQ-WWS-003 (When):** When 사용자가 Save를 제출하면, the web console shall 제출이 실제로 값을 바꾸는 섹션 파일만 기록하고, 값이 바뀌지 않은 섹션 파일은 git 추적 기준 diff 없음(내용 byte-동일)을 유지한다.

**REQ-WWS-004 (Ubiquitous):** No web console code path shall bypass the SPEC-GITSTRATEGY-SAVE-ISOLATION-001 dirty-gate contract on `git-strategy.yaml` (rewrite only upon an in-session `git_strategy` section modification or file absence — 세션 내 섹션 수정 또는 파일 부재에 의해서만 재기록).

### §2.3 포맷 충실도 (formatting fidelity)

**REQ-WWS-005 (Ubiquitous):** A web console save shall preserve the untouched presentation elements of the target section file — key order, blank lines, comments, and keys not modeled by the Go struct (편집되지 않은 표현 요소 보존).

### §2.4 폼 파서 (duplicate-form-value awareness)

**REQ-WWS-006 (When):** When 하나의 POST에 동일 `name`을 가진 폼 필드가 둘 이상 제출되면, the form parser shall 그 중복을 조용히 첫 값으로 해석하지 않는다 — 중복을 감지해 atomic reject에 합류시키거나, 문서화된 중복 해석 규칙을 적용한다.

### §2.5 회귀 가드

**REQ-WWS-007 (Ubiquitous):** The web write paths shall be covered by regression tests that detect the "config written without an explicit save" defect. Each absence-guard test shall be demonstrated RED on the defective code (RED-first) and shall be shown to catch a defect-reintroducing mutant (mutant verification). 셀 구조와 4요소(커맨드·출력·exit code·트리 SHA)는 acceptance.md §D와 `verification-completeness.md` §2를 따른다.

### §2.6 측정-선결 의무 (Class B 게이트)

**REQ-WWS-008 (When):** When a repair to the web write paths (plan.md M4) is designed or implemented, the repair shall cite the two first-measurement conclusions — M-a (write-path attribution: which code path writes without an explicit save) and M-b (the git-strategy dirty-gate bypass mechanism and the feedback.yaml seam blank-line loss cause) — and no repair code shall be authored before both measurements are complete. 각 측정은 커맨드 + 관측 출력 + 트리 SHA로 progress.md §E.2에 귀속된다.

---

## §3 Scope & Boundaries

### §3.1 In Scope

- **쓰기 시점**: 명시적 저장 동작 없는 `.moai/config/**` 기록의 전면 차단
- **쓰기 범위**: 저장이 접촉하는 섹션 파일의 최소화(비편집 섹션 diff 없음)
- **포맷 충실도**: 키 순서·빈 줄·주석·unknown key 보존
- **폼 파서 중복 값 인지**(`internal/web/schemaform.go` 시설 — t509가 read-only mirror 설계로 회피한 대상 자체)
- 위 4축을 검증하는 회귀 가드와 실물 재현 절차

### §3.2 경계 — 형제 카드 (run-phase 드리프트 방지)

- **t509 (codex 패널 B1)**: moai web의 codex 패널 UI/상태 표면은 t509 소관이다. 본 SPEC은 codex 패널을 수리하지 않는다 — 쓰기 안전 축만 소관한다.
- **t510 (쓰기 위치 축)**: "의도하지 않은 곳에 파일을 쓴다"(write location)는 t510 소관이다. 본 카드는 쓰기 **시점**(timing)과 **범위**(scope)만 다룬다. run-phase가 위치 축 결함을 발견하면 본 카드에서 수리하지 않고 리드에 회신해 t510로 이관한다.

### Out of Scope — 부분 실패 롤백 트랜잭션

- `handleSave`의 순차 쓰기가 후행 단계 실패 시 선행 쓰기를 유지하는 현재 동작(코드에 에러 배너로 문서화됨, `internal/web/handlers.go:483-489` 등)은 본 카드에서 재설계하지 않는다. 다만 REQ-WWS-003의 쓰기 최소화가 접촉 단계 수 자체를 줄인다.

### Out of Scope — CSRF/token 인프라

- `internal/web/app.go`의 @MX:NOTE(REQ-WC-009)가 금지한 CSRF/token 도입은 범위 밖이다. loopback bind + Host-check 미들웨어 모델을 보존한다.

### Out of Scope — primary checkout 수복

- 관측으로 오염됐던 두 파일은 이미 explicit-path `git restore`로 복구됐다(1회 복구 목적으로 쓰인 것이며 일반화하지 않는다). 본 카드의 재현·검증은 격리 트리에서만 수행하며 primary checkout에 어떤 쓰기도 하지 않는다.

### Out of Scope — typed-path 재설계의 별도 이슈화

- C3(llm.yaml 매 `Save()` 재기록)은 REQ-WWS-003의 일반 원칙(값 불변 섹션 diff 없음)으로 흡수된다. 별도의 typed-path 재설계 SPEC을 이 카드에서 만들지 않는다.

---

## §4 Constraints

- **검증 환경**: 모든 재현·검증은 격리 트리(fresh worktree 또는 전용 fixture 트리)에서 수행한다. primary checkout(2026-09-03 이래 dirty, 출처 불명 변경 보유)은 절대 건드리지 않는다.
- **스냅샷 우선**: 재현 전 대상 파일을 스냅샷 복사(`cp`)해 두고, 관측은 복사본과의 diff로 판정한다. `git restore`를 본 카드의 검증 도구로 사용하지 않는다.
- **이진 최신성**: 검증에 쓰는 `moai` 바이너리는 측정 시점 코드에서 빌드된 것임을 확인한다. 수리 후 검증은 재빌드 없이는 무효다(`verification-completeness.md` §1.3 deployment 축).
- **측정 귀속**: 모든 측정은 커맨드 + 관측 출력 + 트리 SHA로 귀속한다(`verification-claim-integrity.md` §2). 전달된 수치를 새 측정으로 재사용하지 않는다.

---

## §5 Success Criteria

acceptance.md의 AC 매트릭스가 유일한 판정 기준이다. 요약:

1. 실물 재현(RED)이 격리 트리에서 관측되고, 그 증거가 커맨드·출력·exit code·트리 SHA 4요소로 귀속된다.
2. 첫 측정 2건(M-a 쓰기 경로 귀속, M-b gate 우회·seam 충실도 원인)이 수리 설계에 선행한다(REQ-WWS-008).
3. 모든 부재-가드 AC가 RED-first + 뮤턴트 포착으로 채택된다.
4. 수리 후 같은 재현 절차가 GREEN이고, 비편집 섹션은 diff 없음이다.

---

## §6 Cross-References

- SPEC-GITSTRATEGY-SAVE-ISOLATION-001 — dirty-gate 계약(REQ-GSI-001/002)의 원 SPEC
- SPEC-WEB-CONSOLE-011 — seam/typed 이원 쓰기 구조와 yamlpatch seam의 원 SPEC
- SPEC-WEB-CONSOLE-010 — settings schema SSOT(`internal/settings/schema.go`)
- SPEC-FEEDBACK-AUTO-SUBMIT-001 — feedback.yaml seam-writable 재개 근거
- t509 — codex 패널 소관 카드 / t510 — 쓰기 위치 축 소관 카드
