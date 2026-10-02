---
id: SPEC-WEB-SETTINGS-SAVE-001
title: "moai web 설정 저장 불가 수리 · 감사 pin 기본값 확정 반영 · 세션 워크트리 원격병합 가드"
version: "0.1.0"
status: in-progress
created: 2026-10-01
updated: 2026-10-02
author: GOOS
priority: P1
phase: "v3.2.0"
module: "internal/web, internal/settings, internal/config, internal/cli"
lifecycle: spec-anchored
tags: "web-console, settings-save, lossless-save, audit-pin, session-worktree, remote-merge-guard"
tier: M
---

# SPEC-WEB-SETTINGS-SAVE-001

카드 t1393 (운영자 발행 2026-10-01). 세 스코프를 하나의 SPEC으로 다룬다: ① `moai web` 설정 저장 불가 수리(최우선), ② 감사 pin 기본값의 운영자 확정값 반영, ③ 세션 워크트리 정리의 원격병합 착지 확인 가드.

## HISTORY

| 버전 | 날짜 | 내용 |
|---|---|---|
| 0.1.0 | 2026-10-01 | 초판 — 카드 t1393 plan-phase 산출 (manager-spec). 산출물: spec.md · plan.md · acceptance.md · progress.md · decision-index.md |

## §A 배경과 문제 정의

### §A.1 스코프 ① — 웹 설정 저장 불가 (P1, 운영자 재현 2026-10-01)

운영자가 `moai web` 콘솔에서 **git-worktree 설정 탭과 감사(audit) 설정 탭**에서 토글·값을 바꾼 뒤 "설정 저장" 버튼을 눌러도 아무 일도 일어나지 않는다 — 미저장 표시가 그대로 남고 디스크에 아무것도 기록되지 않는다(스크린샷 관측: worktree `auto_create`·`auto_merge` 토글).

저장 경로의 구조(근거, 본 워크트리 develop f130aa041 기준):

- 쓰기 처리기: `internal/web/handlers.go:346` `handleSave` — `POST /save`의 유일 쓰기 경로. 9개 영속화 시늅(seam)을 순차 실행하고 실패를 `renderErrorPage`(handlers.go:602, 2xx 재렌더 — SPEC-WEB-CONSOLE-017)로 노출한다.
- 실패 탭의 필드 집합: `internal/settings/schema_sections.go:313-316` — `workflow.worktree.{auto_cleanup,auto_create,auto_merge,tmux_preferred}` bool 토글. `:356-367` — `workflow.audit.model` + gates closedSeam. `:381-400` — 감사 6핀 라디오(`workflow.audit.{claude,codex,glm}.{model,effort}`). **실패하는 두 탭이 정확히 workflow.yaml seam 필드의 집합이다.**
- 폼 파서: `internal/web/schemaform.go:364` `parseSchemaForm` — bool은 hidden companion(`<name>__present`)으로 제출/미제출을 구분하고, 중복 폼 값은 원자 거절에 합류한다(REQ-WWS-006). 닫힌 집합 검증은 디스크 현재값 passthrough(RC2)인데, 현재값 읽기 실패 시(handlers.go:398) strict 폴백한다.
- 원자 거절 경로: `handlers.go:434-452` — 2xx 재렌더 + "Validation failed" 배너(카드 t1105 — 4xx는 hx-boost가 폐기하므로 2xx로 응답).
- 영속화 시늅: `handlers.go:507` → `settings.ApplySchemaEdits`(internal/web/app.go:142 배선) — workflow.yaml yamlpatch seam.
- 무음 실패 후보 경로: `handlers.go:189-205` `render()` 자체의 렌더 실패는 500 → hx-boost(htmx 2.0.4)가 비-2xx 본문을 폐기 → 사용자에게 무음. 패닉도 net/http 복구로 500 빈 응답.
- 클라이언트: `internal/web/assets/app.js` — 폼 입력 변경 추적 리스너가 0건이고, 저장 클러스터의 `dirty` 상태(shell.templ:265-269)에 대한 런타임 생산자가 관측되지 않는다. 서버 `settingsSaveState`(internal/web/settings_shell.go:61)는 error|saved|clean만 반환한다.

### §A.2 스코프 ② — 감사 pin 기본값 확정 반영

운영자 확정(2026-10-01): claude `{claude-opus-5-5, high}` · codex `{gpt-6.1-sol, high}` · glm `{glm-5.3, max}`.

현재 상태(근거):

- Go 기본값: `internal/config/defaults.go` `NewDefaultWorkflowConfig` Audit 블록 — claude `{claude-opus-5-5, medium}` · codex `{DefaultCodexAuditModel, high}` · **GLM 핀 부재(빈 값)**.
- resolver 폴백: `internal/cli/mcp_claude.go:20-21` `claudeAuditDefaultEffort = "medium"`. codex는 t1368/t1386이 `{gpt-6.1-sol, high}`로 이미 정합(codexAuditDefault* 상수 — 형제 픽스처가 참조). GLM resolver 폴백 현재값은 `{glm-5.3-flash, effort 빈 값}`이다 — `internal/cli/mcp_glm.go:54` `glmAuditDefaultModel = config.DefaultGLMHigh`, `internal/config/defaults.go:250` `DefaultGLMHigh = DefaultGLM53Flash` = `"glm-5.3-flash"`(defaults.go:248), `resolveGLMAuditModelEffort`(mcp_glm.go:203-211)은 핀 부재 시 `ModelEffort{Model: glmAuditDefaultModel}`(effort 빈 값)를 반환.
- 배포 템플릿: `internal/template/templates/.moai/config/sections/workflow.yaml:112-124` — claude medium · codex high · glm 빈 값.
- 기대 델타: claude effort medium→high · GLM은 단순 "핀 신설"이 아니라 **resolver 폴백 모델 전환(glm-5.3-flash→glm-5.3)+effort 신설(max)**이다 — 핀 없이 돌던 GLM 감사 호출의 실제 모델이 바뀐다 · codex 불변(검증만).
- 선행 SPEC: SPEC-MODEL-MATRIX-UPDATE-001(t1368)이 codex 핀을 "Go 기본값 = 배포 템플릿 = resolver 터미널 폴백" 3중 정합으로 반영한 선례 — 본 스코프는 같은 3중 정합 규율을 claude·glm에 적용한다.

### §A.3 스코프 ③ — 세션 워크트리 정리 원격병합 가드

`internal/cli/session_worktree.go:644` `cleanupSessionWorktree`는 dirty 가드(REQ-SW-010)와 unpushed 가드(card t673)를 통과한 워크트리를 **원격 병합 착지를 확인하지 않고** 제거한다. `gitHasUnpushedReal`(session_worktree.go:741)은 "브랜치가 리모트에 존재하는가"까지만 보므로, **푸시는 됐지만 통합 브랜치(origin/develop)에 병합되지 않은 브랜치**의 워크트리를 제거 대상으로 오판한다 — AGENTS.md §3("원격 병합이 착지 확인되기 전까지 워크트리를 폐기하지 않는다") 위반.

호출 지점: `internal/cli/web.go:115` · `internal/cli/init.go:466` · `internal/cli/profile.go:80`의 `cleanupSessionWorktreeFn` 시늅(profile_setup 경로가 소비). 순서 제약: `internal/cli/session_worktree_automerge.go:137` `sessionExitAutoMerge`가 정리보다 **먼저** 실행된다(merge-first). 자동머지 성공 직후는 "로컬 develop에는 병합됐지만 원격 착지 전" 상태이므로, 가드는 이 상태에서 처분을 **거부하는 것이 의도된 동작**이다(스코프 ③은 향후 `auto_cleanup` 기본값 전환의 선행 조건 — 운영자 의착 기록, decision-index Q4).

## §B 목표

1. (P1) POST /save가 worktree·감사 탭의 편집을 디스크에 기록하게 한다 — 근본원인 수리, 관측-RED 재현 선행, 기존 저장 계약의 회귀 0.
2. 감사 핀 분배 기본값을 운영자 확정값으로 3중 정합한다(Go 기본값 = 배포 템플릿 = resolver 폴백).
3. 세션 워크트리 처분에 원격병합 착지 확인 가드를 넣는다 — 공통 청산 경로는 오늘과 같이 저렴하게 통과.

## §C 요구사항 (GEARS)

### §C.1 스코프 ① — 설정 저장 수리

- REQ-WSS-101: When 스코프 ① 수리 작업이 시작될 때, the implementation shall be preceded by an observed-RED reproduction test — 운영자 증상("POST /save가 `workflow.worktree.*`/`workflow.audit.*` 편집을 디스크에 기록하지 않는다")을 수정 전 트리에서 적색으로 재현하는 테스트 — and the diagnosis record shall name the failing seam with file:line evidence under `.moai/reports/t1393/`.
- REQ-WSS-102: When a POST /save submission carries an edited `workflow.worktree.*` bool field(hidden `__present` companion 포함), the system shall persist the submitted value to `.moai/config/sections/workflow.yaml` through the schema edit seam(`settings.ApplySchemaEdits`).
- REQ-WSS-103: When a POST /save submission carries an edited `workflow.audit.*` field, the system shall persist it likewise.
- REQ-WSS-104: While POST /save의 어느 영속화 시늅이 실패할 때, the system shall surface the failure reason in the 2xx re-rendered page banner and shall not leave the failure unreported — SPEC-WEB-CONSOLE-017 계약(2xx 재렌더)과 `logSaveFailure` stderr 표면을 보존한다.
- REQ-WSS-105: While SPEC-WEB-SAVE-LOSSLESS-001의 무손실 계약이 유효할 때, the save path shall preserve unmodeled keys and hand-maintained comments in every touched section file.
- REQ-WSS-106: The system shall not regress currently-GREEN save-surface behavior — 원자 거절 계약, 중복 폼 값 거절(REQ-WWS-006), 닫힌 집합 검증, GLM/Jev 자격증명 시늅 — while repairing scope ①.

### §C.2 스코프 ② — 감사 핀 기본값

- REQ-WSS-201: The system shall carry the operator-confirmed default audit pins — claude `{claude-opus-5-5, high}`, codex `{gpt-6.1-sol, high}`, glm `{glm-5.3, max}` — in all three pin sources: the Go default(`NewDefaultWorkflowConfig`), the distributed template(`internal/template/templates/.moai/config/sections/workflow.yaml`), and the audit resolver terminal fallbacks.
- REQ-WSS-202: When a project's workflow.yaml carries explicit pin values, the system shall keep explicit values outranking the defaults(기존 우선순위: 사용자 핀 > SSOT 셀 > 빈 값 — 불변).
- REQ-WSS-203: While 스코프 ②의 테스트 갱신이 이루어질 때, the implementation shall observe RED against the new expected values before editing existing assertions — 관측-RED, 맹목 기대치 수정 금지.
- REQ-WSS-204: The system shall not change the audit gates distribution(claude/codex required, glm advisory), the `audit_model` enum, or REQ-AMP-008(핀은 감사 전용, 태스크 위임 불가) while applying scope ②.

### §C.3 스코프 ③ — 원격병합 가드

- REQ-WSS-301: When `cleanupSessionWorktree`가 clean worktree + clean exit로 제거 단계에 도달할 때, the system shall first confirm the branch's remote merge has landed and shall refuse disposal(preserve + notice) where that confirmation is absent.
- REQ-WSS-302: While 원격병합 확인이 성립하지 않을 때 — ① 확인 검사의 오류, ② 원격추적 통합 참조(`refs/remotes/origin/develop`)의 부재, ③ 확인 결과의 이상(참조 소실·비정상 갱신·도달 판정 불신뢰) — the system shall preserve the worktree and emit a notice naming the unconfirmed state(fail-open preserve — dirty/unpushed 가드와 동일 계약). 확인 술어는 **fetch 없는 원격추적 도달성**으로 고정된다(decision-index Q1 DECIDED): `git merge-base --is-ancestor <branch-tip> refs/remotes/origin/develop` 또는 `git cherry refs/remotes/origin/develop <branch>` 공집합(patch-id 등가 — 스쿼시 병합 커버) 중 하나라도 성립하면 착지로 인정한다. <!-- moving-ref-ok: 런타임 술어의 대상(SUBJECT) — 종료 시점 원격추적 참조의 현재값을 판정하는 동작 명세이지 측정 앵커가 아니며, 스테일 참조의 오판 방향은 "못 착지"→보존(fail-open)으로 안전하다 -->
- REQ-WSS-303: The guard shall run after the existing dirty and unpushed guards, and shall not reorder the merge-first contract — `sessionExitAutoMerge`가 `cleanupSessionWorktree`보다 먼저(internal/cli/web.go:114-116 · init.go:466 · profile.go:80) — while adding the confirmation step.
- REQ-WSS-304: When the branch tip is reachable from the remote-tracking integration ref, the system shall proceed with disposal on the same path as today — 공통 청산-착지 경로는 대화형 마찰 없이 저렴하게 유지한다.
- REQ-WSS-305: The guard shall route through the `cleanupSessionWorktreeFn` seam — 세 종료 경로(web/init/profile_setup)가 모두 상속하고 커맨드 수준 테스트가 관측할 수 있게.
- REQ-WSS-306: The system shall not delete a session worktree whose branch the remotes hold but whose merge into the integration branch is unconfirmed — 브랜치 푸시만으로 REQ-WSS-301의 확인이 충족되지 않는다.

## §D 제약

- **재측정 단위**: `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test -timeout 30m ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/settings/... ./internal/profile/...` — 소관 패키지 통째로, 명시 변수 나열 단일 호출. `unset MOAI_*` 글로브는 0-매치 환경에서 판정 명령 자체를 실행시키지 못한다(plan-audit D3 실측: bash `not a valid identifier` rc=1, zsh 스크립트 중단) — 나열 집합은 실행 레인의 생존 변수를 따르며 `env | grep -o '^MOAI_[A-Z0-9_]*' | sort`로 사전 재유도한다. 로컬 전체 스위트 금지(전 패키지 판정은 CI 몫 — AGENTS.local.md §4).
- **Template-First**: REQ-WSS-201이 `internal/template/templates/**`를 건드린다 → 템플릿 원본 수정 후 `make build`(선행 agents-emit-check/commands-emit-check 포함). 스코프 ①·③은 Go 전용 — 템플릿 미러 불요.
- **개발 모드**: `development_mode: tdd`(.moai/config/sections/quality.yaml) — 마일스톤마다 RED-GREEN-REFACTOR, 관측-RED 규율.
- **git 소관**: 커밋·통합은 카드 레인/통합 창 소관. 커밋 메시지에 카드 id(t1393) 명기, 증거 디렉터리 `.moai/reports/t1393/`.
- **sync 검증감사 렌즈**: `--deep`.
- **무음 실패 금지**: 모든 수정 경로에서 사용자 관측 가능한 실패 신호를 유지한다(REQ-WSS-104).

## §E 진단 미지수 (open questions)

1. **dirty 배지 런타임 생산자**: 본 트리(develop f130aa041) 기준 `data-save-state="dirty"`를 만드는 클라이언트 코드가 관측되지 않는다(app.js 폼 변경 리스너 0건; `settingsSaveState`는 error|saved|clean만 반환). 운영자 스크린샷의 "미저장" 배지가 (a) 실제 dirty 상태인지 — 그렇다면 운영자 바이너리가 본 트리와 다르거나 미관측 생산자가 있는 것 — 아니면 (b) 오류 상태 배지("저장하지 못한 필드가 있습니다")의 재발인지를 M1 진단이 정합시킨다. 이 구분이 가설 (a) 대 (b) 판별의 열쇠다.
2. **최소 제출의 재현성(미검증 Gap)**: plan-audit iter 1의 codex 2차 감사가 HEAD f130aa041에서 임시 overlay 테스트로 "최소 형태 제출(토글+`__present`)은 정상 저장했다"고 보고했다(중간 신뢰도 — 본 트리에서 미재검증, **Gap으로 기록**). 참이면 최유력 가설 (c)는 유력성을 잃고 선행 대립 가설은 **"운영자 실행 바이너리가 본 트리와 다르다"**로 옮겨간다 — 위 1번의 배지 정체(decision-index Q2)와 결부된다. 따라서 AC-WSS-001의 재현은 운영자 실효 요청 형태(전체 폼 POST)로 고정하고, 미재현 시 분기(§F M1, plan.md §A.2)를 따른다.

## §F 수용 기준

전체 AC 행렬은 `acceptance.md`(AC-WSS-001..016)에 있다. 요약:

- 스코프 ①: 관측-RED 재현(AC-WSS-001 — 운영자 실효 요청 형태 고정, 미재현 분기 명문화) → worktree 토글 영속(AC-WSS-002) · 감사 필드 영속(AC-WSS-003) · 무음 실패 부재(AC-WSS-004) · t1314 무손실 회귀 0(AC-WSS-005 — 기준선 EV-002). Blocker 행의 RED-now/기준선 셀 처분은 acceptance.md §D.0(§2.1).
- 스코프 ②: Go 기본값 claude high(AC-WSS-006) · glm 핀 신설 3중 정합(AC-WSS-007) · codex 불변(AC-WSS-008) · 관측-RED 증거(AC-WSS-009).
- 스코프 ③: 미확증 거부(AC-WSS-010) · 착지 경로 통과(AC-WSS-011) · fail-open 보존(AC-WSS-012) · merge-first 순서 보존(AC-WSS-013) · 시늅 경유(AC-WSS-015).
- 스코프 ② 보충: 핀 우선순위·gates/enum 무변경(AC-WSS-016).
- 전체: 소관 3패키지 env-scrub 재측정 전량 GREEN(AC-WSS-014).

## Out of Scope

### Out of Scope — M8 PR-병합 정리 경로

- `internal/cli/session_worktree_prmerge.go:201`의 PR-병합 정리도 "푸시 ≠ 병합" 미확증 가족이지만, 본 SPEC의 가드는 세션 종료 처분(`cleanupSessionWorktree`) 한 곳에만 적용한다. PR-병합 경로의 가드 확장은 후속 카드로만 발행한다.

### Out of Scope — auto_cleanup 기본값 전환

- `workflow.worktree.auto_cleanup`의 분배 기본값은 OFF로 유지한다. 본 SPEC의 가드는 향후 전환(운영자 의도)의 선행 조건일 뿐이며, 전환 자체와 그 시점은 다루지 않는다.

### Out of Scope — 감사 gate·컨버전스 로직

- 감사 gate 분포(claude/codex required, glm advisory), `audit_model` enum, multi 컨버전스 엔진은 그대로다(REQ-WSS-204). 핀 `{model, effort}` 기본값만 다룬다.

### Out of Scope — 기존 프로젝트 핀 값 마이그레이션

- 사용자 프로젝트 workflow.yaml에 이미 영속된 핀 값은 재기록하지 않는다. 템플릿 갱신은 템플릿에서 새로 깔리는 프로젝트와 부재 키에만 새 기본값을 봉사한다.

### Out of Scope — 설정웹 UI 개편

- 저장 수리 이외의 콘솔 UI 변경(새 필드·새 탭·디자인 변경)은 하지 않는다.

## §G 교차 참조

- SPEC-WEB-SAVE-LOSSLESS-001 — t1314 무손실 저장(회귀 방지 대상, plan §A)
- SPEC-MODEL-MATRIX-UPDATE-001 — t1368 감사 핀 3중 정합 선례
- SPEC-V3R6-AUDIT-MODEL-PIN-001 — 핀 우선순위·어휘(REQ-AMP-001..009)
- SPEC-SESSION-WORKTREE-001 / SPEC-WORKTREE-KEY-WIRING-001 — 세션 워크트리 M4 정리·automerge 순서
- AGENTS.md §3 — 워크트리 폐기 규율(스코프 ③의 정책 근거)
- 카드 t1393 · 증거 `.moai/reports/t1393/`
