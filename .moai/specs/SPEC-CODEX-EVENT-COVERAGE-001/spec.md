---
id: SPEC-CODEX-EVENT-COVERAGE-001
title: "Codex hook 이벤트 커버리지 — Interrupt 12번째 행 추가(M1) + 미adapted 5종+Interrupt 런타임 발화 실측 캠페인(M2)"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: internal/codexadapter
lifecycle: spec-anchored
tier: M
tags: "codex, hooks, event-table, interrupt, measurement-campaign, codex-cli-0.153.4, t496"
related_specs: [SPEC-CODEX-HOOK-ADAPTER-001, SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-EVENT-COVERAGE-001 — Codex hook 이벤트 커버리지 (Interrupt 행 + 발화 실측 캠페인)

> 카드: **t496** (Class C) · 선행: `SPEC-CODEX-HOOK-ADAPTER-001`(어댑터·REQ-7 불변), `SPEC-CODEX-WIRING-001`(hooks.json 설치)
> 조사 입력: `.moai/reports/t494/codex-doc-survey.md` §3 (sibling worktree t494 — **본 트리·primary 체크아웃 어디에도 이 파일이 존재하지 않는다**). 본 트리에서 재측정한 것은 §C의 M1-M5(트리 내 사실)뿐이며, **공식 12종 열거는 t494 §3이 인용한 외부 문서 주장으로서 본 트리에서 재측정 불가**하다.

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-09-07 | 최초 작성 (plan-phase, 카드 t496). Axis A(M1, Interrupt 행 추가) + Axis B(M2, 발화 실측 캠페인) 이축 구조. 설계 결정 4건은 plan.md §D에 기재 |

## A. 배경

moai의 Codex 어댑터는 `internal/codexadapter/events.go`의 `EventTable`로 Codex 훅 이벤트를 MoAI 디스패처 서브커맨드에 사상한다. 두 개의 독립 축이 있다.

- **Axis A (M1)** — Codex 공식 문서가 훅 이벤트 12종을 열거한다(SessionStart, SessionEnd, SubagentStart, SubagentStop, PreToolUse, PermissionRequest, PostToolUse, PreCompact, PostCompact, UserPromptSubmit, Stop, Interrupt). **이 열거의 근거는 t494 §3이다(외부 문서 주장 — 본 트리 재측정 불가).** moai의 EventTable은 11종을 열거하며(본 트리에서 재측정, §C M2), 누락은 t494 열거 대비 정확히 `Interrupt` 하나다.
- **Axis B (M2)** — EventTable의 `Adapted=false` 5종(PreCompact, PostCompact, PermissionRequest, SubagentStart, SubagentStop)은 "counterpart 부재" 주장이 아니라 **측정 커버리지 결정**(events.go:38-51 doc comment, 0.147.0 시대 `.moai/reports/t83/precondition-measurement.md`)이다. 이 5종+Interrupt가 codex-cli 0.153.4에서 실제로 발화하는지를 실행 기반으로 재측정하고, 이벤트별로 적응(extension) 또는 미발화 문서화를 결정한다.

### 본 트리 재측정 근거 (2026-09-07, 이 워크트리에서 실행)

| # | 주장 | 명령 | 관측 |
|---|---|---|---|
| M1 | `EventInterrupt` 상수가 internal/에 없다 | `grep -rn 'EventInterrupt' internal/` | 매치 0행 (rc=1). 대조군 `EventStop` → `internal/hook/types.go:34` 등 다수 검출 |
| M2 | EventTable은 6 adapted + 5 unadapted = 11행 | `internal/codexadapter/events.go:51-64` 직독 | `{...true}` 6행·`{...false}` 5행, 주장과 일치 |
| M3 | 디스패처에 interrupt 서브커맨드 없음 | `grep -n 'interrupt\|Interrupt' internal/cli/hook.go` | 매치 0행 |
| M4 | 설치기는 Adapted=false 행을 건너뛴다 | `internal/codexwiring/hooks.go:97-100` 직독 | `if !row.Adapted { continue }` 존재 확인 |
| M5 | 이번 머신 codex-cli | `codex --version` | `codex-cli 0.153.4` |
| M6 | 공식 12종 열거의 근거 | — (측정 아님) | t494 §3의 외부 문서 주장을 채택. sibling worktree 밖에서 그 파일이 존재하지 않으므로 본 트리 재측정 불가 — M2의 실행 기반 캠페인이 이 전제의 설계된 후속 검증이다 |

## B. 요구 (GEARS)

### B.1 Axis A — Interrupt 이벤트 (M1)

- **REQ-CEV-001 (Ubiquitous)** — `EventTable` shall carry exactly 12 rows: the existing 11 plus one row whose `CodexEvent` is `"Interrupt"`, `DispatcherArg` is the empty string, and `Adapted` is false.
- **REQ-CEV-002 (Ubiquitous)** — The Interrupt event-name constant shall be defined in package `internal/codexadapter` (e.g. `CodexEventInterrupt hook.EventType = "Interrupt"`), and no `EventInterrupt` constant shall be added to `internal/hook`. 근거: `events.go:8-9`의 불변 — "this package sits in FRONT of the dispatcher rather than inside it, and nothing under internal/hook is modified (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7)". additive 상수라도 REQ-7의 문자 위반이고, internal/hook은 Claude 측 이벤트 어휘이므로 Codex 전용 이름이 들어가면 의도도 위반한다. `hook.EventType`은 string 타입이므로 어댑터가 직접 생성할 수 있다.
- **REQ-CEV-003 (Event-driven)** — **When** `Resolve` receives `"Interrupt"`, it shall return an error wrapping `ErrUnadapted` whose message does NOT assert that a dispatcher argument exists (the current format `(dispatcher arg %q exists; ...)` is false for an empty arg). `ErrUnadapted`의 의도적 재사용을 채택한다 — "recognized but not adapted by this milestone"은 Interrupt에도 여전히 참이고(공식 12종에 인식됨, 본 마일스톤이 적응 안 함), 소비자가 두 refusal을 구별할 필요가 현재 없어 3번째 센티널은 API 표면만 늘린다(plan.md §D1). 메시지가 뉘앙스를 운반한다.
- **REQ-CEV-004 (Event-driven)** — **When** `RenderHooks` renders `.codex/hooks.json`, it shall not install any handler for the Interrupt row (기존 `!row.Adapted` skip 경로가 그대로 지켜야 한다 — M1이 설치 표면을 바꾸지 않는다).
- **REQ-CEV-005 (Unwanted)** — M1 shall not register any new `moai hook` dispatcher subcommand, and shall not modify any file under `internal/hook/`.
- **REQ-CEV-006 (Ubiquitous)** — The doc comments in `events.go` shall state the truth after the change: the table carries 12 rows, one of which (Interrupt) has no MoAI dispatcher counterpart, and the adapted/held-back census matches the rows.

### B.2 Axis B — 발화 실측 캠페인 (M2)

- **REQ-CEV-007 (Ubiquitous)** — The campaign shall measure, by execution on the installed codex-cli (measured `codex-cli 0.153.4`, M5 above), whether each of the 6 events — PreCompact, PostCompact, PermissionRequest, SubagentStart, SubagentStop, Interrupt — fires, with one named trigger method and one per-event capture log per event.
- **REQ-CEV-008 (Unwanted)** — The campaign shall not mutate the operator's real `~/.codex` configuration. 격리는 `CODEX_HOME` 임시 재배치 우선(지원 여부를 캠페인 전제 P0으로 검증), 미지원 시 scratch repo의 project-scope `.codex/hooks.json` + trust 절차로 폴백한다.
- **REQ-CEV-009 (Unwanted)** — "공식 문서 부재"는 "발화하지 않는다"의 근거가 아니다. The campaign record shall base every fired/not-fired verdict on an executed command and its observed output, never on documentation omission (docs omission과 runtime absence는 다른 사실이다).
- **REQ-CEV-010 (Event-driven)** — **When** an event's verdict is "fires", the campaign record shall specify the adapter-extension decision (payload shape observed + dispatcher mapping decision: adapt-now / follow-up card); **When** the verdict is "does not fire", the record shall carry the not-fire documentation with the measured evidence (command + observed output).
- **REQ-CEV-011 (Ubiquitous)** — The campaign record shall re-test the prior "SubagentStop does not fire" observation (0.147.0-era, delegation surfacing as PostToolUse with `tool_name` prefix "collaboration") on 0.153.4 rather than carrying it forward.
- **REQ-CEV-012 (Ubiquitous)** — M1 shall be independently closable: the artifact structure shall allow M1 to close alone if M2 balloons (the lane reports a split to the lead in that case).

## C. 제약

- 구현 최소성: M1의 최소 변경은 테이블 1행 + 어댑터 내 이름 상수 1개 + Resolve 메시지 조건 처리 + 테스트 파일 3종 갱신/신설(plan M1 파일 목록 참조) + doc comment 수정이다. 이 개념적 크기의 ~3배를 넘는 plan은 과잉 설계다.
- `code_comments: en` — 모든 코드 주석/ doc comment는 영어.
- 캠페인은 런 페이즈에서 실행하며, plan 페이즈에서 이를 검증 실행하지 않는다.
- 측정 귀속: 캠페인 판정 each는 실행한 명령 + 관측 출력과 함께 기록된다(verification-claim-integrity §2).

## D. 마일스터리 요약

| 마일스톤 | 축 | 내용 | 우선순위 |
|---|---|---|---|
| M1 | A | Interrupt 행 추가 + 어댑터 상수 + Resolve 메시지 + 테스트/doc 갱신 | High |
| M2 | B | 0.153.4 발화 실측 캠페인 (6종) → `.moai/reports/t496/codex-event-campaign.md` 판정+처분 기록 | High |
| M3 | B(조건부) | M2 처분이 adapt-now로 표시한 이벤트만 테이블+사상 확장. 없으면 no-op으로 M2 기록이 종결 산출물 | Medium |

상세: plan.md §F. 수용기준 전문: acceptance.md.

### AC→REQ 추적 (전 REQ 커버)

- AC-CEV-001 → REQ-CEV-001 · AC-CEV-002 → REQ-CEV-002, REQ-CEV-005 · AC-CEV-003 → REQ-CEV-003 · AC-CEV-004 → REQ-CEV-001, REQ-CEV-003 · AC-CEV-005 → REQ-CEV-004 · AC-CEV-006 → REQ-CEV-006
- AC-CEV-010 → REQ-CEV-007, REQ-CEV-009 · AC-CEV-011 → REQ-CEV-008 · AC-CEV-012 → REQ-CEV-011 · AC-CEV-013 → REQ-CEV-010 · AC-CEV-020 → REQ-CEV-012

## E. 관련 SPEC

- `SPEC-CODEX-HOOK-ADAPTER-001` — 어댑터 패키지 창설, REQ-7 불변("nothing under internal/hook is modified")의 출처
- `SPEC-CODEX-WIRING-001` — hooks.json 렌더/설치(REQ-CW-005 merge model, REQ-CW-003 whitelist gate)
- 입력 리포트: `.moai/reports/t494/codex-doc-survey.md` §3 (12종 열거의 근거 — sibling worktree 존재, 본 트리 무존재) · `.moai/reports/t83/precondition-measurement.md` (0.147.0 캠페인 선례). §C M1-M5의 트리 내 사실만 본 SPEC이 재측정했다.

## F. Out of Scope

### Out of Scope — 새 이벤트의 Claude 측 디스패처 확장
- `moai hook interrupt` 등 Claude 측에 대응 훅이 없는 이벤트를 위한 신규 디스패처 서브커맨드·핸들러 등록 (REQ-CEV-005)
- internal/hook에 `EventInterrupt` 상수 추가 (REQ-CEV-002)
- Interrupt 발화 시 그것을 MoAI 핸들러로 적응하는 것 — M2의 발화 판정과 payload 관측까지만 본 SPEC이 수행하고, 적응 여부는 M2 처분 기록이 결정(adapt-now면 M3, 아니면 후속 카드)

### Out of Scope — 캠페인 범위 밖 이벤트
- 이미 Adapted=true인 6종(PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop, UserPromptSubmit)의 재측정 — 0.147.0에서 payload golden까지 확보된 측정 기반
- Claude Code 측 훅 표면의 변경
- `internal/hook` 디스패처·레지스트리 로직의 변경 (`registry.go` 스위치 포함)

### Out of Scope — 문서·배포
- docs-site / README의 Codex 훅 문서 갱신 (sync 페이즈 소관 판단은 manager-docs)
- 코드 구현 이외의 템플릿(`internal/template/templates/`) 변경 — EventTable은 런타임 Go 코드이지 템플릿이 아니다
