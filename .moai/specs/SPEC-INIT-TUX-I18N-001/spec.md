---
id: SPEC-INIT-TUX-I18N-001
title: "init/update/profile wizard TUX repair — v1 profile wizard absorbed into huh v2, layout repair, remaining English surfaces localized"
version: "0.1.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, tux, wizard, huh-v2, i18n, init, update, profile, layout, pty"
tier: M
related_specs: [SPEC-CLI-TUX-V3-002, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-WEB-CONSOLE-002, SPEC-WEB-CONSOLE-003, SPEC-I18N-GOVERNANCE-001]
---

# SPEC-INIT-TUX-I18N-001 — init/update/profile 위저드 TUX 수리와 잔여 영어 표면 현지화

> 카드: **t586** (Factory lane-2, Class C). 재현 근거: `.moai/reports/t586/verdict.md`.

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-11 | manager-spec | 최초 작성(카드 t586 plan 단계). 재현 판정(`.moai/reports/t586/verdict.md`)과 리드·운영자 결정 D1~D6을 요구사항으로 옮겼다. 프로필 없음 확인창 제거(D1), huh v1 프로필 위저드를 huh v2 위저드로 흡수(D2), 확인창 안쪽 빈 줄 유지(D3), 단계 표시 줄 소실 조사 연기(D4), 옵션 설명 열 표시 폭 기준 정렬(D5), pty 캡처 하네스와 View() 골든 이원 판정(D6)을 반영했다. |

## §A 배경

### A.1 무엇이 문제인가

`moai init`·`moai update`·`moai profile setup` 의 대화형 화면에는 두 종류의 결함이 있다. 재현과 측정은 모두 판정 문서에 있고, 이 SPEC 은 그 측정을 기준선으로 삼는다(수치는 판정 문서에서 읽은 것만 옮긴다).

- **F11 렌더 결함** — 확인 버튼 중앙정렬(들여쓰기 v2 7칸, v1 22칸), 필드 사이 빈 줄(`FieldSeparator` 기본값 `"\n\n"`), 선택 필드 아래 빈 카드 줄(첫 페이지 4줄), 옵션 설명 열 불일치(`Label - Desc` 단순 연결). 네 가지 모두 뷰 문자열과 tmux 80×30 pty 캡처 양쪽에서 재현됐다.
- **F10 영어 고정 표면** — 판정 문서의 I1~I7. 프로필 없음 확인창 제목·설명(I1·I2), v1 확인창 버튼 `Yes`/`No`(I3), v1 프로필 위저드 1단계 언어 선택(I4), 1단계 취소 메시지(I5), 버전 다운그레이드 확인창(I6, 카드 원문에 없던 대상), v1·v2 전 표면의 도움말 줄(I7).

### A.2 결정으로 좁혀진 범위

리드·운영자 결정이 대상 목록을 다음과 같이 바꾼다.

| 판정 대상 | 결정 반영 후 |
|---|---|
| I1·I2 프로필 없음 확인창 | **사라짐** (D1 — 확인창 자체를 없앤다) |
| I3 v1 확인창 버튼 | **사라짐** (D1 로 해당 확인창이 없어짐) |
| I4·I5 v1 프로필 위저드 1단계 | **흡수로 해소** (D2 — v2 위저드의 언어 질문과 취소 경로가 대신한다) |
| I6 다운그레이드 확인창 | **현지화 대상으로 남음** — 이 SPEC 은 이 확인창도 v2 로 옮기기로 정한다(§A.4, §B.3) |
| I7 도움말 줄 | **현지화 대상으로 남음** (F11 의 키맵 조정과 함께) |

### A.3 v1 호출 지점 전수 조사 (D2-a)

트리 `e7a7d4bb3`(워크트리 `WT-init-tux-i18n`)에서 grep 으로 다시 셌다.

`runProfileSetup` 호출·등록 지점 — 4곳, 이 밖에 없음:

| 위치 | 성격 |
|---|---|
| `internal/cli/profile.go:62` | `moai profile --setup`/`-s` 플래그 경로(`runProfileCmd`) |
| `internal/cli/profile_setup.go:214` | `moai profile setup [name]` cobra `RunE` 등록 |
| `internal/cli/init.go:659` | 프로필 없음 확인창 "예" 분기 |
| `internal/cli/update.go:187` | 같은 확인창의 update 쪽 사본 |

(`internal/cli/launcher.go:1104`, `:1106` 은 주석 속 언급이며 호출이 아니다.)

`github.com/charmbracelet/huh`(v1) 을 import 하는 파일 — 5개, 테스트 파일 포함 전수: `huh_theme.go`, `init.go`, `update.go`, `update_version.go`, `profile_setup.go`.

v1 테마 팩토리 `moaiHuhTheme` 소비 지점 — 5곳: `init.go:657`, `update.go:185`, `update_version.go:331`, `profile_setup.go:335`, `profile_setup.go:442`. `moaiHuhStyles` 는 `huh_theme.go` 안과 `huh_theme_test.go` 에서만 쓰인다.

### A.4 흡수 후 사라지는 표면과 남는 표면 (D2-b)

사라지는 것:

- `profile_setup.go` 의 v1 폼 두 개(1단계 언어 폼 `:327-335`, 본 폼 `:354-442`)
- 프로필 없음 확인창 두 개(`init.go:650-662`, `update.go:177-191`)
- v1 테마 팩토리 `huh_theme.go`(`moaiHuhTheme`/`moaiHuhStyles`) — 이 SPEC 이 다운그레이드 확인창(`update_version.go:327`)도 v2 로 옮기므로 흡수 뒤 소비자가 0 이 된다. 다운그레이드 확인창을 v1 에 남겨 두면 v1 테마와 v1 키맵 현지화를 확인창 하나 때문에 따로 유지해야 하므로, 옮기는 쪽을 택한다.

남는 것 (흡수 뒤에도 소비자가 있다):

- `profileSetupText` 구조체와 `getProfileText` — 저장 뒤 요약 출력(`printProfileSummary`, `profile_setup.go:537`), 스키마 옵션 라벨 브리지(`schemaOptionBridge` → `optionLabelFor`, `schema_bridge.go:131-168`), 필드·세그먼트 브리지(`schema_bridge.go:27-111`)가 계속 읽는다.
- `normalizeModel`·`schemaSelectOptions`·`readCurrentProjectConfig`·`persistProjectConfig` 등 v1 폼과 무관한 저장·정규화 함수.

`moai profile setup [name]` 은 명령 표면을 그대로 두고, 폼 실행만 v2 엔진으로 바뀐다(REQ-ITI-007).

### A.5 선행 조건 — 카드 t583 과의 충돌 회피

카드 t583(lane-1, init 위저드 질문 16→4)이 자기 워크트리에서 **커밋하지 않은 상태로** 아래 범위를 고치고 있다. 줄 번호는 lane-1 이 리드에게 보고한 값이며 **이 트리에서 검증하지 않았다**.

| 파일 | t583 이 고치는 범위(보고값) |
|---|---|
| `internal/cli/wizard/questions.go` | `Page3Questions`, `InitQuestions` |
| `internal/cli/wizard/wizard.go` | `397-482`, `35-52`, `176-238` |
| `internal/cli/wizard/translations.go` | (범위 미상) |
| `internal/cli/init.go` | `270-339`, `728-754` |

[HARD] `wizard.go`·`questions.go`·`translations.go`·`init.go` 를 건드리는 run 단계 작업은 **t583 이 develop 에 병합되고 이 워크트리가 그것을 흡수한 뒤에만** 시작한다. 이 네 파일과 겹치지 않는 작업(`huh_theme.go`, `update_version.go`, 검증 하네스와 골든 발판, 새 파일)은 먼저 할 수 있다. 순서는 `plan.md` §F 가 정한다.

### A.6 관련 SPEC 과의 관계

| SPEC | 관계 |
|---|---|
| SPEC-CLI-TUX-V3-002 (completed) | REQ-TUX2-006 단일 멀티그룹 폼, REQ-TUX2-008 동적 단계 표시를 **이어받는다**. REQ-TUX2-012 가 "huh 메이저는 스파이크 결과를 따른다"고 남긴 v1 잔여 표면을 이 SPEC 이 정리한다. 충돌 없음. |
| SPEC-CLI-TUX-INIT-UPDATE-001 (completed) | 배너·체크리스트·진행 표시의 표현 개편. 위저드 폼은 대상이 아니었다. REQ-TUXIU-044(stdout/stderr 채널 규율)는 이 SPEC 에서도 지킨다. 충돌 없음. |
| SPEC-CLI-WIZARD-RESTRUCTURE-001 (completed) | 그 SPEC 의 §C「Out of Scope — wizard rendering engine / theme」가 미뤄 둔 영역을 이 SPEC 이 맡는다. REQ-WIZ-007(언어 변경 즉시 반영)을 프로필 흐름까지 넓힌다(REQ-ITI-008). |
| SPEC-INIT-WIZARD-REPAIR-001 (completed) | 그 SPEC 의 "질문·번역 변경 없음" 제약은 그 SPEC 한정이다. 그 SPEC 이 되살린 세 배선(autonomy tier, 워크플로 토글, audit 블록)은 이 SPEC 이 회귀시키지 않는다. |
| SPEC-WEB-CONSOLE-002 (completed) | REQ-WC2-006 TUI `model_policy` 선택을 흡수 뒤에도 유지한다(REQ-ITI-005). |
| SPEC-WEB-CONSOLE-003 (completed) | REQ-WC3-006 의 `development_mode` 선택과 quality.yaml 저장 경로를 유지한다. 같은 요구의 `git_convention` 선택은 이미 코드에서 빠졌고(`profile_setup.go:427-434` 주석), 이 SPEC 은 되살리지 않는다. |
| SPEC-I18N-GOVERNANCE-001 (completed) | 웹 콘솔 `i18n.js` 카탈로그 관리. 대상 카탈로그가 다르다. 키 동등성 원칙만 빌려 오고 웹 쪽은 건드리지 않는다. |

### A.7 템플릿 영향

이 SPEC 의 변경은 모두 `internal/cli` 와 `internal/cli/wizard` 의 Go 코드와 그 테스트다. `internal/template/templates/**` 는 **바뀌지 않는다**. 따라서 템플릿 중립성 규칙(SPEC ID·날짜·카드 id 금지)이 적용될 산출물이 없다. 사용자 문서(docs-site)에 확인창 문구가 인용돼 있는지는 sync 단계에서 확인한다.

## §B 요구사항 (GEARS)

> 요구 본문은 이 저장소의 GEARS 관례대로 영어 절(`When`/`While`/`Where`/`shall`)로 적는다.

### B.1 프로필 없음 진입 (D1)

- **REQ-ITI-001** (Event-driven): **When** `moai init` or `moai update` runs on an interactive terminal and the active profile is not set up, the CLI **shall** enter the profile setup wizard directly at its first question (conversation language), without presenting any intermediate yes/no confirmation.
- **REQ-ITI-002** (Event-detected): **When** the user cancels the profile setup wizard entered through REQ-ITI-001, or that wizard returns an error, `moai init` and `moai update` **shall** continue their remaining flow exactly as they do today — a cancellation writes no profile and proceeds, an error emits a warning and proceeds — and **shall not** exit non-zero because of the cancellation.
- **REQ-ITI-003** (State-driven): **While** stdin is not a terminal, or `moai init` runs with `--non-interactive`, or `moai update` runs with `--yes`, the CLI **shall not** start the profile setup wizard.

### B.2 v1 프로필 위저드 흡수 (D2)

- **REQ-ITI-004** (Ubiquitous): Every profile setup entry — `moai profile setup [name]`, `moai profile --setup`, and the REQ-ITI-001 entry — **shall** run on the same huh v2 wizard engine as the init wizard, and no non-test Go source **shall** import `github.com/charmbracelet/huh` (v1).
- **REQ-ITI-005** (Ubiquitous): The absorbed profile wizard **shall** collect the same field set as the current wizard (user name; conversation, git-commit, code-comment and documentation language; model; model policy; effort level; permission mode; development mode), **shall** persist each value to the same destination as today (profile `preferences.yaml`, project config sync, `quality.yaml` `development_mode`), and **shall** derive every option value set from the shared settings schema that the web console also reads.
- **REQ-ITI-006** (Ubiquitous — persisted-value preservation): The absorbed profile wizard **shall** preserve these existing behaviors: a deprecated model ID stored in preferences is normalized before it is bound to the model select, so the stored choice is pre-selected rather than silently lost; selecting `acceptEdits` stores an empty permission mode and prints the acceptEdits confirmation line; the development-mode select is initialized from the project's `quality.yaml`, not from preferences; the stored statusline segment map is carried through a save untouched; and every other select pre-selects the currently stored value.
- **REQ-ITI-007** (Ubiquitous): `moai profile setup [name]` **shall** keep its command contract — an optional profile name defaulting to `default`, the session-worktree auto-entry with its scope notice and exit cleanup, the "setup cancelled" exit with status 0 on user abort, and the saved-values summary after a successful save.
- **REQ-ITI-008** (Event-driven): **When** the user answers the conversation-language question in the profile wizard, every subsequently rendered question title, description, option label, confirm button label and help line **shall** render in the selected language.

### B.3 잔여 영어 표면 현지화 (F10)

- **REQ-ITI-009** (Event-driven): **When** `moai update --version <tag>` asks the user to confirm a downgrade, the confirmation **shall** render its title, description, button labels and help line in the conversation language resolved from existing persisted configuration, and **shall** fall back to English when no language resolves.
- **REQ-ITI-010** (Ubiquitous): The help line of every wizard form — the init wizard, the `moai update -c` reconfigure wizard, the profile wizard and the downgrade confirmation — **shall** render its key-action labels in the active locale for each of `en`, `ko`, `ja` and `zh`.
- **REQ-ITI-011** (Unwanted behavior): After this SPEC lands, the wizard and profile translation catalogues **shall not** contain a key that non-test code never reads, and each of `ko`, `ja` and `zh` **shall** define every key the `en` table defines.

### B.4 레이아웃 (F11)

- **REQ-ITI-012** (Ubiquitous): Every confirm field **shall** render its buttons left-aligned with the field card's content column.
- **REQ-ITI-013** (Ubiquitous): A wizard form **shall** render no blank line between consecutive fields, and a select field **shall** render no empty card row below its last option.
- **REQ-ITI-014** (Ubiquitous): **Where** the options of a select field carry descriptions, the field **shall** render the descriptions in one vertically aligned column whose start position is computed from terminal display width, counting each East Asian wide character as two cells.

### B.5 판정 하네스 (D6)

- **REQ-ITI-015** (State-driven): **While** the pty-capture environment variable is set, the test suite **shall** render the wizard surfaces in a real pseudo-terminal and capture the screen for judgement; **while** it is unset, the capture tests **shall** report themselves as skipped. In both states the harness **shall not** write under the real home directory, and every terminal session it starts **shall** be terminated by a registered cleanup.

## §C 제약

- **huh 포크·교체 금지(D3)**: `go.mod` 에 `replace` 지시나 vendor 사본을 추가하지 않는다. huh v2 공개 API 로 되는 조정만 한다.
- **의존성**: 새 모듈을 추가하지 않는다. huh v1 import 가 0 이 된 뒤 `go mod tidy` 로 v1 모듈이 빠지는 것은 허용한다(제거는 새 의존성이 아니다).
- **import 방향**: `internal/cli` → `internal/cli/wizard` 한 방향을 유지한다. 스키마 옵션 라벨은 현재 `internal/cli` 에서 풀리므로, 흡수 뒤에도 옵션 값의 출처는 `settings.FieldOptionDefs` 하나로 남아야 한다.
- **채널 규율**: 요약·경고·확인 문구의 stdout/stderr 배분을 바꾸지 않는다(SPEC-CLI-TUX-INIT-UPDATE-001 REQ-TUXIU-044 승계).
- **테스트 격리**: `t.TempDir()` 만 쓴다. OTEL 환경 변수를 병렬 테스트에서 설정하지 않는다. pty 하네스는 자식 프로세스 환경으로만 HOME 을 바꾸고 `t.Setenv("HOME", …)` 을 쓰지 않는다.
- **로컬 검증 범위**: 변경 패키지(`internal/cli`, `internal/cli/wizard`)만 로컬에서 돌리고, 전체 스위트 판정은 CI 에 맡긴다.
- **크로스 플랫폼**: pty 하네스는 tmux 가 필요한 환경 게이트 테스트이므로 Windows 에서는 건너뛴다. 제품 코드는 `GOOS=windows GOARCH=amd64 go build ./...` 를 통과해야 한다.

## §D 제외 범위

### Out of Scope — 확인창 안쪽 빈 줄 (D3, 알려진 한계)

- huh v2 `field_confirm.go:261-264` 에 박힌 설명과 버튼 사이 `"\n"` 두 번은 그대로 둔다. 포크·`replace`·확인창 자체 구현으로 없애지 않는다.
- 이 한계가 남는 표면: init 위저드와 `moai update -c` 재구성 위저드의 확인형 질문, 흡수된 프로필 위저드에 확인형 질문이 있다면 그 질문, 버전 다운그레이드 확인창. (프로필 없음 확인창 두 개는 D1 로 사라지므로 목록에서 빠진다.)

### Out of Scope — 단계 표시 줄 소실 원인 조사 (D4, 연기)

- 80×30 뷰 문자열에는 있는 단계 표시 줄이 같은 크기 pty 캡처에서 사라지는 현상의 원인 규명과 수리는 이 SPEC 의 요구사항이 아니다. t583 병합 뒤 질문 수가 바뀌므로 그때 다시 재고, 결과를 기록만 한다(`plan.md` §G).

### Out of Scope — 질문 집합 변경

- init 위저드 질문 수·순서 변경은 카드 t583 소관이다. 이 SPEC 은 질문을 더하거나 빼지 않는다.
- 이미 빠진 프로필 위저드 필드(statusline 테마, 세그먼트 MultiSelect, `git_convention`, 중첩 quality·git 자동 감지 필드)는 되살리지 않는다.

### Out of Scope — 웹 콘솔과 스키마 브리지 정리

- 웹 콘솔 화면과 `i18n.js` 카탈로그는 건드리지 않는다.
- `schema_bridge.go` 의 필드·세그먼트 브리지는 현재 비테스트 호출자가 없지만(`uikit.SchemaKeyToTUIField`/`FieldDefTUILabel` 의 비테스트 호출 0건), 삭제 여부는 이 SPEC 이 정하지 않는다. 옵션 라벨 브리지는 흡수된 위저드가 계속 쓴다.

### Out of Scope — 실제 터미널 앱 호환성

- 운영자가 쓰는 개별 터미널 에뮬레이터별 줄바꿈·폭 처리 차이는 판정 대상이 아니다. 판정 기준 환경은 tmux pty 한 가지다.

### Out of Scope — 템플릿과 문서 본문

- `internal/template/templates/**` 변경 없음(§A.7). docs-site 문구 동기화는 sync 단계에서 필요할 때만 한다.

## §E 참조

- `plan.md` — 마일스톤 순서(t583 흡수 게이트 포함), 기술 접근, 위험, 미결 사항
- `acceptance.md` — AC 16개, 판정 방식(pty / 골든 / 기계 검사) 구분, 완료 정의
- 재현 증거: `.moai/reports/t586/verdict.md`, 같은 폴더의 `probe-*.txt`, `tty-*.txt`
