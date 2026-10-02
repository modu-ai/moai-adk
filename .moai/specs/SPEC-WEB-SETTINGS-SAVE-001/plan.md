---
id: SPEC-WEB-SETTINGS-SAVE-001
title: "plan — moai web 설정 저장 불가 수리 · 감사 pin 기본값 · 원격병합 가드"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: GOOS
tier: M
---

# plan — SPEC-WEB-SETTINGS-SAVE-001 (card t1393)

> 이 문서는 상태 축에서 무상태다 — SPEC의 lifecycle은 spec.md frontmatter만 운반한다.

## §A Context

### §A.1 세 스코프와 접근

| 스코프 | 우선순위 | 접근 |
|---|---|---|
| ① 웹 설정 저장 불가 | High | 근본원인 우선(root-cause-first): 관측-RED 재현 → 카드 지정 순서 (a)검증 거절 → (b)저장 경로 예외 → (c)t1314 회귀 탐증 → 최소 수리 |
| ② 감사 핀 기본값 | Medium | t1368 선례의 3중 정합(Go 기본값 = 배포 템플릿 = resolver 폴백), 관측-RED 규율 |
| ③ 원격병합 가드 | High | dirty/unpushed 가드 뒤에 확인 단계 추가, fail-open preserve, merge-first 순서 불변 |

### §A.2 스코프 ① 가설 체인 (plan-phase 근거 정리)

카드 지정 탐증 순서는 (a)→(b)→(c)다. plan-phase에서 모은 증거는 아래와 같다:

- **(a) 요청 스키마 검증 거절**: `parseSchemaForm`(internal/web/schemaform.go:364)의 원자 거절 후보 — 중복 폼 값(REQ-WWS-006), closed-set 검증(현재값 읽기 실패 시 strict 폴백, handlers.go:398), bool companion 누락. **반증 기회**: 거절 경로는 2xx + "Validation failed" 배너를 렌더한다(handlers.go:434-452, t1105). 운영자가 "아무 일도 없었다"고 관측했다면 배너 부재가 이 가설을 누른다 — 단, §E 미지수(dirty 배지 정체)가 해소될 때까지 폐기하지 않는다.
- **(b) 저장 경로 예외**: 9개 시늅의 실패는 `renderErrorPage` 2xx(handlers.go:602)로 노출되나, **재렌더 자체의 실패**는 `render()`가 500(handlers.go:192)을 답하고 hx-boost가 폐기한다 → 무음. 패닉도 500 빈 응답 → 무음.
- **(c) t1314 회귀(SPEC-WEB-SAVE-LOSSLESS-001, 커밋 9be71a4f1)** — **plan-phase 증거상 최유력 후보**. 근거: ① 실패하는 두 탭(git-worktree·audit)이 정확히 workflow.yaml seam 필드의 집합이다(schema_sections.go:313-400). ② t1314 M1/M2가 바로 그 seam 쓰기 경로를 재작성했다 — "seam line-splice routing + difference gate(차이 없으면 쓰기 생략) + per-field no-op 게이트(loaded-value + disk-scalar 동등)". 중첩 경로(`workflow.worktree.auto_create` = workflow→worktree→auto_create 3단)에서 no-op 게이트가 실제 변경을 no-op으로 오판하면 **무음 미기록** — 운영자 관측과 정확히 일치한다. ③ 다른 탭(예: identity/language — profile 스토어 경로)은 영향 보고가 없다.

단, plan-audit iter 1의 codex 2차 감사가 HEAD f130aa041에서 임시 overlay 테스트로 "최소 형태 제출(토글+`__present`)은 정상 저장했다"고 보고했다(중간 신뢰도 — 본 트리에서 미재검증, **Gap으로 기록**). 이 관측이 재확인되면 (c)는 유력성을 잃고 선행 대립 가설은 **"운영자 실행 바이너리 ≠ 본 트리"** 편차로 옮겨간다 — 그래서 M1의 첫 행위는 바이너리 버전 확인 + 실패 페이로드 캡처이고, 미재현 분기는 §F M1에 명문화돼 있다(AC-WSS-001 — decision-index Q2와 결부).

M1은 이 순서대로 탐증하되, 관측-RED 재현 테스트가 어느 가설과 맞는지로 판정한다. 판정이 (c)가 아니면 수정 대상이 M1 판정서를 따른다.

### §A.3 스코프 ③ 설계 공간

**채택됨(리드 카드자율 처분, 2026-10-01 — decision-index Q1 DECIDED)**: fetch 없는 원격추적 도달성. 술어: (i) `git merge-base --is-ancestor <branch-tip> refs/remotes/origin/develop` 또는 (ii) `git cherry refs/remotes/origin/develop <branch>` 공집합(patch-id 등가 — 스쿼시 병합도 커버; SPEC-WORKTREE-SQUASH-MERGE-001 교훈 — 도달성 단독은 스쿼시를 못 본다). 채택 근거: 착지는 단조적이다 — 병합은 취소되지 않으므로 스테일 원격추적 참조가 만들 수 있는 오판은 "아직 못 착지"→보존(안전 방향)뿐이다. 반면 세션 종료마다의 fresh fetch는 네트워크 비용과 새 실패 모드를 모든 종료 경로에 얹어 REQ-WSS-304의 저렴-공통-경로 요구와 충돌한다. 참조 부재·검사 오류·판정 불신뢰는 REQ-WSS-302의 fail-open 보존이 흡수한다. <!-- moving-ref-ok: 채택 술어가 참조하는 origin/develop은 측정 앵커가 아니라 종료 시점 판정의 대상(SUBJECT)이다 — 스테일 오판 방향이 보존이라 안전 -->

참고: `gitHasUnpushedReal`(session_worktree.go:741)의 무-업스트림 분기는 `rev-list --count HEAD --not --remotes`로 리모트 보유 여부만 본다 — "푸시됐지만 미병합"이 통과하는 구멍의 정확한 위치다(REQ-WSS-306이 막는 행동).

## §B Known Issues

1. **dirty 배지 런타임 생산자 부재**(spec §E 미지수): 본 트리 app.js에는 폼 변경 추적이 없다. 운영자 관측과 정합시키는 것이 M1 진단의 일부다. 운영자 실행 바이너리가 develop 팁과 다를 가능성을 배제하지 않는다 — M1 첫 행위가 바이너리 버전 확인이고, AC-WSS-001의 미재현 분기와 결부된다(codex overlay 보고 "최소 제출 정상 저장"이 참이면 본 트리 재현은 안 된다 — 미검증 Gap).
2. **`render()` 500 무음 경로**(handlers.go:192): (b) 가설이 적중하면 이 전파 경로도 함께 수리 대상이다 — SPEC-WEB-CONSOLE-017이 고친 것은 시늅 실패 경로뿐이었다.
3. **`defaults_test.go`에 Audit 블록 직접 단언 부재**: 기본값 단언은 `audit_models_test.go`(round-trip 픽스처, internal/config/audit_models_test.go:103-130 claude medium 명시)와 cli 측 픽스처(audit_pin_test.go 등)에 흩어져 있다 — M4의 관측-RED 대상 파일 목록을 M4 착수 시 `grep -rn "medium\|claude-opus-5-5" internal/config internal/cli --include="*_test.go"`로 재확정한다.
4. **통합 브랜치 결정 — Q1 채택으로 해소**: 채택 술어는 `refs/remotes/origin/develop`을 고정한다. 분산 사용자의 통합 브랜치명이 다른 경우(원격추적 참조 부재)는 REQ-WSS-302 ②의 fail-open 보존이 흡수한다 — 오판 방향이 항상 "보존"이라 안전하다.

## §C Pre-flight

1. 워크트리 항등 확인: `git rev-parse --show-toplevel` → 본 트리, HEAD가 develop 기반이며 다른 세션의 쓰기 없음(레인 배차 원칙).
2. 기준선 측정(수정 전 GREEN 관측): `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test -timeout 30m ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/settings/... ./internal/profile/...` — 적색이면 그 목록이 M1 진단 입력이다(양성 대조 겸용). plan-phase 기준선은 이미 측정돼 acceptance.md §D.8 증거 원장 EV-001·002로 고정됐다 — run-phase는 같은 명령의 수정 후 GREEN 셀을 원장에 쌍으로 추가한다.
3. 선행 SPEC 판독: SPEC-WEB-SAVE-LOSSLESS-001(spec+acceptance — 무손실 계약의 정확한 문면), SPEC-MODEL-MATRIX-UPDATE-001(3중 정합 선례의 AC 문면), SPEC-SESSION-WORKTREE-001(REQ-SW-008/009/010 계약).
4. 증거 디렉터리 생성: `.moai/reports/t1393/`.

## §D Constraints

- 재측정 단위: `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test -timeout 30m ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/settings/... ./internal/profile/...` — 명시 변수 나열 단일 호출(`unset MOAI_*` 글로브 폐기 — 0-매치 환경에서 미실행, plan-audit D3 실측). `-run` 셀렉터로 가드 회피 금지(교훈: 재측정은 소관 패키지 전체가 단위).
- Template-First: M4가 `internal/template/templates/.moai/config/sections/workflow.yaml`를 수정 → `make build` 재생성 + 임베드 드리프트 0 확인(`make build` 후 `git status --short`에 재생성 노이즈 없음).
- 우선순위 라벨만 사용 — 시간 예측 금지. 마일스톤 순서는 "A 완료 후 B 착수".
- 커밋은 레인 소관 — 커밋 메시지에 `t1393` 명기, 커밋 본문에 가드 동사 리터럴 회피(계약 서명 가드 교훈).
- 새 상수는 `internal/config/envkeys.go`/`defaults.go` 단일 원천 규율(하드코딩 방지 §14) — 감사 기본 모델/effort 토큰은 기존 상수(`claudeAuditDefault*`, `DefaultCodexAuditModel`) 재사용, GLM 신설 상수는 같은 파일·같은 이름 규약.
- 테스트 격리: `t.TempDir()` 전용, OTEL env 금지, 병렬 테스트에서 전역 상태 오염 금지.

## §E Self-Verification

- E1: AC 행렬 전 행 재실행(acceptance.md §D) — 각 행의 판정 명령과 원문 출력을 `.moai/reports/t1393/verdict.md`로 반출.
- E2: env-scrub 소관 5패키지(web · cli · config · settings · profile) 재측정 전량 GREEN — 위 단위 명령의 원문 출력 첨부.
- E3: 관측-RED 증거: 스코프 ① 재현 테스트와 스코프 ② 기본값 단언의 수정 전 적색 출력 + 수정 후 녹색 출력 쌍.
- E4: 무손실 회귀: `save_lossless_test.go` 전량 무수정 GREEN.
- E5: 템플릿 임베드: `make build` 후 `git status --short` 클린 + `go test ./internal/template/...` 해당 축 GREEN(M4에서만).
- E6: `go vet` 변경 패키지 + `golangci-lint run`(CI 판 v2.1.6).
- E7: 시늅 배선 grep: `cleanupSessionWorktreeFn` 3호출부와 가드 경유 확인(handlers.go:115/466/profile.go:80 대상).

## §F Milestones

> 순서 원칙: 변동 가능성이 높은 결정(사용자 대면 흐름, 새 동작)을 앞에, 기계적 단계를 뒤에.

### M1 (High) — 스코프 ① 진단 + 관측-RED 재현

첫 행위: 운영자 실행 바이너리 버전 확인(§B.1 질의 — decision-index Q2) + 실패 요청 페이로드 캡처(logSaveFailure stderr 또는 디버그 빌드 — 전체 폼 POST, 교차 탭 hidden 입력 포함). 카드 순서 (a)→(b)→(c) 탐증. 재현 테스트는 캡처한 페이로드 형태를 따른다(AC-WSS-001). 판정서 `.moai/reports/t1393/root-cause.md`(가설 판정 + file:line 증거 + §E 미지수 정합 + 관측-RED 4요소 셀). **미재현 분기(명문화)**: 본 트리(f130aa041)에서 재현되지 않으면 구현을 강행하지 않고 (a)/(b) 재탐증 + 운영자 바이너리 편차 진단으로 분기하며, codex overlay 보고("최소 제출 정상 저장" — 중간 신뢰도·미검증 Gap)의 진위를 이 단계에서 판정한다. 진입 조건: Pre-flight 기준선 측정 완료(EV-001·002).

### M2 (High) — 스코프 ① 수리

M1 판정 대상의 최소 수정 + 재현 테스트 GREEN + 무손실/원자거절 회귀 방지(AC-WSS-004·005). (c)가 확정되면 대상은 t1314 seam/no-op 게이트 경로 — 그래도 검증·렌더 전파 경로(handlers.go:398 현재값 폴백, :192 render 500)의 방어적 정합은 범위 밖으로 밀지 않는다. M1 판정이 다른 가설(바이너리 편차 포함)을 지지하면 이 마일스톤의 수정 대상을 갱신하고 리드에게 보고.

### M3 (High) — 스코프 ③ 원격병합 가드

술어 확정됨(Q1 DECIDED — fetch 없는 원격추적 도달성, §A.3) → 가드 구현(session_worktree.go, dirty/unpushed 가드 뒤) + `cleanupSessionWorktreeFn` 시늅 경유 확인 + 3케이스 테스트(미확증 거부 / 착지 통과[git-cherry patch-id 스쿼시 등가 포함] / fail-open) + merge-first 순서 보존 테스트(AC-WSS-010..013·015). 자동머지 성공-직후 거부가 의도된 동작임을 공지 문구에 명시.

### M4 (Medium) — 스코프 ② 핀 3중 정합

`internal/config/defaults.go` Audit 블록(claude effort→high, GLM `{glm-5.3, max}`) + `internal/cli/mcp_claude.go:21` effort→high + GLM resolver 폴백 전환(glm-5.3-flash→glm-5.3 + effort 신설 — mcp_glm.go:54 `glmAuditDefaultModel`·:203-211 `resolveGLMAuditModelEffort`) + 템플릿 workflow.yaml audit 블록 갱신 → `make build`. 기존 단언 테스트는 적색 관측 후 갱신(AC-WSS-006..009). codex는 불변 검증만(AC-WSS-008).

### M5 (Medium) — 종합 검증 + 증거 반출

E1-E7 전 항 실행, `.moai/reports/t1393/verdict.md` 완성(증거 원장 EV-001·002에 수정 후 GREEN 셀 추가), 리드 완료 보고(병합 SHA는 재측정 통과 직후 보고).

## §G Anti-Patterns

- **맹목 기대치 수정 금지**: 기존 테스트의 기대값을 새 값에 맞추기 전 적색 관측이 없으면 실패(REQ-WSS-203, 교훈 — 동결/부재류 AC 행은 직접 재실행).
- **로컬 전체 스위트 금지**: 소관 3패키지만, env-scrub 1회 호출로(AGENTS.local.md §4).
- **출력 0 = 부재 아님**: 가설 폐기에 양성 대조를 둔다 — 예: (a) 폐기 전에 거절 배너가 정말 렌더되는지 재현 테스트로 확인.
- **-run 셀렉터로 소관 패키지 가드 회피 금지**.
- **커밋 본문에 가드 동사 리터럴 금지**(계약 서명 가드가 본문 텍스트를 잡는다) — 패러프레이즈.
- **요구사항(GEARS)과 AC(Given-When-Then) 혼동 금지** — acceptance.md는 판정 가능한 시나리오만.
- **`moai update` 보호 규율 우회 금지** — 템플릿 갱신은 배포 원본에서만.

## §H Cross-References

- spec.md §A(증거), acceptance.md(AC-WSS-001..014), decision-index.md(미결정 4행)
- SPEC-WEB-SAVE-LOSSLESS-001 · SPEC-MODEL-MATRIX-UPDATE-001 · SPEC-V3R6-AUDIT-MODEL-PIN-001 · SPEC-SESSION-WORKTREE-001
- AGENTS.md §3 · AGENTS.local.md §2(Template-First)·§4(재측정)·§14(하드코딩 방지)
- 카드 t1393 · `.moai/reports/t1393/`
