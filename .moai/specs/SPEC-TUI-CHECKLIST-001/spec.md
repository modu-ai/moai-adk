---
id: SPEC-TUI-CHECKLIST-001
title: "moai init/update 진행 출력을 그룹 체크리스트로 재설계"
version: "0.1.0"
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

## §1 배경과 문제

`moai init` 과 `moai update` 는 단계마다 진행 줄과 결과 줄을 따로 찍는다. 비TTY update 한 번이 stdout 에 60줄을 남기고(`internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:1-60`), init 은 보고기 단계 줄 8개와 11줄짜리 완료 상자를 stderr 에 쌓는다(`internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:1-24`). 무엇이 끝났고 무엇이 문제인지 한 화면에서 읽히지 않는다.

카드가 요구하는 모양은 macOS 정리 도구의 체크리스트다. 머리줄 1개, 환경 요약 1줄, 그룹 제목, 그룹 아래 항목마다 상태 한 줄. 실행 중인 항목만 제자리에서 바뀌고, 상세는 접혀 있다가 실패와 경고만 펼쳐진다. 끝에는 바뀐 수·보존한 수·경고 수·걸린 시간을 담은 요약이 붙는다.

재사용할 부품은 이미 있다. 줄 잔재 없이 제자리에서 다시 그리는 `ProgressLine`(SPEC-V3R6-UPDATE-PROGRESS-001), `StatusIcon`·글리프 상수, `Spinner`, `Theme`·모션 감소 판정, i18n 카탈로그(research.md §1). 새로 만드는 렌더러는 하나뿐이다.

## §2 확정한 결정

| ID | 결정 | 근거 |
|---|---|---|
| D-1 | 비TTY·CI·`--verbose`·`TERM=dumb` 에서는 렌더러를 켜지 않고 지금 출력 경로를 그대로 탄다. 평문 체크리스트를 따로 두지 않는다. | 기존 테스트·CI 로그가 바이트 단위로 그대로 남는다. 리더 확인 요청 항목이다(plan.md §G Q1). |
| D-2 | 그룹 표식은 `◆`, 건너뜀은 `○`. 카드 문구의 `➤`·`–` 는 쓰지 않는다. | `➤`(U+27A4)는 이모지 검사 범위, 둘 다 허용 목록 밖(design.md §1.1). |
| D-3 | 고루틴·티커 없이 이벤트가 올 때만 다시 그린다. | `Spinner` 의 무상태 계약. 동시성 위험을 들이지 않는다. |
| D-4 | 출력은 `cmd.SetOut`·`SetErr` 캡처로 모은다. 항목이 열린 동안의 바이트만 접고, 나머지는 즉시 넘긴다. | init.go 를 거의 건드리지 않고(plan.md §B) 정보 손실을 막는다. |
| D-5 | init 에 `--verbose` 를 새로 둔다. update `--verbose` 는 뜻을 바꾸지 않고 대체 경로를 강제하는 효과만 더한다. | init 에는 지금 플래그가 없다(research.md §2). |
| D-6 | 체크리스트가 켜지면 init 경고 패널과 완료 카드, update 결과 알약을 요약 블록이 대신한다. 대신하는 화면이 담던 데이터는 모두 요약에 옮긴다. | 같은 경고를 두 번 보이면 한 화면에 들어가지 않는다. |
| D-7 | init 요약에는 "보존 수" 를 넣지 않는다. update 요약만 보존 수를 넣는다. | `InitResult` 에 보존 수 필드가 없다(research.md §2). |
| D-8 | 출력 채널은 바꾸지 않는다. init 은 stderr, update 는 stdout. | 상태 채널 규칙과 기존 update 출력 위치(design.md §5). |

## §3 요구사항 (GEARS)

### §3.1 렌더러

- **REQ-TCL-001** — The `internal/tui` package shall provide exactly one new group-checklist renderer that draws a header line, an environment-summary line, group titles, and exactly one status line per item, built from the existing `StatusIcon`/`Glyph*`, `Spinner`, `Theme`, and colour-downsampling primitives.
- **REQ-TCL-002** — **When** an item is running, the renderer shall redraw only that item's line in place, using the existing erase-line prefix, and shall never emit a redraw sequence for a line that already finished.
- **REQ-TCL-003** — **When** an item finishes, the renderer shall end its line with exactly one whitelisted status glyph: `✓` done, `!` done with warnings, `✗` failed, `○` skipped.
- **REQ-TCL-004** — **When** an item fails, the renderer shall expand beneath it, indented, every detail line recorded while that item was open, followed by the failure message.
- **REQ-TCL-005** — **When** a warning is recorded while an item is open, the renderer shall expand that warning beneath the item; the detail lines of an item that finished without warning or failure shall stay collapsed.
- **REQ-TCL-006** — The renderer shall end a completed run with a summary block carrying the changed count, the preserved count where the command has one, the warning count, the elapsed time, and every data field of the completion surface it replaces.
- **REQ-TCL-007** — **While** `NO_COLOR` is set to a non-empty value, the renderer shall emit no SGR colour sequence.
- **REQ-TCL-008** — **While** `MOAI_REDUCED_MOTION` is set to a value other than empty or `0`, the renderer shall draw running items with the static run glyph and shall emit no Braille spinner frame.
- **REQ-TCL-009** — Every user-facing string the renderer draws shall come from a `checklist.*` key present in the en, ko, ja, and zh catalogues, and an unsupported language shall fall back to en.
- **REQ-TCL-010** — The group/item plan shall hold at most 8 items per group, and a clean init or clean update run shall render in at most 24 terminal lines.

### §3.2 CLI 연결과 대체

- **REQ-TCL-011** — **Where** the process stdout and stderr are both terminals, the command writers are those process streams, `CI` is empty, `TERM` is not `dumb`, and `--verbose` is off, `moai init` and the template-sync path of `moai update` shall render the group checklist.
- **REQ-TCL-012** — **When** any activation condition of REQ-TCL-011 does not hold, the command shall emit the same bytes it emits at `cb30e05b3e`.
- **REQ-TCL-013** — `moai init` shall accept `--verbose`; `moai update --verbose` shall keep its current meaning and shall also force the REQ-TCL-012 output path.
- **REQ-TCL-014** — **While** the checklist is active, bytes written to the command writers while an item is open shall be recorded as that item's detail, and bytes written while no item is open shall reach the real writer immediately and unmodified.
- **REQ-TCL-015** — **While** the checklist is active, every warning recorded through the init warning collector or emitted at warn or error severity through the severity-line helper shall appear exactly once in the output, and the consolidated init warning panel shall not be emitted.
- **REQ-TCL-016** — **When** a checklist run ends before the summary — an error return, a wizard cancel, or an early return — the session shall mark the open item failed, write every captured byte not yet shown verbatim to the real writer, and restore the command writers.
- **REQ-TCL-017** — The checklist shall render on stderr for `moai init` and on stdout for `moai update`.
- **REQ-TCL-018** — The checklist session shall not hold an unfinished line while an interactive prompt runs, and shall not change the wizard flow, the survey questions, or their content.
- **REQ-TCL-019** — The checklist shall not change which files are deployed, any reported count, or any command outcome; the summary's numbers shall equal the numbers the legacy surface reports for the same run.
- **REQ-TCL-020** — This SPEC shall not change any file under `internal/template/templates/`.

## §4 제약

- C-1 렌더러 파일은 `internal/tui/checklist.go` 하나다. 다른 tui 파일은 재사용만 하고 고치지 않는다. 예외: 카탈로그 yaml 네 개.
- C-2 init.go 는 plan.md §B 의 R1·R2·R3 세 영역만 고친다. t1966 이 바꾸는 줄과 맞닿은 줄(728, 750, 780-790)은 건드리지 않는다.
- C-3 `.github/workflows/` 를 고치지 않는다.
- C-4 새 코드의 사용자 대상 문자열은 카탈로그 키를 거친다(`internal/tui/i18n.go:81-82` 규칙).
- C-5 테스트 시접을 바꾸는 테스트는 `t.Parallel` 을 쓰지 않고 `t.Cleanup` 으로 원래 값을 되돌린다.
- C-6 기준선 골든은 코드 변경보다 먼저, 별도 커밋으로 들어간다(design.md §8).

## §5 범위 밖

### Out of Scope — 다른 명령과 경로

- `moai doctor`, `moai status`, `moai bundle add` 등 init·update 밖의 명령 출력
- update 의 clean-install 경로, 바이너리 갱신, `--check`, `--shell-env`, `--config`(reconfigure) 경로의 항목화 — 이 경로의 출력은 넘김 규칙으로 지금 그대로 보인다
- init 의 사용자 자산 보장 단계(`init.go:780-790`) 항목화 — t1966 착지 뒤 후속 카드

### Out of Scope — 위저드와 질문

- init 위저드, update 병합 확인, 프로필 위저드의 흐름·질문·문구·스타일
- 대화형 입력 화면을 체크리스트 안에 넣는 일

### Out of Scope — 출력 형식 확장

- 비TTY 용 평문 체크리스트(결정 D-1, 리더 확인 대기)
- JSON 출력 모드, 상세 로그 파일 저장
- 애니메이션 티커, 진행 막대, 남은 시간 추정
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
| AC-TCL-011 | REQ-TCL-010 | M4, M5 |
| AC-TCL-012 | REQ-TCL-011 | M3 |
| AC-TCL-013 | REQ-TCL-012 | M0, M4, M5 |
| AC-TCL-014 | REQ-TCL-013 | M0, M3, M5 |
| AC-TCL-015 | REQ-TCL-014 | M3 |
| AC-TCL-016 | REQ-TCL-015 | M4 |
| AC-TCL-017 | REQ-TCL-016 | M3, M4 |
| AC-TCL-018 | REQ-TCL-017 | M4, M5 |
| AC-TCL-019 | REQ-TCL-018 | M4, M5 |
| AC-TCL-020 | REQ-TCL-019 | M4, M5 |
| AC-TCL-021 | REQ-TCL-020 | 전 구간 |
| AC-TCL-022 | 제약 C-2 | M4 |
| AC-TCL-023 | REQ-TCL-001, REQ-TCL-003 | M2 |
