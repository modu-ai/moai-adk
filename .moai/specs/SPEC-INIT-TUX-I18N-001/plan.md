# SPEC-INIT-TUX-I18N-001 — 구현 계획

> 카드 t586 · Tier L · 개발 방식 TDD(`quality.yaml` `development_mode: tdd`) · plan 감사 통과 기준 0.85

## §A 맥락

v1 프로필 위저드를 v2 위저드에 흡수하고, init·update 의 프로필 없음 확인창을 없애고, 위저드 레이아웃과 남은 영어 표면을 고친다. 요구는 `spec.md` §B, 판정은 `acceptance.md`, 구조 결정은 `design.md`, 조사 원문은 `research.md` 에 있다.

**Tier 판단 — L.** 1회차 감사 결함을 반영한 뒤 다시 셌다.

| 항목 | 수 | Tier M 상한 | 판정 |
|---|---|---|---|
| REQ | 18 | 16 | 초과 |
| AC | 20 | 16 | 초과 |
| 제품 코드 파일(추정) | 14 | 5-15 | 경계 |
| 테스트 파일(기존 수정·삭제 + 새 파일, 추정) | 20 이상 | — | 합산 시 15 초과 |

제품 코드 파일 추정: `profile_setup.go`, `profile_setup_translations.go`, `init.go`, `update.go`, `update_version.go`, `huh_theme.go`(삭제), `wizard/wizard.go`, `wizard/questions.go`, `wizard/translations.go`, 새 프로필 질문 파일, 새 확인창 헬퍼 파일, 새 도움말 키맵 파일, `go.mod`, `go.sum`. 요구 상한을 넘었으므로 SPEC 을 쪼개지 않고 Tier L 로 올려 `design.md`·`research.md` 를 보탰다(리드 판정).

## §B 알려진 사실과 코드에서 드러난 불일치

- **init 은 프로필 위저드 없이도 대화 언어를 이미 묻는다.** init 위저드 첫 질문이 `conversation_language` 이고(`wizard/questions.go:63`), 프로필 값이 있으면 기본값으로만 채운다(`init.go:704` `runWizardFn(rootFlag, opts.ConvLang, opts.UserName)`). 확인창과 `runProfileSetup` 호출을 지우면 대화 언어 질문은 한 번만 나온다.
- **그룹 라벨은 화면에 그려지지 않는다.** `buildFormGroups` 는 라벨을 연속 질문을 묶는 키로만 쓰고(`wizard.go:183-186`), 그룹 머리에는 스테퍼 노트만 둔다(`:168`). 라벨만 바꾸면 사용자가 보는 것은 같다.
- **t583 뒤에는 실제 질문 집합에 확인형 질문이 없다.** 삭제 목록 11개에 확인형 5개가 모두 들어 있다(`research.md` §6). 확인창 버튼 정렬의 실제 표면은 다운그레이드 확인창뿐이고, 위저드 확인형 경로는 실제 빌더를 거치는 픽스처로 판정한다(`acceptance.md` §C).
- **`LangSelectTitle`/`LangSelectDesc` 는 코드에서 참조된다.** `schema_bridge.go:34` 브리지가 읽는다. 비테스트 코드 어디에서도 읽히지 않는 키는 `LangGroupTitle` 뿐이다. REQ-ITI-013 의 "읽히지 않는 키" 판정은 참조 여부로 한다.
- **`profileSetupText` 는 흡수 뒤에도 남는다.** 폼 문자열만 옮기고, 요약·브리지용 필드는 구조체에 남긴다.
- **huh v2 도움말 줄은 키 바인딩별 짧은 라벨로 조립된다.** 기본 라벨은 `keymap.go:111-183`(`next`, `submit`, `back`, `select`, `up`, `down`, `filter`, `toggle` 등). 바인딩별 라벨 표를 두고 문장형 두 키는 지운다(Q2).
- **스테퍼는 `*WizardResult` 에 묶여 있다.** `stepperNote`·`stepperDenominator`·`visibleQuestionIndex` 가 `*WizardResult` 를 받는다(`wizard.go:236`, `:242`, `:261`). 프로필 위저드가 같은 형식을 쓰려면 이 결합을 일반화하거나 같은 `tui.Stepper` 호출을 공유해야 하고, 이것은 `wizard.go` 편집이라 게이트 뒤다(`design.md` §4).
- **huh v2 공개 API 확인(모듈 캐시 `charm.land/huh/v2@v2.0.3`).** `(*Confirm).WithButtonAlignment` `field_confirm.go:361`, 기본값 `lipgloss.Center` `:53`; `(*Select[T]).Height` `field_select.go:256`; `(*Form).WithKeyMap` `form.go:284`; `NewDefaultKeyMap` `keymap.go:107`; `FieldSeparator` `theme.go:27`, 기본값 `"\n\n"` `:104`. 확인창 안쪽 `"\n"` 두 번은 `field_confirm.go:261-264` 에 박혀 있다.
- **폭 계산 도구.** `go.mod` 에 `github.com/mattn/go-runewidth` 와 `charm.land/lipgloss/v2` 가 이미 있다.

## §C 사전 점검 (run 단계 첫 커밋 전)

1. `BASELINE_SHA=$(git rev-parse HEAD)` 를 첫 run 커밋 전에 잡아 progress 기록에 남긴다. t583 흡수 뒤에는 흡수 커밋에서 한 번 더 잡는다.
2. `go test ./internal/cli/wizard/... ./internal/cli/ -run 'Profile|Wizard|HuhTheme|UpdateVersion|TUI' -count=1 -v` 기준선. 선택된 테스트 수가 0 이 아닌지 먼저 센다.
3. `acceptance.md` §D.3 RED 원장 명령 4개를 다시 실행해 원문을 progress 기록에 붙인다.
4. pty 조건(tmux, `TERM=xterm-256color`, 80×30) 확인: `command -v tmux; tmux -V`. tmux 가 없으면 pty AC 는 FAIL 이 아니라 **실행 불가 Gap** 으로 보고하고 run 을 멈춘다(판정은 나지 않는다).
5. t583 병합 여부: `git merge-base --is-ancestor <리드가 보고한 t583 병합 SHA> HEAD`.
6. 게이트 뒤 재측정: §A.7 의 줄 번호 목록을 흡수 트리에서 다시 grep 하고 progress 기록에 새 좌표를 적는다.

## §D 제약 (위반 금지)

- `spec.md` §C 제약 전부.
- t583 겹침 파일(`wizard/questions.go`, `wizard/wizard.go`, `wizard/types.go`, `wizard/translations.go`, `init.go`)은 흡수 게이트 전까지 수정하지 않는다.
- 렌더 결함 수리를 소스 문자열 검사로 판정하지 않는다.
- 커밋 메시지마다 카드 id `t586` 을 넣는다. 레인은 push 하지 않는다.
- 기준선 산출물(RED 캡처·골든 RED 출력)은 그것이 재는 수리보다 **먼저, 별도 커밋으로** 넣는다.

## §E 자기 검증 항목 (run 단계 §E 가 인용)

- E1 `acceptance.md` AC 20개 PASS/FAIL 표 — 판정 방식별 명령과 원문 출력. SKIP 은 PASS 로 세지 않는다
- E2 `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`
- E3 `go test -cover ./internal/cli/wizard/...` 와 `./internal/cli/` 변경 경로 커버리지
- E4 RED 출력(수리 전 골든·단위 테스트 실패 원문, 뮤턴트 실패 원문)
- E5 `go vet ./internal/cli/...`, `golangci-lint run` (변경 패키지)
- E6 pty 캡처 파일 경로(`.moai/reports/t586/` 로 반출한 사본)

## §F 마일스톤

순서 원칙: 게이트 앞에서는 t583 파일과 겹치지 않는 작업만 한다. 각 묶음 안에서는 바뀔 가능성이 큰 결정(데이터 모양·타입·사용자 흐름)을 앞에, 기계적 정리를 뒤에 둔다.

### 흡수 게이트 앞 — t583 과 겹치지 않음

#### M1 — 옵션 모양과 프로필 질문 세트 (데이터·타입 결정)

- `schemaSelectOptions` 가 버전 중립 `{Label, Value}` 목록을 돌려주도록 바꾼다(`design.md` §3). 테스트 3개 파일 6곳의 `.Key` → `.Label` 을 함께 고친다. `model_policy` 옵션도 같은 모양으로 만든다(`template.ValidModelPolicies()` + 빈 값).
- 프로필 위저드 질문 세트를 `wizard` 패키지의 **새 파일**로 정의한다(아직 배선하지 않음). 옵션은 인자로 받는다. 질문 id 집합과 그룹 구성은 `design.md` §2.
- 폼 문자열을 4개 로케일로 새 번역 파일에 둔다. `profileSetupText` 쪽 원본은 아직 지우지 않는다.
- 단위 테스트: 질문 id 집합이 REQ-ITI-004 필드 집합과 같음, 옵션 값 집합이 스키마와 같음(AC-ITI-005 일부), 4개 로케일 키 동등.

#### M2 — 다운그레이드 확인창 v2 이관과 언어 해석 (`update_version.go`)

- `update_version.go:327-334` 의 v1 확인창을 v2 확인창으로 옮긴다. v2 확인창 생성은 `wizard` 패키지의 **새 파일** 헬퍼로 둔다(`design.md` §6).
- 언어 해석 함수(프로젝트 → 활성 프로필 → 영어)를 `cli` 쪽에 두고, 위저드 헬퍼에는 풀린 로케일 문자열만 넘긴다.
- AC-ITI-012 골든 네 사례를 RED 로 먼저 세운다. 기존 `update_version_test.go` 초록 유지.

#### M3 — 판정 발판 (pty 하네스 · 골든 헬퍼)

- `acceptance.md` §B 계약을 따르는 pty 캡처 헬퍼를 새 테스트 파일로 만든다: 환경 변수 게이트, 고유 세션 이름, 세션 생성 직후 `t.Cleanup` 등록, 기준 문자열 대기와 기한 초과 FAIL, 임시 HOME·임시 cwd, 실제 HOME 감시 매니페스트.
- AC-ITI-019·020 의 자기 검증(강제 실패·강제 기한 초과·센티널 세션 생존·tmux 부재)을 이 마일스톤에서 먼저 통과시킨다. 하네스가 믿을 만해진 뒤에만 다른 pty AC 를 쓴다.
- ANSI 제거·표시 폭 열 계산 헬퍼, `View()` 골든 비교 헬퍼를 새 테스트 파일로 만든다.
- 다운그레이드 확인창(M2 전 v1 상태)의 수리 전 pty·골든 RED 를 떠서 `.moai/reports/t586/` 로 반출하고 **별도 커밋**한다.

### 흡수 게이트 — t583 develop 병합 확인 → 이 워크트리에서 develop 흡수 → 흡수 트리에서 §C 재측정

게이트 직후 `go build ./internal/cli/wizard/` 로 M1·M2 가 둔 새 식별자와 t583 식별자의 충돌을 확인한다. 흡수 트리에서 init 첫 페이지·그룹 페이지의 수리 전 pty·골든 RED 를 다시 떠 **별도 커밋**한다(렌더 AC 의 공식 RED).

#### M4 — init·update 흐름에서 프로필 경로 제거 (`init.go`, `update.go`) — 사용자 흐름 결정

- `init.go:647-663`, `update.go:174-192` 의 확인창과 `runProfileSetup` 호출을 지운다(REQ-ITI-001). init 은 곧바로 기존 경로로 이어진다.
- AC-ITI-001·002 를 RED 로 세운 뒤 지운다. init 흐름에서 프로필 위저드 실행 이음새 호출 횟수를 셀 수 있도록 이음새를 둔다.

#### M5 — 흡수 배선과 스테퍼 (`profile_setup.go`, `wizard.go`)

- `runProfileSetup` 이 v2 위저드를 실행하게 바꾼다. 두 명시 진입이 이 함수를 지난다.
- 스테퍼를 프로필 위저드에서도 같은 형식으로 쓰도록 결합을 일반화한다(`design.md` §4, REQ-ITI-008).
- 취소 판별을 `wizard.ErrCancelled` 로 옮긴다. 명령 수준 이음새(위저드 실행, 세션 worktree 진입·정리)를 두고 AC-ITI-007 을 먼저 RED 로 세운다.
- REQ-ITI-005 보존 동작 아홉 가지를 표 테스트로 먼저 RED 로 세운 뒤 옮긴다. `profile_setup.go:40-41` 의 `@MX:NOTE`/`@MX:REASON` 은 새 바인딩 지점으로 옮긴다.
- 소스 스캔 가드 9건을 `design.md` §10 대로 재조준하고 AC-ITI-010 뮤턴트를 돈다.

#### M6 — v1 퇴역과 모듈 정리

- `huh_theme.go`, `huh_theme_test.go` 삭제. `go mod tidy`. AC-ITI-004·011.
- `launcher.go:1094-1110` 주석(`runProfileSetup` 정규화 블록을 가리킴)이 새 구조에서도 참인지 다시 읽고, 거짓이 된 부분만 고친다.

#### M7 — 레이아웃·도움말·그룹 (`wizard.go`, `translations.go`, `questions.go`)

- 확인창 버튼 왼쪽 정렬(REQ-ITI-014), 필드 구분자 축소와 선택 필드 높이(REQ-ITI-015), 옵션 설명 열 정렬(REQ-ITI-016, `wizard.go:290` 단순 연결 대체), 키별 도움말 라벨(REQ-ITI-012), `agent_wiring`·`autonomy_tier` 그룹 재구성(REQ-ITI-017).
- M1·M2 에서 새 파일에 둔 문구를 `translations.go` 체계로 합친다.
- 각 항목은 골든 RED → 수리 → 골든 GREEN, 이어서 pty 캡처로 수리 판정.

#### M8 — 키 정리와 판정 마감 (기계적)

- `profileSetupText` 의 폼 전용 필드, `LangGroupTitle`, `HelpSelect`/`HelpInput` 과 그 테스트(`wizard_test.go:283-298`, `:1256-1260`) 등 읽히지 않게 된 키를 지운다(REQ-ITI-013). 키 참조 스윕·로케일 동등성 테스트와 뮤턴트.
- 골든 확정, pty 판정 캡처 반출, §E 증거 정리.

## §G 연기 항목

- **D4 단계 표시 줄 소실.** t583 흡수 뒤 M3 하네스로 80×30 에서 다시 캡처해, 증상이 남는지와 조건만 progress 기록에 적는다. 원인 규명·수리는 이 SPEC 의 AC 가 아니다.

## §H 결정 기록 (1회차 감사 뒤 리드 판정)

| 질문 | 판정 | 반영 위치 |
|---|---|---|
| Q1 대화 언어 이중 질문 | init 은 프로필 위저드를 부르지 않는다. 대화 언어는 init 위저드 첫 질문에서 한 번만 묻는다. 프로필 위저드는 명시 경로에서만 뜬다. 확인창은 init·update 모두에서 없앤다 | REQ-ITI-001·002, AC-ITI-001·002·003, M4 |
| Q2 도움말 줄 형식 | 키별 짧은 라벨 번역을 쓰고 문장형 `HelpSelect`/`HelpInput` 은 지운다 | REQ-ITI-012, AC-ITI-013, `design.md` §7, M7·M8 |
| Q3 다운그레이드 확인창 언어 | 프로젝트 `language.yaml` → 활성 프로필 → 영어 | REQ-ITI-011, AC-ITI-012, M2 |
| Q4 프로필 위저드 단계 표시 | init 위저드와 같은 형식 | REQ-ITI-008, AC-ITI-009, M5 |

남은 확인 필요 표식은 없다.

## §I 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| t583 병합이 늦어짐 | M4~M8 착수 지연 | M1~M3 을 먼저 끝내 두고, 게이트에서 흡수 트리 재측정 |
| t583 이 약속 밖 구간을 고침 | 수리 지점 이동, 충돌 | 게이트에서 §A.7 좌표를 다시 재고 달라진 곳을 progress 에 기록 |
| 게이트 앞 새 식별자가 t583 식별자와 겹침 | 흡수 직후 빌드 실패 | 새 식별자에 프로필·확인창 접두를 쓰고, 게이트 직후 `go build ./internal/cli/wizard/` |
| 스테퍼 일반화가 init 스테퍼를 바꿈 | init 골든 회귀 | AC-ITI-009 의 init 대조군이 같은 형식 규칙을 검사 |
| 옵션 목록 전달 방식이 import 순환을 만듦 | 빌드 실패 | M1 에서 버전 중립 목록을 `cli` 에서 만들어 인자로 넘기는 모양을 먼저 확정 |
| 선택 필드 높이를 옵션 수에 고정하면 긴 목록이 잘림 | 옵션 가려짐 | 높이는 옵션 수와 터미널 높이를 함께 고려, 골든에 긴 목록 사례 포함 |
| tmux pty 폭 처리와 실제 터미널 차이 | pty 판정이 실제와 어긋남 | 판정 문서 Residual-risk 승계, 표시 폭 계산을 골든으로도 검증 |
| 가드 재조준이 공허 | 제거된 질문 재등장 미검출 | AC-ITI-010 의 v2 형태 뮤턴트와 스캔 대상 기준 문자열 |
| pty 하네스가 사용자 tmux 세션을 지움 | 운영자 작업 손실 | 정확한 이름으로만 종료, `kill-server`·접두 일괄 삭제 금지, 센티널 세션 생존 검사(AC-ITI-020) |
| 기준선과 수리가 같은 커밋 | 순서 주장 검증 불가 | §D 규칙: RED 산출물 선행 별도 커밋 |
