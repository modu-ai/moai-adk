---
id: SPEC-INIT-TUX-I18N-001
title: "init/update/profile wizard TUX repair — v1 profile wizard absorbed into huh v2, layout repair, remaining English surfaces localized"
version: "0.2.3"
status: completed
created: 2026-09-11
updated: 2026-09-13
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, tux, wizard, huh-v2, i18n, init, update, profile, layout, pty"
tier: L
related_specs: [SPEC-CLI-TUI-MODERNIZE-001, SPEC-CLI-TUX-V3-002, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-INIT-HARNESS-PROMPT-001, SPEC-WEB-CONSOLE-002, SPEC-WEB-CONSOLE-003, SPEC-I18N-GOVERNANCE-001]
---

# SPEC-INIT-TUX-I18N-001 — init/update/profile 위저드 TUX 수리와 잔여 영어 표면 현지화

> 카드: **t586** (Factory lane-2, Class C). 재현 근거: `.moai/reports/t586/verdict.md`. 1회차 plan 감사: `.moai/reports/t586/plan-audit.md` (FAIL 0.67, 결함 D1~D17).

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-11 | manager-spec | 최초 작성(카드 t586 plan 단계). 재현 판정과 리드·운영자 결정 D1~D6 을 요구사항으로 옮겼다. |
| 0.2.0 | 2026-09-11 | manager-spec | 1회차 plan 감사 결함 D1~D17 반영. 리드 판정 Q1~Q4 로 미결 4건을 닫았다(init 은 프로필 위저드를 부르지 않음, 도움말은 키별 짧은 라벨, 다운그레이드 확인창 언어 우선순위, 프로필 위저드 단계 표시). SPEC-CLI-TUI-MODERNIZE-001 계약 인수를 명시했다(§A.6.1). 소스 스캔 가드 9건·v1 타입 노출·저장값 보존 4건을 조사와 요구에 넣었다. pty 판정 계약을 공허 초록이 불가능하도록 다시 썼다. t583 충돌 회피 전제를 lane-1 확정 범위로 바꾸고, t583 뒤 남는 그룹 재구성을 새 요구로 넣었다. REQ 18·AC 20 으로 Tier M 요구 상한(16)을 넘어 Tier L 로 올리고 `design.md`·`research.md` 를 보탰다. |
| 0.2.1 | 2026-09-11 | manager-spec | 리드 판정 Q5 반영. REQ-ITI-017 제안(`agent_wiring`·`autonomy_tier` 를 `Agents & Autonomy` 한 그룹으로 묶어 init 위저드를 3페이지에서 2페이지로)을 확정했다. 페이지 수와 스테퍼 분모를 따로 판정하도록 AC 를 나눴다(AC-ITI-018 페이지 수, AC-ITI-021 분모). 그룹 라벨을 읽는 렌더 경로가 없음을 코드에서 재고, 번역 키를 두지 않는 판단을 AC-ITI-022 로 고정했다. `research.md` §13 의 실행 확인 4건을 `plan.md` M1 착수 검증 V-a~V-d 로 옮기고, 그 결과에 기대는 AC 에 선결 표시를 달았다. REQ 18·AC 22, Tier L 유지. |
| 0.2.2 | 2026-09-11 | manager-spec | 2회차 plan 감사 결함 N1~N10 반영. 실제 HOME 무기록 판정을 트리 전체 매니페스트에서 코드로 도출한 감시 목록과 양성 대조군으로 바꿨다(N1). AC-ITI-003 pty 판정에서 단계 표시 줄 조건을 빼고 기준 문자열과 옵션 줄 4개로 도달성을 단정한다(N2). AC-ITI-010 에 양성 가드 성질 제거 뮤턴트 네 개(S2 양성 절·S3·S4·S6)를 더했다(N3). 부재 단정 세 곳(AC-ITI-004 (2)(3), AC-ITI-011 (1))에 같은 형태의 대조군을 붙였다(N4). tmux `-e` 로 변수마다 넘기는 자식 환경 정리 목록과 자식이 기록한 실효 환경 관측을 넣었다(N5). catppuccin 서술 정정(N6), REQ-ITI-017 을 State-driven 으로 고침(N7), tmux 부재 표현을 FAIL 로 통일(N8), t583 SPEC 부재 기록(N9), 세션 목록 비교를 `moai-ptycap-` 접두로 한정(N10). REQ 18·AC 22, Tier L 유지. |
| 0.2.3 | 2026-09-11 | manager-spec | 3회차 plan 감사 결함 F1~F5 반영(운영자가 승인한 좁은 범위 확장). S2 양성 절의 기준을 함수 이름에서 저장 함수 호출 줄(정의 줄 제외)로 바꿔, 호출만 지운 뮤턴트에서 S2 가 실패하도록 했다(F1, `design.md` §10, AC-ITI-010 (2)(4), 측정 `research.md` §17). 조사 기록의 자기 적중 문장을 고치고(F2), `moai update` 잔여 정리 경로의 이관 체크포인트 쓰기를 감시 제외 사유와 함께 기록했으며(F3), REQ-ITI-017 을 Where 패턴으로 바꾸고(F4), 옛 그룹 라벨 판독 문단에 §6.1 대체 표식을 달았다(F5). REQ 18·AC 22, Tier L 유지. |

## §A 배경

### A.1 무엇이 문제인가

`moai init`·`moai update`·`moai profile setup` 의 대화형 화면에는 두 종류의 결함이 있다. 재현과 측정은 판정 문서에 있고, 이 SPEC 은 그 측정을 기준선으로 삼는다.

- **F11 렌더 결함** — 확인 버튼 중앙정렬(들여쓰기 v2 7칸, v1 22칸), 필드 사이 빈 줄(`FieldSeparator` 기본값 `"\n\n"`), 선택 필드 아래 빈 카드 줄(첫 페이지 4줄), 옵션 설명 열 불일치(`Label - Desc` 단순 연결). 네 가지 모두 뷰 문자열과 tmux 80×30 pty 캡처 양쪽에서 재현됐다.
- **F10 영어 고정 표면** — 판정 문서의 I1~I7. 프로필 없음 확인창(I1·I2), v1 확인창 버튼 `Yes`/`No`(I3), v1 프로필 위저드 1단계 언어 선택(I4), 1단계 취소 메시지(I5), 버전 다운그레이드 확인창(I6), v1·v2 전 표면의 도움말 줄(I7).

### A.2 결정으로 좁혀진 범위

리드·운영자 결정 D1~D6(판정 문서)과 1회차 감사 뒤 리드 판정 Q1~Q4 가 대상 목록을 다음과 같이 정한다.

| 판정 대상 | 결정 반영 후 |
|---|---|
| I1·I2 프로필 없음 확인창 | **사라짐** — init·update 모두에서 확인창을 없애고, 두 명령은 프로필 위저드도 부르지 않는다(D1, Q1) |
| I3 v1 확인창 버튼 | **사라짐** — I1·I2 와 함께 없어짐 |
| I4·I5 v1 프로필 위저드 1단계 | **흡수로 해소** — v2 위저드의 대화 언어 질문과 취소 경로가 대신한다(D2) |
| I6 다운그레이드 확인창 | **v2 로 옮기고 현지화** — 언어는 프로젝트 → 활성 프로필 → 영어 순서로 푼다(Q3) |
| I7 도움말 줄 | **키별 짧은 라벨로 현지화** — 문장형 `HelpSelect`/`HelpInput` 은 지운다(Q2) |

Q1 의 결과로 init 에서 대화 언어는 한 번만 묻는다(init 위저드 첫 질문). 프로필 위저드는 사용자가 `moai profile setup [name]` 이나 `moai profile --setup` 을 직접 실행할 때만 뜬다. Q4 에 따라 흡수된 프로필 위저드도 init 위저드와 같은 형식의 단계 표시를 보여 준다.

### A.3 v1 호출 지점 전수 조사

트리 `18144b7ac`(워크트리 `WT-init-tux-i18n`)에서 다시 쟀다. 조사 트리 `e7a7d4bb3` 와 이 트리 사이 `git diff --stat e7a7d4bb3 HEAD -- internal cmd pkg go.mod go.sum` 출력은 비어 있다. 명령과 원문은 `research.md` §1 에 있다.

`runProfileSetup` 호출·등록 지점 — 4곳:

| 위치 | 성격 | 이 SPEC 뒤 |
|---|---|---|
| `internal/cli/profile.go:62` | `moai profile --setup`/`-s` 플래그 경로 | 남음 |
| `internal/cli/profile_setup.go:214` | `moai profile setup [name]` cobra `RunE` 등록 | 남음 |
| `internal/cli/init.go:659` | 프로필 없음 확인창 "예" 분기 | **사라짐** (REQ-ITI-001) |
| `internal/cli/update.go:187` | 같은 확인창의 update 쪽 사본 | **사라짐** (REQ-ITI-001) |

`github.com/charmbracelet/huh`(v1) 을 import 하는 추적 비테스트 Go 파일 — 5개: `huh_theme.go`, `init.go`, `update.go`, `update_version.go`, `profile_setup.go`. v1 을 import 하는 테스트 파일은 `huh_theme_test.go` 하나다.

v1 테마 팩토리 `moaiHuhTheme` 비테스트 소비 지점 — 5곳: `init.go:657`, `update.go:185`, `update_version.go:331`, `profile_setup.go:335`, `profile_setup.go:442`.

### A.4 흡수 뒤 사라지는 표면, 시그니처가 바뀌는 표면, 남는 표면

사라지는 것:

- `profile_setup.go` 의 v1 폼 두 개(1단계 언어 폼 `:327-335`, 본 폼 `:354-442`)
- 프로필 없음 확인창 두 개(`init.go:647-663`, `update.go:174-192`)
- v1 테마 팩토리 `huh_theme.go`(`moaiHuhTheme`·`moaiHuhStyles`·`huhThemeIsDark`)와 `huh_theme_test.go`
- 위저드 번역의 문장형 도움말 `HelpSelect`/`HelpInput`(`wizard/translations.go:18-19`, 로케일별 값 `:542-564`)과 그것을 읽는 `wizard/wizard_test.go:283-298`, `:1256-1260`
- `profileSetupText` 의 폼 전용 필드(1단계 언어 문구, v1 그룹 제목 등) — 흡수 뒤 비테스트 참조가 0 이 되는 것만(REQ-ITI-013)

시그니처가 바뀌는 것 (D4):

| 표면 | 현재 | 처분 | 영향받는 테스트 |
|---|---|---|---|
| `schemaSelectOptions` (`profile_setup.go:124`) | `[]huh.Option[string]`(v1) 반환, `huh.NewOption` 사용 | 버전 중립 `{Label, Value}` 목록을 돌려주도록 바꾸고, v2 위저드에는 이 목록을 인자로 넘긴다(`design.md` §3) | `profile_setup_projectconfig_test.go:160`, `profile_setup_schema_options_test.go:49,68,92,108`, `profile_setup_nested_test.go:107` — 3개 파일, 호출 6곳. `.Key` 는 `.Label` 로 바뀐다 |
| 인라인 언어 옵션 `langOptions` (`profile_setup.go:320-325`) | v1 옵션 리터럴 | 없어짐. 값은 스키마 `languageOptions` 에서, 라벨은 init 위저드와 같은 원어 이름에서 온다(REQ-ITI-004) | — |
| 취소 판별 `errors.Is(err, huh.ErrUserAborted)` (`profile_setup.go:338`, `:445`) | v1 센티널 | v2 위저드의 취소 센티널 `wizard.ErrCancelled`(`wizard.go:135-138` `mapFormErr`)로 판별 | 명령 수준 새 테스트(AC-ITI-007) |
| 다운그레이드 확인창 (`update_version.go:327-334`) | v1 `huh.NewConfirm` + `moaiHuhTheme` | v2 확인창 헬퍼로 이관 | `update_version_test.go` 초록 유지 + 새 골든 |

남는 것:

- `profileSetupText` 구조체와 `getProfileText` — 저장 뒤 요약(`printProfileSummary`), 스키마 옵션 라벨 브리지(`schemaOptionBridge` → `optionLabelFor`, `schema_bridge.go:162-169`), 필드·세그먼트 브리지가 계속 읽는다. 폼 문자열만 위저드 번역 쪽으로 옮긴다.
- `normalizeModel`·`readCurrentProjectConfig`·`persistProjectConfig`·`emitAcceptEditsConfirmation` 등 저장·정규화 함수.
- 명령 표면 `moai profile setup [name]`·`moai profile --setup`.

### A.5 소스 스캔 가드 전수 조사 (D3)

`profile_setup.go` 파일 본문을 읽는 테스트는 9건이다(`git grep -n 'ReadFile("profile_setup.go")' -- 'internal/cli/*_test.go'`). 폼 코드가 옮겨 가면 양성 가드는 실패하고, 음성 가드는 표현이 바뀌어 공허해진다. 가드마다 재조준 대상은 `design.md` §10 이 정한다.

| # | 위치 | 테스트 함수 | 종류 | 단정하는 성질 |
|---|---|---|---|---|
| S1 | `profile_setup_model_policy_test.go:38` | `TestProfileSetup_ModelPolicySelectPresent` | 양성 | model_policy 선택이 `existingPrefs.ModelPolicy` 로 초기화되고 `t.ModelPolicyTitle`·세 정책 값·`ModelPolicy:` 저장을 가진다 |
| S2 | `profile_setup_nested_test.go:27` | `TestTUINestedConfigNoParallelWriter` | 음성+양성 | `yaml.Marshal`·`os.WriteFile` 부재, `persistProjectConfig` 존재 |
| S3 | `profile_setup_nested_test.go:80` | `TestPermissionModeNormalizeAcceptEdits` | 양성 | `permissionMode == defaultPermissionMode` 정규화 존재 |
| S4 | `profile_setup_nested_test.go:117` | `TestTUIEmptyLabelsSchemaSourced` | 양성 | model_policy 빈 라벨을 `settings.EmptyLabelFor("model_policy")` 에서 읽음 |
| S5 | `profile_setup_projectconfig_test.go:145` | `TestProfileSetupConstructsProjectSelects` | 양성 | `&developmentMode` 바인딩 |
| S6 | `schema_bridge_test.go:121` | `TestTUIRendersSchemaFieldSet` | 양성 | 바인딩 9개(`&userName` … `&developmentMode`) |
| S7 | `profile_setup_removed_questions_test.go:61` | `TestWizardOmitsRemovedQuestions` | 음성 | 제거된 질문 표식(`NewMultiSelect`, `&statuslineTheme`, `GitConventionTitle` 등) 부재 |
| S8 | `profile_setup_removed_questions_test.go:87` | `TestWizardWritesNoStatuslineTheme` | 음성 | `StatuslineTheme:` 대입과 테마 이름 부재 |
| S9 | `profile_setup_removed_questions_test.go:116` | `TestWizardCarriesStoredSegmentsIntoPrefs` | 양성 | 세그먼트 맵 통과, 프로젝트 동기화 nil 세그먼트, 빈 convention 인자 |

### A.6 관련 SPEC 과의 관계

| SPEC | 관계 |
|---|---|
| SPEC-CLI-TUI-MODERNIZE-001 (completed) | **계약 일부를 이 SPEC 이 뒤집는다.** 두 팩토리 병존과 v1 쪽 간접 변수를 요구하던 절이 v1 팩토리 삭제와 충돌한다. 인수 내용은 §A.6.1. |
| SPEC-CLI-TUX-V3-002 (completed) | REQ-TUX2-006 단일 멀티그룹 폼, REQ-TUX2-008 동적 단계 표시를 **이어받는다**. REQ-TUX2-012 가 남긴 v1 잔여 표면을 이 SPEC 이 정리한다. 충돌 없음. |
| SPEC-CLI-TUX-INIT-UPDATE-001 (completed) | 배너·체크리스트·진행 표시의 표현 개편. REQ-TUXIU-044(stdout/stderr 채널 규율)는 이 SPEC 에서도 지킨다. 충돌 없음. |
| SPEC-CLI-WIZARD-RESTRUCTURE-001 (completed) | 그 SPEC 의「Out of Scope — wizard rendering engine / theme」가 미뤄 둔 영역을 이 SPEC 이 맡는다. REQ-WIZ-007(언어 변경 즉시 반영)을 프로필 흐름까지 넓힌다(REQ-ITI-007). |
| SPEC-INIT-WIZARD-REPAIR-001 (completed) | 그 SPEC 이 되살린 세 배선(autonomy tier, 워크플로 토글, audit 블록)을 회귀시키지 않는다. |
| SPEC-INIT-HARNESS-PROMPT-001 (completed) | `init.go` 의 프로필 확인창을 언급하지만 REQ-IHP-005 가 질문 총수에 대해 주장하지 않는다고 밝혀 확인창 제거와 충돌하지 않는다. |
| SPEC-WEB-CONSOLE-002 (completed) | REQ-WC2-006 TUI `model_policy` 선택을 흡수 뒤에도 유지한다(REQ-ITI-004). |
| SPEC-WEB-CONSOLE-003 (completed) | REQ-WC3-006 의 `development_mode` 선택과 quality.yaml 저장 경로를 유지한다. `git_convention` 선택은 이미 빠졌고 되살리지 않는다. |
| SPEC-I18N-GOVERNANCE-001 (completed) | 웹 콘솔 `i18n.js` 카탈로그 관리. 키 동등성 원칙만 빌려 오고 웹 쪽은 건드리지 않는다. |

#### A.6.1 SPEC-CLI-TUI-MODERNIZE-001 계약 인수

이 SPEC 은 `internal/cli/huh_theme.go` 를 지운다. 그 SPEC 의 본문은 고치지 않고, 뒤집히는 절과 대체 보장을 여기에 적는다.

| 뒤집히는 절 (SPEC-CLI-TUI-MODERNIZE-001) | 뒤집히는 내용 | 대체 보장 |
|---|---|---|
| 요구 `REQ-TUIM-040` | "huh v1 테마 팩토리와 huh v2 테마 팩토리는 두 개의 분리된 팩토리로 남는다" — v1 팩토리의 존속 | REQ-ITI-009: `internal/cli` 가 그리는 모든 huh 폼은 v2 위저드 테마 팩토리 하나를 쓴다. 병합할 두 번째 팩토리 자체가 없어지므로 "하나로 합치지 않는다"는 금지는 대상이 사라진다 |
| 요구 `REQ-TUIM-045` | "두 팩토리 모두 기존 패키지 수준 간접 변수로 밝기 축을 푼다" — v1 쪽 간접 변수 `huhThemeIsDark` 의 존속 | REQ-ITI-009: 남는 v2 팩토리가 자기 간접 변수(`wizardIsDark`)로 밝기 축을 풀고, 테스트는 환경 변수를 건드리지 않고 두 축을 강제할 수 있다(AC-ITI-011) |
| 인수 기준 `AC-TUIM-026` | `grep -n "func moaiHuhStyles" internal/cli/huh_theme.go` → 1 hit | AC-ITI-011: `huh_theme.go` 부재와 `moaiHuhTheme`·`moaiHuhStyles`·`huhThemeIsDark` 참조 0, 대조군 `var wizardIsDark` 1 hit |
| 인수 기준 `AC-TUIM-029` | `grep -n "var huhThemeIsDark" internal/cli/huh_theme.go` → 1 hit | AC-ITI-011: 다운그레이드 확인창과 프로필 위저드를 `wizardIsDark` 강제 참·거짓으로 각각 그린 골든이 서로 다르고 각 골든과 일치 |

REQ-TUIM-041(두 팩토리에 같은 토큰-역할 배정)은 뒤집지 않는다. 팩토리가 하나가 되므로 그 절은 남는 v2 팩토리의 토큰 배정에 대한 요구로 읽힌다.

**인수 문장**: REQ-TUIM-040 과 REQ-TUIM-045 의 v1 쪽 보장, AC-TUIM-026 과 AC-TUIM-029 의 `internal/cli/huh_theme.go` 판정은 이 SPEC 이 develop 에 병합되는 시점에 이 SPEC 의 REQ-ITI-009 와 AC-ITI-011 로 넘어가며, 그 뒤로는 SPEC-CLI-TUI-MODERNIZE-001 의 해당 절을 현행 계약으로 읽지 않는다.

### A.7 선행 조건 — 카드 t583 과의 충돌 회피

카드 t583(lane-1, SPEC-INIT-QUIET-WIZARD-001, init 질문 18→4)이 같은 위저드 파일을 고친다. 아래 범위는 lane-1 이 확정해 리드에게 넘긴 목록이다(출처: t583 워크트리 `SPEC-INIT-QUIET-WIZARD-001/plan.md` §F.1, 기준 트리 `120436f58`). t583 은 아직 plan 단계이고 커밋되지 않았다. **줄 번호는 t583 병합·흡수 뒤 흡수 트리에서 다시 잰다.** 이 트리에서 함수 시작 줄(`InitQuestions` 296, `Page3Questions` 349, `ReconfigureQuestions` 268, `RunWithDefaults` 35, `buildFormGroups` 158, `stepperDenominator` 236, `saveAnswer` 397, `saveBoolAnswer` 467, `buildConfirmField` 487)은 목록과 일치함을 확인했다(`research.md` §6).

`SPEC-INIT-QUIET-WIZARD-001` 은 t583 이 커밋되기 전이라 이 트리의 `.moai/specs/` 에 없다(HEAD `538b56f19` 에서 `ls -d .moai/specs/SPEC-INIT-QUIET-WIZARD-001` 종료 1, 대조군 `ls -d .moai/specs/SPEC-INIT-TUX-I18N-001` 종료 0, `research.md` §13). 예상된 상태다. 흡수 게이트에서 병합된 SPEC 경로를 다시 확인해 progress 기록에 적는다(`plan.md` §C 5).

| 파일 | t583 이 고치는 범위 | t583 이 건드리지 않는 범위 |
|---|---|---|
| `internal/cli/wizard/questions.go` | `InitQuestions` 본문(296-303), `Page3Questions`(349-527)에서 11문항 리터럴과 딸린 주석 삭제, ID 로 고르는 작은 도우미 추가 | `DefaultQuestions`(48), `GitQuestions`(162), `ReconfigureQuestions`(268-289), `FilteredQuestions`·`TotalVisibleQuestions`·`QuestionByID`, 남는 `agent_wiring`·`autonomy_tier` 리터럴과 그룹 라벨 |
| `internal/cli/wizard/wizard.go` | `saveAnswer` 분기 `project_mode`(434)·`project_continuation`(440)·`audit_model`/`audit_gate_*`(442-449), `saveBoolAnswer`(467) 본문 분기와 위 주석(458-466) | `RunWithDefaults`(35-59), `buildFormGroups`(158~)와 그룹 묶기, 스테퍼(236·242), `buildField`·`buildConfirmField`(487~), 스타일·렌더링 코드 전부 |
| `internal/cli/wizard/types.go` | `WizardResult` 필드 11개 삭제(`ProjectMode`, `WorktreeAutoCreate`, `TodoEnabled`, `FeedbackAutoSubmit`, `ProjectContinuation`, `AuditModel`, `AuditGateClaude`, `AuditGateCodex`, `AuditGateGLM`, `CodexAuditEnabled`, `MCPProvision`) | 나머지 필드 |
| `internal/cli/wizard/translations.go` | ko·ja·zh 표에서 위 11개 ID 항목 삭제 | `project_name`·`model_policy`·`report_format`·Git·`agent_wiring`·`autonomy_tier` 항목, UI 문자열 표(t583 보고값 548 이후; 이 트리의 `var uiStrings` 는 540) |
| t583 위저드 테스트 | t583 plan.md §H 목록만 | 렌더·색·폼 스냅숏 테스트 |
| `internal/cli/init.go` | `applyWizardPage3ToOpts`(277-339) 매핑 삭제, 결과 적용 블록(728-750) 일부 삭제, 대화형 `opts.MCPProvision = true`, `:1004-1011` 규칙과 주석 | 프로필 확인창 블록(647-663)은 목록에 없다 |

t583 뒤 init 질문 집합: Basic 2개(`conversation_language`, `user_name`) / Quality & Workflow 1개(`agent_wiring`) / Autonomy 1개(`autonomy_tier`), 라벨은 그대로다. 확인형(`QuestionTypeConfirm`) 질문 5개(`worktree_auto_create`, `todo_enabled`, `feedback_auto_submit`, `codex_audit_enabled`, `mcp_provision`)는 모두 삭제 목록에 들어 있어, t583 뒤 init·reconfigure 질문 집합에는 확인형 질문이 남지 않는다. `QuestionTypeConfirm` 열거값과 `buildConfirmField` 는 남는다.

**t586 이 새로 맡는 일**: 라벨이 "Quality & Workflow" 인데 `agent_wiring` 하나만 남는 그룹의 정리(REQ-ITI-017). 그룹 라벨은 화면에 그려지지 않고 페이지를 묶는 키로만 쓰이므로(`wizard.go:183-186`), 라벨만 바꾸는 것으로는 사용자가 보는 것이 달라지지 않는다. 리드 판정 Q5 로 두 질문을 `Agents & Autonomy` 한 그룹으로 묶는 안을 확정했다(`design.md` §9, `plan.md` §H). 그 결과 init 위저드는 3페이지에서 2페이지가 되고, 스테퍼 분모는 질문 수 4 그대로다.

[HARD] `questions.go`·`wizard.go`·`types.go`·`translations.go`·`init.go` 를 건드리는 run 단계 작업은 **t583 이 develop 에 병합되고 이 워크트리가 그것을 흡수한 뒤, 흡수 트리에서 기준선을 다시 잰 다음에만** 시작한다. 이 다섯 파일과 겹치지 않는 작업(`update_version.go`, `profile_setup.go` 의 옵션 타입, 새 파일, 검증 하네스와 골든 발판)은 먼저 할 수 있다. 같은 Go 패키지의 최상위 식별자는 한 이름공간을 쓰므로, 게이트 앞에 `wizard` 패키지에 새로 두는 식별자는 흡수 직후 `go build ./internal/cli/wizard/` 로 충돌을 확인한다. 순서는 `plan.md` §F.

### A.8 템플릿 영향

변경은 모두 `internal/cli` 와 `internal/cli/wizard` 의 Go 코드, `go.mod`/`go.sum`, 그 테스트다. `internal/template/templates/**` 는 **바뀌지 않는다**. 사용자 문서(docs-site)에 확인창 문구가 인용돼 있는지는 sync 단계에서 확인한다.

## §B 요구사항 (GEARS)

> 요구 본문은 이 저장소의 GEARS 관례대로 영어 절로 적는다. 패턴 이름은 Ubiquitous / Event-driven / State-driven / Where / Unwanted behavior 중 하나다. Where 는 실행 중에 바뀌는 상태가 아니라 트리 내용·정적 구성에 걸린 조건에 쓴다.

### B.1 프로필 없음 진입 (D1, Q1)

- **REQ-ITI-001** (Event-driven): **When** `moai init` or `moai update` runs while the active profile is not set up, the CLI **shall** continue its flow without presenting any profile yes/no confirmation and without starting the profile setup wizard, whether or not stdin is a terminal and whichever flags were passed.
- **REQ-ITI-002** (Ubiquitous): The interactive `moai init` flow **shall** ask the conversation-language question exactly once, as the first question of the init wizard, and the profile setup wizard **shall** start only when the user runs `moai profile setup [name]` or `moai profile --setup`.

### B.2 v1 프로필 위저드 흡수 (D2, Q4)

- **REQ-ITI-003** (Ubiquitous): Both explicit profile setup entries **shall** run on the same huh v2 wizard engine as the init wizard; no tracked non-test Go source **shall** import `github.com/charmbracelet/huh`, and after `go mod tidy` the `go.mod` file **shall** not require that module.
- **REQ-ITI-004** (Ubiquitous): The absorbed profile wizard **shall** collect the user name; the conversation, git-commit, code-comment and documentation languages; the model; the model policy; the effort level; the permission mode; and the development mode. It **shall** persist them to the same destinations as today — the profile `preferences.yaml`, the project `user.yaml` and `language.yaml` through the project config sync when run inside a MoAI project, and `quality.yaml` `development_mode` — and **shall** take every select's option values from the shared settings schema (`settings.FieldOptionDefs`), except `model_policy`, which the schema no longer declares and whose option values **shall** be `template.ValidModelPolicies()` plus the empty value.
- **REQ-ITI-005** (Ubiquitous): The absorbed profile wizard **shall** preserve these behaviors: (1) a deprecated model ID stored in preferences is normalized before binding, so the stored choice is pre-selected; (2) selecting `acceptEdits` stores an empty permission mode and prints the acceptEdits confirmation line exactly once; (3) the development-mode select is initialized from the project's `quality.yaml`; (4) the stored statusline segment map is written back to `preferences.yaml` unchanged; (5) every other select pre-selects the stored value; (6) the project sync receives a nil segment map, so `statusline.yaml` is left byte-identical; (7) the project config write passes an empty git convention, so `git-convention.yaml` is left byte-identical; (8) the statusline theme is never written, so `statusline_theme` is absent from `preferences.yaml`; (9) an empty stored permission mode pre-selects `acceptEdits`.
- **REQ-ITI-006** (Ubiquitous): `moai profile setup [name]` and `moai profile --setup` **shall** keep their command contract: the profile name defaults to `default` when omitted; session-worktree auto-entry with its scope notice runs before any shared-state read, and its cleanup runs exactly once at exit with a clean-exit flag that is true on success and on cancellation and false on error; a user cancellation prints the localized setup-cancelled line, writes no preferences and exits with status 0; a wizard error returns a non-nil error and writes no preferences; a successful save prints the saved-profile line and the summary.
- **REQ-ITI-007** (Event-driven): **When** the user answers the conversation-language question in the profile wizard, every subsequently rendered group **shall** render its question titles, descriptions, localizable option labels and help-line action labels in the selected language.
- **REQ-ITI-008** (Ubiquitous): The absorbed profile wizard **shall** show a step indicator above each group in the same format as the init wizard's step indicator, with a denominator equal to the number of profile questions currently visible.
- **REQ-ITI-009** (Ubiquitous): Every huh form rendered by `internal/cli` **shall** use the huh v2 wizard theme factory, which **shall** resolve the light/dark axis through its package-level indirection variable, and the huh v1 theme factory and its indirection variable **shall** no longer exist. This requirement supersedes the v1 halves of SPEC-CLI-TUI-MODERNIZE-001 REQ-TUIM-040 and REQ-TUIM-045 and the `internal/cli/huh_theme.go` assertions of AC-TUIM-026 and AC-TUIM-029 (§A.6.1).
- **REQ-ITI-010** (Ubiquitous): Each of the nine source-scan guards listed in §A.5 **shall** keep asserting its property against the source file that holds the corresponding profile question definition or save path after the move, or **shall** be replaced by a behavior test asserting the same property; every positive guard **shall** fail when its asserted construct is removed, and every negative guard **shall** fail when a removed question is reintroduced in the v2 question-definition form.

### B.3 잔여 영어 표면 현지화 (F10, Q2, Q3)

- **REQ-ITI-011** (Event-driven): **When** `moai update --version <tag>` asks the user to confirm a downgrade, the confirmation **shall** render on the huh v2 engine, with its title, description, button labels and help line in the language resolved in this order: the `conversation_language` of `.moai/config/sections/language.yaml` when the working directory holds a `.moai` directory and that value is non-empty; otherwise the active profile's conversation language when non-empty; otherwise English.
- **REQ-ITI-012** (Ubiquitous): The help line of the init wizard, the `moai update -c` reconfigure wizard, the profile wizard and the downgrade confirmation **shall** render each key binding's action label as a short per-key label from the active locale for each of `en`, `ko`, `ja` and `zh`, and the sentence-style help strings `HelpSelect` and `HelpInput` **shall** not exist in any Go source.
- **REQ-ITI-013** (Unwanted behavior): The wizard translation catalogue and `profileSetupText` **shall not** contain a key that non-test code never references, and each of `ko`, `ja` and `zh` **shall** define every key the `en` table defines.

### B.4 레이아웃 (F11, D5)

- **REQ-ITI-014** (Ubiquitous): Every confirm field rendered by `internal/cli`, including the downgrade confirmation, **shall** render its first button at the same display column as its description line.
- **REQ-ITI-015** (Ubiquitous): The init wizard, the reconfigure wizard and the profile wizard **shall** render no blank line between consecutive fields of a group and no empty card row below a select field's last option.
- **REQ-ITI-016** (Event-driven): **When** a select field's options carry descriptions, the field **shall** render the descriptions in one column whose start position is computed from terminal display width, counting each East Asian wide character as two cells.
- **REQ-ITI-017** (Where): **Where** the source tree carries the card t583 init question set, the init question set **shall** place `agent_wiring` and `autonomy_tier` in one group labelled `Agents & Autonomy`, so that the init wizard renders exactly two pages; the init step indicator's denominator **shall** remain the number of visible init questions; no init question **shall** carry the group label `Quality & Workflow`; and no wizard screen **shall** render a question's group label, so the label carries no translation key.

### B.5 판정 하네스 (D6)

- **REQ-ITI-018** (State-driven): **While** the pty-capture environment variable is set, the capture tests **shall** render the wizard surfaces in a fixed-size tmux pseudo-terminal, judge a captured screen only after that surface's anchor text has appeared, and fail on a capture timeout or when tmux is unavailable; **while** it is unset, the capture tests **shall** report themselves as skipped and start no terminal session. In both states the harness **shall** not write under the real home directory, **shall** give every terminal session it starts a unique name, and **shall** terminate only those sessions through a registered cleanup that also runs when a test fails.

## §C 제약

- **huh 포크·교체 금지(D3)**: `go.mod` 에 `replace` 지시나 vendor 사본을 추가하지 않는다. huh v2 공개 API 로 되는 조정만 한다.
- **의존성**: 새 모듈을 추가하지 않는다. huh v1 import 가 0 이 된 뒤 `go mod tidy` 로 v1 모듈과 그것만 끌어오던 간접 의존성 `github.com/charmbracelet/bubbletea v1.3.10`·`github.com/charmbracelet/bubbles v1.0.0`(`go.mod:46-47`)이 빠지는 것은 허용한다. `github.com/catppuccin/go`(`go.mod:44`)는 huh v2 도 요구하므로 tidy 뒤에도 남는다(모듈 그래프 `charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0`, `research.md` §8.1). 남는 줄의 선택 버전은 tidy 를 실행하지 않아 확인하지 않았고, AC-ITI-004 (4) 가 tidy 결과를 커밋본과 대조한다. `github.com/charmbracelet/lipgloss`(v1, `go.mod:15`)는 다른 직접 소비자가 있어 남는다.
- **import 방향**: `internal/cli` → `internal/cli/wizard` 한 방향을 유지한다. 위저드는 `cli` 를 import 하지 않으므로, 스키마 옵션 라벨은 `cli` 에서 풀어 버전 중립 목록으로 위저드에 인자로 넘긴다.
- **채널 규율**: 요약·경고·확인 문구의 stdout/stderr 배분을 바꾸지 않는다(REQ-TUXIU-044 승계).
- **테스트 격리**: `t.TempDir()` 만 쓴다. OTEL 환경 변수를 병렬 테스트에서 설정하지 않는다. pty 하네스는 자식 프로세스 환경으로만 HOME 을 바꾸고 `t.Setenv("HOME", …)` 을 쓰지 않는다. 자식에게는 `acceptance.md` §B 의 자식 환경 정리 목록을 tmux `new-session -e VAR=value` 로 변수마다 넘긴다 — `HOME`·`MOAI_HOME` 은 사례의 임시 경로, `CLAUDE_CONFIG_DIR` 과 `MOAI_KANBAN` 접두 변수 전부는 빈 값. tmux 자식은 테스트 프로세스가 아니라 tmux 서버의 전역 환경을 물려받으므로 부모 쪽 `t.Setenv` 는 정리로 치지 않고, 정리가 닿았는지는 자식이 기록한 실효 환경으로만 판정한다. 인프로세스 골든(AC-ITI-012)은 같은 변수를 `t.Setenv` 로 두고(`HOME` 제외) 병렬로 돌리지 않는다. 실제 HOME 무기록은 트리 전체가 아니라 `acceptance.md` §B P8 감시 목록으로 판정한다.
- **로컬 검증 범위**: 변경 패키지(`internal/cli`, `internal/cli/wizard`)만 로컬에서 돌리고, 전체 스위트 판정은 CI 에 맡긴다.
- **크로스 플랫폼**: pty 하네스는 tmux 가 필요한 환경 게이트 테스트이므로 Windows 에서는 건너뛴다. 제품 코드는 `GOOS=windows GOARCH=amd64 go build ./...` 를 통과해야 한다.
- **t583 게이트**: §A.7 [HARD] 규칙.

## §D 제외 범위

### Out of Scope — 확인창 안쪽 빈 줄 (D3, 알려진 한계)

- huh v2 `field_confirm.go:261-264` 에 박힌 설명과 버튼 사이 `"\n"` 두 번은 그대로 둔다. 포크·`replace`·확인창 자체 구현으로 없애지 않는다.
- 이 한계가 남는 표면: 버전 다운그레이드 확인창, 그리고 앞으로 `QuestionTypeConfirm` 으로 추가될 위저드 질문. t583 뒤 init·reconfigure 질문 집합과 흡수된 프로필 위저드에는 확인형 질문이 없다.

### Out of Scope — 단계 표시 줄 소실 원인 조사 (D4, 연기)

- (sync 재판정, 2026-09-13) 이 제외가 처음 놓였던 관측 — "80×30 뷰 문자열에는 있는 단계 표시 줄이 같은 크기 pty 캡처에서 사라진다" — 는 반증됐다. 같은 80×30 pty 재측정에서 단계 표시 줄 `● ○ ○ ○ 1 / 4` 는 그대로 그려진다(첫 화면 캡처 14번째 줄; 증거 `.moai/reports/t586/sync-d4-recheck.txt`·`.moai/reports/t586/sync-d4-recheck-frame/init-first-screen.txt`, M4 뮤턴트 캡처 `.moai/reports/t586/ac003-init-first-screen-m4mutant-green.txt` 도 같은 결과). 소실이 재현되지 않으므로 그 원인 규명은 대상이 아니고, 기록해 둔 재고(`plan.md` §G)는 여기서 닫는다. 남는 것은 판정 배분이다: 단계 표시 줄의 성질은 AC-ITI-021 이 판정하고 AC-ITI-003 은 판정하지 않으며, REQ-ITI-008 은 뷰 문자열(골든) 기준이라는 점은 그대로다.

### Out of Scope — 질문 집합 변경

- init 위저드 질문의 추가·삭제는 카드 t583 소관이다. 이 SPEC 은 질문을 더하거나 빼지 않는다. REQ-ITI-017 은 남는 질문의 묶음만 바꾼다.
- 이미 빠진 프로필 위저드 필드(statusline 테마, 세그먼트 MultiSelect, `git_convention`, 중첩 quality·git 자동 감지 필드)는 되살리지 않는다.

### Out of Scope — 웹 콘솔과 스키마 브리지 정리

- 웹 콘솔 화면과 `i18n.js` 카탈로그는 건드리지 않는다.
- `schema_bridge.go` 의 필드·세그먼트 브리지의 삭제 여부는 정하지 않는다. 옵션 라벨 브리지는 흡수된 위저드가 계속 쓴다.
- `model_policy` 빈 옵션 라벨이 스키마에서 빈 문자열로 오는 기존 결함(`profile_setup_nested_test.go` 주석)은 고치지 않는다.

### Out of Scope — 실제 터미널 앱 호환성

- 개별 터미널 에뮬레이터별 줄바꿈·폭 처리 차이는 판정 대상이 아니다. 판정 기준 환경은 tmux pty 한 가지다.

### Out of Scope — 템플릿과 문서 본문

- `internal/template/templates/**` 변경 없음(§A.8). docs-site 문구 동기화는 sync 단계에서 필요할 때만 한다.

## §E 참조

- `plan.md` — 마일스톤 순서(t583 흡수 게이트 포함), 기술 접근, 위험
- `acceptance.md` — AC 22개, pty 공통 계약, 픽스처·실제 질문 구분, 완료 정의
- `design.md` — 흡수 구조, 옵션 전달 모양, 도움말 라벨 표, 그룹 재구성 결정, 가드 재조준, pty 하네스
- `research.md` — 이 트리에서 다시 잰 조사 원문
- `progress.md` — 단계별 증거 기록
- 재현 증거: `.moai/reports/t586/verdict.md`, 같은 폴더의 `probe-*.txt`, `tty-*.txt`
