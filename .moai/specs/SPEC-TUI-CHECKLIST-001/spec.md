---
id: SPEC-TUI-CHECKLIST-001
title: "moai init/update 진행 출력을 그룹 체크리스트로 재설계"
version: "0.1.2"
status: draft
created: 2026-10-27
updated: 2026-10-27
author: Goos Kim
priority: P1
phase: "v3.3.0 target"
module: "internal/tui, internal/cli"
lifecycle: spec-anchored
tags: "tui, init, update, progress, checklist, i18n, t1969"
tier: L
related_specs: [SPEC-V3R6-UPDATE-PROGRESS-001, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-V3R6-UPDATE-NOISE-001, SPEC-INIT-QUIET-WIZARD-001]
---

## HISTORY

- 2026-10-27 — v0.1.0 초안 (manager-spec, 카드 t1969 plan 단계, Tier L). 근거는 같은 디렉터리의 research.md(파일:행 사실, HEAD `cb30e05b3e`)와 design.md(렌더러 API·이벤트 모델·대체 규칙·i18n 키·골든 전략)다. 형제 카드 t1966 의 init.go 변경 범위는 공유 브랜치 ref `WT-moai-init-moai`(`d7f4d5c288`)에서 읽었다(research.md §7).
- 2026-10-27 — v0.1.1 plan 델타 1회차 (manager-spec). 1회차 plan 감사(FAIL 0.71, 감사 HEAD `d241917d17`)의 결함 D1~D20 을 고쳤다. REQ 20 → 23 (REQ-TCL-021 init.go 변경 영역 — 제약 C-2 를 REQ 로 올림, REQ-TCL-022 열린 항목 안 경고·오류·안내 줄 보존, REQ-TCL-023 성공 조기 반환은 정상 종료), AC 23 → 24 (AC-TCL-024 보존 줄 주입 시험). 경고 수 정의를 "수집기 수" 하나로 정했다(D6). 레인 결정 Q1~Q3 을 반영해 "리더 확인 대기" 문구를 지웠다(D16). 대체 골든을 `internal/cli/testdata/checklist/fallback/` 하위 디렉터리로 옮겨 기준선 선행 증거를 디렉터리 단위로 잰다(D2). 증거 원장에 L15~L19 를 더했다(측정 HEAD `d241917d17`).
- 2026-10-27 — v0.1.2 plan 델타 2회차 (manager-spec). 2회차 plan 감사(FAIL 0.75, 감사 HEAD `b422a13d53`)의 결함 N1~N13 과 리더 codex_audit 의 P2(= N8)를 고쳤다. REQ·AC 수는 그대로(23·24). 보존 줄 목록을 `internal/cli` 전체(하위 패키지 포함)에 대한 고정 검색 명령의 출력으로 정하고 23행으로 늘렸으며(N1), AC-TCL-024 는 행마다 하위 테스트 하나와 목록 완전성 하위 테스트를 갖는다(N2). 재개 경로를 실제 실행기의 이벤트 순서(검증 단계 `StepError` 가 먼저)에 맞춰 다시 썼다(N3). update 요약이 그려지는 지점을 템플릿 동기화 본문 끝으로 정했다(결정 D-10). 순서 증거에 O4(M0 트리 재실행)와 카드 id 완전성 검사를 더했다(N6·N7). 증거 원장에 L20~L25 를 더했다(측정 HEAD `b422a13d53`).

## §1 배경과 문제

`moai init` 과 `moai update` 는 단계마다 진행 줄과 결과 줄을 따로 찍는다. 비TTY update 한 번이 stdout 에 60줄을 남기고(`internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:1-60`), init 은 보고기 단계 줄 8개와 11줄짜리 완료 상자를 stderr 에 쌓는다(`internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:1-24`). 무엇이 끝났고 무엇이 문제인지 한 화면에서 읽히지 않는다.

카드가 요구하는 모양은 macOS 정리 도구의 체크리스트다. 머리줄 1개, 환경 요약 1줄, 그룹 제목, 그룹 아래 항목마다 상태 한 줄. 실행 중인 항목만 제자리에서 바뀌고, 상세는 접혀 있다가 실패와 경고만 펼쳐진다. 끝에는 바뀐 수·보존한 수·경고 수·걸린 시간을 담은 요약이 붙는다.

재사용할 부품은 이미 있다. 줄 잔재 없이 제자리에서 다시 그리는 `ProgressLine`(SPEC-V3R6-UPDATE-PROGRESS-001), `StatusIcon`·글리프 상수, `Spinner`, `Theme`·모션 감소 판정, i18n 카탈로그(research.md §1). 새로 만드는 렌더러는 하나뿐이다.

## §2 확정한 결정

D-1·D-6·범위 밖 Q3 은 레인 판단(사다리 마지막 단계)으로 정했다. 되돌릴 수 있는 결정이며, Kickoff 기록에서 리더에게 알린다. 근거 기록은 `.moai/reports/t1969/progress.md` 의 결정 기록이다.

| ID | 결정 | 근거 |
|---|---|---|
| D-1 | 비TTY·CI·`--verbose`·`TERM=dumb` 에서는 렌더러를 켜지 않고 기준선 트리의 출력 경로를 그대로 탄다. 평문 체크리스트를 따로 두지 않는다. | 카드 문구 "TTY 아님/CI/--verbose 면 평문 줄 출력으로 폴백"을 "오늘의 평문 줄 출력으로 폴백"으로 읽었다. 오늘 비TTY 출력은 이스케이프 없는 평문 줄이라 문구를 문자 그대로 충족하고, 기존 버퍼 기반 테스트·CI 로그가 바이트 단위로 남는다. **다른 독해**: 카드가 `--verbose` 만 따로 "현재 상세 그대로"라고 적었으므로 비TTY·CI 를 평문 체크리스트(새 모양)로 바꾸라는 뜻일 수도 있다. 리더가 그 독해를 택하면 REQ-TCL-012·AC-TCL-013 과 M0 기준선을 다시 쓰는 델타 회차를 연다(plan.md §G). |
| D-2 | 그룹 표식은 `◆`, 건너뜀은 `○`. 카드 문구의 `➤`·`–` 는 쓰지 않는다. | `➤`(U+27A4)는 이모지 검사 범위, 둘 다 허용 목록 밖(design.md §1.1). Kickoff 기록에서 리더에게 알린다. |
| D-3 | 고루틴·티커 없이 이벤트가 올 때만 다시 그린다. | `Spinner` 의 무상태 계약. 동시성 위험을 들이지 않는다. |
| D-4 | 출력은 `cmd.SetOut`·`SetErr` 캡처로 모은다. 항목이 열린 동안의 바이트만 접고, 나머지는 즉시 넘긴다. 접히면 안 되는 줄(보존 줄 목록, design.md §3.3.1)은 따로 펼친다. | init.go 를 거의 건드리지 않고(plan.md §B) 정보 손실을 막는다. |
| D-5 | init 에 `--verbose` 를 새로 둔다. update `--verbose` 는 뜻을 바꾸지 않고 대체 경로를 강제하는 효과만 더한다. | init 에는 지금 플래그가 없다(research.md §2). |
| D-6 | 체크리스트가 켜지면 init 경고 패널과 완료 카드, update 결과 알약을 요약 블록이 대신한다. 대신하는 화면이 담던 데이터는 모두 요약에 옮긴다. 항목 밖에서 기록된 수집기 경고는 패널 대신 요약 블록이 다시 나열한다. | 같은 경고를 항목 아래와 패널에 두 번 보이면 한 화면에 들어가지 않는다. 항목 밖 경고(예: 위저드 전 `init.go:427`)는 위저드 화면 위로 밀려나므로 요약이 다시 보여 준다. |
| D-7 | init 요약에는 "보존 수" 를 넣지 않는다. update 요약만 보존 수를 넣는다. | `InitResult` 에 보존 수 필드가 없다(research.md §2). |
| D-8 | 출력 채널은 바꾸지 않는다. init 은 stderr, update 는 stdout. | 상태 채널 규칙과 기존 update 출력 위치(design.md §5). |
| D-9 | 요약의 경고 수는 init 경고 수집기의 수(`warnCollector.Count()`, `internal/cli/init_warnings.go:59-63`)다. 심각도 줄(`emitSeverityLine`)과 보존 줄은 펼쳐 보이되 수에 더하지 않는다. | 기존 완료 카드도 같은 수를 보고한다(`init.go:1087` `p.Count()`). 수를 바꾸면 "보고하는 수를 바꾸지 않는다"(REQ-TCL-019)를 어긴다. |
| D-10 | 열린 항목은 다음 항목이 열릴 때, 요약이 그려질 때, 세션이 끝날 때, 그리고 단계 보고기가 실패를 알릴 때만 닫힌다. update 요약은 템플릿 동기화 본문(`runTemplateSyncWithReporter`)이 `return nil`(`internal/cli/update_template_sync.go:1218`) 하기 직전, 즉 `warnStampSkew`(`:1216`) 뒤에 한 번 그린다. 단계 루프 뒤의 꼬리(`:1081-1216`)는 마지막으로 열린 항목 안에서 돈다. | 꼬리의 훅 설치 줄(`internal/cli/hook_install.go:262` — update 마다 나옴)이 요약 뒤로 넘어가면 깨끗한 update 가 24줄을 넘는다. 꼬리의 직접 쓰기 경고(`update_migrate.go:154`·`:161`, `update_template_sync_guard.go:105`)는 보존 줄 목록이 지킨다(design.md §3.2·§3.3.1). |

## §3 요구사항 (GEARS)

### §3.1 렌더러

- **REQ-TCL-001** — The `internal/tui` package shall provide exactly one new group-checklist renderer that draws a header line, an environment-summary line, group titles, and exactly one status line per item, built from the existing `StatusIcon`/`Glyph*`, `Spinner`, `Theme`, and colour-downsampling primitives.
- **REQ-TCL-002** — **When** an item is running, the renderer shall redraw only that item's line in place, using the existing erase-line prefix, and shall never emit a redraw sequence for a line that already finished.
- **REQ-TCL-003** — **When** an item finishes, the renderer shall end its line with exactly one whitelisted status glyph: `✓` done, `!` done with warnings, `✗` failed, `○` skipped.
- **REQ-TCL-004** — **When** an item fails, the renderer shall expand beneath it, indented, every detail line recorded while that item was open, followed by the failure message.
- **REQ-TCL-005** — **When** a warning is recorded while an item is open, the renderer shall expand that warning beneath the item; the detail lines of an item that finished without warning or failure shall stay collapsed.
- **REQ-TCL-006** — The renderer shall end a completed run with a summary block carrying the changed count, the preserved count where the command has one, the warning count defined by decision D-9, the elapsed time, and every data field of the completion surface it replaces.
- **REQ-TCL-007** — **While** `NO_COLOR` is set to a non-empty value, the renderer shall emit no SGR colour sequence.
- **REQ-TCL-008** — **While** `MOAI_REDUCED_MOTION` is set to a value other than empty or `0`, the renderer shall draw running items with the static run glyph and shall emit no Braille spinner frame.
- **REQ-TCL-009** — Every user-facing string the renderer draws shall come from a `checklist.*` key present in the en, ko, ja, and zh catalogues, the ja and zh values shall differ from the en value except for an allow-listed set of identifier-only values, and an unsupported language shall fall back to en.
- **REQ-TCL-010** — The group/item plan shall hold at most 8 items per group, and a clean init run and a clean update run, as defined in §3.3, shall each render in at most 24 terminal lines.

### §3.2 CLI 연결과 대체

- **REQ-TCL-011** — **While** the process stdout and stderr are both terminals, the command writers are those process streams, `CI` is empty, `TERM` is not `dumb`, and `--verbose` is off, `moai init` and the template-sync path of `moai update` shall render the group checklist; the five inputs shall be gathered by one function that both commands call.
- **REQ-TCL-012** — **When** any activation condition of REQ-TCL-011 does not hold, the command shall emit the same bytes the baseline tree emits — the tree the fallback goldens were captured from (`cb30e05b3e`, or the re-baseline tree recorded under plan.md §B).
- **REQ-TCL-013** — `moai init` shall accept `--verbose`; `moai update --verbose` shall keep its current meaning, including its verbose-only stderr lines, and shall also force the REQ-TCL-012 output path.
- **REQ-TCL-014** — **While** the checklist is active, bytes written to the command writers while an item is open shall be recorded as that item's detail, and bytes written while no item is open shall reach the real writer immediately and unmodified.
- **REQ-TCL-015** — **While** the checklist is active, every warning recorded through the init warning collector or emitted at warn or error severity through the severity-line helper while an item is open shall appear exactly once, expanded under that item; every collector warning recorded while no item is open shall keep its legacy immediate line, if it had one, and shall be listed once in the summary block; and the consolidated init warning panel shall not be emitted.
- **REQ-TCL-016** — **When** a checklist run ends before the summary with a non-nil error or a wizard cancel, the session shall mark the open item failed, write every captured byte not yet shown verbatim to the real writer, and restore the command writers.
- **REQ-TCL-017** — The checklist shall render on stderr for `moai init` and on stdout for `moai update`.
- **REQ-TCL-018** — **While** an interactive prompt runs, the checklist session shall hold no unfinished line, and the wizard flow, the survey questions, and their content shall stay byte-identical to the baseline tree.
- **REQ-TCL-019** — The deployed file set, every reported count, and every command outcome shall equal those of the legacy surface for the same run, and the summary's warning count shall equal the collector count the legacy completion card reports.
- **REQ-TCL-020** — Every file under `internal/template/templates/` shall stay byte-identical to the baseline tree.
- **REQ-TCL-021** — This card's commits shall change `internal/cli/init.go` only inside the regions R1, R2, and R3 of plan.md §B, measured against the region base recorded there, and R3 changes shall be pure insertions.
- **REQ-TCL-022** — The preserved-line inventory of design.md §3.3.1 shall classify every hit of its fixed source sweep over `internal/cli` and its subpackages, and **while** the checklist is active, every inventoried line — direct-write warnings, error lines, and user-action notices that bypass the warning signal — shall appear exactly once on the real writer even when its enclosing item finishes successfully; a preserved warning or error line written while an item is open shall end that item with `!`, and a preserved notice shall change neither the item glyph nor the warning count.
- **REQ-TCL-023** — **When** a checklist run returns early with a nil error, the session shall close the open item, if one is open, as done, write every captured preserved line not yet shown, leave every finished line — a failed one included — as drawn, and restore the command writers.

### §3.3 깨끗한 실행의 정의 (REQ-TCL-010)

- 깨끗한 init: 빈 임시 디렉터리에 처음 하는 init, 이전 `.moai` 없음, git 있음, 세션 워크트리 꺼짐, 지연 갱신 확인이 "새 버전 없음", 모든 꼬리 단계 성공, 수집기 경고 0, 보존 줄 0.
- 깨끗한 update: 같은 바이너리로 init 한 프로젝트, 사용자 수정 없음, 바이너리 갱신 없음, 병합 충돌·병합 대체 없음, 종료 행동 장부가 비어 있음(`internal/cli/update_action_block.go:89-91` 이 아무것도 내지 않는 조건).
- 줄 수에 넣는 것: 명령이 두 작성기에 쓴 모든 줄 — 넘김 규칙으로 보이는 줄(update 정체성 띠 `internal/cli/update_tux.go:77-82` 포함), 체크리스트 머리·환경·그룹·항목 줄, 요약 블록. `\r\x1b[2K` 로 덮인 줄은 최종 화면 한 줄로 접어 센다.

## §4 제약

- C-1 렌더러 파일은 `internal/tui/checklist.go` 하나다. 다른 tui 파일은 재사용만 하고 고치지 않는다. 예외: 카탈로그 yaml 네 개.
- C-2 init.go 변경 영역은 REQ-TCL-021 로 올렸다. t1966 이 바꾸는 줄과 맞닿은 줄(728, 750, 780-790)은 건드리지 않는다.
- C-3 `.github/workflows/` 를 고치지 않는다.
- C-4 새 코드의 사용자 대상 문자열은 카탈로그 키를 거친다(`internal/tui/i18n.go:81-82` 규칙).
- C-5 테스트 시접을 바꾸는 테스트는 `t.Parallel` 을 쓰지 않고 `t.Cleanup` 으로 원래 값을 되돌린다.
- C-6 대체 골든과 그 테스트는 코드 변경보다 먼저, 그 둘만 담은 커밋 하나(M0)로 들어간다(design.md §8, AC-TCL-013).

## §5 범위 밖

### Out of Scope — 다른 명령과 경로

- `moai doctor`, `moai status`, `moai bundle add` 등 init·update 밖의 명령 출력
- update 의 clean-install 경로, 바이너리 갱신, `--check`, `--shell-env`, `--config`(reconfigure) 경로의 항목화 — 이 경로의 출력은 넘김 규칙으로 지금 그대로 보인다
- init 의 사용자 자산 보장 단계(`init.go:780-790`) 항목화 — 레인 결정 Q3: t1966 이 `:775-780` 을 바꾸므로 그 카드 착지 뒤 후속 카드로 넘긴다. 이 단계는 `executor.Execute`(`init.go:829`)보다 먼저 돌아 열린 항목이 없으므로 출력이 넘김 규칙으로 지금 그대로 보인다

### Out of Scope — 위저드와 질문

- init 위저드, update 병합 확인, 프로필 위저드의 흐름·질문·문구·스타일
- 대화형 입력 화면을 체크리스트 안에 넣는 일

### Out of Scope — 출력 형식 확장

- 비TTY 용 평문 체크리스트(결정 D-1, 레인 결정 Q1. 다른 독해를 택하면 델타 회차로 다시 연다)
- JSON 출력 모드, 상세 로그 파일 저장
- 애니메이션 티커, 진행 막대, 남은 시간 추정
- 경고 수 정의 변경(심각도 줄·보존 줄을 경고 수에 더하는 일, 결정 D-9)
- ja·zh 문구의 원어민 검수(초안은 이 카드가 넣는다)
- en·ko·ja·zh 밖 11개 언어 카탈로그

### Out of Scope — 배포·워크플로

- `internal/template/templates/` 아래 배포 템플릿
- `.github/workflows/` 와 CI 설정

## §6 추적표

AC 와 REQ 의 대응 선언은 acceptance.md 각 기준 제목의 `maps` 목록이 원본이다. 이 표는 마일스톤을 함께 보이기 위한 사본이다.

| AC | 대응 REQ | 마일스톤 |
|---|---|---|
| AC-TCL-001 | REQ-TCL-001, REQ-TCL-003 | M2 |
| AC-TCL-002 | REQ-TCL-002 | M2 |
| AC-TCL-003 | REQ-TCL-004 | M2 |
| AC-TCL-004 | REQ-TCL-005 | M2 |
| AC-TCL-005 | REQ-TCL-006 | M2 |
| AC-TCL-006 | REQ-TCL-007 | M2 |
| AC-TCL-007 | REQ-TCL-008 | M2 |
| AC-TCL-008 | REQ-TCL-009 | M1 |
| AC-TCL-009 | REQ-TCL-009 | M2 |
| AC-TCL-010 | REQ-TCL-010 | M3 |
| AC-TCL-011 | REQ-TCL-010, REQ-TCL-009 | M4, M5 |
| AC-TCL-012 | REQ-TCL-011 | M3, M4, M5 |
| AC-TCL-013 | REQ-TCL-012 | M0, M4, M5 |
| AC-TCL-014 | REQ-TCL-013 | M0, M3, M5 |
| AC-TCL-015 | REQ-TCL-014 | M3 |
| AC-TCL-016 | REQ-TCL-015 | M4 |
| AC-TCL-017 | REQ-TCL-016, REQ-TCL-023 | M3, M4 |
| AC-TCL-018 | REQ-TCL-017 | M4, M5 |
| AC-TCL-019 | REQ-TCL-018 | M4, M5 |
| AC-TCL-020 | REQ-TCL-019 | M4, M5 |
| AC-TCL-021 | REQ-TCL-020 | 전 구간 |
| AC-TCL-022 | REQ-TCL-021 | M4 |
| AC-TCL-023 | REQ-TCL-001, REQ-TCL-003, REQ-TCL-009 | M2 |
| AC-TCL-024 | REQ-TCL-022 | M4, M5 |
